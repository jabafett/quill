package ui

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

type ProgressSpinner struct {
	model ProcessModel
}

func NewProgressSpinner() *ProgressSpinner {
	return &ProgressSpinner{
		model: NewProcessModel(),
	}
}

func (s *ProgressSpinner) Start(message string) {
	s.model.message = message
	p := tea.NewProgram(s.model)
	go p.Run()
}

func (s *ProgressSpinner) Stop() {
	s.model.done = true
	fmt.Println()
}

func (s *ProgressSpinner) Error(err error) {
	s.model.err = err
	fmt.Fprintln(os.Stderr, styleError.Render("Error: "+err.Error()))
}

func (s *ProgressSpinner) Success(message string) {
	s.model.done = true
	s.model.message = message
	fmt.Println(styleSuccess.Render("✓ " + message))
} 