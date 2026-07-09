package fixtures

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"gopkg.in/yaml.v3"

	"github.com/AI2HU/gego/internal/db"
)

const EnvFixtures = "GEGO_FIXTURES"

type Loader struct {
	Set string
}

type Summary struct {
	LLMs           int
	Prompts        int
	Schedules      int
	Brands         int
	ExclusionWords int
	Responses      int
}

func (l *Loader) Reset(ctx context.Context, database db.Database) error {
	hybrid, ok := database.(*db.HybridDB)
	if !ok {
		return fmt.Errorf("fixtures reset requires hybrid database")
	}
	return hybrid.CleanDev(ctx)
}

func (l *Loader) Load(ctx context.Context, database db.Database) (Summary, error) {
	set, err := l.loadSet()
	if err != nil {
		return Summary{}, err
	}

	summary := Summary{}

	for _, item := range set.LLMs {
		if err := database.CreateLLM(ctx, item.toModel()); err != nil {
			return summary, fmt.Errorf("create llm %s: %w", item.ID, err)
		}
		summary.LLMs++
	}

	for _, item := range set.Prompts {
		if err := database.CreatePrompt(ctx, item.toModel()); err != nil {
			return summary, fmt.Errorf("create prompt %s: %w", item.ID, err)
		}
		summary.Prompts++
	}

	for _, item := range set.Brands {
		brand := item.toModel()
		if err := database.CreateBrand(ctx, brand); err != nil {
			return summary, fmt.Errorf("create brand %s: %w", item.ID, err)
		}
		summary.Brands++
		for _, alias := range brand.Aliases {
			if alias.ID == "" {
				alias.ID = uuid.New().String()
			}
			if err := database.CreateBrandAlias(ctx, alias); err != nil {
				return summary, fmt.Errorf("create brand alias %s: %w", alias.Alias, err)
			}
		}
	}

	for _, item := range set.ExclusionWords {
		if err := database.CreateExclusionWord(ctx, item.toModel()); err != nil {
			return summary, fmt.Errorf("create exclusion word %s: %w", item.Word, err)
		}
		summary.ExclusionWords++
	}

	for _, item := range set.Schedules {
		if err := database.CreateSchedule(ctx, item.toModel()); err != nil {
			return summary, fmt.Errorf("create schedule %s: %w", item.ID, err)
		}
		summary.Schedules++
	}

	responses := generateResponses(set)
	for _, response := range responses {
		if err := database.CreateResponse(ctx, response); err != nil {
			return summary, fmt.Errorf("create response %s: %w", response.ID, err)
		}
		summary.Responses++
	}

	return summary, nil
}

func (l *Loader) loadSet() (*fixtureSet, error) {
	switch l.Set {
	case "", "dev":
		return loadDevSet()
	default:
		return nil, fmt.Errorf("unknown fixture set %q", l.Set)
	}
}

func loadDevSet() (*fixtureSet, error) {
	set := &fixtureSet{}

	if err := loadYAMLFile("dev/llms.yaml", &set.LLMs); err != nil {
		return nil, err
	}
	if err := loadYAMLFile("dev/prompts.yaml", &set.Prompts); err != nil {
		return nil, err
	}
	if err := loadYAMLFile("dev/schedules.yaml", &set.Schedules); err != nil {
		return nil, err
	}
	if err := loadYAMLFile("dev/brands.yaml", &set.Brands); err != nil {
		return nil, err
	}
	if err := loadYAMLFile("dev/exclusion_words.yaml", &set.ExclusionWords); err != nil {
		return nil, err
	}

	return set, nil
}

func loadYAMLFile(path string, target any) error {
	data, err := devFiles.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read fixture %s: %w", path, err)
	}
	if err := yaml.Unmarshal(data, target); err != nil {
		return fmt.Errorf("parse fixture %s: %w", path, err)
	}
	return nil
}
