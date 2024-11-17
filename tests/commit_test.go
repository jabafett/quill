package tests

import (
	"strings"
	"testing"

	"github.com/jabafett/quill/internal/prompts"
)

func TestGetCommitPrompt(t *testing.T) {
	testDiff := "diff --git a/test.go b/test.go\n+test content"
	prompt := prompts.GetCommitPrompt(testDiff)

	// Check that prompt contains essential elements
	requiredElements := []string{
		"conventional commit message",
		"<type>(<scope>): <description>",
		"feat:",
		"fix:",
		"docs:",
		testDiff,
	}

	for _, element := range requiredElements {
		if !strings.Contains(prompt, element) {
			t.Errorf("Expected prompt to contain '%s'", element)
		}
	}
} 