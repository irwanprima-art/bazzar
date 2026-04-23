package usecase

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/xuri/excelize/v2"

	invDomain "github.com/irwan/bazzar/internal/modules/inventory/domain"
	invUC "github.com/irwan/bazzar/internal/modules/inventory/usecase"
	eventRepo "github.com/irwan/bazzar/internal/modules/event/repository"
	"github.com/irwan/bazzar/internal/modules/inbound/domain"
	"github.com/irwan/bazzar/internal/modules/inbound/repository"
	skuRepo "github.com/irwan/bazzar/internal/modules/sku/repository"
	skuDomain "github.com/irwan/bazzar/internal/modules/sku/domain"
)

type InboundUsecase struct {
	repo      *repository.InboundRepository
	skuRepo   *skuRepo.SKURepository
	eventRepo *eventRepo.EventRepository
	invUC     *invUC.InventoryUsecase
}

func NewInboundUsecase(
	repo *repository.InboundRepository,
	sr *skuRepo.SKURepository,
	er *eventRepo.EventRepository,
	iu *invUC.InventoryUsecase,
) *InboundUsecase {
	return &InboundUsecase{repo: repo, skuRepo: sr, eventRepo: er, invUC: iu}
}

func (u *InboundUsecase) ImportFromExcel(ctx context.Context, reader io.Reader, eventID uuid.UUID, refNumber string, userID uuid.UUID) (*domain.InboundOrder, error) {
	f, err := excelize.OpenReader(reader)
	if err != nil {
		return nil, errors.New("failed to parse Excel file")
	}
	defer f.Close()

	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		return nil, errors.New("no sheets found")
	}

	rows, err := f.GetRows(sheets[0])
	if err != nil || len(rows) < 2 {
		return nil, errors.New("file is empty")
	}

	headers := rows[0]
	colMap := mapColumns(headers)

	order := &domain.InboundOrder{
		ID:              uuid.New(),
		EventID:         eventID,
		ReferenceNumber: refNumber,
		Status:          "pending",
		ImportedBy:      &userID,
	}

	if err := u.repo.Create(ctx, order); err != nil {
		return nil, errors.New("failed to create inbound order (duplicate reference?)")
	}

	for _, row := range rows[1:] {
		skuCode := getCell(row, colMap["sku"])
		if skuCode == "" {
			skuCode = getCell(row, colMap["sku_code"])
		}
		if skuCode == "" {
			continue
		}

		qtyStr := getCell(row, colMap["qty"])
		if qtyStr == "" {
			qtyStr = getCell(row, colMap["quantity"])
		}
		qty, _ := strconv.Atoi(qtyStr)
		if qty <= 0 {
			qty = 1
		}

		// Auto-create SKU if not exists
		sku, err := u.skuRepo.GetBySKUCode(ctx, skuCode)
		if err != nil {
			name := getCell(row, colMap["name"])
			if name == "" {
				name = getCell(row, colMap["product_name"])
			}
			if name == "" {
				name = skuCode
			}
			barcode := getCell(row, colMap["barcode"])

			newSku := &skuDomain.SKU{
				ID:             uuid.New(),
				SKUCode:        skuCode,
				Name:           name,
				ReplenishLimit: 5,
			}
			if barcode != "" {
				newSku.Barcode = &barcode
			}
			u.skuRepo.UpsertBySKUCode(ctx, newSku)
			sku, _ = u.skuRepo.GetBySKUCode(ctx, skuCode)
			if sku == nil {
				continue
			}
		}

		item := &domain.InboundItem{
			ID:             uuid.New(),
			InboundOrderID: order.ID,
			SKUID:          sku.ID,
			QtyExpected:    qty,
		}
		u.repo.CreateItem(ctx, item)
	}

	items, _ := u.repo.GetItems(ctx, order.ID)
	order.Items = items
	return order, nil
}

