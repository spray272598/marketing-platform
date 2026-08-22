package common

import (
	"context"
	"os"
	"os/signal"
	"syscall"
)

func WaitForShutdown(cancel context.CancelFunc) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	cancel()
}
