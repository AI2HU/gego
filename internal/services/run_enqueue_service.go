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
	db         interface {
		GetSchedule(ctx context.Context, id string) (*models.Schedule, error)
		GetPrompt(ctx context.Context, id string) (*models.Prompt, error)
		GetLLM(ctx context.Context, id string) (*models.LLMConfig, error)
	}
	store jobs.Store
	keys  *etcdstore.Keys
}

func NewRunEnqueueService(database interface {
	GetSchedule(ctx context.Context, id string) (*models.Schedule, error)
	GetPrompt(ctx context.Context, id string) (*models.Prompt, error)
	GetLLM(ctx context.Context, id string) (*models.LLMConfig, error)
}, store jobs.Store, keys *etcdstore.Keys) *RunEnqueueService {
	return &RunEnqueueService{db: database, store: store, keys: keys}
}

func (s *RunEnqueueService) EnqueueSchedule(ctx context.Context, scheduleID string, trigger models.ScheduleRunTrigger, cronSlot string, dedupTTL time.Duration) (string, error) {
	schedule, err := s.db.GetSchedule(ctx, scheduleID)
	if err != nil {
		return "", fmt.Errorf("get schedule: %w", err)
	}

	prompts := make([]*models.Prompt, 0, len(schedule.PromptIDs))
	for _, promptID := range schedule.PromptIDs {
		prompt, err := s.db.GetPrompt(ctx, promptID)
		if err != nil {
			return "", fmt.Errorf("get prompt %s: %w", promptID, err)
		}
		prompts = append(prompts, prompt)
	}

	llms := make([]*models.LLMConfig, 0, len(schedule.LLMIDs))
	providerLLMs := make(map[string]*models.LLMConfig)
	providerOrder := make([]string, 0)
	for _, llmID := range schedule.LLMIDs {
		llmConfig, err := s.db.GetLLM(ctx, llmID)
		if err != nil {
			return "", fmt.Errorf("get llm %s: %w", llmID, err)
		}
		if !llmConfig.Enabled {
			continue
		}
		if _, exists := providerLLMs[llmConfig.Provider]; exists {
			continue
		}
		providerLLMs[llmConfig.Provider] = llmConfig
		providerOrder = append(providerOrder, llmConfig.Provider)
		llms = append(llms, llmConfig)
	}

	if len(prompts) == 0 {
		return "", fmt.Errorf("no prompts found for schedule")
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

	jobList := make([]*models.ScheduleJob, 0, len(prompts)*len(providerOrder))
	for _, prompt := range prompts {
		for _, provider := range providerOrder {
			llmConfig := providerLLMs[provider]
			temperature := schedule.Temperature
			if schedule.Temperature == -1.0 {
				temperature = rand.Float64()
			}

			jobList = append(jobList, &models.ScheduleJob{
				ID:          uuid.New().String(),
				RunID:       runID,
				ScheduleID:  scheduleID,
				PromptID:    prompt.ID,
				LLMID:       llmConfig.ID,
				Provider:    llmConfig.Provider,
				Temperature: temperature,
				Status:      models.ScheduleJobStatusPending,
				MaxAttempts: 3,
			})
		}
	}

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
