package services

import (
	"context"
	"fmt"
	"time"

	"github.com/AI2HU/gego/internal/db"
	"github.com/AI2HU/gego/internal/models"
)

// ScheduleService provides business logic for schedule management
type ScheduleService struct {
	db db.Database
}

// NewScheduleService creates a new schedule service
func NewScheduleService(database db.Database) *ScheduleService {
	return &ScheduleService{db: database}
}

// ValidateSchedule validates schedule configuration
func (s *ScheduleService) ValidateSchedule(schedule *models.Schedule) error {
	if schedule.Name == "" {
		return fmt.Errorf("schedule name is required")
	}
	if len(schedule.PromptIDs) == 0 {
		return fmt.Errorf("at least one prompt is required")
	}
	if len(schedule.LLMIDs) == 0 {
		return fmt.Errorf("at least one LLM is required")
	}
	if schedule.CronExpr == "" {
		return fmt.Errorf("cron expression is required")
	}
	if schedule.Temperature < 0.0 || schedule.Temperature > 1.0 {
		return fmt.Errorf("temperature must be between 0.0 and 1.0, got: %.2f", schedule.Temperature)
	}

	// Validate that all referenced prompts exist
	for _, promptID := range schedule.PromptIDs {
		if _, err := s.db.GetPrompt(context.Background(), promptID); err != nil {
			return fmt.Errorf("prompt %s not found: %w", promptID, err)
		}
	}

	// Validate that all referenced LLMs exist
	for _, llmID := range schedule.LLMIDs {
		if _, err := s.db.GetLLM(context.Background(), llmID); err != nil {
			return fmt.Errorf("LLM %s not found: %w", llmID, err)
		}
	}

	return nil
}

// CreateSchedule creates a new schedule
func (s *ScheduleService) CreateSchedule(ctx context.Context, schedule *models.Schedule) error {
	if err := s.ValidateSchedule(schedule); err != nil {
		return err
	}
	return s.db.CreateSchedule(ctx, schedule)
}

// UpdateSchedule updates an existing schedule
func (s *ScheduleService) UpdateSchedule(ctx context.Context, schedule *models.Schedule) error {
	if err := s.ValidateSchedule(schedule); err != nil {
		return err
	}
	return s.db.UpdateSchedule(ctx, schedule)
}

// GetSchedule retrieves a schedule by ID
func (s *ScheduleService) GetSchedule(ctx context.Context, id string) (*models.Schedule, error) {
	return s.db.GetSchedule(ctx, id)
}

// ListSchedules lists schedules with optional filtering
func (s *ScheduleService) ListSchedules(ctx context.Context, enabled *bool) ([]*models.Schedule, error) {
	return s.db.ListSchedules(ctx, enabled)
}

// DeleteSchedule deletes a schedule
func (s *ScheduleService) DeleteSchedule(ctx context.Context, id string) error {
	return s.db.DeleteSchedule(ctx, id)
}

// EnableSchedule enables a schedule
func (s *ScheduleService) EnableSchedule(ctx context.Context, id string) error {
	schedule, err := s.db.GetSchedule(ctx, id)
	if err != nil {
		return err
	}
	schedule.Enabled = true
	return s.db.UpdateSchedule(ctx, schedule)
}

// DisableSchedule disables a schedule
func (s *ScheduleService) DisableSchedule(ctx context.Context, id string) error {
	schedule, err := s.db.GetSchedule(ctx, id)
	if err != nil {
		return err
	}
	schedule.Enabled = false
	return s.db.UpdateSchedule(ctx, schedule)
}

// GetEnabledSchedules returns only enabled schedules
func (s *ScheduleService) GetEnabledSchedules(ctx context.Context) ([]*models.Schedule, error) {
	enabled := true
	return s.db.ListSchedules(ctx, &enabled)
}

// UpdateLastRun updates the last run time for a schedule
func (s *ScheduleService) UpdateLastRun(ctx context.Context, id string, runTime time.Time) error {
	schedule, err := s.db.GetSchedule(ctx, id)
	if err != nil {
		return err
	}
	schedule.LastRun = &runTime
	return s.db.UpdateSchedule(ctx, schedule)
}

// UpdateNextRun updates the next run time for a schedule
func (s *ScheduleService) UpdateNextRun(ctx context.Context, id string, nextRun time.Time) error {
	schedule, err := s.db.GetSchedule(ctx, id)
	if err != nil {
		return err
	}
	schedule.NextRun = &nextRun
	return s.db.UpdateSchedule(ctx, schedule)
}

// ValidateCronExpression validates cron expression format
func (s *ScheduleService) ValidateCronExpression(cronExpr string) error {
	if cronExpr == "" {
		return fmt.Errorf("cron expression is required")
	}

	// Basic validation - check if it has 5 parts
	parts := splitCronExpression(cronExpr)
	if len(parts) != 5 {
		return fmt.Errorf("invalid cron expression: %s (must have 5 parts)", cronExpr)
	}

	// Additional validation could be added here using a cron library
	return nil
}

// splitCronExpression splits a cron expression into its components
func splitCronExpression(cronExpr string) []string {
	// Simple split by space - in production, use a proper cron parser
	var parts []string
	var current string

	for _, char := range cronExpr {
		if char == ' ' {
			if current != "" {
				parts = append(parts, current)
				current = ""
			}
		} else {
			current += string(char)
		}
	}
	if current != "" {
		parts = append(parts, current)
	}

	return parts
}

// GetScheduleExecutionPlan returns the execution plan for a schedule
func (s *ScheduleService) GetScheduleExecutionPlan(ctx context.Context, scheduleID string) (*ScheduleExecutionPlan, error) {
	schedule, err := s.db.GetSchedule(ctx, scheduleID)
	if err != nil {
		return nil, err
	}

	plan := &ScheduleExecutionPlan{
		ScheduleID:   scheduleID,
		ScheduleName: schedule.Name,
		Temperature:  schedule.Temperature,
		Prompts:      make([]*models.Prompt, 0, len(schedule.PromptIDs)),
		LLMs:         make([]*models.LLMConfig, 0, len(schedule.LLMIDs)),
	}

	// Get all prompts
	for _, promptID := range schedule.PromptIDs {
		prompt, err := s.db.GetPrompt(ctx, promptID)
		if err != nil {
			return nil, fmt.Errorf("failed to get prompt %s: %w", promptID, err)
		}
		plan.Prompts = append(plan.Prompts, prompt)
	}

	// Get all LLMs
	for _, llmID := range schedule.LLMIDs {
		llm, err := s.db.GetLLM(ctx, llmID)
		if err != nil {
			return nil, fmt.Errorf("failed to get LLM %s: %w", llmID, err)
		}
		plan.LLMs = append(plan.LLMs, llm)
	}

	return plan, nil
}

// ScheduleExecutionPlan represents the execution plan for a schedule
type ScheduleExecutionPlan struct {
	ScheduleID      string              `json:"schedule_id"`
	ScheduleName    string              `json:"schedule_name"`
	Temperature     float64             `json:"temperature"`
	Prompts         []*models.Prompt    `json:"prompts"`
	LLMs            []*models.LLMConfig `json:"llms"`
	TotalExecutions int                 `json:"total_executions"`
}

// CalculateTotalExecutions calculates the total number of executions for a plan
func (plan *ScheduleExecutionPlan) CalculateTotalExecutions() int {
	return len(plan.Prompts) * len(plan.LLMs)
}
