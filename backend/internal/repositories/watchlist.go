package repositories

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type WatchlistItem struct {
	ID          string    `json:"id"`
	UserID      string    `json:"userId"`
	Symbol      string    `json:"symbol"`
	CompanyName string    `json:"companyName"`
	CreatedAt   time.Time `json:"createdAt"`
}

type WatchlistRepository struct {
	pool *pgxpool.Pool
}

func NewWatchlistRepository(pool *pgxpool.Pool) *WatchlistRepository {
	return &WatchlistRepository{pool: pool}
}

func (r *WatchlistRepository) List(ctx context.Context, userID string) ([]WatchlistItem, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id::text, user_id, symbol, company_name, created_at
		FROM watchlist_items
		WHERE user_id = $1
		ORDER BY created_at ASC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]WatchlistItem, 0)
	for rows.Next() {
		var item WatchlistItem
		if err := rows.Scan(&item.ID, &item.UserID, &item.Symbol, &item.CompanyName, &item.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return items, rows.Err()
}

func (r *WatchlistRepository) Add(ctx context.Context, userID, symbol, companyName string) (*WatchlistItem, error) {
	id := uuid.NewString()
	var item WatchlistItem
	err := r.pool.QueryRow(ctx, `
		INSERT INTO watchlist_items (id, user_id, symbol, company_name)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (user_id, symbol) DO UPDATE
		SET company_name = EXCLUDED.company_name
		RETURNING id::text, user_id, symbol, company_name, created_at
	`, id, userID, symbol, companyName).Scan(
		&item.ID, &item.UserID, &item.Symbol, &item.CompanyName, &item.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &item, nil
}

func (r *WatchlistRepository) Remove(ctx context.Context, userID, symbol string) error {
	_, err := r.pool.Exec(ctx, `
		DELETE FROM watchlist_items WHERE user_id = $1 AND symbol = $2
	`, userID, symbol)
	return err
}

func (r *WatchlistRepository) Symbols(ctx context.Context, userID string) ([]string, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT symbol FROM watchlist_items WHERE user_id = $1 ORDER BY created_at ASC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	symbols := make([]string, 0)
	for rows.Next() {
		var symbol string
		if err := rows.Scan(&symbol); err != nil {
			return nil, err
		}
		symbols = append(symbols, symbol)
	}

	return symbols, rows.Err()
}
