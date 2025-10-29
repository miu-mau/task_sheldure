package models

import "time"

type Task struct {
	ID          string    `json:"id"`
	Payload     string    `json:"payload"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	ScheduledAt time.Time `json:"scheduled_at"`
	Attempt     int       `json:"attempt"`
	LastError   string    `json:"last_error"`
	RequestID   string    `json:"request_id"`
}
