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

// Provider represents available LLM providers
type Provider int

const (
	OpenAI Provider = iota + 1
	Anthropic
	Ollama
	Google
)

// String returns the string representation of the provider
func (p Provider) String() string {
	switch p {
	case OpenAI:
		return "openai"
	case Anthropic:
		return "anthropic"
	case Ollama:
		return "ollama"
	case Google:
		return "google"
	default:
		return "unknown"
	}
}

// FromString converts a string to a Provider
func FromString(s string) Provider {
	switch s {
	case "openai":
		return OpenAI
	case "anthropic":
		return Anthropic
	case "ollama":
		return Ollama
	case "google":
		return Google
	default:
		return 0 // Unknown provider
	}
}

// DisplayName returns the display name for the provider with model info
func (p Provider) DisplayName() string {
	switch p {
	case OpenAI:
		return "OpenAI (ChatGPT)"
	case Anthropic:
		return "Anthropic (Claude)"
	case Ollama:
		return "Ollama (local models)"
	case Google:
		return "Google (Gemini)"
	default:
		return "Unknown"
	}
}

// AllProviders returns a slice of all available providers
func AllProviders() []Provider {
	return []Provider{OpenAI, Anthropic, Ollama, Google}
}

// GetConsoleURL returns the console URL where API keys can be generated for the provider
func (p Provider) GetConsoleURL() string {
	switch p {
	case OpenAI:
		return "https://platform.openai.com/api-keys"
	case Anthropic:
		return "https://console.anthropic.com/"
	case Google:
		return "https://makersuite.google.com/app/apikey"
	case Ollama:
		return "https://ollama.ai/" // Ollama doesn't need API keys, but provides setup info
	default:
		return ""
	}
}

var llmCmd = &cobra.Command{
	Use:   "llm",
	Short: "Manage LLM providers",
	Long:  `Add, list, update, and delete LLM provider configurations.`,
}

var llmAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a new LLM provider",
	RunE:  runLLMAdd,
}

var llmListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all LLM providers",
	RunE:  runLLMList,
}

var llmGetCmd = &cobra.Command{
	Use:   "get [id]",
	Short: "Get details of an LLM provider",
	Args:  cobra.ExactArgs(1),
	RunE:  runLLMGet,
}

var llmDeleteCmd = &cobra.Command{
	Use:   "delete [id]",
	Short: "Delete an LLM provider or all LLMs",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runLLMDelete,
}

var llmEnableCmd = &cobra.Command{
	Use:   "enable [id]",
	Short: "Enable an LLM provider",
	Args:  cobra.ExactArgs(1),
	RunE:  runLLMEnable,
}

var llmDisableCmd = &cobra.Command{
	Use:   "disable [id]",
	Short: "Disable an LLM provider",
	Args:  cobra.ExactArgs(1),
	RunE:  runLLMDisable,
}

var llmUpdateCmd = &cobra.Command{
	Use:   "update [id]",
	Short: "Update an LLM provider configuration",
	Args:  cobra.ExactArgs(1),
	RunE:  runLLMUpdate,
}

func init() {
	llmCmd.AddCommand(llmAddCmd)
	llmCmd.AddCommand(llmListCmd)
	llmCmd.AddCommand(llmGetCmd)
	llmCmd.AddCommand(llmUpdateCmd)
	llmCmd.AddCommand(llmDeleteCmd)
	llmCmd.AddCommand(llmEnableCmd)
	llmCmd.AddCommand(llmDisableCmd)
}

