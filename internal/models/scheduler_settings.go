package models

import "time"

type SchedulerSettings struct {
	ID             string    `json:"id"`
	DesiredRunning bool      `json:"desired_running"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}
