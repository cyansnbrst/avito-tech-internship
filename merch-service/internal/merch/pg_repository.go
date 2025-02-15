package merch

import (
	"context"

	m "cyansnbrst/merch-service/internal/models"
)

// Merch repository interface
type Repository interface {
	GetCoinsAndInventory(ctx context.Context, userID int64) (*m.CoinsInventory, error)
	GetTransactionHistory(ctx context.Context, userID int64) (*m.TransactionHistory, error)
	SendCoins(ctx context.Context, fromUser int64, toUser string, amount int64) error
	BuyItem(ctx context.Context, userID int64, itemName string) error
	GetUserIDByUsername(ctx context.Context, username string) (int64, error)
}
