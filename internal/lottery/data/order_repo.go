package data

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/marketing-platform/internal/lottery/biz"
	"github.com/marketing-platform/internal/lottery/data/ent/lotteryorder"
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

// entSegmentStore 用原生的 *sql.DB 在事务内原子推进 id_segment.max_id。
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
