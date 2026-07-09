package fixtures

import (
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/AI2HU/gego/internal/models"
)

const responseCount = 60

func generateResponses(set *fixtureSet) []*models.Response {
	if len(set.Prompts) == 0 || len(set.LLMs) == 0 {
		return nil
	}

	brandNames := make([]string, 0, len(set.Brands))
	for _, brand := range set.Brands {
		brandNames = append(brandNames, brand.Name)
	}

	now := time.Now().UTC()
	responses := make([]*models.Response, 0, responseCount)

	for i := 0; i < responseCount; i++ {
		prompt := set.Prompts[i%len(set.Prompts)]
		llm := set.LLMs[i%len(set.LLMs)]

		daysAgo := 29 - (i * 29 / (responseCount - 1))
		createdAt := now.AddDate(0, 0, -daysAgo).Add(time.Duration(i%12) * time.Hour)

		primaryBrand := brandNames[i%len(brandNames)]
		secondaryBrand := brandNames[(i+1)%len(brandNames)]
		responseText := fmt.Sprintf(
			"Based on current market trends, %s and %s are among the leading brands mentioned by analysts. %s continues to innovate while %s expands its global footprint.",
			primaryBrand, secondaryBrand, primaryBrand, secondaryBrand,
		)

		var scheduleID, runID string
		if len(set.Schedules) > 0 && i%3 == 0 {
			schedule := set.Schedules[i%len(set.Schedules)]
			scheduleID = schedule.ID
			runID = fmt.Sprintf("run-%s-%d", schedule.ID, daysAgo)
		}

		responses = append(responses, &models.Response{
			ID:           fmt.Sprintf("resp-dev-%03d", i+1),
			PromptID:     prompt.ID,
			PromptText:   prompt.Template,
			LLMID:        llm.ID,
			LLMName:      llm.Name,
			LLMProvider:  llm.Provider,
			LLMModel:     llm.Model,
			ResponseText: responseText,
			Temperature:  0.5 + float64(i%5)*0.1,
			TokensUsed:   180 + (i % 120),
			ScheduleID:   scheduleID,
			RunID:        runID,
			JobID:        uuid.New().String(),
			SearchURLs: []models.SearchURL{
				{
					SearchQuery:   prompt.Template,
					URL:           fmt.Sprintf("https://www.example.com/articles/%s-%d", primaryBrand, i+1),
					Title:         fmt.Sprintf("%s market analysis", primaryBrand),
					CitationIndex: 1,
				},
				{
					SearchQuery:   prompt.Template,
					URL:           fmt.Sprintf("https://news.example.org/brands/%s", secondaryBrand),
					Title:         fmt.Sprintf("Why %s matters in 2025", secondaryBrand),
					CitationIndex: 2,
				},
			},
			CreatedAt: createdAt,
		})
	}

	return responses
}
