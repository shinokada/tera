package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Voting System Design:
//
// The Radio Browser API enforces a 10-minute cooldown between votes for the same station.
// However, we store voted stations PERMANENTLY (not just for 10 minutes) because:
//
// 1. User Experience: Once you vote, you should always see "✓ You voted" to avoid confusion
// 2. Prevent Duplicates: Users won't accidentally vote multiple times for the same station
// 3. API Compliance: We still check the 10-minute cooldown before allowing re-votes
//
// After 10 minutes, users CAN vote again (API allows it), but they'll see they already voted.
// This is intentional - most users only want to vote once per station.
//
// To clear vote history: Use ClearAll() or RemoveVote() methods.

const voteCooldown = 10 * time.Minute

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

	// Handle empty file gracefully
	if len(data) == 0 {
		return &VotedStations{Stations: []VotedStation{}}, nil
	}

	var voted VotedStations
	if err := json.Unmarshal(data, &voted); err != nil {
		return nil, fmt.Errorf("failed to parse voted stations: %w", err)
	}

	// Note: We no longer clean up old votes automatically.
	// Voted status is now permanent to prevent duplicate votes.

	return &voted, nil
}

// saveLocked persists state to disk. Caller must hold v.mu.
func (v *VotedStations) saveLocked() error {
	path, err := GetVotedStationsPath()
	if err != nil {
		return err
	}

	// Ensure config directory exists (0700 for security)
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal voted stations: %w", err)
	}

	// Atomic write: write to temp file then rename
	tmpPath := path + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write voted stations: %w", err)
	}

	if err := os.Rename(tmpPath, path); err != nil {
		return fmt.Errorf("failed to rename voted stations file: %w", err)
	}

	return nil
}

// Save saves the list of voted stations (public API - acquires lock)
func (v *VotedStations) Save() error {
	v.mu.Lock()
	defer v.mu.Unlock()
	return v.saveLocked()
}

// AddVote adds a vote for a station or updates the timestamp if already voted
func (v *VotedStations) AddVote(stationUUID string) error {
	v.mu.Lock()
	defer v.mu.Unlock()
	
	// Update existing vote timestamp if found
	found := false
	for i := 0; i < len(v.Stations); i++ {
		if v.Stations[i].StationUUID == stationUUID {
			v.Stations[i].VotedAt = time.Now()
			found = true
			break
		}
	}

	if !found {
		// Add new vote if not found
		v.Stations = append(v.Stations, VotedStation{
			StationUUID: stationUUID,
			VotedAt:     time.Now(),
		})
	}

	// Persist changes using saveLocked (we already hold the lock)
	return v.saveLocked()
}

// HasVoted checks if a station has been voted for (permanent record)
func (v *VotedStations) HasVoted(stationUUID string) bool {
	v.mu.RLock()
	defer v.mu.RUnlock()

	for _, station := range v.Stations {
		if station.StationUUID == stationUUID {
			return true
		}
	}
	return false
}

// CanVoteAgain checks if enough time has passed to vote again (10 minutes)
func (v *VotedStations) CanVoteAgain(stationUUID string) bool {
	v.mu.RLock()
	defer v.mu.RUnlock()

	cutoff := time.Now().Add(-voteCooldown)
	for _, station := range v.Stations {
		if station.StationUUID == stationUUID {
			// If voted more than 10 minutes ago, can vote again
			return station.VotedAt.Before(cutoff)
		}
	}
	// Never voted, so can vote
	return true
}

// CleanupOldVotes removes votes older than specified duration
// Note: This is kept for potential future use, but is not called automatically.
// Votes are now stored permanently to prevent accidental duplicate voting.
func (v *VotedStations) CleanupOldVotes(olderThan time.Duration) {
	v.mu.Lock()
	defer v.mu.Unlock()

	cutoff := time.Now().Add(-olderThan)
	filtered := []VotedStation{}
	for _, station := range v.Stations {
		if station.VotedAt.After(cutoff) {
			filtered = append(filtered, station)
		}
	}
	v.Stations = filtered
}

// ClearAll removes all vote history
func (v *VotedStations) ClearAll() error {
	v.mu.Lock()
	defer v.mu.Unlock()
	
	v.Stations = []VotedStation{}
	return v.saveLocked()
}

// RemoveVote removes a specific station from vote history
func (v *VotedStations) RemoveVote(stationUUID string) error {
	v.mu.Lock()
	defer v.mu.Unlock()
	
	found := false
	for i := 0; i < len(v.Stations); i++ {
		if v.Stations[i].StationUUID == stationUUID {
			v.Stations = append(v.Stations[:i], v.Stations[i+1:]...)
			found = true
			break
		}
	}
	
	if found {
		// Persist using saveLocked (we already hold the lock)
		return v.saveLocked()
	}
	return nil
}
