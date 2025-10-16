package cli

import (
	"bufio"
	"context"
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/AI2HU/gego/internal/services"
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

	prompts, err := promptService.GetEnabledPrompts(ctx)
	if err != nil {
		return fmt.Errorf("failed to list prompts: %w", err)
	}

	llms, err := llmService.GetEnabledLLMs(ctx)
	if err != nil {
		return fmt.Errorf("failed to list LLMs: %w", err)
	}

	if len(prompts) == 0 {
		return fmt.Errorf("no enabled prompts found")
	}

	if len(llms) == 0 {
		return fmt.Errorf("no enabled LLMs found")
	}

	fmt.Printf("%s🔄 Running all prompts with all LLMs%s\n", InfoStyle, Reset)
	fmt.Printf("%s====================================%s\n", DimStyle, Reset)
	fmt.Printf("%sPrompts: %s%s\n", LabelStyle, FormatCount(len(prompts)), Reset)
	fmt.Printf("%sLLMs: %s%s\n", LabelStyle, FormatCount(len(llms)), Reset)
	fmt.Printf("%sTotal executions: %s%s\n", LabelStyle, FormatCount(len(prompts)*len(llms)), Reset)
	fmt.Println()

	reader := bufio.NewReader(os.Stdin)
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
