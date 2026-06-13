package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"strconv"
	"strings"
	"time"

	"stock-market/backend/internal/notifications"
	"stock-market/backend/internal/provider/marketdata"
	"stock-market/backend/internal/repositories"
)

const (
	unusualMoveThreshold = 5.0
	alertCooldown        = 24 * time.Hour
)

type AlertEngine struct {
	provider      marketdata.Provider
	alerts        *repositories.AlertsRepository
	notifications *repositories.NotificationsRepository
	hub           *notifications.Hub
	telegram      TelegramNotifier
}

type TelegramNotifier interface {
	Enabled() bool
	SendMessage(ctx context.Context, message string) error
	SendImportantNews(ctx context.Context, symbol, headline string) error
	SendAlert(ctx context.Context, title, body string) error
}

func NewAlertEngine(
	provider marketdata.Provider,
	alerts *repositories.AlertsRepository,
	notificationsRepo *repositories.NotificationsRepository,
	hub *notifications.Hub,
	telegramNotifier TelegramNotifier,
) *AlertEngine {
	return &AlertEngine{
		provider:      provider,
		alerts:        alerts,
		notifications: notificationsRepo,
		hub:           hub,
		telegram:      telegramNotifier,
	}
}

func (e *AlertEngine) Evaluate(ctx context.Context) {
	activeAlerts, err := e.alerts.ListActive(ctx)
	if err != nil {
		log.Printf("alert engine: list alerts failed: %v", err)
		return
	}

	for _, alert := range activeAlerts {
		if alert.Symbol == "" && alert.AlertType != "new_filing" {
			continue
		}

		if !shouldEvaluateOnSchedule(alert, time.Now().UTC()) {
			continue
		}

		if alert.LastTriggeredAt != nil && time.Since(*alert.LastTriggeredAt) < alertCooldown {
			continue
		}

		recentExists, err := e.notifications.ExistsRecentForAlert(ctx, alert.ID, time.Now().Add(-alertCooldown))
		if err == nil && recentExists {
			continue
		}

		triggered, title, message, err := e.evaluateAlert(ctx, alert)
		if err != nil {
			log.Printf("alert engine: evaluate %s failed: %v", alert.ID, err)
			continue
		}
		if !triggered {
			continue
		}

		created, err := e.notifications.Create(ctx, alert.UserID, alert.ID, alert.Symbol, title, message)
		if err != nil {
			log.Printf("alert engine: create notification failed: %v", err)
			continue
		}

		unread, err := e.notifications.CountUnread(ctx, alert.UserID)
		if err == nil && e.hub != nil {
			e.hub.PublishNotification(alert.UserID, *created, unread)
		}

		if err := e.alerts.MarkTriggered(ctx, alert.ID); err != nil {
			log.Printf("alert engine: mark triggered failed: %v", err)
		}

		e.notifyTelegram(alert, title, message)

		log.Printf("alert engine: triggered alert %s for user %s", alert.ID, alert.UserID)
	}
}

func (e *AlertEngine) notifyTelegram(alert repositories.Alert, title, message string) {
	if e.telegram == nil || !e.telegram.Enabled() {
		return
	}

	go func() {
		sendCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := e.telegram.SendAlert(sendCtx, title, message); err != nil {
			log.Printf("telegram: alert notification failed for %s: %v", alert.Symbol, err)
		}
	}()
}

func (e *AlertEngine) evaluateAlert(ctx context.Context, alert repositories.Alert) (bool, string, string, error) {
	switch alert.AlertType {
	case "earnings_days":
		return e.evaluateEarningsDays(ctx, alert)
	case "price_above":
		return e.evaluatePriceAbove(ctx, alert)
	case "price_below":
		return e.evaluatePriceBelow(ctx, alert)
	case "new_filing":
		return e.evaluateNewFiling(ctx, alert)
	case "unusual_move":
		return e.evaluateUnusualMove(ctx, alert)
	case "on_date":
		return e.evaluateOnDate(ctx, alert)
	default:
		return false, "", "", nil
	}
}

func shouldEvaluateOnSchedule(alert repositories.Alert, now time.Time) bool {
	today := truncateUTCDate(now)
	notifyDate, hasDate := alertNotifyDate(alert)

	if alert.AlertType == "on_date" {
		return hasDate && notifyDate.Equal(today)
	}

	if hasDate {
		return !today.Before(notifyDate)
	}

	return true
}

func alertNotifyDate(alert repositories.Alert) (time.Time, bool) {
	raw, ok := alert.Params["notifyDate"]
	if !ok || raw == nil {
		return time.Time{}, false
	}

	text := strings.TrimSpace(fmt.Sprint(raw))
	if text == "" {
		return time.Time{}, false
	}

	parsed, err := time.Parse("2006-01-02", text)
	if err != nil {
		return time.Time{}, false
	}

	return truncateUTCDate(parsed), true
}

func truncateUTCDate(value time.Time) time.Time {
	value = value.UTC()
	return time.Date(value.Year(), value.Month(), value.Day(), 0, 0, 0, 0, time.UTC)
}

