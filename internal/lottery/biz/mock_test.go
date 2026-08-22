package biz

import (
	"context"
	"fmt"
	"sync"
)

type mockLTActivityRepo struct {
	activities map[string]*LotteryActivity
}

func newMockLTActivityRepo() *mockLTActivityRepo {
	return &mockLTActivityRepo{
		activities: make(map[string]*LotteryActivity),
	}
}

func (m *mockLTActivityRepo) GetActivity(ctx context.Context, activityID string) (*LotteryActivity, error) {
	if a, ok := m.activities[activityID]; ok {
		return a, nil
	}
	return nil, fmt.Errorf("activity not found")
}

type mockStrategyRepo struct {
	mu      sync.RWMutex
	strategies map[string]*LotteryStrategy
	awards  map[string][]*StrategyAward
}

func newMockStrategyRepo() *mockStrategyRepo {
	return &mockStrategyRepo{
		strategies: make(map[string]*LotteryStrategy),
		awards:     make(map[string][]*StrategyAward),
	}
}

func (m *mockStrategyRepo) GetStrategy(ctx context.Context, strategyID string) (*LotteryStrategy, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if s, ok := m.strategies[strategyID]; ok {
		return s, nil
	}
	return nil, fmt.Errorf("strategy not found")
}

func (m *mockStrategyRepo) GetStrategyAwards(ctx context.Context, strategyID string) ([]*StrategyAward, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if a, ok := m.awards[strategyID]; ok {
		return a, nil
	}
	return nil, fmt.Errorf("awards not found")
}

type mockLTOrderRepo struct {
	mu     sync.RWMutex
	orders map[string]*LotteryOrder
	counts map[string]int32
}

func newMockLTOrderRepo() *mockLTOrderRepo {
	return &mockLTOrderRepo{
		orders: make(map[string]*LotteryOrder),
		counts: make(map[string]int32),
	}
}

func (m *mockLTOrderRepo) CreateOrder(ctx context.Context, order *LotteryOrder) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.orders[order.OrderID] = order
	key := fmt.Sprintf("%d_%s", order.UserID, order.ActivityID)
	m.counts[key]++
	return nil
}

func (m *mockLTOrderRepo) GetUserActivityCount(ctx context.Context, userID int64, activityID string) (int32, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	key := fmt.Sprintf("%d_%s", userID, activityID)
	return m.counts[key], nil
}
