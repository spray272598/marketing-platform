package data

import (
	"context"

	"github.com/marketing-platform/internal/conf"
	"github.com/marketing-platform/internal/lottery/data/ent"

	_ "github.com/go-sql-driver/mysql"
	"github.com/google/wire"
)

var ProviderSet = wire.NewSet(NewData, NewActivityRepo, NewStrategyRepo, NewOrderRepo)

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
	cleanup := func() { db.Close() }
	return &Data{db: db}, cleanup, nil
}

func (d *Data) HealthCheck(ctx context.Context) map[string]bool {
	healthy := false
	if d.db != nil {
		// 真正 ping 一下数据库，而不是只判断指针非空
		if _, err := d.db.LotteryActivity.Query().Limit(1).All(ctx); err == nil {
			healthy = true
		}
	}
	return map[string]bool{"mysql": healthy}
}
