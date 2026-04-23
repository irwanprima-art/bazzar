package transport

import (
	"github.com/gofiber/fiber/v2"
	"github.com/irwan/bazzar/internal/middleware"
	"github.com/irwan/bazzar/internal/modules/auth/domain"
	"github.com/irwan/bazzar/internal/modules/auth/usecase"
	"github.com/irwan/bazzar/internal/shared"
)

type AuthHandler struct {
	usecase *usecase.AuthUsecase
}

func NewAuthHandler(uc *usecase.AuthUsecase) *AuthHandler {
	return &AuthHandler{usecase: uc}
}

func (h *AuthHandler) RegisterRoutes(app *fiber.App, authMw *middleware.AuthMiddleware) {
	auth := app.Group("/api/v1/auth")
	auth.Post("/login", h.Login)

	// Protected routes
	auth.Use(authMw.Protected())
	auth.Get("/me", h.GetMe)

	// Admin only
	users := app.Group("/api/v1/users", authMw.Protected(), authMw.AdminOnly())
	users.Post("", h.CreateUser)
	users.Get("", h.ListUsers)
}

func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req domain.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(shared.ErrorResponse("Invalid request body"))
	}

	resp, err := h.usecase.Login(c.Context(), req)
	if err != nil {
		return c.Status(401).JSON(shared.ErrorResponse(err.Error()))
	}

	return c.JSON(shared.SuccessResponse(resp))
}

func (h *AuthHandler) GetMe(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	user, err := h.usecase.GetMe(c.Context(), userID)
	if err != nil {
		return c.Status(404).JSON(shared.ErrorResponse("User not found"))
	}
	return c.JSON(shared.SuccessResponse(user))
}

func (h *AuthHandler) CreateUser(c *fiber.Ctx) error {
	var req domain.CreateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(shared.ErrorResponse("Invalid request body"))
	}

	user, err := h.usecase.CreateUser(c.Context(), req)
	if err != nil {
		return c.Status(400).JSON(shared.ErrorResponse(err.Error()))
	}

	return c.Status(201).JSON(shared.SuccessResponse(user))
}

func (h *AuthHandler) ListUsers(c *fiber.Ctx) error {
	users, err := h.usecase.ListUsers(c.Context())
	if err != nil {
		return c.Status(500).JSON(shared.ErrorResponse("Failed to list users"))
	}
	return c.JSON(shared.SuccessResponse(users))
}
