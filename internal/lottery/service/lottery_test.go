package service

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/marketing-platform/internal/lottery/biz"
	"github.com/marketing-platform/pkg/auth"
)

type mockLActivityRepo struct {
	activities map[string]*biz.LotteryActivity
}

func (m *mockLActivityRepo) GetActivity(ctx context.Context, activityID string) (*biz.LotteryActivity, error) {
	if a, ok := m.activities[activityID]; ok {
		return a, nil
	}
	return nil, nil
}

type mockLStrategyRepo struct {
	strategies map[string]*biz.LotteryStrategy
	awards     map[string][]*biz.StrategyAward
}

func (m *mockLStrategyRepo) GetStrategy(ctx context.Context, strategyID string) (*biz.LotteryStrategy, error) {
	if s, ok := m.strategies[strategyID]; ok {
		return s, nil
	}
	return nil, nil
}

func (m *mockLStrategyRepo) GetStrategyAwards(ctx context.Context, strategyID string) ([]*biz.StrategyAward, error) {
	if a, ok := m.awards[strategyID]; ok {
		return a, nil
	}
	return nil, nil
}

func (m *mockLStrategyRepo) DeductAwardStock(ctx context.Context, awardID string) (bool, error) {
	for _, list := range m.awards {
		for _, a := range list {
			if a.AwardID == awardID {
				if a.AwardCount > 0 {
					a.AwardCount--
					return true, nil
				}
				return false, nil
			}
		}
	}
	return false, nil
}

func (m *mockLStrategyRepo) RestoreAwardStock(ctx context.Context, awardID string) error {
	for _, list := range m.awards {
		for _, a := range list {
			if a.AwardID == awardID {
				a.AwardCount++
				return nil
			}
		}
	}
	return nil
}

type mockLOrderRepo struct {
	orders map[string]*biz.LotteryOrder
	seq    int64
}

func (m *mockLOrderRepo) NextOrderID(ctx context.Context, bizTag string) (int64, error) {
	m.seq++
	return m.seq, nil
}

func (m *mockLOrderRepo) CreateOrder(ctx context.Context, order *biz.LotteryOrder) error {
	m.orders[order.OrderID] = order
	return nil
}

func (m *mockLOrderRepo) GetUserActivityCount(ctx context.Context, userID int64, activityID string) (int32, error) {
	return 0, nil
}

func setupLTestService() *LotteryService {
	activityRepo := &mockLActivityRepo{
		activities: map[string]*biz.LotteryActivity{
			"act_001": {
				ActivityID:    "act_001",
				ActivityName:  "测试抽奖活动",
				StrategyID:    "str_001",
				ActivityState: 1,
			},
		},
	}
	strategyRepo := &mockLStrategyRepo{
		strategies: map[string]*biz.LotteryStrategy{
			"str_001": {
				StrategyID: "str_001",
			},
		},
		awards: map[string][]*biz.StrategyAward{
			"str_001": {
				{AwardID: "award_001", AwardName: "一等奖", AwardRate: 0.1, AwardCount: 100},
				{AwardID: "award_002", AwardName: "二等奖", AwardRate: 0.9, AwardCount: 100},
			},
		},
	}
	orderRepo := &mockLOrderRepo{orders: make(map[string]*biz.LotteryOrder)}

	raffleSvc := biz.NewRaffleService(activityRepo, strategyRepo, orderRepo)
	return NewLotteryService(raffleSvc)
}

func TestRaffleHTTP_Success(t *testing.T) {
	svc := setupLTestService()

	body := `{"activity_id":"act_001"}`
	req := httptest.NewRequest("POST", "/api/v1/lottery/raffle", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	// 身份由鉴权中间件写入 context，这里模拟一个已认证用户。
	req = req.WithContext(auth.WithUserID(req.Context(), 1001))
	w := httptest.NewRecorder()

	svc.RaffleHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["code"] != "0000" {
		t.Errorf("expected code 0000, got %v", resp["code"])
	}
}

// TestRaffleHTTP_Unauthorized 未认证不得抽奖：请求体自带的 user_id 不再被采信。
func TestRaffleHTTP_Unauthorized(t *testing.T) {
	svc := setupLTestService()

	body := `{"activity_id":"act_001","user_id":1001}`
	req := httptest.NewRequest("POST", "/api/v1/lottery/raffle", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	svc.RaffleHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401 without identity, got %d", w.Code)
	}
}

func TestRaffleHTTP_InvalidJSON(t *testing.T) {
	svc := setupLTestService()

	req := httptest.NewRequest("POST", "/api/v1/lottery/raffle", bytes.NewBufferString("invalid"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	svc.RaffleHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["code"] != "C0001" {
		t.Errorf("expected code C0001, got %v", resp["code"])
	}
}

func TestQueryLotteryActivityHTTP(t *testing.T) {
	svc := setupLTestService()

	req := httptest.NewRequest("GET", "/api/v1/lottery/activity/query?activity_id=act_001", nil)
	w := httptest.NewRecorder()

	svc.QueryLotteryActivityHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["code"] != "0000" {
		t.Errorf("expected code 0000, got %v", resp["code"])
	}
}
