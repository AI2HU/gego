package cli

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"sort"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/AI2HU/gego/internal/shared"
)

var (
	statsLimit   int
	statsKeyword string
)

var statsCmd = &cobra.Command{
	Use:   "stats",
	Short: "View statistics and insights",
	Long:  `View various statistics and insights about keyword mentions.`,
}

var statsKeywordsCmd = &cobra.Command{
	Use:   "keywords",
	Short: "View top keywords by mentions",
	RunE:  runStatsKeywords,
}

var statsKeywordCmd = &cobra.Command{
	Use:   "keyword [name]",
	Short: "View statistics for a specific keyword",
	Args:  cobra.ExactArgs(1),
	RunE:  runStatsKeyword,
}

var statsResetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Reset all statistics by clearing all responses",
	Long:  `Reset all statistics by deleting all responses from the database. This will clear all keyword statistics, prompt statistics, and LLM statistics. Prompts and LLMs will remain intact.`,
	Args:  cobra.NoArgs,
	RunE:  runStatsReset,
}

func init() {
	statsCmd.AddCommand(statsKeywordsCmd)
	statsCmd.AddCommand(statsKeywordCmd)
	statsCmd.AddCommand(statsResetCmd)

	statsCmd.PersistentFlags().IntVarP(&statsLimit, "limit", "l", 10, "Limit number of results")
	statsKeywordCmd.Flags().StringVarP(&statsKeyword, "keyword", "k", "", "Keyword name")
}

func runStatsKeywords(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	keywords, err := database.GetTopKeywords(ctx, statsLimit, nil, nil)
	if err != nil {
		return fmt.Errorf("failed to get top keywords: %w", err)
	}

	if len(keywords) == 0 {
		fmt.Printf("%sNo keyword statistics available yet. Run some schedules first!%s\n", WarningStyle, Reset)
		return nil
	}

	// Calculate total mentions for percentage calculation
	totalMentions := 0
	for _, keyword := range keywords {
		totalMentions += keyword.Count
	}

	fmt.Printf("%s📊 Top Keywords by Mentions%s\n", HeaderStyle, Reset)
	fmt.Printf("%s===========================%s\n", DimStyle, Reset)
	fmt.Println()

	w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "%sRANK\tKEYWORD\tMENTIONS%s\n", LabelStyle, Reset)
	fmt.Fprintf(w, "%s────\t───────\t────────%s\n", DimStyle, Reset)

	for i, keyword := range keywords {
		percentage := float64(keyword.Count) / float64(totalMentions) * 100
		fmt.Fprintf(w, "%s\t%s\t%s\n",
			FormatCount(i+1),
			FormatValue(keyword.Keyword),
			fmt.Sprintf("%s%d (%.1f%%)%s", CountStyle, keyword.Count, percentage, Reset),
		)
	}

	w.Flush()
	return nil
}

