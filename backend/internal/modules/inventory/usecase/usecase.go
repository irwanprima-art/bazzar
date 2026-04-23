package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/irwan/bazzar/internal/modules/inventory/domain"
	"github.com/irwan/bazzar/internal/modules/inventory/repository"
)

type InventoryUsecase struct {
	repo *repository.InventoryRepository
}

func NewInventoryUsecase(repo *repository.InventoryRepository) *InventoryUsecase {
	return &InventoryUsecase{repo: repo}
}

func (u *InventoryUsecase) Adjust(ctx context.Context, req domain.AdjustRequest, userID uuid.UUID) error {
	if err := u.repo.SetStock(ctx, req.SKUID, req.LocationID, req.Qty); err != nil {
		return errors.New("failed to adjust stock")
	}

	log := &domain.InventoryLog{
		ID:            uuid.New(),
		SKUID:         req.SKUID,
		LocationID:    req.LocationID,
		EventID:       req.EventID,
		Action:        "adjust",
		QtyChange:     req.Qty,
		ReferenceType: "manual",
		UserID:        &userID,
		Notes:         req.Notes,
	}
	u.repo.CreateLog(ctx, log)
	return nil
}

func (u *InventoryUsecase) Replenish(ctx context.Context, req domain.ReplenishRequest, eventLocID, storageLocID uuid.UUID, userID uuid.UUID) error {
	// Check storage has enough
	storageInv, err := u.repo.GetBySkuAndLocation(ctx, req.SKUID, storageLocID)
	if err != nil || storageInv.QtyOnhand < req.Qty {
		return errors.New("insufficient stock in storage")
	}

	// Deduct from storage
	if err := u.repo.DeductOnhand(ctx, req.SKUID, storageLocID, req.Qty); err != nil {
		return errors.New("failed to deduct from storage")
	}

	// Add to event
	if err := u.repo.UpsertStock(ctx, req.SKUID, eventLocID, req.Qty); err != nil {
		return errors.New("failed to add to event stock")
	}

	// Log both movements
	outLog := &domain.InventoryLog{
		ID:            uuid.New(),
		SKUID:         req.SKUID,
		LocationID:    storageLocID,
		EventID:       req.EventID,
		Action:        "replenish_out",
		QtyChange:     -req.Qty,
		ReferenceType: "replenish",
		UserID:        &userID,
		Notes:         req.Notes,
	}
	u.repo.CreateLog(ctx, outLog)

	inLog := &domain.InventoryLog{
		ID:            uuid.New(),
		SKUID:         req.SKUID,
		LocationID:    eventLocID,
		EventID:       req.EventID,
		Action:        "replenish_in",
		QtyChange:     req.Qty,
		ReferenceType: "replenish",
		UserID:        &userID,
		Notes:         req.Notes,
	}
	u.repo.CreateLog(ctx, inLog)

	return nil
}

func (u *InventoryUsecase) Transfer(ctx context.Context, req domain.TransferRequest, eventLocID, storageLocID uuid.UUID, userID uuid.UUID) error {
	var fromLocID, toLocID uuid.UUID
	var actionOut, actionIn string

	if req.Direction == "event_to_storage" {
		fromLocID = eventLocID
		toLocID = storageLocID
		actionOut = "transfer_out"
		actionIn = "transfer_in"
	} else {
		// default: storage_to_event (same as replenish)
		fromLocID = storageLocID
		toLocID = eventLocID
		actionOut = "replenish_out"
		actionIn = "replenish_in"
	}

	// Check source has enough
	sourceInv, err := u.repo.GetBySkuAndLocation(ctx, req.SKUID, fromLocID)
	if err != nil || sourceInv.QtyOnhand < req.Qty {
		available := 0
		if err == nil {
			available = sourceInv.QtyOnhand
		}
		return fmt.Errorf("stok tidak cukup (tersedia: %d, dibutuhkan: %d)", available, req.Qty)
	}

	// Deduct from source
	if err := u.repo.DeductOnhand(ctx, req.SKUID, fromLocID, req.Qty); err != nil {
		return errors.New("failed to deduct from source")
	}

	// Add to destination
	if err := u.repo.UpsertStock(ctx, req.SKUID, toLocID, req.Qty); err != nil {
		return errors.New("failed to add to destination")
	}

	// Log both movements
	u.repo.CreateLog(ctx, &domain.InventoryLog{
		ID: uuid.New(), SKUID: req.SKUID, LocationID: fromLocID, EventID: req.EventID,
		Action: actionOut, QtyChange: -req.Qty, ReferenceType: "transfer", UserID: &userID, Notes: req.Notes,
	})
	u.repo.CreateLog(ctx, &domain.InventoryLog{
		ID: uuid.New(), SKUID: req.SKUID, LocationID: toLocID, EventID: req.EventID,
		Action: actionIn, QtyChange: req.Qty, ReferenceType: "transfer", UserID: &userID, Notes: req.Notes,
	})

	return nil
}

func (u *InventoryUsecase) ListByEvent(ctx context.Context, eventID uuid.UUID, locationCode string) ([]domain.Inventory, error) {
	return u.repo.ListByEvent(ctx, eventID, locationCode)
}

func (u *InventoryUsecase) GetReplenishAlerts(ctx context.Context, eventID uuid.UUID) ([]domain.ReplenishAlert, error) {
	return u.repo.GetReplenishAlerts(ctx, eventID)
}

func (u *InventoryUsecase) GetLogs(ctx context.Context, eventID uuid.UUID, skuID *uuid.UUID, limit int) ([]domain.InventoryLog, error) {
	if limit <= 0 {
		limit = 50
	}
	return u.repo.GetLogs(ctx, eventID, skuID, limit)
}

func (u *InventoryUsecase) GetSalesReport(ctx context.Context, eventID uuid.UUID) ([]domain.SalesReport, error) {
	return u.repo.GetSalesReport(ctx, eventID, nil, nil)
}

// GetBySkuAndLocation exposes repo method for other usecases
func (u *InventoryUsecase) GetBySkuAndLocation(ctx context.Context, skuID, locationID uuid.UUID) (*domain.Inventory, error) {
	return u.repo.GetBySkuAndLocation(ctx, skuID, locationID)
}

func (u *InventoryUsecase) AddAllocated(ctx context.Context, skuID, locationID uuid.UUID, qty int) error {
	return u.repo.AddAllocated(ctx, skuID, locationID, qty)
}

func (u *InventoryUsecase) RemoveAllocated(ctx context.Context, skuID, locationID uuid.UUID, qty int) error {
	return u.repo.RemoveAllocated(ctx, skuID, locationID, qty)
}

func (u *InventoryUsecase) DeductOnhand(ctx context.Context, skuID, locationID uuid.UUID, qty int) error {
	return u.repo.DeductOnhand(ctx, skuID, locationID, qty)
}

func (u *InventoryUsecase) CreateLog(ctx context.Context, log *domain.InventoryLog) error {
	return u.repo.CreateLog(ctx, log)
}

func (u *InventoryUsecase) UpsertStock(ctx context.Context, skuID, locationID uuid.UUID, qty int) error {
	return u.repo.UpsertStock(ctx, skuID, locationID, qty)
}
