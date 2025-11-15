-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS schedules (
    id TEXT PRIMARY KEY,
    cron TEXT NOT NULL,
    enabled INTEGER NOT NULL DEFAULT 1,
    payload TEXT NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_schedules_enabled ON schedules(enabled);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_schedules_enabled;
DROP TABLE IF EXISTS schedules;
-- +goose StatementEnd

