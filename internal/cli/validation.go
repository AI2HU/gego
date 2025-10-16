package cli

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// validateSelection validates comma-separated selection input
func validateSelection(input string, maxCount int) ([]int, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return nil, fmt.Errorf("selection is required")
	}

	if strings.ToLower(input) == "all" {
		var all []int
		for i := 1; i <= maxCount; i++ {
			all = append(all, i)
		}
		return all, nil
	}

	var selections []int
	parts := strings.Split(input, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		num, err := strconv.Atoi(part)
		if err != nil {
			return nil, fmt.Errorf("invalid selection: %s (must be numbers 1-%d or 'all')", part, maxCount)
		}
		if num < 1 || num > maxCount {
			return nil, fmt.Errorf("invalid selection: %d (must be between 1 and %d)", num, maxCount)
		}
		selections = append(selections, num)
	}

	return selections, nil
}

// validateLanguageCode validates language code input
func validateLanguageCode(input string) (string, error) {
	input = strings.ToUpper(strings.TrimSpace(input))
	if input == "" {
		return "", fmt.Errorf("language code is required")
	}
	if len(input) < 2 || len(input) > 3 {
		return "", fmt.Errorf("language code should be 2-3 characters (e.g., FR, EN, IT)")
	}
	return input, nil
}

// validateCronExpression validates cron expression input
func validateCronExpression(input string) (string, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return "", fmt.Errorf("cron expression is required")
	}

	// Basic validation - check if it has 5 parts
	parts := strings.Fields(input)
	if len(parts) != 5 {
		return "", fmt.Errorf("invalid cron expression: %s (must have 5 parts)", input)
	}

	return input, nil
}

// validateAPIKey validates API key input
func validateAPIKey(input string, provider string) (string, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return "", fmt.Errorf("API key is required for %s", provider)
	}
	if len(input) < 10 {
		return "", fmt.Errorf("API key seems too short")
	}
	return input, nil
}

// validateBaseURL validates base URL input
func validateBaseURL(input string) (string, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return "http://localhost:11434", nil // Default for Ollama
	}
	if !strings.HasPrefix(input, "http://") && !strings.HasPrefix(input, "https://") {
		return "", fmt.Errorf("base URL must start with http:// or https://")
	}
	return input, nil
}

// validateNumber validates numeric input within a range
func validateNumber(input string, min, max int) (int, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return min, nil
	}

	num, err := strconv.Atoi(input)
	if err != nil {
		return 0, fmt.Errorf("invalid number: %s (enter a positive integer)", input)
	}

	if num < min || num > max {
		return 0, fmt.Errorf("number must be between %d and %d, got: %d", min, max, num)
	}

	return num, nil
}

// maskSensitiveData masks sensitive data for display
func maskSensitiveData(data string, maskChar string) string {
	if data == "" {
		return "(not set)"
	}
	if len(data) <= 8 {
		return strings.Repeat(maskChar, 3)
	}
	return data[:4] + "..." + data[len(data)-4:]
}

// formatDuration formats a duration for display
func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%.1fs", d.Seconds())
	}
	if d < time.Hour {
		return fmt.Sprintf("%.1fm", d.Minutes())
	}
	return fmt.Sprintf("%.1fh", d.Hours())
}

// formatCount formats a count for display
func formatCount(count int) string {
	if count < 1000 {
		return fmt.Sprintf("%d", count)
	}
	if count < 1000000 {
		return fmt.Sprintf("%.1fK", float64(count)/1000)
	}
	return fmt.Sprintf("%.1fM", float64(count)/1000000)
}
