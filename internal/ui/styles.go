package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/lipgloss"
	"github.com/shinokada/tera/internal/api"
	"github.com/shinokada/tera/internal/theme"
)

// Global header renderer instance (initialized in app.go)
var globalHeaderRenderer *HeaderRenderer

// InitializeHeaderRenderer initializes the global header renderer
func InitializeHeaderRenderer() {
	globalHeaderRenderer = NewHeaderRenderer()
}

// Color accessors - these call theme.Current() to get current theme values
func colorCyan() lipgloss.Color   { t := theme.Current(); return t.PrimaryColor() }
func colorBlue() lipgloss.Color   { t := theme.Current(); return t.SecondaryColor() }
func colorYellow() lipgloss.Color { t := theme.Current(); return t.HighlightColor() }
func colorRed() lipgloss.Color    { t := theme.Current(); return t.ErrorColor() }
func colorGreen() lipgloss.Color  { t := theme.Current(); return t.SuccessColor() }
func colorGray() lipgloss.Color   { t := theme.Current(); return t.MutedColor() }

// getPadding returns current theme padding values
func getPadding() theme.PaddingConfig {
	return theme.Current().Padding
}

// Style functions - return styles with current theme colors
// These are functions rather than vars to support dynamic theme changes

func titleStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(colorCyan()).
		Bold(true)
}

func listTitleStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(colorCyan()).
		Bold(true).
		MarginLeft(-2) // Compensate for the page left padding
}

func subtitleStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(colorYellow()).
		Bold(true)
}

func highlightStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(colorYellow()).
		Bold(true)
}

func errorStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(colorRed()).
		Bold(true)
}

func successStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(colorGreen()).
		Bold(true)
}

func helpStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(colorGray())
}

func infoStyle() lipgloss.Style {
	t := theme.Current()
	return lipgloss.NewStyle().
		Foreground(t.TextColor())
}

func boldStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Bold(true)
}

func selectedItemStyle() lipgloss.Style {
	p := getPadding()
	return lipgloss.NewStyle().
		Foreground(colorYellow()).
		Bold(true).
		PaddingLeft(p.ListItemLeft)
}

func normalItemStyle() lipgloss.Style {
	p := getPadding()
	return lipgloss.NewStyle().
		PaddingLeft(p.ListItemLeft)
}

func paginationStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(colorCyan())
}

func stationNameStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(colorCyan()).
		Bold(true)
}

func stationFieldStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(colorGray())
}

func stationValueStyle() lipgloss.Style {
	t := theme.Current()
	return lipgloss.NewStyle().
		Foreground(t.TextColor())
}

func quickFavoritesStyle() lipgloss.Style {
	return titleStyle().Foreground(lipgloss.Color("99"))
}

func docStyle() lipgloss.Style {
	p := getPadding()
	return helpStyle().Padding(p.PageVertical, p.PageHorizontal)
}

func docStyleNoTopPadding() lipgloss.Style {
	p := getPadding()
	return helpStyle().PaddingTop(0).PaddingBottom(p.PageVertical).PaddingLeft(p.PageHorizontal).PaddingRight(p.PageHorizontal)
}

// createStyledDelegate creates a list delegate with single-line items and consistent styling
func createStyledDelegate() list.DefaultDelegate {
	delegate := list.NewDefaultDelegate()
	delegate.SetHeight(1)            // Single line per item
	delegate.SetSpacing(0)           // Remove spacing between items
	delegate.ShowDescription = false // Hide description text below items
	// Remove vertical padding from delegate styles
	delegate.Styles.NormalTitle = lipgloss.NewStyle()
	delegate.Styles.SelectedTitle = lipgloss.NewStyle().Foreground(colorYellow()).Bold(true)
	return delegate
}

// wrapPageWithHeader wraps content with header at the top and applies consistent padding
func wrapPageWithHeader(content string) string {
	var b strings.Builder

	// Render header using configuration
	if globalHeaderRenderer != nil {
		header := globalHeaderRenderer.Render()
		if header != "" {
			b.WriteString(header)
			// Header should already have proper spacing/newlines from renderer
		}
	} else {
		// Fallback to default if renderer not initialized
		header := lipgloss.NewStyle().
			Width(50).
			Align(lipgloss.Center).
			Foreground(colorBlue()).
			Bold(true).
			PaddingTop(1).
			Render("TERA")
		b.WriteString(header)
		b.WriteString("\n")
	}

	b.WriteString(content)
	// Use style without top padding since header already has it
	return docStyleNoTopPadding().Render(b.String())
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
		b.WriteString(titleStyle().Render(layout.Title))
	}
	b.WriteString("\n")

	// Subtitle section - always takes up one line (empty or not) for consistency
	if layout.Subtitle != "" {
		b.WriteString(subtitleStyle().Render(layout.Subtitle))
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
		b.WriteString(helpStyle().Render(layout.Help))
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
		b.WriteString(titleStyle().Render(layout.Title))
	}
	b.WriteString("\n")

	// Subtitle section - always takes up one line (empty or not) for consistency
	if layout.Subtitle != "" {
		b.WriteString(subtitleStyle().Render(layout.Subtitle))
	}
	b.WriteString("\n")

	// Main content
	if layout.Content != "" {
		b.WriteString(layout.Content)
	}

	// Calculate how many lines we've used so far
	// First, count actual header lines from the renderer
	var teraHeaderLines int
	if globalHeaderRenderer != nil {
		header := globalHeaderRenderer.Render()
		teraHeaderLines = strings.Count(header, "\n")
	} else {
		teraHeaderLines = 3 // Default fallback
	}
	
	// Count content lines
	contentLines := strings.Count(b.String(), "\n")
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
		b.WriteString(helpStyle().Render(layout.Help))
	}

	return wrapPageWithHeader(b.String())
}

// RenderStationDetails renders station details in a formatted way
func RenderStationDetails(station api.Station) string {
	return RenderStationDetailsWithVote(station, false)
}

// RenderStationDetailsWithVote renders station details with optional voted indicator
func RenderStationDetailsWithVote(station api.Station, voted bool) string {
	var s strings.Builder

	s.WriteString(fmt.Sprintf("Name:    %s\n", boldStyle().Render(station.TrimName())))

	if station.Tags != "" {
		s.WriteString(fmt.Sprintf("Tags:    %s\n", station.Tags))
	}

	if station.Country != "" {
		s.WriteString(fmt.Sprintf("Country: %s", station.Country))
		if station.State != "" {
			s.WriteString(fmt.Sprintf(", %s", station.State))
		}
		s.WriteString("\n")
	}

	if station.Language != "" {
		s.WriteString(fmt.Sprintf("Language: %s\n", station.Language))
	}

	// Votes with voted indicator
	s.WriteString(fmt.Sprintf("Votes:   %d", station.Votes))
	if voted {
		s.WriteString("  ")
		s.WriteString(successStyle().Render("✓ You voted"))
	}
	s.WriteString("\n")

	if station.Codec != "" {
		s.WriteString(fmt.Sprintf("Codec:   %s", station.Codec))
		if station.Bitrate > 0 {
			s.WriteString(fmt.Sprintf(" @ %d kbps", station.Bitrate))
		}
		s.WriteString("\n")
	}

	return s.String()
}
