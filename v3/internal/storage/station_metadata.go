package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"github.com/shinokada/tera/v3/internal/api"
)

// StationMetadata tracks listening statistics for a station
type StationMetadata struct {
	PlayCount            int       `json:"play_count"`
	LastPlayed           time.Time `json:"last_played"`
	FirstPlayed          time.Time `json:"first_played"`
	TotalDurationSeconds int64     `json:"total_duration_seconds"`
}

// MetadataStore holds all station metadata (no mutex - protected by manager)
type MetadataStore struct {
	Stations map[string]*StationMetadata `json:"stations"`
	Version  int                         `json:"version"`
}

// StationWithMetadata combines station info with its metadata for display
type StationWithMetadata struct {
	Station  api.Station
	Metadata *StationMetadata
}

// MetadataManager manages station play statistics
type MetadataManager struct {
	dataPath      string
	store         *MetadataStore
	mu            sync.RWMutex
	savePending   atomic.Bool
	currentPlay   string    // Track current playing station to prevent duplicates
	playStartTime time.Time // When current play started
	ctx           context.Context
	cancel        context.CancelFunc
	wg            sync.WaitGroup
}

// NewMetadataManager creates a new metadata manager
func NewMetadataManager(dataPath string) (*MetadataManager, error) {
	ctx, cancel := context.WithCancel(context.Background())
	m := &MetadataManager{
		dataPath: dataPath,
		store: &MetadataStore{
			Stations: make(map[string]*StationMetadata),
			Version:  1,
		},
		ctx:    ctx,
		cancel: cancel,
	}

	// Load existing metadata
	if err := m.Load(); err != nil {
		// If file doesn't exist or is corrupted, start fresh
		// Log warning but don't fail
		_ = err // Silent failure - start with empty store
	}

	// Start background save goroutine
	m.wg.Add(1)
	go m.saveLoop()

	return m, nil
}

// getMetadataFilePath returns the full path to the metadata file
func (m *MetadataManager) getMetadataFilePath() string {
	return filepath.Join(m.dataPath, "station_metadata.json")
}

// Load loads metadata from disk
func (m *MetadataManager) Load() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	filePath := m.getMetadataFilePath()
	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			// File doesn't exist yet, use empty store
			return nil
		}
		return fmt.Errorf("failed to read metadata file: %w", err)
	}

	var store MetadataStore
	if err := json.Unmarshal(data, &store); err != nil {
		return fmt.Errorf("failed to parse metadata file: %w", err)
	}

	// Ensure stations map is initialized
	if store.Stations == nil {
		store.Stations = make(map[string]*StationMetadata)
	}

	m.store = &store
	return nil
}

// Save saves metadata to disk
func (m *MetadataManager) Save() error {
	m.mu.RLock()
	data, err := json.MarshalIndent(m.store, "", "  ")
	m.mu.RUnlock()

	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	filePath := m.getMetadataFilePath()

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		return fmt.Errorf("failed to create metadata directory: %w", err)
	}

	// Write to a temp file then rename for crash-safety: a direct write would
	// leave a truncated/corrupt file if the process dies mid-write.
	tmpPath := filePath + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write metadata temp file: %w", err)
	}
	if err := os.Rename(tmpPath, filePath); err != nil {
		_ = os.Remove(tmpPath) // best-effort cleanup
		return fmt.Errorf("failed to rename metadata file: %w", err)
	}

	return nil
}

// StartPlay records that a station started playing
func (m *MetadataManager) StartPlay(stationUUID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()

	// Deduplicate: ignore if same station already playing
	if m.currentPlay == stationUUID {
		return nil
	}

	// Stop previous if exists (record duration)
	if m.currentPlay != "" {
		m.stopPlayLocked(m.currentPlay)
	}

	m.currentPlay = stationUUID
	m.playStartTime = now

	// Get or create metadata for this station
	metadata, exists := m.store.Stations[stationUUID]
	if !exists {
		metadata = &StationMetadata{
			FirstPlayed: now,
		}
		m.store.Stations[stationUUID] = metadata
	}

	// Increment play count and update last played
	metadata.PlayCount++
	metadata.LastPlayed = now

	// Mark save as pending
	m.savePending.Store(true)

	return nil
}

// StopPlay records that a station stopped playing
func (m *MetadataManager) StopPlay(stationUUID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Only stop if this is the current playing station
	if m.currentPlay != stationUUID {
		return nil
	}

	m.stopPlayLocked(stationUUID)
	return nil
}

// stopPlayLocked stops play for a station (must be called with lock held)
func (m *MetadataManager) stopPlayLocked(stationUUID string) {
	// Calculate duration
	if !m.playStartTime.IsZero() {
		duration := time.Since(m.playStartTime)
		if metadata, exists := m.store.Stations[stationUUID]; exists {
			metadata.TotalDurationSeconds += int64(duration.Seconds())
		}
	}

	m.currentPlay = ""
	m.playStartTime = time.Time{}
	m.savePending.Store(true)
}

// GetMetadata returns metadata for a station, or nil if not found
func (m *MetadataManager) GetMetadata(stationUUID string) *StationMetadata {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if metadata, exists := m.store.Stations[stationUUID]; exists {
		// Return a copy to prevent external modification
		metaCopy := *metadata
		return &metaCopy
	}
	return nil
}

