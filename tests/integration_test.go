package tests

import (
	"context"
	"strings"
	"testing"

	"github.com/jabafett/quill/internal/ai"
	"github.com/jabafett/quill/tests/mocks"
)

func TestGenerateWorkflow(t *testing.T) {
	mockProvider := &mocks.MockGeminiProvider{
		GenerateFunc: func(ctx context.Context, prompt string, opts ai.GenerateOptions) ([]string, error) {
			return []string{
				"feat(test): add new testing infrastructure",
				"test(core): expand test coverage",
			}, nil
		},
	}

	ctx := context.Background()
	testDiff := "diff --git a/test.go b/test.go\n+test content"

	responses, err := mockProvider.Generate(ctx, testDiff, ai.GenerateOptions{
		MaxCandidates: 2,
		Temperature:   &float32Value,
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
}

var float32Value float32 = 0.3

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

func TestGenerateWorkflowWithOptions(t *testing.T) {
	mockProvider := &mocks.MockGeminiProvider{
		GenerateFunc: func(ctx context.Context, prompt string, opts ai.GenerateOptions) ([]string, error) {
			// Verify options are passed correctly
			if opts.MaxCandidates != 3 || *opts.Temperature != 0.7 {
				t.Errorf("Expected MaxCandidates=3, Temperature=0.7, got %d, %f", 
					opts.MaxCandidates, *opts.Temperature)
			}
			return []string{"feat(test): test message"}, nil
		},
	}

	ctx := context.Background()
	temp := float32(0.7)
	_, err := mockProvider.Generate(ctx, "test diff", ai.GenerateOptions{
		MaxCandidates: 3,
		Temperature:   &temp,
	})

	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}
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
