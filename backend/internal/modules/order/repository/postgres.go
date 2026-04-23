package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/irwan/bazzar/internal/modules/order/domain"
)

type OrderRepository struct {
	db *pgxpool.Pool
}

func NewOrderRepository(db *pgxpool.Pool) *OrderRepository {
	return &OrderRepository{db: db}
}

func (r *OrderRepository) UpsertOrder(ctx context.Context, order *domain.Order) (bool, error) {
	tag, err := r.db.Exec(ctx, `
		INSERT INTO orders (id, event_id, order_number, platform_status, status,
			buyer_name, buyer_username, shipping_option, tracking_number,
			product_name, variation_name, notes, total_payment, imported_by, imported_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,NOW())
		ON CONFLICT (event_id, order_number) DO NOTHING
	`, order.ID, order.EventID, order.OrderNumber, order.PlatformStatus, order.Status,
		order.BuyerName, order.BuyerUsername, order.ShippingOption, order.TrackingNumber,
		order.ProductName, order.VariationName, order.Notes, order.TotalPayment, order.ImportedBy)
	if err != nil {
		return false, err
	}
	return tag.RowsAffected() > 0, nil
}

func (r *OrderRepository) UpsertOrderItem(ctx context.Context, item *domain.OrderItem) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO order_items (id, order_id, sku_id, sku_code, product_name, variation_name, qty_ordered)
		VALUES ($1,$2,$3,$4,$5,$6,$7)
		ON CONFLICT DO NOTHING
	`, item.ID, item.OrderID, item.SKUID, item.SKUCode, item.ProductName, item.VariationName, item.QtyOrdered)
	return err
}

func (r *OrderRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Order, error) {
	var o domain.Order
	err := r.db.QueryRow(ctx, `
		SELECT o.id, o.event_id, o.order_number, COALESCE(o.platform_status,''), o.status,
			COALESCE(o.buyer_name,''), COALESCE(o.buyer_username,''),
			COALESCE(o.shipping_option,''), COALESCE(o.tracking_number,''),
			COALESCE(o.product_name,''), COALESCE(o.variation_name,''),
			COALESCE(o.notes,''), COALESCE(o.total_payment,0),
			o.assigned_picker_id, o.imported_by, o.printed_by, o.picked_by, o.shipped_by,
			o.imported_at, o.allocated_at, o.printed_at, o.picking_started_at, o.picked_at, o.shipped_at,
			o.created_at, COALESCE(u.full_name,'')
		FROM orders o LEFT JOIN users u ON u.id = o.assigned_picker_id
		WHERE o.id = $1
	`, id).Scan(
		&o.ID, &o.EventID, &o.OrderNumber, &o.PlatformStatus, &o.Status,
		&o.BuyerName, &o.BuyerUsername, &o.ShippingOption, &o.TrackingNumber,
		&o.ProductName, &o.VariationName, &o.Notes, &o.TotalPayment,
		&o.AssignedPickerID, &o.ImportedBy, &o.PrintedBy, &o.PickedBy, &o.ShippedBy,
		&o.ImportedAt, &o.AllocatedAt, &o.PrintedAt, &o.PickingStartedAt, &o.PickedAt, &o.ShippedAt,
		&o.CreatedAt, &o.PickerName,
	)
	if err != nil {
		return nil, err
	}
	return &o, nil
}

func (r *OrderRepository) GetByOrderNumber(ctx context.Context, eventID uuid.UUID, orderNum string) (*domain.Order, error) {
	var o domain.Order
	err := r.db.QueryRow(ctx, `
		SELECT o.id, o.event_id, o.order_number, COALESCE(o.platform_status,''), o.status,
			COALESCE(o.buyer_name,''), COALESCE(o.buyer_username,''),
			COALESCE(o.shipping_option,''), COALESCE(o.tracking_number,''),
			COALESCE(o.product_name,''), COALESCE(o.variation_name,''),
			COALESCE(o.notes,''), COALESCE(o.total_payment,0),
			o.assigned_picker_id, o.imported_by, o.printed_by, o.picked_by, o.shipped_by,
			o.imported_at, o.allocated_at, o.printed_at, o.picking_started_at, o.picked_at, o.shipped_at,
			o.created_at, COALESCE(u.full_name,'')
		FROM orders o LEFT JOIN users u ON u.id = o.assigned_picker_id
		WHERE o.event_id = $1 AND o.order_number = $2
	`, eventID, orderNum).Scan(
		&o.ID, &o.EventID, &o.OrderNumber, &o.PlatformStatus, &o.Status,
		&o.BuyerName, &o.BuyerUsername, &o.ShippingOption, &o.TrackingNumber,
		&o.ProductName, &o.VariationName, &o.Notes, &o.TotalPayment,
		&o.AssignedPickerID, &o.ImportedBy, &o.PrintedBy, &o.PickedBy, &o.ShippedBy,
		&o.ImportedAt, &o.AllocatedAt, &o.PrintedAt, &o.PickingStartedAt, &o.PickedAt, &o.ShippedAt,
		&o.CreatedAt, &o.PickerName,
	)
	if err != nil {
		return nil, err
	}
	return &o, nil
}

func (r *OrderRepository) List(ctx context.Context, f domain.OrderFilter) ([]domain.Order, int64, error) {
	offset := (f.Page - 1) * f.PageSize
	where := `WHERE o.event_id = $1`
	args := []interface{}{f.EventID}
	idx := 2

	if f.Status != "" {
		where += fmt.Sprintf(` AND o.status = $%d`, idx)
		args = append(args, f.Status)
		idx++
	}
	if f.Search != "" {
		where += fmt.Sprintf(` AND (o.order_number ILIKE $%d OR o.buyer_name ILIKE $%d)`, idx, idx)
		args = append(args, "%"+f.Search+"%")
		idx++
	}

	var total int64
	r.db.QueryRow(ctx, `SELECT COUNT(*) FROM orders o `+where, args...).Scan(&total)

	q := fmt.Sprintf(`
		SELECT o.id, o.event_id, o.order_number, COALESCE(o.platform_status,''), o.status,
			COALESCE(o.buyer_name,''), COALESCE(o.buyer_username,''),
			COALESCE(o.shipping_option,''), COALESCE(o.tracking_number,''),
			COALESCE(o.product_name,''), COALESCE(o.variation_name,''),
			COALESCE(o.notes,''), COALESCE(o.total_payment,0),
			o.assigned_picker_id, o.imported_by, o.printed_by, o.picked_by, o.shipped_by,
			o.imported_at, o.allocated_at, o.printed_at, o.picking_started_at, o.picked_at, o.shipped_at,
			o.created_at, COALESCE(u.full_name,'')
		FROM orders o LEFT JOIN users u ON u.id = o.assigned_picker_id
		%s ORDER BY o.created_at DESC LIMIT $%d OFFSET $%d
	`, where, idx, idx+1)
	args = append(args, f.PageSize, offset)

	rows, err := r.db.Query(ctx, q, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var orders []domain.Order
	for rows.Next() {
		var o domain.Order
		rows.Scan(
			&o.ID, &o.EventID, &o.OrderNumber, &o.PlatformStatus, &o.Status,
			&o.BuyerName, &o.BuyerUsername, &o.ShippingOption, &o.TrackingNumber,
			&o.ProductName, &o.VariationName, &o.Notes, &o.TotalPayment,
			&o.AssignedPickerID, &o.ImportedBy, &o.PrintedBy, &o.PickedBy, &o.ShippedBy,
			&o.ImportedAt, &o.AllocatedAt, &o.PrintedAt, &o.PickingStartedAt, &o.PickedAt, &o.ShippedAt,
			&o.CreatedAt, &o.PickerName,
		)
		orders = append(orders, o)
	}
	return orders, total, nil
}

func (r *OrderRepository) GetOrderItems(ctx context.Context, orderID uuid.UUID) ([]domain.OrderItem, error) {
	rows, err := r.db.Query(ctx, `
		SELECT oi.id, oi.order_id, oi.sku_id, oi.sku_code,
			COALESCE(oi.product_name,''), COALESCE(oi.variation_name,''),
			oi.qty_ordered, oi.qty_picked, oi.created_at,
			COALESCE(s.name,'Unknown'), s.barcode
		FROM order_items oi LEFT JOIN skus s ON s.id = oi.sku_id
		WHERE oi.order_id = $1 ORDER BY oi.created_at
	`, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []domain.OrderItem
	for rows.Next() {
		var it domain.OrderItem
		rows.Scan(&it.ID, &it.OrderID, &it.SKUID, &it.SKUCode,
			&it.ProductName, &it.VariationName,
			&it.QtyOrdered, &it.QtyPicked, &it.CreatedAt,
			&it.SKUName, &it.Barcode)
		items = append(items, it)
	}
	return items, nil
}

func (r *OrderRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status string, updates map[string]interface{}) error {
	q := `UPDATE orders SET status = $1, updated_at = NOW()`
	args := []interface{}{status}
	idx := 2
	for col, val := range updates {
		q += fmt.Sprintf(`, %s = $%d`, col, idx)
		args = append(args, val)
		idx++
	}
	q += fmt.Sprintf(` WHERE id = $%d`, idx)
	args = append(args, id)
	_, err := r.db.Exec(ctx, q, args...)
	return err
}

func (r *OrderRepository) UpdateItemPicked(ctx context.Context, itemID uuid.UUID, qtyPicked int) error {
	_, err := r.db.Exec(ctx, `UPDATE order_items SET qty_picked = $1 WHERE id = $2`, qtyPicked, itemID)
	return err
}

func (r *OrderRepository) GetStatusCounts(ctx context.Context, eventID uuid.UUID) (map[string]int, error) {
	rows, err := r.db.Query(ctx, `SELECT status, COUNT(*) FROM orders WHERE event_id=$1 GROUP BY status`, eventID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	counts := make(map[string]int)
	for rows.Next() {
		var s string
		var c int
		rows.Scan(&s, &c)
		counts[s] = c
	}
	return counts, nil
}

func (r *OrderRepository) BatchUpdateStatus(ctx context.Context, ids []uuid.UUID, status string, updates map[string]interface{}) error {
	batch := &pgx.Batch{}
	for _, id := range ids {
		q := `UPDATE orders SET status = $1, updated_at = NOW()`
		args := []interface{}{status}
		idx := 2
		for col, val := range updates {
			q += fmt.Sprintf(`, %s = $%d`, col, idx)
			args = append(args, val)
			idx++
		}
		q += fmt.Sprintf(` WHERE id = $%d`, idx)
		args = append(args, id)
		batch.Queue(q, args...)
	}
	br := r.db.SendBatch(ctx, batch)
	defer br.Close()
	for range ids {
		if _, err := br.Exec(); err != nil {
			return err
		}
	}
	return nil
}
