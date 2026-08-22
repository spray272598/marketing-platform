package biz

import (
	"context"
	"encoding/json"
	"fmt"
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
	dataBytes, _ := json.Marshal(data)
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
	dataBytes, _ := json.Marshal(data)
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
			s.handleTaskError(ctx, task, err)
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
	// TODO: Implement HTTP callback
	return s.notifyRepo.UpdateTaskStatus(ctx, task.TaskID, NotifyStatusSuccess)
}

func (s *NotifyService) handleTaskError(ctx context.Context, task *NotifyTask, err error) {
	task.RetryCount++
	if task.RetryCount >= task.MaxRetry {
		s.notifyRepo.UpdateTaskStatus(ctx, task.TaskID, NotifyStatusFailed)
		return
	}

	nextTime := time.Now().Add(time.Duration(task.RetryCount*task.RetryCount) * time.Minute).UnixMilli()
	s.notifyRepo.UpdateTaskRetry(ctx, task.TaskID, task.RetryCount, nextTime)
}
