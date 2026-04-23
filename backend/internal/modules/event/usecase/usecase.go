package usecase

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/irwan/bazzar/internal/modules/event/domain"
	"github.com/irwan/bazzar/internal/modules/event/repository"
)

type EventUsecase struct {
	repo *repository.EventRepository
}

func NewEventUsecase(repo *repository.EventRepository) *EventUsecase {
	return &EventUsecase{repo: repo}
}

func (u *EventUsecase) Create(ctx context.Context, req domain.CreateEventRequest) (*domain.Event, error) {
	event := &domain.Event{
		ID:          uuid.New(),
		Name:        req.Name,
		Description: req.Description,
		StartDate:   &req.StartDate,
		EndDate:     &req.EndDate,
		IsActive:    true,
	}

	if err := u.repo.Create(ctx, event); err != nil {
		return nil, errors.New("failed to create event")
	}

	// Auto-create EVENT and STORAGE locations
	if err := u.repo.CreateLocations(ctx, event.ID); err != nil {
		return nil, errors.New("failed to create locations")
	}

	return event, nil
}

func (u *EventUsecase) GetByID(ctx context.Context, id uuid.UUID) (*domain.Event, error) {
	return u.repo.GetByID(ctx, id)
}

func (u *EventUsecase) List(ctx context.Context) ([]domain.Event, error) {
	return u.repo.List(ctx)
}

func (u *EventUsecase) GetActiveEvent(ctx context.Context) (*domain.Event, error) {
	return u.repo.GetActiveEvent(ctx)
}

func (u *EventUsecase) GetLocations(ctx context.Context, eventID uuid.UUID) ([]domain.Location, error) {
	return u.repo.GetLocations(ctx, eventID)
}

func (u *EventUsecase) GetLocationByCode(ctx context.Context, eventID uuid.UUID, code string) (*domain.Location, error) {
	return u.repo.GetLocationByCode(ctx, eventID, code)
}

// EnsureDefaultEvent creates a default event if none exists
func (u *EventUsecase) EnsureDefaultEvent(ctx context.Context) {
	events, err := u.repo.List(ctx)
	if err != nil || len(events) == 0 {
		event := &domain.Event{
			ID:          uuid.New(),
			Name:        "Bazzar Makuku",
			Description: "Bazzar Makuku Event",
			IsActive:    true,
		}
		sd := "2026-04-23"
		ed := "2026-04-30"
		event.StartDate = &sd
		event.EndDate = &ed
		u.repo.Create(ctx, event)
		u.repo.CreateLocations(ctx, event.ID)
	} else {
		// Ensure locations exist for existing events
		for _, e := range events {
			u.repo.CreateLocations(ctx, e.ID)
		}
	}
}
