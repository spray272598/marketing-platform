package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/marketing-platform/internal/groupbuy/service"
	"github.com/marketing-platform/pkg/middleware"
	"github.com/marketing-platform/pkg/observability"
)

type GroupBuyServer struct {
	httpServer *http.Server
}

func NewGroupBuyServer(groupbuySvc *service.GroupBuyService, metrics *observability.Metrics, chain *middleware.MiddlewareChain) *GroupBuyServer {
	mux := http.NewServeMux()

	mux.HandleFunc("/api/v1/groupbuy/activity/query", groupbuySvc.QueryGroupBuyActivityHTTP)
	mux.HandleFunc("/api/v1/groupbuy/trial", groupbuySvc.TrialGroupBuyMarketHTTP)
	mux.HandleFunc("/api/v1/groupbuy/lock", groupbuySvc.LockMarketPayOrderHTTP)
	mux.HandleFunc("/api/v1/groupbuy/settlement", groupbuySvc.SettlementMarketPayOrderHTTP)
	mux.HandleFunc("/api/v1/groupbuy/refund", groupbuySvc.RefundMarketPayOrderHTTP)

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok","service":"group-buy-market"}`))
	})

	if metrics != nil {
		mux.Handle("/metrics", metrics.PrometheusHandler())
	}

	var handler http.Handler = mux
	if chain != nil {
		handler = chain.Apply(mux)
	} else {
		defaultChain := middleware.NewMiddlewareChain(nil, metrics).DefaultChain()
		handler = defaultChain.Apply(mux)
	}

	srv := &http.Server{
		Addr:         ":18092",
		Handler:      handler,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	return &GroupBuyServer{httpServer: srv}
}

func (s *GroupBuyServer) Run() error {
	fmt.Println("GroupBuy server listening on :18092")
	return s.httpServer.ListenAndServe()
}

func (s *GroupBuyServer) Stop(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}
