package data

import (
	"context"
	"fmt"

	"github.com/marketing-platform/internal/lottery/biz"
)

type activityRepo struct {
	db *Data
}

func NewActivityRepo(data *Data) biz.ActivityRepo {
	return &activityRepo{db: data}
}

func (r *activityRepo) GetActivity(ctx context.Context, activityID string) (*biz.LotteryActivity, error) {
	query := `SELECT id, activity_id, activity_name, strategy_id, activity_state 
			  FROM lottery_activity WHERE activity_id = ?`

	activity := &biz.LotteryActivity{}
	err := r.db.db.QueryRowContext(ctx, query, activityID).Scan(
		&activity.ID, &activity.ActivityID, &activity.ActivityName,
		&activity.StrategyID, &activity.ActivityState,
	)
	if err != nil {
		return nil, fmt.Errorf("activity not found: %w", err)
	}
	return activity, nil
}
