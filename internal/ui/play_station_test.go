package ui

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/shinokada/tera/v2/internal/api"
	"github.com/shinokada/tera/v2/internal/blocklist"
)

func TestStationListItem(t *testing.T) {
	station := api.Station{
		Name:    "  Jazz FM  ",
		Country: "USA",
		Codec:   "MP3",
		Bitrate: 128,
	}

	item := stationListItem{station: station}

	// Test FilterValue
	if item.FilterValue() != "  Jazz FM  " {
		t.Errorf("Expected FilterValue '  Jazz FM  ', got '%s'", item.FilterValue())
	}

	// Test Title (now returns combined info in single line)
	expectedTitle := "Jazz FM • USA • MP3 128kbps"
	if item.Title() != expectedTitle {
		t.Errorf("Expected Title '%s', got '%s'", expectedTitle, item.Title())
	}

	// Test Description (now returns empty string)
	desc := item.Description()
	if desc != "" {
		t.Errorf("Expected empty description, got '%s'", desc)
	}
}

func TestStationListItem_Blocked(t *testing.T) {
	station := api.Station{
		Name: "Blocked Station",
	}

	item := stationListItem{station: station, isBlocked: true}

	// Title should include the 🚫 icon
	expectedTitle := "🚫 Blocked Station"
	if item.Title() != expectedTitle {
		t.Errorf("Expected Title '%s', got '%s'", expectedTitle, item.Title())
	}
}

func TestStationListItem_EmptyFields(t *testing.T) {
	station := api.Station{
		Name: "Test Station",
	}

	item := stationListItem{station: station}

	// Title should just be the name when no other fields present
	if item.Title() != "Test Station" {
		t.Errorf("Expected Title 'Test Station', got '%s'", item.Title())
	}

	desc := item.Description()
	if desc != "" {
		t.Errorf("Expected empty description, got '%s'", desc)
	}
}

