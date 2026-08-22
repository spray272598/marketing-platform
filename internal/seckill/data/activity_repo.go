package data

import (
	"context"
	"fmt"

	"github.com/marketing-platform/internal/seckill/biz"
)

type activityRepo struct {
	db *Data
}

func NewActivityRepo(data *Data) biz.ActivityRepo {
	return &activityRepo{db: data}
}

func (r *activityRepo) GetActivity(ctx context.Context, activityID string) (*biz.SeckillActivity, error) {
	query := `SELECT id, activity_id, activity_name, sku_id, total_count, limit_count, 
			  activity_state, start_time, end_time 
			  FROM seckill_activity WHERE activity_id = ?`

	activity := &biz.SeckillActivity{}
	err := r.db.db.QueryRowContext(ctx, query, activityID).Scan(
		&activity.ID, &activity.ActivityID, &activity.ActivityName,
		&activity.SkuID, &activity.TotalCount, &activity.LimitCount,
		&activity.ActivityState, &activity.StartTime, &activity.EndTime,
	)
	if err != nil {
		return nil, fmt.Errorf("activity not found: %w", err)
	}
	return activity, nil
}

func (r *activityRepo) UpdateActivityStock(ctx context.Context, activityID string, stock int32) error {
	query := `UPDATE seckill_activity SET total_count = ? WHERE activity_id = ?`
	_, err := r.db.db.ExecContext(ctx, query, stock, activityID)
	return err
}
