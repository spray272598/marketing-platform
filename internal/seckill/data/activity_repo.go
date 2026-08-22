package data

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/marketing-platform/internal/seckill/biz"
)

type ActivityRepo struct {
	db *sql.DB
}

func NewActivityRepo(db *sql.DB) *ActivityRepo {
	return &ActivityRepo{db: db}
}

func (r *ActivityRepo) GetActivity(ctx context.Context, activityID string) (*biz.SeckillActivity, error) {
	query := `SELECT id, activity_id, activity_name, sku_id, total_count, limit_count, activity_state, start_time, end_time 
			  FROM seckill_activity WHERE activity_id = ?`

	activity := &biz.SeckillActivity{}
	err := r.db.QueryRowContext(ctx, query, activityID).Scan(
		&activity.ID, &activity.ActivityID, &activity.ActivityName,
		&activity.SkuID, &activity.TotalCount, &activity.LimitCount,
		&activity.ActivityState, &activity.StartTime, &activity.EndTime,
	)
	if err != nil {
		return nil, fmt.Errorf("activity not found: %w", err)
	}
	return activity, nil
}

func (r *ActivityRepo) UpdateActivityStock(ctx context.Context, activityID string, stock int32) error {
	query := `UPDATE seckill_activity SET total_count = ? WHERE activity_id = ?`
	_, err := r.db.ExecContext(ctx, query, stock, activityID)
	return err
}
