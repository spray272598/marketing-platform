package service

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/marketing-platform/internal/groupbuy/biz"
)

type mockGBActivityRepo struct {
	activities map[string]*biz.GroupBuyActivity
	discounts  map[string]*biz.GroupBuyDiscount
}

func (m *mockGBActivityRepo) GetActivity(ctx context.Context, activityID string) (*biz.GroupBuyActivity, error) {
	if a, ok := m.activities[activityID]; ok {
		return a, nil
	}
	return nil, nil
}

func (m *mockGBActivityRepo) GetDiscount(ctx context.Context, discountID string) (*biz.GroupBuyDiscount, error) {
	if d, ok := m.discounts[discountID]; ok {
		return d, nil
	}
	return nil, nil
}

type mockGBOrderRepo struct {
	orders map[string]*biz.GroupBuyOrder
}

func (m *mockGBOrderRepo) CreateOrder(ctx context.Context, order *biz.GroupBuyOrder) error {
	m.orders[order.OrderID] = order
	return nil
}

func (m *mockGBOrderRepo) GetOrder(ctx context.Context, orderID string) (*biz.GroupBuyOrder, error) {
	if o, ok := m.orders[orderID]; ok {
		return o, nil
	}
	return nil, nil
}

func (m *mockGBOrderRepo) UpdateOrderState(ctx context.Context, orderID string, state int32) error {
	if o, ok := m.orders[orderID]; ok {
		o.OrderState = state
	}
	return nil
}

type mockGBTeamRepo struct {
	teams map[string]*biz.GroupBuyTeam
}

func (m *mockGBTeamRepo) CreateTeam(ctx context.Context, team *biz.GroupBuyTeam) error {
	m.teams[team.TeamID] = team
	return nil
}

func (m *mockGBTeamRepo) GetTeam(ctx context.Context, teamID string) (*biz.GroupBuyTeam, error) {
	if t, ok := m.teams[teamID]; ok {
		return t, nil
	}
	return nil, nil
}

func (m *mockGBTeamRepo) IncrementTeamComplete(ctx context.Context, teamID string) (int32, error) {
	if t, ok := m.teams[teamID]; ok {
		t.CompleteCount++
		return t.CompleteCount, nil
	}
	return 0, nil
}

type mockGBRedisRepo struct{}

func (m *mockGBRedisRepo) LockOrder(ctx context.Context, orderKey string) (bool, error) {
	return true, nil
}

func (m *mockGBRedisRepo) UnlockOrder(ctx context.Context, orderKey string) error {
	return nil
}

type mockGBMQRepo struct{}

func (m *mockGBMQRepo) PublishTeamSuccessMessage(ctx context.Context, teamID string) error {
	return nil
}

func (m *mockGBMQRepo) PublishRefundMessage(ctx context.Context, orderID string) error {
	return nil
}

type mockGBNotifyTaskRepo struct {
	tasks map[string]*biz.NotifyTask
}

func (m *mockGBNotifyTaskRepo) CreateTask(ctx context.Context, task *biz.NotifyTask) error {
	m.tasks[task.TaskID] = task
	return nil
}

func (m *mockGBNotifyTaskRepo) GetTask(ctx context.Context, taskID string) (*biz.NotifyTask, error) {
	if t, ok := m.tasks[taskID]; ok {
		return t, nil
	}
	return nil, nil
}

func (m *mockGBNotifyTaskRepo) GetPendingTasks(ctx context.Context, limit int) ([]*biz.NotifyTask, error) {
	return nil, nil
}

func (m *mockGBNotifyTaskRepo) UpdateTaskStatus(ctx context.Context, taskID string, status int32) error {
	if t, ok := m.tasks[taskID]; ok {
		t.NotifyStatus = status
	}
	return nil
}

func (m *mockGBNotifyTaskRepo) UpdateTaskRetry(ctx context.Context, taskID string, retryCount int32, nextTime int64) error {
	if t, ok := m.tasks[taskID]; ok {
		t.RetryCount = retryCount
		t.NextTime = nextTime
	}
	return nil
}

func (m *mockGBNotifyTaskRepo) GetTaskByUUID(ctx context.Context, uuid string) (*biz.NotifyTask, error) {
	return nil, nil
}

type mockGBStockClient struct{}

func (m *mockGBStockClient) DeductStock(ctx context.Context, stockKey string, count int32) error {
	return nil
}

func (m *mockGBStockClient) RestoreStock(ctx context.Context, stockKey string, count int32) error {
	return nil
}

