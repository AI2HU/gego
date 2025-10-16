package cli

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/google/uuid"
	"github.com/spf13/cobra"

	"github.com/AI2HU/gego/internal/models"
)

var scheduleCmd = &cobra.Command{
	Use:   "schedule",
	Short: "Manage schedules",
	Long:  `Add, list, update, and delete execution schedules.`,
}

var scheduleAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a new schedule",
	RunE:  runScheduleAdd,
}

var scheduleListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all schedules",
	RunE:  runScheduleList,
}

var scheduleGetCmd = &cobra.Command{
	Use:   "get [id]",
	Short: "Get details of a schedule",
	Args:  cobra.ExactArgs(1),
	RunE:  runScheduleGet,
}

var scheduleDeleteCmd = &cobra.Command{
	Use:   "delete [id]",
	Short: "Delete a schedule or all schedules",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runScheduleDelete,
}

var scheduleEnableCmd = &cobra.Command{
	Use:   "enable [id]",
	Short: "Enable a schedule",
	Args:  cobra.ExactArgs(1),
	RunE:  runScheduleEnable,
}

var scheduleDisableCmd = &cobra.Command{
	Use:   "disable [id]",
	Short: "Disable a schedule",
	Args:  cobra.ExactArgs(1),
	RunE:  runScheduleDisable,
}

var scheduleRunCmd = &cobra.Command{
	Use:   "run [id]",
	Short: "Run a schedule immediately",
	Args:  cobra.ExactArgs(1),
	RunE:  runScheduleRun,
}

func init() {
	scheduleCmd.AddCommand(scheduleAddCmd)
	scheduleCmd.AddCommand(scheduleListCmd)
	scheduleCmd.AddCommand(scheduleGetCmd)
	scheduleCmd.AddCommand(scheduleDeleteCmd)
	scheduleCmd.AddCommand(scheduleEnableCmd)
	scheduleCmd.AddCommand(scheduleDisableCmd)
	scheduleCmd.AddCommand(scheduleRunCmd)
}

