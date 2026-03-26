package ui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/shinokada/tera/v3/internal/api"
	"github.com/shinokada/tera/v3/internal/player"
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

// sleepTimerActivateMsg is sent by a player screen when the user presses Z to set.
type sleepTimerActivateMsg struct {
	Minutes int
}

// sleepTimerCancelMsg is sent by a player screen when the user presses Z to cancel.
type sleepTimerCancelMsg struct{}

// sleepTimerExtendMsg is sent by a player screen when the user presses + to extend.
// Minutes carries the extension duration so the handler does not need a hard-coded value.
type sleepTimerExtendMsg struct {
	Minutes int
}

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

// handoffPlaybackMsg is sent by a play screen when ContinueOnNavigate is on
// and the user navigates away. App takes ownership of the player and station.
type handoffPlaybackMsg struct {
	player       *player.MPVPlayer
	station      *api.Station
	contextLabel string
}

// stopActivePlaybackMsg is sent when any screen wants to stop the app-level
// active player (e.g. main menu Esc while a handoff is in progress, or a new
// station starting on any screen).
type stopActivePlaybackMsg struct{}

// tickEverySecond returns a command that ticks every second
func tickEverySecond() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}
