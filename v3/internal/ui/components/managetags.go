package components

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/shinokada/tera/v3/internal/theme"
)

// ManageTagsDoneMsg is dispatched when the user finishes managing tags.
// Tags contains the final (possibly modified) tag list for the station.
type ManageTagsDoneMsg struct {
	Tags []string
}

// ManageTagsCancelledMsg is dispatched when the user cancels without changes.
type ManageTagsCancelledMsg struct{}

// tagEntry holds a tag and its checked state for display.
type tagEntry struct {
	tag     string
	checked bool // true = currently on this station
}

// ManageTags is a bubbletea component for toggling tags on/off for a station.
//
//	Usage:
//	  1. Create via NewManageTags.
//	  2. Forward tea.KeyMsg to Update(); it may emit ManageTagsDoneMsg or ManageTagsCancelledMsg.
//	  3. Render via View().
type ManageTags struct {
	stationName string
	entries     []tagEntry // merged list: current tags first, then available
	cursor      int
	width       int
	addingNew   bool       // true when "Add new tag…" row is selected and Enter was pressed
	tagInput    TagInput   // reused for adding a brand-new tag
}

// NewManageTags creates a ManageTags dialog.
//
//	currentTags  – tags already assigned to the station.
//	allTags      – every tag known to the system (for autocomplete / listing).
//	stationName  – display name shown in the dialog title.
//	width        – terminal width used for layout.
func NewManageTags(stationName string, currentTags, allTags []string, width int) ManageTags {
	// Build a set of current tags for fast lookup.
	currentSet := make(map[string]bool, len(currentTags))
	for _, t := range currentTags {
		currentSet[t] = true
	}

	// Current tags first (checked), then remaining existing tags (unchecked).
	var entries []tagEntry
	for _, t := range currentTags {
		entries = append(entries, tagEntry{tag: t, checked: true})
	}
	for _, t := range allTags {
		if !currentSet[t] {
			entries = append(entries, tagEntry{tag: t, checked: false})
		}
	}

	return ManageTags{
		stationName: stationName,
		entries:     entries,
		width:       width,
	}
}

// Init satisfies the bubbletea model interface.
func (m ManageTags) Init() tea.Cmd { return nil }

// Update handles keyboard input.
func (m ManageTags) Update(msg tea.Msg) (ManageTags, tea.Cmd) {
	// If we're in "add new tag" sub-mode, delegate to TagInput.
	if m.addingNew {
		var cmd tea.Cmd
		m.tagInput, cmd = m.tagInput.Update(msg)
		return m, cmd
	}

	kMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}

	// Total items = len(entries) + 1 for the "Add new tag…" row.
	total := len(m.entries) + 1

	switch kMsg.Type {
	case tea.KeyEsc:
		return m, func() tea.Msg { return ManageTagsCancelledMsg{} }

	case tea.KeyEnter:
		if m.cursor == len(m.entries) {
			// "Add new tag…" row selected.
			m.addingNew = true
			allTags := m.existingTags()
			m.tagInput = NewTagInput(allTags, m.width-4)
			return m, nil
		}
		// Toggle the selected entry.
		m.entries[m.cursor].checked = !m.entries[m.cursor].checked
		return m, nil

	case tea.KeyRunes:
		switch kMsg.String() {
		case " ":
			if m.cursor < len(m.entries) {
				m.entries[m.cursor].checked = !m.entries[m.cursor].checked
			}
		case "j":
			if m.cursor < total-1 {
				m.cursor++
			}
		case "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "d":
			// Submit with the current selections.
			return m, func() tea.Msg { return ManageTagsDoneMsg{Tags: m.selectedTags()} }
		case "q":
			return m, func() tea.Msg { return ManageTagsCancelledMsg{} }
		}

	case tea.KeyUp:
		if m.cursor > 0 {
			m.cursor--
		}
	case tea.KeyDown:
		if m.cursor < total-1 {
			m.cursor++
		}
	}

	return m, nil
}

// HandleTagSubmitted processes a TagSubmittedMsg from the inner TagInput.
// The caller must call this when it receives a TagSubmittedMsg while in
// ManageTags mode.  Returns (updated model, save-cmd).
func (m ManageTags) HandleTagSubmitted(tag string) (ManageTags, tea.Cmd) {
	m.addingNew = false
	if tag == "" {
		return m, nil
	}
	// Check if already in entries.
	for i, e := range m.entries {
		if e.tag == tag {
			m.entries[i].checked = true
			return m, nil
		}
	}
	// Prepend new tag as checked.
	m.entries = append([]tagEntry{{tag: tag, checked: true}}, m.entries...)
	return m, nil
}

// HandleTagCancelled cancels the sub-input mode.
func (m ManageTags) HandleTagCancelled() ManageTags {
	m.addingNew = false
	return m
}

// Done emits ManageTagsDoneMsg with the current selections.
func (m ManageTags) Done() tea.Cmd {
	tags := m.selectedTags()
	return func() tea.Msg { return ManageTagsDoneMsg{Tags: tags} }
}

// View renders the dialog box.
func (m ManageTags) View() string {
	// If sub-input is active, show it instead.
	if m.addingNew {
		return m.tagInput.View()
	}

	th := theme.Current()

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(th.HighlightColor()).
		Padding(0, 2).
		Width(m.width - 4)

	labelStyle := lipgloss.NewStyle().
		Foreground(th.HighlightColor()).
		Bold(true)

	dimStyle := lipgloss.NewStyle().
		Foreground(th.MutedColor())

	checkedStyle := lipgloss.NewStyle().
		Foreground(th.SecondaryColor())

	selectedStyle := lipgloss.NewStyle().
		Foreground(th.HighlightColor()).
		Bold(true)

	var sb strings.Builder
	sb.WriteString(labelStyle.Render(fmt.Sprintf("Manage Tags: %s", m.stationName)))
	sb.WriteString("\n\n")

	if len(m.entries) == 0 {
		sb.WriteString(dimStyle.Render("No tags yet."))
		sb.WriteString("\n")
	} else {
		sb.WriteString(dimStyle.Render("Space/Enter: Toggle  ↑↓/jk: Navigate  d: Done  Esc: Cancel"))
		sb.WriteString("\n\n")
		for i, e := range m.entries {
			label := e.tag
			var box string
			if e.checked {
				box = checkedStyle.Render("[✓]")
			} else {
				box = dimStyle.Render("[ ]")
			}
			line := fmt.Sprintf("%s %s", box, label)
			if i == m.cursor {
				sb.WriteString(selectedStyle.Render("> " + line))
			} else {
				sb.WriteString("  " + line)
			}
			sb.WriteString("\n")
		}
	}

	// "Add new tag…" row.
	addRow := "[+ Add new tag...]"
	if m.cursor == len(m.entries) {
		sb.WriteString(selectedStyle.Render("> " + addRow))
	} else {
		sb.WriteString(dimStyle.Render("  " + addRow))
	}
	sb.WriteString("\n")

	return boxStyle.Render(sb.String())
}

// selectedTags returns a slice of all checked tags.
func (m ManageTags) selectedTags() []string {
	var out []string
	for _, e := range m.entries {
		if e.checked {
			out = append(out, e.tag)
		}
	}
	return out
}

// existingTags returns all tag strings (for autocomplete).
func (m ManageTags) existingTags() []string {
	out := make([]string, len(m.entries))
	for i, e := range m.entries {
		out[i] = e.tag
	}
	return out
}
