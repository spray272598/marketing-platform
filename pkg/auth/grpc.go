package auth

import (
	"context"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// UnaryServerInterceptor 返回 gRPC 一元拦截器：从入站 metadata 的
// Authorization 头提取 Bearer 令牌，校验通过后把 user_id 注入 context。
//
// 行为与 HTTP 中间件保持一致——服务侧只信任令牌解析出的 user_id，绝不采信
// 请求体或 RPC 参数自带的 user_id。skip 用于放行不需要鉴权的 method
// （如健康检查）；内部接口请改用 InternalTokenUnaryInterceptor。
func UnaryServerInterceptor(a Authenticator, skip ...SkipMethod) grpc.UnaryServerInterceptor {
	// 鉴权关闭（authenticator 为 nil）时直接放行，与 HTTP 中间件行为一致。
	if a == nil {
		return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
			return handler(ctx, req)
		}
	}
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		for _, s := range skip {
			if s != nil && s(info.FullMethod) {
				return handler(ctx, req)
			}
		}
		token, err := extractBearerFromMD(ctx)
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, commonUnauthorizedMsg)
		}
		uid, err := a.Verify(ctx, token)
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, commonUnauthorizedMsg)
		}
		return handler(WithUserID(ctx, uid), req)
	}
}

// SkipMethod 判断某个 gRPC method（形如 /pkg.Service/Method）是否跳过鉴权。
type SkipMethod func(fullMethod string) bool

// SkipMethods 精确匹配若干 full method 跳过鉴权。
func SkipMethods(methods ...string) SkipMethod {
	set := make(map[string]struct{}, len(methods))
	for _, m := range methods {
		set[m] = struct{}{}
	}
	return func(fullMethod string) bool {
		_, ok := set[fullMethod]
		return ok
	}
}

// InternalTokenUnaryInterceptor 保护"只对内部服务开放"的 gRPC method：
// 从入站 metadata 的 X-Internal-Token 头校验共享内部令牌。令牌未配置时
// 拒绝全部请求（fail closed）。
func InternalTokenUnaryInterceptor(token string, methods ...string) grpc.UnaryServerInterceptor {
	set := make(map[string]struct{}, len(methods))
	for _, m := range methods {
		set[m] = struct{}{}
	}
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if _, ok := set[info.FullMethod]; !ok {
			return handler(ctx, req)
		}
		if token == "" {
			return nil, status.Error(codes.PermissionDenied, commonForbiddenMsg)
		}
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.PermissionDenied, commonForbiddenMsg)
		}
		values := md.Get(InternalTokenHeader)
		if len(values) == 0 || subtleCompare(values[0], token) != 1 {
			return nil, status.Error(codes.PermissionDenied, commonForbiddenMsg)
		}
		return handler(ctx, req)
	}
}

// extractBearerFromMD 从 gRPC metadata 的 Authorization 头提取 Bearer 令牌。
func extractBearerFromMD(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", errMissingToken
	}
	values := md.Get("authorization")
	if len(values) == 0 {
		return "", errMissingToken
	}
	header := values[0]
	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return "", errMissingToken
	}
	return strings.TrimSpace(parts[1]), nil
}

var (
	errMissingToken    = status.Error(codes.Unauthenticated, commonUnauthorizedMsg)
	commonUnauthorizedMsg = "auth: missing or invalid bearer token"
	commonForbiddenMsg    = "auth: internal token required"
)
