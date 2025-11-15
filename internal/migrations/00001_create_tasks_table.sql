-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS tasks (
    id TEXT PRIMARY KEY,
    payload TEXT NOT NULL,
    status INTEGER NOT NULL DEFAULT 0,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    scheduled_at DATETIME NOT NULL,
    attempt INTEGER NOT NULL DEFAULT 0,
    last_error TEXT,
    request_id TEXT,
    requirements_tags TEXT, 
    requirements_region TEXT,
    requirements_cpu REAL,
    requirements_memory REAL,
    requirements_requires_gpu INTEGER DEFAULT 0,
    priority INTEGER DEFAULT 0
);

CREATE INDEX IF NOT EXISTS idx_tasks_status ON tasks(status);
CREATE INDEX IF NOT EXISTS idx_tasks_scheduled_at ON tasks(scheduled_at);
CREATE INDEX IF NOT EXISTS idx_tasks_request_id ON tasks(request_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_tasks_request_id;
DROP INDEX IF EXISTS idx_tasks_scheduled_at;
DROP INDEX IF EXISTS idx_tasks_status;
DROP TABLE IF EXISTS tasks;
-- +goose StatementEnd

