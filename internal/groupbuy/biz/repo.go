package biz

import "context"

type ActivityRepo interface {
	GetActivity(ctx context.Context, activityID string) (*GroupBuyActivity, error)
	GetDiscount(ctx context.Context, discountID string) (*GroupBuyDiscount, error)
}

type OrderRepo interface {
	CreateOrder(ctx context.Context, order *GroupBuyOrder) error
	GetOrder(ctx context.Context, orderID string) (*GroupBuyOrder, error)
	UpdateOrderState(ctx context.Context, orderID string, state int32) error
}

type TeamRepo interface {
	CreateTeam(ctx context.Context, team *GroupBuyTeam) error
	GetTeam(ctx context.Context, teamID string) (*GroupBuyTeam, error)
	IncrementTeamComplete(ctx context.Context, teamID string) (int32, error)
}

type RedisRepo interface {
	LockOrder(ctx context.Context, orderKey string) (bool, error)
	UnlockOrder(ctx context.Context, orderKey string) error
}

type MQRepo interface {
	PublishTeamSuccessMessage(ctx context.Context, teamID string) error
	PublishRefundMessage(ctx context.Context, orderID string) error
}
