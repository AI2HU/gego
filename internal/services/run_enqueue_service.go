package services

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/google/uuid"

	"github.com/AI2HU/gego/internal/jobs"
	etcdstore "github.com/AI2HU/gego/internal/jobs/etcd"
	"github.com/AI2HU/gego/internal/models"
)

type RunEnqueueService struct {
	db    scheduleEnqueueDB
	store jobs.Store
	keys  *etcdstore.Keys
}

type scheduleEnqueueDB interface {
	GetSchedule(ctx context.Context, id string) (*models.Schedule, error)
	GetPrompt(ctx context.Context, id string) (*models.Prompt, error)
	GetLLM(ctx context.Context, id string) (*models.LLMConfig, error)
}

func NewRunEnqueueService(database scheduleEnqueueDB, store jobs.Store, keys *etcdstore.Keys) *RunEnqueueService {
	return &RunEnqueueService{db: database, store: store, keys: keys}
}

func (s *RunEnqueueService) EnqueueSchedule(ctx context.Context, scheduleID string, trigger models.ScheduleRunTrigger, cronSlot string, dedupTTL time.Duration) (string, error) {
	schedule, err := s.db.GetSchedule(ctx, scheduleID)
	if err != nil {
		return "", fmt.Errorf("get schedule: %w", err)
	}

	if err := s.validateScheduleTargets(ctx, schedule); err != nil {
		return "", err
	}

	llms, err := s.loadEnabledLLMs(ctx, schedule.LLMIDs)
	if err != nil {
		return "", err
	}
	if len(llms) == 0 {
		return "", fmt.Errorf("no enabled LLMs found for schedule")
	}

	runID := uuid.New().String()
	run := &models.ScheduleRun{
		ID:         runID,
		ScheduleID: scheduleID,
		Trigger:    trigger,
		Status:     models.ScheduleRunStatusPending,
		CronSlot:   cronSlot,
	}

	jobList := buildModelJobs(runID, scheduleID, schedule.PromptIDs, llms, schedule.Temperature)
	run.TotalJobs = len(jobList)

	dedupKey := ""
	if cronSlot != "" && s.keys != nil {
		dedupKey = s.keys.Dedup(scheduleID, cronSlot)
	}

	if err := s.store.CreateRun(ctx, run, jobList, dedupKey, dedupTTL); err != nil {
		return "", err
	}

	return runID, nil
}

func (s *RunEnqueueService) validateScheduleTargets(ctx context.Context, schedule *models.Schedule) error {
	if len(schedule.PromptIDs) == 0 {
		return fmt.Errorf("no prompts found for schedule")
	}
	for _, promptID := range schedule.PromptIDs {
		if _, err := s.db.GetPrompt(ctx, promptID); err != nil {
			return fmt.Errorf("get prompt %s: %w", promptID, err)
		}
	}
	return nil
}

func (s *RunEnqueueService) loadEnabledLLMs(ctx context.Context, llmIDs []string) ([]*models.LLMConfig, error) {
	llms := make([]*models.LLMConfig, 0, len(llmIDs))
	for _, llmID := range llmIDs {
		llmConfig, err := s.db.GetLLM(ctx, llmID)
		if err != nil {
			return nil, fmt.Errorf("get llm %s: %w", llmID, err)
		}
		if llmConfig.Enabled {
			llms = append(llms, llmConfig)
		}
	}
	return llms, nil
}

func buildModelJobs(runID, scheduleID string, promptIDs []string, llms []*models.LLMConfig, scheduleTemperature float64) []*models.ScheduleJob {
	promptIDs = append([]string(nil), promptIDs...)
	jobs := make([]*models.ScheduleJob, 0, len(llms))

	for _, llmConfig := range llms {
		temperature := scheduleTemperature
		if scheduleTemperature == -1.0 {
			temperature = rand.Float64()
		}

		jobs = append(jobs, &models.ScheduleJob{
			ID:          uuid.New().String(),
			RunID:       runID,
			ScheduleID:  scheduleID,
			PromptIDs:   append([]string(nil), promptIDs...),
			LLMID:       llmConfig.ID,
			Provider:    llmConfig.Provider,
			Temperature: temperature,
			Status:      models.ScheduleJobStatusPending,
			MaxAttempts: 3,
		})
	}

	return jobs
}

func (s *RunEnqueueService) CancelRunsForSchedule(ctx context.Context, scheduleID string) error {
	runs, _, err := s.store.ListRuns(ctx, jobs.RunFilter{ScheduleID: scheduleID, Limit: 100})
	if err != nil {
		return err
	}
	for _, run := range runs {
		if run.Status == models.ScheduleRunStatusPending || run.Status == models.ScheduleRunStatusRunning {
			if err := s.store.CancelRun(ctx, run.ID); err != nil {
				return err
			}
		}
	}
	return nil
}
