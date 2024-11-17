package tests

import (
	"os"
	"testing"

	"github.com/jabafett/quill/internal/keyring"
)

func TestKeyringFallback(t *testing.T) {
	// Clear any existing environment variables
	os.Unsetenv("GEMINI_API_KEY")
	defer os.Unsetenv("GEMINI_API_KEY")

	testKey := "test-api-key"
	provider := keyring.Provider{
		Name:    "test",
		KeyName: "GEMINI_API_KEY",
	}

	// Set environment variable
	os.Setenv(provider.KeyName, testKey)

	// Get key (should fall back to env var)
	key, err := keyring.GetAPIKey(provider)
	if err != nil {
		t.Fatalf("Failed to get API key: %v", err)
	}

	if key != testKey {
		t.Errorf("Expected key %s, got %s", testKey, key)
	}
} 