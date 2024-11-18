package ui

import "github.com/charmbracelet/lipgloss"

var (
	// Colors
	primaryColor   = lipgloss.Color("#8B5CF6") // Vibrant purple
	successColor   = lipgloss.Color("#10B981") // Emerald green
	errorColor     = lipgloss.Color("#EF4444") // Red
	textColor      = lipgloss.Color("#1F2937") // Dark gray
	dimmedColor    = lipgloss.Color("#6B7280") // Medium gray

	// Spinner style
	styleSpinner = lipgloss.NewStyle().
		Foreground(primaryColor).
		Bold(true)

	// Error and success styles
	styleError = lipgloss.NewStyle().
		Foreground(errorColor).
		Bold(true).
		Padding(0, 1)

	styleSuccess = lipgloss.NewStyle().
		Foreground(successColor).
		Bold(true).
		Padding(0, 1)

	// Heading style
	styleHeading = lipgloss.NewStyle().
		Bold(true).
		Foreground(primaryColor).
		MarginBottom(1).
		Padding(0, 2).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderBottom(true).
		BorderForeground(primaryColor)

	// List styles
	styleListTitle = lipgloss.NewStyle().
		MarginLeft(2).
		MarginBottom(1).
		Foreground(primaryColor).
		Bold(true).
		Padding(0, 1)

	styleListItem = lipgloss.NewStyle().
		PaddingLeft(4).
		Foreground(textColor).
		MarginLeft(2)

	styleSelectedItem = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(primaryColor).
		Bold(true).
		Padding(0, 2).
		MarginLeft(2)

	// Help and input styles
	styleHelp = lipgloss.NewStyle().
		Foreground(dimmedColor).
		Padding(1, 2).
		MarginTop(1).
		Align(lipgloss.Center)

	styleInput = lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(primaryColor).
		Padding(0, 1).
		MarginLeft(2).
		MarginRight(2)
)
