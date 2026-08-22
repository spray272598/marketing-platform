package main

import (
	"context"
	"database/sql"
	"os"

	_ "github.com/go-sql-driver/mysql"

	"github.com/marketing-platform/internal/stock/biz"
	"github.com/marketing-platform/internal/stock/data"
	stockserver "github.com/marketing-platform/internal/stock/server"
	"github.com/marketing-platform/internal/stock/service"
	"github.com/marketing-platform/pkg/common"
	"github.com/marketing-platform/pkg/config"
	"github.com/marketing-platform/pkg/log"
)

func main() {
	cfg := config.LoadOrDefault("")
	logger := log.NewLogger(cfg.Log.Level, cfg.Log.Format)

	mysqlDSN := cfg.GetMySQLDSN("root:root@tcp(127.0.0.1:3306)/marketing_stock?charset=utf8mb4&parseTime=True&loc=Local")

	db, err := sql.Open("mysql", mysqlDSN)
	if err != nil {
		logger.Error("failed to open db", log.Fields{"error": err})
		os.Exit(1)
	}
	defer db.Close()

	dataLayer := data.NewData(db)
	stockRepo := data.NewStockRepo(dataLayer)
	stockSvc := biz.NewStockService(stockRepo)
	svc := service.NewStockService(stockSvc)

	stockServer := stockserver.NewServer(svc)

	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	go common.WaitForShutdown(cancel)

	logger.Info("Stock service starting...", log.Fields{"addr": ":18094"})
	if err := stockServer.Run(); err != nil {
		logger.Error("failed to run server", log.Fields{"error": err})
		os.Exit(1)
	}
}
