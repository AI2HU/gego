package cli

import (
	"bufio"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"
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

// promptOptional prompts for optional input with default value
func promptOptional(reader *bufio.Reader, prompt string, defaultValue string) (string, error) {
	return promptWithRetry(reader, prompt, func(input string) (string, error) {
		if input == "" {
			return defaultValue, nil
		}
		return input, nil
	})
}

// promptTemperature prompts for temperature input with validation
func promptTemperature(reader *bufio.Reader) (float64, error) {
	fmt.Printf("%s🌡️  Temperature Control%s\n", LabelStyle, Reset)
	fmt.Printf("%sTemperature controls the randomness of LLM responses:%s\n", DimStyle, Reset)
	fmt.Printf("  %s• 0.0: Very deterministic, consistent responses%s\n", DimStyle, Reset)
	fmt.Printf("  %s• 0.7: Balanced creativity and consistency (default)%s\n", DimStyle, Reset)
	fmt.Printf("  %s• 1.0: Very creative, diverse responses%s\n", DimStyle, Reset)
	fmt.Printf("  %s• 'random': Random temperature between 0.0 and 1.0%s\n", DimStyle, Reset)
	fmt.Println()

	result, err := promptWithRetry(reader, fmt.Sprintf("%sEnter temperature (0.0-1.0) or 'random' [0.7]: %s", LabelStyle, Reset), func(input string) (string, error) {
		if input == "" {
			return "0.7", nil
		}

		if strings.ToLower(input) == "random" {
			return "random", nil
		}

		temp, err := strconv.ParseFloat(input, 64)
		if err != nil {
			return "", fmt.Errorf("invalid temperature: %s (must be a number between 0.0 and 1.0 or 'random')", input)
		}

		if temp < 0.0 || temp > 1.0 {
			return "", fmt.Errorf("temperature must be between 0.0 and 1.0, got: %.2f", temp)
		}

		return input, nil
	})

	if err != nil {
		return 0, err
	}

	if result == "random" {
		rand.Seed(time.Now().UnixNano())
		return -1.0, nil // Special value to indicate random temperature
	}

	return strconv.ParseFloat(result, 64)
}
