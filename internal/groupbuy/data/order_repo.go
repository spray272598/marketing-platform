package data

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/marketing-platform/internal/groupbuy/biz"
	"github.com/marketing-platform/internal/groupbuy/data/ent"
	"github.com/marketing-platform/internal/groupbuy/data/ent/groupbuyorder"
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
