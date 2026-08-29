package server

import (
	"context"
	"net"
	"net/url"
	"os"

	"github.com/go-kratos/kratos/v3/transport"
	"github.com/marketing-platform/internal/conf"
	"github.com/marketing-platform/internal/groupbuy/service"
	v1 "github.com/marketing-platform/api/groupbuy/v1"
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

// groupbuyInternalMethods 是仅允许内部服务调用的 gRPC method，
// 通过 X-Internal-Token 校验，不接受用户 JWT。
var groupbuyInternalMethods = []string{
	"/groupbuy.v1.GroupBuyService/SettlementMarketPayOrder",
	"/groupbuy.v1.GroupBuyService/RefundMarketPayOrder",
}

// NewGRPCServer 创建并配置 groupbuy 的 gRPC server：
//   - recovery 拦截器兜底 panic；
//   - JWT 拦截器作用于用户侧 RPC，内部 RPC（结算/退款）跳过；
//   - 内部 RPC 由 InternalToken 拦截器用共享令牌校验；
//   - 注册 GroupBuyService 实现。
func NewGRPCServer(c *conf.Server, groupbuySvc *service.GroupBuyService, logger *slog.Logger) transport.Server {
	authenticator, err := auth.NewFromEnv()
	if err != nil {
		slog.Error("auth: failed to initialize authenticator for grpc", slog.Any("error", err))
	}
	internalToken := os.Getenv(auth.EnvInternalToken)

	opts := []grpc.ServerOption{
		grpc.ChainUnaryInterceptor(
			middleware.RecoveryUnaryInterceptor(logger),
			auth.UnaryServerInterceptor(authenticator, auth.SkipMethods(groupbuyInternalMethods...)),
			auth.InternalTokenUnaryInterceptor(internalToken, groupbuyInternalMethods...),
		),
	}

	gs := grpc.NewServer(opts...)
	v1.RegisterGroupBuyServiceServer(gs, groupbuySvc)

	addr := c.GetGrpc().GetAddr()
	if addr == "" {
		addr = ":18096"
	}
	return &grpcServerWrapper{Server: gs, addr: addr}
}
