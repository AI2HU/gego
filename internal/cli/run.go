package cli

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/AI2HU/gego/internal/logger"
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
