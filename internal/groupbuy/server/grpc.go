package server

import (
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/marketing-platform/internal/groupbuy/service"
	groupbuyv1 "github.com/marketing-platform/api/groupbuy/v1"
)

type GRPCServer struct {
	*grpc.Server
}

func NewGRPCServer(groupBuySvc *service.GroupBuyService) *GRPCServer {
	srv := grpc.NewServer(
		grpc.Address(":18082"),
		grpc.Middleware(recovery.Recovery()),
	)
	groupbuyv1.RegisterGroupBuyServiceServer(srv.Server, groupBuySvc)
	return &GRPCServer{Server: srv}
}
