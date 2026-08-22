package main

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/go-sql-driver/mysql"

	"github.com/marketing-platform/internal/prize/biz"
	"github.com/marketing-platform/internal/prize/data"
	"github.com/marketing-platform/internal/prize/service"
	prizeserver "github.com/marketing-platform/internal/prize/server"
)

func main() {
	mysqlDSN := getEnv("MYSQL_DSN", "root:root@tcp(127.0.0.1:3306)/marketing_prize?charset=utf8mb4&parseTime=True&loc=Local")

	db, err := sql.Open("mysql", mysqlDSN)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to open db: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	dataLayer := data.NewData(db)
	prizeRepo := data.NewPrizeRepo(dataLayer)
	prizeSvc := biz.NewPrizeService(prizeRepo)
	svc := service.NewPrizeService(prizeSvc)

	server := prizeserver.NewServer(svc)

	fmt.Println("Prize starting...")
	if err := server.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to run server: %v\n", err)
		os.Exit(1)
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
