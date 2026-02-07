package blocklist

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/shinokada/tera/internal/api"
)

func TestNewManager(t *testing.T) {
	manager := NewManager("/tmp/test-blocklist.json")
	if manager == nil {
		t.Fatal("NewManager returned nil")
	}
	if manager.blocklistPath != "/tmp/test-blocklist.json" {
		t.Errorf("Expected path /tmp/test-blocklist.json, got %s", manager.blocklistPath)
	}
}

func TestLoadEmptyBlocklist(t *testing.T) {
	// Create temp directory
	tmpDir := t.TempDir()
	blocklistPath := filepath.Join(tmpDir, "blocklist.json")

	manager := NewManager(blocklistPath)
	ctx := context.Background()

	// Load should succeed even if file doesn't exist
	if err := manager.Load(ctx); err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if manager.Count() != 0 {
		t.Errorf("Expected empty blocklist, got %d stations", manager.Count())
	}
}

func TestBlockStation(t *testing.T) {
	tmpDir := t.TempDir()
	blocklistPath := filepath.Join(tmpDir, "blocklist.json")

	manager := NewManager(blocklistPath)
	ctx := context.Background()

	station := &api.Station{
		StationUUID: "test-uuid-123",
		Name:        "Test Radio",
		Country:     "USA",
		Language:    "english",
		Tags:        "rock, classic",
		Codec:       "MP3",
		Bitrate:     128,
	}

	// Block the station
	msg, err := manager.Block(ctx, station)
	if err != nil {
		t.Fatalf("Block failed: %v", err)
	}

	if msg == "" {
		t.Error("Expected non-empty message")
	}

	// Verify it's blocked
	if !manager.IsBlocked(station.StationUUID) {
		t.Error("Station should be blocked")
	}

	if manager.Count() != 1 {
		t.Errorf("Expected 1 blocked station, got %d", manager.Count())
	}

	// Verify file was created
	if _, err := os.Stat(blocklistPath); os.IsNotExist(err) {
		t.Error("Blocklist file was not created")
	}
}

func TestBlockDuplicateStation(t *testing.T) {
	tmpDir := t.TempDir()
	blocklistPath := filepath.Join(tmpDir, "blocklist.json")

	manager := NewManager(blocklistPath)
	ctx := context.Background()

	station := &api.Station{
		StationUUID: "test-uuid-123",
		Name:        "Test Radio",
	}

	// Block the station
	_, err := manager.Block(ctx, station)
	if err != nil {
		t.Fatalf("First block failed: %v", err)
	}

	// Try to block again
	_, err = manager.Block(ctx, station)
	if err != ErrStationAlreadyBlocked {
		t.Errorf("Expected ErrStationAlreadyBlocked, got %v", err)
	}

	if manager.Count() != 1 {
		t.Errorf("Expected 1 blocked station, got %d", manager.Count())
	}
}

func TestUnblockStation(t *testing.T) {
	tmpDir := t.TempDir()
	blocklistPath := filepath.Join(tmpDir, "blocklist.json")

	manager := NewManager(blocklistPath)
	ctx := context.Background()

	station := &api.Station{
		StationUUID: "test-uuid-123",
		Name:        "Test Radio",
	}

	// Block the station
	_, err := manager.Block(ctx, station)
	if err != nil {
		t.Fatalf("Block failed: %v", err)
	}

	// Unblock the station
	err = manager.Unblock(ctx, station.StationUUID)
	if err != nil {
		t.Fatalf("Unblock failed: %v", err)
	}

	// Verify it's not blocked
	if manager.IsBlocked(station.StationUUID) {
		t.Error("Station should not be blocked")
	}

	if manager.Count() != 0 {
		t.Errorf("Expected 0 blocked stations, got %d", manager.Count())
	}
}

func TestUnblockNonExistentStation(t *testing.T) {
	tmpDir := t.TempDir()
	blocklistPath := filepath.Join(tmpDir, "blocklist.json")

	manager := NewManager(blocklistPath)
	ctx := context.Background()

	err := manager.Unblock(ctx, "non-existent-uuid")
	if err != ErrStationNotBlocked {
		t.Errorf("Expected ErrStationNotBlocked, got %v", err)
	}
}

func TestGetAll(t *testing.T) {
	tmpDir := t.TempDir()
	blocklistPath := filepath.Join(tmpDir, "blocklist.json")

	manager := NewManager(blocklistPath)
	ctx := context.Background()

	// Block multiple stations
	stations := []*api.Station{
		{StationUUID: "uuid-1", Name: "Station 1"},
		{StationUUID: "uuid-2", Name: "Station 2"},
		{StationUUID: "uuid-3", Name: "Station 3"},
	}

	for _, station := range stations {
		_, err := manager.Block(ctx, station)
		if err != nil {
			t.Fatalf("Block failed: %v", err)
		}
		// Small delay to ensure different timestamps
		time.Sleep(1 * time.Millisecond)
	}

	// Get all blocked stations
	blocked := manager.GetAll()
	if len(blocked) != 3 {
		t.Errorf("Expected 3 blocked stations, got %d", len(blocked))
	}

	// Verify most recent is first (uuid-3)
	if blocked[0].StationUUID != "uuid-3" {
		t.Errorf("Expected most recent station (uuid-3) first, got %s", blocked[0].StationUUID)
	}
}

