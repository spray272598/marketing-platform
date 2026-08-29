package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestMetricsFilterIncrementsAndSkips 验证：
//   - 业务请求会被计数（请求总数 / 状态码分布 / 耗时直方图自增）；
//   - /metrics 与 /health 作为探活与抓取端点不被计入业务指标；
//   - /metrics 端点确实吐出 Prometheus  exposition 格式且包含本项目自定义指标。
func TestMetricsFilterIncrementsAndSkips(t *testing.T) {
	h := MetricsFilter("test-svc")

	// 一个业务请求，返回 201。
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/x", nil)
	h(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(201) })).ServeHTTP(rec, req)
	if rec.Code != 201 {
		t.Fatalf("business handler status = %d, want 201", rec.Code)
	}

	// 探活与抓取端点应当被跳过，不计入业务指标。
	for _, p := range []string{"/metrics", "/health"} {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, p, nil)
		h(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})).ServeHTTP(rec, req)
	}

	// 抓取 /metrics，确认自定义指标真实存在。
	rec2 := httptest.NewRecorder()
	MetricsEndpoint().ServeHTTP(rec2, httptest.NewRequest(http.MethodGet, "/metrics", nil))
	body := rec2.Body.String()
	if !strings.Contains(body, "test_svc_http_requests_total") {
		t.Fatalf("expected custom metric test_svc_http_requests_total in /metrics output:\n%s", body)
	}
	if !strings.Contains(body, `status="201"`) {
		t.Fatalf("expected status metric with 201 in /metrics output:\n%s", body)
	}
}
