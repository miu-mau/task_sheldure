package repository

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"task_shelduler/internal/models"
)

type TaskRepository interface {
	CreateTask(task *models.Task) error
	GetTask(id string) (*models.Task, error)
	ListTasks(status models.TaskStatus, limit int, offset int) ([]*models.Task, error)
	UpdateTaskStatus(id string, status models.TaskStatus) error
	DeleteTask(id string) error
}

type taskRepository struct {
	db *sql.DB
}

func NewTaskRepository(db *sql.DB) TaskRepository {
	return &taskRepository{db: db}
}

func (r *taskRepository) CreateTask(task *models.Task) error {
	jsonPayload, err := json.Marshal(task.Payload)
	if err != nil {
		return fmt.Errorf("failed to marshal task payload: %w", err)
	}

	_, err = r.db.Exec(`
		INSERT INTO tasks (id, payload, status, created_at, updated_at, scheduled_at, attempt, last_error, request_id, requirements, priority)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, task.ID, jsonPayload, task.Status, task.CreatedAt, task.UpdatedAt, task.ScheduledAt, task.Attempt, task.LastError, task.RequestID, task.Requirements, task.Priority)
	if err != nil {
		return fmt.Errorf("failed to create task: %w", err)
	}
	return nil
}

func (r *taskRepository) GetTask(id string) (*models.Task, error) {
	return nil, nil
}

func (r *taskRepository) ListTasks(status models.TaskStatus, limit int, offset int) ([]*models.Task, error) {
	return nil, nil
}

func (r *taskRepository) UpdateTaskStatus(id string, status models.TaskStatus) error {
	return nil
}

func (r *taskRepository) DeleteTask(id string) error {
	return nil
}
