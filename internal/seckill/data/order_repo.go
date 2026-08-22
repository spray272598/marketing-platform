package data

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/marketing-platform/internal/seckill/biz"
)

type OrderRepo struct {
	db *sql.DB
}

func NewOrderRepo(db *sql.DB) *OrderRepo {
	return &OrderRepo{db: db}
}

func (r *OrderRepo) CreateOrder(ctx context.Context, order *biz.SeckillOrder) error {
	query := `INSERT INTO seckill_order (order_id, activity_id, user_id, sku_id, order_state, order_time) 
			  VALUES (?, ?, ?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query,
		order.OrderID, order.ActivityID, order.UserID,
		order.SkuID, order.OrderState, order.OrderTime,
	)
	return err
}

func (r *OrderRepo) GetOrder(ctx context.Context, orderID string) (*biz.SeckillOrder, error) {
	query := `SELECT id, order_id, activity_id, user_id, sku_id, order_state, order_time, pay_time 
			  FROM seckill_order WHERE order_id = ?`

	order := &biz.SeckillOrder{}
	err := r.db.QueryRowContext(ctx, query, orderID).Scan(
		&order.ID, &order.OrderID, &order.ActivityID,
		&order.UserID, &order.SkuID, &order.OrderState,
		&order.OrderTime, &order.PayTime,
	)
	if err != nil {
		return nil, fmt.Errorf("order not found: %w", err)
	}
	return order, nil
}

func (r *OrderRepo) GetUserActivityOrder(ctx context.Context, userID int64, activityID string) (*biz.SeckillOrder, error) {
	query := `SELECT id, order_id, activity_id, user_id, sku_id, order_state, order_time, pay_time 
			  FROM seckill_order WHERE user_id = ? AND activity_id = ?`

	order := &biz.SeckillOrder{}
	err := r.db.QueryRowContext(ctx, query, userID, activityID).Scan(
		&order.ID, &order.OrderID, &order.ActivityID,
		&order.UserID, &order.SkuID, &order.OrderState,
		&order.OrderTime, &order.PayTime,
	)
	if err != nil {
		return nil, err
	}
	return order, nil
}
