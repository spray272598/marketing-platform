package data

import (
	"context"
	"fmt"
	"strconv"

	"github.com/redis/go-redis/v9"
)

type RedisRepo struct {
	client *redis.Client
}

func NewRedisRepo(client *redis.Client) *RedisRepo {
	return &RedisRepo{client: client}
}

func (r *RedisRepo) GetStock(ctx context.Context, activityID string) (int32, error) {
	key := fmt.Sprintf("seckill:stock:%s", activityID)
	val, err := r.client.Get(ctx, key).Result()
	if err != nil {
		return 0, err
	}
	stock, _ := strconv.ParseInt(val, 10, 32)
	return int32(stock), nil
}

func (r *RedisRepo) DecrStock(ctx context.Context, activityID string) (bool, error) {
	key := fmt.Sprintf("seckill:stock:%s", activityID)

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

func (r *RedisRepo) SetStock(ctx context.Context, activityID string, stock int32) error {
	key := fmt.Sprintf("seckill:stock:%s", activityID)
	return r.client.Set(ctx, key, stock, 0).Err()
}
