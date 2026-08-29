// Package auth 提供基于 JWT 的身份认证能力：签发令牌、校验令牌，
// 以及把解析出的 user_id 注入 context 的 HTTP 中间件 / gRPC 拦截器。
//
// 设计要点：
//   - 服务侧只信任令牌里解析出的 user_id，绝不使用请求体或 RPC 参数自带的 user_id，
//     否则任何调用方都能冒用他人身份（越权下单、越权查询）。
//   - Authenticator 是接口（依赖倒置），便于替换为其它实现或在测试中注入。
//   - 中间件是标准 net/http 风格（func(http.Handler) http.Handler），
//     与 pkg/middleware 的链路风格保持一致。
package auth

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/marketing-platform/pkg/common"
)

// 认证失败的错误值。注意这些只用于服务内部判断，不直接回传给客户端。
var (
	// ErrMissingToken 请求未携带 Bearer 令牌。
	ErrMissingToken = errors.New("auth: missing bearer token")
	// ErrInvalidToken 令牌无效（签名错误、过期、主体缺失等）。
	ErrInvalidToken = errors.New("auth: invalid token")
)

// DefaultTTL 签发令牌时若未指定有效期则使用的默认值。
const DefaultTTL = 2 * time.Hour

// Claims 是业务身份声明，除标准字段外只携带 user_id。
type Claims struct {
	UserID int64 `json:"user_id"`
	jwt.RegisteredClaims
}

// Authenticator 抽象令牌的签发与校验，便于替换实现与测试。
type Authenticator interface {
	// Sign 为指定用户签发一个 ttl 有效期的令牌。
	Sign(ctx context.Context, userID int64, ttl time.Duration) (string, error)
	// Verify 校验令牌并返回其中的 user_id。
	Verify(ctx context.Context, token string) (int64, error)
}

// JWTAuthenticator 是基于 HMAC-SHA256 的 JWT 实现。
type JWTAuthenticator struct {
	secret []byte
	issuer string
	ttl    time.Duration
}

// NewJWTAuthenticator 创建 JWT 签发器。secret 不能为空，否则无法保证令牌不可伪造。
func NewJWTAuthenticator(secret string, issuer string) (*JWTAuthenticator, error) {
	if secret == "" {
		return nil, errors.New("auth: signing secret must not be empty")
	}
	return &JWTAuthenticator{secret: []byte(secret), issuer: issuer, ttl: DefaultTTL}, nil
}

// Sign 签发令牌。ttl <= 0 时使用 DefaultTTL。
func (a *JWTAuthenticator) Sign(ctx context.Context, userID int64, ttl time.Duration) (string, error) {
	if userID <= 0 {
		return "", errors.New("auth: user id must be positive")
	}
	if ttl <= 0 {
		ttl = a.ttl
		if ttl <= 0 {
			ttl = DefaultTTL
		}
	}
	now := time.Now()
	claims := Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    a.issuer,
			Subject:   strconv.FormatInt(userID, 10),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(a.secret)
}

// Verify 校验令牌并返回 user_id。
//
// 安全约束：
//   - 强制要求 exp，拒绝永不过期的令牌；
//   - 只接受 HMAC 系列算法，防御 alg 混淆（如把 RS256 改成 HS256 或用 none）；
//   - 配置了 issuer 时校验签发者。
func (a *JWTAuthenticator) Verify(ctx context.Context, tokenStr string) (int64, error) {
	claims := &Claims{}
	opts := []jwt.ParserOption{jwt.WithExpirationRequired()}
	if a.issuer != "" {
		opts = append(opts, jwt.WithIssuer(a.issuer))
	}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("auth: unexpected signing method %v", t.Header["alg"])
		}
		return a.secret, nil
	}, opts...)
	if err != nil || !token.Valid {
		return 0, ErrInvalidToken
	}
	if claims.UserID <= 0 {
		return 0, ErrInvalidToken
	}
	return claims.UserID, nil
}

// ---------- context 传递身份 ----------

type userIDKey struct{}

// WithUserID 把 user_id 写入 context，供下游 biz/service 层取用。
func WithUserID(ctx context.Context, userID int64) context.Context {
	return context.WithValue(ctx, userIDKey{}, userID)
}

// UserID 从 context 取出已认证用户的 ID。未认证或值非法时返回 false。
func UserID(ctx context.Context) (int64, bool) {
	uid, ok := ctx.Value(userIDKey{}).(int64)
	if !ok || uid <= 0 {
		return 0, false
	}
	return uid, true
}

// ---------- HTTP 中间件 ----------

// SkipFunc 判断某个请求是否跳过鉴权（例如健康检查、登录接口）。
type SkipFunc func(r *http.Request) bool

