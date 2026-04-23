package domain

import "github.com/google/uuid"

type StartPickingRequest struct {
	OrderID uuid.UUID `json:"order_id"`
}

type ScanRequest struct {
	Barcode string `json:"barcode"`
	Qty     int    `json:"qty"`
}

type ScanResult struct {
	ItemID      uuid.UUID `json:"item_id"`
	SKUCode     string    `json:"sku_code"`
	SKUName     string    `json:"sku_name"`
	QtyOrdered  int       `json:"qty_ordered"`
	QtyPicked   int       `json:"qty_picked"`
	QtyRemain   int       `json:"qty_remaining"`
	IsComplete  bool      `json:"is_complete"`
	Message     string    `json:"message"`
}
