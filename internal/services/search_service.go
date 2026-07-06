package services

import (
	"context"
	"time"

	"github.com/AI2HU/gego/internal/db"
	"github.com/AI2HU/gego/internal/models"
	"github.com/AI2HU/gego/internal/shared"
)

// SearchService provides business logic for searching responses.
type SearchService struct {
	db db.Database
}

// NewSearchService creates a new search service.
func NewSearchService(database db.Database) *SearchService {
	return &SearchService{db: database}
}

// SearchKeyword searches for a keyword and returns statistics.
func (s *SearchService) SearchKeyword(ctx context.Context, keyword string, startTime, endTime *time.Time, promptIDs []string) (*models.KeywordStats, error) {
	return s.db.SearchKeyword(ctx, keyword, startTime, endTime, promptIDs)
}

// ListResponses lists responses with filtering.
func (s *SearchService) ListResponses(ctx context.Context, filter shared.ResponseFilter) ([]*models.Response, error) {
	return s.db.ListResponses(ctx, filter)
}

// CountResponses counts responses matching the filter.
func (s *SearchService) CountResponses(ctx context.Context, filter shared.ResponseFilter) (int64, error) {
	return s.db.CountResponses(ctx, filter)
}
