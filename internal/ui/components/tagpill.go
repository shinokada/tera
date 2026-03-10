package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/shinokada/tera/v3/internal/theme"
)

// TagRenderer provides consistent tag display across all views.
type TagRenderer struct {
	theme      *theme.Theme
	maxDisplay int // Max tags to show inline before truncating with +N.
}

// NewTagRenderer creates a TagRenderer with a default max-display of 3.
func NewTagRenderer() *TagRenderer {
	return &TagRenderer{
		theme:      theme.Current(),
		maxDisplay: 3,
	}
}

// RenderPill returns a styled [tag] pill.
func (r *TagRenderer) RenderPill(tag string) string {
	bracketStyle := lipgloss.NewStyle().Foreground(r.theme.MutedColor())
	textStyle := lipgloss.NewStyle().Foreground(r.theme.SecondaryColor())
	return fmt.Sprintf("%s%s%s", bracketStyle.Render("["), textStyle.Render(tag), bracketStyle.Render("]"))
}

// RenderPills returns inline tag pills; shows at most maxDisplay tags then +N for the rest.
func (r *TagRenderer) RenderPills(tags []string) string {
	if len(tags) == 0 {
		return ""
	}

	var pills []string
	displayTags := tags
	overflow := 0

	if len(tags) > r.maxDisplay {
		displayTags = tags[:r.maxDisplay]
		overflow = len(tags) - r.maxDisplay
	}

	for _, tag := range displayTags {
		pills = append(pills, r.RenderPill(tag))
	}

	result := strings.Join(pills, " ")
	if overflow > 0 {
		countStyle := lipgloss.NewStyle().Foreground(r.theme.MutedColor())
		result += countStyle.Render(fmt.Sprintf(" +%d", overflow))
	}
	return result
}

// RenderList renders all tags as space-separated pills (for the info panel).
// Returns a dim "No tags" string when the slice is empty.
func (r *TagRenderer) RenderList(tags []string) string {
	if len(tags) == 0 {
		return lipgloss.NewStyle().Foreground(r.theme.MutedColor()).Render("No tags")
	}

	var pills []string
	for _, tag := range tags {
		pills = append(pills, r.RenderPill(tag))
	}
	return strings.Join(pills, " ")
}
