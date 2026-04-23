package domain

import (
	"time"

	"github.com/google/uuid"
)

type InboundOrder struct {
	ID              uuid.UUID     `json:"id"`
	EventID         uuid.UUID     `json:"event_id"`
	ReferenceNumber string        `json:"reference_number"`
	Status          string        `json:"status"`
	Notes           string        `json:"notes"`
	ImportedBy      *uuid.UUID    `json:"imported_by"`
	CreatedAt       time.Time     `json:"created_at"`
	UpdatedAt       time.Time     `json:"updated_at"`
	Items           []InboundItem `json:"items,omitempty"`
	ImportedByName  string        `json:"imported_by_name,omitempty"`
}

type InboundItem struct {
	ID              uuid.UUID `json:"id"`
	InboundOrderID  uuid.UUID `json:"inbound_order_id"`
	SKUID           uuid.UUID `json:"sku_id"`
	QtyExpected     int       `json:"qty_expected"`
	QtyReceived     int       `json:"qty_received"`
	QtyRemaining    int       `json:"qty_remaining"`
	CreatedAt       time.Time `json:"created_at"`
	SKUCode         string    `json:"sku_code,omitempty"`
	SKUName         string    `json:"sku_name,omitempty"`
	Barcode         *string   `json:"barcode,omitempty"`
}

type InboundScanRequest struct {
	Barcode string `json:"barcode"`
	Qty     int    `json:"qty"`
}

type InboundScanResult struct {
	ItemID       uuid.UUID `json:"item_id"`
	SKUCode      string    `json:"sku_code"`
	SKUName      string    `json:"sku_name"`
	QtyExpected  int       `json:"qty_expected"`
	QtyReceived  int       `json:"qty_received"`
	QtyRemaining int       `json:"qty_remaining"`
	Message      string    `json:"message"`
}

type CreateInboundRequest struct {
	EventID         uuid.UUID `json:"event_id"`
	ReferenceNumber string    `json:"reference_number"`
	Notes           string    `json:"notes"`
	Items           []struct {
		SKUCode     string `json:"sku_code"`
		QtyExpected int    `json:"qty_expected"`
	} `json:"items"`
}
