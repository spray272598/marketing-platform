//go:build wireinject
// +build wireinject

package main

import (
	"github.com/google/wire"
	"github.com/marketing-platform/internal/lottery/biz"
	"github.com/marketing-platform/internal/lottery/data"
	"github.com/marketing-platform/internal/lottery/server"
	"github.com/marketing-platform/internal/lottery/service"
)

func InitializeLotteryServer() (*server.LotteryServer, func(), error) {
	wire.Build(
		data.NewData,
		data.NewActivityRepo,
		data.NewStrategyRepo,
		data.NewOrderRepo,
		biz.NewRaffleService,
		service.NewLotteryService,
		server.NewLotteryServer,
	)
	return nil, nil, nil
}
