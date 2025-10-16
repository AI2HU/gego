package services

import (
	"context"
	"fmt"
	"strings"

	"github.com/AI2HU/gego/internal/db"
	"github.com/AI2HU/gego/internal/models"
)

// PromptManagementService provides business logic for prompt management
type PromptManagementService struct {
	db db.Database
}

// NewPromptManagementService creates a new prompt management service
func NewPromptManagementService(database db.Database) *PromptManagementService {
	return &PromptManagementService{db: database}
}

// ValidatePrompt validates prompt configuration
func (s *PromptManagementService) ValidatePrompt(prompt *models.Prompt) error {
	if prompt.Template == "" {
		return fmt.Errorf("prompt template is required")
	}
	if len(strings.TrimSpace(prompt.Template)) == 0 {
		return fmt.Errorf("prompt template cannot be empty")
	}
	return nil
}

// CreatePrompt creates a new prompt
func (s *PromptManagementService) CreatePrompt(ctx context.Context, prompt *models.Prompt) error {
	if err := s.ValidatePrompt(prompt); err != nil {
		return err
	}
	return s.db.CreatePrompt(ctx, prompt)
}

// UpdatePrompt updates an existing prompt
func (s *PromptManagementService) UpdatePrompt(ctx context.Context, prompt *models.Prompt) error {
	if err := s.ValidatePrompt(prompt); err != nil {
		return err
	}
	return s.db.UpdatePrompt(ctx, prompt)
}

// GetPrompt retrieves a prompt by ID
func (s *PromptManagementService) GetPrompt(ctx context.Context, id string) (*models.Prompt, error) {
	return s.db.GetPrompt(ctx, id)
}

// ListPrompts lists prompts with optional filtering
func (s *PromptManagementService) ListPrompts(ctx context.Context, enabled *bool) ([]*models.Prompt, error) {
	return s.db.ListPrompts(ctx, enabled)
}

// DeletePrompt deletes a prompt
func (s *PromptManagementService) DeletePrompt(ctx context.Context, id string) error {
	return s.db.DeletePrompt(ctx, id)
}

// EnablePrompt enables a prompt
func (s *PromptManagementService) EnablePrompt(ctx context.Context, id string) error {
	prompt, err := s.db.GetPrompt(ctx, id)
	if err != nil {
		return err
	}
	prompt.Enabled = true
	return s.db.UpdatePrompt(ctx, prompt)
}

// DisablePrompt disables a prompt
func (s *PromptManagementService) DisablePrompt(ctx context.Context, id string) error {
	prompt, err := s.db.GetPrompt(ctx, id)
	if err != nil {
		return err
	}
	prompt.Enabled = false
	return s.db.UpdatePrompt(ctx, prompt)
}

// GetEnabledPrompts returns only enabled prompts
func (s *PromptManagementService) GetEnabledPrompts(ctx context.Context) ([]*models.Prompt, error) {
	enabled := true
	return s.db.ListPrompts(ctx, &enabled)
}

// ValidatePromptTags validates prompt tags
func (s *PromptManagementService) ValidatePromptTags(tags []string) error {
	for _, tag := range tags {
		tag = strings.TrimSpace(tag)
		if tag == "" {
			return fmt.Errorf("empty tags are not allowed")
		}
		if len(tag) > 50 {
			return fmt.Errorf("tag too long: %s (max 50 characters)", tag)
		}
	}
	return nil
}

// SearchPrompts searches prompts by template content
func (s *PromptManagementService) SearchPrompts(ctx context.Context, query string) ([]*models.Prompt, error) {
	// This would typically implement search logic
	// For now, we'll get all prompts and filter client-side
	prompts, err := s.db.ListPrompts(ctx, nil)
	if err != nil {
		return nil, err
	}

	var results []*models.Prompt
	queryLower := strings.ToLower(query)

	for _, prompt := range prompts {
		if strings.Contains(strings.ToLower(prompt.Template), queryLower) {
			results = append(results, prompt)
		}
	}

	return results, nil
}

// GetPromptsByTags returns prompts filtered by tags
func (s *PromptManagementService) GetPromptsByTags(ctx context.Context, tags []string) ([]*models.Prompt, error) {
	prompts, err := s.db.ListPrompts(ctx, nil)
	if err != nil {
		return nil, err
	}

	var results []*models.Prompt
	for _, prompt := range prompts {
		for _, searchTag := range tags {
			for _, promptTag := range prompt.Tags {
				if strings.EqualFold(promptTag, searchTag) {
					results = append(results, prompt)
					break
				}
			}
		}
	}

	return results, nil
}
