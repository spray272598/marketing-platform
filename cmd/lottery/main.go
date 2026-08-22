package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"

	_ "go.uber.org/automaxprocs"
	_ "github.com/go-sql-driver/mysql"

	"github.com/marketing-platform/internal/lottery/biz"
	"github.com/marketing-platform/internal/lottery/data"
	"github.com/marketing-platform/internal/lottery/server"
	"github.com/marketing-platform/internal/lottery/service"
)

func main() {
	// Init MySQL
	db, err := sql.Open("mysql", "root:root@tcp(127.0.0.1:3306)/marketing_lottery?charset=utf8mb4&parseTime=True&loc=Local")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to open db: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	dataLayer := data.NewData(db, nil)

	activityRepo := data.NewActivityRepo(dataLayer)
	strategyRepo := data.NewStrategyRepo(dataLayer)
	orderRepo := data.NewOrderRepo(dataLayer)

	raffleSvc := biz.NewRaffleService(activityRepo, strategyRepo, orderRepo)
	svc := service.NewLotteryService(raffleSvc)

	app := server.NewLotteryServer(svc)

	fmt.Println("Lottery market starting...")
	if err := app.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to run server: %v\n", err)
		os.Exit(1)
	}

	_ = context.Background()
}
