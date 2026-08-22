package main

import (
	"context"
	"database/sql"
	"os"

	_ "github.com/go-sql-driver/mysql"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/redis/go-redis/v9"

	"github.com/marketing-platform/internal/groupbuy/biz"
	"github.com/marketing-platform/internal/groupbuy/data"
	"github.com/marketing-platform/internal/groupbuy/server"
	"github.com/marketing-platform/internal/groupbuy/service"
	"github.com/marketing-platform/pkg/common"
	"github.com/marketing-platform/pkg/config"
	"github.com/marketing-platform/pkg/log"
	"github.com/marketing-platform/pkg/stockclient"
)

func main() {
	cfg := config.LoadOrDefault("")
	logger := log.NewLogger(cfg.Log.Level, cfg.Log.Format)

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
	groupbuyServer := server.NewGroupBuyServer(svc)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go common.WaitForShutdown(cancel)

	consumer := biz.NewNotifyConsumer(notifyTaskRepo)
	go consumer.Start(ctx)

	logger.Info("GroupBuy market starting...", log.Fields{"addr": cfg.GetServerAddr()})
	if err := groupbuyServer.Run(); err != nil {
		logger.Error("failed to run server", log.Fields{"error": err})
		os.Exit(1)
	}
}
