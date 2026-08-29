package biz

import "context"

type ActivityRepo interface {
	GetActivity(ctx context.Context, activityID string) (*LotteryActivity, error)
}

type StrategyRepo interface {
	GetStrategy(ctx context.Context, strategyID string) (*LotteryStrategy, error)
	GetStrategyAwards(ctx context.Context, strategyID string) ([]*StrategyAward, error)
	// DeductAwardStock 原子扣减某奖品的剩余库存，库存不足（<=0）时返回 false，防止超发。
	DeductAwardStock(ctx context.Context, awardID string) (bool, error)
	// RestoreAwardStock 回补奖品库存（下单失败补偿用）。
	RestoreAwardStock(ctx context.Context, awardID string) error
}

type OrderRepo interface {
	CreateOrder(ctx context.Context, order *LotteryOrder) error
	GetUserActivityCount(ctx context.Context, userID int64, activityID string) (int32, error)
}