func runScheduleAdd(cmd *cobra.Command, args []string) error {
	reader := bufio.NewReader(os.Stdin)
	ctx := context.Background()

	fmt.Printf("%s➕ Add New Schedule%s\n", FormatHeader(""), Reset)
	fmt.Printf("%s==================%s\n", DimStyle, Reset)
	fmt.Println()

	schedule := &models.Schedule{
		ID:      uuid.New().String(),
		Enabled: true,
	}

	fmt.Printf("%sName: %s", LabelStyle, Reset)
	name, _ := reader.ReadString('\n')
	schedule.Name = strings.TrimSpace(name)

	fmt.Printf("\n%sAvailable Prompts:%s\n", LabelStyle, Reset)
	prompts, err := database.ListPrompts(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to list prompts: %w", err)
	}

	if len(prompts) == 0 {
		return fmt.Errorf("no prompts available. Create prompts first with 'gego prompt add'")
	}

	for i, p := range prompts {
		fmt.Printf("  %s%d. %s%s\n", CountStyle, i+1, Reset, FormatValue(p.Template))
	}

	fmt.Printf("\n%sSelect prompts (comma-separated numbers or 'all'): %s", LabelStyle, Reset)
	promptSelection, _ := reader.ReadString('\n')
	promptSelection = strings.TrimSpace(promptSelection)

	if promptSelection == "all" {
		for _, p := range prompts {
			schedule.PromptIDs = append(schedule.PromptIDs, p.ID)
		}
	} else {
		selections := strings.Split(promptSelection, ",")
		for _, sel := range selections {
			sel = strings.TrimSpace(sel)
			var idx int
			fmt.Sscanf(sel, "%d", &idx)
			if idx > 0 && idx <= len(prompts) {
				schedule.PromptIDs = append(schedule.PromptIDs, prompts[idx-1].ID)
			}
		}
	}

	fmt.Printf("\n%sAvailable LLMs:%s\n", LabelStyle, Reset)
	llms, err := database.ListLLMs(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to list LLMs: %w", err)
	}

	if len(llms) == 0 {
		return fmt.Errorf("no LLMs available. Create LLMs first with 'gego llm add'")
	}

	for i, l := range llms {
		fmt.Printf("  %s%d. %s (%s - %s)%s\n", CountStyle, i+1, FormatValue(l.Name), FormatSecondary(l.Provider), FormatSecondary(l.Model), Reset)
	}

	fmt.Printf("\n%sSelect LLMs (comma-separated numbers or 'all'): %s", LabelStyle, Reset)
	llmSelection, _ := reader.ReadString('\n')
	llmSelection = strings.TrimSpace(llmSelection)

	if llmSelection == "all" {
		for _, l := range llms {
			schedule.LLMIDs = append(schedule.LLMIDs, l.ID)
		}
	} else {
		selections := strings.Split(llmSelection, ",")
		for _, sel := range selections {
			sel = strings.TrimSpace(sel)
			var idx int
			fmt.Sscanf(sel, "%d", &idx)
			if idx > 0 && idx <= len(llms) {
				schedule.LLMIDs = append(schedule.LLMIDs, llms[idx-1].ID)
			}
		}
	}

	fmt.Printf("\n%sSchedule Frequency:%s\n", LabelStyle, Reset)
	fmt.Printf("  %s1. Every day%s\n", CountStyle, Reset)
	fmt.Printf("  %s2. Every week%s\n", CountStyle, Reset)
	fmt.Printf("  %s3. Every month%s\n", CountStyle, Reset)
	fmt.Printf("  %s4. Custom%s\n", CountStyle, Reset)

	cronChoice, err := promptWithRetry(reader, fmt.Sprintf("\n%sSelect frequency (1-4): %s", LabelStyle, Reset), func(input string) (string, error) {
		switch input {
		case "1", "2", "3", "4":
			return input, nil
		default:
			return "", fmt.Errorf("invalid choice: %s (choose 1-4)", input)
		}
	})
	if err != nil {
		return err
	}

	var cronExpr string
	switch cronChoice {
	case "1":
		cronExpr = "0 9 * * *"
		fmt.Printf("%sSelected: Every day%s\n", SuccessStyle, Reset)
	case "2":
		cronExpr = "0 9 * * MON"
		fmt.Printf("%sSelected: Every week%s\n", SuccessStyle, Reset)
	case "3":
		cronExpr = "0 9 1 * *"
		fmt.Printf("%sSelected: Every month%s\n", SuccessStyle, Reset)
	case "4":
		fmt.Printf("\n%sCron Expression Examples:%s\n", LabelStyle, Reset)
		fmt.Printf("  %s*/15 * * * *%s    - Every 15 minutes\n", FormatSecondary(""), Reset)
		fmt.Printf("  %s0 9 * * *%s       - Every day at 9am\n", FormatSecondary(""), Reset)
		fmt.Printf("  %s0 9 * * MON%s     - Every Monday at 9am\n", FormatSecondary(""), Reset)
		fmt.Printf("  %s0 0 1 * *%s       - First day of every month\n", FormatSecondary(""), Reset)
		customCron, err := promptWithRetry(reader, fmt.Sprintf("\n%sEnter custom cron expression: %s", LabelStyle, Reset), func(input string) (string, error) {
			if input == "" {
				return "", fmt.Errorf("cron expression is required")
			}
			return input, nil
		})
		if err != nil {
			return err
		}
		cronExpr = customCron
	}

	schedule.CronExpr = cronExpr

	temperature, err := promptTemperature(reader)
	if err != nil {
		return fmt.Errorf("failed to get temperature: %w", err)
	}
	schedule.Temperature = temperature

	if err := database.CreateSchedule(ctx, schedule); err != nil {
		return fmt.Errorf("failed to create schedule: %w", err)
	}

	fmt.Printf("\n%s✅ Schedule added successfully!%s\n", SuccessStyle, Reset)
	fmt.Printf("%sID: %s\n", LabelStyle, FormatSecondary(schedule.ID))
	fmt.Printf("%sPrompts: %s\n", LabelStyle, FormatCount(len(schedule.PromptIDs)))
	fmt.Printf("%sLLMs: %s\n", LabelStyle, FormatCount(len(schedule.LLMIDs)))
	fmt.Printf("%sTemperature: %s\n", LabelStyle, FormatValue(fmt.Sprintf("%.1f", schedule.Temperature)))
	fmt.Printf("\n%sRestart the scheduler to apply changes: %s%s\n", InfoStyle, FormatSecondary("gego scheduler start"), Reset)

	return nil
}

