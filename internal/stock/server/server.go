package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/marketing-platform/internal/stock/service"
)

type Server struct {
	httpServer *http.Server
}

func NewServer(stockSvc *service.StockService) *Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/stock/deduct", stockSvc.DeductStockHTTP)
	mux.HandleFunc("/api/v1/stock/query", stockSvc.GetStockHTTP)
	mux.HandleFunc("/api/v1/stock/restore", stockSvc.RestoreStockHTTP)
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok","service":"stock"}`))
	})
	srv := &http.Server{Addr: ":18094", Handler: mux, ReadTimeout: 5 * time.Second, WriteTimeout: 10 * time.Second}
	return &Server{httpServer: srv}
}

func (s *Server) Run() error {
	fmt.Println("Stock server listening on :18094")
	return s.httpServer.ListenAndServe()
}

func (s *Server) Stop(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}
