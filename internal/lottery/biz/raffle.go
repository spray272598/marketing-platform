package biz

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/marketing-platform/pkg/common"
)

type RaffleService struct {
	activityRepo ActivityRepo
	strategyRepo StrategyRepo
	orderRepo    OrderRepo
}

func NewRaffleService(
	activityRepo ActivityRepo,
	strategyRepo StrategyRepo,
	orderRepo OrderRepo,
) *RaffleService {
	return &RaffleService{
		activityRepo: activityRepo,
		strategyRepo: strategyRepo,
		orderRepo:    orderRepo,
	}
}

type RaffleResult struct {
	AwardID    string `json:"award_id"`
	AwardName  string `json:"award_name"`
	AwardState int32  `json:"award_state"`
	AwardTime  string `json:"award_time"`
}

func (s *RaffleService) Raffle(ctx context.Context, activityID string, userID int64) (*RaffleResult, error) {
	activity, err := s.activityRepo.GetActivity(ctx, activityID)
	if err != nil {
		return nil, fmt.Errorf(common.LotteryActivityNotExist.Code+": %w", err)
	}

	// 获取策略
	strategy, err := s.strategyRepo.GetStrategy(ctx, activity.StrategyID)
	if err != nil {
		return nil, fmt.Errorf(common.LotteryStrategyNotExist.Code+": %w", err)
	}
	_ = strategy

	// 获取奖品列表
	awards, err := s.strategyRepo.GetStrategyAwards(ctx, activity.StrategyID)
	if err != nil {
		return nil, err
	}

	// 根据概率抽奖
	award := s.drawAward(awards)

	// 创建订单
	order := &LotteryOrder{
		OrderID:    fmt.Sprintf("lt_%d_%d", time.Now().UnixMilli(), userID),
		ActivityID: activityID,
		UserID:     userID,
		AwardID:    award.AwardID,
		AwardState: 1,
		AwardTime:  time.Now().Format("2006-01-02 15:04:05"),
	}

	if err := s.orderRepo.CreateOrder(ctx, order); err != nil {
		return nil, err
	}

	return &RaffleResult{
		AwardID:    award.AwardID,
		AwardName:  award.AwardName,
		AwardState: order.AwardState,
		AwardTime:  order.AwardTime,
	}, nil
}

func (s *RaffleService) drawAward(awards []*StrategyAward) *StrategyAward {
	totalRate := 0.0
	for _, a := range awards {
		totalRate += a.AwardRate
	}

	r := rand.Float64() * totalRate
	cumulative := 0.0
	for _, a := range awards {
		cumulative += a.AwardRate
		if r <= cumulative {
			return a
		}
	}

	return awards[0]
}
