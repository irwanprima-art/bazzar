package middleware

import (
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/irwan/bazzar/internal/shared"
)

type AuthMiddleware struct {
	jwtSecret string
}

type JWTClaims struct {
	UserID   uuid.UUID `json:"user_id"`
	Username string    `json:"username"`
	Role     string    `json:"role"`
	FullName string    `json:"full_name"`
	jwt.RegisteredClaims
}

func NewAuthMiddleware(secret string) *AuthMiddleware {
	return &AuthMiddleware{jwtSecret: secret}
}

func (m *AuthMiddleware) GenerateToken(userID uuid.UUID, username, role, fullName string) (string, error) {
	claims := JWTClaims{
		UserID:   userID,
		Username: username,
		Role:     role,
		FullName: fullName,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(m.jwtSecret))
}

func (m *AuthMiddleware) Protected() fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(401).JSON(shared.ErrorResponse("Missing authorization header"))
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			return c.Status(401).JSON(shared.ErrorResponse("Invalid authorization format"))
		}

		tokenStr := parts[1]
		claims := &JWTClaims{}

		token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
			return []byte(m.jwtSecret), nil
		})

		if err != nil || !token.Valid {
			return c.Status(401).JSON(shared.ErrorResponse("Invalid or expired token"))
		}

		// Store user info in context
		c.Locals("user_id", claims.UserID)
		c.Locals("username", claims.Username)
		c.Locals("role", claims.Role)
		c.Locals("full_name", claims.FullName)

		return c.Next()
	}
}

func (m *AuthMiddleware) AdminOnly() fiber.Handler {
	return func(c *fiber.Ctx) error {
		role, ok := c.Locals("role").(string)
		if !ok || role != "admin" {
			return c.Status(403).JSON(shared.ErrorResponse("Admin access required"))
		}
		return c.Next()
	}
}

// Helper to extract user ID from context
func GetUserID(c *fiber.Ctx) uuid.UUID {
	if id, ok := c.Locals("user_id").(uuid.UUID); ok {
		return id
	}
	return uuid.Nil
}

func GetUsername(c *fiber.Ctx) string {
	if u, ok := c.Locals("username").(string); ok {
		return u
	}
	return ""
}

func GetUserRole(c *fiber.Ctx) string {
	if r, ok := c.Locals("role").(string); ok {
		return r
	}
	return ""
}