// GetTopPlayed returns stations sorted by play count (most played first)
func (m *MetadataManager) GetTopPlayed(limit int) []StationWithMetadata {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Collect all stations with metadata
	var result []StationWithMetadata
	for uuid, metadata := range m.store.Stations {
		metaCopy := *metadata
		result = append(result, StationWithMetadata{
			Station:  api.Station{StationUUID: uuid},
			Metadata: &metaCopy,
		})
	}

	// Sort by play count (descending)
	sort.Slice(result, func(i, j int) bool {
		return result[i].Metadata.PlayCount > result[j].Metadata.PlayCount
	})

	// Limit results
	if limit > 0 && len(result) > limit {
		result = result[:limit]
	}

	return result
}

// GetRecentlyPlayed returns stations sorted by last played time (most recent first)
func (m *MetadataManager) GetRecentlyPlayed(limit int) []StationWithMetadata {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Collect all stations with metadata
	var result []StationWithMetadata
	for uuid, metadata := range m.store.Stations {
		metaCopy := *metadata
		result = append(result, StationWithMetadata{
			Station:  api.Station{StationUUID: uuid},
			Metadata: &metaCopy,
		})
	}

	// Sort by last played (most recent first)
	sort.Slice(result, func(i, j int) bool {
		return result[i].Metadata.LastPlayed.After(result[j].Metadata.LastPlayed)
	})

	// Limit results
	if limit > 0 && len(result) > limit {
		result = result[:limit]
	}

	return result
}

// GetFirstPlayed returns stations sorted by first played time (oldest first)
func (m *MetadataManager) GetFirstPlayed(limit int) []StationWithMetadata {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []StationWithMetadata
	for uuid, metadata := range m.store.Stations {
		metaCopy := *metadata
		result = append(result, StationWithMetadata{
			Station:  api.Station{StationUUID: uuid},
			Metadata: &metaCopy,
		})
	}

	// Sort by first played (oldest first)
	sort.Slice(result, func(i, j int) bool {
		if result[i].Metadata.FirstPlayed.IsZero() {
			return false
		}
		if result[j].Metadata.FirstPlayed.IsZero() {
			return true
		}
		return result[i].Metadata.FirstPlayed.Before(result[j].Metadata.FirstPlayed)
	})

	if limit > 0 && len(result) > limit {
		result = result[:limit]
	}

	return result
}

// GetAllStationUUIDs returns all station UUIDs with metadata
func (m *MetadataManager) GetAllStationUUIDs() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	uuids := make([]string, 0, len(m.store.Stations))
	for uuid := range m.store.Stations {
		uuids = append(uuids, uuid)
	}
	return uuids
}

// GetTotalStations returns the count of stations with metadata
func (m *MetadataManager) GetTotalStations() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.store.Stations)
}

// ClearAll removes all metadata (for testing or user request)
func (m *MetadataManager) ClearAll() error {
	m.mu.Lock()
	m.store.Stations = make(map[string]*StationMetadata)
	m.currentPlay = ""
	m.playStartTime = time.Time{}
	m.mu.Unlock()

	m.savePending.Store(true)
	return m.Save()
}

// saveLoop runs in the background and saves periodically when changes are pending
func (m *MetadataManager) saveLoop() {
	defer m.wg.Done()

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if m.savePending.CompareAndSwap(true, false) {
				if err := m.Save(); err != nil {
					// Re-set pending so next tick retries the save
					m.savePending.Store(true)
				}
			}
		case <-m.ctx.Done():
			// Save any pending changes before exiting
			if m.savePending.Load() {
				_ = m.Save()
			}
			return
		}
	}
}

// Close stops the background save goroutine and saves any pending changes
func (m *MetadataManager) Close() error {
	// Stop any current play to record duration
	m.mu.Lock()
	if m.currentPlay != "" {
		m.stopPlayLocked(m.currentPlay)
	}
	m.mu.Unlock()

	// Signal background goroutine to stop
	m.cancel()

	// Wait for goroutine to finish
	m.wg.Wait()

	return nil
}

// FormatLastPlayed formats the last played time as a human-readable relative string
func FormatLastPlayed(t time.Time) string {
	if t.IsZero() {
		return "Never"
	}

	now := time.Now()
	diff := now.Sub(t)

	switch {
	case diff < time.Minute:
		return "Just now"
	case diff < time.Hour:
		minutes := int(diff.Minutes())
		if minutes == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", minutes)
	case diff < 24*time.Hour:
		hours := int(diff.Hours())
		if hours == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hours)
	case diff < 7*24*time.Hour:
		days := int(diff.Hours() / 24)
		if days == 1 {
			return "Yesterday"
		}
		return fmt.Sprintf("%d days ago", days)
	case diff < 30*24*time.Hour:
		weeks := int(diff.Hours() / (24 * 7))
		if weeks == 1 {
			return "1 week ago"
		}
		return fmt.Sprintf("%d weeks ago", weeks)
	default:
		// Format as date for older entries
		return t.Format("Jan 2, 2006")
	}
}

// FormatDuration formats duration in seconds as a human-readable string
func FormatDuration(seconds int64) string {
	if seconds < 60 {
		return fmt.Sprintf("%ds", seconds)
	}
	if seconds < 3600 {
		return fmt.Sprintf("%dm", seconds/60)
	}
	hours := seconds / 3600
	minutes := (seconds % 3600) / 60
	if minutes == 0 {
		return fmt.Sprintf("%dh", hours)
	}
	return fmt.Sprintf("%dh %dm", hours, minutes)
}
