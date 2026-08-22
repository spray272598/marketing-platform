package biz

import (
	"context"
	"fmt"
	"time"

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

	lockKey := fmt.Sprintf("groupbuy:lock:%d:%s", userID, activityID)
	locked, err := s.redisRepo.LockOrder(ctx, lockKey)
	if err != nil || !locked {
		return nil, fmt.Errorf("lock failed")
	}
	defer s.redisRepo.UnlockOrder(ctx, lockKey)

	order := &GroupBuyOrder{
		OrderID:    fmt.Sprintf("gb_%d_%d", time.Now().UnixMilli(), userID),
		TeamID:     fmt.Sprintf("team_%s_%d", activity.ActivityID, time.Now().UnixMilli()),
		UserID:     userID,
		ActivityID: activity.ActivityID,
		BizID:      fmt.Sprintf("%d_%s", userID, activity.ActivityID),
		OrderState: common.OrderStateInit,
		CreateAt:   time.Now().Format("2006-01-02 15:04:05"),
	}

	if err := s.orderRepo.CreateOrder(ctx, order); err != nil {
		return nil, err
	}

	team := &GroupBuyTeam{
		TeamID:        order.TeamID,
		ActivityID:    activity.ActivityID,
		TargetCount:   activity.TargetCount,
		CompleteCount: 1,
		LockCount:     1,
		TeamState:     common.TeamStateBuilding,
	}
	_ = s.teamRepo.CreateTeam(ctx, team)

	return order, nil
}
