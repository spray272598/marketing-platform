package middleware

import (
	"context"
	"log/slog"
	"runtime/debug"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// RecoveryUnaryInterceptor 是 gRPC 一元拦截器版的 recover：捕获 handler 中的
// panic，记录堆栈并返回 codes.Internal，避免单个请求把整个 gRPC server 拖垮。
// 对应 HTTP 侧的 pkg/middleware.Recovery()。
func RecoveryUnaryInterceptor(logger *slog.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		defer func() {
			if r := recover(); r != nil {
				if logger != nil {
					logger.Error("grpc recovery: panic recovered",
						slog.String("method", info.FullMethod),
						slog.Any("panic", r),
						slog.String("stack", string(debug.Stack())),
					)
				}
				err = status.Errorf(codes.Internal, "internal server error")
			}
		}()
		return handler(ctx, req)
	}
}
