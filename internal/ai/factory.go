package ai

import (
	"fmt"
	"os"
	"strings"

	"github.com/jabafett/quill/internal/config"
)

// NewProvider creates a new AI provider based on the provider name and configuration
func NewProvider(providerName string, cfg *config.Config) (Provider, error) {
	providerCfg, ok := cfg.Providers[providerName]
	if !ok {
		return nil, fmt.Errorf("provider %s not found in configuration", providerName)
	}

	// Get API key from environment variable
	apiKey := os.Getenv(fmt.Sprintf("%s_API_KEY", strings.ToUpper(providerName)))
	if apiKey == "" {
		return nil, fmt.Errorf("API key not found in environment variables for provider %s", providerName)
	}

	options := Options{
		Model:       providerCfg.Model,
		MaxTokens:   providerCfg.MaxTokens,
		Temperature: providerCfg.Temperature,
		APIKey:      apiKey,
	}

	switch providerName {
	case "gemini":
		return NewGeminiProvider(options)
	// Add other providers here as they're implemented
	default:
		return nil, fmt.Errorf("unsupported provider: %s", providerName)
	}
} 