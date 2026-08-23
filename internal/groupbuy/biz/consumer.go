package biz

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/marketing-platform/pkg/common"
)

type NotifyConsumer struct {
	notifyRepo NotifyTaskRepo
	mqRepo     MQRepo
	interval   time.Duration
	logger     *slog.Logger
	stopCh     chan struct{}
}

func NewNotifyConsumer(notifyRepo NotifyTaskRepo, mqRepo MQRepo) *NotifyConsumer {
	return &NotifyConsumer{
		notifyRepo: notifyRepo,
		mqRepo:     mqRepo,
		interval:   5 * time.Second,
		logger:     slog.Default(),
		stopCh:     make(chan struct{}),
	}
}

func (c *NotifyConsumer) SetLogger(logger *slog.Logger) {
	c.logger = logger
}

func (c *NotifyConsumer) Start(ctx context.Context) {
	c.logger.Info("NotifyConsumer started",
		slog.String("interval", c.interval.String()),
	)

	ticker := time.NewTicker(c.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			c.logger.Info("NotifyConsumer context cancelled, shutting down")
			return
		case <-c.stopCh:
			c.logger.Info("NotifyConsumer stop signal received, shutting down")
			return
		case <-ticker.C:
			c.processPendingTasks(ctx)
		}
	}
}

func (c *NotifyConsumer) Stop() {
	close(c.stopCh)
}

func (c *NotifyConsumer) processPendingTasks(ctx context.Context) {
	tasks, err := c.notifyRepo.GetPendingTasks(ctx, 100)
	if err != nil {
		c.logger.Error("Failed to get pending tasks",
			slog.String("error", err.Error()),
		)
		return
	}

	if len(tasks) == 0 {
		return
	}

	c.logger.Info("Processing pending notify tasks",
		slog.Int("count", len(tasks)),
	)

	for _, task := range tasks {
		if err := c.processTask(ctx, task); err != nil {
			c.logger.Error("Failed to process notify task",
				slog.String("task_id", task.TaskID),
				slog.String("error", err.Error()),
			)
		}
	}
}

func (c *NotifyConsumer) processTask(ctx context.Context, task *NotifyTask) error {
	// 检查是否到达重试时间
	if task.NotifyStatus == common.NotifyStatusInit || task.NotifyStatus == common.NotifyStatusRetry {
		if time.Now().UnixMilli() < task.NextTime {
			return nil
		}
	}

	// 检查是否超过最大重试次数
	if task.RetryCount >= task.MaxRetry {
		c.logger.Warn("Notify task exceeded max retries",
			slog.String("task_id", task.TaskID),
			slog.Int("retry_count", int(task.RetryCount)),
			slog.Int("max_retry", int(task.MaxRetry)),
		)
		return c.notifyRepo.UpdateTaskStatus(ctx, task.TaskID, common.NotifyStatusFailed)
	}

	var err error
	switch task.NotifyType {
	case common.NotifyTypeHTTP:
		err = c.sendHTTPNotify(ctx, task)
	case common.NotifyTypeMQ:
		err = c.sendMQNotify(ctx, task)
	default:
		err = fmt.Errorf("unknown notify type: %s", task.NotifyType)
	}

	if err != nil {
		return c.handleNotifyError(ctx, task, err)
	}

	// 成功 → 标记完成
	c.logger.Info("Notify task completed successfully",
		slog.String("task_id", task.TaskID),
		slog.String("notify_type", task.NotifyType),
	)
	return c.notifyRepo.UpdateTaskStatus(ctx, task.TaskID, common.NotifyStatusSuccess)
}

func (c *NotifyConsumer) sendHTTPNotify(ctx context.Context, task *NotifyTask) error {
	// TODO: 实现 HTTP 回调通知
	return fmt.Errorf("HTTP notify not implemented yet")
}

func (c *NotifyConsumer) sendMQNotify(ctx context.Context, task *NotifyTask) error {
	if c.mqRepo == nil {
		return fmt.Errorf("mqRepo is nil, cannot send MQ notification")
	}

	var data map[string]interface{}
	if err := json.Unmarshal([]byte(task.NotifyData), &data); err != nil {
		return fmt.Errorf("failed to unmarshal notify data: %w", err)
	}

	if teamID, ok := data["team_id"].(string); ok {
		c.logger.Info("Sending team success MQ notification",
			slog.String("task_id", task.TaskID),
			slog.String("team_id", teamID),
		)
		return c.mqRepo.PublishTeamSuccessMessage(ctx, teamID)
	}

	if orderID, ok := data["order_id"].(string); ok {
		c.logger.Info("Sending refund MQ notification",
			slog.String("task_id", task.TaskID),
			slog.String("order_id", orderID),
		)
		return c.mqRepo.PublishRefundMessage(ctx, orderID)
	}

	return fmt.Errorf("unknown MQ notify data format for task: %s", task.TaskID)
}

func (c *NotifyConsumer) handleNotifyError(ctx context.Context, task *NotifyTask, err error) error {
	newRetryCount := task.RetryCount + 1

	// 指数退避: 1s, 2s, 4s, 8s...
	backoff := time.Duration(1<<newRetryCount) * time.Second
	nextTime := time.Now().Add(backoff).UnixMilli()

	c.logger.Warn("Notify task failed, scheduling retry",
		slog.String("task_id", task.TaskID),
		slog.Int("retry_count", int(newRetryCount)),
		slog.Duration("backoff", backoff),
		slog.String("error", err.Error()),
	)

	if newRetryCount >= task.MaxRetry {
		return c.notifyRepo.UpdateTaskStatus(ctx, task.TaskID, common.NotifyStatusFailed)
	}

	return c.notifyRepo.UpdateTaskRetry(ctx, task.TaskID, newRetryCount, nextTime)
}