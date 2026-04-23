package transport

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/irwan/bazzar/internal/middleware"
	"github.com/irwan/bazzar/internal/modules/picking/domain"
	"github.com/irwan/bazzar/internal/modules/picking/usecase"
	"github.com/irwan/bazzar/internal/shared"
)

type PickingHandler struct {
	usecase *usecase.PickingUsecase
}

func NewPickingHandler(uc *usecase.PickingUsecase) *PickingHandler {
	return &PickingHandler{usecase: uc}
}

func (h *PickingHandler) RegisterRoutes(app *fiber.App, authMw *middleware.AuthMiddleware) {
	picking := app.Group("/api/v1/picking", authMw.Protected())
	picking.Post("/:orderId/start", h.StartPicking)
	picking.Post("/:orderId/scan", h.ScanItem)
	picking.Post("/:orderId/complete", h.CompletePicking)

	handover := app.Group("/api/v1/handover", authMw.Protected())
	handover.Post("/scan", h.Ship)
}

func (h *PickingHandler) StartPicking(c *fiber.Ctx) error {
	orderID, err := uuid.Parse(c.Params("orderId"))
	if err != nil {
		return c.Status(400).JSON(shared.ErrorResponse("Invalid order ID"))
	}

	userID := middleware.GetUserID(c)
	if err := h.usecase.StartPicking(c.Context(), orderID, userID); err != nil {
		return c.Status(400).JSON(shared.ErrorResponse(err.Error()))
	}

	return c.JSON(shared.SuccessMessageResponse("Picking started"))
}

func (h *PickingHandler) ScanItem(c *fiber.Ctx) error {
	orderID, err := uuid.Parse(c.Params("orderId"))
	if err != nil {
		return c.Status(400).JSON(shared.ErrorResponse("Invalid order ID"))
	}

	var req domain.ScanRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(shared.ErrorResponse("Invalid request body"))
	}

	userID := middleware.GetUserID(c)
	result, err := h.usecase.ScanItem(c.Context(), orderID, req, userID)
	if err != nil {
		return c.Status(400).JSON(shared.ErrorResponse(err.Error()))
	}

	return c.JSON(shared.SuccessResponse(result))
}

func (h *PickingHandler) CompletePicking(c *fiber.Ctx) error {
	orderID, err := uuid.Parse(c.Params("orderId"))
	if err != nil {
		return c.Status(400).JSON(shared.ErrorResponse("Invalid order ID"))
	}

	userID := middleware.GetUserID(c)
	if err := h.usecase.CompletePicking(c.Context(), orderID, userID); err != nil {
		return c.Status(400).JSON(shared.ErrorResponse(err.Error()))
	}

	return c.JSON(shared.SuccessMessageResponse("Picking completed"))
}

func (h *PickingHandler) Ship(c *fiber.Ctx) error {
	var req struct {
		OrderNumber string    `json:"order_number"`
		EventID     uuid.UUID `json:"event_id"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(shared.ErrorResponse("Invalid request body"))
	}

	userID := middleware.GetUserID(c)
	if err := h.usecase.Ship(c.Context(), req.OrderNumber, req.EventID, userID); err != nil {
		return c.Status(400).JSON(shared.ErrorResponse(err.Error()))
	}

	return c.JSON(shared.SuccessMessageResponse("Order shipped"))
}
