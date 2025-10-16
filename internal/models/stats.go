package models

import (
	"time"
)

// TimeSeriesPoint represents a point in time-series data
type TimeSeriesPoint struct {
	Timestamp time.Time `json:"timestamp"`
	Count     int       `json:"count"`
}

// KeywordCount represents a keyword and its mention count
type KeywordCount struct {
	Keyword string `json:"keyword"`
	Count   int    `json:"count"`
}

// PromptStats represents aggregated statistics for a prompt
type PromptStats struct {
	PromptID       string         `json:"prompt_id"`
	TotalResponses int            `json:"total_responses"`
	UniqueLLMs     int            `json:"unique_llms"`
	LLMCounts      map[string]int `json:"llm_counts"`
	AvgTokens      float64        `json:"avg_tokens"`
	UpdatedAt      time.Time      `json:"updated_at"`
}

// LLMStats represents aggregated statistics for an LLM
type LLMStats struct {
	LLMID          string         `json:"llm_id"`
	TotalResponses int            `json:"total_responses"`
	UniquePrompts  int            `json:"unique_prompts"`
	PromptCounts   map[string]int `json:"prompt_counts"`
	AvgTokens      float64        `json:"avg_tokens"`
	UpdatedAt      time.Time      `json:"updated_at"`
}

// KeywordStats represents on-demand calculated statistics for a keyword search
type KeywordStats struct {
	Keyword       string         `json:"keyword"`
	TotalMentions int            `json:"total_mentions"`
	UniquePrompts int            `json:"unique_prompts"`
	UniqueLLMs    int            `json:"unique_llms"`
	ByPrompt      map[string]int `json:"by_prompt"`   // prompt_id -> count
	ByLLM         map[string]int `json:"by_llm"`      // llm_id -> count
	ByProvider    map[string]int `json:"by_provider"` // provider -> count
	FirstSeen     time.Time      `json:"first_seen"`
	LastSeen      time.Time      `json:"last_seen"`
}
