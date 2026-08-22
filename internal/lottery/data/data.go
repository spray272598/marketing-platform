package data

import (
	"database/sql"

	"github.com/go-kratos/kratos/v2/log"
)

type Data struct {
	db     *sql.DB
	logger *log.Helper
}

func NewData(db *sql.DB, logger log.Logger) *Data {
	return &Data{
		db:     db,
		logger: log.NewHelper(logger),
	}
}

func (d *Data) Close() error {
	if d.db != nil {
		if err := d.db.Close(); err != nil {
			d.logger.Errorf("failed to close db: %v", err)
		}
	}
	return nil
}
