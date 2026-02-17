package ui

import (
	"fmt"
	"strings"
	"sync"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/lipgloss"
	"github.com/shinokada/tera/v3/internal/api"
	"github.com/shinokada/tera/v3/internal/storage"
	"github.com/shinokada/tera/v3/internal/theme"
)

// Global header renderer instance (initialized in app.go)
var (
	globalHeaderRenderer *HeaderRenderer
	headerRendererMu     sync.RWMutex
)

// InitializeHeaderRenderer initializes the global header renderer
func InitializeHeaderRenderer() {
	headerRendererMu.Lock()
	defer headerRendererMu.Unlock()
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

// renderHeader renders the header with fallback (thread-safe)
func renderHeader() string {
	headerRendererMu.RLock()
	renderer := globalHeaderRenderer
	headerRendererMu.RUnlock()

	if renderer != nil {
		result := renderer.Render()
		// Ensure header ends with newline for proper layout
		if result != "" && !strings.HasSuffix(result, "\n") {
			result += "\n"
		}
		return result
	}
	// Fallback to default if renderer not initialized
	return lipgloss.NewStyle().
		Width(50).
		Align(lipgloss.Center).
		Foreground(colorBlue()).
		Bold(true).
		PaddingTop(1).
		Render("TERA") + "\n"
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
	// Assemble page content using shared helper
	content := assemblePageContent(layout)

	// Add help text (if provided)
	if layout.Help != "" {
		if layout.Content != "" {
			content += "\n"
		}
		content += helpStyle().Render(layout.Help)
	}

	// Wrap with header and styling using shared helper
	return wrapWithHeaderAndStyle(content)
}

// assemblePageContent assembles page content with consistent structure (title, subtitle, content)
func assemblePageContent(layout PageLayout) string {
	var b strings.Builder

	// Add one blank line after TERA header for breathing room
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

	return b.String()
}

// wrapWithHeaderAndStyle combines header, content, and applies styling
func wrapWithHeaderAndStyle(content string) string {
	header := renderHeader()
	var fullContent strings.Builder
	fullContent.WriteString(header)
	fullContent.WriteString(content)
	return docStyleNoTopPadding().Render(fullContent.String())
}

// RenderPageWithBottomHelp renders a page with help text at the bottom of the screen
func RenderPageWithBottomHelp(layout PageLayout, terminalHeight int) string {
	// Assemble page content
	content := assemblePageContent(layout)

	// Get the rendered header for line counting
	header := renderHeader()
	teraHeaderLines := strings.Count(header, "\n")

	// Count content lines
	contentLines := strings.Count(content, "\n")
	p := getPadding()
	totalUsed := teraHeaderLines + contentLines + p.PageVertical // padding from docStyleNoTopPadding

	// Calculate remaining space for help text to be at bottom
	// Reserve 1 line for help text itself
	remainingLines := terminalHeight - totalUsed - 1
	if remainingLines < 0 {
		remainingLines = 0
	}

	// Add spacing to push help text to bottom
	var b strings.Builder
	b.WriteString(content)
	for i := 0; i < remainingLines; i++ {
		b.WriteString("\n")
	}

	// Help text (if provided)
	if layout.Help != "" {
		b.WriteString(helpStyle().Render(layout.Help))
	}

	// Wrap with header and styling using shared helper
	return wrapWithHeaderAndStyle(b.String())
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

// RenderStationDetailsWithMetadata renders station details with play statistics
func RenderStationDetailsWithMetadata(station api.Station, voted bool, metadata *storage.StationMetadata) string {
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

	// Play statistics (only show if metadata exists and has data)
	if metadata != nil && metadata.PlayCount > 0 {
		s.WriteString("\n")
		// Use dim style for metadata to not overwhelm
		dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
		s.WriteString(dimStyle.Render(fmt.Sprintf("🎵 Played %d times", metadata.PlayCount)))
		s.WriteString("\n")
		if !metadata.LastPlayed.IsZero() {
			s.WriteString(dimStyle.Render(fmt.Sprintf("🕐 Last played: %s", storage.FormatLastPlayed(metadata.LastPlayed))))
			s.WriteString("\n")
		}
	}

	return s.String()
}
