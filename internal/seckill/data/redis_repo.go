package data

import (
	"context"
	"fmt"
	"strconv"

	"github.com/redis/go-redis/v9"
	"github.com/marketing-platform/internal/seckill/biz"
)

type redisRepo struct {
	client *redis.Client
}

func NewRedisRepo(client *redis.Client) biz.RedisRepo {
	return &redisRepo{client: client}
}

func (r *redisRepo) GetStock(ctx context.Context, activityID string) (int32, error) {
	key := fmt.Sprintf("seckill:stock:%s", activityID)
	val, err := r.client.Get(ctx, key).Result()
	if err != nil {
		return 0, err
	}
	stock, _ := strconv.ParseInt(val, 10, 32)
	return int32(stock), nil
}

func (r *redisRepo) DecrStock(ctx context.Context, activityID string) (bool, error) {
	key := fmt.Sprintf("seckill:stock:%s", activityID)

	// Lua脚本: 原子扣减库存
	script := redis.NewScript(`
		local stock = tonumber(redis.call('GET', KEYS[1]))
		if stock == nil then
			return 0
		end
		if stock <= 0 then
			return 0
		end
		redis.call('DECR', KEYS[1])
		return 1
	`)

	result, err := script.Run(ctx, r.client, []string{key}).Int()
	if err != nil {
		return false, err
	}
	return result == 1, nil
}

func (r *redisRepo) SetStock(ctx context.Context, activityID string, stock int32) error {
	key := fmt.Sprintf("seckill:stock:%s", activityID)
	return r.client.Set(ctx, key, stock, 0).Err()
}
