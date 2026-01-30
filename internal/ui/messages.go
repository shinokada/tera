package ui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// Common message types used across UI components

// tickMsg is sent on a timer for countdown/animation purposes
type tickMsg time.Time

// backToMainMsg signals return to main menu from any screen
type backToMainMsg struct{}

// tickEverySecond returns a command that ticks every second
func tickEverySecond() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}
