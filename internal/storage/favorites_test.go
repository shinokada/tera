package storage

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/shinokada/tera/internal/api"
)

func TestStorage_SaveList(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewStorage(tmpDir)

	list := &FavoritesList{
		Name: "test",
		Stations: []api.Station{
			{StationUUID: "1", Name: "Test Station"},
		},
	}

	err := store.SaveList(context.Background(), list)
	if err != nil {
		t.Fatalf("SaveList failed: %v", err)
	}

	// Verify file was created
	path := filepath.Join(tmpDir, "test.json")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Error("Expected file to be created")
	}

	// Verify content
	data, _ := os.ReadFile(path)
	var stations []api.Station
	json.Unmarshal(data, &stations)

	if len(stations) != 1 {
		t.Errorf("Expected 1 station, got %d", len(stations))
	}
}

func TestStorage_AddStation(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewStorage(tmpDir)

	station := api.Station{
		StationUUID: "test-1",
		Name:        "Test Station",
	}

	// Add to new list
	err := store.AddStation(context.Background(), "favorites", station)
	if err != nil {
		t.Fatalf("AddStation failed: %v", err)
	}

	// Verify it was added
	list, err := store.LoadList(context.Background(), "favorites")
	if err != nil {
		t.Fatalf("LoadList failed: %v", err)
	}

	if len(list.Stations) != 1 {
		t.Errorf("Expected 1 station, got %d", len(list.Stations))
	}

	if list.Stations[0].StationUUID != "test-1" {
		t.Errorf("Expected UUID 'test-1', got '%s'", list.Stations[0].StationUUID)
	}
}

func TestStorage_AddStation_Duplicate(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewStorage(tmpDir)

	station := api.Station{
		StationUUID: "test-1",
		Name:        "Test Station",
	}

	// Add first time
	err := store.AddStation(context.Background(), "favorites", station)
	if err != nil {
		t.Fatalf("First AddStation failed: %v", err)
	}

	// Try to add again (should fail)
	err = store.AddStation(context.Background(), "favorites", station)
	if err != ErrDuplicateStation {
		t.Errorf("Expected ErrDuplicateStation, got %v", err)
	}

	// Verify only one station in list
	list, _ := store.LoadList(context.Background(), "favorites")
	if len(list.Stations) != 1 {
		t.Errorf("Expected 1 station after duplicate attempt, got %d", len(list.Stations))
	}
}

func TestStorage_AddStation_ToExistingList(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewStorage(tmpDir)

	// Create initial list
	station1 := api.Station{StationUUID: "1", Name: "Station 1"}
	store.AddStation(context.Background(), "favorites", station1)

	// Add another station
	station2 := api.Station{StationUUID: "2", Name: "Station 2"}
	err := store.AddStation(context.Background(), "favorites", station2)
	if err != nil {
		t.Fatalf("AddStation failed: %v", err)
	}

	// Verify both stations
	list, _ := store.LoadList(context.Background(), "favorites")
	if len(list.Stations) != 2 {
		t.Errorf("Expected 2 stations, got %d", len(list.Stations))
	}
}

func TestStorage_StationExists(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewStorage(tmpDir)

	// Test non-existent list
	exists, err := store.StationExists(context.Background(), "nonexistent", "uuid")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if exists {
		t.Error("Expected station to not exist in non-existent list")
	}

	// Add a station
	station := api.Station{StationUUID: "test-1", Name: "Test"}
	store.AddStation(context.Background(), "favorites", station)

	// Test existing station
	exists, err = store.StationExists(context.Background(), "favorites", "test-1")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if !exists {
		t.Error("Expected station to exist")
	}

	// Test non-existing station in existing list
	exists, err = store.StationExists(context.Background(), "favorites", "other-uuid")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if exists {
		t.Error("Expected station to not exist")
	}
}
