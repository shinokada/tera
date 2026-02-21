package ui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// Common message types used across UI components

// Message display durations (in seconds, 1 tick = 1 second)
const (
	messageDisplayShort      = 3  // 3 seconds
	messageDisplayMedium     = 5  // 5 seconds
	messageDisplayLong       = 10 // 10 seconds
	messageDisplayPersistent = -1 // never auto-clear; must be reset explicitly
)

// tickMsg is sent on a timer for countdown/animation purposes
type tickMsg time.Time

// backToMainMsg signals return to main menu from any screen
type backToMainMsg struct{}

// stationBlockedMsg is sent when a station is blocked
type stationBlockedMsg struct {
	message     string
	stationUUID string
	success     bool
}

// undoBlockSuccessMsg is sent when a block is successfully undone
type undoBlockSuccessMsg struct{}

// undoBlockFailedMsg is sent when a block undo operation fails
type undoBlockFailedMsg struct{}

// tickEverySecond returns a command that ticks every second
func tickEverySecond() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}
