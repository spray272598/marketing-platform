package biz

import (
	"context"
	"testing"
	"time"
)

func validActivity(id, skuID string) *SeckillActivity {
	return &SeckillActivity{
		ActivityID:    id,
		SkuID:         skuID,
		TotalCount:    100,
		ActivityState: 1,
		StartTime:     time.Now().Add(-1 * time.Hour).Format("2006-01-02 15:04:05"),
		EndTime:       time.Now().Add(1 * time.Hour).Format("2006-01-02 15:04:05"),
	}
}

func TestCreateSeckillOrder_Success(t *testing.T) {
	orderRepo := newMockOrderRepo()
	redisRepo := newMockRedisRepo()
	mqRepo := newMockMQRepo()
	activityRepo := newMockActivityRepo()
	stockClient := &mockStockClient{}

	activityRepo.activities["act_001"] = validActivity("act_001", "sku_001")

	tradeSvc := NewTradeService(orderRepo, redisRepo, mqRepo, activityRepo, stockClient)

	activityID := "act_001"
	redisRepo.stocks[activityID] = 10

	order, err := tradeSvc.CreateSeckillOrder(context.Background(), activityID, 1001)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if order == nil {
		t.Fatal("order should not be nil")
	}
	if order.ActivityID != activityID {
		t.Errorf("expected activity_id %s, got %s", activityID, order.ActivityID)
	}
	if order.UserID != 1001 {
		t.Errorf("expected user_id 1001, got %d", order.UserID)
	}
	if order.OrderState != int32(0) {
		t.Errorf("expected order_state 0, got %d", order.OrderState)
	}

	if len(orderRepo.orders) != 1 {
		t.Errorf("expected 1 order, got %d", len(orderRepo.orders))
	}
	if len(mqRepo.messages) != 1 {
		t.Errorf("expected 1 MQ message, got %d", len(mqRepo.messages))
	}
}

func TestCreateSeckillOrder_ActivityNotFound(t *testing.T) {
	orderRepo := newMockOrderRepo()
	redisRepo := newMockRedisRepo()
	mqRepo := newMockMQRepo()
	activityRepo := newMockActivityRepo()
	stockClient := &mockStockClient{}

	tradeSvc := NewTradeService(orderRepo, redisRepo, mqRepo, activityRepo, stockClient)

	_, err := tradeSvc.CreateSeckillOrder(context.Background(), "non_existent", 1001)

	if err == nil {
		t.Error("expected error, got nil")
	}
}

func TestCreateSeckillOrder_StockNotEnough(t *testing.T) {
	orderRepo := newMockOrderRepo()
	redisRepo := newMockRedisRepo()
	mqRepo := newMockMQRepo()
	activityRepo := newMockActivityRepo()
	stockClient := &mockStockClient{}

	activityRepo.activities["act_002"] = &SeckillActivity{
		ActivityID:    "act_002",
		SkuID:         "sku_002",
		TotalCount:    0,
		ActivityState: 1,
		StartTime:     time.Now().Add(-1 * time.Hour).Format("2006-01-02 15:04:05"),
		EndTime:       time.Now().Add(1 * time.Hour).Format("2006-01-02 15:04:05"),
	}

	tradeSvc := NewTradeService(orderRepo, redisRepo, mqRepo, activityRepo, stockClient)

	activityID := "act_002"
	// 库存设置为0
	redisRepo.stocks[activityID] = 0

	order, err := tradeSvc.CreateSeckillOrder(context.Background(), activityID, 1001)

	if err == nil {
		t.Error("expected error for stock not enough, got nil")
	}
	if order != nil {
		t.Error("order should be nil on error")
	}
}

