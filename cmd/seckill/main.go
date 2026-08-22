package main

import (
	"context"
	"fmt"
	"os"

	"github.com/go-kratos/kratos/v2/log"
	_ "go.uber.org/automaxprocs"

	"github.com/marketing-platform/internal/seckill/biz"
	"github.com/marketing-platform/internal/seckill/data"
	"github.com/marketing-platform/internal/seckill/server"
	"github.com/marketing-platform/internal/seckill/service"
)

func main() {
	logger := log.With(log.NewStdLogger(os.Stdout),
		"ts", log.DefaultTimestamp,
		"caller", log.DefaultCaller,
		"service.id", "seckill-market",
		"service.version", "1.0.0",
	)

	// TODO: Initialize database, redis, mq connections
	// TODO: Initialize repos and services
	// TODO: Use Wire for dependency injection

	// Placeholder initialization
	_ = biz.TradeService{}
	_ = data.Data{}

	_ = logger

	// Create service
	tradeSvc := biz.NewTradeService(nil, nil, nil)
	svc := service.NewSeckillService(tradeSvc)

	// Create servers
	grpcSrv := server.NewGRPCServer(svc)
	httpSrv := server.NewHTTPServer(svc)

	// Create app
	app := server.NewSeckillServer(logger, grpcSrv, httpSrv)

	fmt.Println("Seckill market starting...")
	if err := app.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to run server: %v\n", err)
		os.Exit(1)
	}

	_ = context.Background()
}
