package biz

import "context"

type ActivityRepo interface {
	GetActivity(ctx context.Context, activityID string) (*LotteryActivity, error)
}

type StrategyRepo interface {
	GetStrategy(ctx context.Context, strategyID string) (*LotteryStrategy, error)
	GetStrategyAwards(ctx context.Context, strategyID string) ([]*StrategyAward, error)
}

type OrderRepo interface {
	CreateOrder(ctx context.Context, order *LotteryOrder) error
	GetUserActivityCount(ctx context.Context, userID int64, activityID string) (int32, error)
}
