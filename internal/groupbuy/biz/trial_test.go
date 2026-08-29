package biz

import (
	"context"
	"testing"
)

func TestTrialMarket_ZJ(t *testing.T) {
	activityRepo := newMockGBActivityRepo()
	activityRepo.activities["act_001"] = &GroupBuyActivity{ActivityID: "act_001", DiscountID: "disc_001", ActivityState: 1}
	activityRepo.discounts["disc_001"] = &GroupBuyDiscount{MarketPlan: "ZJ", MarketExpr: "100"}

	trialSvc := NewTrialService(activityRepo)
	result, err := trialSvc.TrialMarket(context.Background(), "act_001", 500)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if result.MarketPayAmount != 400 {
		t.Errorf("expected pay 400, got %d", result.MarketPayAmount)
	}
	if result.MarketDiscountAmt != 100 {
		t.Errorf("expected discount 100, got %d", result.MarketDiscountAmt)
	}
}

func TestTrialMarket_ZK(t *testing.T) {
	activityRepo := newMockGBActivityRepo()
	activityRepo.activities["act_002"] = &GroupBuyActivity{ActivityID: "act_002", DiscountID: "disc_002", ActivityState: 1}
	activityRepo.discounts["disc_002"] = &GroupBuyDiscount{MarketPlan: "ZK", MarketExpr: "8.0"}

	trialSvc := NewTrialService(activityRepo)
	result, err := trialSvc.TrialMarket(context.Background(), "act_002", 1000)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if result.MarketPayAmount != 800 {
		t.Errorf("expected pay 800, got %d", result.MarketPayAmount)
	}
}

func TestTrialMarket_N(t *testing.T) {
	activityRepo := newMockGBActivityRepo()
	activityRepo.activities["act_003"] = &GroupBuyActivity{ActivityID: "act_003", DiscountID: "disc_003", ActivityState: 1}
	activityRepo.discounts["disc_003"] = &GroupBuyDiscount{MarketPlan: "N", MarketExpr: "99"}

	trialSvc := NewTrialService(activityRepo)
	result, err := trialSvc.TrialMarket(context.Background(), "act_003", 500)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if result.MarketPayAmount != 99 {
		t.Errorf("expected pay 99, got %d", result.MarketPayAmount)
	}
}

func TestTrialMarket_ActivityNotFound(t *testing.T) {
	activityRepo := newMockGBActivityRepo()
	trialSvc := NewTrialService(activityRepo)

	result, err := trialSvc.TrialMarket(context.Background(), "non_existent", 500)

	if err == nil {
		t.Error("expected error, got nil")
	}
	if result != nil {
		t.Error("result should be nil on error")
	}
}

func TestLockOrder_Success(t *testing.T) {
	activityRepo := newMockGBActivityRepo()
	orderRepo := newMockGBOrderRepo()
	teamRepo := newMockTeamRepo()
	redisRepo := newMockGBRedisRepo()

	activityRepo.activities["act_010"] = &GroupBuyActivity{
		ActivityID:    "act_010",
		TargetCount:   2,
		ActivityState: 1,
	}

	lockSvc := NewLockService(activityRepo, orderRepo, teamRepo, redisRepo)
	order, err := lockSvc.LockOrder(context.Background(), "act_010", 1001, "ch", "src")

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if order == nil {
		t.Fatal("order should not be nil")
	}
	if order.UserID != 1001 {
		t.Errorf("expected user 1001, got %d", order.UserID)
	}
}

func TestSettlement_Success(t *testing.T) {
	teamRepo := newMockTeamRepo()
	orderRepo := newMockGBOrderRepo()
	notifyRepo := newMockNotifyTaskRepo()
	mqRepo := newMockGBMQRepo()
	notifySvc := NewNotifyService(notifyRepo, mqRepo)

	teamRepo.teams["team_001"] = &GroupBuyTeam{
		TeamID:        "team_001",
		TargetCount:   2,
		CompleteCount: 1,
		TeamState:     0,
	}

	settlementSvc := NewSettlementService(teamRepo, orderRepo, notifySvc, &mockGBStockClient{})
	team, err := settlementSvc.Settlement(context.Background(), "team_001")

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if team == nil {
		t.Fatal("team should not be nil")
	}
	if team.CompleteCount != 2 {
		t.Errorf("expected complete 2, got %d", team.CompleteCount)
	}
	if team.TeamState != 1 {
		t.Errorf("expected state 1 (success), got %d", team.TeamState)
	}

	if len(notifyRepo.tasks) != 1 {
		t.Errorf("expected 1 notify task, got %d", len(notifyRepo.tasks))
	}
}

func TestRefund_Success(t *testing.T) {
	orderRepo := newMockGBOrderRepo()
	notifyRepo := newMockNotifyTaskRepo()
	mqRepo := newMockGBMQRepo()
	notifySvc := NewNotifyService(notifyRepo, mqRepo)

	orderRepo.orders["order_001"] = &GroupBuyOrder{
		OrderID:    "order_001",
		OrderState: 1,
	}

	refundSvc := NewRefundService(orderRepo, notifySvc, &mockGBStockClient{})
	order, err := refundSvc.Refund(context.Background(), "order_001")

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if order.OrderState != 2 {
		t.Errorf("expected state 2 (cancelled), got %d", order.OrderState)
	}

	if len(notifyRepo.tasks) != 1 {
		t.Errorf("expected 1 notify task, got %d", len(notifyRepo.tasks))
	}
}

func TestNotifyService_ProcessPendingTasks(t *testing.T) {
	notifyRepo := newMockNotifyTaskRepo()
	mqRepo := newMockGBMQRepo()
	notifySvc := NewNotifyService(notifyRepo, mqRepo)

	notifyRepo.tasks["task_001"] = &NotifyTask{
		TaskID:       "task_001",
		NotifyType:   NotifyTypeMQ,
		NotifyStatus: NotifyStatusInit,
		NotifyData:   `{"team_id":"team_001"}`,
		RetryCount:   0,
		MaxRetry:     3,
	}

	err := notifySvc.ProcessPendingTasks(context.Background())
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if notifyRepo.tasks["task_001"].NotifyStatus != NotifyStatusSuccess {
		t.Errorf("expected task status success, got %d", notifyRepo.tasks["task_001"].NotifyStatus)
	}

	if len(mqRepo.messages) != 1 {
		t.Errorf("expected 1 MQ message, got %d", len(mqRepo.messages))
	}
}
