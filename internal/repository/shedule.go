package repository

import (
	"database/sql"
	"fmt"
	"task_shelduler/internal/models"
)

type ScheduleRepository interface {
	CreateSchedule(schedule *models.Schedule) error
	GetSchedule(id string) (*models.Schedule, error)
	ListSchedules(enabled *bool, limit int, offset int) ([]*models.Schedule, error)
	UpdateSchedule(id string, schedule *models.Schedule) error
	DeleteSchedule(id string) error
	GetActiveSchedules() ([]*models.Schedule, error)
}

type scheduleRepository struct {
	db *sql.DB
}

func NewScheduleRepository(db *sql.DB) ScheduleRepository {
	return &scheduleRepository{db: db}
}

func (r *scheduleRepository) CreateSchedule(schedule *models.Schedule) error {
	enabled := 0
	if schedule.Enabled {
		enabled = 1
	}

	_, err := r.db.Exec(`
		INSERT INTO schedules (
			id_schedules,
			cron,
			enabled_schedules,
			payload,
			created_at
		)
		VALUES (?, ?, ?, ?, ?)
	`,
		schedule.ID,
		schedule.Cron,
		enabled,
		schedule.Payload,
		schedule.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create schedule: %w", err)
	}
	return nil
}

func (r *scheduleRepository) GetSchedule(id string) (*models.Schedule, error) {
	row := r.db.QueryRow(`
		SELECT
			id_schedules,
			cron,
			enabled_schedules,
			payload,
			created_at
		FROM schedules
		WHERE id_schedules = ?
	`, id)

	var schedule models.Schedule
	var enabledInt int

	err := row.Scan(
		&schedule.ID,
		&schedule.Cron,
		&enabledInt,
		&schedule.Payload,
		&schedule.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, err
		}
		return nil, fmt.Errorf("failed to get schedule: %w", err)
	}

	schedule.Enabled = enabledInt == 1
	return &schedule, nil
}

func (r *scheduleRepository) ListSchedules(enabled *bool, limit int, offset int) ([]*models.Schedule, error) {
	query := `
		SELECT
			id_schedules,
			cron,
			enabled_schedules,
			payload,
			created_at
		FROM schedules
	`

	var args []interface{}
	if enabled != nil {
		enabledVal := 0
		if *enabled {
			enabledVal = 1
		}
		query += " WHERE enabled_schedules = ?"
		args = append(args, enabledVal)
	}

	query += " ORDER BY created_at DESC LIMIT ? OFFSET ?"
	args = append(args, limit, offset)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list schedules: %w", err)
	}
	defer rows.Close()

	var schedules []*models.Schedule
	for rows.Next() {
		var schedule models.Schedule
		var enabledInt int

		err := rows.Scan(
			&schedule.ID,
			&schedule.Cron,
			&enabledInt,
			&schedule.Payload,
			&schedule.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan schedule: %w", err)
		}

		schedule.Enabled = enabledInt == 1
		schedules = append(schedules, &schedule)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return schedules, nil
}

func (r *scheduleRepository) UpdateSchedule(id string, schedule *models.Schedule) error {
	enabled := 0
	if schedule.Enabled {
		enabled = 1
	}

	result, err := r.db.Exec(`
		UPDATE schedules
		SET cron = ?, enabled_schedules = ?, payload = ?
		WHERE id_schedules = ?
	`, schedule.Cron, enabled, schedule.Payload, id)
	if err != nil {
		return fmt.Errorf("failed to update schedule: %w", err)
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

func (r *scheduleRepository) DeleteSchedule(id string) error {
	result, err := r.db.Exec(`DELETE FROM schedules WHERE id_schedules = ?`, id)
	if err != nil {
		return fmt.Errorf("failed to delete schedule: %w", err)
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

func (r *scheduleRepository) GetActiveSchedules() ([]*models.Schedule, error) {
	rows, err := r.db.Query(`
		SELECT
			id_schedules,
			cron,
			enabled_schedules,
			payload,
			created_at
		FROM schedules
		WHERE enabled_schedules = 1
		ORDER BY created_at ASC
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to get active schedules: %w", err)
	}
	defer rows.Close()

	var schedules []*models.Schedule
	for rows.Next() {
		var schedule models.Schedule
		var enabledInt int

		err := rows.Scan(
			&schedule.ID,
			&schedule.Cron,
			&enabledInt,
			&schedule.Payload,
			&schedule.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan schedule: %w", err)
		}

		schedule.Enabled = enabledInt == 1
		schedules = append(schedules, &schedule)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return schedules, nil
}