// SkipPaths 生成按路径精确匹配的跳过规则。
func SkipPaths(paths ...string) SkipFunc {
	set := make(map[string]struct{}, len(paths))
	for _, p := range paths {
		set[p] = struct{}{}
	}
	return func(r *http.Request) bool {
		_, ok := set[r.URL.Path]
		return ok
	}
}

// Middleware 返回 net/http 风格的鉴权中间件：解析 Bearer 令牌，
// 校验通过后把 user_id 注入 context 再交给后续 handler；否则返回 401。
//
// 返回的中间件与 pkg/middleware 的链路风格一致，可直接用 Kratos 的
// http.Filter(...) 挂载（Kratos v3 的 http.Middleware() 不会作用于
// srv.HandleFunc 注册的原生 handler，故不能用它挂中间件）。
func Middleware(a Authenticator, skip ...SkipFunc) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			for _, s := range skip {
				if s != nil && s(r) {
					next.ServeHTTP(w, r)
					return
				}
			}

			token := extractBearer(r)
			if token == "" {
				common.WriteError(w, http.StatusUnauthorized, common.Unauthorized)
				return
			}
			uid, err := a.Verify(r.Context(), token)
			if err != nil {
				common.WriteError(w, http.StatusUnauthorized, common.Unauthorized)
				return
			}
			next.ServeHTTP(w, r.WithContext(WithUserID(r.Context(), uid)))
		})
	}
}

// ---------- 从环境变量装配 ----------

// 环境变量名。密钥属于敏感配置，不应硬编码进代码或提交进配置仓库。
const (
	// EnvSecret JWT 签名密钥（生产环境必须设置）。
	EnvSecret = "MARKETING_AUTH_SECRET"
	// EnvIssuer 可选，JWT 签发者标识。
	EnvIssuer = "MARKETING_AUTH_ISSUER"
	// EnvDisabled 设为 true 时完全关闭鉴权，仅供本地联调使用。
	EnvDisabled = "MARKETING_AUTH_DISABLED"
)

// devSecret 是本地开发用的默认密钥。生产环境务必通过 EnvSecret 注入强随机密钥，
// 否则任何人都能用该公开默认值伪造任意用户身份。
const devSecret = "marketing-platform-development-secret"

// EnvInternalToken 内部服务间调用使用的共享令牌。
const EnvInternalToken = "MARKETING_INTERNAL_TOKEN"

// InternalTokenHeader 内部令牌的请求头名称。
const InternalTokenHeader = "X-Internal-Token"

// InternalToken 保护"只对内部服务开放"的接口（如支付成功后的结算、退款回调）。
// 这类接口不接受用户令牌，必须校验共享的内部令牌，否则任何人都能对任意
// team_id / order_id 发起结算或退款。
//
// 令牌未配置时会拒绝全部请求（fail closed），避免误把内部接口暴露到公网。
// 返回的是 func(http.HandlerFunc) http.HandlerFunc，便于直接包裹
// srv.HandleFunc 注册的 handler。
func InternalToken(token string) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			if token == "" || subtleCompare(r.Header.Get(InternalTokenHeader), token) != 1 {
				common.WriteError(w, http.StatusForbidden, common.Forbidden)
				return
			}
			next(w, r)
		}
	}
}

// subtleCompare 以常量时间比较两个字符串，降低时序侧信道风险。
// 返回 1 表示相等，0 表示不等。
func subtleCompare(a, b string) int {
	if len(a) != len(b) {
		return 0
	}
	diff := byte(0)
	for i := 0; i < len(a); i++ {
		diff |= a[i] ^ b[i]
	}
	if diff == 0 {
		return 1
	}
	return 0
}

// NewFromEnv 依据环境变量构造 Authenticator：
//   - EnvDisabled=true 时返回 (nil, nil)，表示不启用鉴权；
//   - 设置了 EnvSecret 时使用该密钥；
//   - 否则回退到开发默认密钥并输出告警日志，保证本地可直接启动。
func NewFromEnv() (*JWTAuthenticator, error) {
	if strings.EqualFold(os.Getenv(EnvDisabled), "true") {
		slog.Warn("auth: authentication is DISABLED via " + EnvDisabled + ", never enable this in production")
		return nil, nil
	}
	secret := os.Getenv(EnvSecret)
	if secret == "" {
		slog.Warn("auth: " + EnvSecret + " is not set, falling back to the development secret; " +
			"set a strong random secret before deploying to production")
		secret = devSecret
	}
	return NewJWTAuthenticator(secret, os.Getenv(EnvIssuer))
}

// extractBearer 从 Authorization 头提取 Bearer 令牌，格式不符时返回空串。
func extractBearer(r *http.Request) string {
	header := r.Header.Get("Authorization")
	if header == "" {
		return ""
	}
	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}
	return strings.TrimSpace(parts[1])
}
