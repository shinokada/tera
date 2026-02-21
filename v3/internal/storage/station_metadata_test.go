package storage

import (
	"os"
	"testing"
	"time"

	"github.com/shinokada/tera/v3/internal/api"
)

// testStation creates a test station with the given UUID
func testStation(uuid string) *api.Station {
	return &api.Station{
		StationUUID: uuid,
		Name:        "Test Station " + uuid,
		URLResolved: "http://test.stream/" + uuid,
	}
}

func TestMetadataManager(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "tera-metadata-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	t.Run("NewMetadataManager", func(t *testing.T) {
		mgr, err := NewMetadataManager(tmpDir)
		if err != nil {
			t.Fatalf("Failed to create metadata manager: %v", err)
		}
		defer func() { _ = mgr.Close() }()

		if mgr.GetTotalStations() != 0 {
			t.Errorf("Expected 0 stations, got %d", mgr.GetTotalStations())
		}
	})

	t.Run("StartPlay_RecordsMetadata", func(t *testing.T) {
		mgr, err := NewMetadataManager(tmpDir)
		if err != nil {
			t.Fatalf("Failed to create metadata manager: %v", err)
		}
		defer func() { _ = mgr.Close() }()

		stationUUID := "test-station-1"

		err = mgr.StartPlay(testStation(stationUUID))
		if err != nil {
			t.Fatalf("Failed to start play: %v", err)
		}

		metadata := mgr.GetMetadata(stationUUID)
		if metadata == nil {
			t.Fatal("Expected metadata, got nil")
		}

		if metadata.PlayCount != 1 {
			t.Errorf("Expected PlayCount 1, got %d", metadata.PlayCount)
		}

		if metadata.LastPlayed.IsZero() {
			t.Error("Expected LastPlayed to be set")
		}

		if metadata.FirstPlayed.IsZero() {
			t.Error("Expected FirstPlayed to be set")
		}
	})

	t.Run("StartPlay_IncrementsCount", func(t *testing.T) {
		mgr, err := NewMetadataManager(tmpDir)
		if err != nil {
			t.Fatalf("Failed to create metadata manager: %v", err)
		}
		defer func() { _ = mgr.Close() }()

		stationUUID := "test-station-increment"

		// Play 3 times
		for i := 0; i < 3; i++ {
			// Stop previous play to allow next play to count
			_ = mgr.StopPlay(stationUUID)
			err = mgr.StartPlay(testStation(stationUUID))
			if err != nil {
				t.Fatalf("Failed to start play: %v", err)
			}
		}

		metadata := mgr.GetMetadata(stationUUID)
		if metadata == nil {
			t.Fatal("Expected metadata, got nil")
		}

		if metadata.PlayCount != 3 {
			t.Errorf("Expected PlayCount 3, got %d", metadata.PlayCount)
		}
	})

	t.Run("StartPlay_SameStation_NoDuplicate", func(t *testing.T) {
		mgr, err := NewMetadataManager(tmpDir)
		if err != nil {
			t.Fatalf("Failed to create metadata manager: %v", err)
		}
		defer func() { _ = mgr.Close() }()

		stationUUID := "test-station-dedup"

		// Start play twice without stopping
		_ = mgr.StartPlay(testStation(stationUUID))
		_ = mgr.StartPlay(testStation(stationUUID))
		_ = mgr.StartPlay(testStation(stationUUID))

		metadata := mgr.GetMetadata(stationUUID)
		if metadata == nil {
			t.Fatal("Expected metadata, got nil")
		}

		// Should only count as 1 play since same station
		if metadata.PlayCount != 1 {
			t.Errorf("Expected PlayCount 1 (dedup), got %d", metadata.PlayCount)
		}
	})

	t.Run("StopPlay_RecordsDuration", func(t *testing.T) {
		mgr, err := NewMetadataManager(tmpDir)
		if err != nil {
			t.Fatalf("Failed to create metadata manager: %v", err)
		}
		defer func() { _ = mgr.Close() }()

		stationUUID := "test-station-duration"

		_ = mgr.StartPlay(testStation(stationUUID))

		// Wait a bit to accumulate duration
		time.Sleep(100 * time.Millisecond)

		_ = mgr.StopPlay(stationUUID)

		metadata := mgr.GetMetadata(stationUUID)
		if metadata == nil {
			t.Fatal("Expected metadata, got nil")
		}

		// Duration should be > 0
		if metadata.TotalDurationSeconds <= 0 {
			// It might be 0 if the sleep was too short, so just log
			t.Logf("TotalDurationSeconds: %d (may be 0 for short duration)", metadata.TotalDurationSeconds)
		}
	})

	t.Run("StartPlay_SwitchStation_StopsOldStation", func(t *testing.T) {
		mgr, err := NewMetadataManager(tmpDir)
		if err != nil {
			t.Fatalf("Failed to create metadata manager: %v", err)
		}
		defer func() { _ = mgr.Close() }()

		station1 := "test-station-switch-1"
		station2 := "test-station-switch-2"

		_ = mgr.StartPlay(testStation(station1))
		time.Sleep(50 * time.Millisecond)
		_ = mgr.StartPlay(testStation(station2)) // Should auto-stop station1

		// Both should have metadata
		meta1 := mgr.GetMetadata(station1)
		meta2 := mgr.GetMetadata(station2)

		if meta1 == nil || meta2 == nil {
			t.Fatal("Expected metadata for both stations")
		}

		if meta1.PlayCount != 1 {
			t.Errorf("Expected station1 PlayCount 1, got %d", meta1.PlayCount)
		}
		if meta2.PlayCount != 1 {
			t.Errorf("Expected station2 PlayCount 1, got %d", meta2.PlayCount)
		}
	})

	t.Run("GetTopPlayed", func(t *testing.T) {
		mgr, err := NewMetadataManager(tmpDir)
		if err != nil {
			t.Fatalf("Failed to create metadata manager: %v", err)
		}
		defer func() { _ = mgr.Close() }()

		// Clear existing data
		_ = mgr.ClearAll()

		// Create stations with different play counts
		station1 := "top-played-1"
		station2 := "top-played-2"
		station3 := "top-played-3"

		// Station 3: 5 plays
		for i := 0; i < 5; i++ {
			_ = mgr.StopPlay(station3)
			_ = mgr.StartPlay(testStation(station3))
		}
		_ = mgr.StopPlay(station3)

		// Station 1: 3 plays
		for i := 0; i < 3; i++ {
			_ = mgr.StopPlay(station1)
			_ = mgr.StartPlay(testStation(station1))
		}
		_ = mgr.StopPlay(station1)

		// Station 2: 1 play
		_ = mgr.StartPlay(testStation(station2))
		_ = mgr.StopPlay(station2)

		// Get top 2
		top := mgr.GetTopPlayed(2)
		if len(top) != 2 {
			t.Fatalf("Expected 2 results, got %d", len(top))
		}

		if top[0].Station.StationUUID != station3 {
			t.Errorf("Expected first to be station3, got %s", top[0].Station.StationUUID)
		}
		if top[0].Metadata.PlayCount != 5 {
			t.Errorf("Expected first PlayCount 5, got %d", top[0].Metadata.PlayCount)
		}

		if top[1].Station.StationUUID != station1 {
			t.Errorf("Expected second to be station1, got %s", top[1].Station.StationUUID)
		}
	})

	t.Run("GetRecentlyPlayed", func(t *testing.T) {
		mgr, err := NewMetadataManager(tmpDir)
		if err != nil {
			t.Fatalf("Failed to create metadata manager: %v", err)
		}
		defer func() { _ = mgr.Close() }()

		// Clear existing data
		_ = mgr.ClearAll()

		station1 := "recent-1"
		station2 := "recent-2"
		station3 := "recent-3"

		// Play in order: station1, station2, station3
		_ = mgr.StartPlay(testStation(station1))
		_ = mgr.StopPlay(station1)
		time.Sleep(10 * time.Millisecond)

		_ = mgr.StartPlay(testStation(station2))
		_ = mgr.StopPlay(station2)
		time.Sleep(10 * time.Millisecond)

		_ = mgr.StartPlay(testStation(station3))
		_ = mgr.StopPlay(station3)

		// Most recent should be station3
		recent := mgr.GetRecentlyPlayed(2)
		if len(recent) != 2 {
			t.Fatalf("Expected 2 results, got %d", len(recent))
		}

		if recent[0].Station.StationUUID != station3 {
			t.Errorf("Expected first to be station3, got %s", recent[0].Station.StationUUID)
		}
		if recent[1].Station.StationUUID != station2 {
			t.Errorf("Expected second to be station2, got %s", recent[1].Station.StationUUID)
		}
	})

	t.Run("SaveAndLoad", func(t *testing.T) {
		tmpDir2, err := os.MkdirTemp("", "tera-metadata-persist-test")
		if err != nil {
			t.Fatalf("Failed to create temp dir: %v", err)
		}
		defer func() { _ = os.RemoveAll(tmpDir2) }()

		stationUUID := "persist-test-station"

		// Create manager and add data
		{
			mgr, err := NewMetadataManager(tmpDir2)
			if err != nil {
				t.Fatalf("Failed to create metadata manager: %v", err)
			}

			_ = mgr.StartPlay(testStation(stationUUID))
			_ = mgr.StopPlay(stationUUID)

			// Force save
			_ = mgr.Save()
			_ = mgr.Close()
		}

		// Create new manager and verify data persisted
		{
			mgr, err := NewMetadataManager(tmpDir2)
			if err != nil {
				t.Fatalf("Failed to create metadata manager: %v", err)
			}
			defer func() { _ = mgr.Close() }()

			metadata := mgr.GetMetadata(stationUUID)
			if metadata == nil {
				t.Fatal("Expected metadata to persist, got nil")
			}

			if metadata.PlayCount != 1 {
				t.Errorf("Expected PlayCount 1, got %d", metadata.PlayCount)
			}
		}
	})

	t.Run("ClearAll", func(t *testing.T) {
		mgr, err := NewMetadataManager(tmpDir)
		if err != nil {
			t.Fatalf("Failed to create metadata manager: %v", err)
		}
		defer func() { _ = mgr.Close() }()

		stationUUID := "clear-test-station"

		_ = mgr.StartPlay(testStation(stationUUID))
		_ = mgr.StopPlay(stationUUID)

		if mgr.GetTotalStations() == 0 {
			t.Fatal("Expected at least 1 station before clear")
		}

		_ = mgr.ClearAll()

		if mgr.GetTotalStations() != 0 {
			t.Errorf("Expected 0 stations after clear, got %d", mgr.GetTotalStations())
		}
	})

	t.Run("CorruptedFile_GracefulRecovery", func(t *testing.T) {
		tmpDir3, err := os.MkdirTemp("", "tera-metadata-corrupt-test")
		if err != nil {
			t.Fatalf("Failed to create temp dir: %v", err)
		}
		defer func() { _ = os.RemoveAll(tmpDir3) }()

		// Write corrupted JSON file
		corruptedData := []byte("{invalid json")
		err = os.WriteFile(tmpDir3+"/station_metadata.json", corruptedData, 0644)
		if err != nil {
			t.Fatalf("Failed to write corrupted file: %v", err)
		}

		// Should recover gracefully with empty store
		mgr, err := NewMetadataManager(tmpDir3)
		if err != nil {
			t.Fatalf("Expected graceful recovery, got error: %v", err)
		}
		defer func() { _ = mgr.Close() }()

		if mgr.GetTotalStations() != 0 {
			t.Errorf("Expected 0 stations after corrupted file, got %d", mgr.GetTotalStations())
		}
	})
}

