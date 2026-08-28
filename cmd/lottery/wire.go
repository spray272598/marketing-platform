//go:build wireinject
// +build wireinject

package main

import (
	"log/slog"

	"github.com/marketing-platform/internal/conf"
	"github.com/marketing-platform/internal/lottery/biz"
	"github.com/marketing-platform/internal/lottery/data"
	"github.com/marketing-platform/internal/lottery/server"
	"github.com/marketing-platform/internal/lottery/service"

	"github.com/go-kratos/kratos/v3"
	"github.com/google/wire"
)

func wireApp(*conf.Server, *conf.Data, *slog.Logger) (*kratos.App, func(), error) {
	panic(wire.Build(server.ProviderSet, data.ProviderSet, biz.ProviderSet, service.ProviderSet, newApp))
}
