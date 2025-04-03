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
                key.WithHelp("e", "edit message"),
        ),
        Reload: key.NewBinding(
                key.WithKeys("r"),
                key.WithHelp("r", "reload"),
        ),
        Stage: key.NewBinding(
                key.WithKeys("s"),
                key.WithHelp("s", "stage files"),
        ),
        Unstage: key.NewBinding(
                key.WithKeys("u"),
                key.WithHelp("u", "unstage files"),
        ),
        ViewMore: key.NewBinding(
                key.WithKeys("v"),
                key.WithHelp("v", "view details"),
        ),
        Back: key.NewBinding(
                key.WithKeys("b", "esc"),
                key.WithHelp("b/esc", "back"),
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
        suggestions []helpers.SuggestionGroup
        list        list.Model
        input       textarea.Model
        keys        suggestKeyMap
        selected    *helpers.SuggestionGroup
        quitting    bool
        editing     bool
        viewingDetails bool
        currentDetailIndex int
        width      int
        height     int
}

// NewSuggestModel creates a new model for the suggest UI
func NewSuggestModel(suggestions []helpers.SuggestionGroup) SuggestModel {
        // Create list items
        items := make([]list.Item, len(suggestions))
        for i, suggestion := range suggestions {
                items[i] = SuggestionItem{
                        suggestion: suggestion,
                        index:      i,
                }
        }

        // Create list
        l := list.New(items, list.NewDefaultDelegate(), 0, 0)
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
        }
}

func (m SuggestModel) Init() tea.Cmd {
        return tea.Batch(
                textarea.Blink,
                list.NewStatusMessage("Use arrow keys to navigate, enter to select"),
        )
}

func (m SuggestModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
        var cmds []tea.Cmd

        switch msg := msg.(type) {
        case tea.WindowSizeMsg:
                m.width = msg.Width
                m.height = msg.Height
                
                // Update list dimensions
                h, v := listStyle.GetFrameSize()
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
                }
                
                // Handle mode-specific keys
                if m.editing {
                        switch msg.String() {
                        case "esc":
                                m.editing = false
                                m.input.Blur()
                                return m, nil
                        case "enter":
                                // Update the selected suggestion's message
                                if m.selected != nil {
                                        m.selected.Message = m.input.Value()
                                }
                                m.editing = false
                                return m, nil
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
                        case key.Matches(msg, m.keys.Up), key.Matches(msg, m.keys.Down):
                                // Navigate between details
                                if key.Matches(msg, m.keys.Up) && m.currentDetailIndex > 0 {
                                        m.currentDetailIndex--
                                } else if key.Matches(msg, m.keys.Down) && m.currentDetailIndex < len(m.suggestions)-1 {
                                        m.currentDetailIndex++
                                }
                                return m, nil
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
                                        return m, list.NewStatusMessage("Files marked for staging")
                                }
                        case key.Matches(msg, m.keys.Unstage):
                                i, ok := m.list.SelectedItem().(SuggestionItem)
                                if ok {
                                        m.suggestions[i.index].ShouldStage = false
                                        return m, list.NewStatusMessage("Files unmarked for staging")
                                }
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
                        ),
                )
        }

        // Default list view
        help := lipgloss.JoinHorizontal(lipgloss.Center,
                "↑/↓: navigate",
                " • ",
                "enter: select",
                " • ",
                "e: edit message",
                " • ",
                "v: view details",
                " • ",
                "s: stage",
                " • ",
                "q: quit",
        )

        return lipgloss.JoinVertical(lipgloss.Left,
                m.list.View(),
                styleHelp.Render(help),
        )
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
        
        // Add impact if available
        if suggestion.Impact != "" {
                details = append(details, styleListTitle.Render("Version Impact:"))
                details = append(details, styleListItem.Render(suggestion.Impact))
        }
        
        // Add navigation help
        help := lipgloss.JoinHorizontal(lipgloss.Center,
                "↑/↓: navigate suggestions",
                " • ",
                "esc/b: back to list",
                " • ",
                "q: quit",
        )
        
        details = append(details, styleHelp.Render(help))
        
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

// Define listStyle
var listStyle = lipgloss.NewStyle().
        Border(lipgloss.RoundedBorder()).
        BorderForeground(primaryColor).
        Padding(1, 2)