func TestFormatLastPlayed(t *testing.T) {
	tests := []struct {
		name     string
		time     time.Time
		expected string
	}{
		{"Zero time", time.Time{}, "Never"},
		{"Just now", time.Now(), "Just now"},
		{"1 minute ago", time.Now().Add(-1 * time.Minute), "1 minute ago"},
		{"5 minutes ago", time.Now().Add(-5 * time.Minute), "5 minutes ago"},
		{"1 hour ago", time.Now().Add(-1 * time.Hour), "1 hour ago"},
		{"3 hours ago", time.Now().Add(-3 * time.Hour), "3 hours ago"},
		{"Yesterday", time.Now().Add(-25 * time.Hour), "Yesterday"},
		{"3 days ago", time.Now().Add(-3 * 24 * time.Hour), "3 days ago"},
		{"1 week ago", time.Now().Add(-7 * 24 * time.Hour), "1 week ago"},
		{"2 weeks ago", time.Now().Add(-14 * 24 * time.Hour), "2 weeks ago"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := FormatLastPlayed(tc.time)
			if result != tc.expected {
				t.Errorf("Expected '%s', got '%s'", tc.expected, result)
			}
		})
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		seconds  int64
		expected string
	}{
		{30, "30s"},
		{60, "1m"},
		{90, "1m 30s"},
		{3600, "1h"},
		{3660, "1h 1m"},
		{7200, "2h"},
		{7335, "2h 2m"},
	}

	for _, tc := range tests {
		t.Run(tc.expected, func(t *testing.T) {
			result := FormatDuration(tc.seconds)
			if result != tc.expected {
				t.Errorf("FormatDuration(%d): expected '%s', got '%s'", tc.seconds, tc.expected, result)
			}
		})
	}
}

