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

	"github.com/AI2HU/gego/internal/llm"
	"github.com/AI2HU/gego/internal/llm/anthropic"
	"github.com/AI2HU/gego/internal/llm/google"
	"github.com/AI2HU/gego/internal/llm/ollama"
	"github.com/AI2HU/gego/internal/llm/openai"
	"github.com/AI2HU/gego/internal/models"
)

var promptCmd = &cobra.Command{
	Use:   "prompt",
	Short: "Manage prompts for keyword tracking",
	Long:  `Add, list, update, and delete prompt templates. Prompts are used by Gego to track keywords in LLM outputs.`,
}

var promptAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a new prompt template",
	Long:  `Create a new prompt template that will be used to generate text for LLM analysis and keyword tracking.`,
	RunE:  runPromptAdd,
}

var promptListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all prompt templates",
	Long:  `Display all configured prompt templates used for keyword tracking.`,
	RunE:  runPromptList,
}

var promptGetCmd = &cobra.Command{
	Use:   "get [id]",
	Short: "Get details of a prompt template",
	Long:  `Show detailed information about a specific prompt template used for keyword tracking.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runPromptGet,
}

var promptDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete prompt templates",
	Long:  `Remove prompt templates from the keyword tracking system. Lists all prompts and allows selection by number.`,
	Args:  cobra.NoArgs,
	RunE:  runPromptDelete,
}

var promptEnableCmd = &cobra.Command{
	Use:   "enable [id]",
	Short: "Enable a prompt template",
	Long:  `Activate a prompt template for keyword tracking.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runPromptEnable,
}

