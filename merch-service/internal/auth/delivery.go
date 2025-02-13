package auth

import "github.com/labstack/echo/v4"

// Auth handlers interface
type Handlers interface {
	Authenticate(c echo.Context) error
}
