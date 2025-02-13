package http

import (
	"github.com/labstack/echo/v4"

	"cyansnbrst/merch-service/internal/merch"
)

// Register merch routes
func RegisterMerchRoutes(g *echo.Group, h merch.Handlers) {
	g.GET("/info", h.GetInfo)
	g.POST("/sendCoin", h.SendCoins)
	g.GET("/buy/:item", h.BuyItem)
}
