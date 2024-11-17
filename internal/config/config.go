package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/viper"
)

// Config validation errors
var (
	ErrNoConfig        = fmt.Errorf("no configuration file found")
	ErrInvalidConfig   = fmt.Errorf("invalid configuration")
	ErrNoProvider      = fmt.Errorf("no AI provider configured")
	ErrInvalidProvider = fmt.Errorf("invalid provider configuration")
)

type Config struct {
	Core      CoreConfig             `mapstructure:"core"`
	Providers map[string]AIProvider  `mapstructure:"providers"`
}

type CoreConfig struct {
	DefaultProvider string        `mapstructure:"default_provider"`
	CacheTTL       time.Duration `mapstructure:"cache_ttl"`
	MaxDiffSize    string        `mapstructure:"max_diff_size"`
	DefaultCandidates int        `mapstructure:"default_candidates"`
}

type AIProvider struct {
	Model          string  `mapstructure:"model"`
	MaxTokens      int     `mapstructure:"max_tokens"`
	Temperature    float32 `mapstructure:"temperature"`
	EnableRetries  bool    `mapstructure:"enable_retries"`
	CandidateCount int     `mapstructure:"candidate_count"`
}

// LoadConfig loads and validates the configuration
func LoadConfig() (*Config, error) {
	// Get user's home directory
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	// Set config paths and names
	viper.SetConfigType("toml")
	
	// Add all possible config paths and names
	viper.AddConfigPath(".")                           // current directory
	viper.AddConfigPath(filepath.Join(home, ".config")) // ~/.config/
	viper.AddConfigPath(home)                          // home directory

	// Try both config names
	configNames := []string{"quill", ".quill"}
	var configFound bool

	for _, name := range configNames {
		viper.SetConfigName(name)
		if err := viper.ReadInConfig(); err == nil {
			configFound = true
			break
		}
	}

	if !configFound {
		return nil, fmt.Errorf("%w: run 'quill init' to create one", ErrNoConfig)
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidConfig, err)
	}

	if err := validateConfig(&config); err != nil {
		return nil, err
	}

	return &config, nil
}

func validateConfig(cfg *Config) error {
	if cfg.Core.DefaultProvider == "" {
		return fmt.Errorf("%w: default provider not set", ErrNoProvider)
	}

	if _, ok := cfg.Providers[cfg.Core.DefaultProvider]; !ok {
		return fmt.Errorf("%w: provider '%s' not configured", ErrInvalidProvider, cfg.Core.DefaultProvider)
	}

	return nil
} 