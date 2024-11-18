package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
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

type keyMap struct {
	Up     key.Binding
	Down   key.Binding
	Enter  key.Binding
	Quit   key.Binding
	Edit   key.Binding
	Reload key.Binding
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Up, k.Down, k.Enter, k.Edit, k.Quit}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Enter},
		{k.Edit, k.Reload, k.Quit},
	}
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
	list     list.Model
	keys     keyMap
	selected string
	quitting bool
	editing  bool
}

func NewCommitMessageModel(messages []string) CommitMessageModel {
	items := make([]list.Item, len(messages))
	for i, msg := range messages {
		items[i] = commitMessage{
			message: msg,
		}
	}

	delegate := list.NewDefaultDelegate()
	delegate.Styles.SelectedTitle = styleSelectedItem
	delegate.Styles.NormalTitle = styleListItem

	l := list.New(items, delegate, 0, 0)
	l.Title = fmt.Sprintf("Select from %d commit message variations", len(messages))
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Styles.Title = styleListTitle
	l.AdditionalFullHelpKeys = func() []key.Binding {
		return []key.Binding{
			keys.Edit,
			keys.Reload,
		}
	}

	return CommitMessageModel{
		list: l,
		keys: keys,
	}
}

func (m CommitMessageModel) Init() tea.Cmd {
	return nil
}

func (m CommitMessageModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		// Quit the application
		case key.Matches(msg, m.keys.Quit):
			m.quitting = true
			return m, tea.Quit
		// Select the selected commit message
		case key.Matches(msg, m.keys.Enter):
			if i, ok := m.list.SelectedItem().(commitMessage); ok {
				m.selected = i.message
			}
			return m, tea.Quit
		// Edit the selected commit message
		case key.Matches(msg, m.keys.Edit):
			m.editing = true
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
		return styleError.Render("Operation cancelled") + "\n"
	}

	if m.selected != "" {
		var status string
		if m.editing {
			status = styleHeading.Render("✎ Edit Commit Message")
			status += "\n" + styleListItem.Render("Edit and use the following commit message:") + "\n\n"
		} else {
			status = styleHeading.Render("✓ Selected Commit Message")
			status += "\n" + styleListItem.Render("Selected commit message:") + "\n\n"
		}
		return status + styleSelectedItem.Render(m.selected) + "\n"
	}

	help := strings.Join([]string{
		"↑/↓: navigate",
		"enter: select",
		"e: edit",
		"r: reload",
		"q: quit",
	}, " • ")

	return strings.Join([]string{
		styleHeading.Render("✨ Select Commit Message"),
		m.list.View(),
		styleHelp.Render(help),
	}, "\n")
}

func (m CommitMessageModel) IsEditing() bool {
	return m.editing
}

func (m CommitMessageModel) Selected() string {
	return m.selected
} 