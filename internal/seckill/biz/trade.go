package biz

import (
	"context"
	"fmt"
	"time"

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
	// 1. 检查活动是否存在
	activity, err := s.activityRepo.GetActivity(ctx, activityID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", common.SeckillActivityNotExist.Code, err)
	}

	// 2. 原子操作：检查用户是否已下单 + 检查库存 + 扣减库存 + 标记用户
	// 这一步同时解决了一人一单和超卖问题
	result, err := s.redisRepo.DecrStockWithUserCheck(ctx, activityID, userID)
	if err != nil {
		return nil, fmt.Errorf("redis atomic operation failed: %w", err)
	}

	switch result {
	case 0:
		return nil, fmt.Errorf("%s: %s", common.SeckillStockNotEnough.Code, common.SeckillStockNotEnough.Info)
	case 2:
		return nil, fmt.Errorf("%s: %s", common.SeckillOrderDuplicate.Code, common.SeckillOrderDuplicate.Info)
	}

	// result == 1, 扣减成功，继续后续流程

	// 3. 调用stock服务扣减持久化库存（可选，如果需要双写保证）
	stockKey := fmt.Sprintf("product:%s", activity.SkuID)
	if err := s.stockClient.DeductStock(ctx, stockKey, 1); err != nil {
		// Redis已扣减成功，但stock服务失败，需要回滚Redis
		// 这里为了演示简化，实际应补偿Redis库存
		return nil, fmt.Errorf("%s: %w", common.SeckillStockNotEnough.Code, err)
	}

	// 4. 创建订单
	order := &SeckillOrder{
		OrderID:    fmt.Sprintf("sk_%d_%d", time.Now().UnixMilli(), userID),
		ActivityID: activityID,
		UserID:     userID,
		SkuID:      activity.SkuID,
		OrderState: common.OrderStateInit,
		OrderTime:  time.Now().Format("2006-01-02 15:04:05"),
	}

	if err := s.orderRepo.CreateOrder(ctx, order); err != nil {
		return nil, err
	}

	// 5. 异步发送消息（MQ）
	_ = s.mqRepo.PublishOrderMessage(ctx, order)

	return order, nil
}

func (s *TradeService) GetOrder(ctx context.Context, orderID string) (*SeckillOrder, error) {
	return s.orderRepo.GetOrder(ctx, orderID)
}