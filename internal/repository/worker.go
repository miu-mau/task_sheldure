package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type WorkerRepository interface {
	RegisterWorker(ctx context.Context, baseName string) (string, error)
}

type workerRepository struct {
	db *sql.DB
}

func NewWorkerRepository(db *sql.DB) WorkerRepository {
	return &workerRepository{db: db}
}

func (r *workerRepository) RegisterWorker(ctx context.Context, baseName string) (string, error) {
	if baseName == "" {
		baseName = "worker"
	}

	res, err := r.db.ExecContext(ctx,
		`INSERT INTO workers (name, created_at) VALUES (?, ?)`,
		baseName,
		time.Now().UTC(),
	)
	if err != nil {
		return "", fmt.Errorf("failed to insert worker: %w", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return "", fmt.Errorf("failed to get last insert id for worker: %w", err)
	}

	workerID := fmt.Sprintf("%s%d", baseName, id)
	return workerID, nil
}
