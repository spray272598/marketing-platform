package biz

import (
	"context"
	"fmt"
)

const (
	StockTypeProduct = "product"
	StockTypeTeam    = "team"
	StockTypePrize   = "prize"
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
	DeductStockAtomic(ctx context.Context, stockKey string, count int32) (bool, error)
	RestoreStockAtomic(ctx context.Context, stockKey string, count int32) (bool, error)
}

type StockService struct {
	stockRepo StockRepo
}

func NewStockService(stockRepo StockRepo) *StockService {
	return &StockService{stockRepo: stockRepo}
}

func (s *StockService) DeductStock(ctx context.Context, stockKey string, count int32) (bool, error) {
	if count <= 0 {
		return false, fmt.Errorf("count must be positive, got %d", count)
	}

	ok, err := s.stockRepo.DeductStockAtomic(ctx, stockKey, count)
	if err != nil {
		return false, fmt.Errorf("deduct stock failed: %w", err)
	}
	if !ok {
		return false, fmt.Errorf("stock not enough: %s", stockKey)
	}
	return true, nil
}

func (s *StockService) GetStock(ctx context.Context, stockKey string) (int32, error) {
	item, err := s.stockRepo.GetStock(ctx, stockKey)
	if err != nil {
		return 0, fmt.Errorf("get stock failed: %w", err)
	}
	if item == nil {
		return 0, fmt.Errorf("stock not found: %s", stockKey)
	}
	return item.Stock, nil
}

func (s *StockService) RestoreStock(ctx context.Context, stockKey string, count int32) error {
	if count <= 0 {
		return fmt.Errorf("count must be positive, got %d", count)
	}
	ok, err := s.stockRepo.RestoreStockAtomic(ctx, stockKey, count)
	if err != nil {
		return fmt.Errorf("restore stock failed: %w", err)
	}
	if !ok {
		return fmt.Errorf("stock not found: %s", stockKey)
	}
	return nil
}
