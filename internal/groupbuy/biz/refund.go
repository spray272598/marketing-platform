package biz

import (
	"context"
	"fmt"

	"github.com/marketing-platform/pkg/common"
)

type RefundStrategy interface {
	Refund(ctx context.Context, order *GroupBuyOrder) error
}

type UnpaidRefundStrategy struct{}

func (s *UnpaidRefundStrategy) Refund(ctx context.Context, order *GroupBuyOrder) error {
	return nil
}

type PaidRefundStrategy struct {
	orderRepo   OrderRepo
	notifySvc   *NotifyService
	stockClient StockClient
}

func (s *PaidRefundStrategy) Refund(ctx context.Context, order *GroupBuyOrder) error {
	if err := s.orderRepo.UpdateOrderState(ctx, order.OrderID, 2); err != nil {
		return err
	}

	stockKey := fmt.Sprintf("team:%s", order.TeamID)
	if err := s.stockClient.RestoreStock(ctx, stockKey, 1); err != nil {
		return fmt.Errorf("restore stock failed: %w", err)
	}

	return s.notifySvc.CreateRefundNotify(ctx, order.OrderID, map[string]interface{}{
		"order_id":    order.OrderID,
		"user_id":     order.UserID,
		"activity_id": order.ActivityID,
	})
}

type RefundService struct {
	strategies  map[string]RefundStrategy
	orderRepo   OrderRepo
	stockClient StockClient
}

func NewRefundService(orderRepo OrderRepo, notifySvc *NotifyService, stockClient StockClient) *RefundService {
	return &RefundService{
		strategies: map[string]RefundStrategy{
			"unpaid": &UnpaidRefundStrategy{},
			"paid":   &PaidRefundStrategy{orderRepo: orderRepo, notifySvc: notifySvc, stockClient: stockClient},
		},
		orderRepo:   orderRepo,
		stockClient: stockClient,
	}
}

func (s *RefundService) Refund(ctx context.Context, orderID string) (*GroupBuyOrder, error) {
	order, err := s.orderRepo.GetOrder(ctx, orderID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", common.GroupBuyOrderNotExist.Code, err)
	}

	var strategy RefundStrategy
	if order.OrderState == 0 {
		strategy = s.strategies["unpaid"]
	} else {
		strategy = s.strategies["paid"]
	}

	if err := strategy.Refund(ctx, order); err != nil {
		return nil, fmt.Errorf("%s: %w", common.GroupBuyRefundFail.Code, err)
	}

	order.OrderState = 2
	return order, nil
}
