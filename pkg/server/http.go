package server

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

type ServerOption struct {
	Addr         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	Middlewares  []TransportMiddleware
}

type TransportMiddleware func(handler http.Handler) http.Handler

type KratosHTTPServer struct {
	server *http.Server
	mux    *http.ServeMux
}

func NewKratosHTTPServer(cfg ServerOption) *KratosHTTPServer {
	mux := http.NewServeMux()

	var handler http.Handler = mux
	for i := len(cfg.Middlewares) - 1; i >= 0; i-- {
		handler = cfg.Middlewares[i](handler)
	}

	srv := &http.Server{
		Addr:         cfg.Addr,
		Handler:      handler,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
	}

	return &KratosHTTPServer{
		server: srv,
		mux:    mux,
	}
}

func (s *KratosHTTPServer) HandleFunc(pattern string, handlerFunc http.HandlerFunc) {
	s.mux.HandleFunc(pattern, handlerFunc)
}

func (s *KratosHTTPServer) Handle(pattern string, handler http.Handler) {
	s.mux.Handle(pattern, handler)
}

func (s *KratosHTTPServer) Run() error {
	fmt.Printf("HTTP server listening on %s\n", s.server.Addr)
	return s.server.ListenAndServe()
}

func (s *KratosHTTPServer) Stop(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}

func (s *KratosHTTPServer) Endpoint() string {
	return s.server.Addr
}
