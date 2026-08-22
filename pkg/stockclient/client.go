package stockclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Client struct {
	baseURL    string
	httpClient *http.Client
}

func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
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

func (c *Client) DeductStock(ctx context.Context, stockKey string, count int32) error {
	reqBody := DeductRequest{StockKey: stockKey, Count: count}
	bodyBytes, _ := json.Marshal(reqBody)

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/api/v1/stock/deduct", bytes.NewReader(bodyBytes))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("stock service call failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	var stockResp StockResponse
	if err := json.Unmarshal(respBody, &stockResp); err != nil {
		return fmt.Errorf("invalid stock response: %w", err)
	}

	if stockResp.Code != "0000" {
		return fmt.Errorf("stock deduct failed: %s - %s", stockResp.Code, stockResp.Info)
	}
	return nil
}

func (c *Client) GetStock(ctx context.Context, stockKey string) (int32, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+"/api/v1/stock/query?stock_key="+stockKey, nil)
	if err != nil {
		return 0, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("stock service call failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	var stockResp StockResponse
	if err := json.Unmarshal(respBody, &stockResp); err != nil {
		return 0, fmt.Errorf("invalid stock response: %w", err)
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
	bodyBytes, _ := json.Marshal(reqBody)

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/api/v1/stock/restore", bytes.NewReader(bodyBytes))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("stock service call failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	var stockResp StockResponse
	if err := json.Unmarshal(respBody, &stockResp); err != nil {
		return fmt.Errorf("invalid stock response: %w", err)
	}

	if stockResp.Code != "0000" {
		return fmt.Errorf("stock restore failed: %s - %s", stockResp.Code, stockResp.Info)
	}
	return nil
}
