package main

import (
	"fmt"
	"os"

	"github.com/go-kratos/kratos/v2/log"
	_ "go.uber.org/automaxprocs"

	"github.com/marketing-platform/internal/groupbuy/biz"
	"github.com/marketing-platform/internal/groupbuy/server"
	"github.com/marketing-platform/internal/groupbuy/service"
)

func main() {
	logger := log.With(log.NewStdLogger(os.Stdout),
		"ts", log.DefaultTimestamp,
		"caller", log.DefaultCaller,
		"service.id", "groupbuy-market",
		"service.version", "1.0.0",
	)

	_ = logger

	// Create services
	trialSvc := biz.NewTrialService(nil)
	lockSvc := biz.NewLockService(nil, nil, nil, nil)
	settlementSvc := biz.NewSettlementService(nil, nil, nil)
	refundSvc := biz.NewRefundService(nil, nil)

	svc := service.NewGroupBuyService(trialSvc, lockSvc, settlementSvc, refundSvc)

	// Create gRPC server
	grpcSrv := server.NewGRPCServer(svc)

	fmt.Println("GroupBuy market starting...")

	// TODO: Create and run Kratos app
	_ = grpcSrv
}
