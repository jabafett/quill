package ui

import "github.com/charmbracelet/lipgloss"

var (
	// Colors - improved for better visibility on both light and dark backgrounds
	primaryColor   = lipgloss.Color("#8B5CF6") // Vibrant purple
	primaryLight   = lipgloss.Color("#A78BFA") // Lighter purple for dark backgrounds
	primaryDark    = lipgloss.Color("#6D28D9") // Darker purple for light backgrounds
	successColor   = lipgloss.Color("#10B981") // Emerald green
	warningColor   = lipgloss.Color("#F59E0B") // Amber - more visible than soft yellow
	errorColor     = lipgloss.Color("#EF4444") // Red
	textColor      = lipgloss.Color("#E2E8F0") // Light gray for dark backgrounds
	textColorDark  = lipgloss.Color("#1F2937") // Dark gray for light backgrounds
	dimmedColor    = lipgloss.Color("#94A3B8") // Slate gray - more visible on both backgrounds
	highlightColor = lipgloss.Color("#FFFFFF") // White for highlights
	borderColor    = lipgloss.Color("#6D28D9") // Darker purple for borders
	bgColor        = lipgloss.Color("#1E293B") // Dark slate blue for backgrounds
	bgHighlight    = lipgloss.Color("#334155") // Lighter slate for highlighted backgrounds

	// Spinner style
	styleSpinner = lipgloss.NewStyle().
			Foreground(primaryLight).
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
			Foreground(primaryLight).
			MarginBottom(1).
			Padding(1, 2).
			BorderStyle(lipgloss.RoundedBorder()).
			BorderBottom(true).
			BorderTop(true).
			BorderRight(true).
			BorderLeft(true).
			BorderForeground(primaryLight)

	// List styles
	styleListTitle = lipgloss.NewStyle().
			MarginLeft(2).
			MarginBottom(1).
			Foreground(primaryLight).
			Bold(true).
			Padding(0, 1)

	styleList = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(primaryLight).
			Padding(1, 2)

	// Regular list item - works on both light and dark backgrounds
	styleListItem = lipgloss.NewStyle().
			PaddingLeft(4).
			Foreground(textColor).
			MarginLeft(2).
			MarginTop(1)

	// Selected item with high contrast
	styleSelectedItem = lipgloss.NewStyle().
				Foreground(highlightColor).
				Background(primaryDark).
				Bold(true).
				Padding(1, 2).
				MarginLeft(2).
				MarginTop(1)

	// File list item style
	styleFileItem = lipgloss.NewStyle().
			PaddingLeft(6).
			Foreground(dimmedColor).
			MarginLeft(2)

	// Section style for grouping related content
	styleSection = lipgloss.NewStyle().
			BorderLeft(true).
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(primaryLight).
			PaddingLeft(2).
			MarginLeft(2).
			MarginBottom(1)

	// Help and input styles
	styleHelp = lipgloss.NewStyle().
			Foreground(dimmedColor).
			Padding(1, 2).
			MarginTop(1).
			Align(lipgloss.Center)

	styleInput = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(primaryLight).
			Padding(1, 2).
			MarginLeft(2).
			MarginRight(2)

	// Card style for detail views
	styleCard = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(primaryLight).
			Padding(1, 2).
			MarginBottom(1)

	// Commit message style
	styleCommitMsg = lipgloss.NewStyle().
			Foreground(successColor).
			Bold(true).
			PaddingLeft(4).
			MarginLeft(2)

	// Description style
	styleDescription = lipgloss.NewStyle().
				Foreground(primaryLight).
				Bold(true).
				PaddingLeft(4).
				MarginLeft(2).
				MarginBottom(1)
)
