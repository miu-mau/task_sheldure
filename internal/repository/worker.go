package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// WorkerRepository отвечает за регистрацию воркеров в БД.
type WorkerRepository interface {
	// RegisterWorker регистрирует воркер с базовым именем (например, "worker")
	// и возвращает сгенерированный workerID вида "worker1", "worker2" и т.д.
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

	// Вставляем запись в таблицу workers. AUTOINCREMENT id даст нам уникальный номер.
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

	workerID := fmt.Sprintf("%s%d", baseName, id) // worker1, worker2, ...
	return workerID, nil
}
