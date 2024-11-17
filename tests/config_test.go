package tests

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jabafett/quill/internal/config"
)

func TestConfigLoading(t *testing.T) {
	// Create temporary config file
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, ".config")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("Failed to create config directory: %v", err)
	}

	// Try both config names
	configNames := []string{"quill.toml", ".quill.toml"}
	configPath := filepath.Join(configDir, configNames[0]) // Use first name by default

	// Read example config content
	exampleConfig, err := os.ReadFile("../example_quill.toml")
	if err != nil {
		t.Fatalf("Failed to read example config: %v", err)
	}

	err = os.WriteFile(configPath, exampleConfig, 0644)
	if err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	// Set config path for test
	os.Setenv("HOME", tmpDir)

	// Test loading config
	cfg, err := config.LoadConfig()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Verify core config values match example config
	if cfg.Core.DefaultProvider != "gemini" {
		t.Errorf("Expected default provider 'gemini', got %s", cfg.Core.DefaultProvider)
	}

	if cfg.Core.DefaultCandidates != 2 {
		t.Errorf("Expected default candidates 2, got %d", cfg.Core.DefaultCandidates)
	}

	// Verify all providers from example config
	expectedProviders := []string{"anthropic", "openai", "gemini", "ollama"}
	for _, provider := range expectedProviders {
		if _, ok := cfg.Providers[provider]; !ok {
			t.Errorf("Expected provider '%s' not found in config", provider)
		}
	}

	// Verify specific provider settings
	gemini := cfg.Providers["gemini"]
	if gemini.Model != "gemini-1.5-flash-002" {
		t.Errorf("Expected model 'gemini-1.5-flash', got %s", gemini.Model)
	}
	if gemini.Temperature != 0.3 {
		t.Errorf("Expected temperature 0.3, got %f", gemini.Temperature)
	}
}

func TestProviderSpecificConfig(t *testing.T) {
	// Create temporary config file
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, ".config")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("Failed to create config directory: %v", err)
	}

	configPath := filepath.Join(configDir, "quill.toml")

	configContent := `
[core]
default_provider = "anthropic"

[providers.anthropic]
model = "claude-3-sonnet-20240229"
max_tokens = 4096
temperature = 0.3
enable_retries = true
candidate_count = 2
`
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	os.Setenv("HOME", tmpDir)

	cfg, err := config.LoadConfig()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	provider, ok := cfg.Providers["anthropic"]
	if !ok {
		t.Fatal("Anthropic provider config not found")
	}

	if provider.Model != "claude-3-sonnet-20240229" {
		t.Errorf("Expected model 'claude-3-sonnet-20240229', got %s", provider.Model)
	}
}

// Add a test for alternate config name
func TestConfigLoadingAlternateName(t *testing.T) {
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, ".config")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("Failed to create config directory: %v", err)
	}

	// Use alternate config name
	configPath := filepath.Join(configDir, ".quill.toml")

	exampleConfig, err := os.ReadFile("../example_quill.toml")
	if err != nil {
		t.Fatalf("Failed to read example config: %v", err)
	}

	err = os.WriteFile(configPath, exampleConfig, 0644)
	if err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	os.Setenv("HOME", tmpDir)

	cfg, err := config.LoadConfig()
	if err != nil {
		t.Fatalf("Failed to load config with alternate name: %v", err)
	}

	if cfg.Core.DefaultProvider != "gemini" {
		t.Errorf("Failed to load correct config from alternate filename")
	}
}
