package tests

import (
	"fmt"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/jabafett/quill/internal/ui"
)

// Package tests provides testing for the UI components
func TestCommitMessageModel(t *testing.T) {
	messages := []string{
		"feat(test): add new feature",
		"fix(bug): resolve issue",
	}

	model := ui.NewCommitMessageModel(messages)

	// Initialize the model with a window size
	model.Init()
	newModel, _ := model.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	model = newModel.(ui.CommitMessageModel)

	// Test initial state
	view := model.View()
	for _, msg := range messages {
		if !strings.Contains(view, msg) {
			t.Errorf("Expected view to contain message: %s", msg)
		}
	}

	// Test key handling
	testCases := []struct {
		name     string
		msg      tea.Msg
		contains string
	}{
		{
			name: "quit",
			msg:  tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}},
		},
		{
			name: "edit",
			msg:  tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}},
		},
		{
			name: "enter",
			msg:  tea.KeyMsg{Type: tea.KeyEnter},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			newModel, _ := model.Update(tc.msg)
			updatedModel := newModel.(ui.CommitMessageModel)
			view := updatedModel.View()
			if tc.contains != "" && !strings.Contains(view, tc.contains) {
				t.Errorf("Expected view to contain %q", tc.contains)
			}
		})
	}
}

// TestProgressSpinner verifies the behavior of the ProgressSpinner
func TestProgressSpinner(t *testing.T) {
	spinner := ui.NewProgressSpinner()

	// Test spinner messages
	testMessage := "Testing progress"
	spinner.Start(testMessage)
	view := spinner.View()
	if !strings.Contains(view, testMessage) {
		t.Errorf("Expected spinner view to contain message: %s", testMessage)
	}

	// Test success message
	successMessage := "Operation completed"
	spinner.Success(successMessage)
	view = spinner.View()
	if !strings.Contains(view, successMessage) {
		t.Errorf("Expected spinner view to contain success message: %s", successMessage)
	}

	// Test error handling
	testError := fmt.Errorf("test error")
	spinner.Error(testError)
	view = spinner.View()
	if !strings.Contains(view, testError.Error()) {
		t.Errorf("Expected spinner view to contain error: %s", testError)
	}
}

// Add test for editing functionality
func TestCommitMessageEditing(t *testing.T) {
	messages := []string{"feat(test): initial message"}
	model := ui.NewCommitMessageModel(messages)

	// Simulate edit key press
	newModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})
	updatedModel := newModel.(ui.CommitMessageModel)

	if !updatedModel.IsEditing() {
		t.Error("Expected model to be in editing mode")
	}
}

// Add test for progress spinner with retry
func TestProgressSpinnerRetry(t *testing.T) {
	spinner := ui.NewProgressSpinner()

	retryMessage := "Retrying operation..."
	spinner.Retry(retryMessage)

	view := spinner.View()
	if !strings.Contains(view, retryMessage) {
		t.Errorf("Expected spinner view to contain retry message: %s", retryMessage)
	}
}
