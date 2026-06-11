CREATE TABLE IF NOT EXISTS watchlist_items (
    id UUID PRIMARY KEY,
    user_id TEXT NOT NULL,
    symbol TEXT NOT NULL,
    company_name TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (user_id, symbol)
);

CREATE TABLE IF NOT EXISTS alerts (
    id UUID PRIMARY KEY,
    user_id TEXT NOT NULL,
    symbol TEXT,
    alert_type TEXT NOT NULL,
    params JSONB NOT NULL DEFAULT '{}',
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    last_triggered_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_watchlist_user ON watchlist_items (user_id);
CREATE INDEX IF NOT EXISTS idx_alerts_user ON alerts (user_id);
CREATE INDEX IF NOT EXISTS idx_alerts_active ON alerts (is_active);
