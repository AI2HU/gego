package services

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/google/uuid"

	"github.com/AI2HU/gego/internal/db"
	"github.com/AI2HU/gego/internal/models"
	"github.com/AI2HU/gego/internal/shared"
)

// BrandsService manages brand mappings stored in PostgreSQL.
type BrandsService struct {
	db db.Database
}

// NewBrandsService creates a new brands service.
func NewBrandsService(database db.Database) *BrandsService {
	return &BrandsService{db: database}
}

// Initialize loads brand aliases into the in-memory cache.
func (s *BrandsService) Initialize(ctx context.Context) error {
	return s.RefreshCache(ctx)
}

// RefreshCache reloads brand aliases from the database into the shared cache.
func (s *BrandsService) RefreshCache(ctx context.Context) error {
	brands, err := s.db.ListBrands(ctx)
	if err != nil {
		return fmt.Errorf("failed to list brands: %w", err)
	}

	entries := make([]shared.BrandAliasEntry, 0)
	for _, brand := range brands {
		entries = append(entries, shared.BrandAliasEntry{
			Canonical:     brand.Name,
			Alias:         brand.Name,
			CaseSensitive: false,
		})
		for _, alias := range brand.Aliases {
			if strings.EqualFold(alias.Alias, brand.Name) {
				continue
			}
			entries = append(entries, shared.BrandAliasEntry{
				Canonical:     brand.Name,
				Alias:         alias.Alias,
				CaseSensitive: alias.CaseSensitive,
			})
		}
	}
	shared.ReloadBrandAliases(entries)
	return nil
}

// ListBrands returns all brands with aliases.
func (s *BrandsService) ListBrands(ctx context.Context) ([]*models.Brand, error) {
	return s.db.ListBrands(ctx)
}

// GetBrand returns a brand by ID.
func (s *BrandsService) GetBrand(ctx context.Context, id string) (*models.Brand, error) {
	return s.db.GetBrand(ctx, id)
}

// CreateBrand creates a brand with optional aliases.
func (s *BrandsService) CreateBrand(ctx context.Context, name string, aliases []models.BrandAliasRequest) (*models.Brand, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, fmt.Errorf("name is required")
	}

	existing, err := s.db.GetBrandByName(ctx, name)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, fmt.Errorf("brand already exists: %s", name)
	}

	brand := &models.Brand{
		ID:   uuid.New().String(),
		Name: name,
	}
	if err := s.db.CreateBrand(ctx, brand); err != nil {
		return nil, err
	}

	for _, aliasReq := range aliases {
		if _, err := s.createAlias(ctx, brand.ID, aliasReq.Alias, aliasReq.CaseSensitive); err != nil {
			return nil, err
		}
	}

	if err := s.RefreshCache(ctx); err != nil {
		return nil, err
	}
	return s.db.GetBrand(ctx, brand.ID)
}

// UpdateBrand updates a brand's canonical name.
func (s *BrandsService) UpdateBrand(ctx context.Context, id, name string) (*models.Brand, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, fmt.Errorf("name is required")
	}

	brand, err := s.db.GetBrand(ctx, id)
	if err != nil {
		return nil, err
	}

	existing, err := s.db.GetBrandByName(ctx, name)
	if err != nil {
		return nil, err
	}
	if existing != nil && existing.ID != id {
		return nil, fmt.Errorf("brand already exists: %s", name)
	}

	brand.Name = name
	if err := s.db.UpdateBrand(ctx, brand); err != nil {
		return nil, err
	}

	if err := s.RefreshCache(ctx); err != nil {
		return nil, err
	}
	return s.db.GetBrand(ctx, id)
}

// DeleteBrand removes a brand and its aliases.
func (s *BrandsService) DeleteBrand(ctx context.Context, id string) error {
	if err := s.db.DeleteBrand(ctx, id); err != nil {
		return err
	}
	return s.RefreshCache(ctx)
}

// AddAlias adds an alias to an existing brand.
func (s *BrandsService) AddAlias(ctx context.Context, brandID, alias string, caseSensitive bool) (*models.BrandAlias, error) {
	entry, err := s.createAlias(ctx, brandID, alias, caseSensitive)
	if err != nil {
		return nil, err
	}
	if err := s.RefreshCache(ctx); err != nil {
		return nil, err
	}
	return entry, nil
}

