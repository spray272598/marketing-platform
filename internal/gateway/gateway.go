package gateway

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"sync"
	"time"
)

type ServiceConfig struct {
	Name      string   `json:"name"`
	Endpoints []string `json:"endpoints"`
}

type Service struct {
	services map[string]*ServiceConfig
	client   *http.Client
	mu       sync.RWMutex
}

func NewService() *Service {
	return &Service{
		services: map[string]*ServiceConfig{
			"seckill":  {Name: "seckill", Endpoints: []string{"http://localhost:18091"}},
			"groupbuy": {Name: "groupbuy", Endpoints: []string{"http://localhost:18092"}},
			"lottery":  {Name: "lottery", Endpoints: []string{"http://localhost:18093"}},
			"stock":    {Name: "stock", Endpoints: []string{"http://localhost:18094"}},
		},
		client: &http.Client{
			Timeout: 10 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 10,
				IdleConnTimeout:     90 * time.Second,
			},
		},
	}
}

func (s *Service) RegisterService(name string, endpoints []string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.services[name] = &ServiceConfig{
		Name:      name,
		Endpoints: endpoints,
	}
}

func (s *Service) UnregisterService(name string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.services, name)
}

func (s *Service) GetService(name string) (*ServiceConfig, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	svc, ok := s.services[name]
	if !ok {
		return nil, fmt.Errorf("service not found: %s", name)
	}
	if len(svc.Endpoints) == 0 {
		return nil, fmt.Errorf("no endpoints for service: %s", name)
	}
	return svc, nil
}

func (s *Service) SelectEndpoint(svc *ServiceConfig) string {
	if len(svc.Endpoints) == 1 {
		return svc.Endpoints[0]
	}
	return svc.Endpoints[rand.Intn(len(svc.Endpoints))]
}

type ServiceError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (s *Service) ProxyRequest(ctx context.Context, serviceName, method, path string, body interface{}) (map[string]interface{}, error) {
	svc, err := s.GetService(serviceName)
	if err != nil {
		return nil, err
	}

	endpoint := s.SelectEndpoint(svc)
	url := endpoint + path

	var bodyReader io.Reader
	if body != nil {
		bodyBytes, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(bodyBytes)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Gateway-Service", serviceName)

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("proxy request failed: %w", err)
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return result, nil
}

func (s *Service) ListServices() map[string]*ServiceConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make(map[string]*ServiceConfig)
	for k, v := range s.services {
		result[k] = v
	}
	return result
}
