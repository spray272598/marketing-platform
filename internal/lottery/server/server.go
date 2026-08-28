package server

import (
	"github.com/marketing-platform/internal/conf"
	"github.com/marketing-platform/internal/lottery/service"

	"github.com/go-kratos/kratos/v3/middleware/recovery"
	"github.com/go-kratos/kratos/v3/transport/http"
)

func NewHTTPServer(c *conf.Server, lotterySvc *service.LotteryService) *http.Server {
	var opts = []http.ServerOption{
		http.Middleware(recovery.Recovery()),
	}
	if c.GetHttp().GetNetwork() != "" {
		opts = append(opts, http.Network(c.GetHttp().GetNetwork()))
	}
	if c.GetHttp().GetAddr() != "" {
		opts = append(opts, http.Address(c.GetHttp().GetAddr()))
	}
	if c.GetHttp().GetTimeout() != 0 {
		opts = append(opts, http.Timeout(c.GetHttp().GetTimeout()))
	}
	srv := http.NewServer(opts...)

	srv.HandleFunc("/api/v1/lottery/activity/query", lotterySvc.QueryLotteryActivityHTTP)
	srv.HandleFunc("/api/v1/lottery/strategy/query", lotterySvc.QueryLotteryStrategyHTTP)
	srv.HandleFunc("/api/v1/lottery/raffle", lotterySvc.RaffleHTTP)
	srv.HandleFunc("/api/v1/lottery/order/query", lotterySvc.QueryUserRaffleOrderHTTP)
	srv.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok","service":"lottery-market"}`))
	})

	return srv
}
