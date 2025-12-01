package repository

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"task_shelduler/internal/models"
)

type TaskRepository interface {
	CreateTask(task *models.Task) error
	GetTask(id string) (*models.Task, error)
	ListTasks(status models.TaskStatus, limit int, offset int) ([]*models.Task, error)
	UpdateTaskStatus(id string, status models.TaskStatus) error
	UpdateTaskStatusWithError(id string, status models.TaskStatus, errorMsg string) error
	DeleteTask(id string) error
	GetReadyTasks(limit int) ([]*models.Task, error)
	FindByRequestID(requestID string) (*models.Task, error)
}

type taskRepository struct {
	db *sql.DB
}

func NewTaskRepository(db *sql.DB) TaskRepository {
	return &taskRepository{db: db}
}

func (r *taskRepository) CreateTask(task *models.Task) error {
	reqTags, reqRegion, reqCPU, reqMemory, reqGPU := encodeRequirements(task.Requirements)

	_, err := r.db.Exec(`
		INSERT INTO tasks (
			id_tasks,
			payload,
			status_tasks,
			created_at,
			updated_at,
			scheduled_at,
			attempt,
			last_error,
			request_id,
			requirements_tags,
			requirements_region,
			requirements_cpu,
			requirements_memory,
			requirements_requires_gpu,
			priority_tasks
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		task.ID,
		task.Payload,
		task.Status,
		task.CreatedAt,
		task.UpdatedAt,
		task.ScheduledAt,
		task.Attempt,
		task.LastError,
		task.RequestID,
		reqTags,
		reqRegion,
		reqCPU,
		reqMemory,
		reqGPU,
		task.Priority,
	)
	if err != nil {
		return fmt.Errorf("failed to create task: %w", err)
	}
	return nil
}

func (r *taskRepository) GetTask(id string) (*models.Task, error) {
	row := r.db.QueryRow(`
		SELECT
			id_tasks,
			payload,
			status_tasks,
			created_at,
			updated_at,
			scheduled_at,
			attempt,
			last_error,
			request_id,
			requirements_tags,
			requirements_region,
			requirements_cpu,
			requirements_memory,
			requirements_requires_gpu,
			priority_tasks
		FROM tasks
		WHERE id_tasks = ?
	`, id)

	task, err := scanTask(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, err
		}
		return nil, fmt.Errorf("failed to get task: %w", err)
	}
	return task, nil
}

func (r *taskRepository) ListTasks(status models.TaskStatus, limit int, offset int) ([]*models.Task, error) {
	query := `
		SELECT
			id_tasks,
			payload,
			status_tasks,
			created_at,
			updated_at,
			scheduled_at,
			attempt,
			last_error,
			request_id,
			requirements_tags,
			requirements_region,
			requirements_cpu,
			requirements_memory,
			requirements_requires_gpu,
			priority_tasks
		FROM tasks
	`

	var args []interface{}
	if status != models.TaskStatusUnspecified {
		query += " WHERE status_tasks = ?"
		args = append(args, status)
	}

	query += " ORDER BY created_at DESC LIMIT ? OFFSET ?"
	args = append(args, limit, offset)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list tasks: %w", err)
	}
	defer rows.Close()

	var tasks []*models.Task
	for rows.Next() {
		task, err := scanTask(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan task: %w", err)
		}
		tasks = append(tasks, task)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return tasks, nil
}

func (r *taskRepository) UpdateTaskStatus(id string, status models.TaskStatus) error {
	result, err := r.db.Exec(`
		UPDATE tasks
		SET status_tasks = ?, updated_at = ?
		WHERE id_tasks = ?
	`, status, time.Now().UTC(), id)
	if err != nil {
		return fmt.Errorf("failed to update task status: %w", err)
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

func (r *taskRepository) DeleteTask(id string) error {
	result, err := r.db.Exec(`DELETE FROM tasks WHERE id_tasks = ?`, id)
	if err != nil {
		return fmt.Errorf("failed to delete task: %w", err)
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

func (r *taskRepository) UpdateTaskStatusWithError(id string, status models.TaskStatus, errorMsg string) error {
	result, err := r.db.Exec(`
		UPDATE tasks
		SET status_tasks = ?, updated_at = ?, last_error = ?
		WHERE id_tasks = ?
	`, status, time.Now().UTC(), errorMsg, id)
	if err != nil {
		return fmt.Errorf("failed to update task status with error: %w", err)
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

// GetReadyTasks возвращает задачи, готовые к выполнению (DRAFT статус и scheduled_at <= now)
// с блокировкой для предотвращения дублирования
func (r *taskRepository) GetReadyTasks(limit int) ([]*models.Task, error) {
	query := `
		SELECT
			id_tasks,
			payload,
			status_tasks,
			created_at,
			updated_at,
			scheduled_at,
			attempt,
			last_error,
			request_id,
			requirements_tags,
			requirements_region,
			requirements_cpu,
			requirements_memory,
			requirements_requires_gpu,
			priority_tasks
		FROM tasks
		WHERE status_tasks = ? AND scheduled_at <= ?
		ORDER BY priority_tasks DESC, scheduled_at ASC
		LIMIT ?
	`

	rows, err := r.db.Query(query, models.TaskStatusDraft, time.Now().UTC(), limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get ready tasks: %w", err)
	}
	defer rows.Close()

	var tasks []*models.Task
	for rows.Next() {
		task, err := scanTask(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan task: %w", err)
		}
		tasks = append(tasks, task)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return tasks, nil
}

// FindByRequestID находит задачу по request_id для идемпотентности
func (r *taskRepository) FindByRequestID(requestID string) (*models.Task, error) {
	row := r.db.QueryRow(`
		SELECT
			id_tasks,
			payload,
			status_tasks,
			created_at,
			updated_at,
			scheduled_at,
			attempt,
			last_error,
			request_id,
			requirements_tags,
			requirements_region,
			requirements_cpu,
			requirements_memory,
			requirements_requires_gpu,
			priority_tasks
		FROM tasks
		WHERE request_id = ?
		ORDER BY created_at DESC
		LIMIT 1
	`, requestID)

	task, err := scanTask(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Не найдено - это нормально
		}
		return nil, fmt.Errorf("failed to find task by request_id: %w", err)
	}
	return task, nil
}

func scanTask(scanner interface {
	Scan(dest ...interface{}) error
}) (*models.Task, error) {
	var (
		task models.Task

		statusInt               int32
		tagsJSON                sql.NullString
		requirementsRegion      sql.NullString
		requirementsCPU         sql.NullFloat64
		requirementsMemory      sql.NullFloat64
		requirementsRequiresGPU sql.NullInt64
	)

	err := scanner.Scan(
		&task.ID,
		&task.Payload,
		&statusInt,
		&task.CreatedAt,
		&task.UpdatedAt,
		&task.ScheduledAt,
		&task.Attempt,
		&task.LastError,
		&task.RequestID,
		&tagsJSON,
		&requirementsRegion,
		&requirementsCPU,
		&requirementsMemory,
		&requirementsRequiresGPU,
		&task.Priority,
	)
	if err != nil {
		return nil, err
	}

	task.Status = models.TaskStatus(statusInt)
	task.Requirements = decodeRequirements(tagsJSON, requirementsRegion, requirementsCPU, requirementsMemory, requirementsRequiresGPU)

	return &task, nil
}

func encodeRequirements(req *models.TaskRequirements) (tags sql.NullString, region sql.NullString, cpu sql.NullFloat64, memory sql.NullFloat64, requiresGPU sql.NullInt64) {
	if req == nil {
		return
	}

	if len(req.Tags) > 0 {
		if data, err := json.Marshal(req.Tags); err == nil {
			tags = sql.NullString{String: string(data), Valid: true}
		}
	}
	if req.Region != "" {
		region = sql.NullString{String: req.Region, Valid: true}
	}
	if req.CPU != 0 {
		cpu = sql.NullFloat64{Float64: req.CPU, Valid: true}
	}
	if req.Memory != 0 {
		memory = sql.NullFloat64{Float64: req.Memory, Valid: true}
	}
	if req.RequiresGPU {
		requiresGPU = sql.NullInt64{Int64: 1, Valid: true}
	} else {
		requiresGPU = sql.NullInt64{Int64: 0, Valid: true}
	}

	return
}

func decodeRequirements(tags sql.NullString, region sql.NullString, cpu sql.NullFloat64, memory sql.NullFloat64, requiresGPU sql.NullInt64) *models.TaskRequirements {
	if !tags.Valid && !region.Valid && !cpu.Valid && !memory.Valid && !requiresGPU.Valid {
		return nil
	}

	req := &models.TaskRequirements{}

	if tags.Valid {
		var list []string
		if err := json.Unmarshal([]byte(tags.String), &list); err == nil {
			req.Tags = list
		}
	}
	if region.Valid {
		req.Region = region.String
	}
	if cpu.Valid {
		req.CPU = cpu.Float64
	}
	if memory.Valid {
		req.Memory = memory.Float64
	}
	if requiresGPU.Valid {
		req.RequiresGPU = requiresGPU.Int64 == 1
	}

	return req
}
