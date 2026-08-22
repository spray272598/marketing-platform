//go:build wireinject
// +build wireinject

package main

import (
	"github.com/google/wire"
	"github.com/marketing-platform/internal/seckill/biz"
	"github.com/marketing-platform/internal/seckill/data"
	"github.com/marketing-platform/internal/seckill/server"
	"github.com/marketing-platform/internal/seckill/service"
)

func InitializeSeckillServer() (*server.SeckillServer, func(), error) {
	wire.Build(
		data.NewData,
		data.NewActivityRepo,
		data.NewOrderRepo,
		data.NewRedisRepo,
		data.NewMQRepo,
		biz.NewTradeService,
		service.NewSeckillService,
		server.NewSeckillServer,
	)
	return nil, nil, nil
}
