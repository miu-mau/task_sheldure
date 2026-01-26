package models

import "time"

type TaskStatus int32

const (
	TaskStatusUnspecified TaskStatus = 0
	TaskStatusDraft       TaskStatus = 1
	TaskStatusQueued      TaskStatus = 2
	TaskStatusRunning     TaskStatus = 3
	TaskStatusSuccess     TaskStatus = 4
	TaskStatusFailed      TaskStatus = 5
	TaskStatusCancelled   TaskStatus = 6
)

type AttemptStatus int32

const (
	AttemptStatusUnspecified AttemptStatus = 0
	AttemptStatusStarted     AttemptStatus = 1
	AttemptStatusSuccess     AttemptStatus = 2
	AttemptStatusFailed      AttemptStatus = 3
)

type TaskRequirements struct {
	Tags        []string `json:"tags"`
	Region      string   `json:"region"`
	CPU         float64  `json:"cpu"`
	Memory      float64  `json:"memory"`
	RequiresGPU bool     `json:"requires_gpu"`
}

type Task struct {
	ID           string            `json:"id"`
	Payload      string            `json:"payload"`
	Status       TaskStatus        `json:"status"`
	CreatedAt    time.Time         `json:"created_at"`
	UpdatedAt    time.Time         `json:"updated_at"`
	ScheduledAt  time.Time         `json:"scheduled_at"`
	Attempt      int               `json:"attempt"`
	MaxAttempts  int               `json:"max_attempts"`
	LastError    string            `json:"last_error"`
	RequestID    string            `json:"request_id"`
	Requirements *TaskRequirements `json:"requirements,omitempty"`
	Priority     int32             `json:"priority,omitempty"`
	WorkerID     string            `json:"worker_id,omitempty"`
}

type Schedule struct {
	ID        string    `json:"id"`
	Cron      string    `json:"cron"`
	Enabled   bool      `json:"enabled"`
	Payload   string    `json:"payload"`
	CreatedAt time.Time `json:"created_at"`
}

type Attempt struct {
	ID         string        `json:"id"`
	TaskID     string        `json:"task_id"`
	StartedAt  time.Time     `json:"started_at"`
	FinishedAt *time.Time    `json:"finished_at,omitempty"`
	Status     AttemptStatus `json:"status"`
	Error      string        `json:"error"`
}
