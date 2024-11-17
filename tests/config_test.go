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
	configPath := filepath.Join(tmpDir, ".quill.toml")

	configContent := `
[core]
default_provider = "gemini"
cache_ttl = "24h"
default_candidates = 2
max_diff_size = "1MB"
retry_attempts = 3

[providers.gemini]
model = "gemini-1.5-flash-002"
max_tokens = 4096
temperature = 0.3
retry_attempts = 3
candidate_count = 2
`

	err := os.WriteFile(configPath, []byte(configContent), 0644)
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

	// Verify core config values
	if cfg.Core.DefaultProvider != "gemini" {
		t.Errorf("Expected default provider 'gemini', got %s", cfg.Core.DefaultProvider)
	}

	if cfg.Core.DefaultCandidates != 2 {
		t.Errorf("Expected default candidates 2, got %d", cfg.Core.DefaultCandidates)
	}

	// Verify provider config
	provider, ok := cfg.Providers["gemini"]
	if !ok {
		t.Fatal("Gemini provider config not found")
	}

	if provider.Model != "gemini-1.5-flash-002" {
		t.Errorf("Expected model 'gemini-1.5-flash-002', got %s", provider.Model)
	}

	if provider.Temperature != 0.3 {
		t.Errorf("Expected temperature 0.3, got %f", provider.Temperature)
	}

	if provider.CandidateCount != 2 {
		t.Errorf("Expected candidate count 2, got %d", provider.CandidateCount)
	}
}

func TestProviderSpecificConfig(t *testing.T) {
	// Create temporary config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ".quill.toml")

	configContent := `
[core]
default_provider = "anthropic"

[providers.anthropic]
model = "claude-3-sonnet-20240229"
max_tokens = 4096
temperature = 0.3
retry_attempts = 3
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
