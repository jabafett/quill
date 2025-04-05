package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/jabafett/quill/internal/utils/helpers"
)

// Key map for suggest UI
type suggestKeyMap struct {
	Up      key.Binding
	Down    key.Binding
	Enter   key.Binding
	Quit    key.Binding
	Edit    key.Binding
	Reload  key.Binding
	Stage   key.Binding
	Unstage key.Binding
	Back    key.Binding
	Help    key.Binding
}

var suggestKeys = suggestKeyMap{
	Up: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("↑/k", "move up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓/j", "move down"),
	),
	Enter: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "select/apply"),
	),
	Edit: key.NewBinding(
		key.WithKeys("e"),
		key.WithHelp("e", "edit commit message"),
	),
	Reload: key.NewBinding(
		key.WithKeys("r"),
		key.WithHelp("r", "regenerate suggestions"),
	),
	Stage: key.NewBinding(
		key.WithKeys("s"),
		key.WithHelp("s", "mark for staging"),
	),
	Unstage: key.NewBinding(
		key.WithKeys("u"),
		key.WithHelp("u", "unmark for staging"),
	),
	Back: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "cancel edit"),
	),
	Help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "toggle help"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
}

type SuggestionItem struct {
	suggestion helpers.SuggestionGroup
	index      int
}

func (i SuggestionItem) Title() string       { return i.suggestion.Description }
func (i SuggestionItem) Description() string {
	count := len(i.suggestion.Files)
	if count == 0 {
		return "No files"
	}
	if count == 1 {
		return fmt.Sprintf("1 file: %s", i.suggestion.Files[0])
	}
	return fmt.Sprintf("%d files: %s, ...", count, i.suggestion.Files[0])
}
func (i SuggestionItem) FilterValue() string { return i.suggestion.Description }

type SuggestModel struct {
	suggestions        []helpers.SuggestionGroup
	list               list.Model
	input              textarea.Model
	keys               suggestKeyMap
	selected           *helpers.SuggestionGroup
	quitting           bool
	editing            bool
	showHelp           bool
	width, height      int
	statusMessage      string
	statusMessageTimer int
}

func NewSuggestModel(suggestions []helpers.SuggestionGroup) SuggestModel {
	items := make([]list.Item, len(suggestions))
	for i, s := range suggestions {
		items[i] = SuggestionItem{s, i}
	}
	delegate := list.NewDefaultDelegate()
	delegate.ShowDescription = true
	delegate.Styles.SelectedTitle = styleSelectedItem
	delegate.Styles.SelectedDesc = styleSelectedItem.Copy().Foreground(lipgloss.Color("#FFFFFF"))
	delegate.Styles.NormalTitle = styleListItem
	delegate.Styles.NormalDesc = styleListItem.Copy().Foreground(dimmedColor)

	l := list.New(items, delegate, 0, 0)
	l.Title = "✨ Commit Grouping Suggestions"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Styles.Title = styleHeading.Copy().Padding(0, 1)
	l.Styles.PaginationStyle = styleHelp
	l.Styles.HelpStyle = styleHelp
	l.SetShowHelp(false)

	ta := textarea.New()
	ta.Placeholder = "Edit commit message..."
	ta.SetWidth(50)
	ta.SetHeight(5)
	ta.ShowLineNumbers = true
	ta.Prompt = "┃ "
	ta.FocusedStyle.CursorLine = lipgloss.NewStyle().Background(lipgloss.Color("#333333"))

	return SuggestModel{
		suggestions: suggestions,
		list:        l,
		input:       ta,
		keys:        suggestKeys,
	}
}

func (m SuggestModel) Init() tea.Cmd {
	m.statusMessage = "Use ↑/↓ to navigate, e to edit, s to stage, enter to select"
	m.statusMessageTimer = 5
	return tea.Batch(textarea.Blink, m.tickStatus())
}

func (m SuggestModel) tickStatus() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg { return statusTickMsg{} })
}

type statusTickMsg struct{}

