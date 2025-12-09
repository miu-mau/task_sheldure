package repository

import (
	"database/sql"
	"fmt"

	"task_shelduler/internal/models"
)

type AttemptRepository interface {
	CreateAttempt(attempt *models.Attempt) error
	GetAttempt(id string) (*models.Attempt, error)
	ListAttempts(limit int, offset int) ([]*models.Attempt, error)
	ListAttemptsByTaskID(taskID string) ([]*models.Attempt, error)
	UpdateAttempt(id string, attempt *models.Attempt) error
	DeleteAttempt(id string) error
	GetLastAttempt(taskID string) (*models.Attempt, error)
}

type attemptRepository struct {
	db *sql.DB
}

func NewAttemptRepository(db *sql.DB) AttemptRepository {
	return &attemptRepository{db: db}
}

func (r *attemptRepository) CreateAttempt(attempt *models.Attempt) error {
	var finishedAt interface{}
	if attempt.FinishedAt != nil {
		finishedAt = attempt.FinishedAt.UTC()
	}

	_, err := r.db.Exec(`
		INSERT INTO attempts (
			id_attempts,
			task_id,
			started_at,
			finished_at,
			status_attempts,
			error_attempts
		)
		VALUES (?, ?, ?, ?, ?, ?)
	`,
		attempt.ID,
		attempt.TaskID,
		attempt.StartedAt.UTC(),
		finishedAt,
		attempt.Status,
		attempt.Error,
	)
	if err != nil {
		return fmt.Errorf("failed to create attempt: %w", err)
	}
	return nil
}

func (r *attemptRepository) GetAttempt(id string) (*models.Attempt, error) {
	row := r.db.QueryRow(`
		SELECT
			id_attempts,
			task_id,
			started_at,
			finished_at,
			status_attempts,
			error_attempts
		FROM attempts
		WHERE id_attempts = ?
	`, id)

	attempt, err := scanAttempt(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, err
		}
		return nil, fmt.Errorf("failed to get attempt: %w", err)
	}
	return attempt, nil
}

func (r *attemptRepository) ListAttempts(limit int, offset int) ([]*models.Attempt, error) {
	rows, err := r.db.Query(`
		SELECT
			id_attempts,
			task_id,
			started_at,
			finished_at,
			status_attempts,
			error_attempts
		FROM attempts
		ORDER BY started_at DESC
		LIMIT ? OFFSET ?
	`, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list attempts: %w", err)
	}
	defer rows.Close()

	var attempts []*models.Attempt
	for rows.Next() {
		attempt, err := scanAttempt(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan attempt: %w", err)
		}
		attempts = append(attempts, attempt)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return attempts, nil
}

func (r *attemptRepository) ListAttemptsByTaskID(taskID string) ([]*models.Attempt, error) {
	rows, err := r.db.Query(`
		SELECT
			id_attempts,
			task_id,
			started_at,
			finished_at,
			status_attempts,
			error_attempts
		FROM attempts
		WHERE task_id = ?
		ORDER BY started_at DESC
	`, taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to list attempts by task_id: %w", err)
	}
	defer rows.Close()

	var attempts []*models.Attempt
	for rows.Next() {
		attempt, err := scanAttempt(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan attempt: %w", err)
		}
		attempts = append(attempts, attempt)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return attempts, nil
}

func (r *attemptRepository) UpdateAttempt(id string, attempt *models.Attempt) error {
	var finishedAt interface{}
	if attempt.FinishedAt != nil {
		finishedAt = attempt.FinishedAt.UTC()
	}

	result, err := r.db.Exec(`
		UPDATE attempts
		SET task_id = ?,
			started_at = ?,
			finished_at = ?,
			status_attempts = ?,
			error_attempts = ?
		WHERE id_attempts = ?
	`,
		attempt.TaskID,
		attempt.StartedAt.UTC(),
		finishedAt,
		attempt.Status,
		attempt.Error,
		id,
	)
	if err != nil {
		return fmt.Errorf("failed to update attempt: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check affected rows: %w", err)
	}
	if affected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *attemptRepository) DeleteAttempt(id string) error {
	result, err := r.db.Exec(`DELETE FROM attempts WHERE id_attempts = ?`, id)
	if err != nil {
		return fmt.Errorf("failed to delete attempt: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check affected rows: %w", err)
	}
	if affected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *attemptRepository) GetLastAttempt(taskID string) (*models.Attempt, error) {
	row := r.db.QueryRow(`
		SELECT
			id_attempts,
			task_id,
			started_at,
			finished_at,
			status_attempts,
			error_attempts
		FROM attempts
		WHERE task_id = ?
		ORDER BY started_at DESC
		LIMIT 1
	`, taskID)

	attempt, err := scanAttempt(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Не найдено - это нормально
		}
		return nil, fmt.Errorf("failed to get last attempt: %w", err)
	}
	return attempt, nil
}

func scanAttempt(scanner interface {
	Scan(dest ...interface{}) error
}) (*models.Attempt, error) {
	var attempt models.Attempt
	var statusInt int32
	var finishedAt sql.NullTime

	err := scanner.Scan(
		&attempt.ID,
		&attempt.TaskID,
		&attempt.StartedAt,
		&finishedAt,
		&statusInt,
		&attempt.Error,
	)
	if err != nil {
		return nil, err
	}

	attempt.Status = models.AttemptStatus(statusInt)
	if finishedAt.Valid {
		attempt.FinishedAt = &finishedAt.Time
	}

	return &attempt, nil
}
