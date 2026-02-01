package api

import (
	"encoding/json"
	"testing"
)

func TestStation_Unmarshal(t *testing.T) {
	jsonData := `{
        "stationuuid": "test-123",
        "name": "  Jazz FM  ",
        "url_resolved": "http://example.com",
        "votes": 100,
        "codec": "MP3",
        "bitrate": 128
    }`

	var station Station
	err := json.Unmarshal([]byte(jsonData), &station)
	if err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if station.StationUUID != "test-123" {
		t.Errorf("Expected UUID test-123, got %s", station.StationUUID)
	}

	if station.TrimName() != "Jazz FM" {
		t.Errorf("Expected trimmed name 'Jazz FM', got '%s'", station.TrimName())
	}
}

func TestStation_VolumeNilByDefault(t *testing.T) {
	// When JSON has no volume field, Volume should be nil (not 0)
	jsonData := `{
        "stationuuid": "test-123",
        "name": "Test Station"
    }`

	var station Station
	err := json.Unmarshal([]byte(jsonData), &station)
	if err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if station.Volume != nil {
		t.Errorf("Expected Volume to be nil for station without volume, got %d", *station.Volume)
	}

	// GetVolume should return -1 for nil
	if station.GetVolume() != -1 {
		t.Errorf("Expected GetVolume() to return -1 for nil, got %d", station.GetVolume())
	}
}

func TestStation_VolumeWithValue(t *testing.T) {
	// When JSON has volume field, it should be set
	jsonData := `{
        "stationuuid": "test-123",
        "name": "Test Station",
        "volume": 75
    }`

	var station Station
	err := json.Unmarshal([]byte(jsonData), &station)
	if err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if station.Volume == nil {
		t.Fatal("Expected Volume to be set, got nil")
	}

	if *station.Volume != 75 {
		t.Errorf("Expected Volume to be 75, got %d", *station.Volume)
	}

	if station.GetVolume() != 75 {
		t.Errorf("Expected GetVolume() to return 75, got %d", station.GetVolume())
	}
}

func TestStation_VolumeZero(t *testing.T) {
	// Volume of 0 should be valid (mute), not treated as "not set"
	jsonData := `{
        "stationuuid": "test-123",
        "name": "Test Station",
        "volume": 0
    }`

	var station Station
	err := json.Unmarshal([]byte(jsonData), &station)
	if err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if station.Volume == nil {
		t.Fatal("Expected Volume to be set (to 0), got nil")
	}

	if *station.Volume != 0 {
		t.Errorf("Expected Volume to be 0, got %d", *station.Volume)
	}
}

func TestStation_SetVolume(t *testing.T) {
	station := Station{Name: "Test"}

	// Initially nil
	if station.Volume != nil {
		t.Error("Expected initial Volume to be nil")
	}

	// Set volume
	station.SetVolume(50)

	if station.Volume == nil {
		t.Fatal("Expected Volume to be set after SetVolume, got nil")
	}

	if *station.Volume != 50 {
		t.Errorf("Expected Volume to be 50, got %d", *station.Volume)
	}

	// Change volume
	station.SetVolume(80)

	if *station.Volume != 80 {
		t.Errorf("Expected Volume to be 80 after second SetVolume, got %d", *station.Volume)
	}
}

func TestStation_VolumeOmittedInJSON(t *testing.T) {
	// When serializing, nil volume should be omitted from JSON
	station := Station{
		StationUUID: "test-123",
		Name:        "Test Station",
		// Volume is nil
	}

	data, err := json.Marshal(station)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	jsonStr := string(data)
	if contains(jsonStr, "volume") {
		t.Errorf("Expected JSON to omit volume field when nil, got: %s", jsonStr)
	}

	// With volume set, it should be included
	station.SetVolume(50)
	data, err = json.Marshal(station)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	jsonStr = string(data)
	if !contains(jsonStr, `"volume":50`) {
		t.Errorf("Expected JSON to contain volume:50, got: %s", jsonStr)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
