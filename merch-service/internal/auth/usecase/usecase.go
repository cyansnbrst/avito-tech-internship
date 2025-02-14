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
		user, err = u.createUser(ctx, username, password)
		if err != nil {
			return "", err
		}
	} else {
		if err := u.validatePassword(user, password); err != nil {
			return "", err
		}
	}

	return u.GenerateJWT(user)
}

// Create a new user
func (u *authUC) createUser(ctx context.Context, username, password string) (*models.User, error) {
	hashedPassword, err := argon2id.CreateHash(password, argon2id.DefaultParams)
	if err != nil {
		return nil, err
	}

	return u.authRepo.CreateUser(ctx, username, hashedPassword)
}

// Validate password
func (u *authUC) validatePassword(user *models.User, password string) error {
	match, err := argon2id.ComparePasswordAndHash(password, user.PasswordHash)
	if err != nil {
		return err
	}
	if !match {
		return ErrIncorrectPassword
	}
	return nil
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
