package data

import (
	"context"
	"fmt"

	"github.com/marketing-platform/internal/seckill/biz"
	"github.com/marketing-platform/internal/seckill/data/ent"
	"github.com/marketing-platform/internal/seckill/data/ent/seckillactivity"
)

type activityRepo struct {
	data *Data
}

func NewActivityRepo(data *Data) biz.ActivityRepo {
	return &activityRepo{data: data}
}

func (r *activityRepo) GetActivity(ctx context.Context, activityID string) (*biz.SeckillActivity, error) {
	po, err := r.data.db.SeckillActivity.Query().
		Where(seckillactivity.ActivityIDEQ(activityID)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("activity not found: %s", activityID)
		}
		return nil, err
	}
	return toBizActivity(po), nil
}

func (r *activityRepo) UpdateActivityStock(ctx context.Context, activityID string, stock int32) error {
	_, err := r.data.db.SeckillActivity.Update().
		Where(seckillactivity.ActivityIDEQ(activityID)).
		SetTotalCount(stock).
		Save(ctx)
	return err
}

func toBizActivity(po *ent.SeckillActivity) *biz.SeckillActivity {
	if po == nil {
		return nil
	}
	return &biz.SeckillActivity{
		ActivityID:    po.ActivityID,
		ActivityName:  po.ActivityName,
		SkuID:         po.SkuID,
		TotalCount:    po.TotalCount,
		LimitCount:    po.LimitCount,
		ActivityState: po.ActivityState,
		StartTime:     po.StartTime.Format("2006-01-02 15:04:05"),
		EndTime:       po.EndTime.Format("2006-01-02 15:04:05"),
	}
}
