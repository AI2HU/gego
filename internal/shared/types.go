package shared

import (
	"time"
)

// ResponseFilter provides filtering options for listing responses
type ResponseFilter struct {
	PromptID   string
	LLMID      string
	ScheduleID string
	Keyword    string
	StartTime  *time.Time
	EndTime    *time.Time
	Limit      int
	Offset     int
}
