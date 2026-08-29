package data

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/marketing-platform/internal/seckill/biz"
	"github.com/marketing-platform/internal/seckill/data/ent"
	"github.com/marketing-platform/internal/seckill/data/ent/seckillorder"
	"github.com/marketing-platform/pkg/idgen"
)

const orderIDStep = 1000

type orderRepo struct {
	data *Data
	gen  *idgen.Generator
}

func NewOrderRepo(data *Data) biz.OrderRepo {
	return &orderRepo{
		data: data,
		gen:  idgen.NewGenerator(newEntSegmentStore(data.sqldb), orderIDStep),
	}
}

// NextOrderID 返回当前服务订单的下一个号段 ID（基于美团 Leaf 号段模式）。
func (r *orderRepo) NextOrderID(ctx context.Context, bizTag string) (int64, error) {
	return r.gen.Next(ctx, bizTag)
}

// entSegmentStore 用原生 *sql.DB 在事务内原子推进 id_segment.max_id。
type entSegmentStore struct {
	db *sql.DB
}

func newEntSegmentStore(db *sql.DB) idgen.SegmentStore {
	return &entSegmentStore{db: db}
}

func (s *entSegmentStore) AllocMax(ctx context.Context, bizTag string, step int64) (int64, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("idgen: begin tx: %w", err)
	}
	defer tx.Rollback()

	var cur int64
	if err := tx.QueryRowContext(ctx,
		"SELECT max_id FROM id_segment WHERE biz_tag = ? FOR UPDATE", bizTag,
	).Scan(&cur); err != nil {
		return 0, fmt.Errorf("idgen: select max_id: %w", err)
	}
	newMax := cur + step
	if _, err := tx.ExecContext(ctx,
		"UPDATE id_segment SET max_id = ? WHERE biz_tag = ?", newMax, bizTag,
	); err != nil {
		return 0, fmt.Errorf("idgen: update max_id: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("idgen: commit: %w", err)
	}
	return newMax, nil
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

func (r *orderRepo) UpdateOrderState(ctx context.Context, orderID string, state int32) error {
	_, err := r.data.db.SeckillOrder.Update().
		Where(seckillorder.OrderIDEQ(orderID)).
		SetOrderState(state).
		Save(ctx)
	return err
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