func setupGBTestService() *GroupBuyService {
	activityRepo := &mockGBActivityRepo{
		activities: map[string]*biz.GroupBuyActivity{
			"act_001": {
				ActivityID:   "act_001",
				ActivityName: "test",
				DiscountID:   "disc_001",
				TargetCount:  2,
			},
		},
		discounts: map[string]*biz.GroupBuyDiscount{
			"disc_001": {DiscountID: "disc_001", MarketPlan: "ZJ", MarketExpr: "10"},
		},
	}
	orderRepo := &mockGBOrderRepo{orders: make(map[string]*biz.GroupBuyOrder)}
	teamRepo := &mockGBTeamRepo{teams: make(map[string]*biz.GroupBuyTeam)}
	notifyTaskRepo := &mockGBNotifyTaskRepo{tasks: make(map[string]*biz.NotifyTask)}
	stockClient := &mockGBStockClient{}
	mqRepo := &mockGBMQRepo{}

	notifySvc := biz.NewNotifyService(notifyTaskRepo, mqRepo)
	lockSvc := biz.NewLockService(activityRepo, orderRepo, teamRepo, &mockGBRedisRepo{})
	trialSvc := biz.NewTrialService(activityRepo)
	settlementSvc := biz.NewSettlementService(teamRepo, orderRepo, notifySvc, stockClient)
	refundSvc := biz.NewRefundService(orderRepo, notifySvc, stockClient)

	return NewGroupBuyService(trialSvc, lockSvc, settlementSvc, refundSvc)
}

func TestTrialGroupBuyMarketHTTP_Success(t *testing.T) {
	svc := setupGBTestService()
	body := `{"activity_id":"act_001","user_id":1001,"market_original_price":100}`
	req := httptest.NewRequest("POST", "/api/v1/groupbuy/trial", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	svc.TrialGroupBuyMarketHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["code"] != "0000" {
		t.Errorf("expected code 0000, got %v", resp["code"])
	}
}

func TestTrialGroupBuyMarketHTTP_InvalidJSON(t *testing.T) {
	svc := setupGBTestService()
	req := httptest.NewRequest("POST", "/api/v1/groupbuy/trial", bytes.NewBufferString("invalid"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	svc.TrialGroupBuyMarketHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["code"] != "400" {
		t.Errorf("expected code 400, got %v", resp["code"])
	}
}

func TestLockMarketPayOrderHTTP_Success(t *testing.T) {
	svc := setupGBTestService()
	body := `{"activity_id":"act_001","user_id":1001,"channel":"app","source":"homepage"}`
	req := httptest.NewRequest("POST", "/api/v1/groupbuy/order/lock", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	svc.LockMarketPayOrderHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["code"] != "0000" {
		t.Errorf("expected code 0000, got %v", resp["code"])
	}
}

func TestSettlementMarketPayOrderHTTP_Success(t *testing.T) {
	orderRepo := &mockGBOrderRepo{orders: make(map[string]*biz.GroupBuyOrder)}
	teamRepo := &mockGBTeamRepo{teams: map[string]*biz.GroupBuyTeam{
		"team_001": {TeamID: "team_001", ActivityID: "act_001", TargetCount: 2, CompleteCount: 0},
	}}
	notifyTaskRepo := &mockGBNotifyTaskRepo{tasks: make(map[string]*biz.NotifyTask)}
	stockClient := &mockGBStockClient{}
	mqRepo := &mockGBMQRepo{}
	notifySvc := biz.NewNotifyService(notifyTaskRepo, mqRepo)
	settlementSvc := biz.NewSettlementService(teamRepo, orderRepo, notifySvc, stockClient)

	svc := &GroupBuyService{settlementSvc: settlementSvc}
	body := `{"team_id":"team_001"}`
	req := httptest.NewRequest("POST", "/api/v1/groupbuy/order/settlement", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	svc.SettlementMarketPayOrderHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestRefundMarketPayOrderHTTP_Success(t *testing.T) {
	orderRepo := &mockGBOrderRepo{orders: map[string]*biz.GroupBuyOrder{
		"order_001": {OrderID: "order_001", TeamID: "team_001", OrderState: 1},
	}}
	notifyTaskRepo := &mockGBNotifyTaskRepo{tasks: make(map[string]*biz.NotifyTask)}
	stockClient := &mockGBStockClient{}
	mqRepo := &mockGBMQRepo{}
	notifySvc := biz.NewNotifyService(notifyTaskRepo, mqRepo)
	refundSvc := biz.NewRefundService(orderRepo, notifySvc, stockClient)

	svc := &GroupBuyService{refundSvc: refundSvc}
	body := `{"order_id":"order_001"}`
	req := httptest.NewRequest("POST", "/api/v1/groupbuy/order/refund", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	svc.RefundMarketPayOrderHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}
