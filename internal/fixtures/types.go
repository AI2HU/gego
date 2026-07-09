package fixtures

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/AI2HU/gego/internal/models"
)

type fixtureLLM struct {
	ID       string            `yaml:"id"`
	Name     string            `yaml:"name"`
	Provider string            `yaml:"provider"`
	Model    string            `yaml:"model"`
	APIKey   string            `yaml:"api_key,omitempty"`
	BaseURL  string            `yaml:"base_url,omitempty"`
	Config   map[string]string `yaml:"config,omitempty"`
	Enabled  bool              `yaml:"enabled"`
}

type fixturePrompt struct {
	ID       string   `yaml:"id"`
	Template string   `yaml:"template"`
	Tags     []string `yaml:"tags,omitempty"`
	Enabled  bool     `yaml:"enabled"`
}

type fixtureSchedule struct {
	ID          string   `yaml:"id"`
	Name        string   `yaml:"name"`
	PromptIDs   []string `yaml:"prompt_ids"`
	LLMIDs      []string `yaml:"llm_ids"`
	CronExpr    string   `yaml:"cron_expr"`
	Temperature float64  `yaml:"temperature,omitempty"`
	Enabled     bool     `yaml:"enabled"`
}

type fixtureBrandAlias struct {
	ID            string `yaml:"id"`
	Alias         string `yaml:"alias"`
	CaseSensitive bool   `yaml:"case_sensitive"`
}

type fixtureBrand struct {
	ID      string              `yaml:"id"`
	Name    string              `yaml:"name"`
	Aliases []fixtureBrandAlias `yaml:"aliases,omitempty"`
}

type fixtureExclusionWord struct {
	ID   string `yaml:"id"`
	Word string `yaml:"word"`
}

type fixtureSet struct {
	LLMs           []fixtureLLM           `yaml:"llms"`
	Prompts        []fixturePrompt        `yaml:"prompts"`
	Schedules      []fixtureSchedule      `yaml:"schedules"`
	Brands         []fixtureBrand         `yaml:"brands"`
	ExclusionWords []fixtureExclusionWord `yaml:"exclusion_words"`
}

func parseRelativeTime(value string, now time.Time) (time.Time, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return now, nil
	}
	if !strings.HasPrefix(value, "-") {
		return time.Time{}, fmt.Errorf("unsupported time format %q: use relative offsets like -7d", value)
	}

	unit := value[len(value)-1]
	numStr := value[1 : len(value)-1]
	amount, err := strconv.Atoi(numStr)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid relative time %q: %w", value, err)
	}

	switch unit {
	case 'd':
		return now.AddDate(0, 0, -amount), nil
	case 'h':
		return now.Add(-time.Duration(amount) * time.Hour), nil
	default:
		return time.Time{}, fmt.Errorf("unsupported time unit in %q", value)
	}
}

func (f fixtureLLM) toModel() *models.LLMConfig {
	return &models.LLMConfig{
		ID:       f.ID,
		Name:     f.Name,
		Provider: f.Provider,
		Model:    f.Model,
		APIKey:   f.APIKey,
		BaseURL:  f.BaseURL,
		Config:   f.Config,
		Enabled:  f.Enabled,
	}
}

func (f fixturePrompt) toModel() *models.Prompt {
	return &models.Prompt{
		ID:       f.ID,
		Template: f.Template,
		Tags:     f.Tags,
		Enabled:  f.Enabled,
	}
}

func (f fixtureSchedule) toModel() *models.Schedule {
	temp := f.Temperature
	if temp == 0 {
		temp = 0.7
	}
	return &models.Schedule{
		ID:          f.ID,
		Name:        f.Name,
		PromptIDs:   f.PromptIDs,
		LLMIDs:      f.LLMIDs,
		CronExpr:    f.CronExpr,
		Temperature: temp,
		Enabled:     f.Enabled,
	}
}

func (f fixtureBrand) toModel() *models.Brand {
	brand := &models.Brand{
		ID:   f.ID,
		Name: f.Name,
	}
	for _, alias := range f.Aliases {
		brand.Aliases = append(brand.Aliases, &models.BrandAlias{
			ID:            alias.ID,
			BrandID:       f.ID,
			Alias:         alias.Alias,
			CaseSensitive: alias.CaseSensitive,
		})
	}
	return brand
}

func (f fixtureExclusionWord) toModel() *models.ExclusionWord {
	return &models.ExclusionWord{
		ID:   f.ID,
		Word: f.Word,
	}
}
