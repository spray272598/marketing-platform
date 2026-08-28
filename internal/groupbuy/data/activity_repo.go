package data

import (
	"context"
	"fmt"

	"github.com/marketing-platform/internal/groupbuy/biz"
	"github.com/marketing-platform/internal/groupbuy/data/ent"
	"github.com/marketing-platform/internal/groupbuy/data/ent/groupbuyactivity"
	"github.com/marketing-platform/internal/groupbuy/data/ent/groupbuydiscount"
)

type activityRepo struct {
	data *Data
}

func NewActivityRepo(data *Data) biz.ActivityRepo {
	return &activityRepo{data: data}
}

func (r *activityRepo) GetActivity(ctx context.Context, activityID string) (*biz.GroupBuyActivity, error) {
	po, err := r.data.db.GroupBuyActivity.Query().
		Where(groupbuyactivity.ActivityIDEQ(activityID)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("activity not found: %s", activityID)
		}
		return nil, err
	}
	return &biz.GroupBuyActivity{
		ActivityID:    po.ActivityID,
		ActivityName:  po.ActivityName,
		DiscountID:    po.DiscountID,
		GroupType:     po.GroupType,
		TargetCount:   po.TargetCount,
		ValidTime:     po.ValidTime,
		ActivityState: po.ActivityState,
	}, nil
}

func (r *activityRepo) GetDiscount(ctx context.Context, discountID string) (*biz.GroupBuyDiscount, error) {
	po, err := r.data.db.GroupBuyDiscount.Query().
		Where(groupbuydiscount.DiscountIDEQ(discountID)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("discount not found: %s", discountID)
		}
		return nil, err
	}
	return &biz.GroupBuyDiscount{
		DiscountID: po.DiscountID,
		MarketPlan: po.MarketPlan,
		MarketExpr: po.MarketExpr,
	}, nil
}
