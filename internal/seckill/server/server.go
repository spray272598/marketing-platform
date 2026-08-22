package server

import (
	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/log"
)

type SeckillServer struct {
	*kratos.App
}

func NewSeckillServer(
	logger log.Logger,
	grpcSrv *GRPCServer,
	httpSrv *HTTPServer,
) *SeckillServer {
	app := kratos.New(
		kratos.Name("seckill-market"),
		kratos.Logger(logger),
		kratos.Server(
			grpcSrv.Server,
			httpSrv.Server,
		),
	)
	return &SeckillServer{App: app}
}

func (s *SeckillServer) Run() error {
	return s.App.Run()
}
