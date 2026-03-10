package ui

import (
	"fmt"
	"testing"

	"github.com/shinokada/tera/v3/internal/api"
	"github.com/shinokada/tera/v3/internal/config"
	"github.com/shinokada/tera/v3/internal/storage"
)

// newTestApp creates a minimal App for unit testing (no filesystem side-effects).
func newTestApp() *App {
	return &App{
		screen:         screenMainMenu,
		playHistoryCfg: config.DefaultPlayHistoryConfig(),
	}
}

// makeRecentlyPlayed builds a slice of StationWithMetadata from names.
func makeRecentlyPlayed(names ...string) []storage.StationWithMetadata {
	result := make([]storage.StationWithMetadata, len(names))
	for i, name := range names {
		result[i] = storage.StationWithMetadata{
			Station: api.Station{StationUUID: name, Name: name},
		}
	}
	return result
}

// ---------------------------------------------------------------------------
// loadRecentlyPlayed
// ---------------------------------------------------------------------------

func TestLoadRecentlyPlayed_DisabledReturnsNil(t *testing.T) {
	app := newTestApp()
	app.playHistoryCfg.Enabled = false
	app.loadRecentlyPlayed()
	if app.recentlyPlayed != nil {
		t.Errorf("expected nil when disabled, got %d entries", len(app.recentlyPlayed))
	}
}

func TestLoadRecentlyPlayed_NilManagerReturnsNil(t *testing.T) {
	app := newTestApp()
	app.playHistoryCfg.Enabled = true
	app.metadataManager = nil
	app.loadRecentlyPlayed()
	if app.recentlyPlayed != nil {
		t.Errorf("expected nil when manager is nil, got %d entries", len(app.recentlyPlayed))
	}
}

func TestLoadRecentlyPlayed_EmptyHistory(t *testing.T) {
	tmpDir := t.TempDir()
	mgr, err := storage.NewMetadataManager(tmpDir)
	if err != nil {
		t.Fatalf("NewMetadataManager: %v", err)
	}
	defer mgr.Close() //nolint:errcheck

	app := newTestApp()
	app.playHistoryCfg.Enabled = true
	app.playHistoryCfg.Size = 5
	app.metadataManager = mgr

	app.loadRecentlyPlayed()
	if len(app.recentlyPlayed) != 0 {
		t.Errorf("expected 0 entries for empty history, got %d", len(app.recentlyPlayed))
	}
}

func TestLoadRecentlyPlayed_SizeLimit(t *testing.T) {
	tmpDir := t.TempDir()
	mgr, err := storage.NewMetadataManager(tmpDir)
	if err != nil {
		t.Fatalf("NewMetadataManager: %v", err)
	}
	defer mgr.Close() //nolint:errcheck

	// Record 10 distinct stations.
	for i := 0; i < 10; i++ {
		s := api.Station{StationUUID: fmt.Sprintf("station-%02d", i), Name: fmt.Sprintf("Station %d", i)}
		if err := mgr.StartPlay(&s); err != nil {
			t.Fatalf("StartPlay: %v", err)
		}
		if err := mgr.StopPlay(s.StationUUID); err != nil {
			t.Fatalf("StopPlay: %v", err)
		}
	}

	app := newTestApp()
	app.playHistoryCfg.Enabled = true
	app.playHistoryCfg.Size = 3
	app.metadataManager = mgr

	app.loadRecentlyPlayed()
	if len(app.recentlyPlayed) != 3 {
		t.Errorf("expected 3 entries (size limit), got %d", len(app.recentlyPlayed))
	}
}

// ---------------------------------------------------------------------------
// playRecentStation
// ---------------------------------------------------------------------------

func TestPlayRecentStation_ValidIndex(t *testing.T) {
	app := newTestApp()
	app.recentlyPlayed = makeRecentlyPlayed("Radio A", "Radio B", "Radio C")

	_, _ = app.playRecentStation(1)

	if app.playingStation == nil {
		t.Fatal("expected playingStation to be set")
	}
	if app.playingStation.Name != "Radio B" {
		t.Errorf("expected 'Radio B', got '%s'", app.playingStation.Name)
	}
	if !app.playingFromMain {
		t.Error("expected playingFromMain to be true")
	}
}

