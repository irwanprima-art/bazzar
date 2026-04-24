package domain

import (
	"time"

	"github.com/google/uuid"
)

type Order struct {
	ID               uuid.UUID  `json:"id"`
	EventID          uuid.UUID  `json:"event_id"`
	OrderNumber      string     `json:"order_number"`
	PlatformStatus   string     `json:"platform_status"`
	Status           string     `json:"status"`
	BuyerName        string     `json:"buyer_name"`
	BuyerUsername     string     `json:"buyer_username"`
	ShippingOption   string     `json:"shipping_option"`
	TrackingNumber   string     `json:"tracking_number"`
	ProductName      string     `json:"product_name"`
	VariationName    string     `json:"variation_name"`
	Notes            string     `json:"notes"`
	TotalPayment     float64    `json:"total_payment"`
	AssignedPickerID *uuid.UUID `json:"assigned_picker_id"`
	ImportedBy       *uuid.UUID `json:"imported_by"`
	PrintedBy        *uuid.UUID `json:"printed_by"`
	PickedBy         *uuid.UUID `json:"picked_by"`
	ShippedBy        *uuid.UUID `json:"shipped_by"`
	ImportedAt       *time.Time `json:"imported_at"`
	AllocatedAt      *time.Time `json:"allocated_at"`
	PrintedAt        *time.Time `json:"printed_at"`
	PickingStartedAt *time.Time `json:"picking_started_at"`
	PickedAt         *time.Time `json:"picked_at"`
	ShippedAt        *time.Time `json:"shipped_at"`
	CreatedAt        time.Time  `json:"created_at"`

	// Joined
	PickerName string      `json:"picker_name,omitempty"`
	Items      []OrderItem `json:"items,omitempty"`
}

type OrderItem struct {
	ID            uuid.UUID  `json:"id"`
	OrderID       uuid.UUID  `json:"order_id"`
	SKUID         *uuid.UUID `json:"sku_id"`
	SKUCode       string     `json:"sku_code"`
	ProductName   string     `json:"product_name"`
	VariationName string     `json:"variation_name"`
	QtyOrdered    int        `json:"qty_ordered"`
	QtyPicked     int        `json:"qty_picked"`
	CreatedAt     time.Time  `json:"created_at"`

	// Joined
	SKUName string  `json:"sku_name,omitempty"`
	Barcode *string `json:"barcode,omitempty"`
}

type ImportResult struct {
	TotalRows      int      `json:"total_rows"`
	Imported       int      `json:"imported"`
	Updated        int      `json:"updated"`
	Skipped        int      `json:"skipped"`
	Duplicates     int      `json:"duplicates"`
	Errors         int      `json:"errors"`
	ErrorDetails   []string `json:"error_details,omitempty"`
	SkippedDetails []string `json:"skipped_details,omitempty"`
}

type OrderFilter struct {
	EventID uuid.UUID
	Status  string
	Search  string
	Page    int
	PageSize int
}

// Shopee statuses that should be marked as 'issue' and not processed
var IssueStatuses = map[string]bool{
	"Pembatalan diajukan": true,
	"Batal":               true,
	"Belum Bayar":         true,
}

type AllocateResult struct {
	TotalOrders int               `json:"total_orders"`
	Allocated   int               `json:"allocated"`
	Failed      []AllocateFailure `json:"failed,omitempty"`
}

type AllocateFailure struct {
	OrderNumber string `json:"order_number"`
	SKUCode     string `json:"sku_code"`
	Reason      string `json:"reason"`
}