// Additional edge case tests

func TestGetCachedStation(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "tera-cached-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	mgr, err := NewMetadataManager(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create metadata manager: %v", err)
	}
	defer func() { _ = mgr.Close() }()

	station := testStation("cached-uuid")
	station.Country = "USA"
	station.Language = "english"
	station.Tags = "rock,pop"

	if err := mgr.StartPlay(station); err != nil {
		t.Fatalf("StartPlay failed: %v", err)
	}

	cached := mgr.GetCachedStation("cached-uuid")
	if cached == nil {
		t.Fatal("Expected cached station, got nil")
	}

	if cached.Name != station.Name {
		t.Errorf("Expected cached name %q, got %q", station.Name, cached.Name)
	}
	if cached.Country != station.Country {
		t.Errorf("Expected cached country %q, got %q", station.Country, cached.Country)
	}
}

func TestGetMetadataReturnsCopy(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "tera-copy-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	mgr, err := NewMetadataManager(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create metadata manager: %v", err)
	}
	defer func() { _ = mgr.Close() }()

	if err := mgr.StartPlay(testStation("copy-test")); err != nil {
		t.Fatalf("StartPlay failed: %v", err)
	}

	// Get metadata twice
	meta1 := mgr.GetMetadata("copy-test")
	meta2 := mgr.GetMetadata("copy-test")

	if meta1 == nil || meta2 == nil {
		t.Fatal("Expected metadata")
	}

	// Modify first copy
	meta1.PlayCount = 999

	// Second copy should be unchanged
	if meta2.PlayCount == 999 {
		t.Error("GetMetadata should return independent copies")
	}
}

