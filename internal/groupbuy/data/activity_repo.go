package data

import (
	"context"
	"fmt"

	"github.com/marketing-platform/internal/groupbuy/biz"
)

type activityRepo struct {
	db *Data
}

func NewActivityRepo(data *Data) biz.ActivityRepo {
	return &activityRepo{db: data}
}

func (r *activityRepo) GetActivity(ctx context.Context, activityID string) (*biz.GroupBuyActivity, error) {
	query := `SELECT id, activity_id, activity_name, discount_id, group_type, 
			  target_count, valid_time, tag_id, activity_state 
			  FROM groupbuy_activity WHERE activity_id = ?`

	activity := &biz.GroupBuyActivity{}
	err := r.db.db.QueryRowContext(ctx, query, activityID).Scan(
		&activity.ID, &activity.ActivityID, &activity.ActivityName,
		&activity.DiscountID, &activity.GroupType, &activity.TargetCount,
		&activity.ValidTime, &activity.TagID, &activity.ActivityState,
	)
	if err != nil {
		return nil, fmt.Errorf("activity not found: %w", err)
	}
	return activity, nil
}

func (r *activityRepo) GetDiscount(ctx context.Context, discountID string) (*biz.GroupBuyDiscount, error) {
	query := `SELECT id, discount_id, market_plan, market_expr 
			  FROM groupbuy_discount WHERE discount_id = ?`

	discount := &biz.GroupBuyDiscount{}
	err := r.db.db.QueryRowContext(ctx, query, discountID).Scan(
		&discount.ID, &discount.DiscountID, &discount.MarketPlan, &discount.MarketExpr,
	)
	if err != nil {
		return nil, fmt.Errorf("discount not found: %w", err)
	}
	return discount, nil
}
