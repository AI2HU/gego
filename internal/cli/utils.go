package cli

import (
	"bufio"
	"fmt"
	"strings"
)

// promptWithRetry prompts the user for input and retries on invalid input
func promptWithRetry(reader *bufio.Reader, prompt string, validator func(string) (string, error)) (string, error) {
	for {
		fmt.Print(prompt)
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		result, err := validator(input)
		if err == nil {
			return result, nil
		}

		fmt.Printf("❌ %s\n\n", err.Error())
	}
}

// promptYesNo prompts for yes/no input with retry
func promptYesNo(reader *bufio.Reader, prompt string) (bool, error) {
	result, err := promptWithRetry(reader, prompt, func(input string) (string, error) {
		lower := strings.ToLower(input)
		if lower == "y" || lower == "yes" || lower == "n" || lower == "no" || lower == "" {
			return lower, nil
		}
		return "", fmt.Errorf("invalid input: %s (enter y/yes/n/no or press Enter for no)", input)
	})
	if err != nil {
		return false, err
	}

	return result == "y" || result == "yes", nil
}

// promptRequired prompts for required input with retry
func promptRequired(reader *bufio.Reader, prompt string) (string, error) {
	return promptWithRetry(reader, prompt, func(input string) (string, error) {
		if input == "" {
			return "", fmt.Errorf("this field is required")
		}
		return input, nil
	})
}

// promptOptional prompts for optional input with default value
func promptOptional(reader *bufio.Reader, prompt string, defaultValue string) (string, error) {
	return promptWithRetry(reader, prompt, func(input string) (string, error) {
		if input == "" {
			return defaultValue, nil
		}
		return input, nil
	})
}
