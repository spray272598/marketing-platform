package data

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/marketing-platform/internal/groupbuy/biz"
)

type redisRepo struct {
	client *redis.Client
}

func NewRedisRepo(client *redis.Client) biz.RedisRepo {
	return &redisRepo{client: client}
}

func (r *redisRepo) LockOrder(ctx context.Context, orderKey string) (bool, error) {
	ok, err := r.client.SetNX(ctx, orderKey, "1", 30*time.Second).Result()
	if err != nil {
		return false, fmt.Errorf("lock failed: %w", err)
	}
	return ok, nil
}

func (r *redisRepo) UnlockOrder(ctx context.Context, orderKey string) error {
	return r.client.Del(ctx, orderKey).Err()
}
