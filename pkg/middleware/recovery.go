package middleware

import (
	"log/slog"
	"net/http"

	"github.com/marketing-platform/pkg/common"
)

// Recovery 返回独立的 net/http panic 恢复中间件，不依赖 MiddlewareChain，
// 便于直接挂到 Kratos 的 http.Filter 上。
//
// 背景：Kratos v3 里 http.Middleware() 配置的中间件（如 recovery.Recovery()）
// 不会作用于 srv.HandleFunc 注册的原生 handler，因此原先的挂载实际并未生效。
// 只有改成 http.Filter(Recovery()) 才能真正兜住 panic，避免一次异常打挂整个进程。
func Recovery() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					slog.Error("panic recovered",
						slog.String("method", r.Method),
						slog.String("path", r.URL.Path),
						slog.Any("error", err),
					)
					// 只回传通用错误码，不泄露 panic 细节。
					common.WriteError(w, http.StatusInternalServerError, common.InternalError)
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}
