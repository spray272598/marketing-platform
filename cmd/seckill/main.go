package main

import (
	"fmt"
	"os"

	_ "go.uber.org/automaxprocs"
	_ "github.com/go-sql-driver/mysql"
)

func main() {
	seckillServer, cleanup, err := InitializeSeckillServer()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to initialize server: %v\n", err)
		os.Exit(1)
	}
	defer cleanup()

	fmt.Println("Seckill market starting...")
	if err := seckillServer.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to run server: %v\n", err)
		os.Exit(1)
	}
}
