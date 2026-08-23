package biz

import "context"

type ActivityRepo interface {
	GetActivity(ctx context.Context, activityID string) (*SeckillActivity, error)
	UpdateActivityStock(ctx context.Context, activityID string, stock int32) error
}

type OrderRepo interface {
	CreateOrder(ctx context.Context, order *SeckillOrder) error
	GetOrder(ctx context.Context, orderID string) (*SeckillOrder, error)
	GetUserActivityOrder(ctx context.Context, userID int64, activityID string) (*SeckillOrder, error)
}

type RedisRepo interface {
	GetStock(ctx context.Context, activityID string) (int32, error)
	DecrStock(ctx context.Context, activityID string) (bool, error)
	SetStock(ctx context.Context, activityID string, stock int32) error
	// DecrStockWithUserCheck 原子操作：检查用户是否已下单 + 检查库存 + 扣减库存 + 标记用户
	// 返回值: 1=成功, 0=库存不足, 2=用户已下单
	DecrStockWithUserCheck(ctx context.Context, activityID string, userID int64) (int64, error)
}

type MQRepo interface {
	PublishOrderMessage(ctx context.Context, order *SeckillOrder) error
}