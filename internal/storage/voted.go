package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// VotedStation represents a station that has been voted for
type VotedStation struct {
	StationUUID string    `json:"station_uuid"`
	VotedAt     time.Time `json:"voted_at"`
}

// VotedStations manages the list of voted stations
type VotedStations struct {
	Stations []VotedStation `json:"stations"`
	mu       sync.RWMutex
}

// GetVotedStationsPath returns the path to the voted stations file
func GetVotedStationsPath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("failed to get config directory: %w", err)
	}

	return filepath.Join(configDir, "tera", "voted_stations.json"), nil
}

// LoadVotedStations loads the list of voted stations
func LoadVotedStations() (*VotedStations, error) {
	path, err := GetVotedStationsPath()
	if err != nil {
		return nil, err
	}

	// If file doesn't exist, return empty list
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return &VotedStations{Stations: []VotedStation{}}, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read voted stations: %w", err)
	}

	var voted VotedStations
	if err := json.Unmarshal(data, &voted); err != nil {
		return nil, fmt.Errorf("failed to parse voted stations: %w", err)
	}

	// Clean up old entries (older than 10 minutes)
	voted.CleanupOldVotes()

	return &voted, nil
}

// SaveVotedStations saves the list of voted stations
func (v *VotedStations) Save() error {
	v.mu.Lock()
	defer v.mu.Unlock()

	path, err := GetVotedStationsPath()
	if err != nil {
		return err
	}

	// Ensure config directory exists
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal voted stations: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write voted stations: %w", err)
	}

	return nil
}

// AddVote adds a vote for a station
func (v *VotedStations) AddVote(stationUUID string) {
	v.mu.Lock()
	defer v.mu.Unlock()

	// Remove any existing vote for this station
	for i := 0; i < len(v.Stations); i++ {
		if v.Stations[i].StationUUID == stationUUID {
			v.Stations = append(v.Stations[:i], v.Stations[i+1:]...)
			break
		}
	}

	// Add new vote
	v.Stations = append(v.Stations, VotedStation{
		StationUUID: stationUUID,
		VotedAt:     time.Now(),
	})
}

// HasVoted checks if a station has been voted for recently (within 10 minutes)
func (v *VotedStations) HasVoted(stationUUID string) bool {
	v.mu.RLock()
	defer v.mu.RUnlock()

	cutoff := time.Now().Add(-10 * time.Minute)
	for _, station := range v.Stations {
		if station.StationUUID == stationUUID && station.VotedAt.After(cutoff) {
			return true
		}
	}
	return false
}

// CleanupOldVotes removes votes older than 10 minutes
func (v *VotedStations) CleanupOldVotes() {
	v.mu.Lock()
	defer v.mu.Unlock()

	cutoff := time.Now().Add(-10 * time.Minute)
	filtered := []VotedStation{}
	for _, station := range v.Stations {
		if station.VotedAt.After(cutoff) {
			filtered = append(filtered, station)
		}
	}
	v.Stations = filtered
}
