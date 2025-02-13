package http

import (
	"github.com/labstack/echo/v4"

	"cyansnbrst/merch-service/internal/auth"
)

// Register auth routes
func RegisterAuthRoutes(g *echo.Group, h auth.Handlers) {
	g.POST("/auth", h.Authenticate)
}
