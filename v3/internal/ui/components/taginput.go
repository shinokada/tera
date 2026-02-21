package components

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/shinokada/tera/v3/internal/theme"
)

// TagSubmittedMsg is sent when the user confirms a tag.
type TagSubmittedMsg struct {
	Tag string
}

// TagCancelledMsg is sent when the user cancels tag input.
type TagCancelledMsg struct{}

// TagInput is a Bubble Tea component for entering a tag with autocomplete.
type TagInput struct {
	input       textinput.Model
	allTags     []string
	suggestions []string
	selectedSug int
	maxSuggest  int
	width       int
}

// NewTagInput creates a TagInput pre-loaded with all known tags.
func NewTagInput(allTags []string, width int) TagInput {
	ti := textinput.New()
	ti.Placeholder = "type a tag…"
	ti.Focus()
	ti.Width = width - 4

	return TagInput{
		input:       ti,
		allTags:     allTags,
		selectedSug: -1,
		maxSuggest:  5,
		width:       width,
	}
}

// Init implements tea.Model.
func (t TagInput) Init() tea.Cmd {
	return textinput.Blink
}

// Update implements tea.Model.
func (t TagInput) Update(msg tea.Msg) (TagInput, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEsc:
			return t, func() tea.Msg { return TagCancelledMsg{} }

		case tea.KeyEnter:
			tag := t.resolveTag()
			if tag == "" {
				return t, nil
			}
			t.selectedSug = -1
			t.suggestions = nil
			return t, func() tea.Msg { return TagSubmittedMsg{Tag: tag} }

		case tea.KeyTab:
			return t.autocomplete()

		case tea.KeyUp:
			if len(t.suggestions) == 0 {
				return t, nil
			}
			if t.selectedSug <= 0 {
				t.selectedSug = len(t.suggestions) - 1
			} else {
				t.selectedSug--
			}
			return t, nil

		case tea.KeyDown:
			if len(t.suggestions) == 0 {
				return t, nil
			}
			if t.selectedSug >= len(t.suggestions)-1 {
				t.selectedSug = 0
			} else {
				t.selectedSug++
			}
			return t, nil

		default:
			// Any other key goes to the text input; reset suggestion selection.
			t.selectedSug = -1
			var cmd tea.Cmd
			t.input, cmd = t.input.Update(msg)
			t.suggestions = t.filterSuggestions(t.input.Value())
			return t, cmd
		}
	}

	var cmd tea.Cmd
	t.input, cmd = t.input.Update(msg)
	return t, cmd
}

// View implements tea.Model.
func (t TagInput) View() string {
	th := theme.Current()
	labelStyle := lipgloss.NewStyle().Foreground(th.PrimaryColor()).Bold(true)
	mutedStyle := lipgloss.NewStyle().Foreground(th.MutedColor())
	hlStyle := lipgloss.NewStyle().Foreground(th.HighlightColor())

	var sb strings.Builder
	sb.WriteString(labelStyle.Render("Add Tag") + "\n")
	sb.WriteString(t.input.View() + "\n")

	if len(t.suggestions) > 0 {
		sb.WriteString(mutedStyle.Render("Suggestions:") + "\n")
		for i, s := range t.suggestions {
			if i == t.selectedSug {
				sb.WriteString(hlStyle.Render("> "+s) + "\n")
			} else {
				sb.WriteString(mutedStyle.Render("  "+s) + "\n")
			}
		}
	} else {
		sb.WriteString(mutedStyle.Render("Enter to confirm • Esc to cancel • Tab to autocomplete"))
	}

	return sb.String()
}

// Value returns the current text-input value.
func (t TagInput) Value() string {
	return t.input.Value()
}

// filterSuggestions returns tags from allTags that start with query (case-insensitive),
// excluding exact matches and capping at maxSuggest. An empty or whitespace-only
// query returns nil.
func (t TagInput) filterSuggestions(query string) []string {
	q := strings.TrimSpace(strings.ToLower(query))
	if q == "" {
		return nil
	}

	var out []string
	for _, tag := range t.allTags {
		lower := strings.ToLower(tag)
		if lower == q {
			// Exact match: not a suggestion
			continue
		}
		if strings.HasPrefix(lower, q) {
			out = append(out, tag)
			if len(out) >= t.maxSuggest {
				break
			}
		}
	}
	return out
}

// resolveTag returns the tag to submit: the selected suggestion if one is
// highlighted, otherwise the raw text-input value (trimmed).
func (t TagInput) resolveTag() string {
	if t.selectedSug >= 0 && t.selectedSug < len(t.suggestions) {
		return t.suggestions[t.selectedSug]
	}
	return strings.TrimSpace(t.input.Value())
}

// autocomplete fills the input with the best suggestion and resets selection.
func (t TagInput) autocomplete() (TagInput, tea.Cmd) {
	if len(t.suggestions) == 0 {
		return t, nil
	}

	var chosen string
	if t.selectedSug >= 0 && t.selectedSug < len(t.suggestions) {
		chosen = t.suggestions[t.selectedSug]
	} else {
		chosen = t.suggestions[0]
	}

	t.input.SetValue(chosen)
	// Move cursor to end
	t.input.CursorEnd()
	t.selectedSug = -1
	t.suggestions = t.filterSuggestions(chosen)

	var cmd tea.Cmd
	t.input, cmd = t.input.Update(nil)
	return t, cmd
}
