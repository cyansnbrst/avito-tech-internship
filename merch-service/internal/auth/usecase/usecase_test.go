package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/alexedwards/argon2id"
	"github.com/golang-jwt/jwt/v5"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"cyansnbrst/merch-service/config"
	mock_auth "cyansnbrst/merch-service/internal/auth/mock"
	"cyansnbrst/merch-service/internal/models"
	"cyansnbrst/merch-service/pkg/db"
)

var ErrRandomDBError = errors.New("db error")

func TestAuthUC_LoginOrRegister(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock_auth.NewMockRepository(ctrl)
	cfg := &config.Config{
		App: config.App{
			JWTSecretKey: "secret",
			JWTTokenTTL:  time.Hour * 1,
		},
	}

	authUC := NewAuthUseCase(cfg, mockRepo)

	tests := []struct {
		name          string
		username      string
		password      string
		mockSetup     func()
		expectedError error
	}{
		{
			name:     "user found, correct password",
			username: "user",
			password: "password",
			mockSetup: func() {
				hashedPassword, _ := argon2id.CreateHash("password", argon2id.DefaultParams)
				mockRepo.EXPECT().GetUserByUsername(gomock.Any(), "user").Return(&models.User{
					ID:           1,
					Username:     "user",
					PasswordHash: hashedPassword,
				}, nil)
			},
			expectedError: nil,
		},
		{
			name:     "user found, incorrect password",
			username: "user",
			password: "wrong_password",
			mockSetup: func() {
				hashedPassword, _ := argon2id.CreateHash("password", argon2id.DefaultParams)
				mockRepo.EXPECT().GetUserByUsername(gomock.Any(), "user").Return(&models.User{
					ID:           1,
					Username:     "user",
					PasswordHash: hashedPassword,
				}, nil)
			},
			expectedError: ErrIncorrectPassword,
		},
		{
			name:     "user not found, successful registration",
			username: "user",
			password: "password",
			mockSetup: func() {
				mockRepo.EXPECT().GetUserByUsername(gomock.Any(), "user").Return(nil, db.ErrUserNotFound)
				mockRepo.EXPECT().CreateUser(gomock.Any(), "user", gomock.Any()).Return(&models.User{
					ID:       2,
					Username: "user",
				}, nil)
			},
			expectedError: nil,
		},
		{
			name:     "user not found, error creating user",
			username: "user",
			password: "password",
			mockSetup: func() {
				mockRepo.EXPECT().GetUserByUsername(gomock.Any(), "user").Return(nil, db.ErrUserNotFound)
				mockRepo.EXPECT().CreateUser(gomock.Any(), "user", gomock.Any()).Return(nil, ErrRandomDBError)
			},
			expectedError: ErrRandomDBError,
		},
		{
			name:     "error getting user",
			username: "user",
			password: "password",
			mockSetup: func() {
				mockRepo.EXPECT().GetUserByUsername(gomock.Any(), "user").Return(nil, ErrRandomDBError)
			},
			expectedError: ErrRandomDBError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			result, err := authUC.LoginOrRegister(context.Background(), tt.username, tt.password)

			assert.Equal(t, tt.expectedError, err)

			if tt.expectedError == nil {
				assert.NotNil(t, result)
				assert.NotEmpty(t, result)
			}
		})
	}
}

func TestAuthUC_GenerateJWT(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cfg := &config.Config{
		App: config.App{
			JWTSecretKey: "secret",
			JWTTokenTTL:  time.Hour * 1,
		},
	}

	authUC := NewAuthUseCase(cfg, nil)

	user := &models.User{
		ID:       1,
		Username: "user",
	}

	token, err := authUC.GenerateJWT(user)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		return []byte(cfg.App.JWTSecretKey), nil
	})
	assert.NoError(t, err)
	assert.True(t, parsedToken.Valid)
}
