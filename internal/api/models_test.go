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
