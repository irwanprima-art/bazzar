package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/irwan/bazzar/internal/modules/inventory/domain"
)

type InventoryRepository struct {
	db *pgxpool.Pool
}

func NewInventoryRepository(db *pgxpool.Pool) *InventoryRepository {
	return &InventoryRepository{db: db}
}

func (r *InventoryRepository) GetBySkuAndLocation(ctx context.Context, skuID, locationID uuid.UUID) (*domain.Inventory, error) {
	var inv domain.Inventory
	err := r.db.QueryRow(ctx, `
		SELECT i.id, i.sku_id, i.location_id, i.qty_onhand, i.qty_allocated,
			   (i.qty_onhand - i.qty_allocated) as available, i.updated_at,
			   s.sku_code, s.name, s.barcode, l.code, l.name
		FROM inventory i
		JOIN skus s ON s.id = i.sku_id
		JOIN locations l ON l.id = i.location_id
		WHERE i.sku_id = $1 AND i.location_id = $2
	`, skuID, locationID).Scan(
		&inv.ID, &inv.SKUID, &inv.LocationID, &inv.QtyOnhand, &inv.QtyAllocated,
		&inv.Available, &inv.UpdatedAt,
		&inv.SKUCode, &inv.SKUName, &inv.Barcode, &inv.LocationCode, &inv.LocationName,
	)
	if err != nil {
		return nil, err
	}
	return &inv, nil
}

func (r *InventoryRepository) UpsertStock(ctx context.Context, skuID, locationID uuid.UUID, qtyOnhand int) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO inventory (id, sku_id, location_id, qty_onhand)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (sku_id, location_id) DO UPDATE SET
			qty_onhand = inventory.qty_onhand + EXCLUDED.qty_onhand,
			updated_at = NOW()
	`, uuid.New(), skuID, locationID, qtyOnhand)
	return err
}

func (r *InventoryRepository) SetStock(ctx context.Context, skuID, locationID uuid.UUID, qtyOnhand int) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO inventory (id, sku_id, location_id, qty_onhand)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (sku_id, location_id) DO UPDATE SET
			qty_onhand = $4,
			updated_at = NOW()
	`, uuid.New(), skuID, locationID, qtyOnhand)
	return err
}

func (r *InventoryRepository) AddAllocated(ctx context.Context, skuID, locationID uuid.UUID, qty int) error {
	_, err := r.db.Exec(ctx, `
		UPDATE inventory SET qty_allocated = qty_allocated + $1, updated_at = NOW()
		WHERE sku_id = $2 AND location_id = $3
	`, qty, skuID, locationID)
	return err
}

func (r *InventoryRepository) RemoveAllocated(ctx context.Context, skuID, locationID uuid.UUID, qty int) error {
	_, err := r.db.Exec(ctx, `
		UPDATE inventory SET qty_allocated = GREATEST(qty_allocated - $1, 0), updated_at = NOW()
		WHERE sku_id = $2 AND location_id = $3
	`, qty, skuID, locationID)
	return err
}

func (r *InventoryRepository) DeductOnhand(ctx context.Context, skuID, locationID uuid.UUID, qty int) error {
	_, err := r.db.Exec(ctx, `
		UPDATE inventory SET 
			qty_onhand = qty_onhand - $1,
			qty_allocated = GREATEST(qty_allocated - $1, 0),
			updated_at = NOW()
		WHERE sku_id = $2 AND location_id = $3
	`, qty, skuID, locationID)
	return err
}

func (r *InventoryRepository) ListByEvent(ctx context.Context, eventID uuid.UUID, locationCode string) ([]domain.Inventory, error) {
	query := `
		SELECT i.id, i.sku_id, i.location_id, i.qty_onhand, i.qty_allocated,
			   (i.qty_onhand - i.qty_allocated) as available, i.updated_at,
			   s.sku_code, s.name, s.barcode, l.code, l.name
		FROM inventory i
		JOIN skus s ON s.id = i.sku_id
		JOIN locations l ON l.id = i.location_id
		WHERE l.event_id = $1`

	args := []interface{}{eventID}
	if locationCode != "" {
		query += ` AND l.code = $2`
		args = append(args, locationCode)
	}
	query += ` ORDER BY s.sku_code`

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []domain.Inventory
	for rows.Next() {
		var inv domain.Inventory
		if err := rows.Scan(
			&inv.ID, &inv.SKUID, &inv.LocationID, &inv.QtyOnhand, &inv.QtyAllocated,
			&inv.Available, &inv.UpdatedAt,
			&inv.SKUCode, &inv.SKUName, &inv.Barcode, &inv.LocationCode, &inv.LocationName,
		); err != nil {
			return nil, err
		}
		items = append(items, inv)
	}
	return items, nil
}

