package biz

import (
	"context"
	"fmt"
	"sync"
)

type mockGBActivityRepo struct {
	activities map[string]*GroupBuyActivity
	discounts  map[string]*GroupBuyDiscount
}

func newMockGBActivityRepo() *mockGBActivityRepo {
	return &mockGBActivityRepo{
		activities: make(map[string]*GroupBuyActivity),
		discounts:  make(map[string]*GroupBuyDiscount),
	}
}

func (m *mockGBActivityRepo) GetActivity(ctx context.Context, activityID string) (*GroupBuyActivity, error) {
	if a, ok := m.activities[activityID]; ok {
		return a, nil
	}
	return nil, fmt.Errorf("activity not found")
}

func (m *mockGBActivityRepo) GetDiscount(ctx context.Context, discountID string) (*GroupBuyDiscount, error) {
	if d, ok := m.discounts[discountID]; ok {
		return d, nil
	}
	return nil, fmt.Errorf("discount not found")
}

type mockGBOrderRepo struct {
	mu     sync.RWMutex
	orders map[string]*GroupBuyOrder
}

func newMockGBOrderRepo() *mockGBOrderRepo {
	return &mockGBOrderRepo{
		orders: make(map[string]*GroupBuyOrder),
	}
}

func (m *mockGBOrderRepo) CreateOrder(ctx context.Context, order *GroupBuyOrder) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.orders[order.OrderID] = order
	return nil
}

func (m *mockGBOrderRepo) GetOrder(ctx context.Context, orderID string) (*GroupBuyOrder, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if o, ok := m.orders[orderID]; ok {
		return o, nil
	}
	return nil, fmt.Errorf("order not found")
}

func (m *mockGBOrderRepo) UpdateOrderState(ctx context.Context, orderID string, state int32) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if o, ok := m.orders[orderID]; ok {
		o.OrderState = state
		return nil
	}
	return fmt.Errorf("order not found")
}

type mockTeamRepo struct {
	mu     sync.RWMutex
	teams  map[string]*GroupBuyTeam
}

func newMockTeamRepo() *mockTeamRepo {
	return &mockTeamRepo{
		teams: make(map[string]*GroupBuyTeam),
	}
}

func (m *mockTeamRepo) CreateTeam(ctx context.Context, team *GroupBuyTeam) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.teams[team.TeamID] = team
	return nil
}

func (m *mockTeamRepo) GetTeam(ctx context.Context, teamID string) (*GroupBuyTeam, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if t, ok := m.teams[teamID]; ok {
		return t, nil
	}
	return nil, fmt.Errorf("team not found")
}

func (m *mockTeamRepo) IncrementTeamComplete(ctx context.Context, teamID string) (int32, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if t, ok := m.teams[teamID]; ok {
		t.CompleteCount++
		return t.CompleteCount, nil
	}
	return 0, fmt.Errorf("team not found")
}

type mockGBRedisRepo struct {
	mu     sync.RWMutex
	locked map[string]bool
}

func newMockGBRedisRepo() *mockGBRedisRepo {
	return &mockGBRedisRepo{
		locked: make(map[string]bool),
	}
}

func (m *mockGBRedisRepo) LockOrder(ctx context.Context, orderKey string, lockValue string) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.locked[orderKey] {
		return false, nil
	}
	m.locked[orderKey] = true
	return true, nil
}

func (m *mockGBRedisRepo) UnlockOrder(ctx context.Context, orderKey string, lockValue string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.locked, orderKey)
	return nil
}

type mockGBMQRepo struct {
	mu      sync.RWMutex
	messages []string
}

func newMockGBMQRepo() *mockGBMQRepo {
	return &mockGBMQRepo{
		messages: make([]string, 0),
	}
}

func (m *mockGBMQRepo) PublishTeamSuccessMessage(ctx context.Context, teamID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.messages = append(m.messages, "team:"+teamID)
	return nil
}

func (m *mockGBMQRepo) PublishRefundMessage(ctx context.Context, orderID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.messages = append(m.messages, "refund:"+orderID)
	return nil
}

type mockNotifyTaskRepo struct {
	mu    sync.RWMutex
	tasks map[string]*NotifyTask
}

func newMockNotifyTaskRepo() *mockNotifyTaskRepo {
	return &mockNotifyTaskRepo{
		tasks: make(map[string]*NotifyTask),
	}
}

func (m *mockNotifyTaskRepo) CreateTask(ctx context.Context, task *NotifyTask) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.tasks[task.TaskID] = task
	return nil
}

func (m *mockNotifyTaskRepo) GetTask(ctx context.Context, taskID string) (*NotifyTask, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if t, ok := m.tasks[taskID]; ok {
		return t, nil
	}
	return nil, fmt.Errorf("task not found")
}

func (m *mockNotifyTaskRepo) GetPendingTasks(ctx context.Context, limit int) ([]*NotifyTask, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var tasks []*NotifyTask
	for _, t := range m.tasks {
		if t.NotifyStatus == 0 || t.NotifyStatus == 2 {
			tasks = append(tasks, t)
			if len(tasks) >= limit {
				break
			}
		}
	}
	return tasks, nil
}

func (m *mockNotifyTaskRepo) UpdateTaskStatus(ctx context.Context, taskID string, status int32) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if t, ok := m.tasks[taskID]; ok {
		t.NotifyStatus = status
		return nil
	}
	return fmt.Errorf("task not found")
}

func (m *mockNotifyTaskRepo) UpdateTaskRetry(ctx context.Context, taskID string, retryCount int32, nextTime int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if t, ok := m.tasks[taskID]; ok {
		t.RetryCount = retryCount
		t.NextTime = nextTime
		return nil
	}
	return fmt.Errorf("task not found")
}

func (m *mockNotifyTaskRepo) GetTaskByUUID(ctx context.Context, uuid string) (*NotifyTask, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, t := range m.tasks {
		if t.UUID == uuid {
			return t, nil
		}
	}
	return nil, fmt.Errorf("task not found")
}

type mockGBStockClient struct{}

func (m *mockGBStockClient) DeductStock(ctx context.Context, stockKey string, count int32) error {
	return nil
}

func (m *mockGBStockClient) RestoreStock(ctx context.Context, stockKey string, count int32) error {
	return nil
}
