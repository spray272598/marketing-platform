package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/marketing-platform/internal/stock/service"
	"github.com/marketing-platform/pkg/middleware"
	"github.com/marketing-platform/pkg/observability"
)

type StockServer struct {
	httpServer *http.Server
}

func NewStockServer(stockSvc *service.StockService, metrics *observability.Metrics, chain *middleware.MiddlewareChain) *StockServer {
	mux := http.NewServeMux()

	mux.HandleFunc("/api/v1/stock/query", stockSvc.GetStockHTTP)
	mux.HandleFunc("/api/v1/stock/deduct", stockSvc.DeductStockHTTP)
	mux.HandleFunc("/api/v1/stock/restore", stockSvc.RestoreStockHTTP)

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok","service":"stock-market"}`))
	})

	if metrics != nil {
		mux.Handle("/metrics", metrics.PrometheusHandler())
	}

	var handler http.Handler = mux
	if chain != nil {
		handler = chain.DefaultChain().Apply(mux)
	} else {
		defaultChain := middleware.NewMiddlewareChain(nil, metrics).DefaultChain()
		handler = defaultChain.Apply(mux)
	}

	srv := &http.Server{
		Addr:         ":18094",
		Handler:      handler,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	return &StockServer{httpServer: srv}
}

func (s *StockServer) Run() error {
	fmt.Println("Stock server listening on :18094")
	return s.httpServer.ListenAndServe()
}

func (s *StockServer) Stop(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}
