package repository

import (
	"context"
	"encoding/json"

	"github.com/go-redis/redis/v8"

	"cyansnbrst/merch-service/config"
	"cyansnbrst/merch-service/internal/merch"
	m "cyansnbrst/merch-service/internal/models"
)

// Merch redis repository
type merchRedisRepo struct {
	cfg         *config.Config
	redisClient *redis.Client
}

// Merch repository constructor
func NewMerchRedisRepo(cfg *config.Config, redisClient *redis.Client) merch.RedisRepository {
	return &merchRedisRepo{cfg: cfg,
		redisClient: redisClient,
	}
}

// Get info for user
func (r *merchRedisRepo) GetInfo(ctx context.Context, key string) (*m.InfoResponse, error) {
	infoBytes, err := r.redisClient.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, err
	}

	var info m.InfoResponse
	if err = json.Unmarshal(infoBytes, &info); err != nil {
		return nil, err
	}

	return &info, nil
}

// Cache info for user
func (r *merchRedisRepo) SetInfo(ctx context.Context, key string, info *m.InfoResponse) error {
	infoBytes, err := json.Marshal(info)
	if err != nil {
		return err
	}

	if err = r.redisClient.Set(ctx, key, infoBytes, r.cfg.Redis.CacheTTL).Err(); err != nil {
		return err
	}

	return nil
}

// Delete info for user
func (r *merchRedisRepo) DeleteInfo(ctx context.Context, key string) error {
	if err := r.redisClient.Del(ctx, key).Err(); err != nil {
		return err
	}
	return nil
}
