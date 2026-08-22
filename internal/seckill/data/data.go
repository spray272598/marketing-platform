package data

import (
	"database/sql"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/redis/go-redis/v9"
)

type Data struct {
	db       *sql.DB
	rdb      *redis.Client
	conn     *amqp.Connection
	channel  *amqp.Channel
}

func NewData(
	db *sql.DB,
	rdb *redis.Client,
	conn *amqp.Connection,
	ch *amqp.Channel,
) *Data {
	return &Data{
		db:      db,
		rdb:     rdb,
		conn:    conn,
		channel: ch,
	}
}

func (d *Data) Close() error {
	var errs []error
	if d.db != nil {
		errs = append(errs, d.db.Close())
	}
	if d.rdb != nil {
		errs = append(errs, d.rdb.Close())
	}
	if d.channel != nil {
		errs = append(errs, d.channel.Close())
	}
	if d.conn != nil {
		errs = append(errs, d.conn.Close())
	}
	if len(errs) > 0 {
		return fmt.Errorf("close errors: %v", errs)
	}
	return nil
}
