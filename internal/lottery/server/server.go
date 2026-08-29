package server

import (
	"log/slog"

	"github.com/marketing-platform/internal/conf"
	"github.com/marketing-platform/internal/lottery/service"
	"github.com/marketing-platform/pkg/auth"
	"github.com/marketing-platform/pkg/middleware"

	"github.com/go-kratos/kratos/v3/transport/http"
)

func NewHTTPServer(c *conf.Server, lotterySvc *service.LotteryService) *http.Server {
	// 注意：Kratos v3 的 http.Middleware() 不会作用于 srv.HandleFunc 注册的原生
	// handler，所以中间件必须统一走 http.Filter()（net/http 风格）才能生效。
	filters := []http.FilterFunc{middleware.Recovery()}

	authenticator, err := auth.NewFromEnv()
	if err != nil {
		slog.Error("auth: failed to initialize authenticator", slog.Any("error", err))
	} else if authenticator != nil {
		// 探活、活动/策略浏览接口与 /metrics 抓取端点无需登录；抽奖与查单必须携带令牌。
		filters = append(filters, auth.Middleware(authenticator, auth.SkipPaths(
			"/health",
			"/metrics",
			"/api/v1/lottery/activity/query",
			"/api/v1/lottery/strategy/query",
		)))
	}

	// 请求级指标（QPS、耗时、状态码）随每个业务请求自增，供 Grafana 大盘消费。
	filters = append(filters, middleware.MetricsFilter("lottery-market"))

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

	srv.HandleFunc("/api/v1/lottery/activity/query", lotterySvc.QueryLotteryActivityHTTP)
	srv.HandleFunc("/api/v1/lottery/strategy/query", lotterySvc.QueryLotteryStrategyHTTP)
	srv.HandleFunc("/api/v1/lottery/raffle", lotterySvc.RaffleHTTP)
	srv.HandleFunc("/api/v1/lottery/order/query", lotterySvc.QueryUserRaffleOrderHTTP)
	srv.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok","service":"lottery-market"}`))
	})
	srv.HandleFunc("/metrics", middleware.MetricsEndpoint().ServeHTTP)

	return srv
}
