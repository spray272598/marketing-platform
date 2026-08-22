package biz

import (
	"context"
	"fmt"
	"sync"
)

const (
	StockTypeProduct = "product" // 商品库存
	StockTypeTeam    = "team"    // 团位库存
	StockTypePrize   = "prize"   // 奖品库存
)

type StockItem struct {
	ID        int64  `json:"id"`
	StockKey  string `json:"stock_key"`
	StockName string `json:"stock_name"`
	StockType string `json:"stock_type"`
	Stock     int32  `json:"stock"`
	Total     int32  `json:"total"`
}

type StockRepo interface {
	GetStock(ctx context.Context, stockKey string) (*StockItem, error)
	UpdateStock(ctx context.Context, stockKey string, stock int32) error
}

type StockService struct {
	stockRepo StockRepo
	mu        sync.RWMutex
}

func NewStockService(stockRepo StockRepo) *StockService {
	return &StockService{stockRepo: stockRepo}
}

func (s *StockService) DeductStock(ctx context.Context, stockKey string, count int32) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	item, err := s.stockRepo.GetStock(ctx, stockKey)
	if err != nil {
		return false, fmt.Errorf("stock not found: %w", err)
	}

	if item.Stock < count {
		return false, fmt.Errorf("stock not enough: %d < %d", item.Stock, count)
	}

	newStock := item.Stock - count
	if err := s.stockRepo.UpdateStock(ctx, stockKey, newStock); err != nil {
		return false, err
	}

	return true, nil
}

func (s *StockService) GetStock(ctx context.Context, stockKey string) (int32, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	item, err := s.stockRepo.GetStock(ctx, stockKey)
	if err != nil {
		return 0, err
	}

	return item.Stock, nil
}

func (s *StockService) RestoreStock(ctx context.Context, stockKey string, count int32) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	item, err := s.stockRepo.GetStock(ctx, stockKey)
	if err != nil {
		return fmt.Errorf("stock not found: %w", err)
	}

	newStock := item.Stock + count
	return s.stockRepo.UpdateStock(ctx, stockKey, newStock)
}
