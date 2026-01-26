package ui

import (
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/lipgloss"
)

// Color palette matching bash version
var (
	colorCyan   = lipgloss.Color("6") // Cyan for titles
	colorYellow = lipgloss.Color("3") // Yellow for highlights
	colorRed    = lipgloss.Color("9") // Red for errors
	colorGreen  = lipgloss.Color("2") // Green for success
	colorGray   = lipgloss.Color("8") // Gray for secondary text
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

	teraHeaderStyle = lipgloss.NewStyle().
			Foreground(colorCyan).
			Bold(true).
			Align(lipgloss.Center)

	quickFavoritesStyle = titleStyle.Copy().
				Foreground(lipgloss.Color("99"))

	docStyle = helpStyle.Copy().Padding(1, 2)
)

// createStyledDelegate creates a list delegate with single-line items and consistent styling
func createStyledDelegate() list.DefaultDelegate {
	delegate := list.NewDefaultDelegate()
	delegate.SetHeight(1)            // Single line per item
	delegate.SetSpacing(0)           // Remove spacing between items
	delegate.ShowDescription = false // Hide cursor indicator
	// Remove vertical padding from delegate styles
	delegate.Styles.NormalTitle = lipgloss.NewStyle()
	delegate.Styles.SelectedTitle = lipgloss.NewStyle().Foreground(colorYellow).Bold(true)
	return delegate
}

// wrapPageWithHeader wraps content with TERA header at the top and applies consistent padding
func wrapPageWithHeader(content string) string {
	var b strings.Builder
	// Center TERA header with proper width
	header := lipgloss.NewStyle().
		Width(50).
		Align(lipgloss.Center).
		Foreground(colorCyan).
		Bold(true).
		Render("TERA")
	b.WriteString(header)
	b.WriteString("\n\n")
	b.WriteString(content)
	return docStyle.Render(b.String())
}
