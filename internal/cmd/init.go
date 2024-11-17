package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize Quill configuration",
	RunE:  runInit,
}

const defaultConfig = `[core]
default_command = "generate"
default_provider = "gemini"
cache_ttl = "24h"
retry_attempts = 3
candidate_count = 3

[providers]
  [providers.gemini]
    model = "gemini-1.5-flash"
    max_tokens = 4096
    temperature = 0.3
`

func runInit(cmd *cobra.Command, args []string) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	configDir := filepath.Join(homeDir, ".quill")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	configPath := filepath.Join(configDir, "config.toml")
	if _, err := os.Stat(configPath); err == nil {
		return fmt.Errorf("configuration file already exists at %s", configPath)
	}

	if err := os.WriteFile(configPath, []byte(defaultConfig), 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	fmt.Printf("Configuration file created at %s\n", configPath)
	fmt.Println("Please set your API keys as environment variables:")
	fmt.Println("export GEMINI_API_KEY=your_api_key")
	
	return nil
} 