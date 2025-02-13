package http

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"

	"cyansnbrst/merch-service/internal/auth"
	"cyansnbrst/merch-service/internal/auth/usecase"
	m "cyansnbrst/merch-service/internal/models"
	hh "cyansnbrst/merch-service/pkg/http_helpers"
)

// Auth handlers struct
type authHandlers struct {
	authUC auth.UseCase
	logger *zap.Logger
}

// Auth handlers constructor
func NewAuthHandlers(authUC auth.UseCase, logger *zap.Logger) auth.Handlers {
	return &authHandlers{
		authUC: authUC,
		logger: logger,
	}
}

// Authenticate user
func (h *authHandlers) Authenticate(c echo.Context) error {
	var input m.AuthRequest
	if err := c.Bind(&input); err != nil {
		return hh.BadRequestResponse(c, err)
	}

	if err := c.Validate(input); err != nil {
		return hh.BadRequestResponse(c, err)
	}

	token, err := h.authUC.LoginOrRegister(c.Request().Context(), input.Username, input.Password)
	if err != nil {
		if errors.Is(err, usecase.ErrIncorrectPassword) {
			return hh.InvalidCredentialsResponse(c)
		}
		return hh.ServerErrorResponse(c, h.logger, err)
	}

	return c.JSON(http.StatusOK, m.AuthResponse{Token: token})
}
