package redis

import (
	"github.com/go-redis/redis/v8"

	"cyansnbrst/merch-service/config"
)

func NewRedisClient(cfg *config.Config) *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:         cfg.Redis.Addr,
		MinIdleConns: cfg.Redis.MinIdleConns,
		PoolSize:     cfg.Redis.PoolSize,
		PoolTimeout:  cfg.Redis.PoolTimeout,
		DB:           cfg.Redis.DB,
	})

	return client
}
