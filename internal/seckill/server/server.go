package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/marketing-platform/internal/seckill/service"
)

type SeckillServer struct {
	httpServer *http.Server
}

func NewSeckillServer(seckillSvc *service.SeckillService) *SeckillServer {
	mux := http.NewServeMux()

	// Register HTTP handlers
	mux.HandleFunc("/api/v1/seckill/activity/query", seckillSvc.QuerySeckillActivityHTTP)
	mux.HandleFunc("/api/v1/seckill/order/create", seckillSvc.CreateSeckillOrderHTTP)
	mux.HandleFunc("/api/v1/seckill/order/query", seckillSvc.QuerySeckillOrderHTTP)

	// Health check
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok","service":"seckill-market"}`))
	})

	srv := &http.Server{
		Addr:         ":18091",
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	return &SeckillServer{httpServer: srv}
}

func (s *SeckillServer) Run() error {
	fmt.Println("Seckill server listening on :18091")
	return s.httpServer.ListenAndServe()
}

func (s *SeckillServer) Stop(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}
