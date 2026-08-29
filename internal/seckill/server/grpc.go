package server

import (
	"context"
	"net"
	"net/url"

	"github.com/go-kratos/kratos/v3/transport"
	"github.com/marketing-platform/internal/conf"
	"github.com/marketing-platform/internal/seckill/service"
	"github.com/marketing-platform/pkg/auth"
	v1 "github.com/marketing-platform/api/seckill/v1"
	"github.com/marketing-platform/pkg/middleware"
	"google.golang.org/grpc"
	"log/slog"
)

// grpcServerWrapper 把 *grpc.Server 适配为 kratos 的 transport.Server，
// 使其能直接接入 kratos.App 的生命周期（Start/Stop/Endpoint）。
type grpcServerWrapper struct {
	*grpc.Server
	addr string
}

func (w *grpcServerWrapper) Start(ctx context.Context) error {
	lis, err := net.Listen("tcp", w.addr)
	if err != nil {
		return err
	}
	return w.Serve(lis)
}

func (w *grpcServerWrapper) Stop(ctx context.Context) error {
	w.GracefulStop()
	return nil
}

func (w *grpcServerWrapper) Endpoint() (*url.URL, error) {
	return url.Parse("grpc://" + w.addr)
}

// NewGRPCServer 创建并配置 seckill 的 gRPC server：
//   - recovery 拦截器兜底 panic；
//   - JWT 拦截器从 metadata 取 Bearer 并注入 user_id（与 HTTP 中间件一致）；
//   - 注册 SeckillService 实现。
func NewGRPCServer(c *conf.Server, seckillSvc *service.SeckillService, logger *slog.Logger) transport.Server {
	authenticator, err := auth.NewFromEnv()
	if err != nil {
		slog.Error("auth: failed to initialize authenticator for grpc", slog.Any("error", err))
	}

	opts := []grpc.ServerOption{
		grpc.ChainUnaryInterceptor(
			middleware.RecoveryUnaryInterceptor(logger),
			auth.UnaryServerInterceptor(authenticator),
		),
	}

	gs := grpc.NewServer(opts...)
	v1.RegisterSeckillServiceServer(gs, seckillSvc)

	addr := c.GetGrpc().GetAddr()
	if addr == "" {
		addr = ":18095"
	}
	return &grpcServerWrapper{Server: gs, addr: addr}
}

// 确保编译期满足 kratos transport.Server 接口。
var _ transport.Server = (*grpcServerWrapper)(nil)
