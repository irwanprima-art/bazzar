package transport

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/irwan/bazzar/internal/middleware"
	"github.com/irwan/bazzar/internal/modules/inventory/domain"
	"github.com/irwan/bazzar/internal/modules/inventory/usecase"
	eventRepo "github.com/irwan/bazzar/internal/modules/event/repository"
	"github.com/irwan/bazzar/internal/shared"
)

type InventoryHandler struct {
	usecase   *usecase.InventoryUsecase
	eventRepo *eventRepo.EventRepository
}

func NewInventoryHandler(uc *usecase.InventoryUsecase, er *eventRepo.EventRepository) *InventoryHandler {
	return &InventoryHandler{usecase: uc, eventRepo: er}
}

func (h *InventoryHandler) RegisterRoutes(app *fiber.App, authMw *middleware.AuthMiddleware) {
	inv := app.Group("/api/v1/inventory", authMw.Protected())
	inv.Get("", h.List)
	inv.Get("/alerts", h.GetAlerts)
	inv.Get("/logs", h.GetLogs)
	inv.Get("/sales-report", h.GetSalesReport)
	inv.Post("/adjust", authMw.AdminOnly(), h.Adjust)
	inv.Post("/replenish", authMw.AdminOnly(), h.Replenish)
}

func (h *InventoryHandler) List(c *fiber.Ctx) error {
	eventID, err := uuid.Parse(c.Query("event_id"))
	if err != nil {
		return c.Status(400).JSON(shared.ErrorResponse("event_id is required"))
	}

	locationCode := c.Query("location")
	items, err := h.usecase.ListByEvent(c.Context(), eventID, locationCode)
	if err != nil {
		return c.Status(500).JSON(shared.ErrorResponse("Failed to list inventory"))
	}

	return c.JSON(shared.SuccessResponse(items))
}

func (h *InventoryHandler) Adjust(c *fiber.Ctx) error {
	var req domain.AdjustRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(shared.ErrorResponse("Invalid request body"))
	}

	userID := middleware.GetUserID(c)
	if err := h.usecase.Adjust(c.Context(), req, userID); err != nil {
		return c.Status(400).JSON(shared.ErrorResponse(err.Error()))
	}

	return c.JSON(shared.SuccessMessageResponse("Stock adjusted"))
}

func (h *InventoryHandler) Replenish(c *fiber.Ctx) error {
	var req domain.ReplenishRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(shared.ErrorResponse("Invalid request body"))
	}

	eventLoc, err := h.eventRepo.GetLocationByCode(c.Context(), req.EventID, "EVENT")
	if err != nil {
		return c.Status(400).JSON(shared.ErrorResponse("Event location not found"))
	}
	storageLoc, err := h.eventRepo.GetLocationByCode(c.Context(), req.EventID, "STORAGE")
	if err != nil {
		return c.Status(400).JSON(shared.ErrorResponse("Storage location not found"))
	}

	userID := middleware.GetUserID(c)
	if err := h.usecase.Replenish(c.Context(), req, eventLoc.ID, storageLoc.ID, userID); err != nil {
		return c.Status(400).JSON(shared.ErrorResponse(err.Error()))
	}

	return c.JSON(shared.SuccessMessageResponse("Stock replenished"))
}

func (h *InventoryHandler) GetAlerts(c *fiber.Ctx) error {
	eventID, err := uuid.Parse(c.Query("event_id"))
	if err != nil {
		return c.Status(400).JSON(shared.ErrorResponse("event_id is required"))
	}

	alerts, err := h.usecase.GetReplenishAlerts(c.Context(), eventID)
	if err != nil {
		return c.Status(500).JSON(shared.ErrorResponse("Failed to get alerts"))
	}

	return c.JSON(shared.SuccessResponse(alerts))
}

func (h *InventoryHandler) GetLogs(c *fiber.Ctx) error {
	eventID, err := uuid.Parse(c.Query("event_id"))
	if err != nil {
		return c.Status(400).JSON(shared.ErrorResponse("event_id is required"))
	}

	var skuID *uuid.UUID
	if s := c.Query("sku_id"); s != "" {
		id, err := uuid.Parse(s)
		if err == nil {
			skuID = &id
		}
	}

	limit := c.QueryInt("limit", 50)
	logs, err := h.usecase.GetLogs(c.Context(), eventID, skuID, limit)
	if err != nil {
		return c.Status(500).JSON(shared.ErrorResponse("Failed to get logs"))
	}

	return c.JSON(shared.SuccessResponse(logs))
}

func (h *InventoryHandler) GetSalesReport(c *fiber.Ctx) error {
	eventID, err := uuid.Parse(c.Query("event_id"))
	if err != nil {
		return c.Status(400).JSON(shared.ErrorResponse("event_id is required"))
	}

	report, err := h.usecase.GetSalesReport(c.Context(), eventID)
	if err != nil {
		return c.Status(500).JSON(shared.ErrorResponse("Failed to get sales report"))
	}

	return c.JSON(shared.SuccessResponse(report))
}
