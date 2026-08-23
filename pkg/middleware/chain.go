package middleware

import (
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
}

func NewMiddlewareChain(logger *slog.Logger, metrics *observability.Metrics) *MiddlewareChain {
	return &MiddlewareChain{
		middlewares: make([]func(http.Handler) http.Handler, 0),
		logger:      logger,
		metrics:     metrics,
	}
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
		Use(TraceMiddleware(mc.logger)).
		Use(LoggingMiddleware(mc.logger)).
		Use(RecoveryMiddleware(mc.logger)).
		Use(RateLimitMiddleware(mc.logger)).
		Use(MetricsMiddleware(mc.metrics))
}

func TraceMiddleware(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			traceID := r.Header.Get("X-Trace-Id")
			if traceID == "" {
				traceID = uuid.New().String()
			}
			ctx := r.Context()
			_ = ctx

			w.Header().Set("X-Trace-Id", traceID)
			if logger != nil {
				logger.Info("request started",
					slog.String("trace_id", traceID),
					slog.String("method", r.Method),
					slog.String("path", r.URL.Path),
					slog.String("remote_addr", r.RemoteAddr),
				)
			}
			next.ServeHTTP(w, r)
		})
	}
}

func LoggingMiddleware(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

			next.ServeHTTP(wrapped, r)

			duration := time.Since(start)
			if logger != nil {
				logger.Info("request completed",
					slog.String("method", r.Method),
					slog.String("path", r.URL.Path),
					slog.Int("status", wrapped.statusCode),
					slog.Duration("duration", duration),
				)
			}
		})
	}
}

func RecoveryMiddleware(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					if logger != nil {
						logger.Error("panic recovered",
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

func RateLimitMiddleware(logger *slog.Logger) func(http.Handler) http.Handler {
	limiter := NewIPRateLimiter(100, 1000)
	go limiter.Cleanup(10 * time.Minute)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := getClientIP(r)
			if !limiter.Allow(ip) {
				if logger != nil {
					logger.Warn("rate limit exceeded",
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

func MetricsMiddleware(metrics *observability.Metrics) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if metrics != nil {
				metrics.RequestTotal.WithLabelValues(r.Method, r.URL.Path).Inc()

				start := time.Now()
				wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

				next.ServeHTTP(wrapped, r)

				duration := time.Since(start).Seconds()
				metrics.RequestDuration.WithLabelValues(r.Method, r.URL.Path).Observe(duration)
				metrics.RequestStatus.WithLabelValues(r.Method, r.URL.Path, fmt.Sprintf("%d", wrapped.statusCode)).Inc()
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
