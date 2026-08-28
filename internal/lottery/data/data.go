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
	return map[string]bool{"mysql": d.db != nil}
}
