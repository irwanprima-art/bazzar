package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	invDomain "github.com/irwan/bazzar/internal/modules/inventory/domain"
	invUC "github.com/irwan/bazzar/internal/modules/inventory/usecase"
	orderUC "github.com/irwan/bazzar/internal/modules/order/usecase"
	"github.com/irwan/bazzar/internal/modules/picking/domain"
	skuRepo "github.com/irwan/bazzar/internal/modules/sku/repository"
	eventRepo "github.com/irwan/bazzar/internal/modules/event/repository"
)

type PickingUsecase struct {
	orderUC   *orderUC.OrderUsecase
	skuRepo   *skuRepo.SKURepository
	eventRepo *eventRepo.EventRepository
	invUC     *invUC.InventoryUsecase
}

func NewPickingUsecase(
	ou *orderUC.OrderUsecase,
	sr *skuRepo.SKURepository,
	er *eventRepo.EventRepository,
	iu *invUC.InventoryUsecase,
) *PickingUsecase {
	return &PickingUsecase{orderUC: ou, skuRepo: sr, eventRepo: er, invUC: iu}
}

func (u *PickingUsecase) StartPicking(ctx context.Context, orderID, userID uuid.UUID) error {
	order, err := u.orderUC.GetByID(ctx, orderID)
	if err != nil {
		return errors.New("order not found")
	}
	if order.Status != "printed" {
		return fmt.Errorf("order must be printed first (current: %s)", order.Status)
	}

	now := time.Now()
	return u.orderUC.UpdateStatus(ctx, orderID, "picking", map[string]interface{}{
		"assigned_picker_id":  userID,
		"picking_started_at": now,
	})
}

func (u *PickingUsecase) ScanItem(ctx context.Context, orderID uuid.UUID, req domain.ScanRequest, userID uuid.UUID) (*domain.ScanResult, error) {
	order, err := u.orderUC.GetByID(ctx, orderID)
	if err != nil {
		return nil, errors.New("order not found")
	}
	if order.Status != "picking" {
		return nil, fmt.Errorf("order is not in picking state (current: %s)", order.Status)
	}

	// Find SKU by barcode
	sku, err := u.skuRepo.GetByBarcode(ctx, req.Barcode)
	if err != nil {
		return nil, fmt.Errorf("unknown barcode: %s", req.Barcode)
	}

	// Find matching order item
	items, _ := u.orderUC.GetOrderItems(ctx, orderID)
	for _, item := range items {
		if item.SKUID != nil && *item.SKUID == sku.ID {
			scanQty := req.Qty
			if scanQty <= 0 {
				scanQty = 1
			}

			newPicked := item.QtyPicked + scanQty
			if newPicked > item.QtyOrdered {
				return &domain.ScanResult{
					ItemID:     item.ID,
					SKUCode:    sku.SKUCode,
					SKUName:    sku.Name,
					QtyOrdered: item.QtyOrdered,
					QtyPicked:  item.QtyPicked,
					QtyRemain:  item.QtyOrdered - item.QtyPicked,
					Message:    fmt.Sprintf("Cannot exceed ordered qty! Ordered: %d, Already picked: %d", item.QtyOrdered, item.QtyPicked),
				}, nil
			}

			u.orderUC.UpdateItemPicked(ctx, item.ID, newPicked)

			return &domain.ScanResult{
				ItemID:     item.ID,
				SKUCode:    sku.SKUCode,
				SKUName:    sku.Name,
				QtyOrdered: item.QtyOrdered,
				QtyPicked:  newPicked,
				QtyRemain:  item.QtyOrdered - newPicked,
				IsComplete: newPicked == item.QtyOrdered,
				Message:    "OK",
			}, nil
		}
	}

	return nil, fmt.Errorf("item %s not found in this order", sku.SKUCode)
}

func (u *PickingUsecase) CompletePicking(ctx context.Context, orderID, userID uuid.UUID) error {
	order, err := u.orderUC.GetByID(ctx, orderID)
	if err != nil {
		return errors.New("order not found")
	}
	if order.Status != "picking" {
		return fmt.Errorf("order is not in picking state (current: %s)", order.Status)
	}

	// Verify all items are fully picked
	for _, item := range order.Items {
		if item.QtyPicked < item.QtyOrdered {
			return fmt.Errorf("item %s not fully picked (picked: %d/%d)", item.SKUCode, item.QtyPicked, item.QtyOrdered)
		}
	}

	now := time.Now()
	return u.orderUC.UpdateStatus(ctx, orderID, "picked", map[string]interface{}{
		"picked_at": now,
		"picked_by": userID,
	})
}

func (u *PickingUsecase) Ship(ctx context.Context, orderNumber string, eventID, userID uuid.UUID) error {
	order, err := u.orderUC.GetByOrderNumber(ctx, eventID, orderNumber)
	if err != nil {
		return errors.New("order not found")
	}
	if order.Status != "picked" {
		return fmt.Errorf("order must be picked first (current: %s)", order.Status)
	}

	// Deduct inventory
	eventLoc, err := u.eventRepo.GetLocationByCode(ctx, eventID, "EVENT")
	if err == nil {
		for _, item := range order.Items {
			if item.SKUID != nil {
				u.invUC.DeductOnhand(ctx, *item.SKUID, eventLoc.ID, item.QtyOrdered)
				u.invUC.CreateLog(ctx, &invDomain.InventoryLog{
					ID:            uuid.New(),
					SKUID:         *item.SKUID,
					LocationID:    eventLoc.ID,
					EventID:       eventID,
					Action:        "ship",
					QtyChange:     -item.QtyOrdered,
					ReferenceID:   &order.ID,
					ReferenceType: "order",
					UserID:        &userID,
				})
			}
		}
	}

	now := time.Now()
	return u.orderUC.UpdateStatus(ctx, order.ID, "shipped", map[string]interface{}{
		"shipped_at": now,
		"shipped_by": userID,
	})
}