func (r *InventoryRepository) GetReplenishAlerts(ctx context.Context, eventID uuid.UUID) ([]domain.ReplenishAlert, error) {
	rows, err := r.db.Query(ctx, `
		SELECT s.id, s.sku_code, s.name, s.replenish_limit,
			   COALESCE(ie.qty_onhand - ie.qty_allocated, 0) as event_available,
			   COALESCE(is2.qty_onhand, 0) as storage_onhand
		FROM skus s
		LEFT JOIN locations le ON le.event_id = $1 AND le.code = 'EVENT'
		LEFT JOIN locations ls ON ls.event_id = $1 AND ls.code = 'STORAGE'
		LEFT JOIN inventory ie ON ie.sku_id = s.id AND ie.location_id = le.id
		LEFT JOIN inventory is2 ON is2.sku_id = s.id AND is2.location_id = ls.id
		WHERE COALESCE(ie.qty_onhand - ie.qty_allocated, 0) <= s.replenish_limit
		  AND (COALESCE(ie.qty_onhand, 0) > 0 OR COALESCE(is2.qty_onhand, 0) > 0)
		ORDER BY (COALESCE(ie.qty_onhand - ie.qty_allocated, 0)) ASC
	`, eventID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var alerts []domain.ReplenishAlert
	for rows.Next() {
		var a domain.ReplenishAlert
		if err := rows.Scan(&a.SKUID, &a.SKUCode, &a.SKUName, &a.ReplenishLimit,
			&a.EventAvailable, &a.StorageOnhand); err != nil {
			return nil, err
		}
		a.NeedsReplenish = a.EventAvailable <= a.ReplenishLimit
		a.StorageDepleted = a.StorageOnhand == 0
		alerts = append(alerts, a)
	}
	return alerts, nil
}

func (r *InventoryRepository) CreateLog(ctx context.Context, log *domain.InventoryLog) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO inventory_logs (id, sku_id, location_id, event_id, action, qty_change,
			reference_id, reference_type, user_id, notes)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`, log.ID, log.SKUID, log.LocationID, log.EventID, log.Action, log.QtyChange,
		log.ReferenceID, log.ReferenceType, log.UserID, log.Notes)
	return err
}

func (r *InventoryRepository) GetLogs(ctx context.Context, eventID uuid.UUID, skuID *uuid.UUID, search string, limit int) ([]domain.InventoryLog, error) {
	query := `
		SELECT il.id, il.sku_id, il.location_id, il.event_id, il.action, il.qty_change,
			   il.reference_id, COALESCE(il.reference_type,''), il.user_id, COALESCE(il.notes,''), il.created_at,
			   s.sku_code, COALESCE(s.name,''), COALESCE(u.username,'system'),
			   COALESCE(l.code,''),
			   COALESCE(
			     CASE
			       WHEN il.reference_type = 'order' THEN (SELECT order_number FROM orders WHERE id = il.reference_id LIMIT 1)
			       WHEN il.reference_type = 'inbound' THEN (SELECT reference_number FROM inbound_orders WHERE id = il.reference_id LIMIT 1)
			       ELSE ''
			     END, ''
			   ) as ref_number
		FROM inventory_logs il
		JOIN skus s ON s.id = il.sku_id
		LEFT JOIN users u ON u.id = il.user_id
		LEFT JOIN locations l ON l.id = il.location_id
		WHERE il.event_id = $1`

	args := []interface{}{eventID}
	argIdx := 2
	if skuID != nil {
		query += fmt.Sprintf(` AND il.sku_id = $%d`, argIdx)
		args = append(args, *skuID)
		argIdx++
	}
	if search != "" {
		query += fmt.Sprintf(` AND (s.sku_code ILIKE $%d OR s.name ILIKE $%d)`, argIdx, argIdx)
		args = append(args, "%"+search+"%")
		argIdx++
	}
	query += fmt.Sprintf(` ORDER BY il.created_at DESC LIMIT $%d`, argIdx)
	args = append(args, limit)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []domain.InventoryLog
	for rows.Next() {
		var l domain.InventoryLog
		if err := rows.Scan(
			&l.ID, &l.SKUID, &l.LocationID, &l.EventID, &l.Action, &l.QtyChange,
			&l.ReferenceID, &l.ReferenceType, &l.UserID, &l.Notes, &l.CreatedAt,
			&l.SKUCode, &l.SKUName, &l.Username,
			&l.LocationCode, &l.ReferenceNumber,
		); err != nil {
			return nil, err
		}
		logs = append(logs, l)
	}
	return logs, nil
}

func (r *InventoryRepository) GetSalesReport(ctx context.Context, eventID uuid.UUID, from, to *time.Time) ([]domain.SalesReport, error) {
	query := `
		SELECT s.id, s.sku_code, s.name,
			   COALESCE(SUM(CASE WHEN il.action = 'ship' THEN ABS(il.qty_change) ELSE 0 END), 0) as qty_sold,
			   COALESCE(ie.qty_onhand, 0) + COALESCE(is2.qty_onhand, 0) as qty_onhand
		FROM skus s
		LEFT JOIN locations le ON le.event_id = $1 AND le.code = 'EVENT'
		LEFT JOIN locations ls ON ls.event_id = $1 AND ls.code = 'STORAGE'
		LEFT JOIN inventory ie ON ie.sku_id = s.id AND ie.location_id = le.id
		LEFT JOIN inventory is2 ON is2.sku_id = s.id AND is2.location_id = ls.id
		LEFT JOIN inventory_logs il ON il.sku_id = s.id AND il.event_id = $1 AND il.action = 'ship'`

	args := []interface{}{eventID}
	argIdx := 2

	if from != nil {
		query += fmt.Sprintf(` AND il.created_at >= $%d`, argIdx)
		args = append(args, *from)
		argIdx++
	}
	if to != nil {
		query += fmt.Sprintf(` AND il.created_at <= $%d`, argIdx)
		args = append(args, *to)
		argIdx++
	}

	query += ` GROUP BY s.id, s.sku_code, s.name, ie.qty_onhand, is2.qty_onhand
			   HAVING COALESCE(SUM(CASE WHEN il.action = 'ship' THEN ABS(il.qty_change) ELSE 0 END), 0) > 0
			      OR COALESCE(ie.qty_onhand, 0) + COALESCE(is2.qty_onhand, 0) > 0
			   ORDER BY qty_sold DESC`

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reports []domain.SalesReport
	for rows.Next() {
		var rpt domain.SalesReport
		if err := rows.Scan(&rpt.SKUID, &rpt.SKUCode, &rpt.SKUName, &rpt.QtySold, &rpt.QtyOnhand); err != nil {
			return nil, err
		}
		reports = append(reports, rpt)
	}
	return reports, nil
}
