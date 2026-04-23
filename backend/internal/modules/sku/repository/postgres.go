package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/irwan/bazzar/internal/modules/sku/domain"
)

type SKURepository struct {
	db *pgxpool.Pool
}

func NewSKURepository(db *pgxpool.Pool) *SKURepository {
	return &SKURepository{db: db}
}

func (r *SKURepository) Create(ctx context.Context, sku *domain.SKU) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO skus (id, sku_code, barcode, name, description, replenish_limit)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, sku.ID, sku.SKUCode, sku.Barcode, sku.Name, sku.Description, sku.ReplenishLimit)
	return err
}

func (r *SKURepository) Update(ctx context.Context, sku *domain.SKU) error {
	_, err := r.db.Exec(ctx, `
		UPDATE skus SET sku_code=$2, barcode=$3, name=$4, description=$5, 
		replenish_limit=$6, updated_at=NOW()
		WHERE id=$1
	`, sku.ID, sku.SKUCode, sku.Barcode, sku.Name, sku.Description, sku.ReplenishLimit)
	return err
}

func (r *SKURepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx, `DELETE FROM skus WHERE id = $1`, id)
	return err
}

func (r *SKURepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.SKU, error) {
	var sku domain.SKU
	err := r.db.QueryRow(ctx, `
		SELECT id, sku_code, barcode, name, COALESCE(description,''), replenish_limit, created_at, updated_at
		FROM skus WHERE id = $1
	`, id).Scan(&sku.ID, &sku.SKUCode, &sku.Barcode, &sku.Name, &sku.Description,
		&sku.ReplenishLimit, &sku.CreatedAt, &sku.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &sku, nil
}

func (r *SKURepository) GetBySKUCode(ctx context.Context, code string) (*domain.SKU, error) {
	var sku domain.SKU
	err := r.db.QueryRow(ctx, `
		SELECT id, sku_code, barcode, name, COALESCE(description,''), replenish_limit, created_at, updated_at
		FROM skus WHERE sku_code = $1
	`, code).Scan(&sku.ID, &sku.SKUCode, &sku.Barcode, &sku.Name, &sku.Description,
		&sku.ReplenishLimit, &sku.CreatedAt, &sku.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &sku, nil
}

func (r *SKURepository) GetByBarcode(ctx context.Context, barcode string) (*domain.SKU, error) {
	var sku domain.SKU
	err := r.db.QueryRow(ctx, `
		SELECT id, sku_code, barcode, name, COALESCE(description,''), replenish_limit, created_at, updated_at
		FROM skus WHERE barcode = $1
	`, barcode).Scan(&sku.ID, &sku.SKUCode, &sku.Barcode, &sku.Name, &sku.Description,
		&sku.ReplenishLimit, &sku.CreatedAt, &sku.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &sku, nil
}

// GetByBarcodeOrSKUCode tries barcode first, then falls back to sku_code
func (r *SKURepository) GetByBarcodeOrSKUCode(ctx context.Context, input string) (*domain.SKU, error) {
	var sku domain.SKU
	err := r.db.QueryRow(ctx, `
		SELECT id, sku_code, barcode, name, COALESCE(description,''), replenish_limit, created_at, updated_at
		FROM skus WHERE barcode = $1 OR sku_code = $1
		LIMIT 1
	`, input).Scan(&sku.ID, &sku.SKUCode, &sku.Barcode, &sku.Name, &sku.Description,
		&sku.ReplenishLimit, &sku.CreatedAt, &sku.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &sku, nil
}

func (r *SKURepository) List(ctx context.Context, search string, page, pageSize int) ([]domain.SKU, int64, error) {
	offset := (page - 1) * pageSize

	var total int64
	if search != "" {
		err := r.db.QueryRow(ctx, `
			SELECT COUNT(*) FROM skus 
			WHERE sku_code ILIKE $1 OR barcode ILIKE $1 OR name ILIKE $1
		`, "%"+search+"%").Scan(&total)
		if err != nil {
			return nil, 0, err
		}
	} else {
		err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM skus`).Scan(&total)
		if err != nil {
			return nil, 0, err
		}
	}

	var query string
	var args []interface{}

	if search != "" {
		query = `SELECT id, sku_code, barcode, name, COALESCE(description,''), replenish_limit, created_at, updated_at
			FROM skus WHERE sku_code ILIKE $1 OR barcode ILIKE $1 OR name ILIKE $1
			ORDER BY created_at DESC LIMIT $2 OFFSET $3`
		args = []interface{}{"%" + search + "%", pageSize, offset}
	} else {
		query = `SELECT id, sku_code, barcode, name, COALESCE(description,''), replenish_limit, created_at, updated_at
			FROM skus ORDER BY created_at DESC LIMIT $1 OFFSET $2`
		args = []interface{}{pageSize, offset}
	}

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var skus []domain.SKU
	for rows.Next() {
		var s domain.SKU
		if err := rows.Scan(&s.ID, &s.SKUCode, &s.Barcode, &s.Name, &s.Description,
			&s.ReplenishLimit, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, 0, err
		}
		skus = append(skus, s)
	}
	return skus, total, nil
}

func (r *SKURepository) UpsertBySKUCode(ctx context.Context, sku *domain.SKU) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO skus (id, sku_code, barcode, name, description, replenish_limit)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (sku_code) DO UPDATE SET
			name = EXCLUDED.name,
			updated_at = NOW()
	`, sku.ID, sku.SKUCode, sku.Barcode, sku.Name, sku.Description, sku.ReplenishLimit)
	return err
}

func itoa(i int) string {
	return fmt.Sprintf("%d", i)
}
