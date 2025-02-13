package merch

import (
	"context"

	m "cyansnbrst/merch-service/internal/models"
)

// Merch usecase interface
type UseCase interface {
	GetInfo(ctx context.Context, userID int64) (*m.InfoResponse, error)
	SendCoins(ctx context.Context, fromUserID int64, toUser string, amount int64) error
	BuyItem(ctx context.Context, userID int64, item string) error
}
