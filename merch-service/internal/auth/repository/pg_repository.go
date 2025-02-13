package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"cyansnbrst/merch-service/internal/auth"
	m "cyansnbrst/merch-service/internal/models"
	"cyansnbrst/merch-service/pkg/db"
)

// Auth repository struct
type authRepo struct {
	db *pgxpool.Pool
}

// Auth repository constructor
func NewAuthRepo(db *pgxpool.Pool) auth.Repository {
	return &authRepo{db: db}
}

// Create a new user
func (r *authRepo) CreateUser(ctx context.Context, username, passwordHash string) (*m.User, error) {
	query := `
		INSERT INTO users (username, password_hash)
		VALUES ($1, $2)
		RETURNING id, username, password_hash, balance, created_at
	`

	var user m.User
	err := r.db.QueryRow(ctx, query, username, passwordHash).Scan(
		&user.ID,
		&user.Username,
		&user.PasswordHash,
		&user.Balance,
		&user.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("repo - failed to create user: %w", err)
	}

	return &user, nil
}

// Get user by username
func (r *authRepo) GetUserByUsername(ctx context.Context, username string) (*m.User, error) {
	query := `
		SELECT id, username, password_hash, balance, created_at
		FROM users
		WHERE username = $1
	`

	var user m.User
	err := r.db.QueryRow(ctx, query, username).Scan(
		&user.ID,
		&user.Username,
		&user.PasswordHash,
		&user.Balance,
		&user.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, db.ErrUserNotFound
		}
		return nil, fmt.Errorf("repo - failed to get user: %w", err)
	}

	return &user, nil
}
