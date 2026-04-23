package transport

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/irwan/bazzar/internal/middleware"
	"github.com/irwan/bazzar/internal/modules/order/domain"
	"github.com/irwan/bazzar/internal/modules/order/usecase"
	"github.com/irwan/bazzar/internal/shared"
)

type OrderHandler struct {
	usecase *usecase.OrderUsecase
}

func NewOrderHandler(uc *usecase.OrderUsecase) *OrderHandler {
	return &OrderHandler{usecase: uc}
}

func (h *OrderHandler) RegisterRoutes(app *fiber.App, authMw *middleware.AuthMiddleware) {
	orders := app.Group("/api/v1/orders", authMw.Protected())
	orders.Get("", h.List)
	orders.Get("/status-counts", h.GetStatusCounts)
	orders.Get("/:id", h.GetByID)
	orders.Get("/:id/items", h.GetOrderItems)
	orders.Get("/:id/label", h.GetLabel)

	// Admin only
	orders.Post("/import", authMw.AdminOnly(), h.Import)
	orders.Post("/allocate-all", authMw.AdminOnly(), h.AllocateAll)
	orders.Post("/:id/print", authMw.AdminOnly(), h.MarkPrinted)
}

func (h *OrderHandler) Import(c *fiber.Ctx) error {
	eventID, err := uuid.Parse(c.FormValue("event_id"))
	if err != nil {
		return c.Status(400).JSON(shared.ErrorResponse("event_id is required"))
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
	result, err := h.usecase.ImportFromExcel(c.Context(), src, eventID, userID)
	if err != nil {
		return c.Status(400).JSON(shared.ErrorResponse(err.Error()))
	}

	return c.JSON(shared.SuccessResponse(result))
}

func (h *OrderHandler) List(c *fiber.Ctx) error {
	eventID, err := uuid.Parse(c.Query("event_id"))
	if err != nil {
		return c.Status(400).JSON(shared.ErrorResponse("event_id is required"))
	}

	page, pageSize := shared.ParsePagination(c)
	filter := domain.OrderFilter{
		EventID:  eventID,
		Status:   c.Query("status"),
		Search:   c.Query("search"),
		Page:     page,
		PageSize: pageSize,
	}

	orders, total, err := h.usecase.List(c.Context(), filter)
	if err != nil {
		return c.Status(500).JSON(shared.ErrorResponse("Failed to list orders"))
	}

	return c.JSON(shared.SuccessResponseWithMeta(orders, shared.NewPaginationMeta(page, pageSize, total)))
}

func (h *OrderHandler) GetByID(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(shared.ErrorResponse("Invalid order ID"))
	}

	order, err := h.usecase.GetByID(c.Context(), id)
	if err != nil {
		return c.Status(404).JSON(shared.ErrorResponse("Order not found"))
	}

	return c.JSON(shared.SuccessResponse(order))
}

func (h *OrderHandler) GetOrderItems(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(shared.ErrorResponse("Invalid order ID"))
	}

	items, err := h.usecase.GetOrderItems(c.Context(), id)
	if err != nil {
		return c.Status(500).JSON(shared.ErrorResponse("Failed to get order items"))
	}

	return c.JSON(shared.SuccessResponse(items))
}

func (h *OrderHandler) AllocateAll(c *fiber.Ctx) error {
	eventID, err := uuid.Parse(c.Query("event_id", c.FormValue("event_id")))
	if err != nil {
		return c.Status(400).JSON(shared.ErrorResponse("event_id is required"))
	}

	userID := middleware.GetUserID(c)
	result, err := h.usecase.AllocateAll(c.Context(), eventID, userID)
	if err != nil {
		return c.Status(500).JSON(shared.ErrorResponse(err.Error()))
	}

	return c.JSON(shared.SuccessResponse(result))
}

func (h *OrderHandler) MarkPrinted(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(shared.ErrorResponse("Invalid order ID"))
	}

	userID := middleware.GetUserID(c)
	if err := h.usecase.MarkPrinted(c.Context(), id, userID); err != nil {
		return c.Status(400).JSON(shared.ErrorResponse(err.Error()))
	}

	return c.JSON(shared.SuccessMessageResponse("Order marked as printed"))
}

func (h *OrderHandler) GetLabel(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(shared.ErrorResponse("Invalid order ID"))
	}

	order, err := h.usecase.GetByID(c.Context(), id)
	if err != nil {
		return c.Status(404).JSON(shared.ErrorResponse("Order not found"))
	}

	return c.JSON(shared.SuccessResponse(order))
}

func (h *OrderHandler) GetStatusCounts(c *fiber.Ctx) error {
	eventID, err := uuid.Parse(c.Query("event_id"))
	if err != nil {
		return c.Status(400).JSON(shared.ErrorResponse("event_id is required"))
	}

	counts, err := h.usecase.GetStatusCounts(c.Context(), eventID)
	if err != nil {
		return c.Status(500).JSON(shared.ErrorResponse("Failed to get status counts"))
	}

	return c.JSON(shared.SuccessResponse(counts))
}