func (e *AlertEngine) evaluateOnDate(ctx context.Context, alert repositories.Alert) (bool, string, string, error) {
	notifyDate, ok := alertNotifyDate(alert)
	if !ok {
		return false, "", "", nil
	}

	quote, err := e.provider.GetQuote(ctx, alert.Symbol)
	if err != nil {
		return false, "", "", err
	}

	dateLabel := notifyDate.Format("Jan 2, 2006")
	title := fmt.Sprintf("%s scheduled update", alert.Symbol)
	message := fmt.Sprintf(
		"Your scheduled update for %s on %s: $%.2f (%+.2f%% today).",
		alert.Symbol,
		dateLabel,
		quote.CurrentPrice,
		quote.DailyChangePercent,
	)
	return true, title, message, nil
}

func (e *AlertEngine) evaluateEarningsDays(ctx context.Context, alert repositories.Alert) (bool, string, string, error) {
	days := intParam(alert.Params["days"], 3)
	if days < 1 {
		days = 3
	}

	now := time.Now().UTC()
	from := now.Format("2006-01-02")
	to := now.AddDate(0, 0, days+1).Format("2006-01-02")

	events, err := e.provider.GetEarningsCalendar(ctx, from, to)
	if err != nil {
		return false, "", "", err
	}

	target := strings.ToUpper(alert.Symbol)
	for _, event := range events {
		if strings.ToUpper(event.Symbol) != target {
			continue
		}

		eventDate, err := time.Parse("2006-01-02", event.Date)
		if err != nil {
			continue
		}

		daysUntil := int(eventDate.Sub(now.Truncate(24*time.Hour)).Hours() / 24)
		if daysUntil == days {
			title := fmt.Sprintf("%s earnings in %d days", target, days)
			message := fmt.Sprintf("%s reports earnings on %s. Your reminder is set for %d day(s) before.", target, event.Date, days)
			return true, title, message, nil
		}
	}

	return false, "", "", nil
}

func (e *AlertEngine) evaluatePriceAbove(ctx context.Context, alert repositories.Alert) (bool, string, string, error) {
	target := floatParam(alert.Params["price"], 0)
	if target <= 0 {
		return false, "", "", nil
	}

	quote, err := e.provider.GetQuote(ctx, alert.Symbol)
	if err != nil {
		return false, "", "", err
	}

	if quote.CurrentPrice >= target {
		title := fmt.Sprintf("%s above $%.2f", alert.Symbol, target)
		message := fmt.Sprintf("%s is now $%.2f (target: $%.2f).", alert.Symbol, quote.CurrentPrice, target)
		return true, title, message, nil
	}

	return false, "", "", nil
}

func (e *AlertEngine) evaluatePriceBelow(ctx context.Context, alert repositories.Alert) (bool, string, string, error) {
	target := floatParam(alert.Params["price"], 0)
	if target <= 0 {
		return false, "", "", nil
	}

	quote, err := e.provider.GetQuote(ctx, alert.Symbol)
	if err != nil {
		return false, "", "", err
	}

	if quote.CurrentPrice <= target {
		title := fmt.Sprintf("%s below $%.2f", alert.Symbol, target)
		message := fmt.Sprintf("%s is now $%.2f (target: $%.2f).", alert.Symbol, quote.CurrentPrice, target)
		return true, title, message, nil
	}

	return false, "", "", nil
}

func (e *AlertEngine) evaluateNewFiling(ctx context.Context, alert repositories.Alert) (bool, string, string, error) {
	if alert.Symbol == "" {
		return false, "", "", nil
	}

	filings, err := e.provider.GetFilings(ctx, alert.Symbol)
	if err != nil {
		return false, "", "", err
	}
	if len(filings) == 0 {
		return false, "", "", nil
	}

	latest := filings[0]
	filedAt, err := time.Parse("2006-01-02", latest.FiledDate)
	if err != nil {
		return false, "", "", nil
	}

	if time.Since(filedAt) > 48*time.Hour {
		return false, "", "", nil
	}

	title := fmt.Sprintf("New SEC filing: %s", alert.Symbol)
	message := fmt.Sprintf("%s filed %s on %s.", alert.Symbol, latest.Form, latest.FiledDate)
	return true, title, message, nil
}

func (e *AlertEngine) evaluateUnusualMove(ctx context.Context, alert repositories.Alert) (bool, string, string, error) {
	quote, err := e.provider.GetQuote(ctx, alert.Symbol)
	if err != nil {
		return false, "", "", err
	}

	if math.Abs(quote.DailyChangePercent) < unusualMoveThreshold {
		return false, "", "", nil
	}

	direction := "up"
	if quote.DailyChangePercent < 0 {
		direction = "down"
	}

	title := fmt.Sprintf("%s unusual move", alert.Symbol)
	message := fmt.Sprintf("%s is %s %.2f%% today ($%.2f).", alert.Symbol, direction, math.Abs(quote.DailyChangePercent), quote.CurrentPrice)
	return true, title, message, nil
}

func intParam(value any, fallback int) int {
	switch v := value.(type) {
	case float64:
		return int(v)
	case int:
		return v
	case int64:
		return int(v)
	default:
		return fallback
	}
}

func floatParam(value any, fallback float64) float64 {
	switch v := value.(type) {
	case float64:
		return v
	case float32:
		return float64(v)
	case int:
		return float64(v)
	case int64:
		return float64(v)
	case json.Number:
		parsed, err := v.Float64()
		if err == nil {
			return parsed
		}
	case string:
		parsed, err := strconv.ParseFloat(v, 64)
		if err == nil {
			return parsed
		}
	}
	return fallback
}
