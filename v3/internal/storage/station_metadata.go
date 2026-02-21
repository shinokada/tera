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

// CachedStation stores essential station info for display in Most Played
type CachedStation struct {
	Name     string `json:"name"`
	URL      string `json:"url"`
	Country  string `json:"country,omitempty"`
	Language string `json:"language,omitempty"`
	Tags     string `json:"tags,omitempty"`
	Codec    string `json:"codec,omitempty"`
	Bitrate  int    `json:"bitrate,omitempty"`
	Votes    int    `json:"votes,omitempty"`
}

// MetadataStore holds all station metadata (no mutex - protected by manager)
type MetadataStore struct {
	Stations     map[string]*StationMetadata `json:"stations"`
	StationCache map[string]*CachedStation   `json:"station_cache,omitempty"` // Cache station info for display
	Version      int                         `json:"version"`
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
	saveMu        sync.Mutex // serializes concurrent Save() calls
	savePending   atomic.Bool
	currentPlay   string    // Track current playing station to prevent duplicates
	playStartTime time.Time // When current play started
	ctx           context.Context
	cancel        context.CancelFunc
	wg            sync.WaitGroup
}

// NewMetadataManager creates a new metadata manager.
// The error return is always nil in the current implementation: load failures
// are treated as a non-fatal "start fresh" condition. The signature is kept
// for forward-compatibility in case future callers need to detect init errors.
func NewMetadataManager(dataPath string) (*MetadataManager, error) {
	ctx, cancel := context.WithCancel(context.Background())
	m := &MetadataManager{
		dataPath: dataPath,
		store: &MetadataStore{
			Stations:     make(map[string]*StationMetadata),
			StationCache: make(map[string]*CachedStation),
			Version:      1,
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

	// Ensure maps are initialized
	if store.Stations == nil {
		store.Stations = make(map[string]*StationMetadata)
	}
	if store.StationCache == nil {
		store.StationCache = make(map[string]*CachedStation)
	}

	m.store = &store
	return nil
}

// Save saves metadata to disk
func (m *MetadataManager) Save() error {
	m.saveMu.Lock()
	defer m.saveMu.Unlock()

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
		_ = os.Remove(tmpPath) // best-effort cleanup of partial write
		return fmt.Errorf("failed to write metadata temp file: %w", err)
	}
	if err := os.Rename(tmpPath, filePath); err != nil {
		_ = os.Remove(tmpPath) // best-effort cleanup
		return fmt.Errorf("failed to rename metadata file: %w", err)
	}

	return nil
}

// StartPlay records that a station started playing and caches station info
func (m *MetadataManager) StartPlay(station *api.Station) error {
	if station == nil {
		return nil
	}
	stationUUID := station.StationUUID

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

	// Cache station info for later display
	m.store.StationCache[stationUUID] = &CachedStation{
		Name:     station.Name,
		URL:      station.URLResolved,
		Country:  station.Country,
		Language: station.Language,
		Tags:     station.Tags,
		Codec:    station.Codec,
		Bitrate:  station.Bitrate,
		Votes:    station.Votes,
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
		if duration > 0 {
			if metadata, exists := m.store.Stations[stationUUID]; exists {
				metadata.TotalDurationSeconds += int64(duration.Seconds())
			}
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

// GetCachedStation returns cached station info for a station, or nil if not found
func (m *MetadataManager) GetCachedStation(stationUUID string) *CachedStation {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if cached, exists := m.store.StationCache[stationUUID]; exists {
		// Return a copy to prevent external modification
		cachedCopy := *cached
		return &cachedCopy
	}
	return nil
}

// sortedStationsLocked collects all station entries, sorts them by less, and
// truncates to at most limit results (0 = no limit). Must be called with RLock held.
// Station info is populated from the cache if available.
func (m *MetadataManager) sortedStationsLocked(less func(a, b StationWithMetadata) bool, limit int) []StationWithMetadata {
	result := make([]StationWithMetadata, 0, len(m.store.Stations))
	for uuid, metadata := range m.store.Stations {
		metaCopy := *metadata
		station := api.Station{StationUUID: uuid}

		// Populate station info from cache if available
		if cached, ok := m.store.StationCache[uuid]; ok {
			station.Name = cached.Name
			station.URLResolved = cached.URL
			station.Country = cached.Country
			station.Language = cached.Language
			station.Tags = cached.Tags
			station.Codec = cached.Codec
			station.Bitrate = cached.Bitrate
			station.Votes = cached.Votes
		}

		result = append(result, StationWithMetadata{
			Station:  station,
			Metadata: &metaCopy,
		})
	}
	sort.Slice(result, func(i, j int) bool {
		if less(result[i], result[j]) {
			return true
		}
		if less(result[j], result[i]) {
			return false
		}
		// Tiebreak by UUID for deterministic ordering across calls
		return result[i].Station.StationUUID < result[j].Station.StationUUID
	})
	if limit > 0 && len(result) > limit {
		result = result[:limit]
	}
	return result
}

// GetTopPlayed returns stations sorted by play count (most played first)
func (m *MetadataManager) GetTopPlayed(limit int) []StationWithMetadata {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.sortedStationsLocked(func(a, b StationWithMetadata) bool {
		return a.Metadata.PlayCount > b.Metadata.PlayCount
	}, limit)
}

// GetRecentlyPlayed returns stations sorted by last played time (most recent first)
func (m *MetadataManager) GetRecentlyPlayed(limit int) []StationWithMetadata {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.sortedStationsLocked(func(a, b StationWithMetadata) bool {
		return a.Metadata.LastPlayed.After(b.Metadata.LastPlayed)
	}, limit)
}

// GetFirstPlayed returns stations sorted by first played time (oldest first)
func (m *MetadataManager) GetFirstPlayed(limit int) []StationWithMetadata {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.sortedStationsLocked(func(a, b StationWithMetadata) bool {
		if a.Metadata.FirstPlayed.IsZero() {
			return false
		}
		if b.Metadata.FirstPlayed.IsZero() {
			return true
		}
		return a.Metadata.FirstPlayed.Before(b.Metadata.FirstPlayed)
	}, limit)
}

// GetAllStationUUIDs returns all station UUIDs with metadata in sorted order.
func (m *MetadataManager) GetAllStationUUIDs() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	uuids := make([]string, 0, len(m.store.Stations))
	for uuid := range m.store.Stations {
		uuids = append(uuids, uuid)
	}
	sort.Strings(uuids)
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
	m.store.StationCache = make(map[string]*CachedStation)
	m.currentPlay = ""
	m.playStartTime = time.Time{}
	m.mu.Unlock()

	// Save is called unconditionally, so no need to set savePending here.
	return m.Save()
}

// saveLoop runs in the background and saves periodically when changes are pending.
// The final save on shutdown is handled by Close() after wg.Wait() returns.
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
			return
		}
	}
}

