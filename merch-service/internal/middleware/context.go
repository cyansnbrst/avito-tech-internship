package middleware

import (
	"errors"

	"github.com/labstack/echo/v4"
)

const UserContextKey = "user_id"

// Set user ID to the context
func ContextSetUserID(c echo.Context, userID int64) {
	c.Set(UserContextKey, userID)
}

// Get user ID from the context
func ContextGetUserID(c echo.Context) (int64, error) {
	userID, ok := c.Get(UserContextKey).(int64)
	if !ok || userID < 0 {
		return 0, errors.New("incorrect user id")
	}
	return userID, nil
}
