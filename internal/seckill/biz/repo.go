package biz

import "context"

type ActivityRepo interface {
	GetActivity(ctx context.Context, activityID string) (*SeckillActivity, error)
	UpdateActivityStock(ctx context.Context, activityID string, stock int32) error
}

type OrderRepo interface {
	CreateOrder(ctx context.Context, order *SeckillOrder) error
	GetOrder(ctx context.Context, orderID string) (*SeckillOrder, error)
	UpdateOrderState(ctx context.Context, orderID string, state int32) error
	// NextOrderID 基于号段模式返回当前服务的下一个订单号（单调递增）。
	NextOrderID(ctx context.Context, bizTag string) (int64, error)
}

type RedisRepo interface {
	GetStock(ctx context.Context, activityID string) (int32, error)
	DecrStock(ctx context.Context, activityID string) (bool, error)
	SetStock(ctx context.Context, activityID string, stock int32) error
	DecrStockWithUserCheck(ctx context.Context, activityID string, userID int64, limit int32) (int64, error)
	IncrStock(ctx context.Context, activityID string, count int32) error
}

type MQRepo interface {
	PublishOrderMessage(ctx context.Context, order *SeckillOrder) error
}