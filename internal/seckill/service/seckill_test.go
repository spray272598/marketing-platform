package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/marketing-platform/internal/seckill/biz"
)

type mockActivityRepo struct {
	activities map[string]*biz.SeckillActivity
}

func (m *mockActivityRepo) GetActivity(ctx context.Context, activityID string) (*biz.SeckillActivity, error) {
	if a, ok := m.activities[activityID]; ok {
		return a, nil
	}
	return nil, fmt.Errorf("activity not found")
}

func (m *mockActivityRepo) UpdateActivityStock(ctx context.Context, activityID string, stock int32) error {
	return nil
}

type mockOrderRepo struct {
	orders map[string]*biz.SeckillOrder
}

func (m *mockOrderRepo) CreateOrder(ctx context.Context, order *biz.SeckillOrder) error {
	m.orders[order.OrderID] = order
	return nil
}

func (m *mockOrderRepo) GetOrder(ctx context.Context, orderID string) (*biz.SeckillOrder, error) {
	if o, ok := m.orders[orderID]; ok {
		return o, nil
	}
	return nil, nil
}

func (m *mockOrderRepo) GetUserActivityOrder(ctx context.Context, userID int64, activityID string) (*biz.SeckillOrder, error) {
	return nil, nil
}

type mockRedisRepo struct{}

func (m *mockRedisRepo) GetStock(ctx context.Context, activityID string) (int32, error) {
	return 100, nil
}

func (m *mockRedisRepo) DecrStock(ctx context.Context, activityID string) (bool, error) {
	return true, nil
}

func (m *mockRedisRepo) SetStock(ctx context.Context, activityID string, stock int32) error {
	return nil
}

type mockMQRepo struct{}

func (m *mockMQRepo) PublishOrderMessage(ctx context.Context, order *biz.SeckillOrder) error {
	return nil
}

type mockStockClient struct{}

func (m *mockStockClient) DeductStock(ctx context.Context, stockKey string, count int32) error {
	return nil
}

func setupTestService() *SeckillService {
	activityRepo := &mockActivityRepo{
		activities: map[string]*biz.SeckillActivity{
			"act_001": {
				ActivityID:   "act_001",
				ActivityName: "测试秒杀活动",
				SkuID:        "sku_001",
				TotalCount:   100,
				LimitCount:   1,
			},
		},
	}
	orderRepo := &mockOrderRepo{orders: make(map[string]*biz.SeckillOrder)}
	redisRepo := &mockRedisRepo{}
	mqRepo := &mockMQRepo{}
	stockClient := &mockStockClient{}

	tradeSvc := biz.NewTradeService(orderRepo, redisRepo, mqRepo, activityRepo, stockClient)
	return NewSeckillService(tradeSvc, activityRepo)
}

func TestQuerySeckillActivityHTTP(t *testing.T) {
	svc := setupTestService()

	req := httptest.NewRequest("GET", "/api/v1/seckill/activity/query?activity_id=act_001", nil)
	w := httptest.NewRecorder()

	svc.QuerySeckillActivityHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["code"] != "0000" {
		t.Errorf("expected code 0000, got %v", resp["code"])
	}
}

func TestQuerySeckillActivityHTTP_NotFound(t *testing.T) {
	svc := setupTestService()

	req := httptest.NewRequest("GET", "/api/v1/seckill/activity/query?activity_id=non_existent", nil)
	w := httptest.NewRecorder()

	svc.QuerySeckillActivityHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["code"] != "404" {
		t.Errorf("expected code 404, got %v", resp["code"])
	}
}

func TestCreateSeckillOrderHTTP_Success(t *testing.T) {
	svc := setupTestService()

	body := `{"activity_id":"act_001","user_id":1001}`
	req := httptest.NewRequest("POST", "/api/v1/seckill/order/create", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	svc.CreateSeckillOrderHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["code"] != "0000" {
		t.Errorf("expected code 0000, got %v", resp["code"])
	}
}

func TestCreateSeckillOrderHTTP_InvalidJSON(t *testing.T) {
	svc := setupTestService()

	req := httptest.NewRequest("POST", "/api/v1/seckill/order/create", bytes.NewBufferString("invalid"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	svc.CreateSeckillOrderHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["code"] != "400" {
		t.Errorf("expected code 400, got %v", resp["code"])
	}
}

func TestQuerySeckillOrderHTTP_Success(t *testing.T) {
	svc := setupTestService()

	req := httptest.NewRequest("GET", "/api/v1/seckill/order/query?order_id=test_order", nil)
	w := httptest.NewRecorder()

	svc.QuerySeckillOrderHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestQuerySeckillOrderHTTP_MissingOrderID(t *testing.T) {
	svc := setupTestService()

	req := httptest.NewRequest("GET", "/api/v1/seckill/order/query", nil)
	w := httptest.NewRecorder()

	svc.QuerySeckillOrderHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["code"] != "400" {
		t.Errorf("expected code 400, got %v", resp["code"])
	}
}
