package main

import (
	"fmt"
	"os"

	"github.com/go-kratos/kratos/v2/log"
	_ "go.uber.org/automaxprocs"

	"github.com/marketing-platform/internal/lottery/biz"
	"github.com/marketing-platform/internal/lottery/server"
	"github.com/marketing-platform/internal/lottery/service"
)

func main() {
	logger := log.With(log.NewStdLogger(os.Stdout),
		"ts", log.DefaultTimestamp,
		"caller", log.DefaultCaller,
		"service.id", "lottery-market",
		"service.version", "1.0.0",
	)

	_ = logger

	// Create services
	raffleSvc := biz.NewRaffleService(nil, nil, nil)
	svc := service.NewLotteryService(raffleSvc)

	// Create gRPC server
	grpcSrv := server.NewGRPCServer(svc)

	fmt.Println("Lottery market starting...")

	// TODO: Create and run Kratos app
	_ = grpcSrv
}
