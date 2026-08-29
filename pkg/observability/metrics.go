package observability

import (
	"net/http"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Metrics struct {
	RequestTotal    *prometheus.CounterVec
	RequestDuration *prometheus.HistogramVec
	RequestStatus   *prometheus.CounterVec
	ServiceInfo     prometheus.Gauge
	DBConnections   prometheus.Gauge
	CacheHitRate    prometheus.Gauge
	BusinessMetrics map[string]*prometheus.GaugeVec
}

func NewMetrics(serviceName string) *Metrics {
	// Prometheus 指标名只允许 [a-zA-Z_:][a-zA-Z0-9_:]*，服务名里常见的连字符
	// （如 "seckill-market"）会导致 promauto 注册时直接 panic。统一替换为下划线，
	// 既保证安全，也让指标前缀与 Prometheus 的 job 标签保持一致。
	name := sanitizeMetricName(serviceName)
	return &Metrics{
		RequestTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: name + "_http_requests_total",
				Help: "Total number of HTTP requests",
			},
			[]string{"method", "path"},
		),
		RequestDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    name + "_http_request_duration_seconds",
				Help:    "Duration of HTTP requests in seconds",
				Buckets: []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
			},
			[]string{"method", "path"},
		),
		RequestStatus: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: name + "_http_requests_status_total",
				Help: "Total number of HTTP requests by status",
			},
			[]string{"method", "path", "status"},
		),
		ServiceInfo: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: name + "_service_info",
				Help: "Service information",
				ConstLabels: prometheus.Labels{
					"service": name,
				},
			},
		),
		DBConnections: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: name + "_db_connections_active",
				Help: "Number of active database connections",
			},
		),
		CacheHitRate: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: name + "_cache_hit_rate",
				Help: "Cache hit rate",
			},
		),
		BusinessMetrics: make(map[string]*prometheus.GaugeVec),
	}
}

// sanitizeMetricName 把任意服务名规整为合法的 Prometheus 指标名前缀：
// 非 [a-zA-Z0-9_] 的字符统一替换为下划线，并保证首字符合法。
func sanitizeMetricName(s string) string {
	var b strings.Builder
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' {
			b.WriteRune(r)
		} else {
			b.WriteRune('_')
		}
	}
	out := b.String()
	if out == "" {
		return "svc"
	}
	first := rune(out[0])
	if !(first == '_' || (first >= 'a' && first <= 'z') || (first >= 'A' && first <= 'Z')) {
		return "svc_" + out
	}
	return out
}

func (m *Metrics) PrometheusHandler() http.Handler {
	return promhttp.Handler()
}

func (m *Metrics) RegisterBusinessMetric(name, help string, labels []string) *prometheus.GaugeVec {
	gauge := promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: name,
			Help: help,
		},
		labels,
	)
	m.BusinessMetrics[name] = gauge
	return gauge
}

func (m *Metrics) SetServiceInfo(info float64) {
	m.ServiceInfo.Set(info)
}

func (m *Metrics) SetDBConnections(n float64) {
	m.DBConnections.Set(n)
}

func (m *Metrics) SetCacheHitRate(rate float64) {
	m.CacheHitRate.Set(rate)
}
