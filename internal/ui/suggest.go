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

// min returns the smaller of x or y.
func min(x, y int) int {
        if x < y {
                return x
        }
        return y
}

// SuggestKeyMap defines keybindings for the suggest UI
type suggestKeyMap struct {
        Up       key.Binding
        Down     key.Binding
        Enter    key.Binding
        Quit     key.Binding
        Edit     key.Binding
        Reload   key.Binding
        Stage    key.Binding
        Unstage  key.Binding
        ViewMore key.Binding
        Back     key.Binding
        Next     key.Binding
        Prev     key.Binding
        Help     key.Binding
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
        ViewMore: key.NewBinding(
                key.WithKeys("v"),
                key.WithHelp("v", "view details"),
        ),
        Back: key.NewBinding(
                key.WithKeys("esc", "backspace"),
                key.WithHelp("esc", "go back"),
        ),
        Next: key.NewBinding(
                key.WithKeys("tab", "right"),
                key.WithHelp("tab/→", "next suggestion"),
        ),
        Prev: key.NewBinding(
                key.WithKeys("shift+tab", "left"),
                key.WithHelp("shift+tab/←", "previous suggestion"),
        ),
        Help: key.NewBinding(
                key.WithKeys("?", "h"),
                key.WithHelp("?/h", "toggle help"),
        ),
        Quit: key.NewBinding(
                key.WithKeys("q", "ctrl+c"),
                key.WithHelp("q", "quit"),
        ),
}

// SuggestionItem represents a single suggestion in the list
type SuggestionItem struct {
        suggestion helpers.SuggestionGroup
        index      int
}

func (i SuggestionItem) Title() string {
        return i.suggestion.Description
}

func (i SuggestionItem) Description() string {
        fileCount := len(i.suggestion.Files)
        if fileCount == 0 {
                return "No files"
        }

        if fileCount == 1 {
                return fmt.Sprintf("1 file: %s", i.suggestion.Files[0])
        }

        return fmt.Sprintf("%d files: %s, ...", fileCount, i.suggestion.Files[0])
}

func (i SuggestionItem) FilterValue() string {
        return i.suggestion.Description
}

// SuggestModel is the UI model for the suggest command
type SuggestModel struct {
        suggestions        []helpers.SuggestionGroup
        list               list.Model
        input              textarea.Model
        keys               suggestKeyMap
        selected           *helpers.SuggestionGroup
        quitting           bool
        editing            bool
        viewingDetails     bool
        showHelp           bool
        currentDetailIndex int
        width              int
        height             int
        statusMessage      string
        statusMessageTimer int
        debug              bool
}

// NewSuggestModel creates a new model for the suggest UI
func NewSuggestModel(suggestions []helpers.SuggestionGroup, debug bool) SuggestModel {
        // Create list items
        items := make([]list.Item, len(suggestions))
        for i, suggestion := range suggestions {
                items[i] = SuggestionItem{
                        suggestion: suggestion,
                        index:      i,
                }
        }

        // Create list with a custom delegate that doesn't take up the full screen
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
        l.Styles.Title = styleHeading
        l.Styles.PaginationStyle = styleHelp
        l.Styles.HelpStyle = styleHelp

        // Create textarea for editing
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
                debug:       debug,
        }
}

func (m SuggestModel) Init() tea.Cmd {
        m.statusMessage = "Welcome to Quill Suggest! Use arrow keys to navigate, press ? for help"
        m.statusMessageTimer = 5

        return tea.Batch(
                textarea.Blink,
                m.tickStatus(),
        )
}

// tickStatus decrements the status message timer
func (m SuggestModel) tickStatus() tea.Cmd {
        return tea.Tick(time.Second, func(t time.Time) tea.Msg {
                return statusTickMsg{}
        })
}

// statusTickMsg is sent when the status message timer ticks
type statusTickMsg struct{}

