package biz

import (
	"context"
	"fmt"

	"github.com/marketing-platform/pkg/common"
)

type TrialService struct {
	activityRepo ActivityRepo
}

func NewTrialService(activityRepo ActivityRepo) *TrialService {
	return &TrialService{activityRepo: activityRepo}
}

type TrialResult struct {
	ActivityID        string `json:"activity_id"`
	MarketPlan        string `json:"market_plan"`
	MarketRule        string `json:"market_rule"`
	MarketDiscountAmt int32  `json:"market_discount_amount"`
	MarketPayAmount   int32  `json:"market_pay_amount"`
}

func (s *TrialService) TrialMarket(ctx context.Context, activityID string, originalPrice int32) (*TrialResult, error) {
	activity, err := s.activityRepo.GetActivity(ctx, activityID)
	if err != nil {
		return nil, fmt.Errorf(common.GroupBuyActivityNotExist.Code+": %w", err)
	}

	discount, err := s.activityRepo.GetDiscount(ctx, activity.DiscountID)
	if err != nil {
		return nil, err
	}

	result := &TrialResult{
		ActivityID: activity.ActivityID,
		MarketPlan: discount.MarketPlan,
		MarketRule: discount.MarketExpr,
	}

	switch discount.MarketPlan {
	case common.DiscountTypeZJ:
		var discountAmt int32
		fmt.Sscanf(discount.MarketExpr, "%d", &discountAmt)
		result.MarketDiscountAmt = discountAmt
		result.MarketPayAmount = originalPrice - discountAmt
	case common.DiscountTypeMJ:
		result.MarketPayAmount = originalPrice
	case common.DiscountTypeN:
		var nPrice int32
		fmt.Sscanf(discount.MarketExpr, "%d", &nPrice)
		result.MarketPayAmount = nPrice
		result.MarketDiscountAmt = originalPrice - nPrice
	case common.DiscountTypeZK:
		var rate float64
		fmt.Sscanf(discount.MarketExpr, "%f", &rate)
		result.MarketPayAmount = int32(float64(originalPrice) * rate / 10)
		result.MarketDiscountAmt = originalPrice - result.MarketPayAmount
	default:
		result.MarketPayAmount = originalPrice
	}

	return result, nil
}
