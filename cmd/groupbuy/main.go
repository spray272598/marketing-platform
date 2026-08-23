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

	"github.com/marketing-platform/internal/groupbuy/biz"
	"github.com/marketing-platform/internal/groupbuy/data"
	"github.com/marketing-platform/internal/groupbuy/server"
	"github.com/marketing-platform/internal/groupbuy/service"
	"github.com/marketing-platform/pkg/config"
	"github.com/marketing-platform/pkg/log"
	"github.com/marketing-platform/pkg/middleware"
	"github.com/marketing-platform/pkg/observability"
	"github.com/marketing-platform/pkg/stockclient"
)

func main() {
	cfg, err := config.LoadConfig("groupbuy")
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	logger := log.NewLogger(cfg.Log.Level, cfg.Log.Format)
	slog.SetDefault(logger.Slog())

	var metrics *observability.Metrics
	var traceCollector *observability.TraceCollector
	if cfg.Observability.Metrics.Enabled {
		metrics = observability.NewMetrics("groupbuy")
		traceCollector = observability.NewTraceCollector("groupbuy", logger.Slog(), metrics)
	}

	mysqlDSN := cfg.GetMySQLDSN("root:root@tcp(127.0.0.1:3306)/marketing_groupbuy?charset=utf8mb4&parseTime=True&loc=Local")
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
	ch, err := conn.Channel()
	if err != nil {
		logger.Error("failed to open channel", log.Fields{"error": err})
		os.Exit(1)
	}
	defer ch.Close()

	dataLayer := data.NewData(db, rdb, conn, ch, nil)
	activityRepo := data.NewActivityRepo(dataLayer)
	orderRepo := data.NewOrderRepo(dataLayer)
	teamRepo := data.NewTeamRepo(dataLayer)
	redisRepo := data.NewRedisRepo(rdb)
	mqRepo := data.NewMQRepo(conn, ch)
	notifyTaskRepo := data.NewNotifyTaskRepo(dataLayer)
	stockClient := stockclient.NewClient(stockURL)

	notifySvc := biz.NewNotifyService(notifyTaskRepo, mqRepo)
	lockSvc := biz.NewLockService(activityRepo, orderRepo, teamRepo, redisRepo)
	trialSvc := biz.NewTrialService(activityRepo)
	settlementSvc := biz.NewSettlementService(teamRepo, orderRepo, notifySvc, stockClient)
	refundSvc := biz.NewRefundService(orderRepo, notifySvc, stockClient)

	svc := service.NewGroupBuyService(trialSvc, lockSvc, settlementSvc, refundSvc)

	chain := middleware.NewMiddlewareChain(logger.Slog(), metrics)
	if traceCollector != nil {
		chain.SetTraceCollector(traceCollector)
	}
	chain.DefaultChain()
	groupbuyServer := server.NewGroupBuyServer(svc, metrics, chain)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	consumer := biz.NewNotifyConsumer(notifyTaskRepo, mqRepo)
	consumer.SetLogger(logger.Slog())
	go consumer.Start(ctx)

	go func() {
		if err := groupbuyServer.Run(); err != nil {
			logger.Error("server error", log.Fields{"error": err})
			cancel()
		}
	}()

	logger.Info("GroupBuy market started", log.Fields{"addr": cfg.GetServerAddr()})

	select {
	case sig := <-sigCh:
		logger.Info("Received signal, shutting down", log.Fields{"signal": sig.String()})
	case <-ctx.Done():
		logger.Info("Context cancelled, shutting down")
	}

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := groupbuyServer.Stop(shutdownCtx); err != nil {
		logger.Error("Server shutdown error", log.Fields{"error": err})
	}

	logger.Info("GroupBuy market stopped")
}
