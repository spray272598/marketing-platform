package main

import (
	"context"
	"os"

	"github.com/marketing-platform/internal/gateway"
	gwserver "github.com/marketing-platform/internal/gateway/server"
	"github.com/marketing-platform/pkg/common"
	"github.com/marketing-platform/pkg/config"
	"github.com/marketing-platform/pkg/log"
)

func main() {
	cfg := config.LoadOrDefault("")
	logger := log.NewLogger(cfg.Log.Level, cfg.Log.Format)

	gatewaySvc := gateway.NewService()
	gatewayServer := gwserver.NewServer(gatewaySvc)

	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	go common.WaitForShutdown(cancel)

	logger.Info("Gateway starting...", log.Fields{"addr": ":8080"})
	if err := gatewayServer.Run(); err != nil {
		logger.Error("failed to run server", log.Fields{"error": err})
		os.Exit(1)
	}
}
