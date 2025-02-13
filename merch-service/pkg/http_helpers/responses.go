package httphelpers

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

const (
	msgServerError                = "the server encountered a problem and could not process your request"
	msgInvalidCredentials         = "invalid authentication credentials"
	msgInvalidAuthenticationToken = "invalid or missing authentication token"
	msgAuthenticationRequired     = "you must be authenticated to access this resource"
)

// Error response
type ErrorResponse struct {
	Errors string `json:"errors"`
}

// Log an error
func logError(c echo.Context, l *zap.Logger, err error) {
	l.Error("an error occurred",
		zap.String("request_method", c.Request().Method),
		zap.String("request_url", c.Request().URL.String()),
		zap.Error(err),
	)
}

// Error response
func errorResponse(c echo.Context, statusCode int, message string) error {
	return c.JSON(statusCode, ErrorResponse{Errors: message})
}

// Server error response (500)
func ServerErrorResponse(c echo.Context, l *zap.Logger, err error) error {
	logError(c, l, err)
	return errorResponse(c, http.StatusInternalServerError, msgServerError)
}

// Bad request response (400)
func BadRequestResponse(c echo.Context, err error) error {
	return errorResponse(c, http.StatusBadRequest, err.Error())
}

// Invalid credentials response (401)
func InvalidCredentialsResponse(c echo.Context) error {
	return errorResponse(c, http.StatusUnauthorized, msgInvalidCredentials)
}

// Invalid auth token response (401)
func InvalidAuthenticationTokenResponse(c echo.Context) error {
	return errorResponse(c, http.StatusUnauthorized, msgInvalidAuthenticationToken)
}

// Authentication required response (401)
func AuthenticationRequiredResponse(c echo.Context) error {
	return errorResponse(c, http.StatusUnauthorized, msgAuthenticationRequired)
}
