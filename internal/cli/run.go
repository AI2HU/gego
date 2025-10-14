package cli

import (
	"bufio"
	"context"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/spf13/cobra"

	"github.com/AI2HU/gego/internal/logger"
	"github.com/AI2HU/gego/internal/models"
)

// boolPtr returns a pointer to a boolean value
func boolPtr(b bool) *bool {
	return &b
}

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run prompts with LLMs",
	Long:  `Run prompts with LLMs either via scheduler (scheduled execution) or run once (immediate execution of all prompts with all models).`,
	RunE:  runCommand,
}

func runCommand(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	reader := bufio.NewReader(os.Stdin)

	fmt.Printf("%s🚀 Run Prompts with LLMs%s\n", FormatHeader(""), Reset)
	fmt.Printf("%s========================%s\n", DimStyle, Reset)
	fmt.Println()

	// Initialize LLM providers
	if err := initializeLLMProviders(ctx); err != nil {
		return fmt.Errorf("failed to initialize LLM providers: %w", err)
	}

	// Check if there are any schedules
	schedules, err := database.ListSchedules(ctx, boolPtr(true))
	if err != nil {
		return fmt.Errorf("failed to check schedules: %w", err)
	}

	// Check if there are any prompts and LLMs
	prompts, err := database.ListPrompts(ctx, boolPtr(true))
	if err != nil {
		return fmt.Errorf("failed to check prompts: %w", err)
	}

	llms, err := database.ListLLMs(ctx, boolPtr(true))
	if err != nil {
		return fmt.Errorf("failed to check LLMs: %w", err)
	}

	// Show available options
	fmt.Printf("%sChoose how to run prompts:%s\n", LabelStyle, Reset)
	fmt.Printf("  %s1. Scheduler Mode%s - Run scheduled tasks continuously%s\n", CountStyle, Reset, DimStyle+" (requires schedules)"+Reset)
	fmt.Printf("  %s2. Run Once%s - Execute all prompts with all LLMs immediately%s\n", CountStyle, Reset, DimStyle+" (requires prompts and LLMs)"+Reset)
	fmt.Println()

	// Show current status
	if len(schedules) > 0 {
		fmt.Printf("%s📅 Found %s enabled schedule(s)%s\n", InfoStyle, FormatCount(len(schedules)), Reset)
	} else {
		fmt.Printf("%s📅 No enabled schedules found%s\n", WarningStyle, Reset)
	}

	if len(prompts) > 0 && len(llms) > 0 {
		fmt.Printf("%s📝 Found %s prompt(s) and %s LLM(s) for run-once mode%s\n", InfoStyle, FormatCount(len(prompts)), FormatCount(len(llms)), Reset)
	} else {
		fmt.Printf("%s📝 Run-once mode requires prompts and LLMs%s\n", WarningStyle, Reset)
		if len(prompts) == 0 {
			fmt.Printf("   %s• No enabled prompts found%s\n", DimStyle, Reset)
		}
		if len(llms) == 0 {
			fmt.Printf("   %s• No enabled LLMs found%s\n", DimStyle, Reset)
		}
	}
	fmt.Println()

	// Get user choice
	choice, err := promptWithRetry(reader, fmt.Sprintf("%sSelect mode (1 or 2): %s", LabelStyle, Reset), func(input string) (string, error) {
		switch input {
		case "1", "2":
			return input, nil
		default:
			return "", fmt.Errorf("invalid choice: %s (choose 1 or 2)", input)
		}
	})
	if err != nil {
		return err
	}

	if choice == "1" {
		return runSchedulerMode(ctx)
	} else {
		return runOnceMode(ctx)
	}
}

func runSchedulerMode(ctx context.Context) error {
	logger.Info("🚀 Starting Gego Scheduler")
	logger.Info("=========================")

	// Start scheduler
	if err := sched.Start(ctx); err != nil {
		return fmt.Errorf("failed to start scheduler: %w", err)
	}

	logger.Info("✅ Scheduler is running. Press Ctrl+C to stop.")

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	<-sigChan

	logger.Info("⏸️  Stopping scheduler...")
	sched.Stop()

	logger.Info("✅ Scheduler stopped. Goodbye!")
	return nil
}

func runOnceMode(ctx context.Context) error {
	// Get all enabled prompts and LLMs
	prompts, err := database.ListPrompts(ctx, boolPtr(true))
	if err != nil {
		return fmt.Errorf("failed to list prompts: %w", err)
	}

	llms, err := database.ListLLMs(ctx, boolPtr(true))
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

	// Get temperature for this run
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

			// Execute the prompt with the LLM
			if err := executePromptWithLLM(ctx, prompt, llm, currentTemperature); err != nil {
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

func executePromptWithLLM(ctx context.Context, prompt *models.Prompt, llm *models.LLMConfig, temperature float64) error {
	// Get the LLM provider
	provider, ok := llmRegistry.Get(llm.Provider)
	if !ok {
		return fmt.Errorf("LLM provider %s not found", llm.Provider)
	}

	const maxRetries = 3
	const retryDelay = 30 * time.Second

	var lastErr error
	for attempt := 1; attempt <= maxRetries; attempt++ {
		// Generate response
		response, err := provider.Generate(ctx, prompt.Template, map[string]interface{}{
			"model":       llm.Model,
			"temperature": temperature,
		})

		if err != nil {
			lastErr = fmt.Errorf("failed to generate response: %w", err)
			if attempt < maxRetries {
				fmt.Printf("%s❌ Attempt %d/%d failed: %s%s\n", WarningStyle, attempt, maxRetries, FormatValue(err.Error()), Reset)
				fmt.Printf("%s⏳ Waiting %ds before retry attempt %d...%s\n", InfoStyle, retryDelay/time.Second, attempt+1, Reset)
				time.Sleep(retryDelay)
				continue
			}
			return lastErr
		}

		if response.Error != "" {
			lastErr = fmt.Errorf("LLM error: %s", response.Error)
			if attempt < maxRetries {
				fmt.Printf("%s❌ Attempt %d/%d failed: %s%s\n", WarningStyle, attempt, maxRetries, FormatValue(response.Error), Reset)
				fmt.Printf("%s⏳ Waiting %ds before retry attempt %d...%s\n", InfoStyle, retryDelay/time.Second, attempt+1, Reset)
				time.Sleep(retryDelay)
				continue
			}
			return lastErr
		}

		// Success! Save response to database
		responseModel := &models.Response{
			ID:           uuid.New().String(),
			PromptID:     prompt.ID,
			LLMID:        llm.ID,
			PromptText:   prompt.Template,
			ResponseText: response.Text,
			LLMName:      llm.Name,
			LLMProvider:  llm.Provider,
			LLMModel:     llm.Model,
			Temperature:  temperature,
			TokensUsed:   response.TokensUsed,
			LatencyMs:    response.LatencyMs,
			CreatedAt:    time.Now(),
		}

		if err := database.CreateResponse(ctx, responseModel); err != nil {
			return fmt.Errorf("failed to save response: %w", err)
		}

		// Log success with retry info if applicable
		if attempt > 1 {
			fmt.Printf("%s✅ Prompt execution succeeded on attempt %d after %d previous failures%s\n", SuccessStyle, attempt, attempt-1, Reset)
		}

		return nil
	}

	return fmt.Errorf("all %d attempts failed. Last error: %w", maxRetries, lastErr)
}
