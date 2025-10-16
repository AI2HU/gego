package cli

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
)

var schedulerCmd = &cobra.Command{
	Use:   "scheduler",
	Short: "Manage the scheduler",
	Long:  `Manage the Gego scheduler - start and monitor schedules.`,
}

var schedulerStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the scheduler",
	RunE:  runSchedulerStart,
}

func init() {
	schedulerCmd.AddCommand(schedulerStartCmd)
}

func runSchedulerStart(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	fmt.Printf("%s🚀 Start Scheduler%s\n", FormatHeader(""), Reset)
	fmt.Printf("%s================%s\n", DimStyle, Reset)
	fmt.Println()

	if err := initializeLLMProviders(ctx); err != nil {
		return fmt.Errorf("failed to initialize LLM providers: %w", err)
	}

	schedules, err := database.ListSchedules(ctx, boolPtr(true))
	if err != nil {
		return fmt.Errorf("failed to check schedules: %w", err)
	}

	if len(schedules) == 0 {
		fmt.Printf("%s❌ No enabled schedules found%s\n", ErrorStyle, Reset)
		fmt.Printf("%s💡 Use 'gego schedule add' to create schedules%s\n", InfoStyle, Reset)
		return nil
	}

	fmt.Printf("%sStarting Schedules:%s\n", LabelStyle, Reset)
	for i, schedule := range schedules {
		fmt.Printf("  %s%d. %s%s\n", CountStyle, i+1, Reset, FormatValue(schedule.Name))
		fmt.Printf("     %sID: %s | Cron: %s%s\n", DimStyle, schedule.ID, schedule.CronExpr, Reset)
	}
	fmt.Println()

	if err := sched.Start(ctx); err != nil {
		return fmt.Errorf("failed to start scheduler: %w", err)
	}

	fmt.Printf("%s✅ All schedules started successfully%s\n", SuccessStyle, Reset)
	fmt.Printf("%s📅 Running %s schedule(s)%s\n", InfoStyle, FormatCount(len(schedules)), Reset)
	fmt.Printf("%s🔄 Scheduler is now monitoring schedules%s\n", InfoStyle, Reset)
	fmt.Printf("%s📝 Press Ctrl+C to stop the scheduler%s\n", InfoStyle, Reset)
	fmt.Println()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	<-c
	fmt.Printf("\n%s⏹️  Stopping scheduler...%s\n", InfoStyle, Reset)
	sched.Stop()
	fmt.Printf("%s✅ Scheduler stopped%s\n", SuccessStyle, Reset)

	return nil
}
