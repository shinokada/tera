package ui

import "github.com/charmbracelet/lipgloss"

// Color palette matching bash version
var (
	colorCyan   = lipgloss.Color("6")  // Cyan for titles
	colorYellow = lipgloss.Color("3")  // Yellow for highlights
	colorRed    = lipgloss.Color("9")  // Red for errors
	colorGreen  = lipgloss.Color("2")  // Green for success
	colorGray   = lipgloss.Color("8")  // Gray for secondary text
)

// Common styles
var (
	titleStyle = lipgloss.NewStyle().
			Foreground(colorCyan).
			Bold(true).
			Padding(0, 1)

	subtitleStyle = lipgloss.NewStyle().
			Foreground(colorYellow).
			Bold(true)

	highlightStyle = lipgloss.NewStyle().
			Foreground(colorYellow).
			Bold(true)

	errorStyle = lipgloss.NewStyle().
			Foreground(colorRed).
			Bold(true)

	successStyle = lipgloss.NewStyle().
			Foreground(colorGreen).
			Bold(true)

	helpStyle = lipgloss.NewStyle().
			Foreground(colorGray)

	infoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("7"))

	// Text styles
	boldStyle = lipgloss.NewStyle().
			Bold(true)

	subtleStyle = lipgloss.NewStyle().
			Foreground(colorGray)

	// List styles
	selectedItemStyle = lipgloss.NewStyle().
				Foreground(colorYellow).
				Bold(true).
				PaddingLeft(2)

	normalItemStyle = lipgloss.NewStyle().
			PaddingLeft(2)

	paginationStyle = lipgloss.NewStyle().
			Foreground(colorCyan)

	// Box styles
	boxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorCyan).
			Padding(1, 2)

	// Station info styles
	stationNameStyle = lipgloss.NewStyle().
				Foreground(colorCyan).
				Bold(true)

	stationFieldStyle = lipgloss.NewStyle().
				Foreground(colorGray)

	stationValueStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("7"))
)
