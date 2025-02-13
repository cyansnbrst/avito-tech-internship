package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/alexedwards/argon2id"
	"github.com/golang-jwt/jwt/v5"

	"cyansnbrst/merch-service/config"
	"cyansnbrst/merch-service/internal/auth"
	"cyansnbrst/merch-service/internal/models"
	"cyansnbrst/merch-service/pkg/db"
)

var ErrIncorrectPassword = errors.New("incorrect password")

// Auth usecase struct
type authUC struct {
	cfg      *config.Config
	authRepo auth.Repository
}

// Auth usecase constructor
func NewAuthUseCase(cfg *config.Config, authRepo auth.Repository) auth.UseCase {
	return &authUC{
		cfg:      cfg,
		authRepo: authRepo,
	}
}

// Login or register user
func (u *authUC) LoginOrRegister(ctx context.Context, username, password string) (string, error) {
	user, err := u.authRepo.GetUserByUsername(ctx, username)
	if err != nil && !errors.Is(err, db.ErrUserNotFound) {
		return "", err
	}

	if user == nil {
		hashedPassword, err := argon2id.CreateHash(password, argon2id.DefaultParams)
		if err != nil {
			return "", err
		}

		user, err = u.authRepo.CreateUser(ctx, username, string(hashedPassword))
		if err != nil {
			return "", err
		}
	} else {
		match, err := argon2id.ComparePasswordAndHash(password, user.PasswordHash)
		if err != nil {
			return "", err
		}
		if !match {
			return "", ErrIncorrectPassword
		}
	}

	token, err := u.GenerateJWT(user)
	if err != nil {
		return "", err
	}

	return token, nil
}

// Generate JWT token
func (u *authUC) GenerateJWT(user *models.User) (string, error) {
	expirationTime := jwt.NewNumericDate(time.Now().Add(u.cfg.App.JWTTokenTTL))
	claims := jwt.MapClaims{
		"user_id": user.ID,
		"exp":     expirationTime,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signedToken, err := token.SignedString([]byte(u.cfg.App.JWTSecretKey))
	if err != nil {
		return "", fmt.Errorf("uc - failed to sign token: %w", err)
	}

	return signedToken, nil
}
