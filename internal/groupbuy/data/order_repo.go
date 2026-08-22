package data

import (
	"context"
	"fmt"

	"github.com/marketing-platform/internal/groupbuy/biz"
)

type orderRepo struct {
	db *Data
}

func NewOrderRepo(data *Data) biz.OrderRepo {
	return &orderRepo{db: data}
}

func (r *orderRepo) CreateOrder(ctx context.Context, order *biz.GroupBuyOrder) error {
	query := `INSERT INTO groupbuy_order (order_id, team_id, user_id, activity_id, biz_id, order_state, create_at) 
			  VALUES (?, ?, ?, ?, ?, ?, ?)`
	_, err := r.db.db.ExecContext(ctx, query,
		order.OrderID, order.TeamID, order.UserID,
		order.ActivityID, order.BizID, order.OrderState, order.CreateAt,
	)
	return err
}

func (r *orderRepo) GetOrder(ctx context.Context, orderID string) (*biz.GroupBuyOrder, error) {
	query := `SELECT id, order_id, team_id, user_id, activity_id, biz_id, order_state, create_at 
			  FROM groupbuy_order WHERE order_id = ?`

	order := &biz.GroupBuyOrder{}
	err := r.db.db.QueryRowContext(ctx, query, orderID).Scan(
		&order.ID, &order.OrderID, &order.TeamID,
		&order.UserID, &order.ActivityID, &order.BizID,
		&order.OrderState, &order.CreateAt,
	)
	if err != nil {
		return nil, fmt.Errorf("order not found: %w", err)
	}
	return order, nil
}

func (r *orderRepo) UpdateOrderState(ctx context.Context, orderID string, state int32) error {
	query := `UPDATE groupbuy_order SET order_state = ?, update_at = NOW() WHERE order_id = ?`
	_, err := r.db.db.ExecContext(ctx, query, state, orderID)
	return err
}
