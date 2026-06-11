package finnhubprovider

import (
	"context"
	"fmt"
	"strings"
	"time"

	"stock-market/backend/internal/cache"
	finnhubclient "stock-market/backend/internal/finnhub"
	"stock-market/backend/internal/provider/marketdata"
)

type Provider struct {
	client *finnhubclient.Client
	cache  *cache.Redis
}

func New(client *finnhubclient.Client, cache *cache.Redis) *Provider {
	return &Provider{client: client, cache: cache}
}

func (p *Provider) Search(ctx context.Context, query string) ([]marketdata.SearchResult, error) {
	search, err := p.client.SearchSymbols(ctx, query)
	if err != nil {
		return nil, err
	}

	results := make([]marketdata.SearchResult, 0, len(search.Result))
	for _, item := range search.Result {
		results = append(results, marketdata.SearchResult{
			Symbol:      item.Symbol,
			CompanyName: item.Description,
			Exchange:    extractExchange(item.DisplaySymbol, item.Symbol),
			Industry:    "",
			Type:        item.Type,
		})
	}

	return results, nil
}

func (p *Provider) GetQuote(ctx context.Context, symbol string) (*marketdata.Quote, error) {
	cacheKey := "quote:" + strings.ToUpper(symbol)
	var cached marketdata.Quote
	if p.cache != nil {
		if ok, err := p.cache.GetJSON(ctx, cacheKey, &cached); err == nil && ok {
			return &cached, nil
		}
	}

	quote, err := p.client.GetQuote(ctx, symbol)
	if err != nil {
		return nil, err
	}

	result := mapQuote(symbol, quote)
	if p.cache != nil {
		_ = p.cache.SetJSON(ctx, cacheKey, result, 30*time.Second)
	}

	return result, nil
}

func (p *Provider) GetProfile(ctx context.Context, symbol string) (*marketdata.CompanyProfile, error) {
	cacheKey := "profile:" + strings.ToUpper(symbol)
	var cached marketdata.CompanyProfile
	if p.cache != nil {
		if ok, err := p.cache.GetJSON(ctx, cacheKey, &cached); err == nil && ok && cached.Name != "" {
			return &cached, nil
		}
	}

	profile, err := p.client.GetProfile(ctx, symbol)
	if err != nil {
		return nil, err
	}

	result := mapProfile(symbol, profile)
	if p.cache != nil {
		_ = p.cache.SetJSON(ctx, cacheKey, result, 24*time.Hour)
	}

	return result, nil
}

func (p *Provider) GetStockDetails(ctx context.Context, symbol string) (*marketdata.StockDetails, error) {
	quote, err := p.GetQuote(ctx, symbol)
	if err != nil {
		return nil, err
	}

	profile, err := p.GetProfile(ctx, symbol)
	if err != nil {
		profile = &marketdata.CompanyProfile{Symbol: strings.ToUpper(symbol)}
	}

	return &marketdata.StockDetails{
		Symbol:      strings.ToUpper(symbol),
		CompanyName: profile.Name,
		Quote:       *quote,
		Profile:     *profile,
	}, nil
}

func (p *Provider) GetEarningsCalendar(ctx context.Context, from, to string) ([]marketdata.EarningsEvent, error) {
	cacheKey := fmt.Sprintf("earnings:%s:%s", from, to)
	var cached []marketdata.EarningsEvent
	if p.cache != nil {
		if ok, err := p.cache.GetJSON(ctx, cacheKey, &cached); err == nil && ok {
			return cached, nil
		}
	}

	response, err := p.client.GetEarningsCalendar(ctx, from, to)
	if err != nil {
		return nil, err
	}

	results := make([]marketdata.EarningsEvent, 0, len(response.EarningsCalendar))
	for _, item := range response.EarningsCalendar {
		results = append(results, marketdata.EarningsEvent{
			Symbol:          item.Symbol,
			CompanyName:     item.Symbol,
			Date:            item.Date,
			Hour:            item.Hour,
			EPSActual:       item.EPSActual,
			EPSEstimate:     item.EPSEstimate,
			RevenueActual:   item.RevenueActual,
			RevenueEstimate: item.RevenueEstimate,
			Quarter:         item.Quarter,
			Year:            item.Year,
		})
	}

	if p.cache != nil {
		_ = p.cache.SetJSON(ctx, cacheKey, results, 6*time.Hour)
	}

	return results, nil
}

func (p *Provider) GetEarningsSurprises(ctx context.Context, symbol string, limit int) ([]marketdata.EarningsSurprise, error) {
	normalized := strings.ToUpper(strings.TrimSpace(symbol))
	if normalized == "" {
		return nil, fmt.Errorf("symbol is required")
	}
	if limit <= 0 {
		limit = 8
	}

	cacheKey := fmt.Sprintf("earnings-surprises:%s:%d", normalized, limit)
	var cached []marketdata.EarningsSurprise
	if p.cache != nil {
		if ok, err := p.cache.GetJSON(ctx, cacheKey, &cached); err == nil && ok {
			return cached, nil
		}
	}

	entries, err := p.client.GetEarningsSurprises(ctx, normalized, limit)
	if err != nil {
		return nil, err
	}

	results := make([]marketdata.EarningsSurprise, 0, len(entries))
	for _, item := range entries {
		results = append(results, marketdata.EarningsSurprise{
			Symbol:             item.Symbol,
			Period:             item.Period,
			Quarter:            item.Quarter,
			Year:               item.Year,
			EPSActual:          item.Actual,
			EPSEstimate:        item.Estimate,
			EPSSurprise:        item.Surprise,
			EPSSurprisePercent: item.SurprisePercent,
		})
	}

	if p.cache != nil {
		_ = p.cache.SetJSON(ctx, cacheKey, results, 6*time.Hour)
	}

	return results, nil
}

