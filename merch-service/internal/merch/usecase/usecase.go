package usecase

import (
	"context"

	"cyansnbrst/merch-service/internal/merch"
	m "cyansnbrst/merch-service/internal/models"
	"cyansnbrst/merch-service/pkg/db/redis"
)

// Merch usecase struct
type merchUC struct {
	merchRepo      merch.Repository
	merchRedisRepo merch.RedisRepository
}

// Merch usecase constructor
func NewMerchUseCase(merchRepo merch.Repository, merchRedisRepo merch.RedisRepository) merch.UseCase {
	return &merchUC{
		merchRepo:      merchRepo,
		merchRedisRepo: merchRedisRepo,
	}
}

// Get user's info
func (u *merchUC) GetInfo(ctx context.Context, userID int64) (*m.InfoResponse, error) {
	key := redis.GetUserInfoCacheKey(userID)

	cachedInfo, err := u.merchRedisRepo.GetInfo(ctx, key)
	if err != nil {
		return nil, err
	}

	if cachedInfo != nil {
		return cachedInfo, nil
	}

	coinsInventory, err := u.merchRepo.GetCoinsAndInventory(ctx, userID)
	if err != nil {
		return nil, err
	}

	coinHistory, err := u.merchRepo.GetTransactionHistory(ctx, userID)
	if err != nil {
		return nil, err
	}

	info := &m.InfoResponse{
		CoinsInventory: m.CoinsInventory{
			Coins:     coinsInventory.Coins,
			Inventory: coinsInventory.Inventory,
		},
		CoinHistory: coinHistory,
	}

	if err := u.merchRedisRepo.SetInfo(ctx, key, info); err != nil {
		return nil, err
	}

	return info, nil
}

// Send coins to other user
func (u *merchUC) SendCoins(ctx context.Context, fromUserID int64, toUser string, amount int64) error {
	if err := u.merchRepo.SendCoins(ctx, fromUserID, toUser, amount); err != nil {
		return err
	}

	fromUserKey := redis.GetUserInfoCacheKey(fromUserID)
	if err := u.merchRedisRepo.DeleteInfo(ctx, fromUserKey); err != nil {

		return err
	}

	toUserID, err := u.merchRepo.GetUserIDByUsername(ctx, toUser)
	if err != nil {
		return err
	}

	toUserKey := redis.GetUserInfoCacheKey(toUserID)
	if err := u.merchRedisRepo.DeleteInfo(ctx, toUserKey); err != nil {
		return err
	}

	return nil
}

// Buy an item
func (u *merchUC) BuyItem(ctx context.Context, userID int64, item string) error {
	if err := u.merchRepo.BuyItem(ctx, userID, item); err != nil {
		return err
	}

	key := redis.GetUserInfoCacheKey(userID)
	if err := u.merchRedisRepo.DeleteInfo(ctx, key); err != nil {
		return err
	}

	return nil
}
