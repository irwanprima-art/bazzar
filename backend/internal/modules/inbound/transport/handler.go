package transport

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/irwan/bazzar/internal/middleware"
	"github.com/irwan/bazzar/internal/modules/inbound/domain"
	"github.com/irwan/bazzar/internal/modules/inbound/usecase"
	"github.com/irwan/bazzar/internal/shared"
)

type InboundHandler struct {
	usecase *usecase.InboundUsecase
}

func NewInboundHandler(uc *usecase.InboundUsecase) *InboundHandler {
	return &InboundHandler{usecase: uc}
}

func (h *InboundHandler) RegisterRoutes(app *fiber.App, authMw *middleware.AuthMiddleware) {
	inb := app.Group("/api/v1/inbound", authMw.Protected())
	inb.Get("", h.List)
	inb.Get("/:id", h.GetByID)
	inb.Post("/import", authMw.AdminOnly(), h.Import)
	inb.Post("", authMw.AdminOnly(), h.CreateManual)
	inb.Post("/:id/scan", h.ScanItem)
}

func (h *InboundHandler) Import(c *fiber.Ctx) error {
	eventID, err := uuid.Parse(c.FormValue("event_id"))
	if err != nil {
		return c.Status(400).JSON(shared.ErrorResponse("event_id is required"))
	}

	refNumber := c.FormValue("reference_number")
	if refNumber == "" {
		refNumber = "INB-" + uuid.New().String()[:8]
	}

	file, err := c.FormFile("file")
	if err != nil {
		return c.Status(400).JSON(shared.ErrorResponse("No file uploaded"))
	}

	src, err := file.Open()
	if err != nil {
		return c.Status(500).JSON(shared.ErrorResponse("Failed to open file"))
	}
	defer src.Close()

	userID := middleware.GetUserID(c)
	result, err := h.usecase.ImportFromExcel(c.Context(), src, eventID, refNumber, userID)
	if err != nil {
		return c.Status(400).JSON(shared.ErrorResponse(err.Error()))
	}

	return c.Status(201).JSON(shared.SuccessResponse(result))
}

func (h *InboundHandler) CreateManual(c *fiber.Ctx) error {
	var req domain.CreateInboundRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(shared.ErrorResponse("Invalid request body"))
	}

	userID := middleware.GetUserID(c)
	result, err := h.usecase.CreateManual(c.Context(), req, userID)
	if err != nil {
		return c.Status(400).JSON(shared.ErrorResponse(err.Error()))
	}

	return c.Status(201).JSON(shared.SuccessResponse(result))
}

func (h *InboundHandler) List(c *fiber.Ctx) error {
	eventID, err := uuid.Parse(c.Query("event_id"))
	if err != nil {
		return c.Status(400).JSON(shared.ErrorResponse("event_id is required"))
	}

	orders, err := h.usecase.List(c.Context(), eventID)
	if err != nil {
		return c.Status(500).JSON(shared.ErrorResponse("Failed to list inbound orders"))
	}

	return c.JSON(shared.SuccessResponse(orders))
}

func (h *InboundHandler) GetByID(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(shared.ErrorResponse("Invalid inbound ID"))
	}

	order, err := h.usecase.GetByID(c.Context(), id)
	if err != nil {
		return c.Status(404).JSON(shared.ErrorResponse("Inbound order not found"))
	}

	return c.JSON(shared.SuccessResponse(order))
}

func (h *InboundHandler) ScanItem(c *fiber.Ctx) error {
	inboundID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(shared.ErrorResponse("Invalid inbound ID"))
	}

	var req domain.InboundScanRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(shared.ErrorResponse("Invalid request body"))
	}

	eventIDStr := c.Query("event_id", c.FormValue("event_id"))
	eventID, err := uuid.Parse(eventIDStr)
	if err != nil {
		// Try to get from the inbound order itself
		order, oErr := h.usecase.GetByID(c.Context(), inboundID)
		if oErr != nil {
			return c.Status(400).JSON(shared.ErrorResponse("event_id is required"))
		}
		eventID = order.EventID
	}

	userID := middleware.GetUserID(c)
	result, err := h.usecase.ScanItem(c.Context(), inboundID, req, eventID, userID)
	if err != nil {
		return c.Status(400).JSON(shared.ErrorResponse(err.Error()))
	}

	return c.JSON(shared.SuccessResponse(result))
}
