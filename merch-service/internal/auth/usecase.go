package auth

import (
	"context"
	"cyansnbrst/merch-service/internal/models"
)

// Auth usecase interface
type UseCase interface {
	LoginOrRegister(ctx context.Context, username, password string) (string, error)
	GenerateJWT(user *models.User) (string, error)
}
