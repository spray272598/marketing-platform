package data

import (
	"context"
	"fmt"

	"github.com/marketing-platform/internal/lottery/biz"
	"github.com/marketing-platform/internal/lottery/data/ent"
	"github.com/marketing-platform/internal/lottery/data/ent/lotteryactivity"
)

type activityRepo struct{ data *Data }

func NewActivityRepo(data *Data) biz.ActivityRepo { return &activityRepo{data: data} }

func (r *activityRepo) GetActivity(ctx context.Context, activityID string) (*biz.LotteryActivity, error) {
	po, err := r.data.db.LotteryActivity.Query().
		Where(lotteryactivity.ActivityIDEQ(activityID)).Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("activity not found: %s", activityID)
		}
		return nil, err
	}
	return &biz.LotteryActivity{
		ActivityID:    po.ActivityID,
		ActivityName:  po.ActivityName,
		StrategyID:    po.StrategyID,
		ActivityState: po.ActivityState,
	}, nil
}
