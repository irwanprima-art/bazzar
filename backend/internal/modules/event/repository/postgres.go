package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/irwan/bazzar/internal/modules/event/domain"
)

type EventRepository struct {
	db *pgxpool.Pool
}

func NewEventRepository(db *pgxpool.Pool) *EventRepository {
	return &EventRepository{db: db}
}

func (r *EventRepository) Create(ctx context.Context, event *domain.Event) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO events (id, name, description, start_date, end_date, is_active)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, event.ID, event.Name, event.Description, event.StartDate, event.EndDate, event.IsActive)
	return err
}

func (r *EventRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Event, error) {
	var event domain.Event
	err := r.db.QueryRow(ctx, `
		SELECT id, name, COALESCE(description,''), 
			   TO_CHAR(start_date, 'YYYY-MM-DD'), TO_CHAR(end_date, 'YYYY-MM-DD'),
			   is_active, created_at, updated_at
		FROM events WHERE id = $1
	`, id).Scan(
		&event.ID, &event.Name, &event.Description,
		&event.StartDate, &event.EndDate,
		&event.IsActive, &event.CreatedAt, &event.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &event, nil
}

func (r *EventRepository) List(ctx context.Context) ([]domain.Event, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, name, COALESCE(description,''),
			   TO_CHAR(start_date, 'YYYY-MM-DD'), TO_CHAR(end_date, 'YYYY-MM-DD'),
			   is_active, created_at, updated_at
		FROM events ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []domain.Event
	for rows.Next() {
		var e domain.Event
		if err := rows.Scan(&e.ID, &e.Name, &e.Description, &e.StartDate, &e.EndDate,
			&e.IsActive, &e.CreatedAt, &e.UpdatedAt); err != nil {
			return nil, err
		}
		events = append(events, e)
	}
	return events, nil
}

func (r *EventRepository) GetActiveEvent(ctx context.Context) (*domain.Event, error) {
	var event domain.Event
	err := r.db.QueryRow(ctx, `
		SELECT id, name, COALESCE(description,''),
			   TO_CHAR(start_date, 'YYYY-MM-DD'), TO_CHAR(end_date, 'YYYY-MM-DD'),
			   is_active, created_at, updated_at
		FROM events WHERE is_active = true ORDER BY created_at DESC LIMIT 1
	`).Scan(
		&event.ID, &event.Name, &event.Description,
		&event.StartDate, &event.EndDate,
		&event.IsActive, &event.CreatedAt, &event.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &event, nil
}

// CreateLocations creates the two default locations for an event
func (r *EventRepository) CreateLocations(ctx context.Context, eventID uuid.UUID) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO locations (id, event_id, code, name) VALUES
			($1, $2, 'EVENT', 'Event Floor'),
			($3, $2, 'STORAGE', 'Storage Area')
		ON CONFLICT (event_id, code) DO NOTHING
	`, uuid.New(), eventID, uuid.New())
	return err
}

func (r *EventRepository) GetLocations(ctx context.Context, eventID uuid.UUID) ([]domain.Location, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, event_id, code, name, created_at
		FROM locations WHERE event_id = $1 ORDER BY code
	`, eventID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var locs []domain.Location
	for rows.Next() {
		var l domain.Location
		if err := rows.Scan(&l.ID, &l.EventID, &l.Code, &l.Name, &l.CreatedAt); err != nil {
			return nil, err
		}
		locs = append(locs, l)
	}
	return locs, nil
}

func (r *EventRepository) GetLocationByCode(ctx context.Context, eventID uuid.UUID, code string) (*domain.Location, error) {
	var loc domain.Location
	err := r.db.QueryRow(ctx, `
		SELECT id, event_id, code, name, created_at
		FROM locations WHERE event_id = $1 AND code = $2
	`, eventID, code).Scan(&loc.ID, &loc.EventID, &loc.Code, &loc.Name, &loc.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &loc, nil
}
