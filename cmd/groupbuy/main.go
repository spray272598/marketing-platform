package main

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
	_ "go.uber.org/automaxprocs"
	_ "github.com/go-sql-driver/mysql"

	"github.com/marketing-platform/internal/groupbuy/biz"
	"github.com/marketing-platform/internal/groupbuy/data"
	"github.com/marketing-platform/internal/groupbuy/server"
	"github.com/marketing-platform/internal/groupbuy/service"
)

func main() {
	mysqlDSN := getEnv("MYSQL_DSN", "root:root@tcp(127.0.0.1:3306)/marketing_groupbuy?charset=utf8mb4&parseTime=True&loc=Local")
	redisAddr := getEnv("REDIS_ADDR", "127.0.0.1:6379")

	db, err := sql.Open("mysql", mysqlDSN)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to open db: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	rdb := redis.NewClient(&redis.Options{
		Addr:         redisAddr,
		ReadTimeout:  2 * time.Second,
		WriteTimeout: 2 * time.Second,
	})
	defer rdb.Close()

	dataLayer := data.NewData(db, rdb, nil, nil, nil)

	activityRepo := data.NewActivityRepo(dataLayer)
	orderRepo := data.NewOrderRepo(dataLayer)
	teamRepo := data.NewTeamRepo(dataLayer)
	redisRepo := data.NewRedisRepo(rdb)

	trialSvc := biz.NewTrialService(activityRepo)
	lockSvc := biz.NewLockService(activityRepo, orderRepo, teamRepo, redisRepo)
	settlementSvc := biz.NewSettlementService(teamRepo, orderRepo, nil)
	refundSvc := biz.NewRefundService(orderRepo, nil)

	svc := service.NewGroupBuyService(trialSvc, lockSvc, settlementSvc, refundSvc)

	app := server.NewGroupBuyServer(svc)

	fmt.Println("GroupBuy market starting...")
	if err := app.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to run server: %v\n", err)
		os.Exit(1)
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
