//go:build wireinject
// +build wireinject

package main

import (
	"github.com/google/wire"
	"github.com/marketing-platform/internal/groupbuy/biz"
	"github.com/marketing-platform/internal/groupbuy/data"
	"github.com/marketing-platform/internal/groupbuy/server"
	"github.com/marketing-platform/internal/groupbuy/service"
)

func InitializeGroupBuyServer() (*server.GroupBuyServer, func(), error) {
	wire.Build(
		data.NewData,
		data.NewActivityRepo,
		data.NewOrderRepo,
		data.NewTeamRepo,
		data.NewRedisRepo,
		data.NewMQRepo,
		biz.NewTrialService,
		biz.NewLockService,
		biz.NewSettlementService,
		biz.NewRefundService,
		service.NewGroupBuyService,
		server.NewGroupBuyServer,
	)
	return nil, nil, nil
}