// Close stops the background save goroutine and saves any pending changes.
// It returns any error from the final disk write so callers can log or handle it.
func (m *MetadataManager) Close() error {
	// Stop any current play to record duration
	m.mu.Lock()
	if m.currentPlay != "" {
		m.stopPlayLocked(m.currentPlay)
	}
	m.mu.Unlock()

	// Signal background goroutine to stop and wait for it to exit
	m.cancel()
	m.wg.Wait()

	// Perform the final save here (after the goroutine has stopped) so the
	// error is propagated to the caller instead of being silently discarded.
	if m.savePending.Load() {
		return m.Save()
	}
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
	case diff < 365*24*time.Hour:
		months := int(diff.Hours() / (24 * 30))
		if months >= 12 {
			return "About a year ago"
		}
		if months == 1 {
			return "1 month ago"
		}
		return fmt.Sprintf("%d months ago", months)
	default:
		// Format as date for entries older than a year
		return t.Format("Jan 2, 2006")
	}
}

// FormatDuration formats duration in seconds as a human-readable string
func FormatDuration(seconds int64) string {
	if seconds <= 0 {
		return "0s"
	}
	if seconds < 60 {
		return fmt.Sprintf("%ds", seconds)
	}
	if seconds < 3600 {
		m, s := seconds/60, seconds%60
		if s == 0 {
			return fmt.Sprintf("%dm", m)
		}
		return fmt.Sprintf("%dm %ds", m, s)
	}
	hours := seconds / 3600
	minutes := (seconds % 3600) / 60
	if minutes == 0 {
		return fmt.Sprintf("%dh", hours)
	}
	return fmt.Sprintf("%dh %dm", hours, minutes)
}
