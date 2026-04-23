package domain

import (
	"time"

	"github.com/google/uuid"
)

type SKU struct {
	ID             uuid.UUID `json:"id"`
	SKUCode        string    `json:"sku_code"`
	Barcode        *string   `json:"barcode"`
	Name           string    `json:"name"`
	Description    string    `json:"description"`
	ReplenishLimit int       `json:"replenish_limit"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type CreateSKURequest struct {
	SKUCode        string `json:"sku_code"`
	Barcode        string `json:"barcode"`
	Name           string `json:"name"`
	Description    string `json:"description"`
	ReplenishLimit int    `json:"replenish_limit"`
}

type UpdateSKURequest struct {
	SKUCode        string `json:"sku_code"`
	Barcode        string `json:"barcode"`
	Name           string `json:"name"`
	Description    string `json:"description"`
	ReplenishLimit int    `json:"replenish_limit"`
}
