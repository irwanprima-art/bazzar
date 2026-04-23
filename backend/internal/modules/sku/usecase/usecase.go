package usecase

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/irwan/bazzar/internal/modules/sku/domain"
	"github.com/irwan/bazzar/internal/modules/sku/repository"
)

type SKUUsecase struct {
	repo *repository.SKURepository
}

func NewSKUUsecase(repo *repository.SKURepository) *SKUUsecase {
	return &SKUUsecase{repo: repo}
}

func (u *SKUUsecase) Create(ctx context.Context, req domain.CreateSKURequest) (*domain.SKU, error) {
	if req.SKUCode == "" || req.Name == "" {
		return nil, errors.New("SKU code and name are required")
	}

	var barcode *string
	if req.Barcode != "" {
		barcode = &req.Barcode
	}

	limit := req.ReplenishLimit
	if limit <= 0 {
		limit = 5
	}

	sku := &domain.SKU{
		ID:             uuid.New(),
		SKUCode:        req.SKUCode,
		Barcode:        barcode,
		Name:           req.Name,
		Description:    req.Description,
		ReplenishLimit: limit,
	}

	if err := u.repo.Create(ctx, sku); err != nil {
		return nil, errors.New("SKU code or barcode already exists")
	}

	return sku, nil
}

func (u *SKUUsecase) Update(ctx context.Context, id uuid.UUID, req domain.UpdateSKURequest) (*domain.SKU, error) {
	existing, err := u.repo.GetByID(ctx, id)
	if err != nil {
		return nil, errors.New("SKU not found")
	}

	if req.SKUCode != "" {
		existing.SKUCode = req.SKUCode
	}
	if req.Name != "" {
		existing.Name = req.Name
	}
	existing.Description = req.Description

	var barcode *string
	if req.Barcode != "" {
		barcode = &req.Barcode
	}
	existing.Barcode = barcode

	if req.ReplenishLimit > 0 {
		existing.ReplenishLimit = req.ReplenishLimit
	}

	if err := u.repo.Update(ctx, existing); err != nil {
		return nil, errors.New("failed to update SKU")
	}

	return existing, nil
}

func (u *SKUUsecase) Delete(ctx context.Context, id uuid.UUID) error {
	return u.repo.Delete(ctx, id)
}

func (u *SKUUsecase) GetByID(ctx context.Context, id uuid.UUID) (*domain.SKU, error) {
	return u.repo.GetByID(ctx, id)
}

func (u *SKUUsecase) GetByBarcode(ctx context.Context, barcode string) (*domain.SKU, error) {
	return u.repo.GetByBarcode(ctx, barcode)
}

func (u *SKUUsecase) GetBySKUCode(ctx context.Context, code string) (*domain.SKU, error) {
	return u.repo.GetBySKUCode(ctx, code)
}

func (u *SKUUsecase) List(ctx context.Context, search string, page, pageSize int) ([]domain.SKU, int64, error) {
	return u.repo.List(ctx, search, page, pageSize)
}
