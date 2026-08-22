package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/marketing-platform/internal/prize/service"
)

type Server struct {
	httpServer *http.Server
}

func NewServer(prizeSvc *service.PrizeService) *Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/prize/stock/deduct", prizeSvc.DeductStockHTTP)
	mux.HandleFunc("/api/v1/prize/stock/query", prizeSvc.GetStockHTTP)
	mux.HandleFunc("/api/v1/prize/stock/restore", prizeSvc.RestoreStockHTTP)
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok","service":"prize"}`))
	})
	srv := &http.Server{Addr: ":18094", Handler: mux, ReadTimeout: 5 * time.Second, WriteTimeout: 10 * time.Second}
	return &Server{httpServer: srv}
}

func (s *Server) Run() error {
	fmt.Println("Prize server listening on :18094")
	return s.httpServer.ListenAndServe()
}

func (s *Server) Stop(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}
