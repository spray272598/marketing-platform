package biz

import (
	"context"

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
	orderRepo OrderRepo
	mqRepo    MQRepo
}

func (s *PaidRefundStrategy) Refund(ctx context.Context, order *GroupBuyOrder) error {
	if err := s.orderRepo.UpdateOrderState(ctx, order.OrderID, common.OrderStateCancelled); err != nil {
		return err
	}
	return s.mqRepo.PublishRefundMessage(ctx, order.OrderID)
}

type RefundService struct {
	strategies map[string]RefundStrategy
	orderRepo  OrderRepo
}

func NewRefundService(orderRepo OrderRepo, mqRepo MQRepo) *RefundService {
	return &RefundService{
		strategies: map[string]RefundStrategy{
			"unpaid": &UnpaidRefundStrategy{},
			"paid":   &PaidRefundStrategy{orderRepo: orderRepo, mqRepo: mqRepo},
		},
		orderRepo: orderRepo,
	}
}

func (s *RefundService) Refund(ctx context.Context, orderID string) (*GroupBuyOrder, error) {
	order, err := s.orderRepo.GetOrder(ctx, orderID)
	if err != nil {
		return nil, err
	}

	var strategy RefundStrategy
	if order.OrderState == common.OrderStateInit {
		strategy = s.strategies["unpaid"]
	} else {
		strategy = s.strategies["paid"]
	}

	if err := strategy.Refund(ctx, order); err != nil {
		return nil, err
	}

	order.OrderState = common.OrderStateCancelled
	return order, nil
}
