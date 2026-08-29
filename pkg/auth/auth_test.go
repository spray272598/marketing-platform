package auth

import (
	"context"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// b64url 按 JWT 规范做无填充的 base64url 编码，用于手工构造畸形令牌。
func b64url(s string) string {
	return base64.RawURLEncoding.EncodeToString([]byte(s))
}

const (
	testSecret = "unit-test-secret"
	testIssuer = "marketing-platform"
)

func newTestAuthenticator(t *testing.T) *JWTAuthenticator {
	t.Helper()
	a, err := NewJWTAuthenticator(testSecret, testIssuer)
	if err != nil {
		t.Fatalf("NewJWTAuthenticator: %v", err)
	}
	return a
}

func TestNewJWTAuthenticator_EmptySecret(t *testing.T) {
	if _, err := NewJWTAuthenticator("", testIssuer); err == nil {
		t.Fatal("expected error for empty secret")
	}
}

func TestSignVerify_RoundTrip(t *testing.T) {
	a := newTestAuthenticator(t)
	token, err := a.Sign(context.Background(), 10086, time.Hour)
	if err != nil {
		t.Fatalf("Sign: %v", err)
	}
	uid, err := a.Verify(context.Background(), token)
	if err != nil {
		t.Fatalf("Verify: %v", err)
	}
	if uid != 10086 {
		t.Fatalf("expected user 10086, got %d", uid)
	}
}

func TestSign_InvalidUserID(t *testing.T) {
	a := newTestAuthenticator(t)
	if _, err := a.Sign(context.Background(), 0, time.Hour); err == nil {
		t.Fatal("expected error when signing user_id=0")
	}
}

func TestVerify_Expired(t *testing.T) {
	a := newTestAuthenticator(t)
	// 手动构造一个已过期的令牌。
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		UserID: 42,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    testIssuer,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
		},
	})
	s, err := tok.SignedString([]byte(testSecret))
	if err != nil {
		t.Fatalf("SignedString: %v", err)
	}
	if _, err := a.Verify(context.Background(), s); err != ErrInvalidToken {
		t.Fatalf("expected ErrInvalidToken for expired token, got %v", err)
	}
}

func TestVerify_WrongSecret(t *testing.T) {
	a := newTestAuthenticator(t)
	other, err := NewJWTAuthenticator("another-secret", testIssuer)
	if err != nil {
		t.Fatalf("NewJWTAuthenticator: %v", err)
	}
	token, err := other.Sign(context.Background(), 7, time.Hour)
	if err != nil {
		t.Fatalf("Sign: %v", err)
	}
	if _, err := a.Verify(context.Background(), token); err != ErrInvalidToken {
		t.Fatalf("expected ErrInvalidToken for wrong secret, got %v", err)
	}
}

func TestVerify_TamperedPayload(t *testing.T) {
	a := newTestAuthenticator(t)
	token, err := a.Sign(context.Background(), 100, time.Hour)
	if err != nil {
		t.Fatalf("Sign: %v", err)
	}
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		t.Fatalf("unexpected token shape: %v", parts)
	}
	// 替换 payload 段但保留原签名，签名校验必须失败。
	tampered := parts[0] + ".eyJ1c2VyX2lkIjo5OTk5OTk5OTl9." + parts[2]
	if _, err := a.Verify(context.Background(), tampered); err != ErrInvalidToken {
		t.Fatalf("expected ErrInvalidToken for tampered payload, got %v", err)
	}
}

func TestVerify_AlgNoneRejected(t *testing.T) {
	a := newTestAuthenticator(t)
	// 构造 alg=none 的无签名令牌（经典降级攻击）。
	token := b64url(`{"alg":"none","typ":"JWT"}`) + "." +
		b64url(`{"user_id":1,"iss":"`+testIssuer+`","exp":9999999999}`) + "."
	if _, err := a.Verify(context.Background(), token); err != ErrInvalidToken {
		t.Fatalf("expected ErrInvalidToken for alg=none, got %v", err)
	}
}

func TestVerify_MissingToken(t *testing.T) {
	a := newTestAuthenticator(t)
	if _, err := a.Verify(context.Background(), ""); err != ErrInvalidToken {
		t.Fatalf("expected ErrInvalidToken for empty token, got %v", err)
	}
}

func TestUserID_Context(t *testing.T) {
	ctx := context.Background()
	if _, ok := UserID(ctx); ok {
		t.Fatal("expected no user id in empty context")
	}
	ctx = WithUserID(ctx, 555)
	uid, ok := UserID(ctx)
	if !ok || uid != 555 {
		t.Fatalf("expected 555/true, got %d/%v", uid, ok)
	}
}

// newTestHandler 返回记录"是否通过鉴权"以及"拿到的 user_id"的探针 handler。
func newTestHandler(saw *bool, uid *int64) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if id, ok := UserID(r.Context()); ok {
			*uid = id
			*saw = true
		}
		w.WriteHeader(http.StatusOK)
	})
}

func TestMiddleware_ValidToken(t *testing.T) {
	a := newTestAuthenticator(t)
	token, _ := a.Sign(context.Background(), 2024, time.Hour)

	var saw bool
	var uid int64
	h := Middleware(a)(newTestHandler(&saw, &uid))

	req := httptest.NewRequest(http.MethodPost, "/api/v1/lock", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if !saw || uid != 2024 {
		t.Fatalf("expected handler to see user 2024, got saw=%v uid=%d", saw, uid)
	}
}

func TestMiddleware_MissingToken(t *testing.T) {
	a := newTestAuthenticator(t)
	var saw bool
	var uid int64
	h := Middleware(a)(newTestHandler(&saw, &uid))

	req := httptest.NewRequest(http.MethodPost, "/api/v1/lock", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
	if saw {
		t.Fatal("handler must not run without a valid token")
	}
}

func TestMiddleware_InvalidToken(t *testing.T) {
	a := newTestAuthenticator(t)
	var saw bool
	var uid int64
	h := Middleware(a)(newTestHandler(&saw, &uid))

	req := httptest.NewRequest(http.MethodPost, "/api/v1/lock", nil)
	req.Header.Set("Authorization", "Bearer not-a-real-token")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestMiddleware_MalformedAuthorizationHeader(t *testing.T) {
	a := newTestAuthenticator(t)
	var saw bool
	var uid int64
	h := Middleware(a)(newTestHandler(&saw, &uid))

	for _, header := range []string{"", "Basic abc", "Bearer", "bearer   "} {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/lock", nil)
		if header != "" {
			req.Header.Set("Authorization", header)
		}
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("header %q: expected 401, got %d", header, rec.Code)
		}
	}
}

func TestMiddleware_SkipPaths(t *testing.T) {
	a := newTestAuthenticator(t)
	var saw bool
	var uid int64
	h := Middleware(a, SkipPaths("/health"))(newTestHandler(&saw, &uid))

	// /health 跳过鉴权，handler 应直接执行（无身份）。
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected /health to pass without token, got %d", rec.Code)
	}

	// 其它路径仍要求鉴权。
	req2 := httptest.NewRequest(http.MethodPost, "/api/v1/lock", nil)
	rec2 := httptest.NewRecorder()
	h.ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for protected path, got %d", rec2.Code)
	}
}
