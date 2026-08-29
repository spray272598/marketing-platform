package biz

import (
	"context"
	"fmt"
	"sync"
)

type mockActivityRepo struct {
	activities map[string]*SeckillActivity
}

func newMockActivityRepo() *mockActivityRepo {
	return &mockActivityRepo{
		activities: make(map[string]*SeckillActivity),
	}
}

func (m *mockActivityRepo) GetActivity(ctx context.Context, activityID string) (*SeckillActivity, error) {
	if a, ok := m.activities[activityID]; ok {
		return a, nil
	}
	return nil, fmt.Errorf("activity not found")
}

func (m *mockActivityRepo) UpdateActivityStock(ctx context.Context, activityID string, stock int32) error {
	if a, ok := m.activities[activityID]; ok {
		a.TotalCount = stock
		return nil
	}
	return fmt.Errorf("activity not found")
}

type mockOrderRepo struct {
	mu     sync.RWMutex
	orders map[string]*SeckillOrder
}

func newMockOrderRepo() *mockOrderRepo {
	return &mockOrderRepo{
		orders: make(map[string]*SeckillOrder),
	}
}

func (m *mockOrderRepo) CreateOrder(ctx context.Context, order *SeckillOrder) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.orders[order.OrderID] = order
	return nil
}

func (m *mockOrderRepo) GetOrder(ctx context.Context, orderID string) (*SeckillOrder, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if o, ok := m.orders[orderID]; ok {
		return o, nil
	}
	return nil, fmt.Errorf("order not found")
}

func (m *mockOrderRepo) GetUserActivityOrder(ctx context.Context, userID int64, activityID string) (*SeckillOrder, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, o := range m.orders {
		if o.UserID == userID && o.ActivityID == activityID {
			return o, nil
		}
	}
	return nil, fmt.Errorf("order not found")
}

func (m *mockOrderRepo) UpdateOrderState(ctx context.Context, orderID string, state int32) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if o, ok := m.orders[orderID]; ok {
		o.OrderState = state
		return nil
	}
	return fmt.Errorf("order not found")
}

type mockRedisRepo struct {
	mu       sync.RWMutex
	stocks   map[string]int32
	userSets map[string]map[int64]bool
}

func newMockRedisRepo() *mockRedisRepo {
	return &mockRedisRepo{
		stocks:   make(map[string]int32),
		userSets: make(map[string]map[int64]bool),
	}
}

func (m *mockRedisRepo) GetStock(ctx context.Context, activityID string) (int32, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.stocks[activityID], nil
}

func (m *mockRedisRepo) DecrStock(ctx context.Context, activityID string) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if stock, ok := m.stocks[activityID]; ok && stock > 0 {
		m.stocks[activityID]--
		return true, nil
	}
	return false, nil
}

func (m *mockRedisRepo) SetStock(ctx context.Context, activityID string, stock int32) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.stocks[activityID] = stock
	return nil
}

func (m *mockRedisRepo) DecrStockWithUserCheck(ctx context.Context, activityID string, userID int64, limit int32) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 1. 校验用户限购次数（limit<=0 视为不限）
	userKey := activityID
	if m.userSets[userKey] == nil {
		m.userSets[userKey] = make(map[int64]bool)
	}
	if limit > 0 {
		cnt := int32(0)
		for u := range m.userSets[userKey] {
			if u == userID {
				cnt++
			}
		}
		if cnt >= limit {
			return 2, nil // 超过限购
		}
	}

	// 2. 检查库存
	if stock, ok := m.stocks[activityID]; !ok || stock <= 0 {
		return 0, nil // 库存不足
	}

	// 3. 扣减库存
	m.stocks[activityID]--

	// 4. 记录用户购买
	m.userSets[userKey][userID] = true

	return 1, nil // 成功
}

func (m *mockRedisRepo) IncrStock(ctx context.Context, activityID string, count int32) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.stocks[activityID] += count
	return nil
}

type mockMQRepo struct {
	mu       sync.RWMutex
	messages []*SeckillOrder
}

func newMockMQRepo() *mockMQRepo {
	return &mockMQRepo{
		messages: make([]*SeckillOrder, 0),
	}
}

func (m *mockMQRepo) PublishOrderMessage(ctx context.Context, order *SeckillOrder) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.messages = append(m.messages, order)
	return nil
}

type mockStockClient struct{}

func (m *mockStockClient) DeductStock(ctx context.Context, stockKey string, count int32) error {
	return nil
}

func (m *mockStockClient) RestoreStock(ctx context.Context, stockKey string, count int32) error {
	return nil
}