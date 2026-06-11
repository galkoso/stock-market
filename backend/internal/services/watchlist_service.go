package services

import (
	"context"
	"fmt"
	"strings"

	"stock-market/backend/internal/repositories"
)

const MaxWatchlistSize = 50

type WatchlistService struct {
	repo *repositories.WatchlistRepository
}

func NewWatchlistService(repo *repositories.WatchlistRepository) *WatchlistService {
	return &WatchlistService{repo: repo}
}

func (s *WatchlistService) List(ctx context.Context, userID string) ([]repositories.WatchlistItem, error) {
	return s.repo.List(ctx, userID)
}

func (s *WatchlistService) Add(ctx context.Context, userID, symbol, companyName string) (*repositories.WatchlistItem, error) {
	normalized := strings.ToUpper(strings.TrimSpace(symbol))
	if normalized == "" {
		return nil, fmt.Errorf("symbol is required")
	}

	items, err := s.repo.List(ctx, userID)
	if err != nil {
		return nil, err
	}

	for _, item := range items {
		if item.Symbol == normalized {
			return &item, nil
		}
	}

	if len(items) >= MaxWatchlistSize {
		return nil, fmt.Errorf("watchlist is full (%d symbols max)", MaxWatchlistSize)
	}

	return s.repo.Add(ctx, userID, normalized, companyName)
}

func (s *WatchlistService) Remove(ctx context.Context, userID, symbol string) error {
	normalized := strings.ToUpper(strings.TrimSpace(symbol))
	if normalized == "" {
		return fmt.Errorf("symbol is required")
	}
	return s.repo.Remove(ctx, userID, normalized)
}

func (s *WatchlistService) Symbols(ctx context.Context, userID string) ([]string, error) {
	return s.repo.Symbols(ctx, userID)
}
