package components

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"

	tea "github.com/charmbracelet/bubbletea"
)

// SleepTimerSelectedMsg is emitted when the user confirms a duration.
type SleepTimerSelectedMsg struct {
	Minutes int
}

// SleepTimerCancelledMsg is emitted when the user cancels the dialog.
type SleepTimerCancelledMsg struct{}

// sleepTimerState tracks which sub-view the dialog is in.
type sleepTimerState int

const (
	sleepTimerStatePresets sleepTimerState = iota
	sleepTimerStateCustom
)

// presetMinutes are the quick-select options shown to the user.
var presetMinutes = []int{15, 30, 45, 60, 90}

// SleepTimerDialog is a self-contained Bubble Tea component that lets the user
// choose a sleep timer duration from presets or enter a custom value.
type SleepTimerDialog struct {
	state       sleepTimerState
	cursor      int    // selected preset index (0..len(presetMinutes))
	customInput string // raw text for custom input
	customErr   string // validation error
	width       int
}

// NewSleepTimerDialog creates a dialog with the given default selection.
// lastMinutes is pre-selected if it matches a preset, otherwise Custom is highlighted.
func NewSleepTimerDialog(lastMinutes, width int) SleepTimerDialog {
	cursor := 0
	for i, m := range presetMinutes {
		if m == lastMinutes {
			cursor = i
			break
		}
	}
	return SleepTimerDialog{
		state:  sleepTimerStatePresets,
		cursor: cursor,
		width:  width,
	}
}

// Update handles keyboard input for the dialog.
func (d SleepTimerDialog) Update(msg tea.KeyMsg) (SleepTimerDialog, tea.Cmd) {
	switch d.state {
	case sleepTimerStatePresets:
		return d.updatePresets(msg)
	case sleepTimerStateCustom:
		return d.updateCustom(msg)
	}
	return d, nil
}

func (d SleepTimerDialog) updatePresets(msg tea.KeyMsg) (SleepTimerDialog, tea.Cmd) {
	// Total rows: one per preset + one "Custom..." option
	total := len(presetMinutes) + 1
	customIdx := len(presetMinutes)

	switch msg.String() {
	case "up", "k":
		if d.cursor > 0 {
			d.cursor--
		}
	case "down", "j":
		if d.cursor < total-1 {
			d.cursor++
		}
	case "enter":
		if d.cursor == customIdx {
			d.state = sleepTimerStateCustom
			d.customInput = ""
			d.customErr = ""
		} else {
			minutes := presetMinutes[d.cursor]
			return d, func() tea.Msg { return SleepTimerSelectedMsg{Minutes: minutes} }
		}
	case "esc":
		return d, func() tea.Msg { return SleepTimerCancelledMsg{} }
	}
	return d, nil
}

func (d SleepTimerDialog) updateCustom(msg tea.KeyMsg) (SleepTimerDialog, tea.Cmd) {
	switch msg.String() {
	case "enter":
		minutes, err := validateCustomInput(d.customInput)
		if err != nil {
			d.customErr = err.Error()
			return d, nil
		}
		return d, func() tea.Msg { return SleepTimerSelectedMsg{Minutes: minutes} }
	case "esc":
		// Back to preset list
		d.state = sleepTimerStatePresets
		d.customErr = ""
	case "backspace":
		if len(d.customInput) > 0 {
			d.customInput = d.customInput[:len(d.customInput)-1]
			d.customErr = ""
		}
	default:
		// Only allow digit characters
		if len(msg.String()) == 1 && unicode.IsDigit(rune(msg.String()[0])) {
			if len(d.customInput) < 3 { // max 480 fits in 3 digits
				d.customInput += msg.String()
				d.customErr = ""
			}
		}
	}
	return d, nil
}

// validateCustomInput parses and validates the raw custom input string.
func validateCustomInput(raw string) (int, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 0, fmt.Errorf("✗ Enter a number between 1 and 480")
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n < 1 || n > 480 {
		return 0, fmt.Errorf("✗ Enter a number between 1 and 480")
	}
	return n, nil
}

// View renders the sleep timer dialog.
func (d SleepTimerDialog) View() string {
	var b strings.Builder

	switch d.state {
	case sleepTimerStatePresets:
		b.WriteString("Stop playback after:\n\n")
		for i, m := range presetMinutes {
			cursor := "  "
			if i == d.cursor {
				cursor = "> "
			}
			fmt.Fprintf(&b, "%s%d minutes\n", cursor, m)
		}
		// Custom option
		customIdx := len(presetMinutes)
		cursor := "  "
		if d.cursor == customIdx {
			cursor = "> "
		}
		fmt.Fprintf(&b, "%sCustom...\n", cursor)
		b.WriteString("\nEnter: Set • ↑↓/jk: Navigate • Esc: Cancel")

	case sleepTimerStateCustom:
		b.WriteString("Enter duration in minutes:\n\n")
		fmt.Fprintf(&b, "> %s█\n", d.customInput)
		if d.customErr != "" {
			b.WriteString("\n")
			b.WriteString(d.customErr)
			b.WriteString("\n")
		}
		b.WriteString("\nEnter: Set • Esc: Back")
	}

	return b.String()
}
