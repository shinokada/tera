package storage

import (
	"os"
	"testing"
	"time"

	"github.com/shinokada/tera/v3/internal/api"
)

// testRatingStation creates a test station with the given UUID
func testRatingStation(uuid string) *api.Station {
	return &api.Station{
		StationUUID: uuid,
		Name:        "Test Station " + uuid,
		URLResolved: "http://test.stream/" + uuid,
	}
}

func TestRatingsManager(t *testing.T) {
	t.Run("NewRatingsManager", func(t *testing.T) {
		tmpDir := t.TempDir()
		mgr, err := NewRatingsManager(tmpDir)
		if err != nil {
			t.Fatalf("Failed to create ratings manager: %v", err)
		}
		defer func() { _ = mgr.Close() }()

		if mgr.GetTotalRated() != 0 {
			t.Errorf("Expected 0 rated stations, got %d", mgr.GetTotalRated())
		}
	})

	t.Run("SetRating_Valid", func(t *testing.T) {
		tmpDir := t.TempDir()
		mgr, err := NewRatingsManager(tmpDir)
		if err != nil {
			t.Fatalf("Failed to create ratings manager: %v", err)
		}
		defer func() { _ = mgr.Close() }()

		stationUUID := "test-station-1"

		err = mgr.SetRating(testRatingStation(stationUUID), 5)
		if err != nil {
			t.Fatalf("Failed to set rating: %v", err)
		}

		rating := mgr.GetRating(stationUUID)
		if rating == nil {
			t.Fatal("Expected rating, got nil")
		}

		if rating.Rating != 5 {
			t.Errorf("Expected rating 5, got %d", rating.Rating)
		}

		if rating.RatedAt.IsZero() {
			t.Error("Expected RatedAt to be set")
		}

		if rating.UpdatedAt.IsZero() {
			t.Error("Expected UpdatedAt to be set")
		}
	})

	t.Run("SetRating_Validation", func(t *testing.T) {
		tmpDir := t.TempDir()
		mgr, err := NewRatingsManager(tmpDir)
		if err != nil {
			t.Fatalf("Failed to create ratings manager: %v", err)
		}
		defer func() { _ = mgr.Close() }()

		stationUUID := "test-station-validation"

		// Test invalid ratings
		invalidRatings := []int{0, -1, 6, 100, -100}
		for _, r := range invalidRatings {
			err = mgr.SetRating(testRatingStation(stationUUID), r)
			if err == nil {
				t.Errorf("Expected error for invalid rating %d, got nil", r)
			}
		}

		// Test valid ratings
		validRatings := []int{1, 2, 3, 4, 5}
		for _, r := range validRatings {
			err = mgr.SetRating(testRatingStation(stationUUID), r)
			if err != nil {
				t.Errorf("Expected no error for valid rating %d, got %v", r, err)
			}
		}
	})

	t.Run("SetRating_UpdateExisting", func(t *testing.T) {
		tmpDir := t.TempDir()
		mgr, err := NewRatingsManager(tmpDir)
		if err != nil {
			t.Fatalf("Failed to create ratings manager: %v", err)
		}
		defer func() { _ = mgr.Close() }()

		stationUUID := "test-station-update"

		// Set initial rating
		err = mgr.SetRating(testRatingStation(stationUUID), 3)
		if err != nil {
			t.Fatalf("Failed to set initial rating: %v", err)
		}

		initialRating := mgr.GetRating(stationUUID)
		initialRatedAt := initialRating.RatedAt

		// Wait a bit to ensure timestamp difference
		time.Sleep(10 * time.Millisecond)

		// Update rating
		err = mgr.SetRating(testRatingStation(stationUUID), 5)
		if err != nil {
			t.Fatalf("Failed to update rating: %v", err)
		}

		updatedRating := mgr.GetRating(stationUUID)
		if updatedRating == nil {
			t.Fatal("Expected rating, got nil")
		}

		if updatedRating.Rating != 5 {
			t.Errorf("Expected rating 5, got %d", updatedRating.Rating)
		}

		// RatedAt should stay the same
		if !updatedRating.RatedAt.Equal(initialRatedAt) {
			t.Error("RatedAt should not change on update")
		}

		// UpdatedAt should be different
		if updatedRating.UpdatedAt.Equal(initialRatedAt) {
			t.Error("UpdatedAt should change on update")
		}
	})

	t.Run("RemoveRating", func(t *testing.T) {
		tmpDir := t.TempDir()
		mgr, err := NewRatingsManager(tmpDir)
		if err != nil {
			t.Fatalf("Failed to create ratings manager: %v", err)
		}
		defer func() { _ = mgr.Close() }()

		stationUUID := "test-station-remove"

		// Set rating
		err = mgr.SetRating(testRatingStation(stationUUID), 4)
		if err != nil {
			t.Fatalf("Failed to set rating: %v", err)
		}

		// Verify it exists
		if mgr.GetRating(stationUUID) == nil {
			t.Fatal("Expected rating to exist")
		}

		// Remove rating
		err = mgr.RemoveRating(stationUUID)
		if err != nil {
			t.Fatalf("Failed to remove rating: %v", err)
		}

		// Verify it's gone
		if mgr.GetRating(stationUUID) != nil {
			t.Error("Expected rating to be removed")
		}
	})

	t.Run("RemoveRating_NonExistent", func(t *testing.T) {
		tmpDir := t.TempDir()
		mgr, err := NewRatingsManager(tmpDir)
		if err != nil {
			t.Fatalf("Failed to create ratings manager: %v", err)
		}
		defer func() { _ = mgr.Close() }()

		// Remove non-existent rating should not error
		err = mgr.RemoveRating("non-existent-station")
		if err != nil {
			t.Errorf("Expected no error for removing non-existent rating, got %v", err)
		}
	})

	t.Run("GetRating_NotFound", func(t *testing.T) {
		tmpDir := t.TempDir()
		mgr, err := NewRatingsManager(tmpDir)
		if err != nil {
			t.Fatalf("Failed to create ratings manager: %v", err)
		}
		defer func() { _ = mgr.Close() }()

		rating := mgr.GetRating("non-existent-station")
		if rating != nil {
			t.Errorf("Expected nil for non-existent station, got %+v", rating)
		}
	})

	t.Run("GetTopRated", func(t *testing.T) {
		tmpDir := t.TempDir()
		mgr, err := NewRatingsManager(tmpDir)
		if err != nil {
			t.Fatalf("Failed to create ratings manager: %v", err)
		}
		defer func() { _ = mgr.Close() }()

		// Add some ratings
		_ = mgr.SetRating(testRatingStation("station-a"), 3)
		_ = mgr.SetRating(testRatingStation("station-b"), 5)
		_ = mgr.SetRating(testRatingStation("station-c"), 4)
		_ = mgr.SetRating(testRatingStation("station-d"), 5)
		_ = mgr.SetRating(testRatingStation("station-e"), 1)

		// Get top 3
		top := mgr.GetTopRated(3)
		if len(top) != 3 {
			t.Fatalf("Expected 3 results, got %d", len(top))
		}

		// First two should be 5 stars (order by UUID for tie)
		if top[0].Rating.Rating != 5 {
			t.Errorf("Expected first to be 5 stars, got %d", top[0].Rating.Rating)
		}
		if top[1].Rating.Rating != 5 {
			t.Errorf("Expected second to be 5 stars, got %d", top[1].Rating.Rating)
		}
		if top[2].Rating.Rating != 4 {
			t.Errorf("Expected third to be 4 stars, got %d", top[2].Rating.Rating)
		}
	})

	t.Run("GetByMinRating", func(t *testing.T) {
		tmpDir := t.TempDir()
		mgr, err := NewRatingsManager(tmpDir)
		if err != nil {
			t.Fatalf("Failed to create ratings manager: %v", err)
		}
		defer func() { _ = mgr.Close() }()

		// Add some ratings
		_ = mgr.SetRating(testRatingStation("station-a"), 3)
		_ = mgr.SetRating(testRatingStation("station-b"), 5)
		_ = mgr.SetRating(testRatingStation("station-c"), 4)
		_ = mgr.SetRating(testRatingStation("station-d"), 2)
		_ = mgr.SetRating(testRatingStation("station-e"), 1)

		// Get 4+ stars
		highRated := mgr.GetByMinRating(4)
		if len(highRated) != 2 {
			t.Fatalf("Expected 2 results for 4+ stars, got %d", len(highRated))
		}

		// Get 3+ stars
		midRated := mgr.GetByMinRating(3)
		if len(midRated) != 3 {
			t.Fatalf("Expected 3 results for 3+ stars, got %d", len(midRated))
		}

		// Get 5 stars only
		fiveStars := mgr.GetByMinRating(5)
		if len(fiveStars) != 1 {
			t.Fatalf("Expected 1 result for 5 stars only, got %d", len(fiveStars))
		}
	})

	t.Run("GetRecentlyRated", func(t *testing.T) {
		tmpDir := t.TempDir()
		mgr, err := NewRatingsManager(tmpDir)
		if err != nil {
			t.Fatalf("Failed to create ratings manager: %v", err)
		}
		defer func() { _ = mgr.Close() }()

		// Add ratings with delays
		_ = mgr.SetRating(testRatingStation("station-a"), 3)
		time.Sleep(10 * time.Millisecond)
		_ = mgr.SetRating(testRatingStation("station-b"), 4)
		time.Sleep(10 * time.Millisecond)
		_ = mgr.SetRating(testRatingStation("station-c"), 5)

		// Get recent
		recent := mgr.GetRecentlyRated(2)
		if len(recent) != 2 {
			t.Fatalf("Expected 2 results, got %d", len(recent))
		}

		// Most recent should be station-c
		if recent[0].Station.StationUUID != "station-c" {
			t.Errorf("Expected first to be station-c, got %s", recent[0].Station.StationUUID)
		}
	})

	t.Run("GetAllRated", func(t *testing.T) {
		tmpDir := t.TempDir()
		mgr, err := NewRatingsManager(tmpDir)
		if err != nil {
			t.Fatalf("Failed to create ratings manager: %v", err)
		}
		defer func() { _ = mgr.Close() }()

		// Add some ratings
		_ = mgr.SetRating(testRatingStation("station-a"), 3)
		_ = mgr.SetRating(testRatingStation("station-b"), 5)
		_ = mgr.SetRating(testRatingStation("station-c"), 4)

		all := mgr.GetAllRated()
		if len(all) != 3 {
			t.Fatalf("Expected 3 results, got %d", len(all))
		}
	})

	t.Run("ClearAll", func(t *testing.T) {
		tmpDir := t.TempDir()
		mgr, err := NewRatingsManager(tmpDir)
		if err != nil {
			t.Fatalf("Failed to create ratings manager: %v", err)
		}
		defer func() { _ = mgr.Close() }()

		// Add some ratings
		_ = mgr.SetRating(testRatingStation("station-a"), 3)
		_ = mgr.SetRating(testRatingStation("station-b"), 5)

		if mgr.GetTotalRated() != 2 {
			t.Errorf("Expected 2 rated stations, got %d", mgr.GetTotalRated())
		}

		// Clear all
		err = mgr.ClearAll()
		if err != nil {
			t.Fatalf("Failed to clear all: %v", err)
		}

		if mgr.GetTotalRated() != 0 {
			t.Errorf("Expected 0 rated stations after clear, got %d", mgr.GetTotalRated())
		}
	})

	t.Run("SaveAndLoad", func(t *testing.T) {
		tmpDir := t.TempDir()
		// Create first manager and add ratings
		mgr1, err := NewRatingsManager(tmpDir)
		if err != nil {
			t.Fatalf("Failed to create ratings manager: %v", err)
		}

		_ = mgr1.SetRating(testRatingStation("station-persist-1"), 5)
		_ = mgr1.SetRating(testRatingStation("station-persist-2"), 3)

		// Force save
		err = mgr1.Save()
		if err != nil {
			t.Fatalf("Failed to save: %v", err)
		}
		_ = mgr1.Close()

		// Create second manager and verify data persisted
		mgr2, err := NewRatingsManager(tmpDir)
		if err != nil {
			t.Fatalf("Failed to create second ratings manager: %v", err)
		}
		defer func() { _ = mgr2.Close() }()

		rating1 := mgr2.GetRating("station-persist-1")
		if rating1 == nil || rating1.Rating != 5 {
			t.Errorf("Expected persisted rating of 5 for station-persist-1")
		}

		rating2 := mgr2.GetRating("station-persist-2")
		if rating2 == nil || rating2.Rating != 3 {
			t.Errorf("Expected persisted rating of 3 for station-persist-2")
		}
	})

	t.Run("CorruptedFile", func(t *testing.T) {
		// Create a corrupted ratings file
		corruptDir := t.TempDir()

		// Write invalid JSON
		err := os.WriteFile(corruptDir+"/station_ratings.json", []byte("not valid json"), 0644)
		if err != nil {
			t.Fatalf("Failed to write corrupt file: %v", err)
		}

		// Manager should still be created (graceful degradation)
		mgr, err := NewRatingsManager(corruptDir)
		if err != nil {
			t.Fatalf("Expected manager to be created despite corrupt file, got error: %v", err)
		}
		defer func() { _ = mgr.Close() }()

		// Should have empty store
		if mgr.GetTotalRated() != 0 {
			t.Errorf("Expected 0 rated stations on corrupt file, got %d", mgr.GetTotalRated())
		}
	})
}

