package main

import (
	"fmt"
	"os"

	"github.com/marketing-platform/internal/gateway"
	gwserver "github.com/marketing-platform/internal/gateway/server"
)

func main() {
	gatewaySvc := gateway.NewService()
	server := gwserver.NewServer(gatewaySvc)

	fmt.Println("Gateway starting...")
	if err := server.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to run server: %v\n", err)
		os.Exit(1)
	}
}
