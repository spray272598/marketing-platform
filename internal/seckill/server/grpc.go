package server

import (
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/marketing-platform/internal/seckill/service"
	seckillv1 "github.com/marketing-platform/api/seckill/v1"
)

type GRPCServer struct {
	*grpc.Server
}

func NewGRPCServer(seckillSvc *service.SeckillService) *GRPCServer {
	srv := grpc.NewServer(
		grpc.Address(":18081"),
		grpc.Middleware(
			recovery.Recovery(),
		),
	)
	seckillv1.RegisterSeckillServiceServer(srv.Server, seckillSvc)
	return &GRPCServer{Server: srv}
}
