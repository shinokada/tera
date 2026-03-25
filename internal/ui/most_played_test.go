package ui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/shinokada/tera/v3/internal/blocklist"
	"github.com/shinokada/tera/v3/internal/config"
)

// TestMostPlayedConfirmStopRouted verifies that key presses while in the
// mostPlayedStateConfirmStop state are handled (not silently dropped).
// Before the fix, mostPlayedStateConfirmStop was missing from the Update
// switch, so 'y' would do nothing and 'n'/'esc' would do nothing.
func TestMostPlayedConfirmStopRouted(t *testing.T) {
	tmpDir := t.TempDir()
	m := NewMostPlayedModel(nil, tmpDir+"/favorites", blocklist.NewManager(tmpDir+"/blocklist.json"))
	m.playOptsCfg = config.PlayOptionsConfig{ConfirmStop: true}

	// 'n' / 'esc' should return to playing state
	for _, key := range []string{"n", "esc"} {
		m.state = mostPlayedStateConfirmStop
		var msg tea.KeyMsg
		if key == "esc" {
			msg = tea.KeyMsg{Type: tea.KeyEsc}
		} else {
			msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(key)}
		}
		result, _ := m.Update(msg)
		if result.state != mostPlayedStatePlaying {
			t.Errorf("key %q: expected mostPlayedStatePlaying, got %v", key, result.state)
		}
	}

	// 'y' should return to list state and clear selected station
	m.state = mostPlayedStateConfirmStop
	yMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("y")}
	result, _ := m.Update(yMsg)
	if result.state != mostPlayedStateList {
		t.Errorf("key 'y': expected mostPlayedStateList, got %v", result.state)
	}
	if result.selectedStation != nil {
		t.Error("key 'y': expected selectedStation to be cleared")
	}
}

// TestMostPlayedConfirmStopTriggered verifies that when ConfirmStop is ON and
// the user presses Esc or '0' during playback, the model transitions to the
// confirm-stop state rather than stopping immediately.
func TestMostPlayedConfirmStopTriggered(t *testing.T) {
	tmpDir := t.TempDir()
	m := NewMostPlayedModel(nil, tmpDir+"/favorites", blocklist.NewManager(tmpDir+"/blocklist.json"))
	m.playOptsCfg = config.PlayOptionsConfig{ConfirmStop: true}

	for _, key := range []string{"esc", "0"} {
		m.state = mostPlayedStatePlaying
		var msg tea.KeyMsg
		if key == "esc" {
			msg = tea.KeyMsg{Type: tea.KeyEsc}
		} else {
			msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(key)}
		}
		result, _ := m.Update(msg)
		if result.state != mostPlayedStateConfirmStop {
			t.Errorf("key %q with ConfirmStop=true: expected mostPlayedStateConfirmStop, got %v", key, result.state)
		}
	}
}

// ── Top Rated ────────────────────────────────────────────────────────────────

// TestTopRatedConfirmStopRouted verifies that key presses in topRatedStateConfirmStop
// are properly handled.
func TestTopRatedConfirmStopRouted(t *testing.T) {
	tmpDir := t.TempDir()
	m := NewTopRatedModel(nil, nil, nil, tmpDir+"/favorites", blocklist.NewManager(tmpDir+"/blocklist.json"))
	m.playOptsCfg = config.PlayOptionsConfig{ConfirmStop: true}

	// 'n' / 'esc' should return to playing state
	for _, key := range []string{"n", "esc"} {
		m.state = topRatedStateConfirmStop
		var msg tea.KeyMsg
		if key == "esc" {
			msg = tea.KeyMsg{Type: tea.KeyEsc}
		} else {
			msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(key)}
		}
		result, _ := m.Update(msg)
		if result.state != topRatedStatePlaying {
			t.Errorf("key %q: expected topRatedStatePlaying, got %v", key, result.state)
		}
	}

	// 'y' should return to list state
	m.state = topRatedStateConfirmStop
	yMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("y")}
	result, _ := m.Update(yMsg)
	if result.state != topRatedStateList {
		t.Errorf("key 'y': expected topRatedStateList, got %v", result.state)
	}
}

// TestTopRatedConfirmStopTriggered verifies that when ConfirmStop is ON,
// Esc or '0' during playback transitions to the confirm-stop state.
func TestTopRatedConfirmStopTriggered(t *testing.T) {
	tmpDir := t.TempDir()
	m := NewTopRatedModel(nil, nil, nil, tmpDir+"/favorites", blocklist.NewManager(tmpDir+"/blocklist.json"))
	m.playOptsCfg = config.PlayOptionsConfig{ConfirmStop: true}

	for _, key := range []string{"esc", "0"} {
		m.state = topRatedStatePlaying
		var msg tea.KeyMsg
		if key == "esc" {
			msg = tea.KeyMsg{Type: tea.KeyEsc}
		} else {
			msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(key)}
		}
		result, _ := m.Update(msg)
		if result.state != topRatedStateConfirmStop {
			t.Errorf("key %q with ConfirmStop=true: expected topRatedStateConfirmStop, got %v", key, result.state)
		}
	}
}
