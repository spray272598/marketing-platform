package data

import (
	"context"
	"fmt"

	"github.com/marketing-platform/internal/groupbuy/biz"
	"github.com/marketing-platform/internal/groupbuy/data/ent"
	"github.com/marketing-platform/internal/groupbuy/data/ent/groupbuyorder"
	"time"
)

type orderRepo struct {
	data *Data
}

func NewOrderRepo(data *Data) biz.OrderRepo {
	return &orderRepo{data: data}
}

func (r *orderRepo) CreateOrder(ctx context.Context, order *biz.GroupBuyOrder) error {
	t, _ := time.Parse("2006-01-02 15:04:05", order.CreateAt)
	_, err := r.data.db.GroupBuyOrder.Create().
		SetOrderID(order.OrderID).
		SetTeamID(order.TeamID).
		SetUserID(order.UserID).
		SetActivityID(order.ActivityID).
		SetBizID(order.BizID).
		SetOrderState(order.OrderState).
		SetCreatedAt(t).
		Save(ctx)
	return err
}

func (r *orderRepo) GetOrder(ctx context.Context, orderID string) (*biz.GroupBuyOrder, error) {
	po, err := r.data.db.GroupBuyOrder.Query().
		Where(groupbuyorder.OrderIDEQ(orderID)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("order not found: %s", orderID)
		}
		return nil, err
	}
	createAt := ""
	if po.CreatedAt != nil {
		createAt = po.CreatedAt.Format("2006-01-02 15:04:05")
	}
	return &biz.GroupBuyOrder{
		OrderID:    po.OrderID,
		TeamID:     po.TeamID,
		UserID:     po.UserID,
		ActivityID: po.ActivityID,
		BizID:      po.BizID,
		OrderState: po.OrderState,
		CreateAt:   createAt,
	}, nil
}

func (r *orderRepo) UpdateOrderState(ctx context.Context, orderID string, state int32) error {
	_, err := r.data.db.GroupBuyOrder.Update().
		Where(groupbuyorder.OrderIDEQ(orderID)).
		SetOrderState(state).
		Save(ctx)
	return err
}
