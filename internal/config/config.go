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
}

type AIProvider struct {
	Model          string  `mapstructure:"model"`
	MaxTokens      int     `mapstructure:"max_tokens"`
	Temperature    float32 `mapstructure:"temperature"`
	RetryAttempts  int     `mapstructure:"retry_attempts"`
	CandidateCount int     `mapstructure:"candidate_count"`
}

// LoadConfig loads and validates the configuration
func LoadConfig() (*Config, error) {
	// Set default locations
	viper.SetConfigName(".quill")
	viper.SetConfigType("toml")
	viper.AddConfigPath(".")
	viper.AddConfigPath(filepath.Join(os.Getenv("HOME"), ".config"))

	// Set defaults
	viper.SetDefault("core.default_command", "generate")
	viper.SetDefault("core.cache_ttl", "24h")
	viper.SetDefault("core.retry_attempts", 3)

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			return nil, fmt.Errorf("%w: run 'quill init' to create one", ErrNoConfig)
		}
		return nil, fmt.Errorf("failed to read config: %w", err)
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