func TestGetStationsFromList(t *testing.T) {
	// Create temp directory
	tmpDir := t.TempDir()

	// Create test stations
	stations := []api.Station{
		{StationUUID: "3", Name: "Zebra Radio"},
		{StationUUID: "1", Name: "Alpha FM"},
		{StationUUID: "2", Name: "Beta Station"},
	}

	// Write test file
	data, _ := json.Marshal(stations)
	listPath := filepath.Join(tmpDir, "test.json")
	if err := os.WriteFile(listPath, data, 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Test loading
	blocklistManager := blocklist.NewManager(filepath.Join(tmpDir, "blocklist.json"))
	model := NewPlayModel(tmpDir, blocklistManager)
	loaded, err := model.getStationsFromList("test")

	if err != nil {
		t.Fatalf("getStationsFromList failed: %v", err)
	}

	if len(loaded) != 3 {
		t.Errorf("Expected 3 stations, got %d", len(loaded))
	}

	// Verify alphabetical sorting
	if loaded[0].Name != "Alpha FM" {
		t.Errorf("Expected first station 'Alpha FM', got '%s'", loaded[0].Name)
	}
	if loaded[1].Name != "Beta Station" {
		t.Errorf("Expected second station 'Beta Station', got '%s'", loaded[1].Name)
	}
	if loaded[2].Name != "Zebra Radio" {
		t.Errorf("Expected third station 'Zebra Radio', got '%s'", loaded[2].Name)
	}
}

func TestGetStationsFromList_EmptyList(t *testing.T) {
	tmpDir := t.TempDir()

	// Create empty list
	data, _ := json.Marshal([]api.Station{})
	listPath := filepath.Join(tmpDir, "empty.json")
	if err := os.WriteFile(listPath, data, 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	blocklistManager := blocklist.NewManager(filepath.Join(tmpDir, "blocklist.json"))
	model := NewPlayModel(tmpDir, blocklistManager)
	loaded, err := model.getStationsFromList("empty")

	if err != nil {
		t.Fatalf("getStationsFromList failed: %v", err)
	}

	if len(loaded) != 0 {
		t.Errorf("Expected 0 stations, got %d", len(loaded))
	}
}

func TestGetStationsFromList_NonexistentFile(t *testing.T) {
	tmpDir := t.TempDir()

	blocklistManager := blocklist.NewManager(filepath.Join(tmpDir, "blocklist.json"))
	model := NewPlayModel(tmpDir, blocklistManager)
	_, err := model.getStationsFromList("nonexistent")

	if err == nil {
		t.Error("Expected error for nonexistent file, got nil")
	}
}

func TestPlayModel_Update_StationsLoaded(t *testing.T) {
	blocklistManager := blocklist.NewManager("/tmp/blocklist.json")
	model := NewPlayModel("/tmp/favorites", blocklistManager)
	model.width = 80
	model.height = 24
	model.selectedList = "test"

	// Create test stations
	stations := []api.Station{
		{StationUUID: "1", Name: "Test Station 1"},
		{StationUUID: "2", Name: "Test Station 2"},
	}

	// Send stationsLoadedMsg
	updatedModel, _ := model.Update(stationsLoadedMsg{stations: stations})
	m := updatedModel.(PlayModel)

	if len(m.stations) != 2 {
		t.Errorf("Expected 2 stations, got %d", len(m.stations))
	}

	if len(m.stationItems) != 2 {
		t.Errorf("Expected 2 station items, got %d", len(m.stationItems))
	}

	if m.stationListModel.Items() == nil {
		t.Error("Expected stationListModel to be initialized")
	}
}

func TestPlayModel_Update_StationSelectionNavigation(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		keyType  tea.KeyType
		expected playState
	}{
		{"Escape key", "esc", tea.KeyEsc, playStateListSelection},
		{"Other key", "a", tea.KeyRunes, playStateStationSelection},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			blocklistManager := blocklist.NewManager("/tmp/blocklist.json")
			model := NewPlayModel("/tmp/favorites", blocklistManager)
			model.width = 80
			model.height = 24
			model.state = playStateStationSelection
			model.selectedList = "test"

			// Initialize with stations
			updatedModel, _ := model.Update(stationsLoadedMsg{
				stations: []api.Station{
					{StationUUID: "1", Name: "Test"},
				},
			})
			model = updatedModel.(PlayModel)

			// Send key message
			var keyMsg tea.KeyMsg
			if tt.keyType == tea.KeyEsc {
				keyMsg = tea.KeyMsg{Type: tea.KeyEsc}
			} else {
				keyMsg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(tt.key)}
			}

			// Call Update which will route to updateStationSelection
			updatedModel, _ = model.Update(keyMsg)
			m := updatedModel.(PlayModel)

			if m.state != tt.expected {
				t.Errorf("Expected state %v, got %v", tt.expected, m.state)
			}

			// If we navigated back, verify cleanup
			if tt.expected == playStateListSelection {
				if m.stations != nil {
					t.Error("Expected stations to be cleared")
				}
				if m.stationItems != nil {
					t.Error("Expected stationItems to be cleared")
				}
			}
		})
	}
}

func TestPlayModel_View_StationSelection(t *testing.T) {
	blocklistManager := blocklist.NewManager("/tmp/blocklist.json")
	model := NewPlayModel("/tmp/favorites", blocklistManager)
	model.width = 80
	model.height = 24
	model.state = playStateStationSelection
	model.selectedList = "Jazz"

	// Test with stations
	updatedModel, _ := model.Update(stationsLoadedMsg{
		stations: []api.Station{
			{StationUUID: "1", Name: "Jazz FM"},
		},
	})
	model = updatedModel.(PlayModel)

	view := model.viewStationSelection()

	if view == "" {
		t.Error("Expected non-empty view")
	}

	// Should contain list name
	if !contains(view, "Jazz") {
		t.Error("Expected view to contain list name 'Jazz'")
	}
}

func TestPlayModel_View_NoStations(t *testing.T) {
	blocklistManager := blocklist.NewManager("/tmp/blocklist.json")
	model := NewPlayModel("/tmp/favorites", blocklistManager)
	model.state = playStateStationSelection
	model.selectedList = "Empty"

	view := model.viewStationSelection()

	if view == "" {
		t.Error("Expected non-empty view")
	}

	// Should show empty message
	if !contains(view, "empty") {
		t.Error("Expected view to mention empty list")
	}
}

func TestNoStationsView(t *testing.T) {
	view := noStationsView("Test List")

	if view == "" {
		t.Error("Expected non-empty view")
	}

	if !contains(view, "Test List") {
		t.Error("Expected view to contain list name")
	}

	if !contains(view, "empty") {
		t.Error("Expected view to mention empty")
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 &&
		(s == substr || len(s) >= len(substr) && findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
