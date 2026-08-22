package main

import (
	"fmt"
	"os"

	_ "go.uber.org/automaxprocs"
	_ "github.com/go-sql-driver/mysql"
)

func main() {
	lotteryServer, cleanup, err := InitializeLotteryServer()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to initialize server: %v\n", err)
		os.Exit(1)
	}
	defer cleanup()

	fmt.Println("Lottery market starting...")
	if err := lotteryServer.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to run server: %v\n", err)
		os.Exit(1)
	}
}
