package biz

import (
	"context"
	"errors"
)

// ErrTeamNotSettleable 表示团队当前状态不允许继续结算：
// 团队不存在、已结束（成功/失败）或完成人数已满。
// 仓储层用它让上层（Saga）明确感知"本次没有实际推进"，从而触发补偿。
var ErrTeamNotSettleable = errors.New("team is not in a settleable state")

type ActivityRepo interface {
	GetActivity(ctx context.Context, activityID string) (*GroupBuyActivity, error)
	GetDiscount(ctx context.Context, discountID string) (*GroupBuyDiscount, error)
}

type OrderRepo interface {
	CreateOrder(ctx context.Context, order *GroupBuyOrder) error
	GetOrder(ctx context.Context, orderID string) (*GroupBuyOrder, error)
	UpdateOrderState(ctx context.Context, orderID string, state int32) error
	// NextOrderID 基于号段模式返回当前服务的下一个订单号（单调递增）。
	NextOrderID(ctx context.Context, bizTag string) (int64, error)
}

type TeamRepo interface {
	CreateTeam(ctx context.Context, team *GroupBuyTeam) error
	GetTeam(ctx context.Context, teamID string) (*GroupBuyTeam, error)

	// CompleteTeam 原子地把完成数 +1；达到目标人数时把团队状态流转为 successState。
	// 返回 (最新完成数, 是否本次触发成团, error)。
	//
	// "本次触发成团"仅在本次调用真正把状态从"进行中"改为"成团"时为真，
	// 用于保证成团通知只创建一次——重复结算不会重复发奖/通知（幂等）。
	CompleteTeam(ctx context.Context, teamID string, targetCount, successState int32) (int32, bool, error)

	// RollbackTeamComplete 回滚完成数与团队状态，供 Saga 补偿使用。
	// 回滚后团队回到"进行中"，后续重试可以重新成团并补发通知。
	RollbackTeamComplete(ctx context.Context, teamID string, buildingState int32) error
}

type RedisRepo interface {
	LockOrder(ctx context.Context, orderKey string, lockValue string) (bool, error)
	UnlockOrder(ctx context.Context, orderKey string, lockValue string) error
}

type MQRepo interface {
	PublishTeamSuccessMessage(ctx context.Context, teamID string) error
	PublishRefundMessage(ctx context.Context, orderID string) error
}

type NotifyTaskRepo interface {
	CreateTask(ctx context.Context, task *NotifyTask) error
	GetTask(ctx context.Context, taskID string) (*NotifyTask, error)
	GetPendingTasks(ctx context.Context, limit int) ([]*NotifyTask, error)
	UpdateTaskStatus(ctx context.Context, taskID string, status int32) error
	UpdateTaskRetry(ctx context.Context, taskID string, retryCount int32, nextTime int64) error
	GetTaskByUUID(ctx context.Context, uuid string) (*NotifyTask, error)
}
