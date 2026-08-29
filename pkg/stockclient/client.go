package stockclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/marketing-platform/pkg/auth"
)

type Client struct {
	baseURL    string
	httpClient *http.Client
	// internalToken 是访问 stock 内部服务时携带的共享令牌；
	// 未配置时不下发该请求头（stock 侧若要求令牌则会拒绝）。
	internalToken string
}

func NewClient(baseURL string) *Client {
	return &Client{
		baseURL:       baseURL,
		internalToken: os.Getenv(auth.EnvInternalToken),
		httpClient: &http.Client{
			Timeout: 3 * time.Second,
		},
	}
}

type DeductRequest struct {
	StockKey string `json:"stock_key"`
	Count    int32  `json:"count"`
}

type StockResponse struct {
	Code string      `json:"code"`
	Info string      `json:"info"`
	Data interface{} `json:"data,omitempty"`
}

// newRequest 创建请求并统一带上内部服务令牌。
func (c *Client) newRequest(ctx context.Context, method, url string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	if c.internalToken != "" {
		req.Header.Set(auth.InternalTokenHeader, c.internalToken)
	}
	return req, nil
}

// do 执行请求并解析统一的响应信封。
func (c *Client) do(req *http.Request) (*StockResponse, error) {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("stock service call failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	var stockResp StockResponse
	if err := json.Unmarshal(respBody, &stockResp); err != nil {
		return nil, fmt.Errorf("invalid stock response: %w", err)
	}
	return &stockResp, nil
}

func (c *Client) DeductStock(ctx context.Context, stockKey string, count int32) error {
	reqBody := DeductRequest{StockKey: stockKey, Count: count}
	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("marshal deduct request: %w", err)
	}

	req, err := c.newRequest(ctx, "POST", c.baseURL+"/api/v1/stock/deduct", bytes.NewReader(bodyBytes))
	if err != nil {
		return err
	}
	stockResp, err := c.do(req)
	if err != nil {
		return err
	}
	if stockResp.Code != "0000" {
		return fmt.Errorf("stock deduct failed: %s - %s", stockResp.Code, stockResp.Info)
	}
	return nil
}

func (c *Client) GetStock(ctx context.Context, stockKey string) (int32, error) {
	req, err := c.newRequest(ctx, "GET", c.baseURL+"/api/v1/stock/query?stock_key="+stockKey, nil)
	if err != nil {
		return 0, err
	}
	stockResp, err := c.do(req)
	if err != nil {
		return 0, err
	}
	if stockResp.Code != "0000" {
		return 0, fmt.Errorf("stock query failed: %s - %s", stockResp.Code, stockResp.Info)
	}

	if data, ok := stockResp.Data.(map[string]interface{}); ok {
		if stock, ok := data["stock"].(float64); ok {
			return int32(stock), nil
		}
	}
	return 0, nil
}

func (c *Client) RestoreStock(ctx context.Context, stockKey string, count int32) error {
	reqBody := DeductRequest{StockKey: stockKey, Count: count}
	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("marshal restore request: %w", err)
	}

	req, err := c.newRequest(ctx, "POST", c.baseURL+"/api/v1/stock/restore", bytes.NewReader(bodyBytes))
	if err != nil {
		return err
	}
	stockResp, err := c.do(req)
	if err != nil {
		return err
	}
	if stockResp.Code != "0000" {
		return fmt.Errorf("stock restore failed: %s - %s", stockResp.Code, stockResp.Info)
	}
	return nil
}
