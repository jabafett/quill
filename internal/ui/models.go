package ui

import (
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	styleHeading = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#7B2CBF")).
		MarginBottom(1)

	styleError = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#E63946")).
		Bold(true)

	styleSuccess = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#2ECC71")).
		Bold(true)
)

type ProcessModel struct {
	spinner  spinner.Model
	progress progress.Model
	err      error
	done     bool
	message  string
}

func NewProcessModel() ProcessModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	p := progress.New(progress.WithDefaultGradient())

	return ProcessModel{
		spinner:  s,
		progress: p,
	}
}

func (m ProcessModel) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m ProcessModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "q" || msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	case error:
		m.err = msg
		return m, tea.Quit
	}

	var cmd tea.Cmd
	m.spinner, cmd = m.spinner.Update(msg)
	return m, cmd
}

func (m ProcessModel) View() string {
	if m.err != nil {
		return styleError.Render("Error: " + m.err.Error())
	}
	if m.done {
		return styleSuccess.Render("✓ " + m.message)
	}
	return m.spinner.View() + " " + m.message
} 