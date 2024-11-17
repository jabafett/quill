package ui

import (
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type commitMessage struct {
	message string
}

func (c commitMessage) Title() string       { return c.message }
func (c commitMessage) Description() string { return "" }
func (c commitMessage) FilterValue() string { return c.message }

type CommitMessageModel struct {
	list     list.Model
	selected string
	quitting bool
}

func NewCommitMessageModel(messages []string) CommitMessageModel {
	items := make([]list.Item, len(messages))
	for i, msg := range messages {
		items[i] = commitMessage{message: msg}
	}

	l := list.New(items, list.NewDefaultDelegate(), 0, 0)
	l.Title = "Select a commit message"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Styles.Title = styleHeading

	return CommitMessageModel{
		list: l,
	}
}

func (m CommitMessageModel) Init() tea.Cmd {
	return nil
}

func (m CommitMessageModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		case "enter":
			if i, ok := m.list.SelectedItem().(commitMessage); ok {
				m.selected = i.message
			}
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		h, v := lipgloss.NewStyle().Margin(2, 2).GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m CommitMessageModel) View() string {
	if m.quitting {
		return "Operation cancelled\n"
	}
	if m.selected != "" {
		return styleSuccess.Render("Selected commit message:\n") + m.selected + "\n"
	}
	return m.list.View()
} 