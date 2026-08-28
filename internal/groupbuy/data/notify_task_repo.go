package data

import (
	"context"
	"fmt"

	"github.com/marketing-platform/internal/groupbuy/biz"
	"github.com/marketing-platform/internal/groupbuy/data/ent"
	"github.com/marketing-platform/internal/groupbuy/data/ent/notifytask"
	"time"
)

type notifyTaskRepo struct {
	data *Data
}

func NewNotifyTaskRepo(data *Data) biz.NotifyTaskRepo {
	return &notifyTaskRepo{data: data}
}

func (r *notifyTaskRepo) CreateTask(ctx context.Context, task *biz.NotifyTask) error {
	_, err := r.data.db.NotifyTask.Create().
		SetTaskID(task.TaskID).
		SetNotifyType(task.NotifyType).
		SetNotifyStatus(task.NotifyStatus).
		SetNillableNotifyURL(&task.NotifyURL).
		SetNillableNotifyData(&task.NotifyData).
		SetUUID(task.UUID).
		SetRetryCount(task.RetryCount).
		SetMaxRetry(task.MaxRetry).
		SetNextTime(task.NextTime).
		Save(ctx)
	return err
}

func (r *notifyTaskRepo) GetTask(ctx context.Context, taskID string) (*biz.NotifyTask, error) {
	po, err := r.data.db.NotifyTask.Query().
		Where(notifytask.TaskIDEQ(taskID)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("task not found: %s", taskID)
		}
		return nil, err
	}
	return toBizNotifyTask(po), nil
}

func (r *notifyTaskRepo) GetPendingTasks(ctx context.Context, limit int) ([]*biz.NotifyTask, error) {
	now := time.Now().UnixMilli()
	pos, err := r.data.db.NotifyTask.Query().
		Where(
			notifytask.NotifyStatusIn(0, 2),
			notifytask.NextTimeLTE(now),
		).
		Order(ent.Asc(notifytask.FieldID)).
		Limit(limit).
		All(ctx)
	if err != nil {
		return nil, err
	}
	tasks := make([]*biz.NotifyTask, 0, len(pos))
	for _, po := range pos {
		tasks = append(tasks, toBizNotifyTask(po))
	}
	return tasks, nil
}

func (r *notifyTaskRepo) UpdateTaskStatus(ctx context.Context, taskID string, status int32) error {
	_, err := r.data.db.NotifyTask.Update().
		Where(notifytask.TaskIDEQ(taskID)).
		SetNotifyStatus(status).
		Save(ctx)
	return err
}

func (r *notifyTaskRepo) UpdateTaskRetry(ctx context.Context, taskID string, retryCount int32, nextTime int64) error {
	_, err := r.data.db.NotifyTask.Update().
		Where(notifytask.TaskIDEQ(taskID)).
		SetRetryCount(retryCount).
		SetNextTime(nextTime).
		Save(ctx)
	return err
}

func (r *notifyTaskRepo) GetTaskByUUID(ctx context.Context, uuid string) (*biz.NotifyTask, error) {
	po, err := r.data.db.NotifyTask.Query().
		Where(notifytask.UUIDEQ(uuid)).
		Only(ctx)
	if err != nil {
		return nil, err
	}
	return toBizNotifyTask(po), nil
}

func toBizNotifyTask(po *ent.NotifyTask) *biz.NotifyTask {
	if po == nil {
		return nil
	}
	task := &biz.NotifyTask{
		TaskID:       po.TaskID,
		NotifyType:   po.NotifyType,
		NotifyStatus: po.NotifyStatus,
		UUID:         po.UUID,
		RetryCount:   po.RetryCount,
		MaxRetry:     po.MaxRetry,
		NextTime:     po.NextTime,
	}
	if po.NotifyURL != nil {
		task.NotifyURL = *po.NotifyURL
	}
	if po.NotifyData != nil {
		task.NotifyData = *po.NotifyData
	}
	return task
}