func TestRenderStars(t *testing.T) {
	tests := []struct {
		rating     int
		useUnicode bool
		expected   string
	}{
		{0, true, "☆ ☆ ☆ ☆ ☆"},
		{1, true, "★ ☆ ☆ ☆ ☆"},
		{2, true, "★ ★ ☆ ☆ ☆"},
		{3, true, "★ ★ ★ ☆ ☆"},
		{4, true, "★ ★ ★ ★ ☆"},
		{5, true, "★ ★ ★ ★ ★"},
		{-1, true, "☆ ☆ ☆ ☆ ☆"},
		{6, true, "★ ★ ★ ★ ★"},
		{0, false, "- - - - -"},
		{3, false, "* * * - -"},
		{5, false, "* * * * *"},
	}

	for _, tt := range tests {
		result := RenderStars(tt.rating, tt.useUnicode)
		if result != tt.expected {
			t.Errorf("RenderStars(%d, %v) = %q, want %q", tt.rating, tt.useUnicode, result, tt.expected)
		}
	}
}

func TestRenderStarsCompact(t *testing.T) {
	tests := []struct {
		rating     int
		useUnicode bool
		expected   string
	}{
		{0, true, ""},
		{1, true, "★"},
		{2, true, "★ ★"},
		{3, true, "★ ★ ★"},
		{4, true, "★ ★ ★ ★"},
		{5, true, "★ ★ ★ ★ ★"},
		{-1, true, ""},
		{6, true, ""},
		{0, false, ""},
		{3, false, "* * *"},
		{5, false, "* * * * *"},
	}

	for _, tt := range tests {
		result := RenderStarsCompact(tt.rating, tt.useUnicode)
		if result != tt.expected {
			t.Errorf("RenderStarsCompact(%d, %v) = %q, want %q", tt.rating, tt.useUnicode, result, tt.expected)
		}
	}
}

func TestRatingsConcurrentAccess(t *testing.T) {
	tmpDir := t.TempDir()

	mgr, err := NewRatingsManager(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create ratings manager: %v", err)
	}
	defer func() { _ = mgr.Close() }()

	// Run concurrent operations
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(n int) {
			for j := 0; j < 100; j++ {
				stationUUID := "station-concurrent"
				rating := (j % 5) + 1
				_ = mgr.SetRating(testRatingStation(stationUUID), rating)
				_ = mgr.GetRating(stationUUID)
				_ = mgr.GetTopRated(10)
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify no panic and data is consistent
	total := mgr.GetTotalRated()
	if total < 1 {
		t.Errorf("Expected at least 1 rated station, got %d", total)
	}
}