func TestCreateSeckillOrder_DuplicateOrder_Atomic(t *testing.T) {
	orderRepo := newMockOrderRepo()
	redisRepo := newMockRedisRepo()
	mqRepo := newMockMQRepo()
	activityRepo := newMockActivityRepo()
	stockClient := &mockStockClient{}

	activityRepo.activities["act_003"] = &SeckillActivity{
		ActivityID:    "act_003",
		SkuID:         "sku_003",
		TotalCount:    100,
		ActivityState: 1,
		StartTime:     time.Now().Add(-1 * time.Hour).Format("2006-01-02 15:04:05"),
		EndTime:       time.Now().Add(1 * time.Hour).Format("2006-01-02 15:04:05"),
	}

	tradeSvc := NewTradeService(orderRepo, redisRepo, mqRepo, activityRepo, stockClient)

	activityID := "act_003"
	redisRepo.stocks[activityID] = 10

	// 第一次下单成功
	_, err := tradeSvc.CreateSeckillOrder(context.Background(), activityID, 1001)
	if err != nil {
		t.Fatalf("first order failed: %v", err)
	}

	// 第二次同一用户下单应该被原子操作拦截（用户已标记）
	order, err := tradeSvc.CreateSeckillOrder(context.Background(), activityID, 1001)

	if err == nil {
		t.Error("expected error for duplicate order, got nil")
	}
	if order != nil {
		t.Error("order should be nil on duplicate")
	}

	// 验证Redis中用户已被标记
	if !redisRepo.userSets[activityID][1001] {
		t.Error("user should be marked as bought in Redis")
	}

	// 验证只有1个订单被创建
	if len(orderRepo.orders) != 1 {
		t.Errorf("expected 1 order, got %d (duplicate should not create new order)", len(orderRepo.orders))
	}
}

func TestCreateSeckillOrder_ConcurrentStockDecrement(t *testing.T) {
	orderRepo := newMockOrderRepo()
	redisRepo := newMockRedisRepo()
	mqRepo := newMockMQRepo()
	activityRepo := newMockActivityRepo()
	stockClient := &mockStockClient{}

	activityRepo.activities["act_005"] = &SeckillActivity{
		ActivityID:    "act_005",
		SkuID:         "sku_005",
		TotalCount:    100,
		ActivityState: 1,
		StartTime:     time.Now().Add(-1 * time.Hour).Format("2006-01-02 15:04:05"),
		EndTime:       time.Now().Add(1 * time.Hour).Format("2006-01-02 15:04:05"),
	}

	tradeSvc := NewTradeService(orderRepo, redisRepo, mqRepo, activityRepo, stockClient)

	activityID := "act_005"
	redisRepo.stocks[activityID] = 5

	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func(userID int64) {
			_, _ = tradeSvc.CreateSeckillOrder(context.Background(), activityID, userID)
			done <- true
		}(int64(2000 + i))
	}

	for i := 0; i < 10; i++ {
		<-done
	}

	// 库存不应该为负
	stock := redisRepo.stocks[activityID]
	if stock < 0 {
		t.Errorf("stock should not be negative, got %d", stock)
	}

	// 最多只能有5个成功订单（库存为5）
	successfulOrders := 0
	for _, o := range orderRepo.orders {
		if o.ActivityID == activityID {
			successfulOrders++
		}
	}
	if successfulOrders > 5 {
		t.Errorf("should have at most 5 successful orders, got %d", successfulOrders)
	}
}

func TestGetOrder_Success(t *testing.T) {
	orderRepo := newMockOrderRepo()
	redisRepo := newMockRedisRepo()
	mqRepo := newMockMQRepo()
	activityRepo := newMockActivityRepo()
	stockClient := &mockStockClient{}

	activityRepo.activities["act_004"] = &SeckillActivity{
		ActivityID:    "act_004",
		SkuID:         "sku_004",
		TotalCount:    100,
		ActivityState: 1,
		StartTime:     time.Now().Add(-1 * time.Hour).Format("2006-01-02 15:04:05"),
		EndTime:       time.Now().Add(1 * time.Hour).Format("2006-01-02 15:04:05"),
	}

	tradeSvc := NewTradeService(orderRepo, redisRepo, mqRepo, activityRepo, stockClient)

	activityID := "act_004"
	redisRepo.stocks[activityID] = 10
	createdOrder, _ := tradeSvc.CreateSeckillOrder(context.Background(), activityID, 1001)

	order, err := tradeSvc.GetOrder(context.Background(), createdOrder.OrderID)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if order.OrderID != createdOrder.OrderID {
		t.Errorf("expected order_id %s, got %s", createdOrder.OrderID, order.OrderID)
	}
}

func TestGetOrder_NotFound(t *testing.T) {
	orderRepo := newMockOrderRepo()
	redisRepo := newMockRedisRepo()
	mqRepo := newMockMQRepo()
	activityRepo := newMockActivityRepo()
	stockClient := &mockStockClient{}

	tradeSvc := NewTradeService(orderRepo, redisRepo, mqRepo, activityRepo, stockClient)

	order, err := tradeSvc.GetOrder(context.Background(), "non_existent")

	if err == nil {
		t.Error("expected error, got nil")
	}
	if order != nil {
		t.Error("order should be nil on error")
	}
}