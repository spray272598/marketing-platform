package observability

import (
	"net/http"

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
	return &Metrics{
		RequestTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: serviceName + "_http_requests_total",
				Help: "Total number of HTTP requests",
			},
			[]string{"method", "path"},
		),
		RequestDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    serviceName + "_http_request_duration_seconds",
				Help:    "Duration of HTTP requests in seconds",
				Buckets: []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
			},
			[]string{"method", "path"},
		),
		RequestStatus: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: serviceName + "_http_requests_status_total",
				Help: "Total number of HTTP requests by status",
			},
			[]string{"method", "path", "status"},
		),
		ServiceInfo: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: serviceName + "_service_info",
				Help: "Service information",
				ConstLabels: prometheus.Labels{
					"service": serviceName,
				},
			},
		),
		DBConnections: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: serviceName + "_db_connections_active",
				Help: "Number of active database connections",
			},
		),
		CacheHitRate: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: serviceName + "_cache_hit_rate",
				Help: "Cache hit rate",
			},
		),
		BusinessMetrics: make(map[string]*prometheus.GaugeVec),
	}
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
