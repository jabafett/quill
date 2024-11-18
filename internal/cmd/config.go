package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jabafett/quill/internal/utils/keyring"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage Quill configuration",
	Long: `Manage Quill's configuration settings and API keys.

Available Commands:
  get      - Retrieve the value of a specific configuration setting
  set      - Modify a configuration setting
  set-key  - Securely store an API key for an AI provider
  get-key  - Retrieve a stored API key for an AI provider
  list     - Display all current configuration settings

Configuration is stored in ~/.config/quill.toml and API keys are securely stored
in your system's keyring.`,
}

var getCmd = &cobra.Command{
	Use:   "get [key]",
	Short: "Get configuration value",
	Long: `Retrieve the value of a specific configuration setting.
    
The key should be provided in dot notation for nested settings. For example:
  quill config get core.default_provider
  quill config get providers.gemini.temperature`,
	Args:  cobra.ExactArgs(1),
	RunE:  runGet,
}

var setCmd = &cobra.Command{
	Use:   "set [key] [value]",
	Short: "Set configuration value",
	Long: `Modify a configuration setting with a new value.
    
Use dot notation to access nested settings. For example:
  quill config set core.default_provider gemini
  quill config set providers.gemini.temperature 0.7

Changes are immediately written to ~/.config/quill.toml`,
	Args:  cobra.ExactArgs(2),
	RunE:  runSet,
}

var setKeyCmd = &cobra.Command{
	Use:   "set-key [provider] [api-key]",
	Short: "Set API key for a provider",
	Long: `Securely store an API key for an AI provider in your system's keyring.
    
Supported providers:
- gemini   (Google's Gemini API)
- anthropic (Claude API)
- openai    (OpenAI API)

The API key is stored securely in your system's keyring/keychain and is never
written to disk in plaintext.`,
	Args:  cobra.ExactArgs(2),
	RunE:  runSetKey,
}

var getKeyCmd = &cobra.Command{
	Use:   "get-key [provider]",
	Short: "Get API key for a provider",
	Long: `Retrieve a stored API key for an AI provider from your system's keyring.
    
Supported providers:
- gemini
- anthropic
- openai

The key will be displayed in plaintext - use with caution.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runGetKey,
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all configuration settings",
	Long: `Display all current configuration settings in a hierarchical format.
    
This command shows the complete contents of your ~/.config/quill.toml file,
excluding sensitive information like API keys which are stored separately
in your system's keyring.`,
	Args:  cobra.NoArgs,
	RunE:  runList,
}

func init() {
	configCmd.AddCommand(getCmd)
	configCmd.AddCommand(setCmd)
	configCmd.AddCommand(setKeyCmd)
	configCmd.AddCommand(getKeyCmd)
	configCmd.AddCommand(listCmd)
}

func runGet(cmd *cobra.Command, args []string) error {
	key := args[0]
	value := viper.Get(key)
	if value == nil {
		return fmt.Errorf("key '%s' not found in configuration", key)
	}
	fmt.Printf("%v\n", value)
	return nil
}

func runSet(cmd *cobra.Command, args []string) error {
	key := args[0]
	value := args[1]

	// Load existing config
	configFile := filepath.Join(os.Getenv("HOME"), ".config", "quill.toml")
	viper.SetConfigFile(configFile)
	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("failed to read config: %w", err)
	}

	// Set new value
	viper.Set(key, value)

	// Save config
	if err := viper.WriteConfig(); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	fmt.Printf("Set %s = %s\n", key, value)
	return nil
}

func runSetKey(cmd *cobra.Command, args []string) error {
	provider := strings.ToLower(args[0])
	apiKey := args[1]

	var kp keyring.Provider
	switch provider {
	case "gemini":
		kp = keyring.Gemini
	case "anthropic":
		kp = keyring.Anthropic
	case "openai":
		kp = keyring.OpenAI
	default:
		return fmt.Errorf("unknown provider: %s", provider)
	}

	if err := keyring.StoreAPIKey(kp, apiKey); err != nil {
		return fmt.Errorf("failed to store API key: %w", err)
	}

	fmt.Printf("API key for %s has been stored securely\n", provider)
	return nil
}

func runGetKey(cmd *cobra.Command, args []string) error {
	provider := strings.ToLower(args[0])

	var kp keyring.Provider
	switch provider {
	case "gemini":
		kp = keyring.Gemini
	case "anthropic":
		kp = keyring.Anthropic
	case "openai":
		kp = keyring.OpenAI
	default:
		return fmt.Errorf("unknown provider: %s", provider)
	}

	key, err := keyring.GetAPIKey(kp)
	if err != nil {
		return err
	}

	fmt.Printf("%s\n", key)
	return nil
}

func runList(cmd *cobra.Command, args []string) error {
	settings := viper.AllSettings()
	printSettings(settings, 0)
	return nil
}

func printSettings(settings map[string]interface{}, indent int) {
	indentStr := strings.Repeat("  ", indent)
	for k, v := range settings {
		if subMap, ok := v.(map[string]interface{}); ok {
			fmt.Printf("%s%s:\n", indentStr, k)
			printSettings(subMap, indent+1)
		} else {
			fmt.Printf("%s%s: %v\n", indentStr, k, v)
		}
	}
} 