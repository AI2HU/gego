package cli

import (
	"bufio"
	"context"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/AI2HU/gego/internal/models"
	"github.com/AI2HU/gego/internal/services"
	"github.com/AI2HU/gego/internal/shared"
)

// boolPtr returns a pointer to a boolean value
func boolPtr(b bool) *bool {
	return &b
}

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run all prompts with all LLMs once",
	Long:  `Execute all enabled prompts with all enabled LLMs immediately. Use 'gego scheduler start' for scheduled execution.`,
	RunE:  runCommand,
}

func runCommand(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	if err := initializeLLMProviders(ctx); err != nil {
		return fmt.Errorf("failed to initialize LLM providers: %w", err)
	}

	return runOnceMode(ctx)
}

func runOnceMode(ctx context.Context) error {
	promptService := services.NewPromptManagementService(database)
	llmService := services.NewLLMService(database)

	allPrompts, err := promptService.GetEnabledPrompts(ctx)
	if err != nil {
		return fmt.Errorf("failed to list prompts: %w", err)
	}

	llms, err := llmService.GetEnabledLLMs(ctx)
	if err != nil {
		return fmt.Errorf("failed to list LLMs: %w", err)
	}

	if len(allPrompts) == 0 {
		return fmt.Errorf("no enabled prompts found")
	}

	if len(llms) == 0 {
		return fmt.Errorf("no enabled LLMs found")
	}

	reader := bufio.NewReader(os.Stdin)
	runNewOnly, err := promptRunMode(reader)
	if err != nil {
		return fmt.Errorf("failed to get run mode: %w", err)
	}

	var prompts []*models.Prompt
	if runNewOnly {
		prompts, err = filterNewPrompts(ctx, allPrompts)
		if err != nil {
			return fmt.Errorf("failed to filter new prompts: %w", err)
		}
		if len(prompts) == 0 {
			return fmt.Errorf("no new prompts found (all prompts have been run)")
		}
	} else {
		prompts = allPrompts
	}

	modeText := "all"
	if runNewOnly {
		modeText = "new"
	}
	fmt.Printf("%s🔄 Running %s prompts with all LLMs%s\n", InfoStyle, FormatValue(modeText), Reset)
	fmt.Printf("%s====================================%s\n", DimStyle, Reset)
	fmt.Printf("%sPrompts: %s%s\n", LabelStyle, FormatCount(len(prompts)), Reset)
	fmt.Printf("%sLLMs: %s%s\n", LabelStyle, FormatCount(len(llms)), Reset)
	fmt.Printf("%sTotal executions: %s%s\n", LabelStyle, FormatCount(len(prompts)*len(llms)), Reset)
	fmt.Println()

	temperature, err := promptTemperature(reader)
	if err != nil {
		return fmt.Errorf("failed to get temperature: %w", err)
	}

	totalExecutions := len(prompts) * len(llms)
	completedExecutions := 0

	for _, prompt := range prompts {
		currentTemperature := temperature
		if temperature == -1.0 { // random was selected
			rand.Seed(time.Now().UnixNano())
			currentTemperature = rand.Float64()
		}
		for _, llm := range llms {
			fmt.Printf("%s📝 Running prompt: %s%s\n", InfoStyle, FormatValue(prompt.Template), Reset)
			fmt.Printf("%s🤖 Using LLM: %s (%s)%s\n", InfoStyle, FormatValue(llm.Name), FormatSecondary(llm.Provider), Reset)
			fmt.Printf("%s🌡️  Using temperature: %s%s\n", InfoStyle, FormatValue(fmt.Sprintf("%.1f", currentTemperature)), Reset)

			executionService := services.NewExecutionService(database, llmRegistry)
			config := &services.ExecutionConfig{
				Temperature: currentTemperature,
				MaxRetries:  3,
				RetryDelay:  30 * time.Second,
			}

			_, err := executionService.ExecutePromptWithLLM(ctx, prompt, llm, config)
			if err != nil {
				fmt.Printf("%s❌ Failed: %s%s\n", ErrorStyle, FormatValue(err.Error()), Reset)
			} else {
				fmt.Printf("%s✅ Success%s\n", SuccessStyle, Reset)
			}

			completedExecutions++
			fmt.Printf("%sProgress: %s/%s%s\n", DimStyle, FormatCount(completedExecutions), FormatCount(totalExecutions), Reset)
			fmt.Println()
		}
	}

	fmt.Printf("%s🎉 Completed all executions!%s\n", SuccessStyle, Reset)
	return nil
}

// promptRunMode prompts the user to choose between running new prompts or all prompts
func promptRunMode(reader *bufio.Reader) (bool, error) {
	fmt.Printf("%s📋 Run Mode Selection%s\n", LabelStyle, Reset)
	fmt.Printf("%sChoose which prompts to run:%s\n", DimStyle, Reset)
	fmt.Printf("  %s• new: Run only prompts that haven't been run yet%s\n", DimStyle, Reset)
	fmt.Printf("  %s• all: Run all prompts (including already run ones)%s\n", DimStyle, Reset)
	fmt.Println()

	result, err := promptWithRetry(reader, fmt.Sprintf("%sEnter run mode (new/all) [all]: %s", LabelStyle, Reset), func(input string) (string, error) {
		if input == "" {
			return "all", nil
		}

		lower := strings.ToLower(input)
		if lower == "new" || lower == "all" {
			return lower, nil
		}

		return "", fmt.Errorf("invalid input: %s (enter 'new' or 'all')", input)
	})

	if err != nil {
		return false, err
	}

	return result == "new", nil
}

// filterNewPrompts filters prompts to only include those that haven't been run yet
func filterNewPrompts(ctx context.Context, prompts []*models.Prompt) ([]*models.Prompt, error) {
	executionService := services.NewExecutionService(database, llmRegistry)
	var newPrompts []*models.Prompt

	for _, prompt := range prompts {
		responses, err := executionService.ListResponses(ctx, shared.ResponseFilter{
			PromptID: prompt.ID,
			Limit:    1,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to check responses for prompt %s: %w", prompt.ID, err)
		}

		if len(responses) == 0 {
			newPrompts = append(newPrompts, prompt)
		}
	}

	return newPrompts, nil
}
