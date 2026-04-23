package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/irwan/bazzar/internal/modules/auth/domain"
)

type AuthRepository struct {
	db *pgxpool.Pool
}

func NewAuthRepository(db *pgxpool.Pool) *AuthRepository {
	return &AuthRepository{db: db}
}

func (r *AuthRepository) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
	var user domain.User
	err := r.db.QueryRow(ctx, `
		SELECT id, username, password_hash, full_name, role, is_active, created_at, updated_at
		FROM users WHERE username = $1
	`, username).Scan(
		&user.ID, &user.Username, &user.PasswordHash, &user.FullName,
		&user.Role, &user.IsActive, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *AuthRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	var user domain.User
	err := r.db.QueryRow(ctx, `
		SELECT id, username, password_hash, full_name, role, is_active, created_at, updated_at
		FROM users WHERE id = $1
	`, id).Scan(
		&user.ID, &user.Username, &user.PasswordHash, &user.FullName,
		&user.Role, &user.IsActive, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *AuthRepository) Create(ctx context.Context, user *domain.User) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO users (id, username, password_hash, full_name, role, is_active)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, user.ID, user.Username, user.PasswordHash, user.FullName, user.Role, user.IsActive)
	return err
}

func (r *AuthRepository) ListUsers(ctx context.Context) ([]domain.User, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, username, full_name, role, is_active, created_at, updated_at
		FROM users ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []domain.User
	for rows.Next() {
		var u domain.User
		err := rows.Scan(&u.ID, &u.Username, &u.FullName, &u.Role, &u.IsActive, &u.CreatedAt, &u.UpdatedAt)
		if err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, nil
}
