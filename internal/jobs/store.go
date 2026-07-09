package jobs

import (
	"context"
	"time"

	"github.com/AI2HU/gego/internal/models"
)

type RunFilter struct {
	ScheduleID string
	Status     models.ScheduleRunStatus
	Limit      int
	Cursor     string
}

type Store interface {
	CreateRun(ctx context.Context, run *models.ScheduleRun, jobs []*models.ScheduleJob, dedupKey string, dedupTTL time.Duration) error
	WatchPendingJobs(ctx context.Context) (<-chan *models.ScheduleJob, error)
	ClaimJob(ctx context.Context, job *models.ScheduleJob, workerID string, lease time.Duration) (*models.ScheduleJob, error)
	CompleteJob(ctx context.Context, jobID string, responseIDs []string) error
	FailJob(ctx context.Context, jobID string, jobErr error) error
	RetryJob(ctx context.Context, jobID string) error
	CancelRun(ctx context.Context, runID string) error
	ReclaimStaleJobs(ctx context.Context) (int, error)
	UpdateRunProgress(ctx context.Context, runID string) (*models.ScheduleRun, error)
	GetRun(ctx context.Context, runID string) (*models.ScheduleRun, error)
	ListRuns(ctx context.Context, filter RunFilter) ([]*models.ScheduleRun, string, error)
	ListJobs(ctx context.Context, runID string) ([]*models.ScheduleJob, error)
	PendingJobCount(ctx context.Context) (int, error)
	ActiveRunCount(ctx context.Context) (int, error)
	RegisterWorker(ctx context.Context, workerID string, info *models.WorkerInfo, lease time.Duration) error
	UnregisterWorker(ctx context.Context, workerID string) error
	ListWorkers(ctx context.Context) ([]*models.WorkerInfo, error)
	CompactExpiredRuns(ctx context.Context, retention time.Duration) (int, error)
	Ping(ctx context.Context) error
	IsLeader() bool
}

type RateLimiter interface {
	Wait(ctx context.Context, provider string) error
}
