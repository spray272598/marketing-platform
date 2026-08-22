package biz

import (
	"context"
	"fmt"
	"time"

	"github.com/marketing-platform/pkg/common"
)

type TradeService struct {
	orderRepo OrderRepo
	redisRepo RedisRepo
	mqRepo    MQRepo
}

func NewTradeService(
	orderRepo OrderRepo,
	redisRepo RedisRepo,
	mqRepo MQRepo,
) *TradeService {
	return &TradeService{
		orderRepo: orderRepo,
		redisRepo: redisRepo,
		mqRepo:    mqRepo,
	}
}

func (s *TradeService) CreateSeckillOrder(ctx context.Context, activityID string, userID int64) (*SeckillOrder, error) {
	// 1. 检查是否已下单
	existOrder, _ := s.orderRepo.GetUserActivityOrder(ctx, userID, activityID)
	if existOrder != nil {
		return nil, fmt.Errorf(common.SeckillOrderDuplicate.Code + ": " + common.SeckillOrderDuplicate.Info)
	}

	// 2. Redis原子扣减库存
	ok, err := s.redisRepo.DecrStock(ctx, activityID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf(common.SeckillStockNotEnough.Code + ": " + common.SeckillStockNotEnough.Info)
	}

	// 3. 创建订单
	order := &SeckillOrder{
		OrderID:    fmt.Sprintf("sk_%d_%d", time.Now().UnixMilli(), userID),
		ActivityID: activityID,
		UserID:     userID,
		OrderState: common.OrderStateInit,
		OrderTime:  time.Now().Format("2006-01-02 15:04:05"),
	}

	if err := s.orderRepo.CreateOrder(ctx, order); err != nil {
		// 回滚库存
		return nil, err
	}

	// 4. 异步发消息
	_ = s.mqRepo.PublishOrderMessage(ctx, order)

	return order, nil
}

func (s *TradeService) GetOrder(ctx context.Context, orderID string) (*SeckillOrder, error) {
	return s.orderRepo.GetOrder(ctx, orderID)
}
