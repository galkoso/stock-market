CREATE TABLE IF NOT EXISTS notifications (
    id UUID PRIMARY KEY,
    user_id TEXT NOT NULL,
    alert_id UUID REFERENCES alerts(id) ON DELETE SET NULL,
    symbol TEXT,
    title TEXT NOT NULL,
    message TEXT NOT NULL,
    is_read BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_notifications_user ON notifications (user_id);
CREATE INDEX IF NOT EXISTS idx_notifications_user_unread ON notifications (user_id, is_read) WHERE is_read = FALSE;
