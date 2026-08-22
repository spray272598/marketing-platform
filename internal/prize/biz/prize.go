package biz

import (
	"context"
	"fmt"
	"sync"
)

type PrizeItem struct {
	ID        int64  `json:"id"`
	PrizeID   string `json:"prize_id"`
	PrizeName string `json:"prize_name"`
	Stock     int32  `json:"stock"`
	Total     int32  `json:"total"`
}

type PrizeRepo interface {
	GetPrize(ctx context.Context, prizeID string) (*PrizeItem, error)
	UpdateStock(ctx context.Context, prizeID string, stock int32) error
}

type PrizeService struct {
	prizeRepo PrizeRepo
	mu        sync.RWMutex
}

func NewPrizeService(prizeRepo PrizeRepo) *PrizeService {
	return &PrizeService{
		prizeRepo: prizeRepo,
	}
}

func (s *PrizeService) DeductStock(ctx context.Context, prizeID string, count int32) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	prize, err := s.prizeRepo.GetPrize(ctx, prizeID)
	if err != nil {
		return false, fmt.Errorf("prize not found: %w", err)
	}

	if prize.Stock < count {
		return false, fmt.Errorf("stock not enough: %d < %d", prize.Stock, count)
	}

	newStock := prize.Stock - count
	if err := s.prizeRepo.UpdateStock(ctx, prizeID, newStock); err != nil {
		return false, err
	}

	return true, nil
}

func (s *PrizeService) GetStock(ctx context.Context, prizeID string) (int32, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	prize, err := s.prizeRepo.GetPrize(ctx, prizeID)
	if err != nil {
		return 0, err
	}

	return prize.Stock, nil
}

func (s *PrizeService) RestoreStock(ctx context.Context, prizeID string, count int32) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	prize, err := s.prizeRepo.GetPrize(ctx, prizeID)
	if err != nil {
		return fmt.Errorf("prize not found: %w", err)
	}

	newStock := prize.Stock + count
	return s.prizeRepo.UpdateStock(ctx, prizeID, newStock)
}
