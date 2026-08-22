package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
	_ "go.uber.org/automaxprocs"
	_ "github.com/go-sql-driver/mysql"

	"github.com/marketing-platform/internal/seckill/biz"
	"github.com/marketing-platform/internal/seckill/data"
	"github.com/marketing-platform/internal/seckill/server"
	"github.com/marketing-platform/internal/seckill/service"
)

func main() {
	// Init MySQL
	db, err := sql.Open("mysql", "root:root@tcp(127.0.0.1:3306)/marketing_seckill?charset=utf8mb4&parseTime=True&loc=Local")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to open db: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	// Init Redis
	rdb := redis.NewClient(&redis.Options{
		Addr:         "127.0.0.1:6379",
		ReadTimeout:  2 * time.Second,
		WriteTimeout: 2 * time.Second,
	})
	defer rdb.Close()

	dataLayer := data.NewData(db, rdb, nil, nil, nil)

	activityRepo := data.NewActivityRepo(dataLayer)
	orderRepo := data.NewOrderRepo(dataLayer)
	redisRepo := data.NewRedisRepo(rdb)

	tradeSvc := biz.NewTradeService(orderRepo, redisRepo, nil)
	_ = activityRepo

	svc := service.NewSeckillService(tradeSvc)

	app := server.NewSeckillServer(svc)

	fmt.Println("Seckill market starting...")
	if err := app.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to run server: %v\n", err)
		os.Exit(1)
	}

	_ = context.Background()
}
