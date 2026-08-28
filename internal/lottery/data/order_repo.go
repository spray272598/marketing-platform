package data

import (
	"context"
	"fmt"
	"time"

	"github.com/marketing-platform/internal/lottery/biz"
	"github.com/marketing-platform/internal/lottery/data/ent/lotteryorder"
)

type orderRepo struct{ data *Data }

func NewOrderRepo(data *Data) biz.OrderRepo { return &orderRepo{data: data} }

func (r *orderRepo) CreateOrder(ctx context.Context, order *biz.LotteryOrder) error {
	t, _ := time.Parse("2006-01-02 15:04:05", order.AwardTime)
	_, err := r.data.db.LotteryOrder.Create().
		SetOrderID(order.OrderID).
		SetActivityID(order.ActivityID).
		SetUserID(order.UserID).
		SetNillableAwardID(&order.AwardID).
		SetAwardState(order.AwardState).
		SetNillableAwardTime(&t).
		Save(ctx)
	return err
}

func (r *orderRepo) GetUserActivityCount(ctx context.Context, userID int64, activityID string) (int32, error) {
	count, err := r.data.db.LotteryOrder.Query().
		Where(lotteryorder.UserIDEQ(userID), lotteryorder.ActivityIDEQ(activityID)).
		Count(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to count user activity: %w", err)
	}
	return int32(count), nil
}
