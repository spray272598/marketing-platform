package server

import (
	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/log"
)

type GroupBuyServer struct {
	*kratos.App
}

func NewGroupBuyServer(
	logger log.Logger,
	grpcSrv *GRPCServer,
) *GroupBuyServer {
	app := kratos.New(
		kratos.Name("groupbuy-market"),
		kratos.Logger(logger),
		kratos.Server(
			grpcSrv.Server,
		),
	)
	return &GroupBuyServer{App: app}
}

func (s *GroupBuyServer) Run() error {
	return s.App.Run()
}
