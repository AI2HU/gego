package services

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/robfig/cron/v3"

	"github.com/AI2HU/gego/internal/jobs"
	"github.com/AI2HU/gego/internal/logger"
	"github.com/AI2HU/gego/internal/models"
)

type SchedulerService struct {
	store         jobs.Store
	enqueue       *RunEnqueueService
	cron          *cron.Cron
	running       bool
	mu            sync.RWMutex
	scheduleEntries map[string]cron.EntryID
	entriesMu     sync.RWMutex
	cancel        context.CancelFunc
}

func NewSchedulerService(store jobs.Store, enqueue *RunEnqueueService) *SchedulerService {
	c := cron.New(
		cron.WithLocation(time.UTC),
		cron.WithLogger(cron.DefaultLogger),
		cron.WithChain(cron.Recover(cron.DefaultLogger)),
	)

	return &SchedulerService{
		store:           store,
		enqueue:         enqueue,
		cron:            c,
		scheduleEntries: make(map[string]cron.EntryID),
	}
}

func (s *SchedulerService) Start(ctx context.Context, listSchedules func(context.Context) ([]*models.Schedule, error)) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return fmt.Errorf("scheduler already running")
	}

	if err := s.store.Ping(ctx); err != nil {
		return fmt.Errorf("etcd unavailable: %w", err)
	}

	runCtx, cancel := context.WithCancel(ctx)
	s.cancel = cancel

	if etcdStore, ok := s.store.(interface {
		StartElection(context.Context) error
	}); ok {
		if err := etcdStore.StartElection(runCtx); err != nil {
			cancel()
			return err
		}
	}

	schedules, err := listSchedules(ctx)
	if err != nil {
		cancel()
		return fmt.Errorf("failed to load schedules: %w", err)
	}

	registeredCount := 0
	for _, schedule := range schedules {
		if err := s.registerSchedule(runCtx, schedule); err != nil {
			logger.Error("Failed to register schedule %s: %v", schedule.ID, err)
		} else {
			registeredCount++
		}
	}

	logger.Info("Registered %d enabled schedule(s) with cron", registeredCount)

	s.cron.Start()
	s.running = true
	logger.Info("Scheduler started successfully")
	return nil
}

func (s *SchedulerService) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return
	}

	if s.cancel != nil {
		s.cancel()
	}

	if etcdStore, ok := s.store.(interface{ StopElection() }); ok {
		etcdStore.StopElection()
	}

	s.cron.Stop()
	s.running = false

	s.entriesMu.Lock()
	s.scheduleEntries = make(map[string]cron.EntryID)
	s.entriesMu.Unlock()

	logger.Info("Scheduler stopped")
}

func (s *SchedulerService) GetStatus(ctx context.Context) (running bool, enabledSchedules int, isLeader bool, pendingJobs int, activeRuns int, activeWorkers int, err error) {
	s.mu.RLock()
	running = s.running
	s.mu.RUnlock()

	isLeader = s.store.IsLeader()
	pendingJobs, err = s.store.PendingJobCount(ctx)
	if err != nil {
		return running, 0, isLeader, 0, 0, 0, err
	}
	activeRuns, err = s.store.ActiveRunCount(ctx)
	if err != nil {
		return running, 0, isLeader, pendingJobs, 0, 0, err
	}
	workers, err := s.store.ListWorkers(ctx)
	if err != nil {
		return running, 0, isLeader, pendingJobs, activeRuns, 0, err
	}

	return running, enabledSchedules, isLeader, pendingJobs, activeRuns, len(workers), nil
}

func (s *SchedulerService) Reload(ctx context.Context, listSchedules func(context.Context) ([]*models.Schedule, error)) error {
	s.Stop()
	time.Sleep(100 * time.Millisecond)
	return s.Start(ctx, listSchedules)
}

func (s *SchedulerService) registerSchedule(ctx context.Context, schedule *models.Schedule) error {
	scheduleID := schedule.ID
	cronExpr := schedule.CronExpr

	jobFunc := func() {
		if !s.store.IsLeader() {
			return
		}

		cronSlot := time.Now().UTC().Truncate(time.Minute).Format(time.RFC3339)
		runID, err := s.enqueue.EnqueueSchedule(context.Background(), scheduleID, models.ScheduleRunTriggerCron, cronSlot, time.Minute)
		if err != nil {
			logger.Error("Failed to enqueue schedule %s: %v", scheduleID, err)
			return
		}
		logger.Info("Enqueued cron run %s for schedule %s", runID, scheduleID)
	}

	entryID, err := s.cron.AddFunc(cronExpr, jobFunc)
	if err != nil {
		return fmt.Errorf("failed to add cron job: %w", err)
	}

	s.entriesMu.Lock()
	s.scheduleEntries[scheduleID] = entryID
	s.entriesMu.Unlock()

	logger.Info("Registered schedule %s with cron: %s", scheduleID, cronExpr)
	return nil
}
