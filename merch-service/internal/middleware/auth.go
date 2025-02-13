package middleware

import (
	"errors"
	"strings"

	"github.com/labstack/echo/v4"

	"cyansnbrst/merch-service/pkg/auth"
	"cyansnbrst/merch-service/pkg/auth/jwt"
	hh "cyansnbrst/merch-service/pkg/http_helpers"
)

// Authentication middleware
func (mw *MiddlewareManager) Authenticate(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		authHeader := c.Request().Header.Get(echo.HeaderAuthorization)
		if authHeader == "" {
			return hh.AuthenticationRequiredResponse(c)
		}

		token := strings.TrimPrefix(authHeader, "Bearer ")
		if token == authHeader {
			return hh.InvalidAuthenticationTokenResponse(c)
		}

		userID, err := jwt.ParseJWT(token, string(mw.cfg.App.JWTSecretKey))
		if err != nil {
			if errors.Is(err, auth.ErrInvalidToken) {
				return hh.InvalidAuthenticationTokenResponse(c)
			}
			return hh.ServerErrorResponse(c, mw.logger, err)
		}

		ContextSetUserID(c, userID)

		return next(c)
	}
}
