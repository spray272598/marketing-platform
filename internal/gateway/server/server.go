package server

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/marketing-platform/internal/conf"
	"github.com/marketing-platform/internal/gateway"

	"github.com/go-kratos/kratos/v3/middleware/recovery"
	kratoshttp "github.com/go-kratos/kratos/v3/transport/http"
)

func NewHTTPServer(c *conf.Server, gatewaySvc *gateway.Service) *kratoshttp.Server {
	var opts = []kratoshttp.ServerOption{
		kratoshttp.Middleware(recovery.Recovery()),
	}
	if c.GetHttp().GetNetwork() != "" {
		opts = append(opts, kratoshttp.Network(c.GetHttp().GetNetwork()))
	}
	if c.GetHttp().GetAddr() != "" {
		opts = append(opts, kratoshttp.Address(c.GetHttp().GetAddr()))
	}
	if c.GetHttp().GetTimeout() != 0 {
		opts = append(opts, kratoshttp.Timeout(c.GetHttp().GetTimeout()))
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
		result, err := gatewaySvc.ProxyRequest(r.Context(), serviceName, r.Method, path, nil)
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

	return srv
}
