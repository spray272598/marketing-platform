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
	existOrder, _ := s.orderRepo.GetUserActivityOrder(ctx, userID, activityID)
	if existOrder != nil {
		return nil, fmt.Errorf("%s: %s", common.SeckillOrderDuplicate.Code, common.SeckillOrderDuplicate.Info)
	}

	activity, err := s.activityRepo.GetActivity(ctx, activityID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", common.SeckillActivityNotExist.Code, err)
	}

	stockKey := fmt.Sprintf("product:%s", activity.SkuID)
	if err := s.stockClient.DeductStock(ctx, stockKey, 1); err != nil {
		return nil, fmt.Errorf("%s: %w", common.SeckillStockNotEnough.Code, err)
	}

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

	_ = s.mqRepo.PublishOrderMessage(ctx, order)

	return order, nil
}

func (s *TradeService) GetOrder(ctx context.Context, orderID string) (*SeckillOrder, error) {
	return s.orderRepo.GetOrder(ctx, orderID)
}
