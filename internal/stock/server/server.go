package server

import (
	"log/slog"
	nethttp "net/http"
	"os"

	"github.com/marketing-platform/internal/conf"
	"github.com/marketing-platform/internal/stock/service"
	"github.com/marketing-platform/pkg/auth"
	"github.com/marketing-platform/pkg/middleware"

	"github.com/go-kratos/kratos/v3/transport/http"
)

func NewHTTPServer(c *conf.Server, stockSvc *service.StockService) *http.Server {
	// 注意：Kratos v3 的 http.Middleware() 不会作用于 srv.HandleFunc 注册的原生
	// handler，所以中间件必须统一走 http.Filter()（net/http 风格）才能生效。
	var opts = []http.ServerOption{http.Filter(middleware.Recovery())}
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

	// stock 是内部服务，只应被其它微服务通过 stockclient 调用。
	// 配置了内部令牌时强制校验；未配置则保持原样（内网部署假设）并告警。
	internalToken := os.Getenv(auth.EnvInternalToken)
	// 注意用 net/http 的 HandlerFunc：Kratos 里的同名类型是独立定义，二者不兼容。
	wrap := func(h nethttp.HandlerFunc) nethttp.HandlerFunc {
		if internalToken == "" {
			return h
		}
		return auth.InternalToken(internalToken)(h)
	}
	if internalToken == "" {
		slog.Warn("stock: " + auth.EnvInternalToken + " is not set; stock endpoints accept unauthenticated calls. " +
			"set it so only internal services can mutate stock")
	}

	srv.HandleFunc("/api/v1/stock/query", wrap(stockSvc.GetStockHTTP))
	srv.HandleFunc("/api/v1/stock/deduct", wrap(stockSvc.DeductStockHTTP))
	srv.HandleFunc("/api/v1/stock/restore", wrap(stockSvc.RestoreStockHTTP))
	srv.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok","service":"stock-market"}`))
	})

	return srv
}