func runLLMAdd(cmd *cobra.Command, args []string) error {
	reader := bufio.NewReader(os.Stdin)
	ctx := context.Background()

	fmt.Printf("%s➕ Add New LLM Models%s\n", FormatHeader(""), Reset)
	fmt.Printf("%s====================%s\n", DimStyle, Reset)
	fmt.Println()

	// Step 1: Select provider
	fmt.Printf("%sAvailable providers:%s\n", LabelStyle, Reset)
	providers := AllProviders()
	for i, provider := range providers {
		fmt.Printf("  %s%d. %s%s\n", CountStyle, i+1, Reset, FormatValue(provider.DisplayName()))
	}

	providerChoice, err := promptWithRetry(reader, "\nSelect provider (1, 2, 3, or 4): ", func(input string) (string, error) {
		switch input {
		case "1", "2", "3", "4":
			return input, nil
		default:
			return "", fmt.Errorf("invalid provider choice: %s (choose 1, 2, 3, or 4)", input)
		}
	})
	if err != nil {
		return err
	}

	var selectedProvider Provider
	switch providerChoice {
	case "1":
		selectedProvider = OpenAI
	case "2":
		selectedProvider = Anthropic
	case "3":
		selectedProvider = Ollama
	case "4":
		selectedProvider = Google
	}

	providerName := selectedProvider.String()

	// Step 2: Get credentials
	var apiKey, baseURL string

	if selectedProvider == OpenAI || selectedProvider == Anthropic || selectedProvider == Google {
		fmt.Printf("\n🔑 %s API Key Required\n", selectedProvider.DisplayName())
		fmt.Printf("Get your API key from: %s\n", selectedProvider.GetConsoleURL())

		apiKey, err = promptWithRetry(reader, "\nAPI Key: ", func(input string) (string, error) {
			if input == "" {
				return "", fmt.Errorf("API key is required for %s", selectedProvider.DisplayName())
			}
			return input, nil
		})
		if err != nil {
			return err
		}
	}

	if selectedProvider == Ollama {
		fmt.Printf("\n🌐 %s Configuration\n", selectedProvider.DisplayName())
		fmt.Printf("Ollama setup guide: %s\n", selectedProvider.GetConsoleURL())

		baseURL, err = promptWithRetry(reader, "\nBase URL [http://localhost:11434]: ", func(input string) (string, error) {
			if input == "" {
				return "http://localhost:11434", nil
			}
			return input, nil
		})
		if err != nil {
			return err
		}
	}

	// Step 3: List available models
	fmt.Println("\n🔍 Fetching available models...")

	provider, ok := llmRegistry.Get(providerName)
	if !ok {
		return fmt.Errorf("provider not found in registry: %s", providerName)
	}

	availableModels, err := provider.ListModels(ctx, apiKey, baseURL)
	if err != nil {
		return fmt.Errorf("failed to list models: %w", err)
	}

	if len(availableModels) == 0 {
		fmt.Println("\n⚠️  No models found for this provider")
		return nil
	}

	// Step 4: Display models
	fmt.Println("\nAvailable text-to-text models:")
	fmt.Println("==============================")
	for i, model := range availableModels {
		fmt.Printf("%d. %s", i+1, model.Name)
		if model.Description != "" {
			fmt.Printf(" - %s", model.Description)
		}
		fmt.Println()
	}

	// Step 5: Let user select models
	selection, err := promptWithRetry(reader, "\nSelect models (comma-separated numbers, or 'all'): ", func(input string) (string, error) {
		if strings.ToLower(input) == "all" {
			return input, nil
		}

		// Validate comma-separated numbers
		selections := strings.Split(input, ",")
		for _, sel := range selections {
			sel = strings.TrimSpace(sel)
			var idx int
			_, err := fmt.Sscanf(sel, "%d", &idx)
			if err != nil || idx < 1 || idx > len(availableModels) {
				return "", fmt.Errorf("invalid selection: %s (must be numbers 1-%d or 'all')", sel, len(availableModels))
			}
		}
		return input, nil
	})
	if err != nil {
		return err
	}

	var selectedModels []models.ModelInfo

	if strings.ToLower(selection) == "all" {
		selectedModels = availableModels
	} else {
		// Parse comma-separated numbers
		selections := strings.Split(selection, ",")
		for _, sel := range selections {
			sel = strings.TrimSpace(sel)
			var idx int
			fmt.Sscanf(sel, "%d", &idx)
			selectedModels = append(selectedModels, availableModels[idx-1])
		}
	}

	// Step 6: Create LLM entries for each selected model
	fmt.Printf("\n%s📝 Adding %s model(s)...%s\n", InfoStyle, FormatCount(len(selectedModels)), Reset)

	addedCount := 0
	for _, model := range selectedModels {
		llm := &models.LLMConfig{
			ID:        uuid.New().String(),
			Name:      model.Name, // Use model name as the LLM name
			Provider:  providerName,
			Model:     model.ID,
			APIKey:    apiKey,
			BaseURL:   baseURL,
			Enabled:   true,
			Config:    make(map[string]string),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		if err := database.CreateLLM(ctx, llm); err != nil {
			fmt.Printf("%s⚠️  Failed to add %s: %s%s\n", ErrorStyle, FormatValue(model.Name), FormatValue(err.Error()), Reset)
			continue
		}

		fmt.Printf("%s✅ Added: %s (ID: %s)%s\n", SuccessStyle, FormatValue(model.Name), FormatSecondary(llm.ID), Reset)
		addedCount++
	}

	fmt.Printf("\n%s🎉 Successfully added %s/%s model(s)!%s\n", SuccessStyle, FormatCount(addedCount), FormatCount(len(selectedModels)), Reset)
	return nil
}

func runLLMList(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	llms, err := database.ListLLMs(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to list LLMs: %w", err)
	}

	if len(llms) == 0 {
		fmt.Printf("%sNo LLM providers configured. Use '%s' to add one.%s\n", WarningStyle, FormatSecondary("gego llm add"), Reset)
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "%sID\tNAME\tPROVIDER\tMODEL\tENABLED%s\n", LabelStyle, Reset)
	fmt.Fprintf(w, "%s──\t────\t────────\t─────\t───────%s\n", DimStyle, Reset)

	for _, llm := range llms {
		enabled := "Yes"
		if !llm.Enabled {
			enabled = "No"
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
			FormatSecondary(llm.ID),
			FormatValue(llm.Name),
			FormatSecondary(llm.Provider),
			FormatValue(llm.Model),
			FormatValue(enabled),
		)
	}

	w.Flush()
	fmt.Printf("\n%sTotal: %s LLM providers%s\n", InfoStyle, FormatCount(len(llms)), Reset)

	return nil
}

func runLLMGet(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	id := args[0]

	llm, err := database.GetLLM(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get LLM: %w", err)
	}

	fmt.Printf("%sLLM Provider Details%s\n", FormatHeader(""), Reset)
	fmt.Printf("%s===================%s\n", DimStyle, Reset)
	fmt.Printf("%sID: %s\n", LabelStyle, FormatSecondary(llm.ID))
	fmt.Printf("%sName: %s\n", LabelStyle, FormatValue(llm.Name))
	fmt.Printf("%sProvider: %s\n", LabelStyle, FormatSecondary(llm.Provider))
	fmt.Printf("%sModel: %s\n", LabelStyle, FormatValue(llm.Model))
	fmt.Printf("%sAPI Key: %s\n", LabelStyle, FormatSecondary(maskAPIKey(llm.APIKey)))
	if llm.BaseURL != "" {
		fmt.Printf("%sBase URL: %s\n", LabelStyle, FormatSecondary(llm.BaseURL))
	}
	fmt.Printf("%sEnabled: %s\n", LabelStyle, FormatValue(fmt.Sprintf("%v", llm.Enabled)))
	fmt.Printf("%sCreated: %s\n", LabelStyle, FormatMeta(llm.CreatedAt.Format(time.RFC3339)))
	fmt.Printf("%sUpdated: %s\n", LabelStyle, FormatMeta(llm.UpdatedAt.Format(time.RFC3339)))

	if len(llm.Config) > 0 {
		fmt.Printf("\n%sConfiguration:%s\n", SuccessStyle, Reset)
		for k, v := range llm.Config {
			fmt.Printf("  %s: %s\n", FormatLabel(k), FormatValue(v))
		}
	}

	return nil
}

func runLLMDelete(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	reader := bufio.NewReader(os.Stdin)

	// If no ID provided, delete all LLMs
	if len(args) == 0 {
		// First, get all LLMs to show count
		llms, err := database.ListLLMs(ctx, nil)
		if err != nil {
			return fmt.Errorf("failed to list LLMs: %w", err)
		}

		if len(llms) == 0 {
			fmt.Printf("%sNo LLMs to delete.%s\n", WarningStyle, Reset)
			return nil
		}

		// Confirm deletion of all LLMs
		confirmed, err := promptYesNo(reader, fmt.Sprintf("%sAre you sure you want to delete ALL %s LLMs? This action cannot be undone! (y/N): %s", ErrorStyle, FormatCount(len(llms)), Reset))
		if err != nil {
			return err
		}

		if !confirmed {
			fmt.Printf("%sCancelled.%s\n", WarningStyle, Reset)
			return nil
		}

		// Use database delete all method
		deletedCount, err := database.DeleteAllLLMs(ctx)
		if err != nil {
			return fmt.Errorf("failed to delete all LLMs: %w", err)
		}

		fmt.Printf("%s✅ Successfully deleted %s LLMs!%s\n", SuccessStyle, FormatCount(deletedCount), Reset)
		return nil
	}

	// Delete specific LLM (original behavior)
	id := args[0]
	confirmed, err := promptYesNo(reader, fmt.Sprintf("%sAre you sure you want to delete LLM provider %s? (y/N): %s", ErrorStyle, FormatValue(id), Reset))
	if err != nil {
		return err
	}

	if !confirmed {
		fmt.Printf("%sCancelled.%s\n", WarningStyle, Reset)
		return nil
	}

	if err := database.DeleteLLM(ctx, id); err != nil {
		return fmt.Errorf("failed to delete LLM: %w", err)
	}

	fmt.Printf("%s✅ LLM provider deleted successfully!%s\n", SuccessStyle, Reset)
	return nil
}

func runLLMEnable(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	id := args[0]

	llm, err := database.GetLLM(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get LLM: %w", err)
	}

	llm.Enabled = true
	if err := database.UpdateLLM(ctx, llm); err != nil {
		return fmt.Errorf("failed to update LLM: %w", err)
	}

	fmt.Printf("%s✅ LLM provider enabled!%s\n", SuccessStyle, Reset)
	return nil
}

func runLLMDisable(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	id := args[0]

	llm, err := database.GetLLM(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get LLM: %w", err)
	}

	llm.Enabled = false
	if err := database.UpdateLLM(ctx, llm); err != nil {
		return fmt.Errorf("failed to update LLM: %w", err)
	}

	fmt.Printf("%s✅ LLM provider disabled!%s\n", SuccessStyle, Reset)
	return nil
}

func runLLMUpdate(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	id := args[0]
	reader := bufio.NewReader(os.Stdin)

	// Get the existing LLM
	llm, err := database.GetLLM(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get LLM: %w", err)
	}

	fmt.Printf("🔄 Update LLM Provider: %s\n", llm.Name)
	fmt.Println("================================")
	fmt.Println()

	// Show current configuration
	fmt.Printf("Current Configuration:\n")
	fmt.Printf("  Name: %s\n", llm.Name)
	fmt.Printf("  Provider: %s\n", llm.Provider)
	fmt.Printf("  Model: %s\n", llm.Model)
	fmt.Printf("  API Key: %s\n", maskAPIKey(llm.APIKey))
	fmt.Printf("  Base URL: %s\n", llm.BaseURL)
	fmt.Printf("  Enabled: %t\n", llm.Enabled)
	fmt.Println()

	// Update API Key
	provider := FromString(llm.Provider)
	apiKeyURL := provider.GetConsoleURL()
	if apiKeyURL != "" {
		fmt.Printf("Get API key from: %s\n", apiKeyURL)
	}
	fmt.Print("Enter new API key (press Enter to keep current): ")
	apiKey, _ := reader.ReadString('\n')
	apiKey = strings.TrimSpace(apiKey)
	if apiKey != "" {
		llm.APIKey = apiKey
	}

	// Update Base URL (only for Ollama and custom providers)
	if provider == Ollama || llm.Provider == "custom" {
		fmt.Print("Enter new base URL (press Enter to keep current): ")
		baseURL, _ := reader.ReadString('\n')
		baseURL = strings.TrimSpace(baseURL)
		if baseURL != "" {
			llm.BaseURL = baseURL
		}
	}

	// Update enabled status
	fmt.Print("Enable this LLM? (y/N): ")
	enabledStr, _ := reader.ReadString('\n')
	enabledStr = strings.TrimSpace(strings.ToLower(enabledStr))
	if enabledStr == "y" || enabledStr == "yes" {
		llm.Enabled = true
	} else if enabledStr == "n" || enabledStr == "no" {
		llm.Enabled = false
	}

	// Save the updated LLM
	if err := database.UpdateLLM(ctx, llm); err != nil {
		return fmt.Errorf("failed to update LLM: %w", err)
	}

	fmt.Println("\n✅ LLM provider updated successfully!")
	return nil
}

// maskAPIKey masks the API key for display (shows first 4 and last 4 characters)
func maskAPIKey(apiKey string) string {
	if apiKey == "" {
		return "(not set)"
	}
	if len(apiKey) <= 8 {
		return "***"
	}
	return apiKey[:4] + "..." + apiKey[len(apiKey)-4:]
}
