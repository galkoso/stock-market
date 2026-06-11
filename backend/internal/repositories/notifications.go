package repositories

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Notification struct {
	ID        string    `json:"id"`
	UserID    string    `json:"userId"`
	AlertID   string    `json:"alertId,omitempty"`
	Symbol    string    `json:"symbol,omitempty"`
	Title     string    `json:"title"`
	Message   string    `json:"message"`
	IsRead    bool      `json:"isRead"`
	CreatedAt time.Time `json:"createdAt"`
}

type NotificationsRepository struct {
	pool *pgxpool.Pool
}

func NewNotificationsRepository(pool *pgxpool.Pool) *NotificationsRepository {
	return &NotificationsRepository{pool: pool}
}

func (r *NotificationsRepository) Create(ctx context.Context, userID, alertID, symbol, title, message string) (*Notification, error) {
	id := uuid.NewString()
	var notification Notification
	err := r.pool.QueryRow(ctx, `
		INSERT INTO notifications (id, user_id, alert_id, symbol, title, message)
		VALUES ($1, $2, NULLIF($3, '')::uuid, NULLIF($4, ''), $5, $6)
		RETURNING id::text, user_id, COALESCE(alert_id::text, ''), COALESCE(symbol, ''), title, message, is_read, created_at
	`, id, userID, alertID, symbol, title, message).Scan(
		&notification.ID, &notification.UserID, &notification.AlertID, &notification.Symbol,
		&notification.Title, &notification.Message, &notification.IsRead, &notification.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &notification, nil
}

func (r *NotificationsRepository) List(ctx context.Context, userID string, limit int) ([]Notification, error) {
	if limit <= 0 {
		limit = 50
	}

	rows, err := r.pool.Query(ctx, `
		SELECT id::text, user_id, COALESCE(alert_id::text, ''), COALESCE(symbol, ''), title, message, is_read, created_at
		FROM notifications
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]Notification, 0)
	for rows.Next() {
		var item Notification
		if err := rows.Scan(
			&item.ID, &item.UserID, &item.AlertID, &item.Symbol,
			&item.Title, &item.Message, &item.IsRead, &item.CreatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *NotificationsRepository) CountUnread(ctx context.Context, userID string) (int, error) {
	var count int
	err := r.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM notifications WHERE user_id = $1 AND is_read = FALSE
	`, userID).Scan(&count)
	return count, err
}

func (r *NotificationsRepository) MarkRead(ctx context.Context, userID, notificationID string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE notifications SET is_read = TRUE WHERE user_id = $1 AND id = $2
	`, userID, notificationID)
	return err
}

func (r *NotificationsRepository) MarkAllRead(ctx context.Context, userID string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE notifications SET is_read = TRUE WHERE user_id = $1 AND is_read = FALSE
	`, userID)
	return err
}

func (r *NotificationsRepository) ExistsRecentForAlert(ctx context.Context, alertID string, since time.Time) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM notifications WHERE alert_id = $1::uuid AND created_at >= $2
		)
	`, alertID, since).Scan(&exists)
	return exists, err
}
