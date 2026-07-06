package shared

import (
	"bufio"
	"os"
	"regexp"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
)

var (
	exclusionWords     map[string]bool
	exclusionWordsList []string
	exclusionWordsMu   sync.RWMutex

	capitalizedWordRe = regexp.MustCompile(`\b[\p{Lu}][\p{L}]+(?:\s+[\p{Lu}][\p{L}]+)*\b`)
)

// ParseEnabledFilter parses the enabled query parameter and returns a pointer to bool or nil
func ParseEnabledFilter(c *gin.Context) *bool {
	enabledStr := c.Query("enabled")
	if enabledStr == "" {
		return nil
	}

	switch enabledStr {
	case "true":
		return &[]bool{true}[0]
	case "false":
		return &[]bool{false}[0]
	default:
		return nil
	}
}

// SetExclusionWords replaces the in-memory exclusion word cache.
func SetExclusionWords(words []string) {
	exclusionWordsMu.Lock()
	defer exclusionWordsMu.Unlock()

	exclusionWords = make(map[string]bool, len(words))
	exclusionWordsList = make([]string, 0, len(words))
	for _, word := range words {
		normalized := strings.TrimSpace(word)
		if normalized == "" {
			continue
		}
		key := strings.ToLower(normalized)
		if exclusionWords[key] {
			continue
		}
		exclusionWords[key] = true
		exclusionWordsList = append(exclusionWordsList, normalized)
	}
}

// ReloadExclusionWords reloads exclusion words from the provided list.
func ReloadExclusionWords(words []string) {
	SetExclusionWords(words)
}

// GetExclusionWordsList returns a list of all exclusion words.
func GetExclusionWordsList() []string {
	exclusionWordsMu.RLock()
	defer exclusionWordsMu.RUnlock()

	result := make([]string, len(exclusionWordsList))
	copy(result, exclusionWordsList)
	return result
}

func getExclusionWords() map[string]bool {
	exclusionWordsMu.RLock()
	defer exclusionWordsMu.RUnlock()
	return exclusionWords
}

func isExcluded(word string, exclusions map[string]bool) bool {
	if exclusions == nil {
		return false
	}
	return exclusions[strings.ToLower(word)]
}

// LoadExclusionWordsFromFile reads exclusion words from a legacy file path.
func LoadExclusionWordsFromFile(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	defer file.Close()

	var words []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" && !strings.HasPrefix(line, "#") {
			words = append(words, line)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return words, nil
}

// ExtractCapitalizedWords extracts words that start with a capital letter.
func ExtractCapitalizedWords(text string) []string {
	return filterCapitalizedWords(text, getExclusionWords(), true)
}

// ExtractAllCapitalizedWords extracts capitalized words without exclusions or brand resolution.
func ExtractAllCapitalizedWords(text string) []string {
	return filterCapitalizedWords(text, nil, false)
}

func filterCapitalizedWords(text string, exclusions map[string]bool, resolveBrands bool) []string {
	matches := capitalizedWordRe.FindAllString(text, -1)

	var filtered []string
	for _, word := range matches {
		if isExcluded(word, exclusions) {
			continue
		}
		if resolveBrands {
			word = ResolveBrand(word)
		}
		filtered = append(filtered, word)
	}
	return filtered
}

// CountOccurrences counts how many times a keyword appears in text (case-insensitive)
func CountOccurrences(text, keyword string) int {
	lower := strings.ToLower(text)
	lowerKeyword := strings.ToLower(keyword)
	count := 0
	index := 0

	for {
		i := strings.Index(lower[index:], lowerKeyword)
		if i == -1 {
			break
		}
		count++
		index += i + len(lowerKeyword)
	}

	return count
}
