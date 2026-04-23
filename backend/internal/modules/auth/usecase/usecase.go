package usecase

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/irwan/bazzar/internal/middleware"
	"github.com/irwan/bazzar/internal/modules/auth/domain"
	"github.com/irwan/bazzar/internal/modules/auth/repository"
	"golang.org/x/crypto/bcrypt"
)

type AuthUsecase struct {
	repo *repository.AuthRepository
	auth *middleware.AuthMiddleware
}

func NewAuthUsecase(repo *repository.AuthRepository, auth *middleware.AuthMiddleware) *AuthUsecase {
	return &AuthUsecase{repo: repo, auth: auth}
}

func (u *AuthUsecase) Login(ctx context.Context, req domain.LoginRequest) (*domain.LoginResponse, error) {
	user, err := u.repo.GetByUsername(ctx, req.Username)
	if err != nil {
		return nil, errors.New("invalid username or password")
	}

	if !user.IsActive {
		return nil, errors.New("account is deactivated")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, errors.New("invalid username or password")
	}

	token, err := u.auth.GenerateToken(user.ID, user.Username, user.Role, user.FullName)
	if err != nil {
		return nil, errors.New("failed to generate token")
	}

	return &domain.LoginResponse{
		Token: token,
		User:  *user,
	}, nil
}

func (u *AuthUsecase) GetMe(ctx context.Context, userID uuid.UUID) (*domain.User, error) {
	return u.repo.GetByID(ctx, userID)
}

func (u *AuthUsecase) CreateUser(ctx context.Context, req domain.CreateUserRequest) (*domain.User, error) {
	if req.Role != "admin" && req.Role != "picker" {
		return nil, errors.New("role must be 'admin' or 'picker'")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.New("failed to hash password")
	}

	user := &domain.User{
		ID:           uuid.New(),
		Username:     req.Username,
		PasswordHash: string(hash),
		FullName:     req.FullName,
		Role:         req.Role,
		IsActive:     true,
	}

	if err := u.repo.Create(ctx, user); err != nil {
		return nil, errors.New("username already exists")
	}

	return user, nil
}

func (u *AuthUsecase) ListUsers(ctx context.Context) ([]domain.User, error) {
	return u.repo.ListUsers(ctx)
}

// EnsureDefaultAdmin creates the default admin if not exists
func (u *AuthUsecase) EnsureDefaultAdmin(ctx context.Context) {
	_, err := u.repo.GetByUsername(ctx, "admin")
	if err != nil {
		hash, _ := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
		admin := &domain.User{
			ID:           uuid.New(),
			Username:     "admin",
			PasswordHash: string(hash),
			FullName:     "Administrator",
			Role:         "admin",
			IsActive:     true,
		}
		u.repo.Create(ctx, admin)
	}
}
