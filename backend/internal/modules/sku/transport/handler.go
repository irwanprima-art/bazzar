package transport

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/irwan/bazzar/internal/middleware"
	"github.com/irwan/bazzar/internal/modules/sku/domain"
	"github.com/irwan/bazzar/internal/modules/sku/usecase"
	"github.com/irwan/bazzar/internal/shared"
)

type SKUHandler struct {
	usecase *usecase.SKUUsecase
}

func NewSKUHandler(uc *usecase.SKUUsecase) *SKUHandler {
	return &SKUHandler{usecase: uc}
}

func (h *SKUHandler) RegisterRoutes(app *fiber.App, authMw *middleware.AuthMiddleware) {
	skus := app.Group("/api/v1/skus", authMw.Protected())

	// Read endpoints (all roles)
	skus.Get("/barcode/:code", h.GetByBarcode)
	skus.Get("/sku-code/:code", h.GetBySKUCode)
	skus.Get("", h.List)
	skus.Get("/:id", h.GetByID)

	// Write endpoints (admin only)
	skus.Post("", authMw.AdminOnly(), h.Create)
	skus.Put("/:id", authMw.AdminOnly(), h.Update)
	skus.Delete("/:id", authMw.AdminOnly(), h.Delete)
}

func (h *SKUHandler) Create(c *fiber.Ctx) error {
	var req domain.CreateSKURequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(shared.ErrorResponse("Invalid request body"))
	}

	sku, err := h.usecase.Create(c.Context(), req)
	if err != nil {
		return c.Status(400).JSON(shared.ErrorResponse(err.Error()))
	}

	return c.Status(201).JSON(shared.SuccessResponse(sku))
}

func (h *SKUHandler) Update(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(shared.ErrorResponse("Invalid SKU ID"))
	}

	var req domain.UpdateSKURequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(shared.ErrorResponse("Invalid request body"))
	}

	sku, err := h.usecase.Update(c.Context(), id, req)
	if err != nil {
		return c.Status(400).JSON(shared.ErrorResponse(err.Error()))
	}

	return c.JSON(shared.SuccessResponse(sku))
}

func (h *SKUHandler) Delete(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(shared.ErrorResponse("Invalid SKU ID"))
	}

	if err := h.usecase.Delete(c.Context(), id); err != nil {
		return c.Status(400).JSON(shared.ErrorResponse("Failed to delete SKU"))
	}

	return c.JSON(shared.SuccessMessageResponse("SKU deleted"))
}

func (h *SKUHandler) GetByID(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(shared.ErrorResponse("Invalid SKU ID"))
	}

	sku, err := h.usecase.GetByID(c.Context(), id)
	if err != nil {
		return c.Status(404).JSON(shared.ErrorResponse("SKU not found"))
	}

	return c.JSON(shared.SuccessResponse(sku))
}

func (h *SKUHandler) GetByBarcode(c *fiber.Ctx) error {
	sku, err := h.usecase.GetByBarcode(c.Context(), c.Params("code"))
	if err != nil {
		return c.Status(404).JSON(shared.ErrorResponse("SKU not found"))
	}
	return c.JSON(shared.SuccessResponse(sku))
}

func (h *SKUHandler) GetBySKUCode(c *fiber.Ctx) error {
	sku, err := h.usecase.GetBySKUCode(c.Context(), c.Params("code"))
	if err != nil {
		return c.Status(404).JSON(shared.ErrorResponse("SKU not found"))
	}
	return c.JSON(shared.SuccessResponse(sku))
}

func (h *SKUHandler) List(c *fiber.Ctx) error {
	search := c.Query("search")
	page, pageSize := shared.ParsePagination(c)

	skus, total, err := h.usecase.List(c.Context(), search, page, pageSize)
	if err != nil {
		return c.Status(500).JSON(shared.ErrorResponse("Failed to list SKUs"))
	}

	return c.JSON(shared.SuccessResponseWithMeta(skus, shared.NewPaginationMeta(page, pageSize, total)))
}
