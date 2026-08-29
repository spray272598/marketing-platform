package server

import (
	"context"
	"net"
	"net/url"

	"github.com/go-kratos/kratos/v3/transport"
	"github.com/marketing-platform/internal/conf"
	"github.com/marketing-platform/internal/lottery/service"
	v1 "github.com/marketing-platform/api/lottery/v1"
	"github.com/marketing-platform/pkg/auth"
	"github.com/marketing-platform/pkg/middleware"
	"google.golang.org/grpc"
	"log/slog"
)

// grpcServerWrapper 把 *grpc.Server 适配为 kratos 的 transport.Server。
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

var _ transport.Server = (*grpcServerWrapper)(nil)

// NewGRPCServer 创建并配置 lottery 的 gRPC server：
//   - recovery 拦截器兜底 panic；
//   - JWT 拦截器从 metadata 取 Bearer 并注入 user_id（与 HTTP 中间件一致）；
//   - 注册 LotteryService 实现。
func NewGRPCServer(c *conf.Server, lotterySvc *service.LotteryService, logger *slog.Logger) transport.Server {
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
	v1.RegisterLotteryServiceServer(gs, lotterySvc)

	addr := c.GetGrpc().GetAddr()
	if addr == "" {
		addr = ":18097"
	}
	return &grpcServerWrapper{Server: gs, addr: addr}
}
