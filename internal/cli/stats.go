package cli

import (
	"context"
	"fmt"
	"sort"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/AI2HU/gego/internal/db"
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

func init() {
	statsCmd.AddCommand(statsKeywordsCmd)
	statsCmd.AddCommand(statsKeywordCmd)

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

	fmt.Printf("%s📊 Top Keywords by Mentions%s\n", HeaderStyle, Reset)
	fmt.Printf("%s===========================%s\n", DimStyle, Reset)
	fmt.Println()

	w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "%sRANK\tKEYWORD\tMENTIONS%s\n", LabelStyle, Reset)
	fmt.Fprintf(w, "%s────\t───────\t────────%s\n", DimStyle, Reset)

	for i, keyword := range keywords {
		fmt.Fprintf(w, "%s\t%s\t%s\n",
			FormatCount(i+1),
			FormatValue(keyword.Keyword),
			FormatCount(keyword.Count),
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
			// Show the actual prompt content instead of the name
			displayText = prompt.Template
			// Truncate in the middle if too long (show start and end)
			if len(displayText) > 60 {
				start := displayText[:25]
				end := displayText[len(displayText)-25:]
				displayText = start + "..." + end
			}
		}
		fmt.Printf("  %s%d. %s: %s mentions%s\n", CountStyle, i+1, Reset, FormatValue(displayText), CountStyle+fmt.Sprintf(" %d", item.Value)+Reset)
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
		name := item.Key
		if err == nil {
			name = fmt.Sprintf("%s (%s)", llm.Name, llm.Provider)
		}
		fmt.Printf("  %s%d. %s: %s mentions%s\n", CountStyle, i+1, Reset, FormatValue(name), CountStyle+fmt.Sprintf(" %d", item.Value)+Reset)
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
		fmt.Printf("  %s: %s mentions%s\n", FormatValue(item.Key), CountStyle+fmt.Sprintf(" %d", item.Value)+Reset, Reset)
		if i >= 10 {
			break
		}
	}

	return nil
}

// Helper to filter responses
func filterResponses(ctx context.Context, filter db.ResponseFilter) error {
	responses, err := database.ListResponses(ctx, filter)
	if err != nil {
		return err
	}

	fmt.Printf("Found %d responses\n", len(responses))
	for _, resp := range responses {
		promptPreview := resp.PromptText
		if len(promptPreview) > 50 {
			promptPreview = promptPreview[:50] + "..."
		}
		fmt.Printf("- %s: %s\n", resp.LLMName, promptPreview)
	}

	return nil
}
