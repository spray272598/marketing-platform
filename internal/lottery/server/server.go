package server

import (
	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/log"
)

type LotteryServer struct {
	*kratos.App
}

func NewLotteryServer(
	logger log.Logger,
	grpcSrv *GRPCServer,
) *LotteryServer {
	app := kratos.New(
		kratos.Name("lottery-market"),
		kratos.Logger(logger),
		kratos.Server(
			grpcSrv.Server,
		),
	)
	return &LotteryServer{App: app}
}

func (s *LotteryServer) Run() error {
	return s.App.Run()
}
