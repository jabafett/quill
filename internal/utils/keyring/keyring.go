package keyring

import (
	"fmt"
	"os"

	"github.com/zalando/go-keyring"
)

const (
	ServiceName = "quill-ai"
)

// Provider represents an AI provider's keyring entry
type Provider struct {
	Name    string
	KeyName string
}

var (
	Gemini    = Provider{Name: "gemini", KeyName: "GEMINI_API_KEY"}
	Anthropic = Provider{Name: "anthropic", KeyName: "ANTHROPIC_API_KEY"}
	OpenAI    = Provider{Name: "openai", KeyName: "OPENAI_API_KEY"}
)

// StoreAPIKey stores an API key in the system keyring
// Falls back to environment variable if keyring is unavailable
func StoreAPIKey(provider Provider, apiKey string) error {
	err := keyring.Set(ServiceName, provider.KeyName, apiKey)
	if err != nil {
		// Fall back to environment variable
		return os.Setenv(provider.KeyName, apiKey)
	}
	return nil
}

// GetAPIKey retrieves an API key from the system keyring
// Falls back to environment variable if keyring is unavailable
func GetAPIKey(provider Provider) (string, error) {
	key, err := keyring.Get(ServiceName, provider.KeyName)
	if err != nil {
		// Fall back to environment variable
		key = os.Getenv(provider.KeyName)
		if key == "" {
			return "", fmt.Errorf("no API key found for %s in keyring or environment", provider.Name)
		}
		return key, nil
	}
	return key, nil
}

// DeleteAPIKey removes an API key from the system keyring
func DeleteAPIKey(provider Provider) error {
	err := keyring.Delete(ServiceName, provider.KeyName)
	if err != nil {
		// If keyring deletion fails, try to unset environment variable
		return os.Unsetenv(provider.KeyName)
	}
	return nil
} 