func (m SuggestModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
        var cmds []tea.Cmd

        switch msg := msg.(type) {
        case statusTickMsg:
                // Update status message timer
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

                // Update list dimensions
                h, v := styleList.GetFrameSize()
                m.list.SetSize(msg.Width-h, msg.Height-v)

                // Update textarea dimensions
                m.input.SetWidth(msg.Width - 4)

                return m, nil

        case tea.KeyMsg:
                // Handle global keys first
                switch {
                case key.Matches(msg, m.keys.Quit):
                        m.quitting = true
                        return m, tea.Quit
                case key.Matches(msg, m.keys.Help):
                        m.showHelp = !m.showHelp
                        return m, nil
                }

                // Handle mode-specific keys
                if m.editing {
                        switch msg.String() {
                        case "esc":
                                m.editing = false
                                m.input.Blur()
                                m.statusMessage = "Edit cancelled"
                                m.statusMessageTimer = 3
                                return m, m.tickStatus()
                        case "enter":
                                // Update the selected suggestion's message
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
                } else if m.viewingDetails {
                        switch {
                        case key.Matches(msg, m.keys.Back):
                                m.viewingDetails = false
                                return m, nil
                        case key.Matches(msg, m.keys.Next), key.Matches(msg, m.keys.Down):
                                // Navigate to next suggestion
                                if m.currentDetailIndex < len(m.suggestions)-1 {
                                        m.currentDetailIndex++
                                        m.selected = &m.suggestions[m.currentDetailIndex]
                                }
                                return m, nil
                        case key.Matches(msg, m.keys.Prev), key.Matches(msg, m.keys.Up):
                                // Navigate to previous suggestion
                                if m.currentDetailIndex > 0 {
                                        m.currentDetailIndex--
                                        m.selected = &m.suggestions[m.currentDetailIndex]
                                }
                                return m, nil
                        case key.Matches(msg, m.keys.Edit):
                                // Edit the current suggestion's message
                                m.input.SetValue(m.suggestions[m.currentDetailIndex].Message)
                                m.input.Focus()
                                m.editing = true
                                m.viewingDetails = false
                                return m, textarea.Blink
                        case key.Matches(msg, m.keys.Stage):
                                // Mark for staging
                                m.suggestions[m.currentDetailIndex].ShouldStage = true
                                m.statusMessage = "Files marked for staging"
                                m.statusMessageTimer = 3
                                return m, m.tickStatus()
                        case key.Matches(msg, m.keys.Unstage):
                                // Unmark for staging
                                m.suggestions[m.currentDetailIndex].ShouldStage = false
                                m.statusMessage = "Files unmarked for staging"
                                m.statusMessageTimer = 3
                                return m, m.tickStatus()
                        case key.Matches(msg, m.keys.Enter):
                                // Select this suggestion
                                m.selected = &m.suggestions[m.currentDetailIndex]
                                return m, tea.Quit
                        }
                } else {
                        // List view
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
                        case key.Matches(msg, m.keys.ViewMore):
                                i, ok := m.list.SelectedItem().(SuggestionItem)
                                if ok {
                                        m.selected = &m.suggestions[i.index]
                                        m.currentDetailIndex = i.index
                                        m.viewingDetails = true
                                        return m, nil
                                }
                        case key.Matches(msg, m.keys.Stage):
                                i, ok := m.list.SelectedItem().(SuggestionItem)
                                if ok {
                                        m.suggestions[i.index].ShouldStage = true
                                        m.statusMessage = "Files marked for staging"
                                        m.statusMessageTimer = 3
                                        return m, m.tickStatus()
                                }
                        case key.Matches(msg, m.keys.Unstage):
                                i, ok := m.list.SelectedItem().(SuggestionItem)
                                if ok {
                                        m.suggestions[i.index].ShouldStage = false
                                        m.statusMessage = "Files unmarked for staging"
                                        m.statusMessageTimer = 3
                                        return m, m.tickStatus()
                                }
                        case key.Matches(msg, m.keys.Next):
                                // Move to next item in the list
                                if m.list.Index() < len(m.suggestions)-1 {
                                        m.list.Select(m.list.Index() + 1)
                                }
                                return m, nil
                        case key.Matches(msg, m.keys.Prev):
                                // Move to previous item in the list
                                if m.list.Index() > 0 {
                                        m.list.Select(m.list.Index() - 1)
                                }
                                return m, nil
                        }
                }
        }

        // Handle list updates
        if !m.editing && !m.viewingDetails {
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

        // Use a more compact style similar to the generate command
        mainStyle := lipgloss.NewStyle().Padding(1, 2)

        if m.editing {
                return mainStyle.Render(
                        lipgloss.JoinVertical(lipgloss.Left,
                                styleHeading.Render("✎ Edit Commit Message"),
                                styleInput.Render(m.input.View()),
                                styleHelp.Render("enter: save • esc: cancel"),
                                m.renderStatusMessage(),
                        ),
                )
        }

        if m.viewingDetails {
                return m.renderDetailView()
        }

        if m.selected != nil {
                return mainStyle.Render(
                        lipgloss.JoinVertical(lipgloss.Left,
                                styleHeading.Render("✓ Selected Grouping"),
                                styleSelectedItem.Render(m.selected.Description),
                                m.renderFileList(m.selected.Files),
                                styleListItem.Render(fmt.Sprintf("Commit message: %s", m.selected.Message)),
                                styleHelp.Render("Press enter to confirm"),
                                m.renderStatusMessage(),
                        ),
                )
        }

        // Default list view - more compact like generate command
        var content []string
        
        // Add title
        content = append(content, styleHeading.Render("✨ Commit Grouping Suggestions"))
        
        // Add list items in a more compact format
        for i, suggestion := range m.suggestions {
                var itemStyle lipgloss.Style
                if i == m.list.Index() {
                        itemStyle = styleSelectedItem
                } else {
                        itemStyle = styleListItem
                }
                
                // Format description and file count
                desc := suggestion.Description
                fileInfo := fmt.Sprintf("%d file(s)", len(suggestion.Files))
                if len(suggestion.Files) > 0 {
                        fileInfo = fmt.Sprintf("%d file(s): %s", len(suggestion.Files), suggestion.Files[0])
                        if len(suggestion.Files) > 1 {
                                fileInfo += ", ..."
                        }
                }
                
                content = append(content, itemStyle.Render(desc))
                content = append(content, styleListItem.Copy().Foreground(dimmedColor).Render(fileInfo))
                
                // Add a small separator between items
                if i < len(m.suggestions)-1 {
                        content = append(content, "")
                }
        }

        // Show help text
        var helpText string
        if m.showHelp {
                helpText = m.renderFullHelp()
        } else {
                helpText = m.renderCompactHelp()
        }
        content = append(content, styleHelp.Render(helpText))

        // Show status message if any
        if m.statusMessage != "" {
                content = append(content, m.renderStatusMessage())
        }

        return mainStyle.Render(lipgloss.JoinVertical(lipgloss.Left, content...))
}

// renderStatusMessage renders the status message with a timer
func (m SuggestModel) renderStatusMessage() string {
        if m.statusMessage == "" {
                return ""
        }

        statusStyle := lipgloss.NewStyle().
                Foreground(successColor).
                Bold(true).
                Padding(0, 2).
                MarginTop(1)

        return statusStyle.Render(m.statusMessage)
}

// renderCompactHelp renders a compact help message
func (m SuggestModel) renderCompactHelp() string {
        return lipgloss.JoinHorizontal(lipgloss.Center,
                "↑/↓: navigate",
                " • ",
                "enter: select",
                " • ",
                "e: edit",
                " • ",
                "v: details",
                " • ",
                "s: stage",
                " • ",
                "?: help",
                " • ",
                "q: quit",
        )
}

// renderFullHelp renders a detailed help message
func (m SuggestModel) renderFullHelp() string {
        helpItems := []string{
                "↑/↓/j/k: Navigate up/down",
                "tab/→: Next suggestion",
                "shift+tab/←: Previous suggestion",
                "enter: Select and apply suggestion",
                "e: Edit commit message",
                "v: View detailed information",
                "s: Mark files for staging",
                "u: Unmark files for staging",
                "esc: Go back to previous view",
                "?: Toggle help",
                "q: Quit without applying",
        }

        return strings.Join(helpItems, " • ")
}

// renderDetailView renders a detailed view of the current suggestion
func (m SuggestModel) renderDetailView() string {
        suggestion := m.suggestions[m.currentDetailIndex]

        mainStyle := lipgloss.NewStyle().Padding(1, 2)

        // Build the detail view
        var details []string

        // Add description
        details = append(details, styleHeading.Render(fmt.Sprintf("Suggestion %d of %d", m.currentDetailIndex+1, len(m.suggestions))))
        details = append(details, styleListTitle.Render("Description:"))
        details = append(details, styleListItem.Render(suggestion.Description))

        // Add files
        details = append(details, styleListTitle.Render("Files:"))
        details = append(details, m.renderFileList(suggestion.Files))

        // Add commit message
        details = append(details, styleListTitle.Render("Commit Message:"))
        details = append(details, styleListItem.Render(suggestion.Message))

        // Add staging status
        stagingStatus := "Not marked for staging"
        if suggestion.ShouldStage {
                stagingStatus = "Marked for staging"
        }
        details = append(details, styleListTitle.Render("Staging Status:"))
        details = append(details, styleListItem.Render(stagingStatus))

        // Add navigation help
        var helpText string
        if m.showHelp {
                helpText = "↑/↓: Navigate suggestions • e: Edit message • s: Mark for staging • u: Unmark • enter: Select • esc: Back • ?: Hide help • q: Quit"
        } else {
                helpText = "↑/↓: Navigate • e: Edit • s: Stage • u: Unstage • enter: Select • esc: Back • ?: Help • q: Quit"
        }
        details = append(details, styleHelp.Render(helpText))

        // Add status message if any
        if m.statusMessage != "" {
                details = append(details, m.renderStatusMessage())
        }

        return mainStyle.Render(lipgloss.JoinVertical(lipgloss.Left, details...))
}

// renderFileList renders a list of files
func (m SuggestModel) renderFileList(files []string) string {
        if len(files) == 0 {
                return styleListItem.Render("No files")
        }

        var fileItems []string
        for _, file := range files {
                fileItems = append(fileItems, styleListItem.Render("• "+file))
        }

        return strings.Join(fileItems, "\n")
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
