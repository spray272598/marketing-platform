package data

import (
	"context"
	"fmt"
	"time"

	"github.com/marketing-platform/internal/groupbuy/biz"
)

type notifyTaskRepo struct {
	db *Data
}

func NewNotifyTaskRepo(data *Data) biz.NotifyTaskRepo {
	return &notifyTaskRepo{db: data}
}

func (r *notifyTaskRepo) CreateTask(ctx context.Context, task *biz.NotifyTask) error {
	query := `INSERT INTO notify_task (task_id, notify_type, notify_status, notify_url, notify_data, uuid, retry_count, max_retry, next_time) 
			  VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := r.db.db.ExecContext(ctx, query,
		task.TaskID, task.NotifyType, task.NotifyStatus,
		task.NotifyURL, task.NotifyData, task.UUID,
		task.RetryCount, task.MaxRetry, task.NextTime,
	)
	return err
}

func (r *notifyTaskRepo) GetTask(ctx context.Context, taskID string) (*biz.NotifyTask, error) {
	query := `SELECT id, task_id, notify_type, notify_status, notify_url, notify_data, uuid, retry_count, max_retry, next_time 
			  FROM notify_task WHERE task_id = ?`

	task := &biz.NotifyTask{}
	err := r.db.db.QueryRowContext(ctx, query, taskID).Scan(
		&task.ID, &task.TaskID, &task.NotifyType, &task.NotifyStatus,
		&task.NotifyURL, &task.NotifyData, &task.UUID,
		&task.RetryCount, &task.MaxRetry, &task.NextTime,
	)
	if err != nil {
		return nil, fmt.Errorf("task not found: %w", err)
	}
	return task, nil
}

func (r *notifyTaskRepo) GetPendingTasks(ctx context.Context, limit int) ([]*biz.NotifyTask, error) {
	query := `SELECT id, task_id, notify_type, notify_status, notify_url, notify_data, uuid, retry_count, max_retry, next_time 
			  FROM notify_task WHERE notify_status IN (0, 2) AND next_time <= ? 
			  ORDER BY id ASC LIMIT ?`

	rows, err := r.db.db.QueryContext(ctx, query, time.Now().UnixMilli(), limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []*biz.NotifyTask
	for rows.Next() {
		task := &biz.NotifyTask{}
		if err := rows.Scan(
			&task.ID, &task.TaskID, &task.NotifyType, &task.NotifyStatus,
			&task.NotifyURL, &task.NotifyData, &task.UUID,
			&task.RetryCount, &task.MaxRetry, &task.NextTime,
		); err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}
	return tasks, nil
}

func (r *notifyTaskRepo) UpdateTaskStatus(ctx context.Context, taskID string, status int32) error {
	query := `UPDATE notify_task SET notify_status = ?, update_at = NOW() WHERE task_id = ?`
	_, err := r.db.db.ExecContext(ctx, query, status, taskID)
	return err
}

func (r *notifyTaskRepo) UpdateTaskRetry(ctx context.Context, taskID string, retryCount int32, nextTime int64) error {
	query := `UPDATE notify_task SET retry_count = ?, next_time = ?, update_at = NOW() WHERE task_id = ?`
	_, err := r.db.db.ExecContext(ctx, query, retryCount, nextTime, taskID)
	return err
}

func (r *notifyTaskRepo) GetTaskByUUID(ctx context.Context, uuid string) (*biz.NotifyTask, error) {
	query := `SELECT id, task_id, notify_type, notify_status, notify_url, notify_data, uuid, retry_count, max_retry, next_time 
			  FROM notify_task WHERE uuid = ?`

	task := &biz.NotifyTask{}
	err := r.db.db.QueryRowContext(ctx, query, uuid).Scan(
		&task.ID, &task.TaskID, &task.NotifyType, &task.NotifyStatus,
		&task.NotifyURL, &task.NotifyData, &task.UUID,
		&task.RetryCount, &task.MaxRetry, &task.NextTime,
	)
	if err != nil {
		return nil, err
	}
	return task, nil
}
