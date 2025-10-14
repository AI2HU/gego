package cli

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/AI2HU/gego/internal/models"
	"github.com/AI2HU/gego/internal/shared"
)

var (
	searchKeyword       string
	searchLimit         int
	searchCaseSensitive bool
)

var searchCmd = &cobra.Command{
	Use:   "search [keyword]",
	Short: "Search for specific keywords in all responses",
	Long:  `Search for specific keywords in all LLM responses and display the context around each match.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runSearch,
}

func init() {
	searchCmd.Flags().IntVarP(&searchLimit, "limit", "l", 50, "Maximum number of results to display")
	searchCmd.Flags().BoolVarP(&searchCaseSensitive, "case-sensitive", "c", false, "Make search case-sensitive")
}

func runSearch(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	keyword := args[0]

	fmt.Printf("%s🔍 Searching for keyword: \"%s\"%s\n", HeaderStyle, CountStyle+keyword+Reset, Reset)
	fmt.Println()

	// Use the database interface to search for responses containing the keyword
	filter := shared.ResponseFilter{
		Keyword: keyword,
		Limit:   searchLimit * 10, // Get more responses to find matches
	}

	responses, err := database.ListResponses(ctx, filter)
	if err != nil {
		return fmt.Errorf("failed to search responses: %w", err)
	}

	fmt.Printf("%s📊 Found %s responses containing \"%s\"%s\n", InfoStyle, CountStyle+fmt.Sprintf("%d", len(responses))+Reset, CountStyle+keyword+Reset, Reset)
	fmt.Println()

	if len(responses) == 0 {
		fmt.Printf("%s❌ No matches found for keyword \"%s\"%s\n", ErrorStyle, CountStyle+keyword+Reset, Reset)
		return nil
	}

	// Create regex for case-sensitive/insensitive matching
	var regex *regexp.Regexp
	if searchCaseSensitive {
		regex = regexp.MustCompile(regexp.QuoteMeta(keyword))
	} else {
		regex = regexp.MustCompile("(?i)" + regexp.QuoteMeta(keyword))
	}

	// Process matches
	var matches []SearchMatch
	for _, response := range responses {
		matches = append(matches, findMatches(response, regex, keyword)...)
	}

	if len(matches) == 0 {
		fmt.Printf("%s❌ No matches found for keyword \"%s\"%s\n", ErrorStyle, CountStyle+keyword+Reset, Reset)
		return nil
	}

	fmt.Printf("%s✅ Found %s matches for keyword \"%s\"%s\n", SuccessStyle, CountStyle+fmt.Sprintf("%d", len(matches))+Reset, CountStyle+keyword+Reset, Reset)
	fmt.Println()

	// Display results
	displayCount := 0
	for i, match := range matches {
		if displayCount >= searchLimit {
			fmt.Printf("\n%s... and %s more matches (use --limit to see more)%s\n", DimStyle, CountStyle+fmt.Sprintf("%d", len(matches)-displayCount)+Reset, Reset)
			break
		}

		fmt.Printf("%s📄 Match %s:%s\n", TitleStyle, CountStyle+fmt.Sprintf("%d", i+1)+Reset, Reset)
		fmt.Printf("   %s🏷️  Prompt:%s %s\n", LabelStyle, Reset, FormatValue(match.PromptName))
		fmt.Printf("   %s🤖 LLM:%s %s (%s%s%s)\n", LabelStyle, Reset, FormatValue(match.LLMName), SecondaryStyle, match.LLMProvider, Reset)

		// Display temperature with special handling for old responses
		tempDisplay := fmt.Sprintf("%.1f", match.Temperature)
		if match.Temperature == 0.0 {
			tempDisplay = "N/A (legacy)"
		}
		fmt.Printf("   %s🌡️  Temperature:%s %s\n", LabelStyle, Reset, FormatValue(tempDisplay))

		fmt.Printf("   %s📅 Date:%s %s\n", LabelStyle, Reset, FormatMeta(match.CreatedAt.Format("2006-01-02 15:04:05")))
		fmt.Println()

		fmt.Printf("   %s📝 Context:%s\n", SuccessStyle, Reset)
		fmt.Printf("   %s\n", match.Context)
		fmt.Println()

		fmt.Printf("   %s📋 Full Prompt:%s\n", SuccessStyle, Reset)
		fmt.Printf("   %s\n", FormatDim(match.FullPrompt))
		fmt.Println()
		fmt.Printf("   %s%s%s\n", DimStyle, strings.Repeat("─", 80), Reset)
		fmt.Println()

		displayCount++
	}

	return nil
}

type SearchMatch struct {
	ResponseID  string
	PromptID    string
	PromptName  string
	FullPrompt  string
	LLMName     string
	LLMProvider string
	Temperature float64
	Context     string
	CreatedAt   time.Time
}

func findMatches(response *models.Response, regex *regexp.Regexp, keyword string) []SearchMatch {
	var matches []SearchMatch

	// Find all matches in the response text
	indices := regex.FindAllStringIndex(response.ResponseText, -1)

	for _, index := range indices {
		start := index[0]
		end := index[1]

		// Get context: 100 chars before and 100 chars after
		contextStart := start - 100
		if contextStart < 0 {
			contextStart = 0
		}
		contextEnd := end + 100
		if contextEnd > len(response.ResponseText) {
			contextEnd = len(response.ResponseText)
		}

		contextText := response.ResponseText[contextStart:contextEnd]

		// Highlight the keyword in context with ANSI colors
		highlightedContext := strings.ReplaceAll(contextText, keyword, FormatHighlight(keyword))

		// Get prompt name
		promptName := "Unknown Prompt"
		if prompt, err := database.GetPrompt(context.Background(), response.PromptID); err == nil {
			promptName = prompt.Template
		}

		matches = append(matches, SearchMatch{
			ResponseID:  response.ID,
			PromptID:    response.PromptID,
			PromptName:  promptName,
			FullPrompt:  response.PromptText,
			LLMName:     response.LLMName,
			LLMProvider: response.LLMProvider,
			Temperature: response.Temperature,
			Context:     highlightedContext,
			CreatedAt:   response.CreatedAt,
		})
	}

	return matches
}
