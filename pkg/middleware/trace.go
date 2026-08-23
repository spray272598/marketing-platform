package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/marketing-platform/pkg/common"
)

func TraceMiddlewareHTTP(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		traceID := r.Header.Get("X-Trace-Id")
		if traceID == "" {
			traceID = uuid.New().String()
		}
		w.Header().Set("X-Trace-Id", traceID)

		ctx := common.WithTraceID(r.Context(), traceID)
		r = r.WithContext(ctx)

		start := time.Now()
		next.ServeHTTP(w, r)
		_ = time.Since(start)
	})
}

type traceKey struct{}

func TraceIDFromContext(ctx context.Context) string {
	if v, ok := ctx.Value(traceKey{}).(string); ok {
		return v
	}
	return ""
}
