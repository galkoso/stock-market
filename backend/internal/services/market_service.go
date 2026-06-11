package services

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"stock-market/backend/internal/provider/marketdata"
)

type MarketService struct {
	provider marketdata.Provider
}

func NewMarketService(provider marketdata.Provider) *MarketService {
	return &MarketService{provider: provider}
}

func (s *MarketService) Search(ctx context.Context, query string) ([]marketdata.SearchResult, error) {
	trimmed := strings.TrimSpace(query)
	if trimmed == "" {
		return nil, fmt.Errorf("search query is required")
	}
	return s.provider.Search(ctx, trimmed)
}

func (s *MarketService) GetDetails(ctx context.Context, symbol string) (*marketdata.StockDetails, error) {
	trimmed := strings.ToUpper(strings.TrimSpace(symbol))
	if trimmed == "" {
		return nil, fmt.Errorf("symbol is required")
	}
	return s.provider.GetStockDetails(ctx, trimmed)
}

func (s *MarketService) GetEarnings(ctx context.Context, from, to string) ([]marketdata.EarningsEvent, error) {
	if from == "" || to == "" {
		return nil, fmt.Errorf("from and to dates are required")
	}
	return s.provider.GetEarningsCalendar(ctx, from, to)
}

func (s *MarketService) GetEarningsSurprises(ctx context.Context, symbol string, limit int) ([]marketdata.EarningsSurprise, error) {
	trimmed := strings.ToUpper(strings.TrimSpace(symbol))
	if trimmed == "" {
		return nil, fmt.Errorf("symbol is required")
	}
	return s.provider.GetEarningsSurprises(ctx, trimmed, limit)
}

func (s *MarketService) GetEarningsHistory(ctx context.Context, symbols []string, from, to string, limit int) ([]marketdata.EarningsEvent, error) {
	if len(symbols) == 0 {
		return []marketdata.EarningsEvent{}, nil
	}
	if limit <= 0 {
		limit = 8
	}

	events := make([]marketdata.EarningsEvent, 0)
	seen := make(map[string]struct{})

	for _, symbol := range symbols {
		surprises, err := s.provider.GetEarningsSurprises(ctx, symbol, limit)
		if err != nil {
			continue
		}

		for _, item := range surprises {
			date := strings.TrimSpace(item.Period)
			if len(date) > 10 {
				date = date[:10]
			}
			if from != "" && date < from {
				continue
			}
			if to != "" && date > to {
				continue
			}

			key := fmt.Sprintf("%s:%d:%d", strings.ToUpper(item.Symbol), item.Quarter, item.Year)
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}

			actual := item.EPSActual
			estimate := item.EPSEstimate
			surprise := item.EPSSurprise
			surprisePct := item.EPSSurprisePercent

			events = append(events, marketdata.EarningsEvent{
				Symbol:             strings.ToUpper(item.Symbol),
				CompanyName:        strings.ToUpper(item.Symbol),
				Date:               date,
				EPSActual:          &actual,
				EPSEstimate:        &estimate,
				EPSSurprise:        &surprise,
				EPSSurprisePercent: &surprisePct,
				Quarter:            item.Quarter,
				Year:               item.Year,
			})
		}
	}

	sort.Slice(events, func(i, j int) bool {
		return events[i].Date > events[j].Date
	})

	return events, nil
}

func (s *MarketService) GetWatchlistEarnings(ctx context.Context, symbols []string, windowDays int) ([]marketdata.EarningsEvent, error) {
	if len(symbols) == 0 {
		return []marketdata.EarningsEvent{}, nil
	}

	now := time.Now().UTC()
	from := now.Format("2006-01-02")
	to := now.AddDate(0, 0, windowDays).Format("2006-01-02")

	events, err := s.provider.GetEarningsCalendar(ctx, from, to)
	if err != nil {
		return nil, err
	}

	symbolSet := make(map[string]struct{}, len(symbols))
	for _, symbol := range symbols {
		symbolSet[strings.ToUpper(symbol)] = struct{}{}
	}

	filtered := make([]marketdata.EarningsEvent, 0)
	for _, event := range events {
		if _, ok := symbolSet[strings.ToUpper(event.Symbol)]; ok {
			filtered = append(filtered, event)
		}
	}

	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].Date < filtered[j].Date
	})

	return filtered, nil
}

func (s *MarketService) GetNews(ctx context.Context, symbol string) ([]marketdata.NewsArticle, error) {
	trimmed := strings.ToUpper(strings.TrimSpace(symbol))
	if trimmed == "" {
		return nil, fmt.Errorf("symbol is required")
	}

	to := time.Now().UTC().Format("2006-01-02")
	from := time.Now().UTC().AddDate(0, 0, -30).Format("2006-01-02")
	return s.provider.GetCompanyNews(ctx, trimmed, from, to)
}

func (s *MarketService) GetFilings(ctx context.Context, symbol string) ([]marketdata.Filing, error) {
	trimmed := strings.ToUpper(strings.TrimSpace(symbol))
	if trimmed == "" {
		return nil, fmt.Errorf("symbol is required")
	}
	return s.provider.GetFilings(ctx, trimmed)
}

func (s *MarketService) GetRecommendations(ctx context.Context, symbol string) ([]marketdata.Recommendation, error) {
	trimmed := strings.ToUpper(strings.TrimSpace(symbol))
	if trimmed == "" {
		return nil, fmt.Errorf("symbol is required")
	}
	return s.provider.GetRecommendations(ctx, trimmed)
}

func (s *MarketService) GetMovers(ctx context.Context, symbols []string) ([]marketdata.Quote, error) {
	quotes := make([]marketdata.Quote, 0, len(symbols))
	for _, symbol := range symbols {
		quote, err := s.provider.GetQuote(ctx, symbol)
		if err != nil {
			continue
		}
		quotes = append(quotes, *quote)
	}

	sort.Slice(quotes, func(i, j int) bool {
		return quotes[i].DailyChangePercent > quotes[j].DailyChangePercent
	})

	return quotes, nil
}
