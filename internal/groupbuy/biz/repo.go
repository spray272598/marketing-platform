package biz

import "context"

type ActivityRepo interface {
	GetActivity(ctx context.Context, activityID string) (*GroupBuyActivity, error)
	GetDiscount(ctx context.Context, discountID string) (*GroupBuyDiscount, error)
}

type OrderRepo interface {
	CreateOrder(ctx context.Context, order *GroupBuyOrder) error
	GetOrder(ctx context.Context, orderID string) (*GroupBuyOrder, error)
	UpdateOrderState(ctx context.Context, orderID string, state int32) error
}

type TeamRepo interface {
	CreateTeam(ctx context.Context, team *GroupBuyTeam) error
	GetTeam(ctx context.Context, teamID string) (*GroupBuyTeam, error)
	IncrementTeamComplete(ctx context.Context, teamID string) (int32, error)
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
