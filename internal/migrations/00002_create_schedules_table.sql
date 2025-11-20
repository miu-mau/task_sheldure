-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS schedules (
    id_schedules TEXT PRIMARY KEY,
    cron TEXT NOT NULL,
    enabled_schedules INTEGER NOT NULL DEFAULT 1,
    payload TEXT NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_schedules_enabled ON schedules(enabled_schedules);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_schedules_enabled_schedules;
DROP TABLE IF EXISTS schedules;
-- +goose StatementEnd

