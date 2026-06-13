package forex

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"stock-market/backend/internal/cache"
)

const (
	frankfurterURL = "https://api.frankfurter.dev/v1/latest?from=USD&to=ILS"
	openErApiURL   = "https://open.er-api.com/v6/latest/USD"
	cacheKey       = "forex:USD:ILS"
	cacheTTL       = 15 * time.Minute
)

type Client struct {
	httpClient *http.Client
	cache      *cache.Redis
}

func NewClient(redisCache *cache.Redis) *Client {
	return &Client{
		httpClient: &http.Client{Timeout: 10 * time.Second},
		cache:      redisCache,
	}
}

func (c *Client) GetUSDToILS(ctx context.Context) (float64, error) {
	if c.cache != nil {
		var cached float64
		if ok, err := c.cache.GetJSON(ctx, cacheKey, &cached); err == nil && ok && cached > 0 {
			return cached, nil
		}
	}

	rate, err := c.fetchFrankfurter(ctx)
	if err != nil {
		rate, err = c.fetchOpenErAPI(ctx)
		if err != nil {
			return 0, fmt.Errorf("fetch USD/ILS rate: %w", err)
		}
	}

	if c.cache != nil {
		_ = c.cache.SetJSON(ctx, cacheKey, rate, cacheTTL)
	}

	return rate, nil
}

type frankfurterResponse struct {
	Rates map[string]float64 `json:"rates"`
}

func (c *Client) fetchFrankfurter(ctx context.Context) (float64, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, frankfurterURL, nil)
	if err != nil {
		return 0, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return 0, err
	}

	var response frankfurterResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return 0, fmt.Errorf("decode frankfurter response: %w", err)
	}

	rate, ok := response.Rates["ILS"]
	if !ok || rate <= 0 {
		return 0, fmt.Errorf("ILS rate missing from frankfurter response")
	}

	return rate, nil
}

type openErAPIResponse struct {
	Result string             `json:"result"`
	Rates  map[string]float64 `json:"rates"`
}

func (c *Client) fetchOpenErAPI(ctx context.Context) (float64, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, openErApiURL, nil)
	if err != nil {
		return 0, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return 0, err
	}

	var response openErAPIResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return 0, fmt.Errorf("decode open.er-api response: %w", err)
	}

	if response.Result != "success" {
		return 0, fmt.Errorf("open.er-api returned non-success result")
	}

	rate, ok := response.Rates["ILS"]
	if !ok || rate <= 0 {
		return 0, fmt.Errorf("ILS rate missing from open.er-api response")
	}

	return rate, nil
}

func (c *Client) doRequest(req *http.Request) ([]byte, error) {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("forex request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read forex response: %w", err)
	}

	if resp.StatusCode >= http.StatusBadRequest {
		return nil, fmt.Errorf("forex API error (%d): %s", resp.StatusCode, string(body))
	}

	return body, nil
}
