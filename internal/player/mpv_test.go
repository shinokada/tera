package player

import (
	"testing"

	"github.com/shinokada/tera/internal/api"
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
	}

	if station.StationUUID != "test-123" {
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
