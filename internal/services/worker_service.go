package services

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/sourcegraph/conc/pool"

	"github.com/AI2HU/gego/internal/db"
	"github.com/AI2HU/gego/internal/jobs"
	"github.com/AI2HU/gego/internal/logger"
	"github.com/AI2HU/gego/internal/models"
	"github.com/AI2HU/gego/internal/shared"
)

type WorkerService struct {
	db           db.Database
	store        jobs.Store
	executor     *PromptExecutor
	cfg          jobs.Config
	hostname     string
	shutdown     chan struct{}
	shutdownOnce sync.Once
	shutdownWg   sync.WaitGroup
}

func NewWorkerService(database db.Database, store jobs.Store, executor *PromptExecutor, cfg jobs.Config) *WorkerService {
	hostname, _ := os.Hostname()
	if hostname == "" {
		hostname = "worker"
	}
	return &WorkerService{
		db:       database,
		store:    store,
		executor: executor,
		cfg:      cfg,
		hostname: hostname,
		shutdown: make(chan struct{}),
	}
}

func (w *WorkerService) Start(ctx context.Context) error {
	if err := w.store.Ping(ctx); err != nil {
		return fmt.Errorf("etcd unavailable: %w", err)
	}

	workerID := fmt.Sprintf("%s-%d-%s", w.hostname, os.Getpid(), uuid.New().String()[:8])
	logger.Info("worker starting id=%s concurrency=%d", workerID, w.cfg.WorkerConcurrency)

	w.shutdownWg.Add(3)
	go w.heartbeatLoop(ctx, workerID)
	go w.reclaimLoop(ctx)
	go w.compactLoop(ctx)

	jobCh, err := w.store.WatchPendingJobs(ctx)
	if err != nil {
		return err
	}

	p := pool.New().WithErrors().WithMaxGoroutines(w.cfg.WorkerConcurrency)

	for {
		select {
		case <-ctx.Done():
			w.signalShutdown()
			p.Wait()
			_ = w.store.UnregisterWorker(context.Background(), workerID)
			w.shutdownWg.Wait()
			return ctx.Err()
		case <-w.shutdown:
			p.Wait()
			_ = w.store.UnregisterWorker(context.Background(), workerID)
			w.shutdownWg.Wait()
			return nil
		case job, ok := <-jobCh:
			if !ok {
				return fmt.Errorf("job watch channel closed")
			}
			j := job
			p.Go(func() error {
				return w.processJob(ctx, workerID, j)
			})
		}
	}
}

func (w *WorkerService) signalShutdown() {
	w.shutdownOnce.Do(func() {
		close(w.shutdown)
	})
}

func (w *WorkerService) Stop() {
	w.signalShutdown()
}

func (w *WorkerService) heartbeatLoop(ctx context.Context, workerID string) {
	defer w.shutdownWg.Done()
	ticker := time.NewTicker(w.cfg.JobLease / 3)
	defer ticker.Stop()

	info := &models.WorkerInfo{
		ID:        workerID,
		Hostname:  w.hostname,
		StartedAt: time.Now().UTC(),
	}

	for {
		if err := w.store.RegisterWorker(ctx, workerID, info, w.cfg.JobLease); err != nil {
			logger.Error("worker heartbeat failed: %v", err)
		}
		select {
		case <-ctx.Done():
			return
		case <-w.shutdown:
			return
		case <-ticker.C:
		}
	}
}

func (w *WorkerService) reclaimLoop(ctx context.Context) {
	defer w.shutdownWg.Done()
	ticker := time.NewTicker(w.cfg.ReclaimInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-w.shutdown:
			return
		case <-ticker.C:
			n, err := w.store.ReclaimStaleJobs(ctx)
			if err != nil {
				logger.Error("reclaim stale jobs: %v", err)
			} else if n > 0 {
				logger.Info("reclaimed %d stale jobs", n)
			}
		}
	}
}

func (w *WorkerService) compactLoop(ctx context.Context) {
	defer w.shutdownWg.Done()
	ticker := time.NewTicker(w.cfg.CompactInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-w.shutdown:
			return
		case <-ticker.C:
			n, err := w.store.CompactExpiredRuns(ctx, w.cfg.RunRetention)
			if err != nil {
				logger.Error("compact expired runs: %v", err)
			} else if n > 0 {
				logger.Info("compacted %d expired runs", n)
			}
		}
	}
}

func (w *WorkerService) processJob(ctx context.Context, workerID string, job *models.ScheduleJob) error {
	run, err := w.store.GetRun(ctx, job.RunID)
	if err != nil {
		return err
	}
	if run.Status == models.ScheduleRunStatusCancelled {
		return nil
	}

	claimed, err := w.store.ClaimJob(ctx, job, workerID, w.cfg.JobLease)
	if err != nil {
		return nil
	}

	responseIDs, err := w.executeModelJob(ctx, claimed)
	if err != nil {
		if failErr := w.store.FailJob(ctx, claimed.ID, err); failErr != nil {
			logger.Error("fail job %s: %v", claimed.ID, failErr)
		}
		w.finalizeRunIfNeeded(ctx, claimed.RunID)
		return err
	}

	if err := w.store.CompleteJob(ctx, claimed.ID, responseIDs); err != nil {
		return err
	}

	return w.finalizeRunIfNeeded(ctx, claimed.RunID)
}

func (w *WorkerService) executeModelJob(ctx context.Context, job *models.ScheduleJob) ([]string, error) {
	if len(job.PromptIDs) == 0 {
		return nil, fmt.Errorf("job has no prompts")
	}

	llmConfig, err := w.db.GetLLM(ctx, job.LLMID)
	if err != nil {
		return nil, err
	}

	responseIDs := make([]string, 0, len(job.PromptIDs))
	params := PromptExecutionParams{
		ScheduleID:  job.ScheduleID,
		RunID:       job.RunID,
		JobID:       job.ID,
		Temperature: job.Temperature,
	}

	for _, promptID := range job.PromptIDs {
		if w.hasSuccessfulResponse(ctx, job.RunID, job.ID, promptID) {
			continue
		}

		prompt, err := w.db.GetPrompt(ctx, promptID)
		if err != nil {
			return nil, err
		}

		responseID, err := w.executor.Execute(ctx, params, prompt, llmConfig)
		if err != nil {
			return nil, err
		}
		responseIDs = append(responseIDs, responseID)
	}

	return responseIDs, nil
}

func (w *WorkerService) hasSuccessfulResponse(ctx context.Context, runID, jobID, promptID string) bool {
	responses, err := w.db.ListResponses(ctx, shared.ResponseFilter{
		PromptID: promptID,
		RunID:    runID,
		JobID:    jobID,
		Limit:    1,
	})
	if err != nil || len(responses) == 0 {
		return false
	}
	return responses[0].Error == ""
}

func (w *WorkerService) finalizeRunIfNeeded(ctx context.Context, runID string) error {
	run, err := w.store.UpdateRunProgress(ctx, runID)
	if err != nil {
		return err
	}
	if run.Status != models.ScheduleRunStatusCompleted && run.Status != models.ScheduleRunStatusFailed {
		return nil
	}

	schedule, err := w.db.GetSchedule(ctx, run.ScheduleID)
	if err != nil {
		return err
	}
	now := time.Now().UTC()
	schedule.LastRun = &now
	return w.db.UpdateSchedule(ctx, schedule)
}
