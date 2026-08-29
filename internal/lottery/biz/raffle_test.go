package biz

import (
	"context"
	"testing"
)

func TestRaffle_Success(t *testing.T) {
	activityRepo := newMockLTActivityRepo()
	strategyRepo := newMockStrategyRepo()
	orderRepo := newMockLTOrderRepo()

	activityRepo.activities["act_001"] = &LotteryActivity{
		ActivityID:    "act_001",
		StrategyID:    "str_001",
		ActivityState: 1,
	}

	strategyRepo.strategies["str_001"] = &LotteryStrategy{
		StrategyID: "str_001",
	}

	strategyRepo.awards["str_001"] = []*StrategyAward{
		{AwardID: "award_001", AwardName: "一等奖", AwardRate: 0.1, AwardCount: 100},
		{AwardID: "award_002", AwardName: "二等奖", AwardRate: 0.3, AwardCount: 100},
		{AwardID: "award_003", AwardName: "三等奖", AwardRate: 0.6, AwardCount: 100},
	}

	raffleSvc := NewRaffleService(activityRepo, strategyRepo, orderRepo)
	result, err := raffleSvc.Raffle(context.Background(), "act_001", 1001)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("result should not be nil")
	}
	if result.AwardID == "" {
		t.Error("award_id should not be empty")
	}
	if result.AwardState != 1 {
		t.Errorf("expected award_state 1, got %d", result.AwardState)
	}
}

func TestRaffle_ActivityNotFound(t *testing.T) {
	activityRepo := newMockLTActivityRepo()
	strategyRepo := newMockStrategyRepo()
	orderRepo := newMockLTOrderRepo()

	raffleSvc := NewRaffleService(activityRepo, strategyRepo, orderRepo)
	result, err := raffleSvc.Raffle(context.Background(), "non_existent", 1001)

	if err == nil {
		t.Error("expected error, got nil")
	}
	if result != nil {
		t.Error("result should be nil on error")
	}
}

func TestRaffle_DrawLimitOnce(t *testing.T) {
	activityRepo := newMockLTActivityRepo()
	strategyRepo := newMockStrategyRepo()
	orderRepo := newMockLTOrderRepo()

	activityRepo.activities["act_002"] = &LotteryActivity{
		ActivityID:    "act_002",
		StrategyID:    "str_002",
		ActivityState: 1,
	}

	strategyRepo.strategies["str_002"] = &LotteryStrategy{
		StrategyID: "str_002",
	}

	strategyRepo.awards["str_002"] = []*StrategyAward{
		{AwardID: "award_010", AwardName: "大奖", AwardRate: 0.01, AwardCount: 100},
		{AwardID: "award_011", AwardName: "小奖", AwardRate: 0.99, AwardCount: 100},
	}

	raffleSvc := NewRaffleService(activityRepo, strategyRepo, orderRepo)

	// 第一次抽奖应成功
	if _, err := raffleSvc.Raffle(context.Background(), "act_002", 1001); err != nil {
		t.Fatalf("first draw unexpected error: %v", err)
	}

	// 同一用户第二次抽奖应被限流
	if _, err := raffleSvc.Raffle(context.Background(), "act_002", 1001); err == nil {
		t.Error("expected draw-limit error on second draw, got nil")
	}

	// 验证该用户仅产生 1 条抽奖记录
	count, _ := orderRepo.GetUserActivityCount(context.Background(), 1001, "act_002")
	if count != 1 {
		t.Errorf("expected 1 order, got %d", count)
	}
}

func TestRaffle_DifferentUsers(t *testing.T) {
	activityRepo := newMockLTActivityRepo()
	strategyRepo := newMockStrategyRepo()
	orderRepo := newMockLTOrderRepo()

	activityRepo.activities["act_003"] = &LotteryActivity{
		ActivityID:    "act_003",
		StrategyID:    "str_003",
		ActivityState: 1,
	}

	strategyRepo.strategies["str_003"] = &LotteryStrategy{
		StrategyID: "str_003",
	}

	strategyRepo.awards["str_003"] = []*StrategyAward{
		{AwardID: "award_020", AwardName: "奖品", AwardRate: 1.0},
	}

	raffleSvc := NewRaffleService(activityRepo, strategyRepo, orderRepo)

	// Draw for different users
	for userID := int64(1000); userID < 1010; userID++ {
		_, err := raffleSvc.Raffle(context.Background(), "act_003", userID)
		if err != nil {
			t.Errorf("user %d unexpected error: %v", userID, err)
		}
	}

	// Verify each user has 1 order
	for userID := int64(1000); userID < 1010; userID++ {
		count, _ := orderRepo.GetUserActivityCount(context.Background(), userID, "act_003")
		if count != 1 {
			t.Errorf("user %d expected 1 order, got %d", userID, count)
		}
	}
}
