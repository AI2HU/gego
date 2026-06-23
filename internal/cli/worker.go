package cli

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/AI2HU/gego/internal/logger"
	"github.com/AI2HU/gego/internal/services"
)

var workerConcurrency int

var workerCmd = &cobra.Command{
	Use:   "worker",
	Short: "Manage background workers",
}

var workerStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start a schedule worker",
	RunE:  runWorkerStart,
}

func init() {
	workerStartCmd.Flags().IntVar(&workerConcurrency, "concurrency", 0, "Worker concurrency override")
	workerCmd.AddCommand(workerStartCmd)
}

func runWorkerStart(cmd *cobra.Command, args []string) error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	if err := initializeLLMProviders(ctx); err != nil {
		return fmt.Errorf("failed to initialize LLM providers: %w", err)
	}

	if runtime != nil {
		cfg := runtime.Config
		if workerConcurrency > 0 {
			cfg.WorkerConcurrency = workerConcurrency
		}
		worker := runtime.NewWorkerService(database)
		fmt.Printf("%sStarting worker (concurrency=%d)...%s\n", InfoStyle, cfg.WorkerConcurrency, Reset)
		if err := worker.Start(ctx); err != nil && ctx.Err() == nil {
			return fmt.Errorf("worker stopped: %w", err)
		}
		logger.Info("Worker stopped")
		return nil
	}

	runtime, err := services.NewRuntime(database, llmRegistry)
	if err != nil {
		return err
	}
	defer runtime.Close()

	cfg := runtime.Config
	if workerConcurrency > 0 {
		cfg.WorkerConcurrency = workerConcurrency
	}

	worker := runtime.NewWorkerService(database)
	fmt.Printf("%sStarting worker (concurrency=%d)...%s\n", InfoStyle, cfg.WorkerConcurrency, Reset)

	if err := worker.Start(ctx); err != nil && ctx.Err() == nil {
		return fmt.Errorf("worker stopped: %w", err)
	}

	logger.Info("Worker stopped")
	return nil
}
