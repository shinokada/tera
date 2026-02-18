package components

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/shinokada/tera/v3/internal/theme"
)

// StarRenderer provides consistent star rating display across the UI
type StarRenderer struct {
	useUnicode bool
	filledStar string
	emptyStar  string
}

// NewStarRenderer creates a new StarRenderer
// Set useUnicode to false for terminals that don't support unicode well
func NewStarRenderer(useUnicode bool) *StarRenderer {
	if useUnicode {
		return &StarRenderer{
			useUnicode: true,
			filledStar: "★",
			emptyStar:  "☆",
		}
	}
	return &StarRenderer{
		useUnicode: false,
		filledStar: "*",
		emptyStar:  "-",
	}
}

// accentStyle returns the style for filled stars (uses highlight/accent color)
func (s *StarRenderer) accentStyle() lipgloss.Style {
	t := theme.Current()
	return lipgloss.NewStyle().Foreground(t.HighlightColor())
}

// dimStyle returns the style for empty stars (uses muted color)
func (s *StarRenderer) dimStyle() lipgloss.Style {
	t := theme.Current()
	return lipgloss.NewStyle().Foreground(t.MutedColor())
}

// Render returns a styled string of stars for the given rating (e.g., "★ ★ ★ ★ ☆")
// Returns 5 stars with filled ones for the rating value, space-separated
func (s *StarRenderer) Render(rating int) string {
	if rating < 0 {
		rating = 0
	}
	if rating > 5 {
		rating = 5
	}

	var parts []string
	for i := 0; i < rating; i++ {
		parts = append(parts, s.accentStyle().Render(s.filledStar))
	}
	for i := rating; i < 5; i++ {
		parts = append(parts, s.dimStyle().Render(s.emptyStar))
	}
	return strings.Join(parts, " ")
}

// RenderCompact returns only filled stars with spacing (e.g., "★ ★ ★")
// Returns empty string for unrated (rating 0 or invalid)
func (s *StarRenderer) RenderCompact(rating int) string {
	if rating < 1 || rating > 5 {
		return ""
	}

	var parts []string
	for i := 0; i < rating; i++ {
		parts = append(parts, s.filledStar)
	}
	return s.accentStyle().Render(strings.Join(parts, " "))
}

// RenderWithLabel returns stars with label (e.g., "★★★★☆ (4/5)")
func (s *StarRenderer) RenderWithLabel(rating int) string {
	stars := s.Render(rating)
	label := s.dimStyle().Render(" (" + string(rune('0'+rating)) + "/5)")
	return stars + label
}

// RenderPlain returns uncolored stars with spacing (for use in contexts where styling is applied elsewhere)
func (s *StarRenderer) RenderPlain(rating int) string {
	if rating < 0 {
		rating = 0
	}
	if rating > 5 {
		rating = 5
	}

	var parts []string
	for i := 0; i < rating; i++ {
		parts = append(parts, s.filledStar)
	}
	for i := rating; i < 5; i++ {
		parts = append(parts, s.emptyStar)
	}
	return strings.Join(parts, " ")
}

// RenderCompactPlain returns only filled stars without styling, space-separated
func (s *StarRenderer) RenderCompactPlain(rating int) string {
	if rating < 1 || rating > 5 {
		return ""
	}
	var parts []string
	for i := 0; i < rating; i++ {
		parts = append(parts, s.filledStar)
	}
	return strings.Join(parts, " ")
}

// Width returns the character width of a full star rating display (5 stars + 4 spaces)
func (s *StarRenderer) Width() int {
	return 9 // 5 stars + 4 spaces
}

// FilledStar returns the filled star character
func (s *StarRenderer) FilledStar() string {
	return s.filledStar
}

// EmptyStar returns the empty star character
func (s *StarRenderer) EmptyStar() string {
	return s.emptyStar
}

// DefaultStarRenderer returns the default star renderer (with unicode)
func DefaultStarRenderer() *StarRenderer {
	return NewStarRenderer(true)
}
