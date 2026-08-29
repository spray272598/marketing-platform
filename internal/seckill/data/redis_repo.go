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

func NewRedisRepo(data *Data) biz.RedisRepo {
	return &redisRepo{client: data.rdb}
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

// DecrStockWithUserCheck 原子操作：校验用户限购 + 检查库存 + 扣减库存 + 记录用户购买次数
// 返回值: 1=成功, 0=库存不足, 2=超过限购(limit)
func (r *redisRepo) DecrStockWithUserCheck(ctx context.Context, activityID string, userID int64, limit int32) (int64, error) {
	stockKey := fmt.Sprintf("seckill:stock:%s", activityID)
	userKey := fmt.Sprintf("seckill:user:%s", activityID)
	userIDStr := strconv.FormatInt(userID, 10)
	limitStr := strconv.FormatInt(int64(limit), 10)

	script := redis.NewScript(`
		local stock_key = KEYS[1]
		local user_key = KEYS[2]
		local user_id = ARGV[1]
		local limit = tonumber(ARGV[2])

		-- 1. 校验用户限购次数（limit<=0 视为不限，但通常为 1）
		local cnt = tonumber(redis.call('GET', user_key) or '0')
		if limit > 0 and cnt >= limit then
			return 2
		end

		-- 2. 检查库存
		local stock = tonumber(redis.call('GET', stock_key))
		if stock == nil or stock <= 0 then
			return 0
		end

		-- 3. 原子扣减库存
		redis.call('DECR', stock_key)

		-- 4. 记录用户购买次数（设置过期时间防止无限增长）
		redis.call('INCR', user_key)
		redis.call('EXPIRE', user_key, 86400)

		return 1
	`)

	result, err := script.Run(ctx, r.client,
		[]string{stockKey, userKey},
		userIDStr, limitStr,
	).Int64()

	if err != nil {
		return -1, err
	}
	return result, nil
}

func (r *redisRepo) IncrStock(ctx context.Context, activityID string, count int32) error {
	key := fmt.Sprintf("seckill:stock:%s", activityID)
	return r.client.IncrBy(ctx, key, int64(count)).Err()
}