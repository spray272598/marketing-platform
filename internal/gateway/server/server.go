package server

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/marketing-platform/internal/conf"
	"github.com/marketing-platform/internal/gateway"
	"github.com/marketing-platform/pkg/middleware"

	kratoshttp "github.com/go-kratos/kratos/v3/transport/http"
)

func NewHTTPServer(c *conf.Server, gatewaySvc *gateway.Service) *kratoshttp.Server {
	// 注意：Kratos v3 的 http.Middleware() 不会作用于 srv.HandleFunc 注册的原生
	// handler，所以中间件必须统一走 http.Filter()（net/http 风格）才能生效。
	// 请求级指标（QPS、耗时、状态码）随每个业务请求自增，供 Grafana 大盘消费。
	var opts = []kratoshttp.ServerOption{kratoshttp.Filter(middleware.Recovery(), middleware.MetricsFilter("gateway"))}
	if c.GetHttp().GetNetwork() != "" {
		opts = append(opts, kratoshttp.Network(c.GetHttp().GetNetwork()))
	}
	if c.GetHttp().GetAddr() != "" {
		opts = append(opts, kratoshttp.Address(c.GetHttp().GetAddr()))
	}
	if t := c.GetHttp().GetTimeout().AsDuration(); t != 0 {
		opts = append(opts, kratoshttp.Timeout(t))
	}
	srv := kratoshttp.NewServer(opts...)

	srv.HandleFunc("/api/v1/gateway/proxy/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		serviceName := r.URL.Query().Get("service")
		path := r.URL.Query().Get("path")
		if serviceName == "" || path == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]interface{}{"code": "400", "message": "missing service or path param"})
			return
		}
		// 转发原始请求体与请求头：否则后端既收不到参数，也无法用用户令牌鉴权。
		result, err := gatewaySvc.ProxyRequest(r.Context(), serviceName, r.Method, path, r.Body, r.Header)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]interface{}{"code": "500", "message": err.Error()})
			return
		}
		json.NewEncoder(w).Encode(result)
	})

	srv.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "ok", "service": "gateway", "time": time.Now().Format(time.RFC3339),
		})
	})
	srv.HandleFunc("/metrics", middleware.MetricsEndpoint().ServeHTTP)

	return srv
}
