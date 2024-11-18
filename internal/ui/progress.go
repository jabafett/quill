package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
)



// ProgressSpinner represents a loading spinner with status messages
type ProgressSpinner struct {
	spinner spinner.Model
	message string
	err     error
	done    bool
}

// NewProgressSpinner creates and initializes a new progress spinner
func NewProgressSpinner() *ProgressSpinner {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = styleSpinner

	return &ProgressSpinner{
		spinner: s,
	}
}

// Start begins the spinner animation with the given message
func (p *ProgressSpinner) Start(message string) {
	p.message = message
	p.err = nil
	p.done = false
}

// Success displays a success message and stops the spinner
func (p *ProgressSpinner) Success(message string) {
	p.message = message
	p.done = true
	p.err = nil
}

// Error displays an error message and stops the spinner
func (p *ProgressSpinner) Error(err error) {
	p.err = err
	p.done = true
}

// View returns the current view of the spinner
func (p *ProgressSpinner) View() string {
	if p.err != nil {
		return strings.Join([]string{
			styleHeading.Render("✗ Error"),
			styleError.Render("Error: " + p.err.Error()),
		}, "\n")
	}
	if p.done {
		return strings.Join([]string{
			styleHeading.Render("✓ Success"),
			styleSuccess.Render(p.message),
		}, "\n")
	}
	return strings.Join([]string{
		styleHeading.Render("⧗ Progress"),
		fmt.Sprintf("%s %s",
			styleSpinner.Render(p.spinner.View()),
			styleListItem.Render(p.message)),
	}, "\n")
}

// Update handles spinner animation updates
func (p *ProgressSpinner) Update(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	p.spinner, cmd = p.spinner.Update(msg)
	return cmd
}

// Retry begins the spinner animation with the given message
func (p *ProgressSpinner) Retry(message string) {
	p.message = message
	p.err = nil
	p.done = false
}
