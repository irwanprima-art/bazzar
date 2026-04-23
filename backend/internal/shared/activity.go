package shared

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ActivityLogger struct {
	db *pgxpool.Pool
}

func NewActivityLogger(db *pgxpool.Pool) *ActivityLogger {
	return &ActivityLogger{db: db}
}

func (l *ActivityLogger) Log(ctx context.Context, userID, eventID uuid.UUID, action, entityType string, entityID *uuid.UUID, details interface{}, ip string) {
	l.db.Exec(ctx, `
		INSERT INTO activity_logs (id, user_id, event_id, action, entity_type, entity_id, details, ip_address)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, uuid.New(), userID, eventID, action, entityType, entityID, details, ip)
}
