package services

import (
	"context"
	"fmt"
	"strings"

	"stock-market/backend/internal/repositories"
)

var allowedAlertTypes = map[string]bool{
	"earnings_days":  true,
	"price_above":    true,
	"price_below":    true,
	"new_filing":     true,
	"unusual_move":   true,
}

type AlertsService struct {
	repo *repositories.AlertsRepository
}

func NewAlertsService(repo *repositories.AlertsRepository) *AlertsService {
	return &AlertsService{repo: repo}
}

func (s *AlertsService) List(ctx context.Context, userID string) ([]repositories.Alert, error) {
	return s.repo.List(ctx, userID)
}

func (s *AlertsService) Create(ctx context.Context, userID, symbol, alertType string, params map[string]any) (*repositories.Alert, error) {
	normalizedType := strings.TrimSpace(alertType)
	if !allowedAlertTypes[normalizedType] {
		return nil, fmt.Errorf("unsupported alert type")
	}

	normalizedSymbol := strings.ToUpper(strings.TrimSpace(symbol))
	return s.repo.Create(ctx, userID, normalizedSymbol, normalizedType, params)
}

func (s *AlertsService) Delete(ctx context.Context, userID, alertID string) error {
	if strings.TrimSpace(alertID) == "" {
		return fmt.Errorf("alert id is required")
	}
	return s.repo.Delete(ctx, userID, alertID)
}
