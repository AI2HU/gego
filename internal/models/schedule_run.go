package models

import "time"

type ScheduleRunTrigger string

const (
	ScheduleRunTriggerCron   ScheduleRunTrigger = "cron"
	ScheduleRunTriggerManual ScheduleRunTrigger = "manual"
)

type ScheduleRunStatus string

const (
	ScheduleRunStatusPending   ScheduleRunStatus = "pending"
	ScheduleRunStatusRunning   ScheduleRunStatus = "running"
	ScheduleRunStatusCompleted ScheduleRunStatus = "completed"
	ScheduleRunStatusFailed    ScheduleRunStatus = "failed"
	ScheduleRunStatusCancelled ScheduleRunStatus = "cancelled"
)

type ScheduleJobStatus string

const (
	ScheduleJobStatusPending   ScheduleJobStatus = "pending"
	ScheduleJobStatusClaimed   ScheduleJobStatus = "claimed"
	ScheduleJobStatusRunning   ScheduleJobStatus = "running"
	ScheduleJobStatusCompleted ScheduleJobStatus = "completed"
	ScheduleJobStatusFailed    ScheduleJobStatus = "failed"
	ScheduleJobStatusDead      ScheduleJobStatus = "dead"
)

type ScheduleRun struct {
	ID             string             `json:"id"`
	ScheduleID     string             `json:"schedule_id"`
	Trigger        ScheduleRunTrigger `json:"trigger"`
	Status         ScheduleRunStatus  `json:"status"`
	TotalJobs      int                `json:"total_jobs"`
	CompletedJobs  int                `json:"completed_jobs"`
	FailedJobs     int                `json:"failed_jobs"`
	CreatedAt      time.Time          `json:"created_at"`
	StartedAt      *time.Time         `json:"started_at,omitempty"`
	FinishedAt     *time.Time         `json:"finished_at,omitempty"`
	CronSlot       string             `json:"cron_slot,omitempty"`
}

type ScheduleJob struct {
	ID          string            `json:"id"`
	RunID       string            `json:"run_id"`
	ScheduleID  string            `json:"schedule_id"`
	PromptIDs   []string          `json:"prompt_ids"`
	LLMID       string            `json:"llm_id"`
	Provider    string            `json:"provider"`
	Temperature float64           `json:"temperature"`
	Status      ScheduleJobStatus `json:"status"`
	Attempts    int               `json:"attempts"`
	MaxAttempts int               `json:"max_attempts"`
	WorkerID    string            `json:"worker_id,omitempty"`
	ResponseIDs []string          `json:"response_ids,omitempty"`
	Error       string            `json:"error,omitempty"`
	CreatedAt   time.Time         `json:"created_at"`
	ClaimedAt   *time.Time        `json:"claimed_at,omitempty"`
	CompletedAt *time.Time        `json:"completed_at,omitempty"`
	PendingKey  string            `json:"pending_key,omitempty"`
}

type WorkerInfo struct {
	ID        string    `json:"id"`
	StartedAt time.Time `json:"started_at"`
	Hostname  string    `json:"hostname"`
}