func TestPlayRecentStation_OutOfRange(t *testing.T) {
	app := newTestApp()
	app.recentlyPlayed = makeRecentlyPlayed("Radio A")

	_, _ = app.playRecentStation(5)

	if app.playingStation != nil {
		t.Error("expected playingStation to remain nil for out-of-range index")
	}
}

func TestPlayRecentStation_EmptyList(t *testing.T) {
	app := newTestApp()
	app.recentlyPlayed = nil

	_, _ = app.playRecentStation(0)

	if app.playingStation != nil {
		t.Error("expected playingStation to remain nil for empty list")
	}
}

func TestPlayRecentStation_FirstStation(t *testing.T) {
	app := newTestApp()
	app.recentlyPlayed = makeRecentlyPlayed("Jazz FM", "Blues Radio")

	_, _ = app.playRecentStation(0)

	if app.playingStation == nil {
		t.Fatal("expected playingStation to be set")
	}
	if app.playingStation.Name != "Jazz FM" {
		t.Errorf("expected 'Jazz FM', got '%s'", app.playingStation.Name)
	}
}

func TestPlayRecentStation_LastStation(t *testing.T) {
	app := newTestApp()
	app.recentlyPlayed = makeRecentlyPlayed("A", "B", "C")

	_, _ = app.playRecentStation(2)

	if app.playingStation == nil {
		t.Fatal("expected playingStation to be set")
	}
	if app.playingStation.Name != "C" {
		t.Errorf("expected 'C', got '%s'", app.playingStation.Name)
	}
}

// ---------------------------------------------------------------------------
// updateRPViewOffset
// ---------------------------------------------------------------------------

func TestUpdateRPViewOffset_NoRP(t *testing.T) {
	app := newTestApp()
	app.recentlyPlayed = nil
	app.rpViewOffset = 3 // stale value
	app.updateRPViewOffset(5)
	if app.rpViewOffset != 0 {
		t.Errorf("expected 0 when no RP entries, got %d", app.rpViewOffset)
	}
}

func TestUpdateRPViewOffset_CursorAboveRP(t *testing.T) {
	app := newTestApp()
	app.recentlyPlayed = makeRecentlyPlayed("A", "B", "C", "D", "E")
	app.rpViewOffset = 3
	app.unifiedMenuIndex = 0 // cursor in menu, not RP
	app.updateRPViewOffset(3)
	if app.rpViewOffset != 0 {
		t.Errorf("expected offset reset to 0 when cursor above RP, got %d", app.rpViewOffset)
	}
}

func TestUpdateRPViewOffset_ScrollsDown(t *testing.T) {
	app := newTestApp()
	app.recentlyPlayed = makeRecentlyPlayed("A", "B", "C", "D", "E")
	// Cursor at RP entry 4 (index 3 in RP, unifiedIdx = 11+3 = 14)
	rpStart := mainMenuItemCount + len(app.quickFavorites)
	app.unifiedMenuIndex = rpStart + 3
	app.updateRPViewOffset(3) // window of 3
	// entry 3 must be visible: offset should be 1 (shows entries 1,2,3)
	if app.rpViewOffset != 1 {
		t.Errorf("expected offset 1, got %d", app.rpViewOffset)
	}
}

func TestUpdateRPViewOffset_ScrollsUp(t *testing.T) {
	app := newTestApp()
	app.recentlyPlayed = makeRecentlyPlayed("A", "B", "C", "D", "E")
	app.rpViewOffset = 4 // pushed far down
	rpStart := mainMenuItemCount + len(app.quickFavorites)
	app.unifiedMenuIndex = rpStart + 0 // cursor at first RP entry
	app.updateRPViewOffset(3)
	if app.rpViewOffset != 0 {
		t.Errorf("expected offset scrolled back to 0, got %d", app.rpViewOffset)
	}
}

func TestUpdateRPViewOffset_LargeWindowCoversAllEntries(t *testing.T) {
	app := newTestApp()
	app.recentlyPlayed = makeRecentlyPlayed("A", "B", "C")
	// Window larger than list — all items fit, offset stays 0 regardless of cursor position
	rpStart := mainMenuItemCount + len(app.quickFavorites)
	app.unifiedMenuIndex = rpStart + 1
	app.updateRPViewOffset(10)
	if app.rpViewOffset != 0 {
		t.Errorf("expected offset 0 when window covers all entries, got %d", app.rpViewOffset)
	}
}
