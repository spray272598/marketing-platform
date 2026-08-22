package biz

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/marketing-platform/pkg/common"
)

type NotifyConsumer struct {
	notifyRepo NotifyTaskRepo
	interval   time.Duration
	client     *http.Client
}

func NewNotifyConsumer(notifyRepo NotifyTaskRepo) *NotifyConsumer {
	return &NotifyConsumer{
		notifyRepo: notifyRepo,
		interval:   5 * time.Second,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *NotifyConsumer) Start(ctx context.Context) {
	ticker := time.NewTicker(c.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			c.processPendingTasks(ctx)
		}
	}
}

func (c *NotifyConsumer) processPendingTasks(ctx context.Context) {
	tasks, err := c.notifyRepo.GetPendingTasks(ctx, 100)
	if err != nil {
		return
	}

	for _, task := range tasks {
		c.processTask(ctx, task)
	}
}

func (c *NotifyConsumer) processTask(ctx context.Context, task *NotifyTask) {
	if task.NotifyStatus == common.NotifyStatusInit || task.NotifyStatus == common.NotifyStatusRetry {
		if time.Now().UnixMilli() < task.NextTime {
			return
		}
	}

	if task.RetryCount >= task.MaxRetry {
		c.notifyRepo.UpdateTaskStatus(ctx, task.TaskID, common.NotifyStatusFailed)
		return
	}

	var err error
	switch task.NotifyType {
	case common.NotifyTypeHTTP:
		err = c.sendHTTPNotify(ctx, task)
	case common.NotifyTypeMQ:
		err = c.sendMQNotify(ctx, task)
	}

	if err != nil {
		c.handleNotifyError(ctx, task, err)
		return
	}

	c.notifyRepo.UpdateTaskStatus(ctx, task.TaskID, common.NotifyStatusSuccess)
}

func (c *NotifyConsumer) sendHTTPNotify(ctx context.Context, task *NotifyTask) error {
	req, err := http.NewRequestWithContext(ctx, "POST", task.NotifyURL,
		bytes.NewBufferString(task.NotifyData))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("http notify failed with status: %d", resp.StatusCode)
	}

	return nil
}

func (c *NotifyConsumer) sendMQNotify(ctx context.Context, task *NotifyTask) error {
	data, _ := json.Marshal(task.NotifyData)
	fmt.Printf("MQ notify: %s\n", string(data))
	return nil
}

func (c *NotifyConsumer) handleNotifyError(ctx context.Context, task *NotifyTask, err error) {
	newRetryCount := task.RetryCount + 1
	backoff := time.Duration(1<<newRetryCount) * time.Second
	nextTime := time.Now().Add(backoff).UnixMilli()

	if newRetryCount >= task.MaxRetry {
		c.notifyRepo.UpdateTaskStatus(ctx, task.TaskID, common.NotifyStatusFailed)
	} else {
		c.notifyRepo.UpdateTaskRetry(ctx, task.TaskID, newRetryCount, nextTime)
	}
}
