package tests

import (
	"strings"
	"testing"

	"github.com/jabafett/quill/internal/ui"
)

func TestUIStyles(t *testing.T) {
	// Test spinner view with different states
	spinner := ui.NewProgressSpinner()

	// Test normal state
	spinner.Start("Testing...")
	view := spinner.View()
	if !strings.Contains(view, "Testing...") {
		t.Error("Spinner view should contain message")
	}

	// Test success state
	spinner.Success("Done!")
	view = spinner.View()
	if !strings.Contains(view, "✓") || !strings.Contains(view, "Done!") {
		t.Error("Success view should contain checkmark and message")
	}

	// Test error state
	spinner.Error("test error")
	view = spinner.View()
	if !strings.Contains(view, "✗") || !strings.Contains(view, "test error") {
		t.Error("Error view should contain X and error message")
	}
} 