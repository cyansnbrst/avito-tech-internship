package merch

import (
	"context"

	m "cyansnbrst/merch-service/internal/models"
)

// Merch Redis repository interface
type RedisRepository interface {
	GetInfo(ctx context.Context, key string) (*m.InfoResponse, error)
	SetInfo(ctx context.Context, key string, info *m.InfoResponse) error
	DeleteInfo(ctx context.Context, key string) error
}
