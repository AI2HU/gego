package shared

import (
	"time"
)

// ResponseFilter provides filtering options for listing responses
type ResponseFilter struct {
	PromptID   string
	PromptIDs  []string
	LLMID      string
	ScheduleID string
	RunID      string
	JobID      string
	Keyword        string
	HasSearchURLs  bool
	ErrorsOnly     bool
	StartTime  *time.Time
	EndTime    *time.Time
	Limit      int
	Offset     int
}
