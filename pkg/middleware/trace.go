package middleware

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/marketing-platform/pkg/common"
)

func TraceMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		traceID := c.GetHeader("X-Trace-Id")
		if traceID == "" {
			traceID = uuid.New().String()
		}
		c.Set("trace_id", traceID)
		c.Header("X-Trace-Id", traceID)

		ctx := common.WithTraceID(c.Request.Context(), traceID)
		c.Request = c.Request.WithContext(ctx)

		start := time.Now()
		c.Next()
		_ = time.Since(start)

		c.Header("X-Trace-Id", traceID)
	}
}

type traceKey struct{}

func TraceIDFromContext(ctx context.Context) string {
	if v, ok := ctx.Value(traceKey{}).(string); ok {
		return v
	}
	return ""
}
