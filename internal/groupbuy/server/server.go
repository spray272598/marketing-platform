package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/marketing-platform/internal/groupbuy/service"
)

type GroupBuyServer struct {
	httpServer *http.Server
}

func NewGroupBuyServer(groupBuySvc *service.GroupBuyService) *GroupBuyServer {
	mux := http.NewServeMux()

	mux.HandleFunc("/api/v1/groupbuy/activity/query", groupBuySvc.QueryGroupBuyActivityHTTP)
	mux.HandleFunc("/api/v1/groupbuy/trial", groupBuySvc.TrialGroupBuyMarketHTTP)
	mux.HandleFunc("/api/v1/groupbuy/order/lock", groupBuySvc.LockMarketPayOrderHTTP)
	mux.HandleFunc("/api/v1/groupbuy/order/settlement", groupBuySvc.SettlementMarketPayOrderHTTP)
	mux.HandleFunc("/api/v1/groupbuy/order/refund", groupBuySvc.RefundMarketPayOrderHTTP)

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok","service":"groupbuy-market"}`))
	})

	srv := &http.Server{
		Addr:         ":18092",
		Handler:      mux,
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
