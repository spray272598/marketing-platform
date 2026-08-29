package middleware

import (
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/marketing-platform/pkg/observability"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// metricsCache 按服务名缓存 Metrics 实例，保证每个服务只向默认 Prometheus
// 注册表注册一次自定义 collector（重复注册会直接 panic）。
var (
	metricsMu    sync.Mutex
	metricsCache = map[string]*observability.Metrics{}
)

func metricsFor(serviceName string) *observability.Metrics {
	metricsMu.Lock()
	defer metricsMu.Unlock()
	if m, ok := metricsCache[serviceName]; ok {
		return m
	}
	m := observability.NewMetrics(serviceName)
	metricsCache[serviceName] = m
	return m
}

// MetricsEndpoint 返回 Prometheus 抓取端点 handler，暴露默认注册表中的全部指标
// （Go runtime 指标 + 本项目自定义指标）。直接挂到 /metrics 即可被 Prometheus 抓取。
func MetricsEndpoint() http.Handler {
	return promhttp.Handler()
}

// MetricsFilter 是 net/http 风格的请求级指标中间件：对业务请求计数、记录耗时
// 直方图与状态码分布。探活（/health）与指标自身端点（/metrics）不计入业务指标，
// 避免把心跳、抓取流量污染到业务大盘里。
func MetricsFilter(serviceName string) func(http.Handler) http.Handler {
	m := metricsFor(serviceName)
	skip := map[string]struct{}{"/health": {}, "/metrics": {}}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if _, ok := skip[r.URL.Path]; ok {
				next.ServeHTTP(w, r)
				return
			}
			m.RequestTotal.WithLabelValues(r.Method, r.URL.Path).Inc()
			start := time.Now()
			wr := &metricsStatusWriter{ResponseWriter: w, status: http.StatusOK}
			next.ServeHTTP(wr, r)
			m.RequestDuration.WithLabelValues(r.Method, r.URL.Path).Observe(time.Since(start).Seconds())
			m.RequestStatus.WithLabelValues(r.Method, r.URL.Path, strconv.Itoa(wr.status)).Inc()
		})
	}
}

// metricsStatusWriter 捕获下游 handler 写入的 HTTP 状态码，用于状态码分布统计。
type metricsStatusWriter struct {
	http.ResponseWriter
	status int
}

func (sw *metricsStatusWriter) WriteHeader(code int) {
	sw.status = code
	sw.ResponseWriter.WriteHeader(code)
}
