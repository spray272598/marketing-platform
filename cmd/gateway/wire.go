//go:build wireinject
// +build wireinject

package main

import (
	"log/slog"

	"github.com/marketing-platform/internal/conf"
	"github.com/marketing-platform/internal/gateway"
	"github.com/marketing-platform/internal/gateway/server"

	"github.com/go-kratos/kratos/v3"
	"github.com/google/wire"
)

func wireApp(*conf.Server, *conf.Data, *slog.Logger) (*kratos.App, func(), error) {
	panic(wire.Build(server.ProviderSet, gateway.ProviderSet, newApp))
}
