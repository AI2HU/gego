package shared

import (
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
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

// ExtractCapitalizedWords extracts words that start with a capital letter
func ExtractCapitalizedWords(text string) []string {
	re := regexp.MustCompile(`\b[A-Z][a-zA-Z]+(?:\s+[A-Z][a-zA-Z]+)*\b`)
	matches := re.FindAllString(text, -1)

	// Filter common words that can be confused with brand names
	var filtered []string
	commonWords := map[string]bool{
		"The": true, "A": true, "An": true, "And": true, "Or": true,
		"But": true, "In": true, "On": true, "At": true, "To": true,
		"For": true, "Of": true, "With": true, "By": true, "From": true,
		"This": true, "That": true, "These": true, "Those": true,
		"I": true, "You": true, "He": true, "She": true, "It": true,
		"We": true, "They": true, "My": true, "Your": true, "His": true,
		"Her": true, "Its": true, "Our": true, "Their": true,
		"If": true, "While": true, "AI": true, "What": true, "CRM": true, "Here": true,
		"URL": true,
	}

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