func (m SuggestModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case statusTickMsg:
		if m.statusMessageTimer > 0 {
			m.statusMessageTimer--
			if m.statusMessageTimer == 0 {
				m.statusMessage = ""
			}
			return m, m.tickStatus()
		}
		return m, nil

	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.list.SetSize(msg.Width/2-2, msg.Height-4)
		m.input.SetWidth(msg.Width/2 - 4)
		return m, nil

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Quit):
			m.quitting = true
			m.selected = nil
			return m, tea.Quit
		case key.Matches(msg, m.keys.Help):
			m.showHelp = !m.showHelp
			return m, nil
		}

		if m.editing {
			switch msg.String() {
			case "esc":
				m.editing = false
				m.input.Blur()
				m.statusMessage = "Edit cancelled"
				m.statusMessageTimer = 3
				return m, m.tickStatus()
			case "enter":
				if m.selected != nil {
					m.selected.Message = m.input.Value()
					m.statusMessage = "Commit message updated"
					m.statusMessageTimer = 3
				}
				m.editing = false
				return m, m.tickStatus()
			default:
				var cmd tea.Cmd
				m.input, cmd = m.input.Update(msg)
				return m, cmd
			}
		}

		switch {
		case key.Matches(msg, m.keys.Enter):
			i, ok := m.list.SelectedItem().(SuggestionItem)
			if ok {
				m.selected = &m.suggestions[i.index]
				return m, tea.Quit
			}
		case key.Matches(msg, m.keys.Edit):
			i, ok := m.list.SelectedItem().(SuggestionItem)
			if ok {
				m.selected = &m.suggestions[i.index]
				m.input.SetValue(m.selected.Message)
				m.input.Focus()
				m.editing = true
				return m, textarea.Blink
			}
		case key.Matches(msg, m.keys.Stage):
			i, ok := m.list.SelectedItem().(SuggestionItem)
			if ok {
				m.suggestions[i.index].ShouldStage = true
				m.statusMessage = "Marked for staging"
				m.statusMessageTimer = 3
				return m, m.tickStatus()
			}
		case key.Matches(msg, m.keys.Unstage):
			i, ok := m.list.SelectedItem().(SuggestionItem)
			if ok {
				m.suggestions[i.index].ShouldStage = false
				m.statusMessage = "Unmarked for staging"
				m.statusMessageTimer = 3
				return m, m.tickStatus()
			}
		}
	}

	if !m.editing {
		var cmd tea.Cmd
		m.list, cmd = m.list.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m SuggestModel) View() string {
	if m.quitting {
		return styleError.Render("Operation cancelled")
	}

	if m.editing {
		var content []string
		content = append(content, styleHeading.Render("✎ Edit Commit Message"))
		content = append(content, styleCard.Copy().BorderForeground(primaryLight).Render(m.input.View()))
		content = append(content, styleHelp.Render("enter: save • esc: cancel"))
		if m.statusMessage != "" {
			content = append(content, m.renderStatusMessage())
		}
		return lipgloss.NewStyle().Padding(1, 2).Render(lipgloss.JoinVertical(lipgloss.Left, content...))
	}

	left := m.renderListPanel()
	right := m.renderDetailPanel()

	return lipgloss.JoinHorizontal(lipgloss.Top, left, right)
}

func (m SuggestModel) renderListPanel() string {
	return lipgloss.NewStyle().
		Width(m.width/2).
		Height(m.height).
		Padding(1, 1).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(primaryLight).
		Render(m.list.View())
}

func (m SuggestModel) renderDetailPanel() string {
	i, ok := m.list.SelectedItem().(SuggestionItem)
	if !ok {
		return ""
	}
	s := m.suggestions[i.index]

	var content []string
	content = append(content, styleHeading.Render("Details"))
	content = append(content, styleListTitle.Render("Description"))
	content = append(content, styleDescription.Render(s.Description))

	content = append(content, styleListTitle.Render(fmt.Sprintf("Files (%d)", len(s.Files))))
	if len(s.Files) == 0 {
		content = append(content, styleFileItem.Render("No files"))
	} else {
		for _, f := range s.Files {
			content = append(content, styleFileItem.Render("• "+f))
		}
	}

	content = append(content, styleListTitle.Render("Commit Message"))
	msgParts := strings.Split(s.Message, "\n\n")
	if len(msgParts) > 1 {
		content = append(content, styleCommitMsg.Render(msgParts[0]))
		bodyLines := strings.Split(msgParts[1], "\n")
		for _, line := range bodyLines {
			content = append(content, styleListItem.Render(line))
		}
		if len(msgParts) > 2 {
			content = append(content, "")
			content = append(content, styleListItem.Copy().Foreground(warningColor).Bold(true).Render(msgParts[2]))
		}
	} else {
		content = append(content, styleCommitMsg.Render(s.Message))
	}

	status := "Not marked for staging"
	if s.ShouldStage {
		status = "✓ Marked for staging"
	}
	content = append(content, styleListTitle.Render("Staging"))
	content = append(content, styleListItem.Render(status))

	if m.showHelp {
		content = append(content, styleHelp.Render("↑/↓: navigate • enter: select • e: edit • s/u: stage/unstage • q: quit • ?: toggle help"))
	} else {
		content = append(content, styleHelp.Render("?: help • q: quit"))
	}

	if m.statusMessage != "" {
		content = append(content, m.renderStatusMessage())
	}

	return lipgloss.NewStyle().
		Width(m.width/2).
		Height(m.height).
		Padding(1, 1).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(primaryLight).
		Render(lipgloss.JoinVertical(lipgloss.Left, content...))
}

func (m SuggestModel) renderStatusMessage() string {
	return lipgloss.NewStyle().
		Foreground(successColor).
		Bold(true).
		Padding(0, 2).
		MarginTop(1).
		Render(m.statusMessage)
}

func (m SuggestModel) Quitting() bool {
	return m.quitting
}

func (m SuggestModel) HasSelection() bool {
	return m.selected != nil
}

func (m SuggestModel) Selected() *helpers.SuggestionGroup {
	return m.selected
}
