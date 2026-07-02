package services

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/google/uuid"

	"github.com/AI2HU/gego/internal/db"
	"github.com/AI2HU/gego/internal/models"
	"github.com/AI2HU/gego/internal/shared"
)

// ExclusionWordsService manages exclusion words stored in PostgreSQL.
type ExclusionWordsService struct {
	db db.Database
}

// NewExclusionWordsService creates a new exclusion words service.
func NewExclusionWordsService(database db.Database) *ExclusionWordsService {
	return &ExclusionWordsService{db: database}
}

// Initialize loads exclusion words into the in-memory cache and migrates legacy file data if needed.
func (s *ExclusionWordsService) Initialize(ctx context.Context, legacyFilePath string) error {
	count, err := s.db.CountExclusionWords(ctx)
	if err != nil {
		return fmt.Errorf("failed to count exclusion words: %w", err)
	}

	if count == 0 && legacyFilePath != "" {
		if err := s.migrateFromFile(ctx, legacyFilePath); err != nil {
			return fmt.Errorf("failed to migrate exclusion words from file: %w", err)
		}
	}

	return s.RefreshCache(ctx)
}

func (s *ExclusionWordsService) migrateFromFile(ctx context.Context, path string) error {
	words, err := shared.LoadExclusionWordsFromFile(path)
	if err != nil {
		return err
	}
	for _, word := range words {
		if err := s.createIfMissing(ctx, word); err != nil {
			return err
		}
	}
	return nil
}

func (s *ExclusionWordsService) createIfMissing(ctx context.Context, word string) error {
	word = strings.TrimSpace(word)
	if word == "" {
		return nil
	}

	existing, err := s.db.GetExclusionWordByWord(ctx, word)
	if err != nil {
		return err
	}
	if existing != nil {
		return nil
	}

	return s.db.CreateExclusionWord(ctx, &models.ExclusionWord{
		ID:   uuid.New().String(),
		Word: word,
	})
}

// RefreshCache reloads exclusion words from the database into the shared cache.
func (s *ExclusionWordsService) RefreshCache(ctx context.Context) error {
	words, err := s.db.ListExclusionWords(ctx)
	if err != nil {
		return fmt.Errorf("failed to list exclusion words: %w", err)
	}

	wordList := make([]string, len(words))
	for i, word := range words {
		wordList[i] = word.Word
	}
	shared.ReloadExclusionWords(wordList)
	return nil
}

// ListExclusionWords returns all exclusion words.
func (s *ExclusionWordsService) ListExclusionWords(ctx context.Context) ([]*models.ExclusionWord, error) {
	return s.db.ListExclusionWords(ctx)
}

// CreateExclusionWord adds a new exclusion word.
func (s *ExclusionWordsService) CreateExclusionWord(ctx context.Context, word string) (*models.ExclusionWord, error) {
	word = strings.TrimSpace(word)
	if word == "" {
		return nil, fmt.Errorf("word is required")
	}

	existing, err := s.db.GetExclusionWordByWord(ctx, word)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return existing, nil
	}

	entry := &models.ExclusionWord{
		ID:   uuid.New().String(),
		Word: word,
	}
	if err := s.db.CreateExclusionWord(ctx, entry); err != nil {
		return nil, err
	}

	if err := s.RefreshCache(ctx); err != nil {
		return nil, err
	}
	return entry, nil
}

// DeleteExclusionWord removes an exclusion word by ID.
func (s *ExclusionWordsService) DeleteExclusionWord(ctx context.Context, id string) error {
	if err := s.db.DeleteExclusionWord(ctx, id); err != nil {
		return err
	}
	return s.RefreshCache(ctx)
}

// GetSuggestedBrandWords returns capitalized words detected in responses that are not yet excluded.
func (s *ExclusionWordsService) GetSuggestedBrandWords(ctx context.Context, limit int) ([]models.SuggestedBrandWordResponse, error) {
	if limit < 1 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}

	responses, err := s.db.ListResponses(ctx, shared.ResponseFilter{})
	if err != nil {
		return nil, fmt.Errorf("failed to list responses: %w", err)
	}

	excluded := getExcludedWordSet(ctx, s)
	wordCounts := make(map[string]int)

	for _, response := range responses {
		for _, word := range shared.ExtractAllCapitalizedWords(response.ResponseText) {
			if excluded[strings.ToLower(word)] {
				continue
			}
			wordCounts[word]++
		}
	}

	type kv struct {
		word  string
		count int
	}

	sorted := make([]kv, 0, len(wordCounts))
	for word, count := range wordCounts {
		sorted = append(sorted, kv{word: word, count: count})
	}

	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].count == sorted[j].count {
			return sorted[i].word < sorted[j].word
		}
		return sorted[i].count > sorted[j].count
	})

	if len(sorted) > limit {
		sorted = sorted[:limit]
	}

	results := make([]models.SuggestedBrandWordResponse, len(sorted))
	for i, item := range sorted {
		results[i] = models.SuggestedBrandWordResponse{
			Word:  item.word,
			Count: item.count,
		}
	}
	return results, nil
}

func getExcludedWordSet(ctx context.Context, s *ExclusionWordsService) map[string]bool {
	words, err := s.db.ListExclusionWords(ctx)
	if err != nil {
		return map[string]bool{}
	}

	excluded := make(map[string]bool, len(words))
	for _, word := range words {
		excluded[strings.ToLower(word.Word)] = true
	}
	return excluded
}

// ImportFromLegacyFile imports words from a legacy keywords_exclusion file.
func (s *ExclusionWordsService) ImportFromLegacyFile(ctx context.Context, path string) (int, error) {
	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil
		}
		return 0, err
	}
	defer file.Close()

	imported := 0
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		existing, err := s.db.GetExclusionWordByWord(ctx, line)
		if err != nil {
			return imported, err
		}
		if existing != nil {
			continue
		}
		if err := s.db.CreateExclusionWord(ctx, &models.ExclusionWord{
			ID:   uuid.New().String(),
			Word: line,
		}); err != nil {
			return imported, err
		}
		imported++
	}
	if err := scanner.Err(); err != nil {
		return imported, err
	}

	if imported > 0 {
		if err := s.RefreshCache(ctx); err != nil {
			return imported, err
		}
	}
	return imported, nil
}
