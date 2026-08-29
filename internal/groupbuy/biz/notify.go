package biz

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
)

type NotifyService struct {
	notifyRepo NotifyTaskRepo
	mqRepo     MQRepo
}

func NewNotifyService(notifyRepo NotifyTaskRepo, mqRepo MQRepo) *NotifyService {
	return &NotifyService{
		notifyRepo: notifyRepo,
		mqRepo:     mqRepo,
	}
}

func (s *NotifyService) CreateTeamSuccessNotify(ctx context.Context, teamID string, data interface{}) error {
	dataBytes, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("marshal notify data: %w", err)
	}
	taskID := fmt.Sprintf("team_success_%s_%d", teamID, time.Now().UnixMilli())
	uuidVal := uuid.New().String()

	task := &NotifyTask{
		TaskID:       taskID,
		NotifyType:   NotifyTypeMQ,
		NotifyStatus: NotifyStatusInit,
		NotifyData:   string(dataBytes),
		UUID:         uuidVal,
		RetryCount:   0,
		MaxRetry:     3,
		NextTime:     time.Now().UnixMilli(),
	}

	return s.notifyRepo.CreateTask(ctx, task)
}

func (s *NotifyService) CreateRefundNotify(ctx context.Context, orderID string, data interface{}) error {
	dataBytes, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("marshal notify data: %w", err)
	}
	taskID := fmt.Sprintf("refund_%s_%d", orderID, time.Now().UnixMilli())
	uuidVal := uuid.New().String()

	task := &NotifyTask{
		TaskID:       taskID,
		NotifyType:   NotifyTypeMQ,
		NotifyStatus: NotifyStatusInit,
		NotifyData:   string(dataBytes),
		UUID:         uuidVal,
		RetryCount:   0,
		MaxRetry:     3,
		NextTime:     time.Now().UnixMilli(),
	}

	return s.notifyRepo.CreateTask(ctx, task)
}

func (s *NotifyService) ProcessPendingTasks(ctx context.Context) error {
	tasks, err := s.notifyRepo.GetPendingTasks(ctx, 100)
	if err != nil {
		return err
	}

	for _, task := range tasks {
		if err := s.processTask(ctx, task); err != nil {
			// 更新重试状态本身也可能失败，必须记录，否则任务会静默卡在中间态。
			if herr := s.handleTaskError(ctx, task, err); herr != nil {
				slog.Error("notify task failed and retry state could not be persisted",
					slog.String("task_id", task.TaskID),
					slog.Any("task_error", err),
					slog.Any("update_error", herr),
				)
			}
		}
	}
	return nil
}

func (s *NotifyService) processTask(ctx context.Context, task *NotifyTask) error {
	switch task.NotifyType {
	case NotifyTypeMQ:
		return s.processMQTask(ctx, task)
	case NotifyTypeHTTP:
		return s.processHTTPTask(ctx, task)
	default:
		return fmt.Errorf("unknown notify type: %s", task.NotifyType)
	}
}

func (s *NotifyService) processMQTask(ctx context.Context, task *NotifyTask) error {
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(task.NotifyData), &data); err != nil {
		return err
	}

	if teamID, ok := data["team_id"].(string); ok {
		if err := s.mqRepo.PublishTeamSuccessMessage(ctx, teamID); err != nil {
			return err
		}
	} else if orderID, ok := data["order_id"].(string); ok {
		if err := s.mqRepo.PublishRefundMessage(ctx, orderID); err != nil {
			return err
		}
	}

	return s.notifyRepo.UpdateTaskStatus(ctx, task.TaskID, NotifyStatusSuccess)
}

func (s *NotifyService) processHTTPTask(ctx context.Context, task *NotifyTask) error {
	// HTTP 回调尚未实现。这里必须返回错误而不是直接标记成功，
	// 否则通知会被"假装送达"而静默丢失；返回错误可让任务进入重试/失败流程并留下痕迹。
	return fmt.Errorf("http notify not implemented for task %s", task.TaskID)
}

// handleTaskError 记录一次失败并推进重试状态，返回仓储操作的错误，
// 避免"更新重试状态失败"被静默吞掉导致任务永远卡住。
func (s *NotifyService) handleTaskError(ctx context.Context, task *NotifyTask, err error) error {
	task.RetryCount++
	if task.RetryCount >= task.MaxRetry {
		return s.notifyRepo.UpdateTaskStatus(ctx, task.TaskID, NotifyStatusFailed)
	}

	// 指数退避：1, 4, 9... 分钟。限制上限，避免重试次数很大时溢出。
	backoffMin := task.RetryCount * task.RetryCount
	if backoffMin > 60 {
		backoffMin = 60
	}
	nextTime := time.Now().Add(time.Duration(backoffMin) * time.Minute).UnixMilli()
	return s.notifyRepo.UpdateTaskRetry(ctx, task.TaskID, task.RetryCount, nextTime)
}