func runScheduleList(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	schedules, err := database.ListSchedules(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to list schedules: %w", err)
	}

	if len(schedules) == 0 {
		fmt.Printf("%sNo schedules configured. Use '%s' to add one.%s\n", WarningStyle, FormatSecondary("gego schedule add"), Reset)
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "%sID\tNAME\tCRON\tPROMPTS\tLLMs\tTEMP\tLAST RUN\tENABLED%s\n", LabelStyle, Reset)
	fmt.Fprintf(w, "%s──\t────\t────\t───────\t────\t────\t────────\t───────%s\n", DimStyle, Reset)

	for _, schedule := range schedules {
		enabled := "Yes"
		if !schedule.Enabled {
			enabled = "No"
		}

		lastRun := "Never"
		if schedule.LastRun != nil {
			lastRun = schedule.LastRun.Format("01-02 15:04")
		}

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			FormatSecondary(schedule.ID),
			FormatValue(schedule.Name),
			FormatSecondary(schedule.CronExpr),
			FormatCount(len(schedule.PromptIDs)),
			FormatCount(len(schedule.LLMIDs)),
			FormatValue(fmt.Sprintf("%.1f", schedule.Temperature)),
			FormatMeta(lastRun),
			FormatValue(enabled),
		)
	}

	w.Flush()
	fmt.Printf("\n%sTotal: %s schedules%s\n", InfoStyle, FormatCount(len(schedules)), Reset)

	return nil
}

func runScheduleGet(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	id := args[0]

	schedule, err := database.GetSchedule(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get schedule: %w", err)
	}

	fmt.Printf("%sSchedule Details%s\n", FormatHeader(""), Reset)
	fmt.Printf("%s================%s\n", DimStyle, Reset)
	fmt.Printf("%sID: %s\n", LabelStyle, FormatSecondary(schedule.ID))
	fmt.Printf("%sName: %s\n", LabelStyle, FormatValue(schedule.Name))
	fmt.Printf("%sCron Expression: %s\n", LabelStyle, FormatSecondary(schedule.CronExpr))
	fmt.Printf("%sEnabled: %s\n", LabelStyle, FormatValue(fmt.Sprintf("%v", schedule.Enabled)))
	fmt.Printf("%sCreated: %s\n", LabelStyle, FormatMeta(schedule.CreatedAt.Format(time.RFC3339)))
	fmt.Printf("%sUpdated: %s\n", LabelStyle, FormatMeta(schedule.UpdatedAt.Format(time.RFC3339)))

	if schedule.LastRun != nil {
		fmt.Printf("%sLast Run: %s\n", LabelStyle, FormatMeta(schedule.LastRun.Format(time.RFC3339)))
	}
	if schedule.NextRun != nil {
		fmt.Printf("%sNext Run: %s\n", LabelStyle, FormatMeta(schedule.NextRun.Format(time.RFC3339)))
	}

	fmt.Printf("\n%sPrompts (%s):%s\n", SuccessStyle, FormatCount(len(schedule.PromptIDs)), Reset)
	for _, promptID := range schedule.PromptIDs {
		prompt, err := database.GetPrompt(ctx, promptID)
		if err != nil {
			fmt.Printf("  - %s (error: %s)\n", FormatValue(promptID), FormatValue(err.Error()))
		} else {
			template := prompt.Template
			if len(template) > 50 {
				template = template[:47] + "..."
			}
			fmt.Printf("  - %s\n", FormatValue(template))
		}
	}

	fmt.Printf("\n%sLLMs (%s):%s\n", SuccessStyle, FormatCount(len(schedule.LLMIDs)), Reset)
	for _, llmID := range schedule.LLMIDs {
		llm, err := database.GetLLM(ctx, llmID)
		if err != nil {
			fmt.Printf("  - %s (error: %s)\n", FormatValue(llmID), FormatValue(err.Error()))
		} else {
			fmt.Printf("  - %s (%s - %s)\n", FormatValue(llm.Name), FormatSecondary(llm.Provider), FormatSecondary(llm.Model))
		}
	}

	return nil
}

