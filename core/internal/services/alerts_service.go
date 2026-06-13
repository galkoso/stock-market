package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"stock-market/backend/internal/repositories"
)

var allowedAlertTypes = map[string]bool{
	"earnings_days": true,
	"price_above":   true,
	"price_below":   true,
	"new_filing":    true,
	"unusual_move":  true,
	"on_date":       true,
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

	if params == nil {
		params = map[string]any{}
	}

	if err := validateAlertParams(normalizedType, symbol, params); err != nil {
		return nil, err
	}

	normalizedSymbol := strings.ToUpper(strings.TrimSpace(symbol))
	return s.repo.Create(ctx, userID, normalizedSymbol, normalizedType, params)
}

func validateAlertParams(alertType, symbol string, params map[string]any) error {
	if alertType == "on_date" {
		if strings.TrimSpace(symbol) == "" {
			return fmt.Errorf("symbol is required for scheduled updates")
		}
		if _, err := parseNotifyDate(params["notifyDate"]); err != nil {
			return fmt.Errorf("notifyDate is required and must be YYYY-MM-DD")
		}
		return nil
	}

	if raw, ok := params["notifyDate"]; ok && raw != nil && strings.TrimSpace(fmt.Sprint(raw)) != "" {
		if _, err := parseNotifyDate(raw); err != nil {
			return fmt.Errorf("notifyDate must be YYYY-MM-DD")
		}
	} else {
		delete(params, "notifyDate")
	}

	return nil
}

func parseNotifyDate(value any) (time.Time, error) {
	raw := strings.TrimSpace(fmt.Sprint(value))
	if raw == "" {
		return time.Time{}, fmt.Errorf("empty date")
	}
	parsed, err := time.Parse("2006-01-02", raw)
	if err != nil {
		return time.Time{}, err
	}
	return parsed.UTC(), nil
}

func (s *AlertsService) Delete(ctx context.Context, userID, alertID string) error {
	if strings.TrimSpace(alertID) == "" {
		return fmt.Errorf("alert id is required")
	}
	return s.repo.Delete(ctx, userID, alertID)
}
