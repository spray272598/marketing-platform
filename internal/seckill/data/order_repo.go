package data

import (
	"context"
	"fmt"
	"time"

	"github.com/marketing-platform/internal/seckill/biz"
	"github.com/marketing-platform/internal/seckill/data/ent"
	"github.com/marketing-platform/internal/seckill/data/ent/seckillorder"
)

type orderRepo struct {
	data *Data
}

func NewOrderRepo(data *Data) biz.OrderRepo {
	return &orderRepo{data: data}
}

func (r *orderRepo) CreateOrder(ctx context.Context, order *biz.SeckillOrder) error {
	_, err := r.data.db.SeckillOrder.Create().
		SetOrderID(order.OrderID).
		SetActivityID(order.ActivityID).
		SetUserID(order.UserID).
		SetSkuID(order.SkuID).
		SetOrderState(order.OrderState).
		SetOrderTime(parseTime(order.OrderTime)).
		Save(ctx)
	return err
}

func (r *orderRepo) GetOrder(ctx context.Context, orderID string) (*biz.SeckillOrder, error) {
	po, err := r.data.db.SeckillOrder.Query().
		Where(seckillorder.OrderIDEQ(orderID)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("order not found: %s", orderID)
		}
		return nil, err
	}
	return toBizOrder(po), nil
}

func (r *orderRepo) GetUserActivityOrder(ctx context.Context, userID int64, activityID string) (*biz.SeckillOrder, error) {
	po, err := r.data.db.SeckillOrder.Query().
		Where(
			seckillorder.UserIDEQ(userID),
			seckillorder.ActivityIDEQ(activityID),
		).
		Only(ctx)
	if err != nil {
		return nil, err
	}
	return toBizOrder(po), nil
}

func toBizOrder(po *ent.SeckillOrder) *biz.SeckillOrder {
	if po == nil {
		return nil
	}
	o := &biz.SeckillOrder{
		OrderID:    po.OrderID,
		ActivityID: po.ActivityID,
		UserID:     po.UserID,
		SkuID:      po.SkuID,
		OrderState: po.OrderState,
	}
	if po.OrderTime != (time.Time{}) {
		o.OrderTime = po.OrderTime.Format("2006-01-02 15:04:05")
	}
	if po.PayTime != nil {
		o.PayTime = po.PayTime.Format("2006-01-02 15:04:05")
	}
	return o
}

func parseTime(s string) time.Time {
	t, _ := time.Parse("2006-01-02 15:04:05", s)
	return t
}
