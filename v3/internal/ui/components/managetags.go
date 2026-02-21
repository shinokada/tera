package components

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/shinokada/tera/v3/internal/theme"
)

// ManageTagsDoneMsg is sent when the user finishes editing tags.
type ManageTagsDoneMsg struct {
	Tags []string
}

// ManageTagsCancelledMsg is sent when the user cancels tag management.
type ManageTagsCancelledMsg struct{}

// tagEntry is a single row in the checklist.
type tagEntry struct {
	tag     string
	checked bool
}

// ManageTags is a Bubble Tea component for toggling a station's tags.
type ManageTags struct {
	stationName string
	entries     []tagEntry
	cursor      int
	addingNew   bool
	tagInput    TagInput
	width       int
}

// NewManageTags creates a ManageTags component.
//
// currentTags are listed first (checked); remaining tags from allTags are
// appended (unchecked). Duplicates within allTags are silently dropped.
func NewManageTags(stationName string, currentTags []string, allTags []string, width int) ManageTags {
	// Build ordered, deduplicated entry list.
	seen := make(map[string]bool)
	var entries []tagEntry

	// Current tags first – all checked.
	for _, t := range currentTags {
		if !seen[t] {
			entries = append(entries, tagEntry{tag: t, checked: true})
			seen[t] = true
		}
	}
	// Remaining tags from the global pool – unchecked.
	for _, t := range allTags {
		if !seen[t] {
			entries = append(entries, tagEntry{tag: t, checked: false})
			seen[t] = true
		}
	}

	return ManageTags{
		stationName: stationName,
		entries:     entries,
		cursor:      0,
		width:       width,
	}
}

// Init implements tea.Model.
func (m ManageTags) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model.
func (m ManageTags) Update(msg tea.Msg) (ManageTags, tea.Cmd) {
	// While entering a new tag, delegate all input to TagInput.
	if m.addingNew {
		var cmd tea.Cmd
		m.tagInput, cmd = m.tagInput.Update(msg)
		return m, cmd
	}

	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}

	// Total rows = entries + 1 "Add new tag" row.
	lastCursor := len(m.entries)

	switch keyMsg.Type {
	case tea.KeyEsc:
		return m, func() tea.Msg { return ManageTagsCancelledMsg{} }

	case tea.KeyUp:
		if m.cursor > 0 {
			m.cursor--
		}
		return m, nil

	case tea.KeyDown:
		if m.cursor < lastCursor {
			m.cursor++
		}
		return m, nil

	case tea.KeyEnter:
		if m.cursor == lastCursor {
			// "Add new tag" row selected.
			m.addingNew = true
			m.tagInput = NewTagInput(m.existingTags(), m.width-4)
			return m, nil
		}
		// Toggle the entry under the cursor.
		m.entries[m.cursor].checked = !m.entries[m.cursor].checked
		return m, nil

	case tea.KeyRunes:
		switch keyMsg.String() {
		case "j":
			if m.cursor < lastCursor {
				m.cursor++
			}
		case "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case " ":
			// Space only toggles entry rows, not the "Add new tag" row.
			if m.cursor < len(m.entries) {
				m.entries[m.cursor].checked = !m.entries[m.cursor].checked
			}
		case "d":
			return m, m.Done()
		case "q":
			return m, func() tea.Msg { return ManageTagsCancelledMsg{} }
		}
		return m, nil
	}

	return m, nil
}

// View implements tea.Model.
func (m ManageTags) View() string {
	if m.addingNew {
		return m.tagInput.View()
	}

	th := theme.Current()
	titleStyle := lipgloss.NewStyle().Foreground(th.PrimaryColor()).Bold(true)
	mutedStyle := lipgloss.NewStyle().Foreground(th.MutedColor())
	hlStyle := lipgloss.NewStyle().Foreground(th.HighlightColor())
	checkStyle := lipgloss.NewStyle().Foreground(th.SuccessColor())

	var sb strings.Builder
	sb.WriteString(titleStyle.Render(fmt.Sprintf("Manage Tags — %s", m.stationName)) + "\n\n")

	if len(m.entries) == 0 {
		sb.WriteString(mutedStyle.Render("No tags yet") + "\n")
	} else {
		for i, e := range m.entries {
			cursor := "  "
			if i == m.cursor {
				cursor = "> "
			}
			checkbox := "[ ]"
			if e.checked {
				checkbox = checkStyle.Render("[x]")
			}
			line := fmt.Sprintf("%s%s %s", cursor, checkbox, e.tag)
			if i == m.cursor {
				sb.WriteString(hlStyle.Render(line) + "\n")
			} else {
				sb.WriteString(line + "\n")
			}
		}
	}

	// "Add new tag" row.
	addCursor := "  "
	if m.cursor == len(m.entries) {
		addCursor = "> "
	}
	addLine := addCursor + "+ Add new tag…"
	if m.cursor == len(m.entries) {
		sb.WriteString(hlStyle.Render(addLine) + "\n")
	} else {
		sb.WriteString(mutedStyle.Render(addLine) + "\n")
	}

	sb.WriteString("\n" + mutedStyle.Render("j/k or ↑↓ navigate • Space/Enter toggle • d done • q cancel"))
	return sb.String()
}

// Done emits a ManageTagsDoneMsg with all currently checked tags.
func (m ManageTags) Done() tea.Cmd {
	tags := m.selectedTags()
	return func() tea.Msg { return ManageTagsDoneMsg{Tags: tags} }
}

// HandleTagSubmitted processes a TagSubmittedMsg from the embedded TagInput.
// If the tag already exists it is checked; otherwise it is prepended as a new
// checked entry.
func (m ManageTags) HandleTagSubmitted(tag string) (ManageTags, tea.Cmd) {
	m.addingNew = false
	if tag == "" {
		return m, nil
	}
	// Check whether the tag already exists in the list.
	for i, e := range m.entries {
		if strings.EqualFold(e.tag, tag) {
			m.entries[i].checked = true
			return m, nil
		}
	}
	// Prepend as a new checked entry.
	m.entries = append([]tagEntry{{tag: tag, checked: true}}, m.entries...)
	return m, nil
}

// HandleTagCancelled cancels the new-tag flow without adding anything.
func (m ManageTags) HandleTagCancelled() ManageTags {
	m.addingNew = false
	return m
}

// selectedTags returns all currently checked tag names.
func (m ManageTags) selectedTags() []string {
	var out []string
	for _, e := range m.entries {
		if e.checked {
			out = append(out, e.tag)
		}
	}
	return out
}

// existingTags returns all tag names in the list (checked or not).
func (m ManageTags) existingTags() []string {
	out := make([]string, len(m.entries))
	for i, e := range m.entries {
		out[i] = e.tag
	}
	return out
}
