package biz

import (
	"context"
	"testing"
)

type mockStockRepo struct {
	stocks map[string]*StockItem
}

func newMockStockRepo() *mockStockRepo {
	return &mockStockRepo{
		stocks: map[string]*StockItem{
			"product:sku_001": {StockKey: "product:sku_001", StockName: "iPhone", StockType: "product", Stock: 100, Total: 100},
			"team:team_001":   {StockKey: "team:team_001", StockName: "团位", StockType: "team", Stock: 50, Total: 50},
			"prize:award_001": {StockKey: "prize:award_001", StockName: "奖品", StockType: "prize", Stock: 10, Total: 10},
		},
	}
}

func (m *mockStockRepo) GetStock(ctx context.Context, stockKey string) (*StockItem, error) {
	if item, ok := m.stocks[stockKey]; ok {
		return item, nil
	}
	return nil, nil
}

func (m *mockStockRepo) UpdateStock(ctx context.Context, stockKey string, stock int32) error {
	if item, ok := m.stocks[stockKey]; ok {
		item.Stock = stock
	}
	return nil
}

func TestDeductStock_Success(t *testing.T) {
	repo := newMockStockRepo()
	svc := NewStockService(repo)

	ok, err := svc.DeductStock(context.Background(), "product:sku_001", 5)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !ok {
		t.Error("expected deduct success")
	}

	stock, _ := svc.GetStock(context.Background(), "product:sku_001")
	if stock != 95 {
		t.Errorf("expected stock 95, got %d", stock)
	}
}

func TestDeductStock_NotEnough(t *testing.T) {
	repo := newMockStockRepo()
	svc := NewStockService(repo)

	ok, err := svc.DeductStock(context.Background(), "prize:award_001", 20)
	if err == nil {
		t.Error("expected error, got nil")
	}
	if ok {
		t.Error("expected deduct failed")
	}
}

func TestDeductStock_NotFound(t *testing.T) {
	repo := newMockStockRepo()
	svc := NewStockService(repo)

	ok, err := svc.DeductStock(context.Background(), "non_existent", 1)
	if err == nil {
		t.Error("expected error, got nil")
	}
	if ok {
		t.Error("expected deduct failed")
	}
}

func TestGetStock_Success(t *testing.T) {
	repo := newMockStockRepo()
	svc := NewStockService(repo)

	stock, err := svc.GetStock(context.Background(), "team:team_001")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if stock != 50 {
		t.Errorf("expected stock 50, got %d", stock)
	}
}

func TestGetStock_NotFound(t *testing.T) {
	repo := newMockStockRepo()
	svc := NewStockService(repo)

	_, err := svc.GetStock(context.Background(), "non_existent")
	if err == nil {
		t.Error("expected error, got nil")
	}
}

func TestRestoreStock_Success(t *testing.T) {
	repo := newMockStockRepo()
	svc := NewStockService(repo)

	err := svc.RestoreStock(context.Background(), "prize:award_001", 5)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	stock, _ := svc.GetStock(context.Background(), "prize:award_001")
	if stock != 15 {
		t.Errorf("expected stock 15, got %d", stock)
	}
}

func TestDeductStock_AllTypes(t *testing.T) {
	repo := newMockStockRepo()
	svc := NewStockService(repo)

	tests := []struct {
		name     string
		stockKey string
		count    int32
		wantErr  bool
	}{
		{"product stock", "product:sku_001", 1, false},
		{"team stock", "team:team_001", 1, false},
		{"prize stock", "prize:award_001", 1, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ok, err := svc.DeductStock(context.Background(), tt.stockKey, tt.count)
			if (err != nil) != tt.wantErr {
				t.Errorf("DeductStock() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !ok && !tt.wantErr {
				t.Error("expected deduct success")
			}
		})
	}
}
