package auth

import (
	"context"

	m "cyansnbrst/merch-service/internal/models"
)

// Auth repository interface
type Repository interface {
	CreateUser(ctx context.Context, username, passwordHash string) (*m.User, error)
	GetUserByUsername(ctx context.Context, username string) (*m.User, error)
}
