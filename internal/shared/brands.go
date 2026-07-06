package shared

import (
	"strings"
	"sync"
)

// BrandAliasEntry is used when reloading the brand alias cache.
type BrandAliasEntry struct {
	Canonical     string
	Alias         string
	CaseSensitive bool
}

// BrandSearchTerm represents a term to search for when looking up a brand.
type BrandSearchTerm struct {
	Term          string
	CaseSensitive bool
}

var (
	brandInsensitive      map[string]string
	brandSensitive        map[string]string
	brandByCanonical      map[string][]BrandSearchTerm
	brandByToken          map[string][]BrandSearchTerm
	brandBySensitiveToken map[string][]BrandSearchTerm
	brandCanonicals       map[string]string
	brandAliasesMu        sync.RWMutex
)

// ReloadBrandAliases rebuilds the in-memory brand alias cache.
func ReloadBrandAliases(entries []BrandAliasEntry) {
	brandAliasesMu.Lock()
	defer brandAliasesMu.Unlock()

	brandInsensitive = make(map[string]string)
	brandSensitive = make(map[string]string)
	brandByCanonical = make(map[string][]BrandSearchTerm)
	brandByToken = make(map[string][]BrandSearchTerm)
	brandBySensitiveToken = make(map[string][]BrandSearchTerm)
	brandCanonicals = make(map[string]string)

	addSearchTerm := func(canonical, term string, caseSensitive bool) {
		key := strings.ToLower(canonical)
		brandCanonicals[key] = canonical

		seen := false
		for _, existing := range brandByCanonical[key] {
			if existing.Term == term && existing.CaseSensitive == caseSensitive {
				seen = true
				break
			}
		}
		if !seen {
			brandByCanonical[key] = append(brandByCanonical[key], BrandSearchTerm{
				Term:          term,
				CaseSensitive: caseSensitive,
			})
		}
	}

	for _, entry := range entries {
		canonical := strings.TrimSpace(entry.Canonical)
		alias := strings.TrimSpace(entry.Alias)
		if canonical == "" || alias == "" {
			continue
		}

		if entry.CaseSensitive {
			brandSensitive[alias] = canonical
		} else {
			brandInsensitive[strings.ToLower(alias)] = canonical
		}
		addSearchTerm(canonical, alias, entry.CaseSensitive)
	}

	for key, canonical := range brandCanonicals {
		addSearchTerm(canonical, canonical, false)
		_ = key
	}

	for _, terms := range brandByCanonical {
		for _, term := range terms {
			if term.CaseSensitive {
				brandBySensitiveToken[term.Term] = terms
				continue
			}
			brandByToken[strings.ToLower(term.Term)] = terms
		}
	}
}

// ResolveBrand maps an extracted token to its canonical brand name if an alias matches.
func ResolveBrand(token string) string {
	brandAliasesMu.RLock()
	defer brandAliasesMu.RUnlock()

	if canonical, ok := brandSensitive[token]; ok {
		return canonical
	}
	if canonical, ok := brandInsensitive[strings.ToLower(token)]; ok {
		return canonical
	}
	return token
}

// IsKnownBrandToken reports whether the token is a registered alias or canonical name.
func IsKnownBrandToken(token string) bool {
	brandAliasesMu.RLock()
	defer brandAliasesMu.RUnlock()

	if _, ok := brandSensitive[token]; ok {
		return true
	}
	if _, ok := brandInsensitive[strings.ToLower(token)]; ok {
		return true
	}
	if _, ok := brandCanonicals[strings.ToLower(token)]; ok {
		return true
	}
	return false
}

// ExpandBrandSearchTerms returns all search terms for a keyword that matches a brand.
func ExpandBrandSearchTerms(keyword string) []BrandSearchTerm {
	brandAliasesMu.RLock()
	defer brandAliasesMu.RUnlock()

	trimmed := strings.TrimSpace(keyword)
	if trimmed == "" {
		return nil
	}

	if terms, ok := brandByToken[strings.ToLower(trimmed)]; ok {
		return copyBrandSearchTerms(terms)
	}

	if terms, ok := brandBySensitiveToken[trimmed]; ok {
		return copyBrandSearchTerms(terms)
	}

	return []BrandSearchTerm{{Term: trimmed, CaseSensitive: false}}
}

func copyBrandSearchTerms(terms []BrandSearchTerm) []BrandSearchTerm {
	result := make([]BrandSearchTerm, len(terms))
	copy(result, terms)
	return result
}

// CountSearchTermOccurrences counts occurrences of a search term in text.
func CountSearchTermOccurrences(text string, term BrandSearchTerm) int {
	if term.CaseSensitive {
		return countSubstringOccurrences(text, term.Term)
	}
	return CountOccurrences(text, term.Term)
}

func countSubstringOccurrences(text, keyword string) int {
	count := 0
	index := 0
	for {
		i := strings.Index(text[index:], keyword)
		if i == -1 {
			break
		}
		count++
		index += i + len(keyword)
	}
	return count
}

// CountBrandOccurrences counts total occurrences for all expanded brand search terms.
func CountBrandOccurrences(text string, terms []BrandSearchTerm) int {
	total := 0
	for _, term := range terms {
		total += CountSearchTermOccurrences(text, term)
	}
	return total
}

// GetBrandSearchTermStrings returns unique search strings for a keyword, including brand aliases.
func GetBrandSearchTermStrings(keyword string) []string {
	terms := ExpandBrandSearchTerms(keyword)
	if len(terms) == 0 {
		trimmed := strings.TrimSpace(keyword)
		if trimmed == "" {
			return nil
		}
		return []string{trimmed}
	}

	seen := make(map[string]bool)
	result := make([]string, 0, len(terms))
	for _, term := range terms {
		if term.Term == "" || seen[term.Term] {
			continue
		}
		seen[term.Term] = true
		result = append(result, term.Term)
	}
	return result
}
