package usecase

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/xuri/excelize/v2"

	invDomain "github.com/irwan/bazzar/internal/modules/inventory/domain"
	invUC "github.com/irwan/bazzar/internal/modules/inventory/usecase"
	"github.com/irwan/bazzar/internal/modules/order/domain"
	"github.com/irwan/bazzar/internal/modules/order/repository"
	skuRepo "github.com/irwan/bazzar/internal/modules/sku/repository"
	eventRepo "github.com/irwan/bazzar/internal/modules/event/repository"
)

type OrderUsecase struct {
	repo      *repository.OrderRepository
	skuRepo   *skuRepo.SKURepository
	eventRepo *eventRepo.EventRepository
	invUC     *invUC.InventoryUsecase
}

func NewOrderUsecase(
	repo *repository.OrderRepository,
	sr *skuRepo.SKURepository,
	er *eventRepo.EventRepository,
	iu *invUC.InventoryUsecase,
) *OrderUsecase {
	return &OrderUsecase{repo: repo, skuRepo: sr, eventRepo: er, invUC: iu}
}

// ImportFromExcel parses Shopee export and imports orders
func (u *OrderUsecase) ImportFromExcel(ctx context.Context, reader io.Reader, eventID, userID uuid.UUID) (*domain.ImportResult, error) {
	f, err := excelize.OpenReader(reader)
	if err != nil {
		return nil, errors.New("failed to parse Excel file")
	}
	defer f.Close()

	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		return nil, errors.New("no sheets found in file")
	}

	rows, err := f.GetRows(sheets[0])
	if err != nil {
		return nil, errors.New("failed to read sheet rows")
	}

	if len(rows) < 2 {
		return nil, errors.New("file is empty or has no data rows")
	}

	// Map header columns
	headers := rows[0]
	colMap := mapColumns(headers)

	result := &domain.ImportResult{}

	for i, row := range rows[1:] {
		result.TotalRows++
		rowNum := i + 2

		orderNum := getCell(row, colMap["no. pesanan"])
		if orderNum == "" {
			result.Errors++
			result.ErrorDetails = append(result.ErrorDetails, fmt.Sprintf("Row %d: missing order number", rowNum))
			continue
		}

		// Check shipping option - only import "Jasa Kirim Toko"
		shippingOpt := getCell(row, colMap["opsi pengiriman"])
		if !strings.Contains(strings.ToLower(shippingOpt), "jasa kirim toko") {
			result.Skipped++
			result.SkippedDetails = append(result.SkippedDetails,
				fmt.Sprintf("Row %d: %s - shipping: %s (not Jasa Kirim Toko)", rowNum, orderNum, shippingOpt))
			continue
		}

		platformStatus := getCell(row, colMap["status pesanan"])
		skuCode := getCell(row, colMap["nomor referensi sku"])
		productName := getCell(row, colMap["nama produk"])
		variationName := getCell(row, colMap["nama variasi"])
		buyerName := getCell(row, colMap["nama penerima"])
		buyerUsername := getCell(row, colMap["username (pembeli)"])
		trackingNum := getCell(row, colMap["no. resi"])
		notes := getCell(row, colMap["catatan dari pembeli"])
		qtyStr := getCell(row, colMap["jumlah"])
		totalStr := getCell(row, colMap["total pembayaran"])

		qty, _ := strconv.Atoi(qtyStr)
		if qty <= 0 {
			qty = 1
		}
		totalPayment, _ := strconv.ParseFloat(strings.ReplaceAll(totalStr, ".", ""), 64)

		// Determine order status
		orderStatus := "imported"
		if domain.IssueStatuses[platformStatus] {
			orderStatus = "issue"
		}

		order := &domain.Order{
			ID:             uuid.New(),
			EventID:        eventID,
			OrderNumber:    orderNum,
			PlatformStatus: platformStatus,
			Status:         orderStatus,
			BuyerName:      buyerName,
			BuyerUsername:   buyerUsername,
			ShippingOption: shippingOpt,
			TrackingNumber: trackingNum,
			ProductName:    productName,
			VariationName:  variationName,
			Notes:          notes,
			TotalPayment:   totalPayment,
			ImportedBy:     &userID,
		}

		inserted, err := u.repo.UpsertOrder(ctx, order)
		if err != nil {
			result.Errors++
			result.ErrorDetails = append(result.ErrorDetails, fmt.Sprintf("Row %d: %s - %v", rowNum, orderNum, err))
			continue
		}

		if !inserted {
			result.Duplicates++
			continue
		}

		// Look up SKU
		var skuID *uuid.UUID
		if skuCode != "" {
			sku, err := u.skuRepo.GetBySKUCode(ctx, skuCode)
			if err == nil {
				skuID = &sku.ID
			}
		}

		// Create order item
		item := &domain.OrderItem{
			ID:            uuid.New(),
			OrderID:       order.ID,
			SKUID:         skuID,
			SKUCode:       skuCode,
			ProductName:   productName,
			VariationName: variationName,
			QtyOrdered:    qty,
		}
		u.repo.UpsertOrderItem(ctx, item)

		// Auto-allocate if not an issue order and SKU is known
		if orderStatus == "imported" && skuID != nil {
			u.allocateOrder(ctx, order, item, eventID)
		}

		result.Imported++
	}

	return result, nil
}

