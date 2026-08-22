package server

import (
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/marketing-platform/internal/lottery/service"
	lotteryv1 "github.com/marketing-platform/api/lottery/v1"
)

type GRPCServer struct {
	*grpc.Server
}

func NewGRPCServer(lotterySvc *service.LotteryService) *GRPCServer {
	srv := grpc.NewServer(
		grpc.Address(":18083"),
		grpc.Middleware(recovery.Recovery()),
	)
	lotteryv1.RegisterLotteryServiceServer(srv.Server, lotterySvc)
	return &GRPCServer{Server: srv}
}
