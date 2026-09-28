package ui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/shinokada/tera/v3/internal/api"
	"github.com/shinokada/tera/v3/internal/player"
)

// spaceKey is the key message Bubble Tea sends for the space bar.
var spaceKey = tea.KeyMsg{Type: tea.KeySpace, Runes: []rune{' '}}

// TestMostPlayed_SpaceHandledWhilePlaying verifies Space is routed to the
// pause handler (an idle player yields a visible "Pause failed" message
// instead of the key being silently ignored).
func TestMostPlayed_SpaceHandledWhilePlaying(t *testing.T) {
	m := NewMostPlayedModel(nil, t.TempDir(), nil)
	m.state = mostPlayedStatePlaying
	m.selectedStation = &api.Station{StationUUID: "u1", Name: "Test"}

	got, cmd := m.handlePlayingInput(spaceKey)

	if !strings.Contains(got.saveMessage, "Pause failed") {
		t.Errorf("expected Space to reach TogglePause, saveMessage=%q", got.saveMessage)
	}
	if cmd == nil {
		t.Error("expected a tick command to clear the message")
	}
	if got.state != mostPlayedStatePlaying {
		t.Errorf("state changed unexpectedly: %v", got.state)
	}
}

// TestTopRated_SpaceHandledWhilePlaying verifies Space is routed to the
// pause handler on the Top Rated screen.
func TestTopRated_SpaceHandledWhilePlaying(t *testing.T) {
	m := NewTopRatedModel(nil, nil, nil, t.TempDir(), nil)
	m.player = player.NewMPVPlayer()
	m.state = topRatedStatePlaying
	m.selectedStation = &api.Station{StationUUID: "u1", Name: "Test"}

	got, cmd := m.handlePlayingInput(spaceKey)

	if !strings.Contains(got.saveMessage, "Pause failed") {
		t.Errorf("expected Space to reach TogglePause, saveMessage=%q", got.saveMessage)
	}
	if cmd == nil {
		t.Error("expected a tick command to clear the message")
	}
	if got.state != topRatedStatePlaying {
		t.Errorf("state changed unexpectedly: %v", got.state)
	}
}
