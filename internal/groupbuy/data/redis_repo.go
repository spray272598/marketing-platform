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

func NewRedisRepo(data *Data) biz.RedisRepo {
	return &redisRepo{client: data.rdb}
}

func (r *redisRepo) LockOrder(ctx context.Context, orderKey string, lockValue string) (bool, error) {
	ok, err := r.client.SetNX(ctx, orderKey, lockValue, 30*time.Second).Result()
	if err != nil {
		return false, fmt.Errorf("lock failed: %w", err)
	}
	return ok, nil
}

func (r *redisRepo) UnlockOrder(ctx context.Context, orderKey string, lockValue string) error {
	script := redis.NewScript(`
		if redis.call("get", KEYS[1]) == ARGV[1] then
			return redis.call("del", KEYS[1])
		end
		return 0
	`)
	_, err := script.Run(ctx, r.client, []string{orderKey}, lockValue).Result()
	return err
}
