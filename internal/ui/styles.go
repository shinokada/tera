package ui

import (
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/lipgloss"
)

// Color palette matching bash version
var (
	colorCyan   = lipgloss.Color("6")  // Cyan for titles
	colorBlue   = lipgloss.Color("12") // Bright blue for TERA header
	colorYellow = lipgloss.Color("3")  // Yellow for highlights
	colorRed    = lipgloss.Color("9")  // Red for errors
	colorGreen  = lipgloss.Color("2")  // Green for success
	colorGray   = lipgloss.Color("8")  // Gray for secondary text
)

// Common styles
var (
	titleStyle = lipgloss.NewStyle().
			Foreground(colorCyan).
			Bold(true)

	// Title style without padding for list titles
	listTitleStyle = lipgloss.NewStyle().
			Foreground(colorCyan).
			Bold(true).
			MarginLeft(-2) // Compensate for the page left padding

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

	quickFavoritesStyle = titleStyle.Foreground(lipgloss.Color("99"))

	docStyle = helpStyle.Padding(1, 2)

	// Style for content without top padding (used after header)
	docStyleNoTopPadding = helpStyle.PaddingTop(0).PaddingBottom(1).PaddingLeft(2).PaddingRight(2)
)

// createStyledDelegate creates a list delegate with single-line items and consistent styling
func createStyledDelegate() list.DefaultDelegate {
	delegate := list.NewDefaultDelegate()
	delegate.SetHeight(1)            // Single line per item
	delegate.SetSpacing(0)           // Remove spacing between items
	delegate.ShowDescription = false // Hide description text below items
	// Remove vertical padding from delegate styles
	delegate.Styles.NormalTitle = lipgloss.NewStyle()
	delegate.Styles.SelectedTitle = lipgloss.NewStyle().Foreground(colorYellow).Bold(true)
	return delegate
}

// wrapPageWithHeader wraps content with TERA header at the top and applies consistent padding
func wrapPageWithHeader(content string) string {
	var b strings.Builder
	// Center TERA header with proper width - add top padding here
	header := lipgloss.NewStyle().
		Width(50).
		Align(lipgloss.Center).
		Foreground(colorBlue). // Use blue color for TERA
		Bold(true).
		PaddingTop(1).
		Render("TERA")
	b.WriteString(header)
	b.WriteString("\n")
	b.WriteString(content)
	// Use style without top padding since header already has it
	return docStyleNoTopPadding.Render(b.String())
}

// PageLayout represents a consistent page layout structure
type PageLayout struct {
	Title    string // Main title (optional)
	Subtitle string // Subtitle (optional)
	Content  string // Main content area
	Help     string // Help text at bottom
}

// RenderPage renders a page with consistent layout using the template
// This ensures all pages have the same spacing and structure
func RenderPage(layout PageLayout) string {
	var b strings.Builder

	// Add consistent spacing after TERA header
	b.WriteString("\n")

	// Title section - always takes up one line (empty or not) for consistency
	if layout.Title != "" {
		b.WriteString(titleStyle.Render(layout.Title))
	}
	b.WriteString("\n")

	// Subtitle section - always takes up one line (empty or not) for consistency  
	if layout.Subtitle != "" {
		b.WriteString(subtitleStyle.Render(layout.Subtitle))
	}
	b.WriteString("\n")

	// Main content
	if layout.Content != "" {
		b.WriteString(layout.Content)
	}

	// Help text (if provided)
	if layout.Help != "" {
		if layout.Content != "" {
			b.WriteString("\n")
		}
		b.WriteString(helpStyle.Render(layout.Help))
	}

	return wrapPageWithHeader(b.String())
}

// RenderPageWithBottomHelp renders a page with help text at the bottom of the screen
func RenderPageWithBottomHelp(layout PageLayout, terminalHeight int) string {
	var b strings.Builder

	// Add consistent spacing after TERA header
	b.WriteString("\n")

	// Title section - always takes up one line (empty or not) for consistency
	if layout.Title != "" {
		b.WriteString(titleStyle.Render(layout.Title))
	}
	b.WriteString("\n")

	// Subtitle section - always takes up one line (empty or not) for consistency  
	if layout.Subtitle != "" {
		b.WriteString(subtitleStyle.Render(layout.Subtitle))
	}
	b.WriteString("\n")

	// Main content
	if layout.Content != "" {
		b.WriteString(layout.Content)
	}

	// Calculate how many lines we've used so far
	// TERA header (3 lines) + blank line (1) + title (1) + subtitle (1) + content lines + padding (2)
	contentLines := strings.Count(b.String(), "\n")
	teraHeaderLines := 3
	totalUsed := teraHeaderLines + contentLines + 2 // +2 for padding

	// Calculate remaining space for help text to be at bottom
	// Reserve 1 line for help text itself
	remainingLines := terminalHeight - totalUsed - 1
	if remainingLines < 0 {
		remainingLines = 0
	}

	// Add spacing to push help text to bottom
	for i := 0; i < remainingLines; i++ {
		b.WriteString("\n")
	}

	// Help text (if provided)
	if layout.Help != "" {
		b.WriteString(helpStyle.Render(layout.Help))
	}

	return wrapPageWithHeader(b.String())
}