func (p *Provider) GetCompanyNews(ctx context.Context, symbol, from, to string) ([]marketdata.NewsArticle, error) {
	cacheKey := fmt.Sprintf("news:%s:%s:%s", strings.ToUpper(symbol), from, to)
	var cached []marketdata.NewsArticle
	if p.cache != nil {
		if ok, err := p.cache.GetJSON(ctx, cacheKey, &cached); err == nil && ok {
			return cached, nil
		}
	}

	entries, err := p.client.GetCompanyNews(ctx, symbol, from, to)
	if err != nil {
		return nil, err
	}

	results := make([]marketdata.NewsArticle, 0, len(entries))
	for _, item := range entries {
		results = append(results, marketdata.NewsArticle{
			Headline:    item.Headline,
			Summary:     item.Summary,
			Source:      item.Source,
			PublishedAt: time.Unix(item.Datetime, 0).UTC(),
			URL:         item.URL,
		})
	}

	if p.cache != nil {
		_ = p.cache.SetJSON(ctx, cacheKey, results, 15*time.Minute)
	}

	return results, nil
}

func (p *Provider) GetFilings(ctx context.Context, symbol string) ([]marketdata.Filing, error) {
	cacheKey := "filings:" + strings.ToUpper(symbol)
	var cached []marketdata.Filing
	if p.cache != nil {
		if ok, err := p.cache.GetJSON(ctx, cacheKey, &cached); err == nil && ok {
			return cached, nil
		}
	}

	entries, err := p.client.GetFilings(ctx, symbol)
	if err != nil {
		return nil, err
	}

	allowed := map[string]bool{"10-K": true, "10-Q": true, "8-K": true}
	results := make([]marketdata.Filing, 0)
	for _, item := range entries {
		if !allowed[item.Form] {
			continue
		}

		accepted, _ := time.Parse(time.RFC3339, item.AcceptedDate)
		reportURL := item.ReportURL
		if reportURL == "" {
			reportURL = item.FilingURL
		}

		results = append(results, marketdata.Filing{
			Form:         item.Form,
			FiledDate:    item.FiledDate,
			AcceptedDate: accepted,
			ReportURL:    reportURL,
		})
	}

	if p.cache != nil {
		_ = p.cache.SetJSON(ctx, cacheKey, results, 24*time.Hour)
	}

	return results, nil
}

func (p *Provider) GetRecommendations(ctx context.Context, symbol string) ([]marketdata.Recommendation, error) {
	cacheKey := "recommendations:" + strings.ToUpper(symbol)
	var cached []marketdata.Recommendation
	if p.cache != nil {
		if ok, err := p.cache.GetJSON(ctx, cacheKey, &cached); err == nil && ok {
			return cached, nil
		}
	}

	entries, err := p.client.GetRecommendations(ctx, symbol)
	if err != nil {
		return nil, err
	}

	results := make([]marketdata.Recommendation, 0, len(entries))
	for _, item := range entries {
		results = append(results, marketdata.Recommendation{
			Buy:        item.Buy,
			Hold:       item.Hold,
			Sell:       item.Sell,
			StrongBuy:  item.StrongBuy,
			StrongSell: item.StrongSell,
			Period:     item.Period,
		})
	}

	if p.cache != nil {
		_ = p.cache.SetJSON(ctx, cacheKey, results, 24*time.Hour)
	}

	return results, nil
}

func mapQuote(symbol string, quote *finnhubclient.QuoteResponse) *marketdata.Quote {
	lastUpdated := time.Unix(quote.Timestamp, 0).UTC()
	if quote.Timestamp == 0 {
		lastUpdated = time.Now().UTC()
	}

	return &marketdata.Quote{
		Symbol:             strings.ToUpper(symbol),
		CurrentPrice:       quote.CurrentPrice,
		Open:               quote.Open,
		High:               quote.High,
		Low:                quote.Low,
		PreviousClose:      quote.PreviousClose,
		DailyChange:        quote.DailyChange,
		DailyChangePercent: quote.DailyChangePercent,
		LastUpdated:        lastUpdated,
	}
}

func mapProfile(symbol string, profile *finnhubclient.ProfileResponse) *marketdata.CompanyProfile {
	return &marketdata.CompanyProfile{
		Symbol:    strings.ToUpper(symbol),
		Name:      profile.Name,
		Exchange:  profile.Exchange,
		Industry:  profile.FinnhubIndustry,
		Country:   profile.Country,
		MarketCap: profile.MarketCap * 1_000_000,
		Logo:      profile.Logo,
		WebURL:    profile.WebURL,
		IPO:       profile.IPO,
	}
}

func extractExchange(displaySymbol, symbol string) string {
	if strings.Contains(displaySymbol, ":") {
		parts := strings.SplitN(displaySymbol, ":", 2)
		return parts[0]
	}
	return ""
}
