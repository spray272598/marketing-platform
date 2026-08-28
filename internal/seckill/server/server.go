package server

import (
	"github.com/marketing-platform/internal/conf"
	"github.com/marketing-platform/internal/seckill/service"

	"github.com/go-kratos/kratos/v3/middleware/recovery"
	"github.com/go-kratos/kratos/v3/transport/http"
)

// NewHTTPServer new an HTTP server.
func NewHTTPServer(c *conf.Server, seckillSvc *service.SeckillService) *http.Server {
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

	// Register HTTP routes
	srv.HandleFunc("/api/v1/seckill/activity/query", seckillSvc.QuerySeckillActivityHTTP)
	srv.HandleFunc("/api/v1/seckill/order/create", seckillSvc.CreateSeckillOrderHTTP)
	srv.HandleFunc("/api/v1/seckill/order/query", seckillSvc.QuerySeckillOrderHTTP)
	srv.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok","service":"seckill-market"}`))
	})

	return srv
}
