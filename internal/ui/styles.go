package ui

import "github.com/charmbracelet/lipgloss"

var (
	// Colors - using a more cohesive color palette
	primaryColor   = lipgloss.Color("#8B5CF6") // Vibrant purple
	// secondaryColor = lipgloss.Color("#6D28D9") // Darker purple
	successColor   = lipgloss.Color("#10B981") // Emerald green
	errorColor     = lipgloss.Color("#EF4444") // Red
	textColor      = lipgloss.Color("#1F2937") // Dark gray
	dimmedColor    = lipgloss.Color("#6B7280") // Medium gray
	spinnerColor   = lipgloss.Color("#8B5CF6") // Match primary
	bgHighlight    = lipgloss.Color("#F3F4F6") // Light gray background

	// Shared styles
	styleHeading = lipgloss.NewStyle().
		Bold(true).
		Foreground(primaryColor).
		MarginBottom(1).
		MarginTop(1).
		MarginLeft(2).
		Padding(0, 1).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderBottom(true).
		BorderForeground(primaryColor)

	styleError = lipgloss.NewStyle().
		Foreground(errorColor).
		Bold(true).
		Padding(0, 1)

	styleSuccess = lipgloss.NewStyle().
		Foreground(successColor).
		Bold(true).
		Padding(0, 1)

	styleSpinner = lipgloss.NewStyle().
		Foreground(spinnerColor).
		Bold(true)

	styleHelp = lipgloss.NewStyle().
		Foreground(dimmedColor).
		MarginTop(1).
		MarginLeft(2).
		Padding(1, 2).
		Background(bgHighlight).
		BorderStyle(lipgloss.RoundedBorder())

	// List styles
	styleListTitle = lipgloss.NewStyle().
		MarginLeft(2).
		MarginBottom(1).
		Foreground(primaryColor).
		Bold(true).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderBottom(true).
		BorderForeground(primaryColor).
		Padding(0, 1)

	styleListItem = lipgloss.NewStyle().
		PaddingLeft(4).
		Foreground(textColor).
		Padding(0, 1)

	styleSelectedItem = lipgloss.NewStyle().
		PaddingLeft(2).
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(primaryColor).
		Bold(true).
		Padding(0, 1)
)