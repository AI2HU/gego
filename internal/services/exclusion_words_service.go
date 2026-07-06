package services

import (
	"context"
	"fmt"
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
