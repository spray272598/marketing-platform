package data

import (
	"context"
	"database/sql"

	"github.com/marketing-platform/internal/conf"
	"github.com/marketing-platform/internal/seckill/biz"
	"github.com/marketing-platform/internal/seckill/data/ent"
	"github.com/marketing-platform/pkg/stockclient"

	_ "github.com/go-sql-driver/mysql"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/google/wire"
	"github.com/redis/go-redis/v9"
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(
	NewData,
	NewActivityRepo,
	NewOrderRepo,
	NewRedisRepo,
	NewMQRepo,
	NewStockClient,
)

// Data holds the long-lived storage clients shared by repos.
type Data struct {
	db      *ent.Client
	sqldb   *sql.DB
	rdb     *redis.Client
	conn    *amqp.Connection
	channel *amqp.Channel
}

// NewData opens all storage clients and returns them with a cleanup function.
func NewData(c *conf.Data) (*Data, func(), error) {
	dc := c.GetDatabase()
	db, err := ent.Open(dc.GetDriver(), dc.GetSource())
	if err != nil {
		return nil, nil, err
	}
	// 原生 *sql.DB：号段模式(ID 分配)等需要跑原生 SQL 事务的场景使用，
	// 与 Ent 共用同一 DSN，独立连接池。
	sqldb, err := sql.Open("mysql", dc.GetSource())
	if err != nil {
		db.Close()
		return nil, nil, err
	}
	if dc.GetDebug() {
		db = db.Debug()
	}
	if dc.GetAutoMigrate() {
		if err := db.Schema.Create(context.Background()); err != nil {
			sqldb.Close()
			db.Close()
			return nil, nil, err
		}
	}

	rc := c.GetRedis()
	rdb := redis.NewClient(&redis.Options{
		Addr:         rc.GetAddr(),
		Password:     rc.GetPassword(),
		DB:           int(rc.GetDb()),
		ReadTimeout:  rc.GetReadTimeout().AsDuration(),
		WriteTimeout: rc.GetWriteTimeout().AsDuration(),
	})

	var conn *amqp.Connection
	var ch *amqp.Channel
	rmq := c.GetRabbitmq()
	if rmq != nil && rmq.GetUrl() != "" {
		conn, err = amqp.Dial(rmq.GetUrl())
		if err != nil {
			db.Close()
			rdb.Close()
			return nil, nil, err
		}
		ch, err = conn.Channel()
		if err != nil {
			db.Close()
			rdb.Close()
			conn.Close()
			return nil, nil, err
		}
	}

	cleanup := func() {
		if ch != nil {
			ch.Close()
		}
		if conn != nil {
			conn.Close()
		}
		rdb.Close()
		db.Close()
		sqldb.Close()
	}

	return &Data{
		db:      db,
		sqldb:   sqldb,
		rdb:     rdb,
		conn:    conn,
		channel: ch,
	}, cleanup, nil
}

func (d *Data) HealthCheck(ctx context.Context) map[string]bool {
	health := make(map[string]bool)
	if d.db != nil {
		health["mysql"] = true
	}
	if d.rdb != nil {
		health["redis"] = d.rdb.Ping(ctx).Err() == nil
	}
	return health
}

// NewStockClient creates a stock HTTP client from config.
func NewStockClient(c *conf.Data) biz.StockClient {
	stockURL := "http://127.0.0.1:18094"
	if c.GetStock() != nil && c.GetStock().GetUrl() != "" {
		stockURL = c.GetStock().GetUrl()
	}
	return stockclient.NewClient(stockURL)
}
