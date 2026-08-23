package middleware

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/marketing-platform/pkg/observability"
)

type MiddlewareChain struct {
	middlewares []func(http.Handler) http.Handler
	logger      *slog.Logger
	metrics     *observability.Metrics
	traceCollector *observability.TraceCollector
}

func NewMiddlewareChain(logger *slog.Logger, metrics *observability.Metrics) *MiddlewareChain {
	return &MiddlewareChain{
		middlewares:    make([]func(http.Handler) http.Handler, 0),
		logger:         logger,
		metrics:        metrics,
		traceCollector: nil,
	}
}

func (mc *MiddlewareChain) SetTraceCollector(tc *observability.TraceCollector) {
	mc.traceCollector = tc
}

func (mc *MiddlewareChain) Use(mw func(http.Handler) http.Handler) *MiddlewareChain {
	mc.middlewares = append(mc.middlewares, mw)
	return mc
}

func (mc *MiddlewareChain) Apply(handler http.Handler) http.Handler {
	for i := len(mc.middlewares) - 1; i >= 0; i-- {
		handler = mc.middlewares[i](handler)
	}
	return handler
}

func (mc *MiddlewareChain) DefaultChain() *MiddlewareChain {
	return mc.
		Use(mc.TraceMiddleware()).
		Use(mc.LoggingMiddleware()).
		Use(mc.RecoveryMiddleware()).
		Use(mc.RateLimitMiddleware()).
		Use(mc.MetricsMiddleware())
}

func (mc *MiddlewareChain) TraceMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 从请求头获取或生成 TraceID
			traceID := r.Header.Get("X-Trace-Id")
			if traceID == "" {
				traceID = uuid.New().String()
			}

			// 将 TraceID 写入 context
			ctx := r.Context()
			ctx = contextWithTraceID(ctx, traceID)
			r = r.WithContext(ctx)

			// 如果有 TraceCollector，创建根 Span
			var span *observability.TraceContext
			if mc.traceCollector != nil {
				span = mc.traceCollector.StartRootSpan(ctx, r.URL.Path)
				span.AddTag("http.method", r.Method)
				span.AddTag("http.path", r.URL.Path)
				span.AddTag("remote_addr", r.RemoteAddr)
			}

			// 设置响应头
			w.Header().Set("X-Trace-Id", traceID)

			if mc.logger != nil {
				mc.logger.Info("request started",
					slog.String("trace_id", traceID),
					slog.String("method", r.Method),
					slog.String("path", r.URL.Path),
					slog.String("remote_addr", r.RemoteAddr),
				)
			}

			// 执行后续处理
			if span != nil {
				defer mc.traceCollector.CompleteTrace(span)
			}
			next.ServeHTTP(w, r)
		})
	}
}

func (mc *MiddlewareChain) LoggingMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

			next.ServeHTTP(wrapped, r)

			duration := time.Since(start)
			traceID := traceIDFromContext(r.Context())

			if mc.logger != nil {
				mc.logger.Info("request completed",
					slog.String("trace_id", traceID),
					slog.String("method", r.Method),
					slog.String("path", r.URL.Path),
					slog.Int("status", wrapped.statusCode),
					slog.Duration("duration", duration),
				)
			}
		})
	}
}

func (mc *MiddlewareChain) RecoveryMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					traceID := traceIDFromContext(r.Context())
					if mc.logger != nil {
						mc.logger.Error("panic recovered",
							slog.String("trace_id", traceID),
							slog.Any("error", err),
							slog.String("path", r.URL.Path),
						)
					}
					http.Error(w, fmt.Sprintf("Internal Server Error: %v", err), http.StatusInternalServerError)
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}

func (mc *MiddlewareChain) RateLimitMiddleware() func(http.Handler) http.Handler {
	limiter := NewIPRateLimiter(100, 1000)
	go limiter.Cleanup(10 * time.Minute)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := getClientIP(r)
			if !limiter.Allow(ip) {
				traceID := traceIDFromContext(r.Context())
				if mc.logger != nil {
					mc.logger.Warn("rate limit exceeded",
						slog.String("trace_id", traceID),
						slog.String("ip", ip),
						slog.String("path", r.URL.Path),
					)
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusTooManyRequests)
				fmt.Fprint(w, `{"code":"429","info":"请求过于频繁"}`)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func (mc *MiddlewareChain) MetricsMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if mc.metrics != nil {
				mc.metrics.RequestTotal.WithLabelValues(r.Method, r.URL.Path).Inc()

				start := time.Now()
				wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

				next.ServeHTTP(wrapped, r)

				duration := time.Since(start).Seconds()
				mc.metrics.RequestDuration.WithLabelValues(r.Method, r.URL.Path).Observe(duration)
				mc.metrics.RequestStatus.WithLabelValues(r.Method, r.URL.Path, fmt.Sprintf("%d", wrapped.statusCode)).Inc()
			} else {
				next.ServeHTTP(w, r)
			}
		})
	}
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func getClientIP(r *http.Request) string {
	ip := r.Header.Get("X-Forwarded-For")
	if ip != "" {
		return ip
	}
	ip = r.Header.Get("X-Real-IP")
	if ip != "" {
		return ip
	}
	return r.RemoteAddr
}

// Context trace_id 相关辅助函数
type traceIDKey struct{}

func contextWithTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, traceIDKey{}, traceID)
}

func traceIDFromContext(ctx context.Context) string {
	if traceID, ok := ctx.Value(traceIDKey{}).(string); ok {
		return traceID
	}
	return ""
}