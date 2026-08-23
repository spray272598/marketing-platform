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

	"github.com/marketing-platform/internal/lottery/biz"
	"github.com/marketing-platform/internal/lottery/data"
	"github.com/marketing-platform/internal/lottery/server"
	"github.com/marketing-platform/internal/lottery/service"
	"github.com/marketing-platform/pkg/config"
	"github.com/marketing-platform/pkg/log"
	"github.com/marketing-platform/pkg/middleware"
	"github.com/marketing-platform/pkg/observability"
)

func main() {
	cfg, err := config.LoadConfig("lottery")
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	logger := log.NewLogger(cfg.Log.Level, cfg.Log.Format)
	slog.SetDefault(logger.Slog())

	var metrics *observability.Metrics
	if cfg.Observability.Metrics.Enabled {
		metrics = observability.NewMetrics("lottery")
	}

	mysqlDSN := cfg.GetMySQLDSN("root:root@tcp(127.0.0.1:3306)/marketing_lottery?charset=utf8mb4&parseTime=True&loc=Local")

	db, err := sql.Open("mysql", mysqlDSN)
	if err != nil {
		logger.Error("failed to open db", log.Fields{"error": err})
		os.Exit(1)
	}
	defer db.Close()

	dataLayer := data.NewData(db, nil)

	activityRepo := data.NewActivityRepo(dataLayer)
	strategyRepo := data.NewStrategyRepo(dataLayer)
	orderRepo := data.NewOrderRepo(dataLayer)

	raffleSvc := biz.NewRaffleService(activityRepo, strategyRepo, orderRepo)
	svc := service.NewLotteryService(raffleSvc)

	chain := middleware.NewMiddlewareChain(logger.Slog(), metrics)
	lotteryServer := server.NewLotteryServer(svc, metrics, chain)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := lotteryServer.Run(); err != nil {
			logger.Error("server error", log.Fields{"error": err})
			cancel()
		}
	}()

	logger.Info("Lottery market started", log.Fields{"addr": ":18093"})

	select {
	case sig := <-sigCh:
		logger.Info("Received signal, shutting down", log.Fields{"signal": sig.String()})
	case <-ctx.Done():
		logger.Info("Context cancelled, shutting down")
	}

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := lotteryServer.Stop(shutdownCtx); err != nil {
		logger.Error("Server shutdown error", log.Fields{"error": err})
	}

	logger.Info("Lottery market stopped")
}
