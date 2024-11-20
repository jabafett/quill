package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jabafett/quill/internal/utils/debug"
	"github.com/jabafett/quill/internal/utils/keyring"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize Quill configuration",
	Long: `Initialize Quill by creating a default configuration file and setting up API keys.

This command will:
1. Create ~/.config/quill.toml with default settings
2. Set up the configuration directory if it doesn't exist
3. Configure your preferred AI provider and model
4. Store your API key securely

The configuration includes:
- Default AI provider settings
- API configuration for supported providers
- Cache and retry settings
- Token and temperature parameters
- Model specifications`,
	Example: `  quill init`,
	RunE:    runInit,
}

func PromptForProvider() (string, error) {
	fmt.Println("Select an AI provider:")
	fmt.Println("1. Google")
	fmt.Println("2. Anthropic")
	fmt.Println("3. OpenAI")
	fmt.Println("4. Ollama")

	var input string
	fmt.Print("Enter choice (1-4): ")
	fmt.Scanln(&input)

	switch input {
	case "1":
		return "gemini", nil
	case "2":
		return "anthropic", nil
	case "3":
		return "openai", nil
	case "4":
		return "ollama", nil
	default:
		return "", fmt.Errorf("invalid choice: %s (must be 1-4)", input)
	}
}

func PromptForAPIKey(provider string) (string, error) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("Enter API key for %s: ", provider)
	apiKey, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("failed to read API key: %w", err)
	}
	return strings.TrimSpace(apiKey), nil
}

func GetProviderConfig(provider string) string {
	switch provider {
	case "gemini":
		return `    model = "gemini-1.5-flash-002"
    max_tokens = 8192
    temperature = 0.5
    enable_retries = false
    candidate_count = 2`
	case "anthropic":
		return `    model = "claude-3-sonnet-20240229"
    max_tokens = 8192
    temperature = 0.5
    enable_retries = false
    candidate_count = 2`
	case "openai":
		return `    model = "gpt-4"
    max_tokens = 8192
    temperature = 0.5
    enable_retries = false
    candidate_count = 2`
	case "ollama":
		return `    model = "qwen2.5-8b-instruct"
    max_tokens = 8192
    temperature = 0.5
    enable_retries = true
    candidate_count = 3`
	default:
		return ""
	}
}

func GenerateConfig(selectedProvider string) string {
	return fmt.Sprintf(`[core]
# Cache TTL duration
cache_ttl = "24h"
# Number of retry attempts for API calls
retry_attempts = 3
# Default number of candidates to generate (0-3)
default_candidates = 2
# Maximum diff size for processing
max_diff_size = "500MB"
# Default provider
default_provider = "%s"

[providers]
  [providers.%s]
%s
`, selectedProvider, selectedProvider, GetProviderConfig(selectedProvider))
}

func runInit(cmd *cobra.Command, args []string) error {
	debug.Log("Starting init command")

	// Get provider selection
	selectedProvider, err := PromptForProvider()
	if err != nil {
		return fmt.Errorf("invalid provider selection: %w", err)
	}

	// Get API key if provider is not Ollama
	if selectedProvider != "ollama" {
		apiKey, err := PromptForAPIKey(selectedProvider)
		if err != nil {
			return err
		}

		var kp keyring.Provider
		switch selectedProvider {
		case "gemini":
			kp = keyring.Gemini
		case "anthropic":
			kp = keyring.Anthropic
		case "openai":
			kp = keyring.OpenAI
		}

		if err := keyring.StoreAPIKey(kp, apiKey); err != nil {
			return fmt.Errorf("failed to store API key: %w", err)
		}
		cmd.Printf("API key for %s has been stored securely\n", selectedProvider)
	}

	// Create config directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	configDir := filepath.Join(homeDir, ".config")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	configPath := filepath.Join(configDir, "quill.toml")
	debug.Log("Using config path: %s", configPath)

	// Check if config already exists
	if _, err := os.Stat(configPath); err == nil {
		if err := os.Remove(configPath); err != nil {
			return fmt.Errorf("failed to remove existing config: %w", err)
		}
		cmd.Printf("Removed existing config at %s\n", configPath)
	}

	// Generate and write config
	config := GenerateConfig(selectedProvider)
	if err := os.WriteFile(configPath, []byte(config), 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	cmd.Printf("Configuration file created at %s\n", configPath)
	return nil
}

// GetInitCmd exposes the init command for testing
func GetInitCmd() *cobra.Command {
	return initCmd
}
