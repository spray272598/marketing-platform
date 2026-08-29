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

	if activity.ActivityState != common.ActivityStateOpen {
		return nil, fmt.Errorf("%s: %s", common.LotteryActivityNotExist.Code, "activity is not open")
	}

	// 策略必须存在，否则视为活动配置错误（原先获取后直接丢弃，规则被忽略）。
	if _, err := s.strategyRepo.GetStrategy(ctx, activity.StrategyID); err != nil {
		return nil, fmt.Errorf("%s: %w", common.LotteryStrategyNotExist.Code, err)
	}

	drawCount, err := s.orderRepo.GetUserActivityCount(ctx, userID, activityID)
	if err != nil {
		return nil, fmt.Errorf("get draw count failed: %w", err)
	}
	if drawCount > 0 {
		return nil, fmt.Errorf("%s: %s", common.LotteryDrawLimit.Code, common.LotteryDrawLimit.Info)
	}

	awards, err := s.strategyRepo.GetStrategyAwards(ctx, activity.StrategyID)
	if err != nil {
		return nil, err
	}
	if len(awards) == 0 {
		return nil, fmt.Errorf("%s: no awards configured", common.LotteryAwardNotFound.Code)
	}

	// 只在仍有库存的奖品中抽奖，避免抽中已售罄的奖品。
	available := make([]*StrategyAward, 0, len(awards))
	for _, a := range awards {
		if a.AwardCount > 0 {
			available = append(available, a)
		}
	}

	won := false
	var award *StrategyAward
	if len(available) > 0 {
		award = s.drawAward(available)
		// 原子扣减奖品库存；若被并发抢光则返回 false，本次视为未中奖。
		ok, err := s.strategyRepo.DeductAwardStock(ctx, award.AwardID)
		if err != nil {
			return nil, fmt.Errorf("deduct award stock failed: %w", err)
		}
		won = ok
	}

	id, err := s.orderRepo.NextOrderID(ctx, "order")
	if err != nil {
		if won {
			// 补偿失败必须留痕，否则奖品库存会凭空减少且无人知晓。
			compensate("restore_award_stock", s.strategyRepo.RestoreAwardStock(ctx, award.AwardID))
		}
		return nil, fmt.Errorf("generate order id failed: %w", err)
	}
	order := &LotteryOrder{
		OrderID:    fmt.Sprintf("lt_%019d", id),
		ActivityID: activityID,
		UserID:     userID,
		AwardState: 0,
		AwardTime:  time.Now().Format("2006-01-02 15:04:05"),
	}
	if won {
		order.AwardID = award.AwardID
		order.AwardState = 1
	}

	if err := s.orderRepo.CreateOrder(ctx, order); err != nil {
		if won {
			compensate("restore_award_stock", s.strategyRepo.RestoreAwardStock(ctx, award.AwardID))
		}
		return nil, fmt.Errorf("create order failed: %w", err)
	}

	result := &RaffleResult{
		AwardID:    order.AwardID,
		AwardName:  "",
		AwardState: order.AwardState,
		AwardTime:  order.AwardTime,
	}
	if won {
		result.AwardName = award.AwardName
	}
	return result, nil
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

	return awards[len(awards)-1]
}
