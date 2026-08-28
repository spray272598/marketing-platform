package server

import (
	"github.com/marketing-platform/internal/conf"
	"github.com/marketing-platform/internal/groupbuy/service"

	"github.com/go-kratos/kratos/v3/middleware/recovery"
	"github.com/go-kratos/kratos/v3/transport/http"
)

// NewHTTPServer new an HTTP server.
func NewHTTPServer(c *conf.Server, groupbuySvc *service.GroupBuyService) *http.Server {
	var opts = []http.ServerOption{
		http.Middleware(
			recovery.Recovery(),
		),
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

	srv.HandleFunc("/api/v1/groupbuy/activity/query", groupbuySvc.QueryGroupBuyActivityHTTP)
	srv.HandleFunc("/api/v1/groupbuy/trial", groupbuySvc.TrialGroupBuyMarketHTTP)
	srv.HandleFunc("/api/v1/groupbuy/lock", groupbuySvc.LockMarketPayOrderHTTP)
	srv.HandleFunc("/api/v1/groupbuy/settlement", groupbuySvc.SettlementMarketPayOrderHTTP)
	srv.HandleFunc("/api/v1/groupbuy/refund", groupbuySvc.RefundMarketPayOrderHTTP)
	srv.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok","service":"group-buy-market"}`))
	})

	return srv
}
