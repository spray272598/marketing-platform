package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/marketing-platform/internal/gateway"
	"github.com/marketing-platform/pkg/middleware"
	"github.com/marketing-platform/pkg/observability"
)

type Server struct {
	httpServer *http.Server
	gatewaySvc *gateway.Service
	metrics    *observability.Metrics
}

func NewServer(gatewaySvc *gateway.Service, metrics *observability.Metrics) *Server {
	mux := http.NewServeMux()

	s := &Server{
		gatewaySvc: gatewaySvc,
		metrics:    metrics,
	}

	mux.HandleFunc("/api/v1/gateway/proxy/", s.handleProxy)
	mux.HandleFunc("/api/v1/gateway/route", s.handleRouteConfig)

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "ok",
			"service": "gateway",
			"time":    time.Now().Format(time.RFC3339),
		})
	})

	mux.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "ready",
		})
	})

	if metrics != nil {
		mux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			fmt.Fprintf(w, "# HELP marketing_platform_metrics Marketing Platform Metrics\n")
			fmt.Fprintf(w, "# TYPE marketing_platform_up gauge\n")
			fmt.Fprintf(w, "marketing_platform_up 1\n")
		})
	}

	chain := middleware.NewMiddlewareChain(
		nil,
		metrics,
	).DefaultChain()

	handler := chain.Apply(mux)

	srv := &http.Server{
		Addr:         ":8080",
		Handler:      handler,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return &Server{
		httpServer: srv,
	}
}

func (s *Server) handleProxy(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	serviceName := r.URL.Query().Get("service")
	path := r.URL.Query().Get("path")

	if serviceName == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"code":    "400",
			"message": "缺少 service 参数",
		})
		return
	}

	if path == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"code":    "400",
			"message": "缺少 path 参数",
		})
		return
	}

	if s.metrics != nil {
		s.metrics.RequestTotal.WithLabelValues(r.Method, "/api/v1/gateway/proxy").Inc()
	}

	result, err := s.gatewaySvc.ProxyRequest(r.Context(), serviceName, r.Method, path, nil)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"code":    "500",
			"message": err.Error(),
		})
		return
	}

	json.NewEncoder(w).Encode(result)
}

func (s *Server) handleRouteConfig(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case http.MethodGet:
		services := make([]map[string]interface{}, 0)
		allSvcs := s.gatewaySvc.ListServices()
		for name, svc := range allSvcs {
			services = append(services, map[string]interface{}{
				"name":      name,
				"endpoints": svc.Endpoints,
			})
		}
		json.NewEncoder(w).Encode(map[string]interface{}{
			"services": services,
		})

	case http.MethodPost:
		var req struct {
			Name      string   `json:"name"`
			Endpoints []string `json:"endpoints"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"code":    "400",
				"message": err.Error(),
			})
			return
		}
		s.gatewaySvc.RegisterService(req.Name, req.Endpoints)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"code":    "200",
			"message": "Service registered successfully",
		})

	case http.MethodDelete:
		serviceName := r.URL.Query().Get("service")
		if serviceName == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"code":    "400",
				"message": "缺少 service 参数",
			})
			return
		}
		s.gatewaySvc.UnregisterService(serviceName)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"code":    "200",
			"message": "Service unregistered successfully",
		})
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (s *Server) Run() error {
	fmt.Printf("Gateway server listening on %s\n", s.httpServer.Addr)
	return s.httpServer.ListenAndServe()
}

func (s *Server) Stop(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}
