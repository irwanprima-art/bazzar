package transport

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/irwan/bazzar/internal/middleware"
	"github.com/irwan/bazzar/internal/modules/event/domain"
	"github.com/irwan/bazzar/internal/modules/event/usecase"
	"github.com/irwan/bazzar/internal/shared"
)

type EventHandler struct {
	usecase *usecase.EventUsecase
}

func NewEventHandler(uc *usecase.EventUsecase) *EventHandler {
	return &EventHandler{usecase: uc}
}

func (h *EventHandler) RegisterRoutes(app *fiber.App, authMw *middleware.AuthMiddleware) {
	events := app.Group("/api/v1/events", authMw.Protected())
	events.Get("", h.List)
	events.Get("/active", h.GetActive)
	events.Get("/:id", h.GetByID)
	events.Get("/:id/locations", h.GetLocations)

	// Admin only
	events.Post("", authMw.AdminOnly(), h.Create)
}

func (h *EventHandler) Create(c *fiber.Ctx) error {
	var req domain.CreateEventRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(shared.ErrorResponse("Invalid request body"))
	}

	event, err := h.usecase.Create(c.Context(), req)
	if err != nil {
		return c.Status(400).JSON(shared.ErrorResponse(err.Error()))
	}

	return c.Status(201).JSON(shared.SuccessResponse(event))
}

func (h *EventHandler) List(c *fiber.Ctx) error {
	events, err := h.usecase.List(c.Context())
	if err != nil {
		return c.Status(500).JSON(shared.ErrorResponse("Failed to list events"))
	}
	return c.JSON(shared.SuccessResponse(events))
}

func (h *EventHandler) GetByID(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(shared.ErrorResponse("Invalid event ID"))
	}

	event, err := h.usecase.GetByID(c.Context(), id)
	if err != nil {
		return c.Status(404).JSON(shared.ErrorResponse("Event not found"))
	}

	return c.JSON(shared.SuccessResponse(event))
}

func (h *EventHandler) GetActive(c *fiber.Ctx) error {
	event, err := h.usecase.GetActiveEvent(c.Context())
	if err != nil {
		return c.Status(404).JSON(shared.ErrorResponse("No active event found"))
	}
	return c.JSON(shared.SuccessResponse(event))
}

func (h *EventHandler) GetLocations(c *fiber.Ctx) error {
	eventID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(shared.ErrorResponse("Invalid event ID"))
	}

	locs, err := h.usecase.GetLocations(c.Context(), eventID)
	if err != nil {
		return c.Status(500).JSON(shared.ErrorResponse("Failed to get locations"))
	}
	return c.JSON(shared.SuccessResponse(locs))
}
