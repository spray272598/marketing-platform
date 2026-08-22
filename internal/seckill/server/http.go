package server

import (
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/transport/http"
	"github.com/marketing-platform/internal/seckill/service"
)

type HTTPServer struct {
	*http.Server
}

func NewHTTPServer(seckillSvc *service.SeckillService) *HTTPServer {
	srv := http.NewServer(
		http.Address(":18091"),
		http.Middleware(
			recovery.Recovery(),
		),
	)
	_ = seckillSvc
	return &HTTPServer{Server: srv}
}
