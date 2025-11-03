package shared

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/AI2HU/gego/internal/config"
	"github.com/gin-gonic/gin"
)

var (
	exclusionWords       map[string]bool
	exclusionWordsOnce   sync.Once
	exclusionWordsMu     sync.RWMutex
	exclusionFileModTime time.Time
	exclusionFilePath    string
	exclusionFilePathMu  sync.RWMutex
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

// getExclusionWords loads exclusion words from file or returns default words
// It automatically reloads if the file has been modified since last load
func getExclusionWords() map[string]bool {
	exclusionWordsOnce.Do(func() {
		exclusionWordsMu.Lock()
		exclusionWords, exclusionFileModTime = loadExclusionWordsFromFileWithModTime()
		exclusionWordsMu.Unlock()
	})

	exclusionFile := getExclusionFilePath()
	fileInfo, err := os.Stat(exclusionFile)
	if err == nil && !fileInfo.ModTime().IsZero() && fileInfo.ModTime().After(exclusionFileModTime) {
		exclusionWordsMu.Lock()
		exclusionWords, exclusionFileModTime = loadExclusionWordsFromFileWithModTime()
		exclusionWordsMu.Unlock()
	}

	exclusionWordsMu.RLock()
	defer exclusionWordsMu.RUnlock()
	return exclusionWords
}

// ReloadExclusionWords reloads the exclusion words from file
func ReloadExclusionWords() error {
	exclusionWordsMu.Lock()
	defer exclusionWordsMu.Unlock()
	exclusionWords, exclusionFileModTime = loadExclusionWordsFromFileWithModTime()
	return nil
}

// GetExclusionWordsList returns a list of all exclusion words (for debugging/inspection)
func GetExclusionWordsList() []string {
	words := getExclusionWords()
	result := make([]string, 0, len(words))
	for word := range words {
		result = append(result, word)
	}
	return result
}

// loadExclusionWordsFromFileWithModTime loads exclusion words and their modification time
func loadExclusionWordsFromFileWithModTime() (map[string]bool, time.Time) {
	exclusionFile := getExclusionFilePath()

	words := make(map[string]bool)
	var modTime time.Time

	fileInfo, err := os.Stat(exclusionFile)
	if err != nil {
		return words, modTime
	}
	modTime = fileInfo.ModTime()

	file, err := os.Open(exclusionFile)
	if err != nil {
		return words, modTime
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" && !strings.HasPrefix(line, "#") {
			words[line] = true
		}
	}

	if err := scanner.Err(); err != nil {
		return make(map[string]bool), modTime
	}

	return words, modTime
}

// SetExclusionFilePath sets the path to the keywords_exclusion file from config
func SetExclusionFilePath(path string) {
	exclusionFilePathMu.Lock()
	defer exclusionFilePathMu.Unlock()
	exclusionFilePath = path
}

// getExclusionFilePath returns the path to the keywords_exclusion file
func getExclusionFilePath() string {
	exclusionFilePathMu.RLock()
	path := exclusionFilePath
	exclusionFilePathMu.RUnlock()

	if path != "" {
		return path
	}

	var configPath string
	if envPath := os.Getenv("GEGO_CONFIG_PATH"); envPath != "" {
		configPath = envPath
	} else {
		configPath = config.GetConfigPath()
	}
	configDir := filepath.Dir(configPath)
	return filepath.Join(configDir, "keywords_exclusion")
}

// GetExclusionFilePath returns the path to the keywords_exclusion file (exported for CLI)
func GetExclusionFilePath() string {
	return getExclusionFilePath()
}

// ExtractCapitalizedWords extracts words that start with a capital letter
func ExtractCapitalizedWords(text string) []string {
	re := regexp.MustCompile(`\b[A-Z][a-zA-Z]+(?:\s+[A-Z][a-zA-Z]+)*\b`)
	matches := re.FindAllString(text, -1)

	// Filter common words that can be confused with brand names
	var filtered []string
	commonWords := getExclusionWords()

	for _, word := range matches {
		if !commonWords[word] {
			filtered = append(filtered, word)
		}
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
