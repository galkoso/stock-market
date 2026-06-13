package services

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strings"

	"stock-market/backend/internal/forex"
	"stock-market/backend/internal/provider/marketdata"
	"stock-market/backend/internal/repositories"

	"go.mongodb.org/mongo-driver/mongo"
)

type AllocationHolding struct {
	Symbol             string  `json:"symbol"`
	Quantity           float64 `json:"quantity"`
	Price              float64 `json:"price"`
	PriceIls           float64 `json:"priceIls"`
	PreviousClose      float64 `json:"previousClose"`
	DailyChange        float64 `json:"dailyChange"`
	DailyChangePercent float64 `json:"dailyChangePercent"`
	MarketValue        float64 `json:"marketValue"`
	MarketValueIls     float64 `json:"marketValueIls"`
	AllocationPercent  float64 `json:"allocationPercent"`
}

type PortfolioAllocation struct {
	TotalValue    float64             `json:"totalValue"`
	TotalValueIls float64             `json:"totalValueIls"`
	UsdToIls      float64             `json:"usdToIls"`
	Holdings      []AllocationHolding `json:"holdings"`
}

type PortfolioService struct {
	repo     *repositories.PortfolioRepository
	provider marketdata.Provider
	forex    *forex.Client
}

func NewPortfolioService(repo *repositories.PortfolioRepository, provider marketdata.Provider, forexClient *forex.Client) *PortfolioService {
	return &PortfolioService{
		repo:     repo,
		provider: provider,
		forex:    forexClient,
	}
}

func (s *PortfolioService) List(ctx context.Context, userID string) ([]repositories.PortfolioHolding, error) {
	return s.repo.List(ctx, userID)
}

func (s *PortfolioService) Add(ctx context.Context, userID, symbol string, quantity float64) (*repositories.PortfolioHolding, error) {
	normalized := strings.ToUpper(strings.TrimSpace(symbol))
	if normalized == "" {
		return nil, fmt.Errorf("symbol is required")
	}
	if quantity <= 0 {
		return nil, fmt.Errorf("quantity must be greater than zero")
	}

	if _, err := s.provider.GetQuote(ctx, normalized); err != nil {
		return nil, fmt.Errorf("unable to verify symbol %s", normalized)
	}

	if _, err := s.repo.FindBySymbol(ctx, userID, normalized); err == nil {
		return nil, fmt.Errorf("holding for %s already exists", normalized)
	} else if err != mongo.ErrNoDocuments {
		return nil, err
	}

	return s.repo.Create(ctx, userID, normalized, quantity)
}

func (s *PortfolioService) UpdateQuantity(ctx context.Context, userID, symbol string, quantity float64) (*repositories.PortfolioHolding, error) {
	normalized := strings.ToUpper(strings.TrimSpace(symbol))
	if normalized == "" {
		return nil, fmt.Errorf("symbol is required")
	}
	if quantity <= 0 {
		return nil, fmt.Errorf("quantity must be greater than zero")
	}

	holding, err := s.repo.UpdateQuantity(ctx, userID, normalized, quantity)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("holding for %s not found", normalized)
		}
		return nil, err
	}

	return holding, nil
}

func (s *PortfolioService) Remove(ctx context.Context, userID, symbol string) error {
	normalized := strings.ToUpper(strings.TrimSpace(symbol))
	if normalized == "" {
		return fmt.Errorf("symbol is required")
	}

	result, err := s.repo.Remove(ctx, userID, normalized)
	if err != nil {
		return err
	}
	if result == 0 {
		return fmt.Errorf("holding for %s not found", normalized)
	}

	return nil
}

func (s *PortfolioService) GetAllocation(ctx context.Context, userID string) (*PortfolioAllocation, error) {
	holdings, err := s.repo.List(ctx, userID)
	if err != nil {
		return nil, err
	}

	usdToIls, err := s.forex.GetUSDToILS(ctx)
	if err != nil {
		return nil, fmt.Errorf("USD/ILS exchange rate: %w", err)
	}

	if len(holdings) == 0 {
		return &PortfolioAllocation{
			TotalValue:    0,
			TotalValueIls: 0,
			UsdToIls:      roundMoney(usdToIls),
			Holdings:      []AllocationHolding{},
		}, nil
	}

	allocations := make([]AllocationHolding, 0, len(holdings))
	var totalValue float64

	for _, holding := range holdings {
		quote, err := s.provider.GetQuote(ctx, holding.Symbol)
		if err != nil {
			continue
		}

		marketValue := holding.Quantity * quote.CurrentPrice
		totalValue += marketValue

		allocations = append(allocations, AllocationHolding{
			Symbol:             holding.Symbol,
			Quantity:           holding.Quantity,
			Price:              quote.CurrentPrice,
			PriceIls:           roundMoney(quote.CurrentPrice * usdToIls),
			PreviousClose:      quote.PreviousClose,
			DailyChange:        quote.DailyChange,
			DailyChangePercent: quote.DailyChangePercent,
			MarketValue:        marketValue,
			MarketValueIls:     roundMoney(marketValue * usdToIls),
		})
	}

	for i := range allocations {
		if totalValue > 0 {
			allocations[i].AllocationPercent = roundPercent((allocations[i].MarketValue / totalValue) * 100)
		}
	}

	sort.Slice(allocations, func(i, j int) bool {
		return allocations[i].MarketValue > allocations[j].MarketValue
	})

	totalValue = roundMoney(totalValue)

	return &PortfolioAllocation{
		TotalValue:    totalValue,
		TotalValueIls: roundMoney(totalValue * usdToIls),
		UsdToIls:      roundMoney(usdToIls),
		Holdings:      allocations,
	}, nil
}

func roundPercent(value float64) float64 {
	return math.Round(value*100) / 100
}

func roundMoney(value float64) float64 {
	return math.Round(value*100) / 100
}
