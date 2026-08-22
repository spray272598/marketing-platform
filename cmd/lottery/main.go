package main

import (
	"context"
	"database/sql"
	"os"

	_ "github.com/go-sql-driver/mysql"

	"github.com/marketing-platform/internal/lottery/biz"
	"github.com/marketing-platform/internal/lottery/data"
	"github.com/marketing-platform/internal/lottery/server"
	"github.com/marketing-platform/internal/lottery/service"
	"github.com/marketing-platform/pkg/common"
	"github.com/marketing-platform/pkg/config"
	"github.com/marketing-platform/pkg/log"
)

func main() {
	cfg := config.LoadOrDefault("")
	logger := log.NewLogger(cfg.Log.Level, cfg.Log.Format)

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

	lotteryServer := server.NewLotteryServer(svc)

	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	go common.WaitForShutdown(cancel)

	logger.Info("Lottery market starting...", log.Fields{"addr": ":18093"})
	if err := lotteryServer.Run(); err != nil {
		logger.Error("failed to run server", log.Fields{"error": err})
		os.Exit(1)
	}
}
