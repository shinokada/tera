package components

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// HelpSection represents a section in the help menu
type HelpSection struct {
	Title string
	Items []HelpItem
}

// HelpItem represents a single help item
type HelpItem struct {
	Key         string
	Description string
}

// HelpModel represents the help overlay
type HelpModel struct {
	sections []HelpSection
	visible  bool
	width    int
	height   int
}

// NewHelpModel creates a new help model
func NewHelpModel(sections []HelpSection) HelpModel {
	return HelpModel{
		sections: sections,
		visible:  false,
	}
}

// Toggle toggles the help visibility
func (m *HelpModel) Toggle() {
	m.visible = !m.visible
}

// Show shows the help overlay
func (m *HelpModel) Show() {
	m.visible = true
}

// Hide hides the help overlay
func (m *HelpModel) Hide() {
	m.visible = false
}

// IsVisible returns whether the help is visible
func (m *HelpModel) IsVisible() bool {
	return m.visible
}

// SetSize sets the help overlay size
func (m *HelpModel) SetSize(width, height int) {
	m.width = width
	m.height = height
}

// Update handles help model updates
func (m HelpModel) Update(msg tea.Msg) (HelpModel, tea.Cmd) {
	if !m.visible {
		return m, nil
	}

	switch msg.(type) {
	case tea.KeyMsg:
		// Any key closes the help
		m.visible = false
		return m, nil
	}

	return m, nil
}

// View renders the help overlay
func (m HelpModel) View() string {
	if !m.visible {
		return ""
	}

	var b strings.Builder

	// Header style
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("86")). // Cyan
		Align(lipgloss.Center).
		Padding(0, 1)

	// Section title style
	sectionStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("220")). // Yellow
		MarginTop(1)

	// Key style - right-aligned for clean column look
	keyStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("86")). // Cyan
		Width(12).
		Align(lipgloss.Right).
		Bold(true)

	// Description style - left-aligned
	descStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("252")). // Light gray
		PaddingLeft(1)

	// Footer style
	footerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")). // Gray
		Italic(true).
		MarginTop(1)

	// Content width for centering titles
	contentWidth := 44

	// Build help content
	header := lipgloss.NewStyle().Width(contentWidth).Align(lipgloss.Center).Render(headerStyle.Render("═══ TERA Help ═══"))
	b.WriteString(header)
	b.WriteString("\n")

	for _, section := range m.sections {
		b.WriteString(sectionStyle.Render(section.Title))
		b.WriteString("\n")

		for _, item := range section.Items {
			line := keyStyle.Render(item.Key+":") + descStyle.Render(item.Description)
			b.WriteString(line)
			b.WriteString("\n")
		}
	}

	footer := lipgloss.NewStyle().Width(contentWidth).Align(lipgloss.Center).Render(footerStyle.Render("Press any key to close"))
	b.WriteString(footer)

	// Create a box around the content
	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("86")). // Cyan border
		Padding(1, 2)

	content := boxStyle.Render(b.String())

	// Center the help box on screen
	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		content,
	)
}

// CreateMainMenuHelp creates help sections for the main menu
func CreateMainMenuHelp() []HelpSection {
	return []HelpSection{
		{
			Title: "Navigation",
			Items: []HelpItem{
				{"↑↓/jk", "Navigate lists"},
				{"Enter", "Select/Play"},
				{"1-6", "Jump to menu"},
				{"10+", "Quick play"},
				{"Esc", "Stop & Back"},
				{"0", "Main Menu"},
				{"Ctrl+C", "Quit"},
			},
		},
		{
			Title: "Playback Controls",
			Items: []HelpItem{
				{"Space", "Pause/Resume"},
				{"/*", "Adjust volume"},
				{"m", "Toggle mute"},
				{"b", "Block station"},
				{"u", "Undo block"},
			},
		},
	}
}

// CreateFavoritesHelp creates help sections for playing from favorites (no save to list)
func CreateFavoritesHelp() []HelpSection {
	return []HelpSection{
		{
			Title: "Navigation",
			Items: []HelpItem{
				{"Esc", "Stop & Back"},
				{"0", "Main Menu"},
				{"Ctrl+C", "Quit"},
			},
		},
		{
			Title: "Playback Controls",
			Items: []HelpItem{
				{"Space", "Pause/Resume"},
				{"/*", "Adjust volume"},
				{"m", "Toggle mute"},
			},
		},
		{
			Title: "Actions",
			Items: []HelpItem{
				{"f", "Save to Favorites"},
				{"r", "Rate station (1-5)"},
				{"t", "Add tag"},
				{"T", "Manage tags"},
				{"v", "Vote"},
				{"b", "Block station"},
				{"u", "Undo block"},
				{"Z", "Sleep timer"},
				{"+", "Extend sleep timer"},
			},
		},
	}
}

// CreatePlayingHelp creates help sections for playing screens (search/lucky results)
func CreatePlayingHelp() []HelpSection {
	return []HelpSection{
		{
			Title: "Navigation",
			Items: []HelpItem{
				{"Esc", "Stop & Back"},
				{"0", "Main Menu"},
				{"Ctrl+C", "Quit"},
			},
		},
		{
			Title: "Playback Controls",
			Items: []HelpItem{
				{"Space", "Pause/Resume"},
				{"/*", "Adjust volume"},
				{"m", "Toggle mute"},
			},
		},
		{
			Title: "Actions",
			Items: []HelpItem{
				{"f", "Save to Favorites"},
				{"s", "Save to List"},
				{"r", "Rate station (1-5)"},
				{"t", "Add tag"},
				{"T", "Manage tags"},
				{"v", "Vote"},
				{"b", "Block station"},
				{"u", "Undo block"},
				{"Z", "Sleep timer"},
				{"+", "Extend sleep timer"},
			},
		},
	}
}

// CreateTagsPlayingHelp creates help sections for tag-based playing screens
// (Browse by Tag, Tag Playlists) which have no voting, block, or sleep timer.
func CreateTagsPlayingHelp() []HelpSection {
	return []HelpSection{
		{
			Title: "Navigation",
			Items: []HelpItem{
				{"Esc", "Stop & Back"},
				{"0", "Main Menu"},
			},
		},
		{
			Title: "Playback Controls",
			Items: []HelpItem{
				{"Space", "Pause/Resume"},
				{"/*", "Adjust volume"},
				{"m", "Toggle mute"},
			},
		},
		{
			Title: "Actions",
			Items: []HelpItem{
				{"r", "Rate station (1-5)"},
				{"t", "Add tag"},
				{"T", "Manage tags"},
			},
		},
	}
}

func CreateAppearanceHelp() []HelpSection {
	return []HelpSection{
		{
			Title: "Navigation",
			Items: []HelpItem{
				{"↑↓/jk", "Navigate menu"},
				{"Enter", "Select/Edit"},
				{"Esc/q", "Back"},
				{"?", "Toggle help"},
			},
		},
		{
			Title: "Header Modes",
			Items: []HelpItem{
				{"default", "Show 'TERA'"},
				{"text", "Custom text"},
				{"ascii", "ASCII art"},
				{"none", "No header"},
			},
		},
		{
			Title: "Actions",
			Items: []HelpItem{
				{"Save", "Apply changes"},
				{"Preview", "See result"},
				{"Reset", "Restore defaults"},
			},
		},
	}
}
