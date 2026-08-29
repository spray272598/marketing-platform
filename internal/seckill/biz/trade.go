package biz

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/marketing-platform/pkg/common"
)

type TradeService struct {
	orderRepo   OrderRepo
	redisRepo   RedisRepo
	mqRepo      MQRepo
	activityRepo ActivityRepo
	stockClient  StockClient
}

type StockClient interface {
	DeductStock(ctx context.Context, stockKey string, count int32) error
	RestoreStock(ctx context.Context, stockKey string, count int32) error
}

func NewTradeService(
	orderRepo OrderRepo,
	redisRepo RedisRepo,
	mqRepo MQRepo,
	activityRepo ActivityRepo,
	stockClient StockClient,
) *TradeService {
	return &TradeService{
		orderRepo:    orderRepo,
		redisRepo:    redisRepo,
		mqRepo:       mqRepo,
		activityRepo: activityRepo,
		stockClient:  stockClient,
	}
}

func (s *TradeService) CreateSeckillOrder(ctx context.Context, activityID string, userID int64) (*SeckillOrder, error) {
	activity, err := s.activityRepo.GetActivity(ctx, activityID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", common.SeckillActivityNotExist.Code, err)
	}

	if activity.ActivityState != common.ActivityStateOpen {
		return nil, fmt.Errorf("%s: %s", common.SeckillActivityClosed.Code, common.SeckillActivityClosed.Info)
	}

	now := time.Now()
	// 用 ParseInLocation(time.Local) 解析，避免时间字符串按 UTC 解析、
	// 而 time.Now() 为本地时间导致的时区错位（非 UTC 环境会误判活动已关闭）。
	startTime, err := time.ParseInLocation("2006-01-02 15:04:05", activity.StartTime, time.Local)
	if err != nil {
		return nil, fmt.Errorf("invalid activity start_time: %w", err)
	}
	endTime, err := time.ParseInLocation("2006-01-02 15:04:05", activity.EndTime, time.Local)
	if err != nil {
		return nil, fmt.Errorf("invalid activity end_time: %w", err)
	}
	if now.Before(startTime) || now.After(endTime) {
		return nil, fmt.Errorf("%s: %s", common.SeckillActivityClosed.Code, common.SeckillActivityClosed.Info)
	}

	limit := activity.LimitCount
	if limit <= 0 {
		limit = 1
	}
	result, err := s.redisRepo.DecrStockWithUserCheck(ctx, activityID, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("redis atomic operation failed: %w", err)
	}
	switch result {
	case 0:
		return nil, fmt.Errorf("%s: %s", common.SeckillStockNotEnough.Code, common.SeckillStockNotEnough.Info)
	case 2:
		return nil, fmt.Errorf("%s: %s", common.SeckillOrderDuplicate.Code, common.SeckillOrderDuplicate.Info)
	}

	stockKey := fmt.Sprintf("product:%s", activity.SkuID)
	if err := s.stockClient.DeductStock(ctx, stockKey, 1); err != nil {
		_ = s.redisRepo.IncrStock(ctx, activityID, 1)
		return nil, fmt.Errorf("%s: %w", common.SeckillStockNotEnough.Code, err)
	}

	order := &SeckillOrder{
		OrderID:    fmt.Sprintf("sk_%s", uuid.New().String()[:12]),
		ActivityID: activityID,
		UserID:     userID,
		SkuID:      activity.SkuID,
		OrderState: common.OrderStateInit,
		OrderTime:  time.Now().Format("2006-01-02 15:04:05"),
	}

	if err := s.orderRepo.CreateOrder(ctx, order); err != nil {
		_ = s.stockClient.RestoreStock(ctx, stockKey, 1)
		_ = s.redisRepo.IncrStock(ctx, activityID, 1)
		return nil, fmt.Errorf("create order failed: %w", err)
	}

	if err := s.mqRepo.PublishOrderMessage(ctx, order); err != nil {
		_ = s.orderRepo.UpdateOrderState(ctx, order.OrderID, common.OrderStateCancelled)
		_ = s.stockClient.RestoreStock(ctx, stockKey, 1)
		_ = s.redisRepo.IncrStock(ctx, activityID, 1)
	}

	return order, nil
}

func (s *TradeService) GetOrder(ctx context.Context, orderID string) (*SeckillOrder, error) {
	return s.orderRepo.GetOrder(ctx, orderID)
}