package domain

import (
	"time"

	"github.com/google/uuid"
)

type Event struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	StartDate   *string   `json:"start_date"`
	EndDate     *string   `json:"end_date"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type CreateEventRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	StartDate   string `json:"start_date"`
	EndDate     string `json:"end_date"`
}

type Location struct {
	ID        uuid.UUID `json:"id"`
	EventID   uuid.UUID `json:"event_id"`
	Code      string    `json:"code"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}
