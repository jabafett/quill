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
}

var suggestKeys = suggestKeyMap{
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
		key.WithHelp("e", "edit msg"),
	),
	Reload: key.NewBinding(
		key.WithKeys("r"),
		key.WithHelp("r", "regen"),
	),
	Stage: key.NewBinding(
		key.WithKeys("s"),
		key.WithHelp("s", "stage"),
	),
	Unstage: key.NewBinding(
		key.WithKeys("u"),
		key.WithHelp("u", "unstage"),
	),
	Back: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "cancel edit"),
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

func (i SuggestionItem) Title() string {
	marker := "✗"
	if i.suggestion.ShouldStage {
		marker = "✔"
	}
	return fmt.Sprintf("%s %s", marker, i.suggestion.Description)
}
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
	// Handled manually in renderListPanel so it stays pinned
	l.Title = ""
	l.SetShowStatusBar(true)
	l.SetFilteringEnabled(false)
	l.Styles.Title = lipgloss.NewStyle() // Title is rendered manually
	l.Styles.PaginationStyle = styleHelp
	l.Styles.HelpStyle = styleHelp
	l.SetShowHelp(true)
	l.Styles.NoItems = lipgloss.NewStyle().Padding(1, 2) // Style for empty list
	l.Styles.FilterPrompt = lipgloss.NewStyle().Foreground(primaryLight)
	l.Styles.FilterCursor = lipgloss.NewStyle().Foreground(primaryLight)
	l.Styles.StatusBar = styleHelp // Use help style for status bar

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
	m.statusMessage = ""
	m.statusMessageTimer = 0
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
		m.width = msg.Width
		m.height = msg.Height

		// No fixed margin subtraction; use full window size
		totalWidth := m.width
		totalHeight := m.height * 80 / 100

		// Calculate frame sizes dynamically
		listFrameX, listFrameY := lipgloss.NewStyle().Padding(1, 1).Border(lipgloss.RoundedBorder()).GetFrameSize()
		detailFrameX, _ := lipgloss.NewStyle().Padding(1, 1).Border(lipgloss.RoundedBorder()).GetFrameSize()

		// Allocate panel widths so that total including frames fits terminal width
		listPanelContentWidth := (totalWidth / 3) - listFrameX
		if listPanelContentWidth < 10 {
			listPanelContentWidth = 10 // minimum width
		}
		detailPanelContentWidth := (totalWidth - (totalWidth / 3)) - detailFrameX
		if detailPanelContentWidth < 10 {
			detailPanelContentWidth = 10
		}

		// Allocate panel heights so that total including frames fits terminal height
		panelContentHeight := totalHeight - listFrameY
		if panelContentHeight < 5 {
			panelContentHeight = 5
		}

		// Set inner content sizes
		m.list.SetSize(listPanelContentWidth, panelContentHeight)
		m.input.SetWidth(detailPanelContentWidth)
		// keep textarea height fixed (default 5)

		return m, nil

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Quit):
			m.quitting = true
			m.selected = nil
			return m, tea.Quit
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

// GetStagedSuggestions returns all suggestions that are marked for staging
func (m SuggestModel) GetStagedSuggestions() []*helpers.SuggestionGroup {
	var stagedGroups []*helpers.SuggestionGroup
	for i := range m.suggestions {
		if m.suggestions[i].ShouldStage {
			stagedGroups = append(stagedGroups, &m.suggestions[i])
		}
	}
	return stagedGroups
}

func (m SuggestModel) View() string {
	if m.quitting {
		return styleError.Render("Operation cancelled")
	}

	if m.editing {
		var content []string
		content = append(content, styleHeading.Render("✎ Edit Commit Message"))
		content = append(content, styleCard.Copy().BorderForeground(primaryLight).Render(m.input.View()))
		content = append(content, m.renderKeybinds())
		if m.statusMessage != "" {
			content = append(content, m.renderStatusMessage())
		}
		return lipgloss.NewStyle().Padding(1, 2).Render(lipgloss.JoinVertical(lipgloss.Left, content...))
	}

	listPanel := m.renderListPanel()
	detailPanel := m.renderDetailPanel()

	return lipgloss.JoinHorizontal(lipgloss.Top, listPanel, detailPanel)
}

func (m SuggestModel) renderListPanel() string {
	listWidth := m.width / 3

	// Pinned header
	header := styleHeading.Render("✨ Commit Groups")

	// Scrolling list body
	body := lipgloss.NewStyle().
		Width(listWidth).
		Render(m.list.View())

	return lipgloss.JoinVertical(lipgloss.Left, header, body)
}

func (m SuggestModel) renderDetailPanel() string {
	detailWidth := m.width - (m.width / 3)
	var content []string

	content = append(content, styleHeading.Render("Details"))

	i, ok := m.list.SelectedItem().(SuggestionItem)
	if !ok {
		content = append(content, styleListItem.Render("No suggestion selected"))
	} else {
		s := m.suggestions[i.index]

		content = append(content, styleListTitle.Render("Description"))
		content = append(content, styleDescription.Render(s.Description))

		content = append(content, styleListTitle.Render(fmt.Sprintf("Files (%d)", len(s.Files))))
		if len(s.Files) == 0 {
			content = append(content, styleFileItem.Render("No files"))
		} else {
			for _, f := range s.Files {
				var prefix string
				if s.ShouldStage {
					prefix = "✔"
				} else {
					prefix = "✗"
				}
				content = append(content, styleFileItem.Render(prefix+" "+f))
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
	}

	content = append(content, m.renderKeybinds())

	if m.statusMessage != "" {
		content = append(content, m.renderStatusMessage())
	}

	return lipgloss.NewStyle().
		Width(detailWidth).
		Height(m.height).
		Padding(1, 1).
		MarginLeft(1).
		Render(lipgloss.JoinVertical(lipgloss.Left, content...))
}

func (m SuggestModel) renderKeybinds() string {
	return styleHelp.Render("↑/↓: navigate • enter: select • e: edit • s/u: stage/unstage • q: quit")
}

func (m SuggestModel) renderStatusMessage() string {
	separator := lipgloss.NewStyle().
		Foreground(dimmedColor).
		Render(strings.Repeat("─", m.width/2))

	return lipgloss.JoinVertical(lipgloss.Left,
		separator,
		lipgloss.NewStyle().
			Foreground(successColor).
			Bold(true).
			Padding(0, 2).
			MarginTop(1).
			Render(m.statusMessage),
	)
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
