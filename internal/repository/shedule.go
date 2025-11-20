package repository

import (
	"database/sql"
	"task_shelduler/internal/models"
)

type ScheduleRepository interface {
	CreateSchedule(schedule *models.Schedule) error
	GetSchedule(id string) (*models.Schedule, error)
	ListSchedules(limit int, offset int) ([]*models.Schedule, error)
	UpdateSchedule(id string, schedule *models.Schedule) error
	DeleteSchedule(id string) error
}

type scheduleRepository struct {
	db *sql.DB
}

func NewScheduleRepository(db *sql.DB) ScheduleRepository {
	return &scheduleRepository{db: db}
}

func (r *scheduleRepository) CreateSchedule(schedule *models.Schedule) error {
	return nil
}

func (r *scheduleRepository) GetSchedule(id string) (*models.Schedule, error) {
	return nil, nil
}

func (r *scheduleRepository) ListSchedules(limit int, offset int) ([]*models.Schedule, error) {
	return nil, nil
}

func (r *scheduleRepository) UpdateSchedule(id string, schedule *models.Schedule) error {
	return nil
}

func (r *scheduleRepository) DeleteSchedule(id string) error {
	return nil
}
