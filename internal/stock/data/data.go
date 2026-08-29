package data

import (
	"context"

	"github.com/marketing-platform/internal/conf"
	"github.com/marketing-platform/internal/stock/biz"
	"github.com/marketing-platform/internal/stock/data/ent"
	"github.com/marketing-platform/internal/stock/data/ent/stockitem"

	_ "github.com/go-sql-driver/mysql"
	"github.com/google/wire"
)

var ProviderSet = wire.NewSet(NewData, NewStockRepo)

type Data struct {
	db *ent.Client
}

func NewData(c *conf.Data) (*Data, func(), error) {
	dc := c.GetDatabase()
	db, err := ent.Open(dc.GetDriver(), dc.GetSource())
	if err != nil {
		return nil, nil, err
	}
	if dc.GetDebug() {
		db = db.Debug()
	}
	if dc.GetAutoMigrate() {
		if err := db.Schema.Create(context.Background()); err != nil {
			db.Close()
			return nil, nil, err
		}
	}
	return &Data{db: db}, func() { db.Close() }, nil
}

func (d *Data) HealthCheck(ctx context.Context) map[string]bool {
	healthy := false
	if d.db != nil {
		// 真正 ping 一下数据库，而不是只判断指针非空
		if _, err := d.db.StockItem.Query().Limit(1).All(ctx); err == nil {
			healthy = true
		}
	}
	return map[string]bool{"mysql": healthy}
}

type stockRepo struct{ data *Data }

func NewStockRepo(data *Data) biz.StockRepo { return &stockRepo{data: data} }

func (r *stockRepo) GetStock(ctx context.Context, stockKey string) (*biz.StockItem, error) {
	po, err := r.data.db.StockItem.Query().
		Where(stockitem.StockKeyEQ(stockKey)).Only(ctx)
	if err != nil {
		return nil, err
	}
	return &biz.StockItem{
		StockKey:  po.StockKey,
		StockName: po.StockName,
		StockType: po.StockType,
		Stock:     po.Stock,
		Total:     po.Total,
	}, nil
}

func (r *stockRepo) UpdateStock(ctx context.Context, stockKey string, stock int32) error {
	_, err := r.data.db.StockItem.Update().
		Where(stockitem.StockKeyEQ(stockKey)).
		SetStock(stock).
		Save(ctx)
	return err
}

func (r *stockRepo) DeductStockAtomic(ctx context.Context, stockKey string, count int32) (bool, error) {
	n, err := r.data.db.StockItem.Update().
		Where(stockitem.StockKeyEQ(stockKey), stockitem.StockGTE(count)).
		AddStock(-count).
		Save(ctx)
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

func (r *stockRepo) RestoreStockAtomic(ctx context.Context, stockKey string, count int32) (bool, error) {
	po, err := r.data.db.StockItem.Query().
		Where(stockitem.StockKeyEQ(stockKey)).Only(ctx)
	if err != nil {
		return false, err
	}
	newStock := po.Stock + count
	if newStock > po.Total {
		newStock = po.Total
	}
	n, err := r.data.db.StockItem.Update().
		Where(stockitem.StockKeyEQ(stockKey)).
		SetStock(newStock).
		Save(ctx)
	if err != nil {
		return false, err
	}
	return n > 0, nil
}
