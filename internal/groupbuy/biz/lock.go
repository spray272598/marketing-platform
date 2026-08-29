package biz

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/marketing-platform/pkg/common"
)

type LockService struct {
	activityRepo ActivityRepo
	orderRepo    OrderRepo
	teamRepo     TeamRepo
	redisRepo    RedisRepo
}

func NewLockService(
	activityRepo ActivityRepo,
	orderRepo OrderRepo,
	teamRepo TeamRepo,
	redisRepo RedisRepo,
) *LockService {
	return &LockService{
		activityRepo: activityRepo,
		orderRepo:    orderRepo,
		teamRepo:     teamRepo,
		redisRepo:    redisRepo,
	}
}

func (s *LockService) LockOrder(ctx context.Context, activityID string, userID int64, channel, source string) (*GroupBuyOrder, error) {
	activity, err := s.activityRepo.GetActivity(ctx, activityID)
	if err != nil {
		return nil, fmt.Errorf(common.GroupBuyActivityNotExist.Code+": %w", err)
	}

	if activity.ActivityState != common.ActivityStateOpen {
		return nil, fmt.Errorf("%s: activity is not open", common.GroupBuyActivityNotExist.Code)
	}

	lockKey := fmt.Sprintf("groupbuy:lock:%d:%s", userID, activityID)
	lockValue := uuid.New().String()
	locked, err := s.redisRepo.LockOrder(ctx, lockKey, lockValue)
	if err != nil || !locked {
		return nil, fmt.Errorf("lock failed")
	}
	defer s.redisRepo.UnlockOrder(ctx, lockKey, lockValue)

	order := &GroupBuyOrder{
		OrderID:    fmt.Sprintf("gb_%s", uuid.New().String()[:12]),
		TeamID:     fmt.Sprintf("team_%s", uuid.New().String()[:12]),
		UserID:     userID,
		ActivityID: activity.ActivityID,
		BizID:      fmt.Sprintf("%d_%s", userID, activity.ActivityID),
		OrderState: common.OrderStateInit,
		CreateAt:   time.Now().Format("2006-01-02 15:04:05"),
	}

	if err := s.orderRepo.CreateOrder(ctx, order); err != nil {
		return nil, fmt.Errorf("create order failed: %w", err)
	}

	team := &GroupBuyTeam{
		TeamID:        order.TeamID,
		ActivityID:    activity.ActivityID,
		TargetCount:   activity.TargetCount,
		CompleteCount: 1,
		LockCount:     1,
		TeamState:     common.TeamStateBuilding,
	}
	if err := s.teamRepo.CreateTeam(ctx, team); err != nil {
		_ = s.orderRepo.UpdateOrderState(ctx, order.OrderID, common.OrderStateCancelled)
		return nil, fmt.Errorf("create team failed: %w", err)
	}

	return order, nil
}
