package domain

import (
	"time"

	"github.com/google/uuid"
)

type Inventory struct {
	ID           uuid.UUID `json:"id"`
	SKUID        uuid.UUID `json:"sku_id"`
	LocationID   uuid.UUID `json:"location_id"`
	QtyOnhand    int       `json:"qty_onhand"`
	QtyAllocated int       `json:"qty_allocated"`
	Available    int       `json:"available"` // Computed: onhand - allocated
	UpdatedAt    time.Time `json:"updated_at"`

	// Joined fields
	SKUCode      string  `json:"sku_code,omitempty"`
	SKUName      string  `json:"sku_name,omitempty"`
	Barcode      *string `json:"barcode,omitempty"`
	LocationCode string  `json:"location_code,omitempty"`
	LocationName string  `json:"location_name,omitempty"`
}

type InventoryLog struct {
	ID            uuid.UUID `json:"id"`
	SKUID         uuid.UUID `json:"sku_id"`
	LocationID    uuid.UUID `json:"location_id"`
	EventID       uuid.UUID `json:"event_id"`
	Action        string    `json:"action"`
	QtyChange     int       `json:"qty_change"`
	ReferenceID   *uuid.UUID `json:"reference_id"`
	ReferenceType string    `json:"reference_type"`
	UserID        *uuid.UUID `json:"user_id"`
	Notes         string    `json:"notes"`
	CreatedAt     time.Time `json:"created_at"`

	// Joined
	SKUCode  string `json:"sku_code,omitempty"`
	SKUName  string `json:"sku_name,omitempty"`
	Username string `json:"username,omitempty"`
}

type AdjustRequest struct {
	SKUID      uuid.UUID `json:"sku_id"`
	LocationID uuid.UUID `json:"location_id"`
	EventID    uuid.UUID `json:"event_id"`
	Qty        int       `json:"qty"`
	Notes      string    `json:"notes"`
}

type ReplenishRequest struct {
	EventID uuid.UUID `json:"event_id"`
	SKUID   uuid.UUID `json:"sku_id"`
	Qty     int       `json:"qty"`
	Notes   string    `json:"notes"`
}

type TransferRequest struct {
	EventID   uuid.UUID `json:"event_id"`
	SKUID     uuid.UUID `json:"sku_id"`
	Qty       int       `json:"qty"`
	Direction string    `json:"direction"` // "storage_to_event" or "event_to_storage"
	Notes     string    `json:"notes"`
}

type ReplenishAlert struct {
	SKUID            uuid.UUID `json:"sku_id"`
	SKUCode          string    `json:"sku_code"`
	SKUName          string    `json:"sku_name"`
	ReplenishLimit   int       `json:"replenish_limit"`
	EventAvailable   int       `json:"event_available"`
	StorageOnhand    int       `json:"storage_onhand"`
	NeedsReplenish   bool      `json:"needs_replenish"`
	StorageDepleted  bool      `json:"storage_depleted"`
}

type SalesReport struct {
	SKUID     uuid.UUID `json:"sku_id"`
	SKUCode   string    `json:"sku_code"`
	SKUName   string    `json:"sku_name"`
	QtySold   int       `json:"qty_sold"`
	QtyOnhand int       `json:"qty_onhand"`
}