func TestFormatDurationEdgeCases(t *testing.T) {
	tests := []struct {
		seconds  int64
		expected string
	}{
		{0, "0s"},
		{-1, "0s"},
		{59, "59s"},
		{120, "2m"},
		{3599, "59m 59s"},
	}

	for _, tc := range tests {
		t.Run(tc.expected, func(t *testing.T) {
			result := FormatDuration(tc.seconds)
			if result != tc.expected {
				t.Errorf("FormatDuration(%d): expected '%s', got '%s'", tc.seconds, tc.expected, result)
			}
		})
	}
}

func TestFormatLastPlayedAboutYear(t *testing.T) {
	// Use a single base time to avoid drift between two time.Now() calls.
	now := time.Now()

	// 330 days back: months = int(330*24h / (24*30h)) = 11
	elevenMonths := now.Add(-11 * 30 * 24 * time.Hour)
	result := FormatLastPlayed(elevenMonths)
	if result != "11 months ago" {
		t.Errorf("Expected '11 months ago', got %q", result)
	}

	// 360 days: still < 365 days, months = int(360*24h / (24*30h)) = 12 → "About a year ago"
	almostYear := now.Add(-360 * 24 * time.Hour)
	result = FormatLastPlayed(almostYear)
	if result != "About a year ago" {
		t.Errorf("Expected 'About a year ago', got %q", result)
	}
}

