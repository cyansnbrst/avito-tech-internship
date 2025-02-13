package merch

import "github.com/labstack/echo/v4"

// Merch handlers interface
type Handlers interface {
	GetInfo(c echo.Context) error
	SendCoins(c echo.Context) error
	BuyItem(c echo.Context) error
}
