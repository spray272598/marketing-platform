package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/marketing-platform/internal/gateway"
	gwserver "github.com/marketing-platform/internal/gateway/server"
	"github.com/marketing-platform/pkg/config"
	"github.com/marketing-platform/pkg/log"
	"github.com/marketing-platform/pkg/observability"
)

func main() {
	cfg, err := config.LoadConfig("gateway")
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	logger := log.NewLogger(cfg.Log.Level, cfg.Log.Format)
	slog.SetDefault(logger.Slog())

	var metrics *observability.Metrics
	if cfg.Observability.Metrics.Enabled {
		metrics = observability.NewMetrics("gateway")
	}

	gatewaySvc := gateway.NewService()
	gatewayServer := gwserver.NewServer(gatewaySvc, metrics)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := gatewayServer.Run(); err != nil {
			slog.Error("server error", "error", err)
			cancel()
		}
	}()

	logger.Info("Gateway started", log.Fields{"addr": ":8080"})

	select {
	case sig := <-sigCh:
		logger.Info("Received signal, shutting down", log.Fields{"signal": sig.String()})
	case <-ctx.Done():
		logger.Info("Context cancelled, shutting down")
	}

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := gatewayServer.Stop(shutdownCtx); err != nil {
		slog.Error("Server shutdown error", "error", err)
	}

	logger.Info("Gateway stopped")
}
