package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/marketing-platform/internal/lottery/service"
	"github.com/marketing-platform/pkg/middleware"
	"github.com/marketing-platform/pkg/observability"
)

type LotteryServer struct {
	httpServer *http.Server
}

func NewLotteryServer(lotterySvc *service.LotteryService, metrics *observability.Metrics, chain *middleware.MiddlewareChain) *LotteryServer {
	mux := http.NewServeMux()

	mux.HandleFunc("/api/v1/lottery/activity/query", lotterySvc.QueryLotteryActivityHTTP)
	mux.HandleFunc("/api/v1/lottery/strategy/query", lotterySvc.QueryLotteryStrategyHTTP)
	mux.HandleFunc("/api/v1/lottery/raffle", lotterySvc.RaffleHTTP)
	mux.HandleFunc("/api/v1/lottery/order/query", lotterySvc.QueryUserRaffleOrderHTTP)

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok","service":"lottery-market"}`))
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
		Addr:         ":18093",
		Handler:      handler,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	return &LotteryServer{httpServer: srv}
}

func (s *LotteryServer) Run() error {
	fmt.Println("Lottery server listening on :18093")
	return s.httpServer.ListenAndServe()
}

func (s *LotteryServer) Stop(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}
