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
}

type MQRepo interface {
	PublishOrderMessage(ctx context.Context, order *SeckillOrder) error
}