func runStatsKeyword(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	keywordName := args[0]

	stats, err := database.SearchKeyword(ctx, keywordName, nil, nil)
	if err != nil {
		return fmt.Errorf("failed to get keyword stats: %w", err)
	}

	fmt.Printf("%s📊 Keyword Statistics: %s%s\n", HeaderStyle, CountStyle+keywordName+Reset, Reset)
	fmt.Printf("%s========================%s\n", DimStyle, Reset)
	fmt.Println()

	fmt.Printf("%sTotal Mentions: %s\n", LabelStyle, FormatCount(stats.TotalMentions))
	fmt.Printf("%sUnique Prompts: %s\n", LabelStyle, FormatCount(stats.UniquePrompts))
	fmt.Printf("%sUnique LLMs: %s\n", LabelStyle, FormatCount(stats.UniqueLLMs))
	fmt.Printf("%sFirst Seen: %s\n", LabelStyle, FormatMeta(stats.FirstSeen.Format("2006-01-02 15:04:05")))
	fmt.Printf("%sLast Seen: %s\n", LabelStyle, FormatMeta(stats.LastSeen.Format("2006-01-02 15:04:05")))
	fmt.Println()

	// Top prompts for this keyword
	fmt.Printf("%sTop Prompts:%s\n", SuccessStyle, Reset)
	fmt.Printf("%s────────────%s\n", DimStyle, Reset)
	type kv struct {
		Key   string
		Value int
	}
	var promptList []kv
	for k, v := range stats.ByPrompt {
		promptList = append(promptList, kv{k, v})
	}
	sort.Slice(promptList, func(i, j int) bool {
		return promptList[i].Value > promptList[j].Value
	})

	for i, item := range promptList {
		if i >= statsLimit {
			break
		}
		prompt, err := database.GetPrompt(ctx, item.Key)
		displayText := item.Key
		if err == nil {
			// Show the actual prompt content instead of the ID
			displayText = prompt.Template
			// Truncate if too long (show start and end)
			if len(displayText) > 80 {
				start := displayText[:35]
				end := displayText[len(displayText)-35:]
				displayText = start + "..." + end
			}
		} else {
			// Show that this is historical data
			displayText = fmt.Sprintf("[Deleted Prompt: %s]", item.Key[:8])
		}
		percentage := float64(item.Value) / float64(stats.TotalMentions) * 100
		fmt.Printf("  %s%d. %s%s\n", CountStyle, i+1, Reset, FormatValue(displayText))
		fmt.Printf("     %s%d mentions (%.1f%%)%s\n", DimStyle, item.Value, percentage, Reset)
	}

	fmt.Println()

	// Top LLMs for this keyword
	fmt.Printf("%sTop LLMs:%s\n", SuccessStyle, Reset)
	fmt.Printf("%s─────────%s\n", DimStyle, Reset)
	var llmList []kv
	for k, v := range stats.ByLLM {
		llmList = append(llmList, kv{k, v})
	}
	sort.Slice(llmList, func(i, j int) bool {
		return llmList[i].Value > llmList[j].Value
	})

	for i, item := range llmList {
		if i >= statsLimit {
			break
		}
		llm, err := database.GetLLM(ctx, item.Key)
		displayText := item.Key
		if err == nil {
			// Show model name and provider instead of ID
			displayText = fmt.Sprintf("%s (%s)", llm.Model, llm.Provider)
		} else {
			// Show that this is historical data
			displayText = fmt.Sprintf("[Deleted LLM: %s]", item.Key[:8])
		}
		percentage := float64(item.Value) / float64(stats.TotalMentions) * 100
		fmt.Printf("  %s%d. %s%s\n", CountStyle, i+1, Reset, FormatValue(displayText))
		fmt.Printf("     %s%d mentions (%.1f%%)%s\n", DimStyle, item.Value, percentage, Reset)
	}

	fmt.Println()

	// Top providers for this brand
	fmt.Printf("%sBy Provider:%s\n", SuccessStyle, Reset)
	fmt.Printf("%s────────────%s\n", DimStyle, Reset)
	var providerList []kv
	for k, v := range stats.ByProvider {
		providerList = append(providerList, kv{k, v})
	}
	sort.Slice(providerList, func(i, j int) bool {
		return providerList[i].Value > providerList[j].Value
	})

	for i, item := range providerList {
		percentage := float64(item.Value) / float64(stats.TotalMentions) * 100
		fmt.Printf("  %s: %s mentions (%.1f%%)%s\n", FormatValue(item.Key), CountStyle+fmt.Sprintf(" %d", item.Value)+Reset, percentage, Reset)
		if i >= 10 {
			break
		}
	}

	return nil
}

func runStatsReset(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	reader := bufio.NewReader(os.Stdin)

	fmt.Printf("%s🔄 Reset All Statistics%s\n", FormatHeader(""), Reset)
	fmt.Printf("%s========================%s\n", DimStyle, Reset)
	fmt.Println()

	fmt.Printf("%s⚠️  Warning: This will permanently delete ALL responses from the database.%s\n", WarningStyle, Reset)
	fmt.Printf("%sThis action will:%s\n", LabelStyle, Reset)
	fmt.Printf("  %s• Clear all keyword statistics%s\n", DimStyle, Reset)
	fmt.Printf("  %s• Clear all prompt statistics%s\n", DimStyle, Reset)
	fmt.Printf("  %s• Clear all LLM statistics%s\n", DimStyle, Reset)
	fmt.Printf("  %s• Delete all response data%s\n", DimStyle, Reset)
	fmt.Printf("  %s• Keep prompts and LLMs intact%s\n", DimStyle, Reset)
	fmt.Println()

	fmt.Printf("%sThis action cannot be undone!%s\n", ErrorStyle, Reset)
	fmt.Println()

	confirmed, err := promptYesNo(reader, fmt.Sprintf("%sAre you sure you want to reset all statistics? (y/N): %s", ErrorStyle, Reset))
	if err != nil {
		return err
	}

	if !confirmed {
		fmt.Printf("%sCancelled.%s\n", WarningStyle, Reset)
		return nil
	}

	fmt.Printf("\n%s🗑️  Clearing all responses...%s\n", InfoStyle, Reset)

	// Get count of responses before deletion
	responses, err := database.ListResponses(ctx, shared.ResponseFilter{Limit: 1})
	if err != nil {
		return fmt.Errorf("failed to check responses: %w", err)
	}

	if len(responses) == 0 {
		fmt.Printf("%sNo responses found to delete.%s\n", WarningStyle, Reset)
		return nil
	}

	// Delete all responses
	deletedCount, err := database.DeleteAllResponses(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete responses: %w", err)
	}

	fmt.Printf("%s✅ Successfully deleted %s responses!%s\n", SuccessStyle, FormatCount(deletedCount), Reset)
	fmt.Printf("%s🎉 All statistics have been reset.%s\n", SuccessStyle, Reset)
	fmt.Printf("%sYou can now run new prompts to generate fresh statistics.%s\n", InfoStyle, Reset)

	return nil
}
