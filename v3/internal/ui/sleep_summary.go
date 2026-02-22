package ui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	internaltimer "github.com/shinokada/tera/v3/internal/timer"
)

// sleepExpiredMsg is sent by the App when the sleep timer fires.
type sleepExpiredMsg struct{}

// SleepSummaryModel is the full-screen summary shown after the sleep timer expires.
type SleepSummaryModel struct {
	entries       []internaltimer.SessionEntry
	totalDuration time.Duration
	setDuration   time.Duration // the duration the user originally requested
	width         int
	height        int
}

// NewSleepSummaryModel builds the summary from a completed session.
func NewSleepSummaryModel(
	session *internaltimer.SleepSession,
	setDuration time.Duration,
	width, height int,
) SleepSummaryModel {
	session.RecordStop()
	return SleepSummaryModel{
		entries:       session.Entries(),
		totalDuration: session.Total(),
		setDuration:   setDuration,
		width:         width,
		height:        height,
	}
}

// Init satisfies tea.Model.
func (m SleepSummaryModel) Init() tea.Cmd { return nil }

// Update handles navigation: 0 returns to main menu, q/Esc exits TERA.
func (m SleepSummaryModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyMsg:
		switch msg.String() {
		case "0":
			return m, func() tea.Msg { return navigateMsg{screen: screenMainMenu} }
		default:
			return m, tea.Quit
		}
	}
	return m, nil
}

// View renders the session summary.
func (m SleepSummaryModel) View() string {
	var content strings.Builder

	minutes := int(m.setDuration.Minutes())
	fmt.Fprintf(&content, "💤 Playback stopped after %d minutes\n\n", minutes)

	if len(m.entries) == 0 {
		content.WriteString("No stations were played during this session.\n")
	} else {
		content.WriteString("Stations played this session:\n\n")
		for i, e := range m.entries {
			name := e.Station.TrimName()
			dur := formatSessionDuration(e.Duration)
			fmt.Fprintf(&content, "  %d. %-40s %s\n", i+1, truncate(name, 40), dur)
		}
		fmt.Fprintf(&content, "\nTotal: %d station(s) • %s\n",
			len(m.entries), formatSessionDuration(m.totalDuration))
	}

	return RenderPageWithBottomHelp(PageLayout{
		Title:   "💤 Sleep Timer — Session Summary",
		Content: content.String(),
		Help:    "0: Main Menu • Any other key: Exit",
	}, m.height)
}

// formatSessionDuration formats a duration as "Xh Ym" or "Zm" for display.
func formatSessionDuration(d time.Duration) string {
	if d < 0 {
		d = 0
	}
	total := int(d.Minutes())
	if total == 0 {
		secs := int(d.Seconds())
		if secs < 1 {
			secs = 1
		}
		return fmt.Sprintf("%ds", secs)
	}
	hours := total / 60
	mins := total % 60
	if hours > 0 {
		return fmt.Sprintf("%dh %dm", hours, mins)
	}
	return fmt.Sprintf("%dm", mins)
}

// truncate shortens s to at most n runes, appending "…" if trimmed.
func truncate(s string, n int) string {
	runes := []rune(s)
	if len(runes) <= n {
		return s
	}
	return string(runes[:n-1]) + "…"
}
