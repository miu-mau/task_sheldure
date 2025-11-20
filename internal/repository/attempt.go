package repository

import (
	"database/sql"
	"task_shelduler/internal/models"
)

type AttemptRepository interface {
	CreateAttempt(attempt *models.Attempt) error
	GetAttempt(id string) (*models.Attempt, error)
	ListAttempts(limit int, offset int) ([]*models.Attempt, error)
	UpdateAttempt(id string, attempt *models.Attempt) error
	DeleteAttempt(id string) error
}

type attemptRepository struct {
	db *sql.DB
}

func NewAttemptRepository(db *sql.DB) AttemptRepository {
	return &attemptRepository{db: db}
}

func (r *attemptRepository) CreateAttempt(attempt *models.Attempt) error {
	return nil
}

func (r *attemptRepository) GetAttempt(id string) (*models.Attempt, error) {
	return nil, nil
}

func (r *attemptRepository) ListAttempts(limit int, offset int) ([]*models.Attempt, error) {
	return nil, nil
}

func (r *attemptRepository) UpdateAttempt(id string, attempt *models.Attempt) error {
	return nil
}

func (r *attemptRepository) DeleteAttempt(id string) error {
	return nil
}
