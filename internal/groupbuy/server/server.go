package server

import (
	"log/slog"
	"os"

	"github.com/marketing-platform/internal/conf"
	"github.com/marketing-platform/internal/groupbuy/service"
	"github.com/marketing-platform/pkg/auth"
	"github.com/marketing-platform/pkg/middleware"

	"github.com/go-kratos/kratos/v3/transport/http"
)

// internalPaths 是不面向终端用户、只接受内部服务令牌的接口。
var internalPaths = []string{
	"/api/v1/groupbuy/settlement",
	"/api/v1/groupbuy/refund",
}

// NewHTTPServer new an HTTP server.
func NewHTTPServer(c *conf.Server, groupbuySvc *service.GroupBuyService) *http.Server {
	// 注意：Kratos v3 的 http.Middleware() 不会作用于 srv.HandleFunc 注册的原生
	// handler，所以中间件必须统一走 http.Filter()（net/http 风格）才能生效。
	filters := []http.FilterFunc{middleware.Recovery()}

	authenticator, err := auth.NewFromEnv()
	if err != nil {
		slog.Error("auth: failed to initialize authenticator", slog.Any("error", err))
	} else if authenticator != nil {
		// /health 用于探活；内部接口改用内部服务令牌校验；/metrics 供 Prometheus 抓取。
		filters = append(filters, auth.Middleware(authenticator,
			auth.SkipPaths(append([]string{"/health", "/metrics"}, internalPaths...)...)))
	}

	// 请求级指标（QPS、耗时、状态码）随每个业务请求自增，供 Grafana 大盘消费。
	filters = append(filters, middleware.MetricsFilter("groupbuy-market"))

	var opts = []http.ServerOption{http.Filter(filters...)}
	if c.GetHttp().GetNetwork() != "" {
		opts = append(opts, http.Network(c.GetHttp().GetNetwork()))
	}
	if c.GetHttp().GetAddr() != "" {
		opts = append(opts, http.Address(c.GetHttp().GetAddr()))
	}
	if t := c.GetHttp().GetTimeout().AsDuration(); t != 0 {
		opts = append(opts, http.Timeout(t))
	}
	srv := http.NewServer(opts...)

	srv.HandleFunc("/api/v1/groupbuy/activity/query", groupbuySvc.QueryGroupBuyActivityHTTP)
	srv.HandleFunc("/api/v1/groupbuy/trial", groupbuySvc.TrialGroupBuyMarketHTTP)
	srv.HandleFunc("/api/v1/groupbuy/lock", groupbuySvc.LockMarketPayOrderHTTP)
	// 结算/退款可对任意 team_id / order_id 生效，必须校验内部服务令牌。
	requireInternal := auth.InternalToken(os.Getenv(auth.EnvInternalToken))
	srv.HandleFunc("/api/v1/groupbuy/settlement", requireInternal(groupbuySvc.SettlementMarketPayOrderHTTP))
	srv.HandleFunc("/api/v1/groupbuy/refund", requireInternal(groupbuySvc.RefundMarketPayOrderHTTP))
	srv.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok","service":"group-buy-market"}`))
	})
	srv.HandleFunc("/metrics", middleware.MetricsEndpoint().ServeHTTP)

	return srv
}
