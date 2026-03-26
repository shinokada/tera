package player

import (
	"testing"

	"github.com/shinokada/tera/v3/internal/api"
)

func TestNewMPVPlayer(t *testing.T) {
	player := NewMPVPlayer()

	if player == nil {
		t.Fatal("Expected non-nil player")
	}

	if player.IsPlaying() {
		t.Error("Expected player to not be playing initially")
	}

	if player.GetCurrentStation() != nil {
		t.Error("Expected no current station initially")
	}
}

func TestMPVPlayer_IsPlaying(t *testing.T) {
	player := NewMPVPlayer()

	if player.IsPlaying() {
		t.Error("Expected player to not be playing initially")
	}

	// Manually set playing state for testing
	player.mu.Lock()
	player.playing = true
	player.mu.Unlock()

	if !player.IsPlaying() {
		t.Error("Expected player to be playing")
	}

	player.mu.Lock()
	player.playing = false
	player.mu.Unlock()

	if player.IsPlaying() {
		t.Error("Expected player to not be playing")
	}
}

func TestMPVPlayer_GetCurrentStation(t *testing.T) {
	player := NewMPVPlayer()

	if player.GetCurrentStation() != nil {
		t.Error("Expected no current station initially")
	}

	testStation := &api.Station{
		StationUUID: "test-123",
		Name:        "Test Station",
		URLResolved: "http://example.com/stream",
	}

	// Manually set station for testing
	player.mu.Lock()
	player.station = testStation
	player.mu.Unlock()

	station := player.GetCurrentStation()
	if station == nil {
		t.Fatal("Expected current station to be set")
	} else if station.StationUUID != "test-123" {
		t.Errorf("Expected UUID 'test-123', got '%s'", station.StationUUID)
	}
}

func TestMPVPlayer_Stop_WhenNotPlaying(t *testing.T) {
	player := NewMPVPlayer()

	// Stopping when not playing should not error
	err := player.Stop()
	if err != nil {
		t.Errorf("Expected no error when stopping while not playing, got: %v", err)
	}
}

// Note: We can't easily test actual playback without mpv installed
// and without creating integration tests. The core logic is tested above.
// Full playback testing should be done in integration tests or manually.

// ---------------------------------------------------------------------------
// Regression: race condition between Stop() and async Play()
//
// Root cause: when Play() is called asynchronously (via a tea.Cmd goroutine)
// and Stop() is called before that goroutine wakes up, playing==false so
// Stop() used to be a no-op — the subsequent Play() would start mpv anyway.
//
// Fix: Stop() now sets killed=true when the player is not yet playing.
//      Play() refuses to start if killed==true.
// ---------------------------------------------------------------------------

// TestMPVPlayer_Stop_SetsKilledFlag verifies that calling Stop() on a player
// that has not started yet sets the internal killed flag.
func TestMPVPlayer_Stop_SetsKilledFlag(t *testing.T) {
	p := NewMPVPlayer()

	err := p.Stop()
	if err != nil {
		t.Fatalf("Stop() on an idle player should not return an error: %v", err)
	}

	p.mu.Lock()
	killed := p.killed
	p.mu.Unlock()

	if !killed {
		t.Error("Stop() on a not-playing player should set killed=true to prevent a subsequent async Play()")
	}
}

// TestMPVPlayer_KilledPlayer_IgnoresPlay verifies that after Stop() sets the
// killed flag, a subsequent Play() returns nil without starting playback.
// This covers the race where Stop() is called before an async Play() runs.
func TestMPVPlayer_KilledPlayer_IgnoresPlay(t *testing.T) {
	p := NewMPVPlayer()

	// Mark the player as killed (Stop() called before async Play() ran).
	if err := p.Stop(); err != nil {
		t.Fatalf("Stop() should not fail while preparing killed state: %v", err)
	}

	station := &api.Station{
		StationUUID: "test-killed",
		Name:        "Should Not Play",
		URLResolved: "http://example.com/stream",
	}

	err := p.Play(station)
	if err != nil {
		t.Fatalf("Play() on a killed player should return nil, got: %v", err)
	}

	if p.IsPlaying() {
		t.Error("Play() on a killed player must not start playback")
	}
}

func TestMPVPlayer_ThreadSafety(t *testing.T) {
	player := NewMPVPlayer()

	// Test concurrent access to IsPlaying and GetCurrentStation
	done := make(chan bool)

	// Goroutine 1: Read state
	go func() {
		for i := 0; i < 100; i++ {
			_ = player.IsPlaying()
			_ = player.GetCurrentStation()
		}
		done <- true
	}()

	// Goroutine 2: Write state
	go func() {
		for i := 0; i < 100; i++ {
			player.mu.Lock()
			player.playing = i%2 == 0
			player.mu.Unlock()
		}
		done <- true
	}()

	// Wait for both goroutines
	<-done
	<-done

	// If we got here without data races, test passes
}
