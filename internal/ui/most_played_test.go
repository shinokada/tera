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

// TestMostPlayedConfirmStopResolved verifies that confirming the stop prompt
// fully resolves the model (state reset, confirmStopTarget cleared) for both
// the Esc ("back") and 0 ("main") paths.
func TestMostPlayedConfirmStopResolved(t *testing.T) {
	tmpDir := t.TempDir()

	// Esc → confirm state → y: should land on mostPlayedStateList with clean target.
	m := NewMostPlayedModel(nil, tmpDir+"/favorites", blocklist.NewManager(tmpDir+"/blocklist.json"))
	m.playOptsCfg = config.PlayOptionsConfig{ConfirmStop: true}
	m.state = mostPlayedStatePlaying
	m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	// Enter confirm state manually to avoid needing a real player.
	m.state = mostPlayedStateConfirmStop
	m.confirmStopTarget = "back"
	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("y")})
	if result.state != mostPlayedStateList {
		t.Errorf("esc→y: expected mostPlayedStateList, got %v", result.state)
	}
	if result.confirmStopTarget != "" {
		t.Errorf("esc→y: expected confirmStopTarget to be cleared, got %q", result.confirmStopTarget)
	}

	// 0 → confirm state → y: should not be stuck in mostPlayedStateConfirmStop.
	m2 := NewMostPlayedModel(nil, tmpDir+"/favorites", blocklist.NewManager(tmpDir+"/blocklist.json"))
	m2.playOptsCfg = config.PlayOptionsConfig{ConfirmStop: true}
	m2.state = mostPlayedStateConfirmStop
	m2.confirmStopTarget = "main"
	result2, _ := m2.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("y")})
	if result2.state == mostPlayedStateConfirmStop {
		t.Error("0→y: model stuck in mostPlayedStateConfirmStop after confirmation")
	}
	if result2.confirmStopTarget != "" {
		t.Errorf("0→y: expected confirmStopTarget to be cleared, got %q", result2.confirmStopTarget)
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

// TestTopRatedConfirmStopResolved verifies that confirming the stop prompt
// fully resolves the model (state reset, confirmStopTarget cleared) for both
// the Esc ("back") and 0 ("main") paths.
func TestTopRatedConfirmStopResolved(t *testing.T) {
	tmpDir := t.TempDir()

	// Esc path: confirm → y should land on topRatedStateList with a clean target.
	m := NewTopRatedModel(nil, nil, nil, tmpDir+"/favorites", blocklist.NewManager(tmpDir+"/blocklist.json"))
	m.playOptsCfg = config.PlayOptionsConfig{ConfirmStop: true}
	m.state = topRatedStateConfirmStop
	m.confirmStopTarget = "back"
	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("y")})
	if result.state != topRatedStateList {
		t.Errorf("esc→y: expected topRatedStateList, got %v", result.state)
	}
	if result.confirmStopTarget != "" {
		t.Errorf("esc→y: expected confirmStopTarget to be cleared, got %q", result.confirmStopTarget)
	}

	// 0 path: confirm → y should not be stuck in topRatedStateConfirmStop.
	m2 := NewTopRatedModel(nil, nil, nil, tmpDir+"/favorites", blocklist.NewManager(tmpDir+"/blocklist.json"))
	m2.playOptsCfg = config.PlayOptionsConfig{ConfirmStop: true}
	m2.state = topRatedStateConfirmStop
	m2.confirmStopTarget = "main"
	result2, _ := m2.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("y")})
	if result2.state == topRatedStateConfirmStop {
		t.Error("0→y: model stuck in topRatedStateConfirmStop after confirmation")
	}
	if result2.confirmStopTarget != "" {
		t.Errorf("0→y: expected confirmStopTarget to be cleared, got %q", result2.confirmStopTarget)
	}
}
