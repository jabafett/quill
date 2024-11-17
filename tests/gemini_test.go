package ai

import (
	"context"
	"strings"
	"testing"

	"github.com/jabafett/quill/internal/ai"
)

func TestGeminiProvider_GenerateCommitMessage(t *testing.T) {
	// Create a provider with test API key
	provider, err := ai.NewGeminiProvider(ai.Options{
		APIKey:      "AIzaSyDxoYCe5cTP1yjHPCnzRLo5LzsfM744AQA",
		Model:       "gemini-1.5-flash",
		Temperature: 0.7,
	})
	if err != nil {
		t.Fatalf("Failed to create provider: %v", err)
	}

	// Sample git diff
	testDiff := `diff --git a/test.txt b/test.txt
+ This is a fix test line
- This is a fix removed line`

	// Test the generate commit message
	msg, err := provider.GenerateCommitMessage(context.Background(), testDiff)
	if err != nil {
		t.Fatalf("Failed to generate commit message: %v", err)
	}

	// Basic validation
	if msg == "" {
		t.Error("Generated commit message is empty")
	}

	// Check if the message follows conventional commit format
	if !strings.HasPrefix(msg, "fix") && !strings.HasPrefix(msg, "feat") {
		t.Errorf("Generated commit message does not follow conventional format: %s", msg)
	}

	// Check message length
	if len(strings.Split(msg, "\n")[0]) > 72 {
		t.Error("First line of commit message exceeds 72 characters")
	}
}