func (u *InboundUsecase) CreateManual(ctx context.Context, req domain.CreateInboundRequest, userID uuid.UUID) (*domain.InboundOrder, error) {
	order := &domain.InboundOrder{
		ID:              uuid.New(),
		EventID:         req.EventID,
		ReferenceNumber: req.ReferenceNumber,
		Status:          "pending",
		Notes:           req.Notes,
		ImportedBy:      &userID,
	}

	if err := u.repo.Create(ctx, order); err != nil {
		return nil, errors.New("failed to create inbound order")
	}

	for _, it := range req.Items {
		sku, err := u.skuRepo.GetBySKUCode(ctx, it.SKUCode)
		if err != nil {
			continue
		}
		item := &domain.InboundItem{
			ID:             uuid.New(),
			InboundOrderID: order.ID,
			SKUID:          sku.ID,
			QtyExpected:    it.QtyExpected,
		}
		u.repo.CreateItem(ctx, item)
	}

	items, _ := u.repo.GetItems(ctx, order.ID)
	order.Items = items
	return order, nil
}

func (u *InboundUsecase) ScanItem(ctx context.Context, inboundID uuid.UUID, req domain.InboundScanRequest, eventID, userID uuid.UUID) (*domain.InboundScanResult, error) {
	sku, err := u.skuRepo.GetByBarcode(ctx, req.Barcode)
	if err != nil {
		return nil, fmt.Errorf("unknown barcode: %s", req.Barcode)
	}

	items, err := u.repo.GetItems(ctx, inboundID)
	if err != nil {
		return nil, errors.New("failed to get inbound items")
	}

	for _, item := range items {
		if item.SKUID == sku.ID {
			scanQty := req.Qty
			if scanQty <= 0 {
				scanQty = 1
			}

			newReceived := item.QtyReceived + scanQty
			u.repo.UpdateItemReceived(ctx, item.ID, newReceived)

			// Add to inventory at the target location (default: STORAGE for inbound)
			loc, err := u.eventRepo.GetLocationByCode(ctx, eventID, "STORAGE")
			if err == nil {
				u.invUC.UpsertStock(ctx, sku.ID, loc.ID, scanQty)
				u.invUC.CreateLog(ctx, &invDomain.InventoryLog{
					ID:            uuid.New(),
					SKUID:         sku.ID,
					LocationID:    loc.ID,
					EventID:       eventID,
					Action:        "inbound",
					QtyChange:     scanQty,
					ReferenceID:   &inboundID,
					ReferenceType: "inbound",
					UserID:        &userID,
				})
			}

			// Check if all items are complete
			u.checkAndUpdateStatus(ctx, inboundID)

			return &domain.InboundScanResult{
				ItemID:       item.ID,
				SKUCode:      sku.SKUCode,
				SKUName:      sku.Name,
				QtyExpected:  item.QtyExpected,
				QtyReceived:  newReceived,
				QtyRemaining: item.QtyExpected - newReceived,
				Message:      "OK",
			}, nil
		}
	}

	return nil, fmt.Errorf("item %s not in this inbound PO", sku.SKUCode)
}

func (u *InboundUsecase) checkAndUpdateStatus(ctx context.Context, inboundID uuid.UUID) {
	items, err := u.repo.GetItems(ctx, inboundID)
	if err != nil {
		return
	}
	allComplete := true
	anyReceived := false
	for _, it := range items {
		if it.QtyReceived > 0 {
			anyReceived = true
		}
		if it.QtyReceived < it.QtyExpected {
			allComplete = false
		}
	}
	if allComplete {
		u.repo.UpdateStatus(ctx, inboundID, "completed")
	} else if anyReceived {
		u.repo.UpdateStatus(ctx, inboundID, "partial")
	}
}

func (u *InboundUsecase) GetByID(ctx context.Context, id uuid.UUID) (*domain.InboundOrder, error) {
	order, err := u.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	items, _ := u.repo.GetItems(ctx, id)
	order.Items = items
	return order, nil
}

func (u *InboundUsecase) List(ctx context.Context, eventID uuid.UUID) ([]domain.InboundOrder, error) {
	return u.repo.List(ctx, eventID)
}

func mapColumns(headers []string) map[string]int {
	m := make(map[string]int)
	for i, h := range headers {
		m[strings.ToLower(strings.TrimSpace(h))] = i
	}
	return m
}

func getCell(row []string, idx int) string {
	if idx < 0 || idx >= len(row) {
		return ""
	}
	return strings.TrimSpace(row[idx])
}
