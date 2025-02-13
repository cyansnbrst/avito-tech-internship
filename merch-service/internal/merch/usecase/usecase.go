package usecase

import (
	"context"

	"cyansnbrst/merch-service/internal/merch"
	m "cyansnbrst/merch-service/internal/models"
)

// Merch usecase struct
type merchUC struct {
	merchRepo merch.Repository
}

// Merch usecase constructor
func NewMerchUseCase(merchRepo merch.Repository) merch.UseCase {
	return &merchUC{merchRepo: merchRepo}
}

// Get user's info
func (u *merchUC) GetInfo(ctx context.Context, userID int64) (*m.InfoResponse, error) {
	coinsInventory, err := u.merchRepo.GetCoinsAndInventory(ctx, userID)
	if err != nil {
		return nil, err
	}

	coinHistory, err := u.merchRepo.GetTransactionHistory(ctx, userID)
	if err != nil {
		return nil, err
	}

	return &m.InfoResponse{
		CoinsInventory: m.CoinsInventory{
			Coins:     coinsInventory.Coins,
			Inventory: coinsInventory.Inventory,
		},
		CoinHistory: coinHistory,
	}, nil
}

// Send coins to other user
func (u *merchUC) SendCoins(ctx context.Context, fromUserID int64, toUser string, amount int64) error {
	return u.merchRepo.SendCoins(ctx, fromUserID, toUser, amount)
}

// Buy an item
func (u *merchUC) BuyItem(ctx context.Context, userID int64, item string) error {
	return u.merchRepo.BuyItem(ctx, userID, item)
}
