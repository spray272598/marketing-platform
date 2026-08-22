package biz

import (
	"context"
	"testing"
)

func TestCreateSeckillOrder_Success(t *testing.T) {
	orderRepo := newMockOrderRepo()
	redisRepo := newMockRedisRepo()
	mqRepo := newMockMQRepo()
	activityRepo := newMockActivityRepo()
	stockClient := &mockStockClient{}

	activityRepo.activities["act_001"] = &SeckillActivity{
		ActivityID: "act_001",
		SkuID:      "sku_001",
		TotalCount: 100,
	}

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

	activityID := "non_existent"
	redisRepo.stocks[activityID] = 10

	order, err := tradeSvc.CreateSeckillOrder(context.Background(), activityID, 1001)

	if err == nil {
		t.Error("expected error, got nil")
	}
	if order != nil {
		t.Error("order should be nil on error")
	}
}

func TestCreateSeckillOrder_DuplicateOrder(t *testing.T) {
	orderRepo := newMockOrderRepo()
	redisRepo := newMockRedisRepo()
	mqRepo := newMockMQRepo()
	activityRepo := newMockActivityRepo()
	stockClient := &mockStockClient{}

	activityRepo.activities["act_003"] = &SeckillActivity{
		ActivityID: "act_003",
		SkuID:      "sku_003",
		TotalCount: 100,
	}

	tradeSvc := NewTradeService(orderRepo, redisRepo, mqRepo, activityRepo, stockClient)

	activityID := "act_003"
	redisRepo.stocks[activityID] = 10

	_, err := tradeSvc.CreateSeckillOrder(context.Background(), activityID, 1001)
	if err != nil {
		t.Fatalf("first order failed: %v", err)
	}

	order, err := tradeSvc.CreateSeckillOrder(context.Background(), activityID, 1001)

	if err == nil {
		t.Error("expected error for duplicate order, got nil")
	}
	if order != nil {
		t.Error("order should be nil on error")
	}
}

func TestGetOrder_Success(t *testing.T) {
	orderRepo := newMockOrderRepo()
	redisRepo := newMockRedisRepo()
	mqRepo := newMockMQRepo()
	activityRepo := newMockActivityRepo()
	stockClient := &mockStockClient{}

	activityRepo.activities["act_004"] = &SeckillActivity{
		ActivityID: "act_004",
		SkuID:      "sku_004",
		TotalCount: 100,
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

func TestCreateSeckillOrder_ConcurrentStockDecrement(t *testing.T) {
	orderRepo := newMockOrderRepo()
	redisRepo := newMockRedisRepo()
	mqRepo := newMockMQRepo()
	activityRepo := newMockActivityRepo()
	stockClient := &mockStockClient{}

	activityRepo.activities["act_005"] = &SeckillActivity{
		ActivityID: "act_005",
		SkuID:      "sku_005",
		TotalCount: 100,
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

	stock := redisRepo.stocks[activityID]
	if stock < 0 {
		t.Errorf("stock should not be negative, got %d", stock)
	}
}
