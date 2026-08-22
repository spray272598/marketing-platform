package data

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/marketing-platform/internal/stock/biz"
)

type Data struct {
	db *sql.DB
}

func NewData(db *sql.DB) *Data {
	return &Data{db: db}
}

type stockRepo struct {
	data *Data
}

func NewStockRepo(data *Data) biz.StockRepo {
	return &stockRepo{data: data}
}

func (r *stockRepo) GetStock(ctx context.Context, stockKey string) (*biz.StockItem, error) {
	query := `SELECT id, stock_key, stock_name, stock_type, stock, total FROM stock_item WHERE stock_key = ?`
	item := &biz.StockItem{}
	err := r.data.db.QueryRowContext(ctx, query, stockKey).Scan(
		&item.ID, &item.StockKey, &item.StockName, &item.StockType, &item.Stock, &item.Total,
	)
	if err != nil {
		return nil, fmt.Errorf("stock not found: %w", err)
	}
	return item, nil
}

func (r *stockRepo) UpdateStock(ctx context.Context, stockKey string, stock int32) error {
	query := `UPDATE stock_item SET stock = ? WHERE stock_key = ?`
	_, err := r.data.db.ExecContext(ctx, query, stock, stockKey)
	return err
}
