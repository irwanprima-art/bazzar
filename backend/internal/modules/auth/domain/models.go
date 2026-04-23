package domain

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID `json:"id"`
	Username     string    `json:"username"`
	PasswordHash string    `json:"-"`
	FullName     string    `json:"full_name"`
	Role         string    `json:"role"`
	IsActive     bool      `json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token    string `json:"token"`
	User     User   `json:"user"`
}

type CreateUserRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	FullName string `json:"full_name"`
	Role     string `json:"role"`
}
