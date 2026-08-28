//go:build wireinject
// +build wireinject

package main

import (
	"log/slog"

	"github.com/marketing-platform/internal/conf"
	"github.com/marketing-platform/internal/seckill/biz"
	"github.com/marketing-platform/internal/seckill/data"
	"github.com/marketing-platform/internal/seckill/server"
	"github.com/marketing-platform/internal/seckill/service"

	"github.com/go-kratos/kratos/v3"
	"github.com/google/wire"
)

func wireApp(*conf.Server, *conf.Data, *slog.Logger) (*kratos.App, func(), error) {
	panic(wire.Build(server.ProviderSet, data.ProviderSet, biz.ProviderSet, service.ProviderSet, newApp))
}