var promptDisableCmd = &cobra.Command{
	Use:   "disable [id]",
	Short: "Disable a prompt template",
	Long:  `Deactivate a prompt template from keyword tracking.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runPromptDisable,
}

func init() {
	promptCmd.AddCommand(promptAddCmd)
	promptCmd.AddCommand(promptListCmd)
	promptCmd.AddCommand(promptGetCmd)
	promptCmd.AddCommand(promptDeleteCmd)
	promptCmd.AddCommand(promptEnableCmd)
	promptCmd.AddCommand(promptDisableCmd)
}

func runPromptAdd(cmd *cobra.Command, args []string) error {
	reader := bufio.NewReader(os.Stdin)
	ctx := context.Background()

	fmt.Printf("%s➕ Add New Prompt Template%s\n", FormatHeader(""), Reset)
	fmt.Printf("%s==========================%s\n", DimStyle, Reset)
	fmt.Println()
	fmt.Printf("%sThis prompt will be sent to LLMs to generate text for keyword tracking.%s\n", InfoStyle, Reset)
	fmt.Printf("%sThe LLM responses will be analyzed to track brand mentions and keywords.%s\n", InfoStyle, Reset)
	fmt.Println()

	fmt.Printf("%sChoose how to create your prompt:%s\n", LabelStyle, Reset)
	fmt.Printf("  %s1. Generate prompts using LLM%s\n", CountStyle, Reset)
	fmt.Printf("  %s2. Add a custom prompt%s\n", CountStyle, Reset)

	method, err := promptWithRetry(reader, fmt.Sprintf("\n%sSelect method (1 or 2): %s", LabelStyle, Reset), func(input string) (string, error) {
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

	if method == "1" {
		return runPromptGenerate(reader, ctx)
	} else {
		return runPromptCustom(reader, ctx)
	}
}

func runPromptList(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	prompts, err := database.ListPrompts(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to list prompts: %w", err)
	}

	if len(prompts) == 0 {
		fmt.Printf("%sNo prompts configured. Use '%s' to add one.%s\n", WarningStyle, FormatSecondary("gego prompt add"), Reset)
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "%sID\tTEMPLATE\tTAGS\tENABLED%s\n", LabelStyle, Reset)
	fmt.Fprintf(w, "%s──\t────────\t────\t───────%s\n", DimStyle, Reset)

	for _, prompt := range prompts {
		enabled := "Yes"
		if !prompt.Enabled {
			enabled = "No"
		}

		template := prompt.Template
		if len(template) > 50 {
			template = template[:47] + "..."
		}

		tags := strings.Join(prompt.Tags, ",")
		if len(tags) > 20 {
			tags = tags[:17] + "..."
		}

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
			FormatSecondary(prompt.ID),
			FormatDim(template),
			FormatSecondary(tags),
			FormatValue(enabled),
		)
	}

	w.Flush()
	fmt.Printf("\n%sTotal: %s prompts%s\n", InfoStyle, FormatCount(len(prompts)), Reset)

	return nil
}

func runPromptGet(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	id := args[0]

	prompt, err := database.GetPrompt(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get prompt: %w", err)
	}

	fmt.Printf("%sPrompt Details%s\n", FormatHeader(""), Reset)
	fmt.Printf("%s==============%s\n", DimStyle, Reset)
	fmt.Printf("%sID: %s\n", LabelStyle, FormatSecondary(prompt.ID))
	fmt.Printf("%sEnabled: %s\n", LabelStyle, FormatValue(fmt.Sprintf("%v", prompt.Enabled)))
	fmt.Printf("%sTags: %s\n", LabelStyle, FormatSecondary(strings.Join(prompt.Tags, ", ")))
	fmt.Printf("%sCreated: %s\n", LabelStyle, FormatMeta(prompt.CreatedAt.Format(time.RFC3339)))
	fmt.Printf("%sUpdated: %s\n", LabelStyle, FormatMeta(prompt.UpdatedAt.Format(time.RFC3339)))
	fmt.Printf("\n%sTemplate:%s\n", SuccessStyle, Reset)
	fmt.Printf("%s─────────%s\n", DimStyle, Reset)
	fmt.Printf("%s\n", FormatValue(prompt.Template))

	return nil
}

func runPromptDelete(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	reader := bufio.NewReader(os.Stdin)

	fmt.Printf("%s🗑️  Delete Prompts%s\n", FormatHeader(""), Reset)
	fmt.Printf("%s==================%s\n", DimStyle, Reset)

	prompts, err := database.ListPrompts(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to list prompts: %w", err)
	}

	if len(prompts) == 0 {
		fmt.Printf("%sNo prompts to delete.%s\n", WarningStyle, Reset)
		return nil
	}

	fmt.Printf("\n%sAvailable prompts:%s\n", LabelStyle, Reset)
	fmt.Printf("%s==================%s\n", DimStyle, Reset)
	for i, prompt := range prompts {
		fmt.Printf("%s%d. %s%s\n", CountStyle, i+1, Reset, FormatValue(prompt.Template))
		if len(prompt.Tags) > 0 {
			fmt.Printf("   %sTags: %s%s\n", DimStyle, strings.Join(prompt.Tags, ", "), Reset)
		}
		fmt.Printf("   %sID: %s%s\n", DimStyle, FormatSecondary(prompt.ID), Reset)
		fmt.Println()
	}

	fmt.Printf("%sEnter the numbers of prompts you want to delete (comma-separated, e.g., 1,3,5) or 'all' to delete all: %s", LabelStyle, Reset)
	selection, _ := reader.ReadString('\n')
	selection = strings.TrimSpace(selection)

	if selection == "" {
		fmt.Printf("%sNo prompts selected.%s\n", WarningStyle, Reset)
		return nil
	}

	var selectedIndices []int

	if strings.ToLower(selection) == "all" {
		for i := range prompts {
			selectedIndices = append(selectedIndices, i)
		}
		fmt.Printf("%sSelected all %s prompts.%s\n", SuccessStyle, FormatCount(len(prompts)), Reset)
	} else {
		selections := strings.Split(selection, ",")
		for _, sel := range selections {
			sel = strings.TrimSpace(sel)
			var idx int
			_, err := fmt.Sscanf(sel, "%d", &idx)
			if err != nil || idx < 1 || idx > len(prompts) {
				fmt.Printf("%s⚠️  Invalid selection: %s (skipping)%s\n", WarningStyle, FormatValue(sel), Reset)
				continue
			}
			selectedIndices = append(selectedIndices, idx-1)
		}
	}

	if len(selectedIndices) == 0 {
		fmt.Printf("%sNo valid prompts selected.%s\n", WarningStyle, Reset)
		return nil
	}

	fmt.Printf("\n%s⚠️  You are about to delete %s prompt(s). This action cannot be undone!%s\n", ErrorStyle, FormatCount(len(selectedIndices)), Reset)
	confirmed, err := promptYesNo(reader, fmt.Sprintf("%sAre you sure? (y/N): %s", ErrorStyle, Reset))
	if err != nil {
		return err
	}

	if !confirmed {
		fmt.Printf("%sOperation cancelled.%s\n", WarningStyle, Reset)
		return nil
	}

	fmt.Printf("\n%s🗑️  Deleting selected prompts...%s\n", InfoStyle, Reset)
	deletedCount := 0
	for _, idx := range selectedIndices {
		prompt := prompts[idx]
		if err := database.DeletePrompt(ctx, prompt.ID); err != nil {
			fmt.Printf("%s⚠️  Failed to delete prompt %s: %s%s\n", ErrorStyle, FormatCount(idx+1), FormatValue(err.Error()), Reset)
			continue
		}
		deletedCount++
	}

	fmt.Printf("\n%s🎉 Successfully deleted %s prompt(s)!%s\n", SuccessStyle, FormatCount(deletedCount), Reset)
	return nil
}

func runPromptEnable(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	id := args[0]

	prompt, err := database.GetPrompt(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get prompt: %w", err)
	}

	prompt.Enabled = true
	if err := database.UpdatePrompt(ctx, prompt); err != nil {
		return fmt.Errorf("failed to update prompt: %w", err)
	}

	fmt.Printf("%s✅ Prompt enabled!%s\n", SuccessStyle, Reset)
	return nil
}

func runPromptDisable(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	id := args[0]

	prompt, err := database.GetPrompt(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get prompt: %w", err)
	}

	prompt.Enabled = false
	if err := database.UpdatePrompt(ctx, prompt); err != nil {
		return fmt.Errorf("failed to update prompt: %w", err)
	}

	fmt.Printf("%s✅ Prompt disabled!%s\n", SuccessStyle, Reset)
	return nil
}

// runPromptGenerate generates prompts using an LLM
func runPromptGenerate(reader *bufio.Reader, ctx context.Context) error {
	fmt.Printf("\n%s🤖 Generate Prompts Using LLM%s\n", FormatHeader(""), Reset)
	fmt.Printf("%s==============================%s\n", DimStyle, Reset)

	llms, err := database.ListLLMs(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to list LLMs: %w", err)
	}

	if len(llms) == 0 {
		fmt.Printf("%s❌ No LLM providers configured.%s\n", ErrorStyle, Reset)
		fmt.Printf("Please add an LLM provider first using: %s\n", FormatSecondary("gego llm add"))
		return nil
	}

	fmt.Printf("\n%sAvailable LLM providers:%s\n", LabelStyle, Reset)
	for i, llm := range llms {
		fmt.Printf("  %s%d. %s (%s)%s\n", CountStyle, i+1, Reset, FormatValue(llm.Name), FormatSecondary(llm.Provider))
	}

	llmChoice, err := promptWithRetry(reader, fmt.Sprintf("\nSelect a model to prompt generation (1-%d): ", len(llms)), func(input string) (string, error) {
		var idx int
		_, err := fmt.Sscanf(input, "%d", &idx)
		if err != nil || idx < 1 || idx > len(llms) {
			return "", fmt.Errorf("invalid choice: %s (choose 1-%d)", input, len(llms))
		}
		return input, nil
	})
	if err != nil {
		return err
	}

	var idx int
	fmt.Sscanf(llmChoice, "%d", &idx)
	selectedLLM := llms[idx-1]

	languageCode, err := promptWithRetry(reader, fmt.Sprintf("\n%sEnter language code (e.g., FR, EN, IT, ES, DE, etc.): %s", LabelStyle, Reset), func(input string) (string, error) {
		input = strings.ToUpper(strings.TrimSpace(input))
		if input == "" {
			return "", fmt.Errorf("language code is required")
		}
		if len(input) < 2 || len(input) > 3 {
			return "", fmt.Errorf("language code should be 2-3 characters (e.g., FR, EN, IT)")
		}
		return input, nil
	})
	if err != nil {
		return err
	}

	languageNames := map[string]string{
		"EN": "English", "FR": "Français", "IT": "Italiano", "ES": "Español", "DE": "Deutsch",
		"PT": "Português", "RU": "Русский", "JA": "日本語", "KO": "한국어", "ZH": "中文",
		"AR": "العربية", "NL": "Nederlands", "SV": "Svenska", "NO": "Norsk", "DA": "Dansk",
		"FI": "Suomi", "PL": "Polski", "CS": "Čeština", "HU": "Magyar", "RO": "Română",
		"BG": "Български", "HR": "Hrvatski", "SK": "Slovenčina", "SL": "Slovenščina",
		"ET": "Eesti", "LV": "Latviešu", "LT": "Lietuvių", "EL": "Ελληνικά", "TR": "Türkçe",
		"HE": "עברית", "HI": "हिन्दी", "TH": "ไทย", "VI": "Tiếng Việt", "ID": "Bahasa Indonesia",
		"MS": "Bahasa Melayu", "TL": "Filipino",
	}

	languageName := languageNames[languageCode]
	if languageName == "" {
		languageName = languageCode // Fallback to code if name not found
	}

	userInput, err := promptWithRetry(reader, fmt.Sprintf("\n%sDescribe what kind of prompts you need in %s (e.g., 'questions about streaming services'): %s", LabelStyle, FormatValue(languageName), Reset), func(input string) (string, error) {
		if input == "" {
			return "", fmt.Errorf("description is required")
		}
		return input, nil
	})
	if err != nil {
		return err
	}

	promptCountStr, err := promptWithRetry(reader, fmt.Sprintf("\n%sHow many prompts would you like to generate? [20]: %s", LabelStyle, Reset), func(input string) (string, error) {
		input = strings.TrimSpace(input)
		if input == "" {
			return "20", nil
		}
		var count int
		_, err := fmt.Sscanf(input, "%d", &count)
		if err != nil {
			return "", fmt.Errorf("invalid number: %s (enter a positive integer)", input)
		}
		if count < 1 {
			return "", fmt.Errorf("count must be at least 1")
		}
		return input, nil
	})
	if err != nil {
		return err
	}

	var promptCount int
	fmt.Sscanf(promptCountStr, "%d", &promptCount)

	fmt.Printf("\n%s📋 Fetching existing prompts...%s\n", InfoStyle, Reset)
	existingPrompts, err := database.ListPrompts(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to fetch existing prompts: %w", err)
	}

	fmt.Printf("Found %d existing prompts.\n", len(existingPrompts))

	var existingPromptTemplates []string
	for _, prompt := range existingPrompts {
		existingPromptTemplates = append(existingPromptTemplates, prompt.Template)
	}

	fmt.Printf("\n%s🔍 Generating prompts...%s\n", InfoStyle, Reset)

	prePrompt := llm.GenerateGEOPromptTemplate(userInput, existingPromptTemplates, languageCode, promptCount)

	var provider llm.Provider
	switch selectedLLM.Provider {
	case "openai":
		provider = openai.New(selectedLLM.APIKey, selectedLLM.BaseURL)
	case "anthropic":
		provider = anthropic.New(selectedLLM.APIKey, selectedLLM.BaseURL)
	case "ollama":
		provider = ollama.New(selectedLLM.BaseURL)
	case "google":
		provider = google.New(selectedLLM.APIKey, selectedLLM.BaseURL)
	default:
		return fmt.Errorf("unsupported LLM provider: %s", selectedLLM.Provider)
	}

	response, err := provider.Generate(ctx, prePrompt, llm.Config{
		Model: selectedLLM.Model,
	})
	if err != nil {
		return fmt.Errorf("failed to generate prompts: %w", err)
	}

	if response.Error != "" {
		return fmt.Errorf("LLM error: %s", response.Error)
	}

	promptLines := strings.Split(strings.TrimSpace(response.Text), "\n")
	var generatedPrompts []string

	for _, line := range promptLines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		if len(line) > 2 && (line[1] == '.' || line[1] == ')') && (line[0] >= '0' && line[0] <= '9') {
			line = strings.TrimSpace(line[2:])
		}

		if line != "" {
			generatedPrompts = append(generatedPrompts, line)
		}
	}

	if len(generatedPrompts) == 0 {
		return fmt.Errorf("no valid prompts were generated")
	}

	fmt.Printf("\n%s✅ Generated %s prompts:%s\n", SuccessStyle, FormatCount(len(generatedPrompts)), Reset)
	fmt.Printf("%s================================%s\n", DimStyle, Reset)
	for i, prompt := range generatedPrompts {
		fmt.Printf("%s%d. %s%s\n", CountStyle, i+1, Reset, FormatValue(prompt))
	}

	fmt.Printf("\n%sWould you like to save all %s generated prompts? (y/n): %s", LabelStyle, FormatCount(len(generatedPrompts)), Reset)
	choice, _ := reader.ReadString('\n')
	choice = strings.TrimSpace(strings.ToLower(choice))

	if choice != "y" && choice != "yes" {
		fmt.Printf("%sOperation cancelled.%s\n", WarningStyle, Reset)
		return nil
	}

	fmt.Printf("%sSelected all %s prompts.%s\n", SuccessStyle, FormatCount(len(generatedPrompts)), Reset)

	fmt.Printf("\n%s💾 Saving all prompts...%s\n", InfoStyle, Reset)
	savedCount := 0
	for _, promptText := range generatedPrompts {
		prompt := &models.Prompt{
			ID:       uuid.New().String(),
			Template: promptText,
			Tags:     []string{"generated", "llm-created", fmt.Sprintf("lang-%s", languageCode)},
			Enabled:  true,
		}

		if err := database.CreatePrompt(ctx, prompt); err != nil {
			fmt.Printf("%s⚠️  Failed to save prompt: %s%s\n", ErrorStyle, FormatValue(err.Error()), Reset)
			continue
		}

		savedCount++
	}

	fmt.Printf("\n%s🎉 Successfully saved %s prompt(s)!%s\n", SuccessStyle, FormatCount(savedCount), Reset)
	return nil
}

// runPromptCustom allows users to add a custom prompt
func runPromptCustom(reader *bufio.Reader, ctx context.Context) error {
	fmt.Printf("\n%s✏️  Add Custom Prompt%s\n", FormatHeader(""), Reset)
	fmt.Printf("%s=====================%s\n", DimStyle, Reset)

	prompt := &models.Prompt{
		ID:      uuid.New().String(),
		Enabled: true,
	}

	fmt.Println("\nEnter prompt template (press Ctrl+D when done):")
	fmt.Println("Example: What are the top streaming services for watching movies?")
	fmt.Println("Note: This prompt will be used to generate text that will be analyzed for keyword mentions.")
	fmt.Println()

	var templateLines []string
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		templateLines = append(templateLines, scanner.Text())
	}
	prompt.Template = strings.Join(templateLines, "\n")

	if prompt.Template == "" {
		return fmt.Errorf("prompt template cannot be empty")
	}

	tags, err := promptOptional(reader, "\nTags (comma-separated, optional): ", "")
	if err != nil {
		return err
	}
	if tags != "" {
		prompt.Tags = strings.Split(tags, ",")
		for i := range prompt.Tags {
			prompt.Tags[i] = strings.TrimSpace(prompt.Tags[i])
		}
	}

	if err := database.CreatePrompt(ctx, prompt); err != nil {
		return fmt.Errorf("failed to create prompt: %w", err)
	}

	fmt.Println("\n✅ Prompt added successfully!")
	fmt.Printf("ID: %s\n", prompt.ID)

	return nil
}
