package data

import (
	"database/sql"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/redis/go-redis/v9"
	"github.com/go-kratos/kratos/v2/log"
)

type Data struct {
	db      *sql.DB
	rdb     *redis.Client
	conn    *amqp.Connection
	channel *amqp.Channel
	logger  *log.Helper
}

func NewData(
	db *sql.DB,
	rdb *redis.Client,
	conn *amqp.Connection,
	ch *amqp.Channel,
	logger log.Logger,
) *Data {
	return &Data{
		db:      db,
		rdb:     rdb,
		conn:    conn,
		channel: ch,
		logger:  log.NewHelper(logger),
	}
}

func (d *Data) Close() error {
	if d.db != nil {
		if err := d.db.Close(); err != nil {
			d.logger.Errorf("failed to close db: %v", err)
		}
	}
	if d.rdb != nil {
		if err := d.rdb.Close(); err != nil {
			d.logger.Errorf("failed to close redis: %v", err)
		}
	}
	if d.channel != nil {
		if err := d.channel.Close(); err != nil {
			d.logger.Errorf("failed to close mq channel: %v", err)
		}
	}
	if d.conn != nil {
		if err := d.conn.Close(); err != nil {
			d.logger.Errorf("failed to close mq connection: %v", err)
		}
	}
	return nil
}
