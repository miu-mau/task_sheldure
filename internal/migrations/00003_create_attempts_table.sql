-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS attempts (
    id_attempts TEXT PRIMARY KEY,
    task_id TEXT NOT NULL,
    started_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    finished_at DATETIME,
    status_attempts INTEGER NOT NULL DEFAULT 0,
    error_attempts TEXT,
    FOREIGN KEY (task_id) REFERENCES tasks(id_tasks) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_attempts_task_id ON attempts(task_id);
CREATE INDEX IF NOT EXISTS idx_attempts_status_attempts ON attempts(status_attempts);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_attempts_status_attempts;
DROP INDEX IF EXISTS idx_attempts_task_id_attempts;
DROP TABLE IF EXISTS attempts;
-- +goose StatementEnd

