package main

import (
	"flag"
	"log/slog"
	"os"

	"github.com/marketing-platform/internal/conf"

	"github.com/go-kratos/kratos/v3"
	"github.com/go-kratos/kratos/v3/config"
	"github.com/go-kratos/kratos/v3/config/env"
	"github.com/go-kratos/kratos/v3/config/file"
	kratoslog "github.com/go-kratos/kratos/v3/log"
	"github.com/go-kratos/kratos/v3/transport/http"
)

var (
	flagconf string
	id, _    = os.Hostname()
)

func init() {
	flag.StringVar(&flagconf, "conf", "../../configs/gateway", "config path")
}

func newApp(logger *slog.Logger, hs *http.Server) *kratos.App {
	return kratos.New(
		kratos.ID(id), kratos.Name("gateway"),
		kratos.Logger(logger), kratos.Server(hs),
	)
}

func main() {
	flag.Parse()
	logger := kratoslog.NewLogger(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{AddSource: true, Level: slog.LevelInfo}))
	kratoslog.SetDefault(logger)

	c := config.New(config.WithSource(file.NewSource(flagconf), env.NewSource("GATEWAY")))
	defer c.Close()
	if err := c.Load(); err != nil {
		panic(err)
	}
	var bc conf.Bootstrap
	if err := c.Scan(&bc); err != nil {
		panic(err)
	}
	app, cleanup, err := wireApp(bc.Server, bc.Data, logger)
	if err != nil {
		panic(err)
	}
	defer cleanup()
	if err := app.Run(); err != nil {
		panic(err)
	}
}
