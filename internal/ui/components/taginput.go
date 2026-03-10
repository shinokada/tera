package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/shinokada/tera/v3/internal/theme"
)

// TagSubmittedMsg is dispatched when the user confirms a tag.
type TagSubmittedMsg struct{ Tag string }

// TagCancelledMsg is dispatched when the user cancels tag input.
type TagCancelledMsg struct{}

// TagInput is a self-contained BubbleTea component for entering a single tag
// with prefix-based autocomplete drawn from an existing tag list.
type TagInput struct {
	input       textinput.Model
	allTags     []string // existing tags available for autocomplete
	suggestions []string // filtered suggestions matching current input
	selectedSug int      // highlighted suggestion index (-1 = none)
	width       int
	maxSuggest  int // max suggestions to display
}

// NewTagInput creates a new TagInput with autocomplete sourced from allTags.
func NewTagInput(allTags []string, width int) TagInput {
	ti := textinput.New()
	ti.Placeholder = "e.g. chill vibes, gym workout..."
	ti.CharLimit = 50
	ti.Width = width - 6
	ti.Focus()

	return TagInput{
		input:       ti,
		allTags:     allTags,
		maxSuggest:  5,
		width:       width,
		selectedSug: -1,
	}
}

// Init satisfies the BubbleTea model interface.
func (t TagInput) Init() tea.Cmd {
	return textinput.Blink
}

// Update handles key messages for the tag input component.
func (t TagInput) Update(msg tea.Msg) (TagInput, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEsc:
			return t, func() tea.Msg { return TagCancelledMsg{} }

		case tea.KeyEnter:
			tag := strings.TrimSpace(t.input.Value())
			if t.selectedSug >= 0 && t.selectedSug < len(t.suggestions) {
				tag = t.suggestions[t.selectedSug]
			}
			if tag == "" {
				return t, nil
			}
			return t, func() tea.Msg { return TagSubmittedMsg{Tag: tag} }

		case tea.KeyTab:
			if len(t.suggestions) > 0 {
				top := t.suggestions[0]
				if t.selectedSug >= 0 && t.selectedSug < len(t.suggestions) {
					top = t.suggestions[t.selectedSug]
				}
				t.input.SetValue(top)
				t.input.CursorEnd()
				t.selectedSug = -1
				t.suggestions = t.filterSuggestions(top)
			}
			return t, nil

		case tea.KeyUp:
			if len(t.suggestions) > 0 {
				if t.selectedSug <= 0 {
					t.selectedSug = len(t.suggestions) - 1
				} else {
					t.selectedSug--
				}
			}
			return t, nil

		case tea.KeyDown:
			if len(t.suggestions) > 0 {
				t.selectedSug = (t.selectedSug + 1) % len(t.suggestions)
			}
			return t, nil
		}
	}

	// Delegate all other keys to the underlying text input.
	var cmd tea.Cmd
	t.input, cmd = t.input.Update(msg)
	t.suggestions = t.filterSuggestions(t.input.Value())
	t.selectedSug = -1
	return t, cmd
}

// filterSuggestions returns tags that start with query (prefix match).
func (t TagInput) filterSuggestions(query string) []string {
	query = strings.ToLower(strings.TrimSpace(query))
	if query == "" {
		return nil
	}
	var results []string
	for _, tag := range t.allTags {
		if strings.HasPrefix(tag, query) && tag != query {
			results = append(results, tag)
			if len(results) >= t.maxSuggest {
				break
			}
		}
	}
	return results
}

// View renders the tag input box with autocomplete suggestions.
func (t TagInput) View() string {
	th := theme.Current()

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(th.HighlightColor()).
		Padding(0, 1).
		Width(t.width - 2)

	labelStyle := lipgloss.NewStyle().
		Foreground(th.HighlightColor()).
		Bold(true)

	dimStyle := lipgloss.NewStyle().
		Foreground(th.MutedColor())

	var sb strings.Builder
	sb.WriteString(labelStyle.Render("Add Tag"))
	sb.WriteString("\n\n")
	sb.WriteString(t.input.View())

	if len(t.suggestions) > 0 {
		sb.WriteString("\n\n")
		sb.WriteString(dimStyle.Render("Suggestions:"))
		for i, sug := range t.suggestions {
			sb.WriteString("\n")
			if i == t.selectedSug {
				sb.WriteString(lipgloss.NewStyle().
					Foreground(th.HighlightColor()).
					Bold(true).
					Render(fmt.Sprintf("  > %s", sug)))
			} else {
				sb.WriteString(dimStyle.Render(fmt.Sprintf("    %s", sug)))
			}
		}
		sb.WriteString("\n\n")
		sb.WriteString(dimStyle.Render("Tab: Complete • ↑↓: Navigate"))
	} else {
		sb.WriteString("\n\n")
		sb.WriteString(dimStyle.Render("Enter: Add • Esc: Cancel"))
	}

	return boxStyle.Render(sb.String())
}

// Value returns the current raw text input value.
func (t TagInput) Value() string {
	return t.input.Value()
}
