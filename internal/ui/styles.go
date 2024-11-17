package ui

import "github.com/charmbracelet/lipgloss"

var (
	// Colors
	primaryColor   = lipgloss.Color("#7B2CBF")
	successColor   = lipgloss.Color("#2ECC71")
	errorColor     = lipgloss.Color("#E63946")
	textColor      = lipgloss.Color("#1A1A1A")
	dimmedColor    = lipgloss.Color("#626262")
	spinnerColor   = lipgloss.Color("#9D4EDD")
	
	// Shared styles
	styleHeading = lipgloss.NewStyle().
		Bold(true).
		Foreground(primaryColor).
		MarginBottom(1).
		MarginTop(1).
		MarginLeft(2).
		BorderStyle(lipgloss.NormalBorder()).
		BorderBottom(true)

	styleError = lipgloss.NewStyle().
		Foreground(errorColor).
		Bold(true)

	styleSuccess = lipgloss.NewStyle().
		Foreground(successColor).
		Bold(true)

	styleSpinner = lipgloss.NewStyle().
		Foreground(spinnerColor)

	styleHelp = lipgloss.NewStyle().
		Foreground(dimmedColor).
		MarginTop(1).
		MarginLeft(2)

	// List styles
	styleListTitle = lipgloss.NewStyle().
		MarginLeft(2).
		Foreground(primaryColor).
		Bold(true)

	styleListItem = lipgloss.NewStyle().
		PaddingLeft(4).
		Foreground(textColor)

	styleSelectedItem = lipgloss.NewStyle().
		PaddingLeft(2).
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(primaryColor).
		Bold(true)
)