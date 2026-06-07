package finnhub

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const baseURL = "https://finnhub.io/api/v1"

type Client struct {
	apiKey     string
	httpClient *http.Client
}

func NewClient(apiKey string) *Client {
	return &Client{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *Client) SearchSymbols(ctx context.Context, query string) (*SearchResponse, error) {
	endpoint := fmt.Sprintf("%s/search?q=%s&token=%s", baseURL, url.QueryEscape(query), c.apiKey)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("create search request: %w", err)
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	var response SearchResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("decode search response: %w", err)
	}

	return &response, nil
}

func (c *Client) GetQuote(ctx context.Context, symbol string) (*QuoteResponse, error) {
	endpoint := fmt.Sprintf("%s/quote?symbol=%s&token=%s", baseURL, url.QueryEscape(symbol), c.apiKey)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("create quote request: %w", err)
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	var response QuoteResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("decode quote response: %w", err)
	}

	return &response, nil
}

func (c *Client) doRequest(req *http.Request) ([]byte, error) {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("finnhub request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read finnhub response: %w", err)
	}

	if resp.StatusCode >= http.StatusBadRequest {
		message := strings.TrimSpace(string(body))
		if message == "" {
			message = resp.Status
		}
		return nil, fmt.Errorf("finnhub API error (%d): %s", resp.StatusCode, message)
	}

	return body, nil
}
