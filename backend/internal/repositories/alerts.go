package repositories

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Alert struct {
	ID              string         `json:"id"`
	UserID          string         `json:"userId"`
	Symbol          string         `json:"symbol,omitempty"`
	AlertType       string         `json:"alertType"`
	Params          map[string]any `json:"params"`
	IsActive        bool           `json:"isActive"`
	LastTriggeredAt *time.Time     `json:"lastTriggeredAt,omitempty"`
	CreatedAt       time.Time      `json:"createdAt"`
}

type AlertsRepository struct {
	pool *pgxpool.Pool
}

func NewAlertsRepository(pool *pgxpool.Pool) *AlertsRepository {
	return &AlertsRepository{pool: pool}
}

func (r *AlertsRepository) List(ctx context.Context, userID string) ([]Alert, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id::text, user_id, COALESCE(symbol, ''), alert_type, params, is_active, last_triggered_at, created_at
		FROM alerts
		WHERE user_id = $1
		ORDER BY created_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	alerts := make([]Alert, 0)
	for rows.Next() {
		var alert Alert
		var paramsRaw []byte
		if err := rows.Scan(
			&alert.ID, &alert.UserID, &alert.Symbol, &alert.AlertType, &paramsRaw,
			&alert.IsActive, &alert.LastTriggeredAt, &alert.CreatedAt,
		); err != nil {
			return nil, err
		}
		_ = json.Unmarshal(paramsRaw, &alert.Params)
		if alert.Params == nil {
			alert.Params = map[string]any{}
		}
		alerts = append(alerts, alert)
	}

	return alerts, rows.Err()
}

func (r *AlertsRepository) Create(ctx context.Context, userID, symbol, alertType string, params map[string]any) (*Alert, error) {
	if params == nil {
		params = map[string]any{}
	}

	paramsRaw, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}

	id := uuid.NewString()
	var alert Alert
	var paramsOut []byte
	err = r.pool.QueryRow(ctx, `
		INSERT INTO alerts (id, user_id, symbol, alert_type, params)
		VALUES ($1, $2, NULLIF($3, ''), $4, $5)
		RETURNING id::text, user_id, COALESCE(symbol, ''), alert_type, params, is_active, last_triggered_at, created_at
	`, id, userID, symbol, alertType, paramsRaw).Scan(
		&alert.ID, &alert.UserID, &alert.Symbol, &alert.AlertType, &paramsOut,
		&alert.IsActive, &alert.LastTriggeredAt, &alert.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	_ = json.Unmarshal(paramsOut, &alert.Params)

	return &alert, nil
}

func (r *AlertsRepository) Delete(ctx context.Context, userID, alertID string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM alerts WHERE user_id = $1 AND id = $2`, userID, alertID)
	return err
}

func (r *AlertsRepository) MarkTriggered(ctx context.Context, alertID string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE alerts SET last_triggered_at = NOW() WHERE id = $1
	`, alertID)
	return err
}

func (r *AlertsRepository) ListActive(ctx context.Context) ([]Alert, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id::text, user_id, COALESCE(symbol, ''), alert_type, params, is_active, last_triggered_at, created_at
		FROM alerts WHERE is_active = TRUE
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	alerts := make([]Alert, 0)
	for rows.Next() {
		var alert Alert
		var paramsRaw []byte
		if err := rows.Scan(
			&alert.ID, &alert.UserID, &alert.Symbol, &alert.AlertType, &paramsRaw,
			&alert.IsActive, &alert.LastTriggeredAt, &alert.CreatedAt,
		); err != nil {
			return nil, err
		}
		_ = json.Unmarshal(paramsRaw, &alert.Params)
		if alert.Params == nil {
			alert.Params = map[string]any{}
		}
		alerts = append(alerts, alert)
	}

	return alerts, rows.Err()
}