func runScheduleDelete(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	reader := bufio.NewReader(os.Stdin)

	if len(args) == 0 {
		schedules, err := database.ListSchedules(ctx, nil)
		if err != nil {
			return fmt.Errorf("failed to list schedules: %w", err)
		}

		if len(schedules) == 0 {
			fmt.Printf("%sNo schedules found to delete.%s\n", WarningStyle, Reset)
			return nil
		}

		fmt.Printf("%sFound %s schedule(s):%s\n", InfoStyle, FormatCount(len(schedules)), Reset)
		for _, schedule := range schedules {
			fmt.Printf("  - %s (%s)\n", FormatValue(schedule.Name), FormatSecondary(schedule.ID))
		}
		fmt.Println()

		fmt.Printf("%sDo you want to delete ALL schedules? (y/N): %s", ErrorStyle, Reset)
		response, _ := reader.ReadString('\n')
		response = strings.TrimSpace(strings.ToLower(response))

		if response != "y" && response != "yes" {
			fmt.Printf("%sCancelled.%s\n", WarningStyle, Reset)
			return nil
		}

		deletedCount, err := database.DeleteAllSchedules(ctx)
		if err != nil {
			return fmt.Errorf("failed to delete all schedules: %w", err)
		}

		fmt.Printf("%s✅ Successfully deleted %s schedules!%s\n", SuccessStyle, FormatCount(deletedCount), Reset)
		fmt.Printf("%sRestart the scheduler to apply changes: %s%s\n", InfoStyle, FormatSecondary("gego scheduler start"), Reset)
		return nil
	}

	id := args[0]
	fmt.Printf("%sAre you sure you want to delete schedule %s? (y/N): %s", ErrorStyle, FormatValue(id), Reset)
	response, _ := reader.ReadString('\n')
	response = strings.TrimSpace(strings.ToLower(response))

	if response != "y" && response != "yes" {
		fmt.Printf("%sCancelled.%s\n", WarningStyle, Reset)
		return nil
	}

	if err := database.DeleteSchedule(ctx, id); err != nil {
		return fmt.Errorf("failed to delete schedule: %w", err)
	}

	fmt.Printf("%s✅ Schedule deleted successfully!%s\n", SuccessStyle, Reset)
	fmt.Printf("%sRestart the scheduler to apply changes: %s%s\n", InfoStyle, FormatSecondary("gego scheduler start"), Reset)
	return nil
}

func runScheduleEnable(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	id := args[0]

	schedule, err := database.GetSchedule(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get schedule: %w", err)
	}

	schedule.Enabled = true
	if err := database.UpdateSchedule(ctx, schedule); err != nil {
		return fmt.Errorf("failed to update schedule: %w", err)
	}

	fmt.Printf("%s✅ Schedule enabled!%s\n", SuccessStyle, Reset)
	fmt.Printf("%sRestart the scheduler to apply changes: %s%s\n", InfoStyle, FormatSecondary("gego scheduler start"), Reset)
	return nil
}

func runScheduleDisable(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	id := args[0]

	schedule, err := database.GetSchedule(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get schedule: %w", err)
	}

	schedule.Enabled = false
	if err := database.UpdateSchedule(ctx, schedule); err != nil {
		return fmt.Errorf("failed to update schedule: %w", err)
	}

	fmt.Printf("%s✅ Schedule disabled!%s\n", SuccessStyle, Reset)
	fmt.Printf("%sRestart the scheduler to apply changes: %s%s\n", InfoStyle, FormatSecondary("gego scheduler start"), Reset)
	return nil
}

func runScheduleRun(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	id := args[0]

	if err := initializeLLMProviders(ctx); err != nil {
		return fmt.Errorf("failed to initialize LLM providers: %w", err)
	}

	fmt.Printf("%s⏳ Executing schedule %s...%s\n", InfoStyle, FormatValue(id), Reset)

	if err := sched.ExecuteNow(ctx, id); err != nil {
		return fmt.Errorf("failed to execute schedule: %w", err)
	}

	fmt.Printf("%s✅ Schedule execution completed!%s\n", SuccessStyle, Reset)
	return nil
}
