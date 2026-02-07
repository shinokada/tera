package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/shinokada/tera/internal/storage"
)

// HeaderRenderer handles rendering the app header based on configuration
type HeaderRenderer struct {
	config storage.AppearanceConfig
}

// NewHeaderRenderer creates a new header renderer with current config
func NewHeaderRenderer() *HeaderRenderer {
	config, err := storage.LoadAppearanceConfig()
	if err != nil {
		config = storage.DefaultAppearanceConfig()
	}

	return &HeaderRenderer{
		config: config,
	}
}

// Render generates the header content based on configuration
func (h *HeaderRenderer) Render() string {
	switch h.config.Header.Mode {
	case storage.HeaderModeNone:
		return "" // No header

	case storage.HeaderModeText:
		return h.renderText()

	case storage.HeaderModeASCII:
		return h.renderASCII()

	default: // HeaderModeDefault
		return h.renderDefault()
	}
}

// renderDefault renders the default "TERA" header
func (h *HeaderRenderer) renderDefault() string {
	style := h.createBaseStyle()
	return style.Render("TERA")
}

// renderText renders custom text header
func (h *HeaderRenderer) renderText() string {
	if h.config.Header.CustomText == "" {
		return h.renderDefault() // Fallback
	}

	style := h.createBaseStyle()
	return style.Render(h.config.Header.CustomText)
}

// renderASCII renders ASCII art header
func (h *HeaderRenderer) renderASCII() string {
	if h.config.Header.AsciiArt == "" {
		return h.renderDefault() // Fallback if no ASCII art provided
	}

	return h.styleASCII(h.config.Header.AsciiArt)
}

// applyAlignment sets the alignment on a style based on config
func (h *HeaderRenderer) applyAlignment(style lipgloss.Style) lipgloss.Style {
	switch h.config.Header.Alignment {
	case "left":
		return style.Align(lipgloss.Left)
	case "right":
		return style.Align(lipgloss.Right)
	default:
		return style.Align(lipgloss.Center)
	}
}

// applyColor sets the foreground color on a style based on config
func (h *HeaderRenderer) applyColor(style lipgloss.Style) lipgloss.Style {
	if h.config.Header.Color == "auto" {
		return style.Foreground(colorBlue())
	}
	return style.Foreground(lipgloss.Color(h.config.Header.Color))
}

// createBaseStyle creates the base style for text headers
func (h *HeaderRenderer) createBaseStyle() lipgloss.Style {
	style := lipgloss.NewStyle().
		Width(h.config.Header.Width).
		PaddingTop(h.config.Header.PaddingTop).
		PaddingBottom(h.config.Header.PaddingBottom)

	style = h.applyAlignment(style)
	style = h.applyColor(style)

	// Bold
	if h.config.Header.Bold {
		style = style.Bold(true)
	}

	return style
}

// styleASCII applies styling to ASCII art
func (h *HeaderRenderer) styleASCII(art string) string {
	// Trim only newline characters, preserving intentional leading spaces
	art = strings.Trim(art, "\r\n")

	// Create style for each line
	lineStyle := lipgloss.NewStyle().Width(h.config.Header.Width)
	lineStyle = h.applyAlignment(lineStyle)
	lineStyle = h.applyColor(lineStyle)

	// Split into lines and style each
	lines := strings.Split(art, "\n")
	var styledLines []string

	// Style each line
	for _, line := range lines {
		styledLines = append(styledLines, lineStyle.Render(line))
	}

	// Build final result with proper padding
	var result strings.Builder

	// Top padding
	for i := 0; i < h.config.Header.PaddingTop; i++ {
		result.WriteString("\n")
	}

	// Styled content - join with newlines
	result.WriteString(strings.Join(styledLines, "\n"))

	// Always end with a newline after content
	result.WriteString("\n")

	// Bottom padding
	for i := 0; i < h.config.Header.PaddingBottom; i++ {
		result.WriteString("\n")
	}

	return result.String()
}

// Reload reloads the configuration (call after config changes)
func (h *HeaderRenderer) Reload() error {
	config, err := storage.LoadAppearanceConfig()
	if err != nil {
		return err
	}
	h.config = config
	return nil
}
