package components

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/shinokada/tera/v3/internal/theme"
)

// ChecklistConfirmedMsg is sent when the user confirms their selections.
type ChecklistConfirmedMsg struct {
	Items []ChecklistItem
}

// ChecklistCancelledMsg is sent when the user cancels the checklist.
type ChecklistCancelledMsg struct{}

// ChecklistItem represents a single toggleable category row.
type ChecklistItem struct {
	// Key is the internal identifier (e.g. "favorites").
	Key string
	// Label is the display name shown in the UI (e.g. "Favorites (playlists)").
	Label string
	// Detail is optional secondary text shown to the right of the label.
	Detail string
	// Checked controls whether the item is selected.
	Checked bool
}

// ChecklistModel is a self-contained bubbletea component for multi-select
// checklists. It renders a list of toggleable items and emits
// ChecklistConfirmedMsg or ChecklistCancelledMsg when the user is done.
//
// Keybindings:
//
//	↑ / k      move cursor up
//	↓ / j      move cursor down
//	Space      toggle current item
//	a          toggle all items (select all / deselect all)
//	Enter      confirm and emit ChecklistConfirmedMsg
//	Esc / q    cancel and emit ChecklistCancelledMsg
type ChecklistModel struct {
	Title  string
	Items  []ChecklistItem
	cursor int
	width  int
}

// NewChecklistModel creates a ChecklistModel with the given title and items.
func NewChecklistModel(title string, items []ChecklistItem) ChecklistModel {
	return ChecklistModel{
		Title: title,
		Items: items,
	}
}

// SetWidth sets the render width (used for padding and alignment).
func (m *ChecklistModel) SetWidth(w int) {
	m.width = w
}

// SetItems replaces the item list and resets the cursor.
func (m *ChecklistModel) SetItems(items []ChecklistItem) {
	m.Items = items
	m.cursor = 0
}

// CheckedKeys returns the Keys of all currently checked items.
func (m ChecklistModel) CheckedKeys() []string {
	var keys []string
	for _, item := range m.Items {
		if item.Checked {
			keys = append(keys, item.Key)
		}
	}
	return keys
}

// AnyChecked returns true if at least one item is checked.
func (m ChecklistModel) AnyChecked() bool {
	for _, item := range m.Items {
		if item.Checked {
			return true
		}
	}
	return false
}

// allChecked returns true if every item is checked.
func (m ChecklistModel) allChecked() bool {
	for _, item := range m.Items {
		if !item.Checked {
			return false
		}
	}
	return true
}

// Init implements tea.Model.
func (m ChecklistModel) Init() tea.Cmd {
	return nil
}

// Update handles keyboard input.
func (m ChecklistModel) Update(msg tea.Msg) (ChecklistModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.Items)-1 {
				m.cursor++
			}
		case " ":
			if len(m.Items) > 0 {
				m.Items[m.cursor].Checked = !m.Items[m.cursor].Checked
			}
		case "a":
			// Toggle all: if everything is checked, uncheck all; otherwise check all.
			target := !m.allChecked()
			for i := range m.Items {
				m.Items[i].Checked = target
			}
		case "enter":
			return m, func() tea.Msg { return ChecklistConfirmedMsg{Items: m.Items} }
		case "esc", "q":
			return m, func() tea.Msg { return ChecklistCancelledMsg{} }
		}
	}
	return m, nil
}

// View renders the checklist.
func (m ChecklistModel) View() string {
	t := theme.Current()

	titleStyle := lipgloss.NewStyle().
		Foreground(t.HighlightColor()).
		Bold(true)

	checkedStyle := lipgloss.NewStyle().
		Foreground(t.SuccessColor())

	uncheckedStyle := lipgloss.NewStyle().
		Foreground(t.MutedColor())

	cursorStyle := lipgloss.NewStyle().
		Foreground(t.HighlightColor()).
		Bold(true)

	detailStyle := lipgloss.NewStyle().
		Foreground(t.MutedColor())

	helpStyle := lipgloss.NewStyle().
		Foreground(t.MutedColor())

	p := t.Padding
	indent := strings.Repeat(" ", p.ListItemLeft)

	var b strings.Builder

	// Title
	b.WriteString(indent)
	b.WriteString(titleStyle.Render(m.Title))
	b.WriteString("\n\n")

	// Items
	for i, item := range m.Items {
		// Cursor
		cursor := "  "
		if i == m.cursor {
			cursor = cursorStyle.Render("> ")
		}

		// Checkbox
		var checkbox string
		if item.Checked {
			checkbox = checkedStyle.Render("[x]")
		} else {
			checkbox = uncheckedStyle.Render("[ ]")
		}

		// Label
		var label string
		if i == m.cursor {
			label = cursorStyle.Render(item.Label)
		} else {
			label = item.Label
		}

		// Detail
		detail := ""
		if item.Detail != "" {
			detail = "  " + detailStyle.Render(item.Detail)
		}

		fmt.Fprintf(&b, "%s%s %s %s%s\n",
			indent, cursor, checkbox, label, detail)
	}

	// Help bar
	b.WriteString("\n")
	b.WriteString(indent)
	b.WriteString(helpStyle.Render("↑↓/jk: move   Space: toggle   a: toggle all   Enter: confirm   Esc: cancel"))
	b.WriteString("\n")

	return b.String()
}
