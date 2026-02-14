-- +goose Up
-- +goose StatementBegin
ALTER TABLE tasks ADD COLUMN max_attempts INTEGER NOT NULL DEFAULT 3;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE tasks DROP COLUMN max_attempts;
-- +goose StatementEnd
