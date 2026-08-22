package gateway

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type ServiceConfig struct {
	Name    string
	BaseURL string
}

type Service struct {
	services map[string]*ServiceConfig
	client   *http.Client
}

func NewService() *Service {
	return &Service{
		services: map[string]*ServiceConfig{
			"seckill":  {Name: "seckill", BaseURL: "http://localhost:18091"},
			"groupbuy": {Name: "groupbuy", BaseURL: "http://localhost:18092"},
			"lottery":  {Name: "lottery", BaseURL: "http://localhost:18093"},
			"prize":    {Name: "prize", BaseURL: "http://localhost:18094"},
		},
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (s *Service) ProxyRequest(ctx context.Context, serviceName, method, path string, body interface{}) (map[string]interface{}, error) {
	svc, ok := s.services[serviceName]
	if !ok {
		return nil, fmt.Errorf("service not found: %s", serviceName)
	}

	url := svc.BaseURL + path

	var bodyReader io.Reader
	if body != nil {
		bodyBytes, _ := json.Marshal(body)
		bodyReader = bytes.NewReader(bodyBytes)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result, nil
}
