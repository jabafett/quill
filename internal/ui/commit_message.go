package ui

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type keyMap struct {
	Up     key.Binding
	Down   key.Binding
	Enter  key.Binding
	Quit   key.Binding
	Edit   key.Binding
	Reload key.Binding
}

var keys = keyMap{
	Up: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("↑/k", "up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓/j", "down"),
	),
	Enter: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "select"),
	),
	Edit: key.NewBinding(
		key.WithKeys("e"),
		key.WithHelp("e", "edit"),
	),
	Reload: key.NewBinding(
		key.WithKeys("r"),
		key.WithHelp("r", "reload"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "esc", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
}

type CommitMessageModel struct {
	messages []string
	cursor   int
	input    textarea.Model
	keys     keyMap
	selected string
	quitting bool
	editing  bool
	width    int
	height   int
}

func NewCommitMessageModel(messages []string) CommitMessageModel {
	ta := textarea.New()
	ta.Placeholder = "Edit commit message..."
	ta.SetWidth(50)
	ta.SetHeight(5)
	ta.ShowLineNumbers = true
	ta.Prompt = "┃ "
	ta.FocusedStyle.CursorLine = lipgloss.NewStyle().Background(lipgloss.Color("#333333"))

	return CommitMessageModel{
		messages: messages,
		input:    ta,
		keys:     keys,
	}
}

func (m CommitMessageModel) Init() tea.Cmd {
	return textarea.Blink
}

func (m CommitMessageModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.input.SetWidth(m.width - 4)
		return m, nil

	case tea.KeyMsg:
		if m.editing {
			switch msg.String() {
			case "esc":
				m.editing = false
				m.input.Blur()
				return m, nil
			case "enter":
				m.selected = m.input.Value()
				return m, tea.Quit
			default:
				var cmd tea.Cmd
				m.input, cmd = m.input.Update(msg)
				return m, cmd
			}
		}

		switch {
		case key.Matches(msg, m.keys.Quit):
			m.quitting = true
			return m, tea.Quit
		case key.Matches(msg, m.keys.Up):
			if m.cursor > 0 {
				m.cursor--
			}
		case key.Matches(msg, m.keys.Down):
			if m.cursor < len(m.messages)-1 {
				m.cursor++
			}
		case key.Matches(msg, m.keys.Enter):
			m.selected = m.messages[m.cursor]
			return m, tea.Quit
		case key.Matches(msg, m.keys.Edit):
			m.editing = true
			m.input.SetValue(m.messages[m.cursor])
			m.input.Focus()
			return m, textarea.Blink
		}
	}

	return m, nil
}

func (m CommitMessageModel) View() string {
	if m.quitting {
		return styleError.Render("Operation cancelled")
	}

	mainStyle := lipgloss.NewStyle().Padding(1, 2)

	if m.editing {
		return mainStyle.Render(
			lipgloss.JoinVertical(lipgloss.Left,
				styleHeading.Render("✎ Edit Commit Message"),
				styleInput.Render(m.input.View()),
				styleHelp.Render("enter: save • esc: cancel"),
			),
		)
	}

	if m.selected != "" {
		return mainStyle.Render(
			lipgloss.JoinVertical(lipgloss.Left,
				styleHeading.Render("✓ Selected Commit Message"),
				styleSelectedItem.Render(m.selected),
				styleHelp.Render("Press enter to confirm"),
			),
		)
	}

	// Render messages
	var items []string
	for i, msg := range m.messages {
		if i == m.cursor {
			items = append(items, styleSelectedItem.Render(msg))
		} else {
			items = append(items, styleListItem.Render(msg))
		}
	}

	help := lipgloss.JoinHorizontal(lipgloss.Center,
		"↑/↓: navigate",
		" • ",
		"enter: select",
		" • ",
		"e: edit",
		" • ",
		"q: quit",
	)

	return mainStyle.Render(
		lipgloss.JoinVertical(lipgloss.Left,
			styleHeading.Render("✨ Select Commit Message"),
			lipgloss.JoinVertical(lipgloss.Left, items...),
			styleHelp.Render(help),
		),
	)
}

func (m CommitMessageModel) IsEditing() bool {
	return m.editing
}

func (m CommitMessageModel) Selected() string {
	return m.selected
}

func (m CommitMessageModel) Quitting() bool {
	return m.quitting
}