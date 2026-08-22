package data

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/marketing-platform/internal/prize/biz"
)

type Data struct {
	db *sql.DB
}

func NewData(db *sql.DB) *Data {
	return &Data{db: db}
}

func (d *Data) Close() error {
	if d.db != nil {
		return d.db.Close()
	}
	return nil
}

type prizeRepo struct {
	data *Data
}

func NewPrizeRepo(data *Data) biz.PrizeRepo {
	return &prizeRepo{data: data}
}

func (r *prizeRepo) GetPrize(ctx context.Context, prizeID string) (*biz.PrizeItem, error) {
	query := `SELECT id, prize_id, prize_name, stock, total FROM prize_item WHERE prize_id = ?`
	item := &biz.PrizeItem{}
	err := r.data.db.QueryRowContext(ctx, query, prizeID).Scan(
		&item.ID, &item.PrizeID, &item.PrizeName, &item.Stock, &item.Total,
	)
	if err != nil {
		return nil, fmt.Errorf("prize not found: %w", err)
	}
	return item, nil
}

func (r *prizeRepo) UpdateStock(ctx context.Context, prizeID string, stock int32) error {
	query := `UPDATE prize_item SET stock = ? WHERE prize_id = ?`
	_, err := r.data.db.ExecContext(ctx, query, stock, prizeID)
	return err
}
