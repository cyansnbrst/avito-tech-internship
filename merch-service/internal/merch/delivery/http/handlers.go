package http

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"

	"cyansnbrst/merch-service/internal/merch"
	"cyansnbrst/merch-service/internal/middleware"
	m "cyansnbrst/merch-service/internal/models"
	"cyansnbrst/merch-service/pkg/db"
	hh "cyansnbrst/merch-service/pkg/http_helpers"
)

// Merch handlers struct
type merchHandlers struct {
	merchUC merch.UseCase
	logger  *zap.Logger
}

// Merch handlers constructor
func NewMerchHandlers(merchUC merch.UseCase, logger *zap.Logger) merch.Handlers {
	return &merchHandlers{
		merchUC: merchUC,
		logger:  logger,
	}
}

// @Summary		Get user's info
// @Description	Get user's balance and transactions history.
// @Tags		merch
// @Produce		json
// @Success		200	{object}	models.InfoResponse			"successful"
// @Failure		400	{object}	httphelpers.ErrorResponse	"bad request"
// @Failure		401	{object}	httphelpers.ErrorResponse	"authentication required"
// @Failure		500	{object}	httphelpers.ErrorResponse	"internal server error"
// @Security 	JWT
// @Router		/info [get]
func (h *merchHandlers) GetInfo(c echo.Context) error {
	userID, err := middleware.ContextGetUserID(c)
	if err != nil {
		return hh.ServerErrorResponse(c, h.logger, err)
	}

	userInfo, err := h.merchUC.GetInfo(c.Request().Context(), userID)
	if err != nil {
		return hh.ServerErrorResponse(c, h.logger, err)
	}

	return c.JSON(http.StatusOK, userInfo)
}

// @Summary		Send coins
// @Description	Send coins to another user
// @Tags		merch
// @Accept 		json
// @Produce		json
// @Param input body models.SendTransaction true "input"
// @Success		200
// @Failure		400	{object}	httphelpers.ErrorResponse	"bad request"
// @Failure		401	{object}	httphelpers.ErrorResponse	"authentication required"
// @Failure		500	{object}	httphelpers.ErrorResponse	"internal server error"
// @Security 	JWT
// @Router		/sendCoin [post]
func (h *merchHandlers) SendCoins(c echo.Context) error {
	userID, err := middleware.ContextGetUserID(c)
	if err != nil {
		return hh.ServerErrorResponse(c, h.logger, err)
	}

	var input m.SendTransaction
	if err := c.Bind(&input); err != nil {
		return hh.BadRequestResponse(c, err)
	}

	if err := c.Validate(input); err != nil {
		return hh.BadRequestResponse(c, err)
	}

	err = h.merchUC.SendCoins(c.Request().Context(), userID, input.ToUser, input.Amount)
	if err != nil {
		if errors.Is(err, db.ErrInsufficientFunds) || errors.Is(err, db.ErrIncorrectReciever) || errors.Is(err, db.ErrUserNotFound) {
			return hh.BadRequestResponse(c, err)
		}
		return hh.ServerErrorResponse(c, h.logger, err)
	}

	return c.NoContent(http.StatusOK)
}

// @Summary		Buy item
// @Description	Buy an item from the store
// @Tags		merch
// @Produce		json
// @Param   	item  path  string  true  "name of the item to buy"
// @Success		200
// @Failure		400	{object}	httphelpers.ErrorResponse	"bad request"
// @Failure		401	{object}	httphelpers.ErrorResponse	"authentication required"
// @Failure		500	{object}	httphelpers.ErrorResponse	"internal server error"
// @Security 	JWT
// @Router		/buy/{item} [get]
func (h *merchHandlers) BuyItem(c echo.Context) error {
	userID, err := middleware.ContextGetUserID(c)
	if err != nil {
		return hh.ServerErrorResponse(c, h.logger, err)
	}

	item := c.Param("item")

	err = h.merchUC.BuyItem(c.Request().Context(), userID, item)
	if err != nil {
		if errors.Is(err, db.ErrItemtNotFound) || errors.Is(err, db.ErrInsufficientFunds) {
			return hh.BadRequestResponse(c, err)
		}
		return hh.ServerErrorResponse(c, h.logger, err)
	}

	return c.NoContent(http.StatusOK)
}
