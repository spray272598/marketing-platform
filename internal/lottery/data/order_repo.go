package data

import (
	"context"
	"fmt"

	"github.com/marketing-platform/internal/lottery/biz"
)

type orderRepo struct {
	db *Data
}

func NewOrderRepo(data *Data) biz.OrderRepo {
	return &orderRepo{db: data}
}

func (r *orderRepo) CreateOrder(ctx context.Context, order *biz.LotteryOrder) error {
	query := `INSERT INTO lottery_order (order_id, activity_id, user_id, award_id, award_state, award_time) 
			  VALUES (?, ?, ?, ?, ?, ?)`
	_, err := r.db.db.ExecContext(ctx, query,
		order.OrderID, order.ActivityID, order.UserID,
		order.AwardID, order.AwardState, order.AwardTime,
	)
	return err
}

func (r *orderRepo) GetUserActivityCount(ctx context.Context, userID int64, activityID string) (int32, error) {
	query := `SELECT COUNT(*) FROM lottery_order WHERE user_id = ? AND activity_id = ?`
	var count int32
	err := r.db.db.QueryRowContext(ctx, query, userID, activityID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count user activity: %w", err)
	}
	return count, nil
}
