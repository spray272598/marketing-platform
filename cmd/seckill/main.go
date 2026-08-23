package main

import (
	"context"
	"database/sql"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/go-sql-driver/mysql"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/redis/go-redis/v9"

	"github.com/marketing-platform/internal/seckill/biz"
	"github.com/marketing-platform/internal/seckill/data"
	"github.com/marketing-platform/internal/seckill/server"
	"github.com/marketing-platform/internal/seckill/service"
	"github.com/marketing-platform/pkg/config"
	"github.com/marketing-platform/pkg/log"
	"github.com/marketing-platform/pkg/middleware"
	"github.com/marketing-platform/pkg/observability"
	"github.com/marketing-platform/pkg/stockclient"
)

func main() {
	cfg, err := config.LoadConfig("seckill")
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	logger := log.NewLogger(cfg.Log.Level, cfg.Log.Format)
	slog.SetDefault(logger.Slog())

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go config.WatchConfig(ctx, func(cfg *config.BootstrapConfig) {
		logger.Info("配置变更通知", slog.String("service", cfg.ServiceName))
	})

	var metrics *observability.Metrics
	if cfg.Observability.Metrics.Enabled {
		metrics = observability.NewMetrics("seckill")
	}

	mysqlDSN := cfg.GetMySQLDSN("root:root@tcp(127.0.0.1:3306)/marketing_seckill?charset=utf8mb4&parseTime=True&loc=Local")
	redisAddr := cfg.GetRedisAddr("127.0.0.1:6379")
	rabbitMQURL := cfg.GetRabbitMQURL("amqp://guest:guest@127.0.0.1:5672/")
	stockURL := cfg.GetStockURL("http://127.0.0.1:18094")

	db, err := sql.Open("mysql", mysqlDSN)
	if err != nil {
		logger.Error("failed to open db", log.Fields{"error": err})
		os.Exit(1)
	}
	defer db.Close()

	rdb := redis.NewClient(&redis.Options{Addr: redisAddr})
	conn, err := amqp.Dial(rabbitMQURL)
	if err != nil {
		logger.Error("failed to connect rabbitmq", log.Fields{"error": err})
		os.Exit(1)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		logger.Error("failed to create channel", log.Fields{"error": err})
		os.Exit(1)
	}
	defer ch.Close()

	dataLayer := data.NewData(db, rdb, conn, ch, nil)
	activityRepo := data.NewActivityRepo(dataLayer)
	orderRepo := data.NewOrderRepo(dataLayer)
	redisRepo := data.NewRedisRepo(rdb)
	mqRepo := data.NewMQRepo(conn, ch)
	stockClient := stockclient.NewClient(stockURL)

	tradeSvc := biz.NewTradeService(orderRepo, redisRepo, mqRepo, activityRepo, stockClient)
	svc := service.NewSeckillService(tradeSvc, activityRepo)

	chain := middleware.NewMiddlewareChain(logger.Slog(), metrics)
	seckillServer := server.NewSeckillServer(svc, metrics, chain)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := seckillServer.Run(); err != nil {
			logger.Error("server error", log.Fields{"error": err})
			cancel()
		}
	}()

	logger.Info("Seckill market started", log.Fields{"addr": cfg.GetServerAddr()})

	select {
	case sig := <-sigCh:
		logger.Info("Received signal, shutting down", log.Fields{"signal": sig.String()})
	case <-ctx.Done():
		logger.Info("Context cancelled, shutting down")
	}

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := seckillServer.Stop(shutdownCtx); err != nil {
		logger.Error("Server shutdown error", log.Fields{"error": err})
	}

	logger.Info("Seckill market stopped")
}
