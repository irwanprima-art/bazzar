package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/irwan/bazzar/internal/modules/inbound/domain"
)

type InboundRepository struct {
	db *pgxpool.Pool
}

func NewInboundRepository(db *pgxpool.Pool) *InboundRepository {
	return &InboundRepository{db: db}
}

func (r *InboundRepository) Create(ctx context.Context, o *domain.InboundOrder) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO inbound_orders (id, event_id, reference_number, status, notes, imported_by)
		VALUES ($1,$2,$3,$4,$5,$6)
	`, o.ID, o.EventID, o.ReferenceNumber, o.Status, o.Notes, o.ImportedBy)
	return err
}

func (r *InboundRepository) CreateItem(ctx context.Context, item *domain.InboundItem) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO inbound_items (id, inbound_order_id, sku_id, qty_expected, qty_received)
		VALUES ($1,$2,$3,$4,$5)
		ON CONFLICT (inbound_order_id, sku_id) DO UPDATE SET
			qty_expected = inbound_items.qty_expected + EXCLUDED.qty_expected,
			updated_at = NOW()
	`, item.ID, item.InboundOrderID, item.SKUID, item.QtyExpected, item.QtyReceived)
	return err
}

func (r *InboundRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.InboundOrder, error) {
	var o domain.InboundOrder
	err := r.db.QueryRow(ctx, `
		SELECT io.id, io.event_id, io.reference_number, io.status,
			   COALESCE(io.notes,''), io.imported_by, io.created_at, io.updated_at,
			   COALESCE(u.full_name,'')
		FROM inbound_orders io
		LEFT JOIN users u ON u.id = io.imported_by
		WHERE io.id = $1
	`, id).Scan(&o.ID, &o.EventID, &o.ReferenceNumber, &o.Status,
		&o.Notes, &o.ImportedBy, &o.CreatedAt, &o.UpdatedAt, &o.ImportedByName)
	if err != nil {
		return nil, err
	}
	return &o, nil
}

func (r *InboundRepository) List(ctx context.Context, eventID uuid.UUID) ([]domain.InboundOrder, error) {
	rows, err := r.db.Query(ctx, `
		SELECT io.id, io.event_id, io.reference_number, io.status,
			   COALESCE(io.notes,''), io.imported_by, io.created_at, io.updated_at,
			   COALESCE(u.full_name,'')
		FROM inbound_orders io
		LEFT JOIN users u ON u.id = io.imported_by
		WHERE io.event_id = $1 ORDER BY io.created_at DESC
	`, eventID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []domain.InboundOrder
	for rows.Next() {
		var o domain.InboundOrder
		rows.Scan(&o.ID, &o.EventID, &o.ReferenceNumber, &o.Status,
			&o.Notes, &o.ImportedBy, &o.CreatedAt, &o.UpdatedAt, &o.ImportedByName)
		orders = append(orders, o)
	}
	return orders, nil
}

func (r *InboundRepository) GetItems(ctx context.Context, inboundID uuid.UUID) ([]domain.InboundItem, error) {
	rows, err := r.db.Query(ctx, `
		SELECT ii.id, ii.inbound_order_id, ii.sku_id, ii.qty_expected, ii.qty_received,
			   (ii.qty_expected - ii.qty_received) as qty_remaining,
			   ii.created_at, s.sku_code, s.name, s.barcode
		FROM inbound_items ii
		JOIN skus s ON s.id = ii.sku_id
		WHERE ii.inbound_order_id = $1 ORDER BY s.sku_code
	`, inboundID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []domain.InboundItem
	for rows.Next() {
		var it domain.InboundItem
		rows.Scan(&it.ID, &it.InboundOrderID, &it.SKUID, &it.QtyExpected, &it.QtyReceived,
			&it.QtyRemaining, &it.CreatedAt, &it.SKUCode, &it.SKUName, &it.Barcode)
		items = append(items, it)
	}
	return items, nil
}

func (r *InboundRepository) UpdateItemReceived(ctx context.Context, itemID uuid.UUID, qtyReceived int) error {
	_, err := r.db.Exec(ctx, `
		UPDATE inbound_items SET qty_received = $1, updated_at = NOW() WHERE id = $2
	`, qtyReceived, itemID)
	return err
}

func (r *InboundRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	_, err := r.db.Exec(ctx, `
		UPDATE inbound_orders SET status = $1, updated_at = NOW() WHERE id = $2
	`, status, id)
	return err
}
