package cli

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Start the scheduler daemon",
	Long:  `Start the scheduler daemon that will execute schedules based on their cron expressions.`,
	RunE:  runScheduler,
}

func runScheduler(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Initialize LLM providers
	if err := initializeLLMProviders(ctx); err != nil {
		return fmt.Errorf("failed to initialize LLM providers: %w", err)
	}

	fmt.Printf("%s🚀 Starting Gego Scheduler%s\n", FormatHeader(""), Reset)
	fmt.Printf("%s=========================%s\n", DimStyle, Reset)
	fmt.Println()

	// Start scheduler
	if err := sched.Start(ctx); err != nil {
		return fmt.Errorf("failed to start scheduler: %w", err)
	}

	fmt.Printf("%s✅ Scheduler is running. Press Ctrl+C to stop.%s\n", SuccessStyle, Reset)
	fmt.Println()

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	<-sigChan

	fmt.Printf("\n%s⏸️  Stopping scheduler...%s\n", WarningStyle, Reset)
	sched.Stop()

	fmt.Printf("%s✅ Scheduler stopped. Goodbye!%s\n", SuccessStyle, Reset)
	return nil
}