func (u *OrderUsecase) allocateOrder(ctx context.Context, order *domain.Order, item *domain.OrderItem, eventID uuid.UUID) {
	eventLoc, err := u.eventRepo.GetLocationByCode(ctx, eventID, "EVENT")
	if err != nil {
		return
	}

	inv, err := u.invUC.GetBySkuAndLocation(ctx, *item.SKUID, eventLoc.ID)
	if err != nil || inv.Available < item.QtyOrdered {
		return // Not enough stock, stay as 'imported'
	}

	// Allocate
	u.invUC.AddAllocated(ctx, *item.SKUID, eventLoc.ID, item.QtyOrdered)

	now := time.Now()
	u.repo.UpdateStatus(ctx, order.ID, "allocated", map[string]interface{}{
		"allocated_at": now,
	})

	// Log allocation
	u.invUC.CreateLog(ctx, &invDomain.InventoryLog{
		ID:            uuid.New(),
		SKUID:         *item.SKUID,
		LocationID:    eventLoc.ID,
		EventID:       eventID,
		Action:        "allocate",
		QtyChange:     -item.QtyOrdered,
		ReferenceID:   &order.ID,
		ReferenceType: "order",
		UserID:        order.ImportedBy,
	})
}

func (u *OrderUsecase) AllocateAll(ctx context.Context, eventID, userID uuid.UUID) (int, error) {
	filter := domain.OrderFilter{EventID: eventID, Status: "imported", Page: 1, PageSize: 1000}
	orders, _, err := u.repo.List(ctx, filter)
	if err != nil {
		return 0, err
	}

	allocated := 0
	for _, order := range orders {
		items, err := u.repo.GetOrderItems(ctx, order.ID)
		if err != nil || len(items) == 0 {
			continue
		}
		for _, item := range items {
			if item.SKUID == nil {
				continue
			}
			o := order
			o.ImportedBy = &userID
			u.allocateOrder(ctx, &o, &item, eventID)
			allocated++
		}
	}
	return allocated, nil
}

func (u *OrderUsecase) GetByID(ctx context.Context, id uuid.UUID) (*domain.Order, error) {
	order, err := u.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	items, _ := u.repo.GetOrderItems(ctx, id)
	order.Items = items
	return order, nil
}

func (u *OrderUsecase) List(ctx context.Context, filter domain.OrderFilter) ([]domain.Order, int64, error) {
	return u.repo.List(ctx, filter)
}

func (u *OrderUsecase) MarkPrinted(ctx context.Context, orderID, userID uuid.UUID) error {
	order, err := u.repo.GetByID(ctx, orderID)
	if err != nil {
		return errors.New("order not found")
	}
	if order.Status != "allocated" {
		return fmt.Errorf("order must be allocated first (current: %s)", order.Status)
	}
	now := time.Now()
	return u.repo.UpdateStatus(ctx, orderID, "printed", map[string]interface{}{
		"printed_at": now,
		"printed_by": userID,
	})
}

func (u *OrderUsecase) GetStatusCounts(ctx context.Context, eventID uuid.UUID) (map[string]int, error) {
	return u.repo.GetStatusCounts(ctx, eventID)
}

func (u *OrderUsecase) GetOrderItems(ctx context.Context, orderID uuid.UUID) ([]domain.OrderItem, error) {
	return u.repo.GetOrderItems(ctx, orderID)
}

func (u *OrderUsecase) UpdateStatus(ctx context.Context, id uuid.UUID, status string, updates map[string]interface{}) error {
	return u.repo.UpdateStatus(ctx, id, status, updates)
}

func (u *OrderUsecase) UpdateItemPicked(ctx context.Context, itemID uuid.UUID, qty int) error {
	return u.repo.UpdateItemPicked(ctx, itemID, qty)
}

func (u *OrderUsecase) GetByOrderNumber(ctx context.Context, eventID uuid.UUID, orderNum string) (*domain.Order, error) {
	order, err := u.repo.GetByOrderNumber(ctx, eventID, orderNum)
	if err != nil {
		return nil, err
	}
	items, _ := u.repo.GetOrderItems(ctx, order.ID)
	order.Items = items
	return order, nil
}

// Column mapping helpers
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