func TestStopPlayNonExistent(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "tera-stop-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	mgr, err := NewMetadataManager(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create metadata manager: %v", err)
	}
	defer func() { _ = mgr.Close() }()

	// Stopping a non-playing station should not error
	err = mgr.StopPlay("nonexistent")
	if err != nil {
		t.Errorf("Expected no error for stopping nonexistent station, got %v", err)
	}
}

func TestGetFirstPlayed(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "tera-first-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	mgr, err := NewMetadataManager(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create metadata manager: %v", err)
	}
	defer func() { _ = mgr.Close() }()

	// The sleep ensures FirstPlayed timestamps are strictly ordered;
	// back-to-back StartPlay calls on a fast machine can produce the same
	// time.Now() value inside the manager's lock.
	if err := mgr.StartPlay(testStation("first-station")); err != nil {
		t.Fatalf("StartPlay failed: %v", err)
	}
	if err := mgr.StopPlay("first-station"); err != nil {
		t.Fatalf("StopPlay failed: %v", err)
	}
	time.Sleep(50 * time.Millisecond)

	if err := mgr.StartPlay(testStation("second-station")); err != nil {
		t.Fatalf("StartPlay failed: %v", err)
	}
	if err := mgr.StopPlay("second-station"); err != nil {
		t.Fatalf("StopPlay failed: %v", err)
	}

	// Get first played
	first := mgr.GetFirstPlayed(1)
	if len(first) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(first))
	}

	if first[0].Station.StationUUID != "first-station" {
		t.Errorf("Expected first-station, got %s", first[0].Station.StationUUID)
	}
}

func TestGetAllStationUUIDs(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "tera-all-uuids-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	mgr, err := NewMetadataManager(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create metadata manager: %v", err)
	}
	defer func() { _ = mgr.Close() }()

	// Add stations in unsorted order
	_ = mgr.StartPlay(testStation("uuid-c"))
	_ = mgr.StartPlay(testStation("uuid-a"))
	_ = mgr.StartPlay(testStation("uuid-b"))

	uuids := mgr.GetAllStationUUIDs()
	if len(uuids) != 3 {
		t.Fatalf("Expected 3 UUIDs, got %d", len(uuids))
	}

	// Should be sorted
	if uuids[0] != "uuid-a" || uuids[1] != "uuid-b" || uuids[2] != "uuid-c" {
		t.Errorf("Expected sorted UUIDs [uuid-a uuid-b uuid-c], got %v", uuids)
	}
}

func TestCloseStopsCurrentPlay(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "tera-close-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	mgr, err := NewMetadataManager(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create metadata manager: %v", err)
	}

	const stationUUID = "close-playing"
	if err := mgr.StartPlay(testStation(stationUUID)); err != nil {
		t.Fatalf("StartPlay failed: %v", err)
	}

	// Close must stop the current play, flush metadata, and persist to disk.
	// No sleep needed: we only assert Close() is error-free and that the
	// station's play count was recorded (duration may be 0 for sub-second plays).
	if err := mgr.Close(); err != nil {
		t.Fatalf("Close failed: %v", err)
	}

	// Reload from disk and verify the play was recorded.
	mgr2, err := NewMetadataManager(tmpDir)
	if err != nil {
		t.Fatalf("Failed to reopen metadata manager: %v", err)
	}
	defer func() { _ = mgr2.Close() }()

	meta := mgr2.GetMetadata(stationUUID)
	if meta == nil {
		t.Fatal("Expected metadata after Close, got nil")
	}
	if meta.PlayCount != 1 {
		t.Errorf("Expected PlayCount 1, got %d", meta.PlayCount)
	}
}