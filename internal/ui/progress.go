package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
)

// ProgressSpinner represents a loading spinner with status messages
type ProgressSpinner struct {
	spinner  spinner.Model
	message  string
	err      string
	done     bool
	program  *tea.Program
	quitting bool
}

// NewProgressSpinner creates and initializes a new progress spinner
func NewProgressSpinner() *ProgressSpinner {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = styleSpinner

	p := &ProgressSpinner{
		spinner: s,
	}
	
	p.program = tea.NewProgram(p)
	return p
}

// Start begins the spinner animation with the given message
func (p *ProgressSpinner) Start(message string) {
	p.message = message
	p.err = ""
	p.done = false
	p.quitting = false
	
	// Run the program in a goroutine
	go func() {
		if err := p.program.Start(); err != nil {
			fmt.Printf("Error running spinner: %v\n", err)
		}
	}()
}

// Success displays a success message and stops the spinner
func (p *ProgressSpinner) Success(message string) {
	p.message = message
	p.done = true
	p.err = ""
	p.quitting = true
	p.program.Quit()
}

// Error displays an error message and stops the spinner
func (p *ProgressSpinner) Error(s string) {
	p.err = s
	p.done = true
	p.quitting = true
	p.program.Quit()
}

// Init implements tea.Model
func (p *ProgressSpinner) Init() tea.Cmd {
	return p.spinner.Tick
}

// Update implements tea.Model
func (p *ProgressSpinner) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if p.quitting {
		return p, tea.Quit
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return p, tea.Quit
		}
	default:
		var cmd tea.Cmd
		p.spinner, cmd = p.spinner.Update(msg)
		return p, cmd
	}

	return p, nil
}

// View implements tea.Model
func (p *ProgressSpinner) View() string {
	if p.err != "" {
		return strings.Join([]string{
			styleHeading.Render("✗ Error"),
			styleError.Render("Error: " + p.err),
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
			p.spinner.View(),
			styleListItem.Render(p.message)),
	}, "\n")
}

// Retry begins the spinner animation with the given message
func (p *ProgressSpinner) Retry(message string) {
	p.message = message
	p.err = ""
	p.done = false
	p.quitting = false
	
	go func() {
		if err := p.program.Start(); err != nil {
			fmt.Printf("Error running spinner: %v\n", err)
		}
	}()
}