func TestClear(t *testing.T) {
	tmpDir := t.TempDir()
	blocklistPath := filepath.Join(tmpDir, "blocklist.json")

	manager := NewManager(blocklistPath)
	ctx := context.Background()

	// Block multiple stations
	for i := 0; i < 5; i++ {
		station := &api.Station{
			StationUUID: string(rune('a' + i)),
			Name:        "Station",
		}
		_, err := manager.Block(ctx, station)
		if err != nil {
			t.Fatalf("Block failed: %v", err)
		}
	}

	if manager.Count() != 5 {
		t.Errorf("Expected 5 blocked stations, got %d", manager.Count())
	}

	// Clear all
	err := manager.Clear(ctx)
	if err != nil {
		t.Fatalf("Clear failed: %v", err)
	}

	if manager.Count() != 0 {
		t.Errorf("Expected 0 blocked stations after clear, got %d", manager.Count())
	}
}

func TestPersistence(t *testing.T) {
	tmpDir := t.TempDir()
	blocklistPath := filepath.Join(tmpDir, "blocklist.json")

	ctx := context.Background()

	// Create first manager and block stations
	manager1 := NewManager(blocklistPath)
	station1 := &api.Station{StationUUID: "uuid-1", Name: "Station 1", Country: "USA"}
	station2 := &api.Station{StationUUID: "uuid-2", Name: "Station 2", Language: "english"}

	_, err := manager1.Block(ctx, station1)
	if err != nil {
		t.Fatalf("Block failed: %v", err)
	}
	_, err = manager1.Block(ctx, station2)
	if err != nil {
		t.Fatalf("Block failed: %v", err)
	}

	// Create second manager and load from disk
	manager2 := NewManager(blocklistPath)
	err = manager2.Load(ctx)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	// Verify data was persisted
	if manager2.Count() != 2 {
		t.Errorf("Expected 2 blocked stations after load, got %d", manager2.Count())
	}

	if !manager2.IsBlocked("uuid-1") {
		t.Error("uuid-1 should be blocked after load")
	}
	if !manager2.IsBlocked("uuid-2") {
		t.Error("uuid-2 should be blocked after load")
	}

	// Verify metadata was preserved
	blocked := manager2.GetAll()
	for _, b := range blocked {
		if b.StationUUID == "uuid-1" && b.Country != "USA" {
			t.Errorf("Expected Country='USA', got '%s'", b.Country)
		}
		if b.StationUUID == "uuid-2" && b.Language != "english" {
			t.Errorf("Expected Language='english', got '%s'", b.Language)
		}
	}
}

func TestUndoLastBlock(t *testing.T) {
	tmpDir := t.TempDir()
	blocklistPath := filepath.Join(tmpDir, "blocklist.json")

	manager := NewManager(blocklistPath)
	ctx := context.Background()

	station := &api.Station{
		StationUUID: "test-uuid-123",
		Name:        "Test Radio",
	}

	// Block the station
	_, err := manager.Block(ctx, station)
	if err != nil {
		t.Fatalf("Block failed: %v", err)
	}

	// Verify it's blocked
	if !manager.IsBlocked(station.StationUUID) {
		t.Error("Station should be blocked")
	}

	// Undo the block
	undone, err := manager.UndoLastBlock(ctx)
	if err != nil {
		t.Fatalf("UndoLastBlock failed: %v", err)
	}
	if !undone {
		t.Error("Expected undo to succeed")
	}

	// Verify it's unblocked
	if manager.IsBlocked(station.StationUUID) {
		t.Error("Station should be unblocked after undo")
	}

	// Try to undo again (should return false)
	undone, err = manager.UndoLastBlock(ctx)
	if err != nil {
		t.Fatalf("Second UndoLastBlock failed: %v", err)
	}
	if undone {
		t.Error("Expected undo to return false when nothing to undo")
	}
}

func TestWarningMessages(t *testing.T) {
	tmpDir := t.TempDir()
	blocklistPath := filepath.Join(tmpDir, "blocklist.json")

	manager := NewManager(blocklistPath)
	ctx := context.Background()

	// Block stations up to warning threshold
	for i := 0; i < BlockWarningThreshold; i++ {
		station := &api.Station{
			StationUUID: fmt.Sprintf("uuid-warn-%d", i),
			Name:        "Station",
		}
		msg, err := manager.Block(ctx, station)
		if err != nil {
			t.Fatalf("Block %d failed: %v", i, err)
		}

		// Check if warning appears at threshold
		if i == BlockWarningThreshold-1 {
			if msg == "" {
				t.Error("Expected warning message at threshold")
			}
		}
	}
}

func TestConcurrentAccess(t *testing.T) {
	tmpDir := t.TempDir()
	blocklistPath := filepath.Join(tmpDir, "blocklist.json")

	manager := NewManager(blocklistPath)
	ctx := context.Background()

	// Test concurrent reads and writes
	done := make(chan bool)

	// Writer goroutine
	go func() {
		for i := 0; i < 10; i++ {
			station := &api.Station{
				StationUUID: string(rune('a' + i)),
				Name:        "Station",
			}
			_, _ = manager.Block(ctx, station)
			time.Sleep(1 * time.Millisecond)
		}
		done <- true
	}()

	// Reader goroutine
	go func() {
		for i := 0; i < 10; i++ {
			_ = manager.IsBlocked(string(rune('a' + i)))
			_ = manager.Count()
			_ = manager.GetAll()
			time.Sleep(1 * time.Millisecond)
		}
		done <- true
	}()

	// Wait for both to complete
	<-done
	<-done

	// Should not panic and should have blocked stations
	if manager.Count() == 0 {
		t.Error("Expected some blocked stations after concurrent access")
	}
}
