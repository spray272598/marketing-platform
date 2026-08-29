package biz

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/marketing-platform/pkg/common"
)

// compensate 记录补偿操作的失败。
//
// 补偿本身失败意味着系统可能停留在不一致状态（例如库存已扣但订单没建成），
// 绝不能静默吞掉，必须留下错误日志以便告警与人工介入。
func compensate(action string, err error) {
	if err == nil {
		return
	}
	slog.Error("COMPENSATION FAILED, manual intervention required",
		slog.String("service", "seckill"),
		slog.String("action", action),
		slog.Any("error", err),
	)
}

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
		compensate("incr_redis_stock", s.redisRepo.IncrStock(ctx, activityID, 1))
		return nil, fmt.Errorf("%s: %w", common.SeckillStockNotEnough.Code, err)
	}

	id, err := s.orderRepo.NextOrderID(ctx, "order")
	if err != nil {
		compensate("restore_stock", s.stockClient.RestoreStock(ctx, stockKey, 1))
		compensate("incr_redis_stock", s.redisRepo.IncrStock(ctx, activityID, 1))
		return nil, fmt.Errorf("generate order id failed: %w", err)
	}
	order := &SeckillOrder{
		OrderID:    fmt.Sprintf("sk_%019d", id),
		ActivityID: activityID,
		UserID:     userID,
		SkuID:      activity.SkuID,
		OrderState: common.OrderStateInit,
		OrderTime:  time.Now().Format("2006-01-02 15:04:05"),
	}

	if err := s.orderRepo.CreateOrder(ctx, order); err != nil {
		compensate("restore_stock", s.stockClient.RestoreStock(ctx, stockKey, 1))
		compensate("incr_redis_stock", s.redisRepo.IncrStock(ctx, activityID, 1))
		return nil, fmt.Errorf("create order failed: %w", err)
	}

	if err := s.mqRepo.PublishOrderMessage(ctx, order); err != nil {
		compensate("cancel_order", s.orderRepo.UpdateOrderState(ctx, order.OrderID, common.OrderStateCancelled))
		compensate("restore_stock", s.stockClient.RestoreStock(ctx, stockKey, 1))
		compensate("incr_redis_stock", s.redisRepo.IncrStock(ctx, activityID, 1))
	}

	return order, nil
}

func (s *TradeService) GetOrder(ctx context.Context, orderID string) (*SeckillOrder, error) {
	return s.orderRepo.GetOrder(ctx, orderID)
}