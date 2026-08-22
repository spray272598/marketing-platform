package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/marketing-platform/internal/gateway"
)

type Server struct {
	httpServer *http.Server
}

func NewServer(gatewaySvc *gateway.Service) *Server {
	mux := http.NewServeMux()

	mux.HandleFunc("/api/v1/gateway/proxy/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		serviceName := r.URL.Query().Get("service")
		path := r.URL.Query().Get("path")

		result, err := gatewaySvc.ProxyRequest(r.Context(), serviceName, r.Method, path, nil)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"code": "500",
				"info": err.Error(),
			})
			return
		}

		json.NewEncoder(w).Encode(result)
	})

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok","service":"gateway"}`))
	})

	srv := &http.Server{
		Addr:         ":8080",
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	return &Server{httpServer: srv}
}

func (s *Server) Run() error {
	fmt.Println("Gateway server listening on :8080")
	return s.httpServer.ListenAndServe()
}

func (s *Server) Stop(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}
