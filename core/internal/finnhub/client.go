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

func (c *Client) GetProfile(ctx context.Context, symbol string) (*ProfileResponse, error) {
	endpoint := fmt.Sprintf("%s/stock/profile2?symbol=%s&token=%s", baseURL, url.QueryEscape(symbol), c.apiKey)
	return getJSON[ProfileResponse](c, ctx, endpoint, "profile")
}

func (c *Client) GetEarningsCalendar(ctx context.Context, from, to string) (*EarningsCalendarResponse, error) {
	endpoint := fmt.Sprintf("%s/calendar/earnings?from=%s&to=%s&token=%s", baseURL, from, to, c.apiKey)
	return getJSON[EarningsCalendarResponse](c, ctx, endpoint, "earnings calendar")
}

func (c *Client) GetEarningsSurprises(ctx context.Context, symbol string, limit int) ([]EarningsSurpriseEntry, error) {
	endpoint := fmt.Sprintf("%s/stock/earnings?symbol=%s&token=%s", baseURL, url.QueryEscape(symbol), c.apiKey)
	if limit > 0 {
		endpoint += fmt.Sprintf("&limit=%d", limit)
	}
	return getJSONSlice[EarningsSurpriseEntry](c, ctx, endpoint, "earnings surprises")
}

func (c *Client) GetCompanyNews(ctx context.Context, symbol, from, to string) ([]CompanyNewsEntry, error) {
	endpoint := fmt.Sprintf("%s/company-news?symbol=%s&from=%s&to=%s&token=%s", baseURL, url.QueryEscape(symbol), from, to, c.apiKey)
	return getJSONSlice[CompanyNewsEntry](c, ctx, endpoint, "company news")
}

func (c *Client) GetFilings(ctx context.Context, symbol string) ([]FilingEntry, error) {
	endpoint := fmt.Sprintf("%s/stock/filings?symbol=%s&token=%s", baseURL, url.QueryEscape(symbol), c.apiKey)
	return getJSONSlice[FilingEntry](c, ctx, endpoint, "filings")
}

func (c *Client) GetRecommendations(ctx context.Context, symbol string) ([]RecommendationEntry, error) {
	endpoint := fmt.Sprintf("%s/stock/recommendation?symbol=%s&token=%s", baseURL, url.QueryEscape(symbol), c.apiKey)
	return getJSONSlice[RecommendationEntry](c, ctx, endpoint, "recommendations")
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

func getJSON[T any](c *Client, ctx context.Context, endpoint, label string) (*T, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("create %s request: %w", label, err)
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	var response T
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("decode %s response: %w", label, err)
	}

	return &response, nil
}

func getJSONSlice[T any](c *Client, ctx context.Context, endpoint, label string) ([]T, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("create %s request: %w", label, err)
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	var response []T
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("decode %s response: %w", label, err)
	}

	return response, nil
}
