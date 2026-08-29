package biz

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/marketing-platform/pkg/common"
	"github.com/marketing-platform/pkg/saga"
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
	logger      *slog.Logger
}

func (s *PaidRefundStrategy) Refund(ctx context.Context, order *GroupBuyOrder) error {
	stockKey := fmt.Sprintf("team:%s", order.TeamID)
	// 记录退款前的状态，供补偿时精确回滚。
	prevState := order.OrderState

	c := saga.New(
		// 1) 本地：把订单置为已取消。
		saga.Step{
			Name: "mark_order_cancelled",
			Action: func(ctx context.Context) error {
				return s.orderRepo.UpdateOrderState(ctx, order.OrderID, common.OrderStateCancelled)
			},
			Compensate: func(ctx context.Context) error {
				return s.orderRepo.UpdateOrderState(ctx, order.OrderID, prevState)
			},
		},
		// 2) 跨服务：归还库存。失败时把库存重新扣回去，避免库存凭空增加。
		saga.Step{
			Name: "restore_stock",
			Action: func(ctx context.Context) error {
				return s.stockClient.RestoreStock(ctx, stockKey, 1)
			},
			Compensate: func(ctx context.Context) error {
				return s.stockClient.DeductStock(ctx, stockKey, 1)
			},
		},
		// 3) 本地：落一条退款通知任务（本地消息表，由后台任务保证投递）。
		saga.Step{
			Name: "create_refund_notify",
			Action: func(ctx context.Context) error {
				return s.notifySvc.CreateRefundNotify(ctx, order.OrderID, map[string]interface{}{
					"order_id":    order.OrderID,
					"user_id":     order.UserID,
					"activity_id": order.ActivityID,
				})
			},
			// 无需反向补偿：前两步的补偿已让订单与库存回到退款前状态，
			// 通知只是"尚未创建"，后续重试会补上。
		},
	).WithLogger(s.logger).WithLog(saga.NewSlogLog(s.logger))

	if err := c.Run(ctx); err != nil {
		// 补偿失败意味着系统可能停留在不一致状态，必须显式告警以便人工介入。
		var sErr *saga.SagaError
		if errors.As(err, &sErr) && len(sErr.CompErrors) > 0 {
			s.logger.Error("REFUND SAGA COMPENSATION FAILED, manual intervention required",
				slog.String("order_id", order.OrderID),
				slog.String("saga_error", err.Error()),
			)
		}
		return err
	}
	return nil
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
			"paid": &PaidRefundStrategy{
				orderRepo:   orderRepo,
				notifySvc:   notifySvc,
				stockClient: stockClient,
				logger:      slog.Default(),
			},
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
	if order == nil {
		return nil, fmt.Errorf("%s: order not found: %s", common.GroupBuyOrderNotExist.Code, orderID)
	}

	var strategy RefundStrategy
	if order.OrderState == common.OrderStateInit {
		strategy = s.strategies["unpaid"]
	} else {
		strategy = s.strategies["paid"]
	}

	if err := strategy.Refund(ctx, order); err != nil {
		return nil, fmt.Errorf("%s: %w", common.GroupBuyRefundFail.Code, err)
	}

	order.OrderState = common.OrderStateCancelled
	return order, nil
}
