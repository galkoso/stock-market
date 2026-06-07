package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"stock-market/backend/internal/finnhub"
	"stock-market/backend/internal/model"
)

var (
	ErrMissingQuery   = errors.New("search query is required")
	ErrSymbolNotFound = errors.New("no matching stock symbol found")
	ErrInvalidQuote   = errors.New("quote data is unavailable for this symbol")
)

type FinnhubClient interface {
	SearchSymbols(ctx context.Context, query string) (*finnhub.SearchResponse, error)
	GetQuote(ctx context.Context, symbol string) (*finnhub.QuoteResponse, error)
}

type StockService struct {
	client FinnhubClient
}

func NewStockService(client FinnhubClient) *StockService {
	return &StockService{client: client}
}

func (s *StockService) SearchStock(ctx context.Context, query string) (*model.StockQuote, error) {
	trimmed := strings.TrimSpace(query)
	if trimmed == "" {
		return nil, ErrMissingQuery
	}

	search, err := s.client.SearchSymbols(ctx, trimmed)
	if err != nil {
		return nil, fmt.Errorf("search symbols: %w", err)
	}

	match := pickBestMatch(trimmed, search.Result)
	if match == nil {
		return nil, ErrSymbolNotFound
	}

	quote, err := s.client.GetQuote(ctx, match.Symbol)
	if err != nil {
		return nil, fmt.Errorf("get quote: %w", err)
	}

	if quote.Timestamp == 0 && quote.CurrentPrice == 0 {
		return nil, ErrInvalidQuote
	}

	lastUpdated := time.Unix(quote.Timestamp, 0).UTC()
	if quote.Timestamp == 0 {
		lastUpdated = time.Now().UTC()
	}

	return &model.StockQuote{
		Symbol:             match.Symbol,
		CompanyName:        match.Description,
		CurrentPrice:       quote.CurrentPrice,
		DailyChange:        quote.DailyChange,
		DailyChangePercent: quote.DailyChangePercent,
		LastUpdated:        lastUpdated,
	}, nil
}

func (s *StockService) ValidateSymbol(ctx context.Context, symbol string) error {
	trimmed := strings.ToUpper(strings.TrimSpace(symbol))
	if trimmed == "" {
		return ErrMissingQuery
	}

	quote, err := s.client.GetQuote(ctx, trimmed)
	if err != nil {
		return fmt.Errorf("get quote: %w", err)
	}

	if quote.Timestamp == 0 && quote.CurrentPrice == 0 {
		return ErrSymbolNotFound
	}

	return nil
}

func pickBestMatch(query string, results []finnhub.SearchResult) *finnhub.SearchResult {
	if len(results) == 0 {
		return nil
	}

	normalizedQuery := strings.ToUpper(strings.TrimSpace(query))

	for i := range results {
		result := &results[i]
		if strings.EqualFold(result.Symbol, normalizedQuery) ||
			strings.EqualFold(result.DisplaySymbol, normalizedQuery) {
			return result
		}
	}

	for i := range results {
		result := &results[i]
		if strings.Contains(strings.ToUpper(result.Description), normalizedQuery) {
			return result
		}
	}

	return &results[0]
}