func (s *BrandsService) createAlias(ctx context.Context, brandID, alias string, caseSensitive bool) (*models.BrandAlias, error) {
	alias = strings.TrimSpace(alias)
	if alias == "" {
		return nil, fmt.Errorf("alias is required")
	}

	if _, err := s.db.GetBrand(ctx, brandID); err != nil {
		return nil, err
	}

	existing, err := s.db.GetBrandAliasByAlias(ctx, alias)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, fmt.Errorf("alias already exists: %s", alias)
	}

	entry := &models.BrandAlias{
		ID:            uuid.New().String(),
		BrandID:       brandID,
		Alias:         alias,
		CaseSensitive: caseSensitive,
	}
	if err := s.db.CreateBrandAlias(ctx, entry); err != nil {
		return nil, err
	}
	return entry, nil
}

// UpdateAlias updates an alias.
func (s *BrandsService) UpdateAlias(ctx context.Context, brandID, aliasID, alias string, caseSensitive bool) (*models.BrandAlias, error) {
	alias = strings.TrimSpace(alias)
	if alias == "" {
		return nil, fmt.Errorf("alias is required")
	}

	current, err := s.db.GetBrandAlias(ctx, aliasID)
	if err != nil {
		return nil, err
	}
	if current.BrandID != brandID {
		return nil, fmt.Errorf("alias does not belong to brand: %s", brandID)
	}

	existing, err := s.db.GetBrandAliasByAlias(ctx, alias)
	if err != nil {
		return nil, err
	}
	if existing != nil && existing.ID != aliasID {
		return nil, fmt.Errorf("alias already exists: %s", alias)
	}

	current.Alias = alias
	current.CaseSensitive = caseSensitive
	if err := s.db.UpdateBrandAlias(ctx, current); err != nil {
		return nil, err
	}

	if err := s.RefreshCache(ctx); err != nil {
		return nil, err
	}
	return s.db.GetBrandAlias(ctx, aliasID)
}

// DeleteAlias removes an alias.
func (s *BrandsService) DeleteAlias(ctx context.Context, brandID, aliasID string) error {
	current, err := s.db.GetBrandAlias(ctx, aliasID)
	if err != nil {
		return err
	}
	if current.BrandID != brandID {
		return fmt.Errorf("alias does not belong to brand: %s", brandID)
	}

	if err := s.db.DeleteBrandAlias(ctx, aliasID); err != nil {
		return err
	}
	return s.RefreshCache(ctx)
}

// MapFromDetection maps a detected word to a canonical brand name.
func (s *BrandsService) MapFromDetection(ctx context.Context, alias, canonicalName string, caseSensitive bool) (*models.Brand, error) {
	alias = strings.TrimSpace(alias)
	canonicalName = strings.TrimSpace(canonicalName)
	if alias == "" || canonicalName == "" {
		return nil, fmt.Errorf("alias and name are required")
	}

	brand, err := s.db.GetBrandByName(ctx, canonicalName)
	if err != nil {
		return nil, err
	}
	if brand == nil {
		brand, err = s.CreateBrand(ctx, canonicalName, nil)
		if err != nil {
			return nil, err
		}
	}

	if _, err := s.AddAlias(ctx, brand.ID, alias, caseSensitive); err != nil {
		if strings.Contains(err.Error(), "already exists") {
			return s.db.GetBrand(ctx, brand.ID)
		}
		return nil, err
	}
	return s.db.GetBrand(ctx, brand.ID)
}

// GetSuggestedBrandWords returns capitalized words detected in responses that are not excluded or mapped.
func (s *BrandsService) GetSuggestedBrandWords(ctx context.Context, limit int, promptIDs []string, tagFilter bool) ([]models.SuggestedBrandWordResponse, error) {
	if limit < 1 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}

	if tagFilter && len(promptIDs) == 0 {
		return []models.SuggestedBrandWordResponse{}, nil
	}

	filter := shared.ResponseFilter{Limit: 10000}
	if len(promptIDs) > 0 {
		filter.PromptIDs = promptIDs
	}

	responses, err := s.db.ListResponses(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to list responses: %w", err)
	}

	excluded := getExcludedWordSetFromCache()
	wordCounts := make(map[string]int)

	for _, response := range responses {
		for _, word := range shared.ExtractAllCapitalizedWords(response.ResponseText) {
			if excluded[strings.ToLower(word)] {
				continue
			}
			if shared.IsKnownBrandToken(word) {
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

func getExcludedWordSetFromCache() map[string]bool {
	words := shared.GetExclusionWordsList()
	excluded := make(map[string]bool, len(words))
	for _, word := range words {
		excluded[strings.ToLower(word)] = true
	}
	return excluded
}
