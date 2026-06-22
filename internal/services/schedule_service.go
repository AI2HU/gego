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

	for _, promptID := range schedule.PromptIDs {
		if _, err := s.db.GetPrompt(context.Background(), promptID); err != nil {
			return fmt.Errorf("prompt %s not found: %w", promptID, err)
		}
	}

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
