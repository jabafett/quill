package tests

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/jabafett/quill/internal/factories"
	"github.com/jabafett/quill/internal/utils/ai"
	"github.com/jabafett/quill/tests/mocks"
)

func TestGenerateWorkflow(t *testing.T) {
	mock := &mocks.MockGeminiProvider{
		GenerateFunc: func(ctx context.Context, prompt string, opts ai.GenerateOptions) ([]string, error) {
			return []string{
				"feat(test): add new testing infrastructure",
				"test(core): expand test coverage",
			}, nil
		},
	}

	provider := factories.GetRateLimitedProvider(mock, true)

	ctx := context.Background()
	testDiff := "diff --git a/test.go b/test.go\n+test content"

	responses, err := provider.Generate(ctx, testDiff, ai.GenerateOptions{
		MaxCandidates: 2,
	})

	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	if len(responses) != 2 {
		t.Errorf("Expected 2 responses, got %d", len(responses))
	}

	// Verify response format
	for _, resp := range responses {
		if !isValidCommitMessage(resp) {
			t.Errorf("Invalid commit message format: %s", resp)
		}
	}

	// Verify rate limiting
	start := time.Now()
	_, err = provider.Generate(ctx, testDiff, ai.GenerateOptions{})
	if err != nil {
		t.Fatalf("Second generate call failed: %v", err)
	}

	duration := time.Since(start)
	if duration < time.Second {
		t.Error("Rate limiting not enforced")
	}
}

func isValidCommitMessage(msg string) bool {
	// Verify conventional commit format
	parts := strings.Split(msg, ":")
	if len(parts) != 2 {
		return false
	}

	// Check type and scope
	typeScope := parts[0]
	if !strings.Contains(typeScope, "feat") &&
		!strings.Contains(typeScope, "fix") &&
		!strings.Contains(typeScope, "docs") &&
		!strings.Contains(typeScope, "style") &&
		!strings.Contains(typeScope, "refactor") &&
		!strings.Contains(typeScope, "test") &&
		!strings.Contains(typeScope, "chore") {
		return false
	}

	// Check description
	description := strings.TrimSpace(parts[1])
	return len(description) > 0 && description[len(description)-1] != '.'
}

func TestBreakingChangeCommitMessage(t *testing.T) {
	messages := []string{
		"feat(api)!: remove deprecated endpoints",
		"feat(core): add new feature",
	}

	for _, msg := range messages {
		if !isValidCommitMessage(msg) {
			t.Errorf("Expected valid commit message: %s", msg)
		}
	}
}
