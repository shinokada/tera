package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/shinokada/tera/v3/internal/api"
)

// StationTags represents tags for a single station.
type StationTags struct {
	Tags      []string  `json:"tags"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TagPlaylist represents a dynamic tag-based playlist.
type TagPlaylist struct {
	Tags      []string  `json:"tags"`
	MatchMode string    `json:"match_mode"` // "any" or "all"
	CreatedAt time.Time `json:"created_at"`
}

// TagsStore holds all station tags and playlists (no mutex — protected by manager).
type TagsStore struct {
	StationTags  map[string]*StationTags `json:"station_tags"`
	AllTags      []string                `json:"all_tags"`
	TagPlaylists map[string]*TagPlaylist `json:"tag_playlists"`
	Version      int                     `json:"version"`
}

// StationWithTags combines station data with its custom tags for display.
type StationWithTags struct {
	Station api.Station
	Tags    *StationTags
}

// TagsManager manages custom station tags with thread-safe access and debounced saves.
type TagsManager struct {
	dataPath    string
	store       *TagsStore
	mu          sync.RWMutex
	savePending atomic.Bool
	lastSave    time.Time
}

var tagRegex = regexp.MustCompile(`^[a-z0-9][a-z0-9\-_ ]*[a-z0-9]$|^[a-z0-9]$`)

// NewTagsManager creates a new TagsManager and starts the background save goroutine.
func NewTagsManager(dataPath string) (*TagsManager, error) {
	tm := &TagsManager{
		dataPath: dataPath,
		store: &TagsStore{
			StationTags:  make(map[string]*StationTags),
			AllTags:      []string{},
			TagPlaylists: make(map[string]*TagPlaylist),
			Version:      1,
		},
	}

	if err := tm.Load(); err != nil && !os.IsNotExist(err) {
		// Corrupted file — log warning and continue with empty store.
		_ = err
	}

	go tm.saveLoop(context.Background())
	return tm, nil
}

// Load reads tags from disk.
func (t *TagsManager) Load() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	filePath := filepath.Join(t.dataPath, "station_tags.json")
	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, t.store)
}

// Save writes tags to disk atomically via a temp file.
func (t *TagsManager) Save() error {
	t.mu.RLock()
	data, err := json.MarshalIndent(t.store, "", "  ")
	t.mu.RUnlock()
	if err != nil {
		return err
	}

	filePath := filepath.Join(t.dataPath, "station_tags.json")
	if err := os.MkdirAll(t.dataPath, 0755); err != nil {
		return err
	}

	tmpPath := filePath + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0644); err != nil {
		return err
	}
	return os.Rename(tmpPath, filePath)
}

// saveLoop saves pending changes every 5 seconds.
func (t *TagsManager) saveLoop(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			if t.savePending.Load() {
				_ = t.Save()
			}
			return
		case <-ticker.C:
			if t.savePending.Load() {
				if err := t.Save(); err == nil {
					t.savePending.Store(false)
					t.lastSave = time.Now()
				}
			}
		}
	}
}

// normalizeTag lowercases, trims, and validates a tag string.
func normalizeTag(tag string) (string, error) {
	tag = strings.TrimSpace(strings.ToLower(tag))
	if tag == "" {
		return "", fmt.Errorf("tag cannot be empty")
	}
	if len(tag) > 50 {
		return "", fmt.Errorf("tag cannot exceed 50 characters")
	}
	if !tagRegex.MatchString(tag) {
		return "", fmt.Errorf("tag contains invalid characters")
	}
	return tag, nil
}

// AddTag adds a tag to a station (idempotent — no error if already present).
func (t *TagsManager) AddTag(stationUUID string, tag string) error {
	normalized, err := normalizeTag(tag)
	if err != nil {
		return err
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	now := time.Now()
	existing := t.store.StationTags[stationUUID]

	if existing == nil {
		t.store.StationTags[stationUUID] = &StationTags{
			Tags:      []string{normalized},
			CreatedAt: now,
			UpdatedAt: now,
		}
	} else {
		for _, existingTag := range existing.Tags {
			if existingTag == normalized {
				return nil
			}
		}
		if len(existing.Tags) >= 20 {
			return fmt.Errorf("station cannot have more than 20 tags")
		}
		existing.Tags = append(existing.Tags, normalized)
		existing.UpdatedAt = now
	}

	t.addToAllTags(normalized)
	t.savePending.Store(true)
	return nil
}

// RemoveTag removes a tag from a station.
func (t *TagsManager) RemoveTag(stationUUID string, tag string) error {
	normalized, err := normalizeTag(tag)
	if err != nil {
		return err
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	existing := t.store.StationTags[stationUUID]
	if existing == nil {
		return nil
	}

	newTags := make([]string, 0, len(existing.Tags))
	found := false
	for _, tg := range existing.Tags {
		if tg != normalized {
			newTags = append(newTags, tg)
		} else {
			found = true
		}
	}

	if found {
		existing.Tags = newTags
		existing.UpdatedAt = time.Now()
		t.savePending.Store(true)
	}
	return nil
}

// SetTags replaces all tags for a station.
func (t *TagsManager) SetTags(stationUUID string, tags []string) error {
	var normalized []string
	for _, tag := range tags {
		n, err := normalizeTag(tag)
		if err != nil {
			return err
		}
		isDup := false
		for _, existing := range normalized {
			if existing == n {
				isDup = true
				break
			}
		}
		if !isDup {
			normalized = append(normalized, n)
		}
	}
	if len(normalized) > 20 {
		return fmt.Errorf("station cannot have more than 20 tags")
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	now := time.Now()
	existing := t.store.StationTags[stationUUID]
	if existing == nil {
		t.store.StationTags[stationUUID] = &StationTags{
			Tags:      normalized,
			CreatedAt: now,
			UpdatedAt: now,
		}
	} else {
		existing.Tags = normalized
		existing.UpdatedAt = now
	}

	for _, tag := range normalized {
		t.addToAllTags(tag)
	}
	t.savePending.Store(true)
	return nil
}

// ClearTags removes all tags from a station.
func (t *TagsManager) ClearTags(stationUUID string) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if _, ok := t.store.StationTags[stationUUID]; ok {
		delete(t.store.StationTags, stationUUID)
		t.savePending.Store(true)
	}
	return nil
}

// GetTags returns a copy of the tags for a station (empty slice if none).
func (t *TagsManager) GetTags(stationUUID string) []string {
	t.mu.RLock()
	defer t.mu.RUnlock()

	if st, ok := t.store.StationTags[stationUUID]; ok {
		tags := make([]string, len(st.Tags))
		copy(tags, st.Tags)
		return tags
	}
	return []string{}
}

// GetAllTags returns a sorted copy of all unique tags.
func (t *TagsManager) GetAllTags() []string {
	t.mu.RLock()
	defer t.mu.RUnlock()

	tags := make([]string, len(t.store.AllTags))
	copy(tags, t.store.AllTags)
	return tags
}

// GetStationsByTag returns all station UUIDs tagged with the given tag.
func (t *TagsManager) GetStationsByTag(tag string) []string {
	normalized, _ := normalizeTag(tag)
	if normalized == "" {
		return nil
	}

	t.mu.RLock()
	defer t.mu.RUnlock()

	var uuids []string
	for uuid, st := range t.store.StationTags {
		for _, tg := range st.Tags {
			if tg == normalized {
				uuids = append(uuids, uuid)
				break
			}
		}
	}
	return uuids
}

// GetStationsByTags returns station UUIDs matching given tags.
// If matchAll is true, the station must have ALL tags; otherwise ANY tag.
func (t *TagsManager) GetStationsByTags(tags []string, matchAll bool) []string {
	if len(tags) == 0 {
		return nil
	}

	normalizedTags := make([]string, 0, len(tags))
	for _, tag := range tags {
		if n, err := normalizeTag(tag); err == nil {
			normalizedTags = append(normalizedTags, n)
		}
	}
	if len(normalizedTags) == 0 {
		return nil
	}

	t.mu.RLock()
	defer t.mu.RUnlock()

	var uuids []string
	for uuid, st := range t.store.StationTags {
		matches := 0
		for _, targetTag := range normalizedTags {
			for _, stationTag := range st.Tags {
				if stationTag == targetTag {
					matches++
					break
				}
			}
		}
		if matchAll {
			if matches == len(normalizedTags) {
				uuids = append(uuids, uuid)
			}
		} else {
			if matches > 0 {
				uuids = append(uuids, uuid)
			}
		}
	}
	return uuids
}

// GetTaggedStations returns all station UUIDs that have at least one tag.
func (t *TagsManager) GetTaggedStations() []string {
	t.mu.RLock()
	defer t.mu.RUnlock()

	uuids := make([]string, 0, len(t.store.StationTags))
	for uuid, st := range t.store.StationTags {
		if len(st.Tags) > 0 {
			uuids = append(uuids, uuid)
		}
	}
	return uuids
}

// CreatePlaylist creates a new named tag-based playlist.
func (t *TagsManager) CreatePlaylist(name string, tags []string, matchMode string) error {
	if name == "" {
		return fmt.Errorf("playlist name cannot be empty")
	}
	if len(tags) == 0 {
		return fmt.Errorf("playlist must have at least one tag")
	}
	if matchMode != "any" && matchMode != "all" {
		return fmt.Errorf("match mode must be 'any' or 'all'")
	}

	normalizedTags := make([]string, 0, len(tags))
	for _, tag := range tags {
		n, err := normalizeTag(tag)
		if err != nil {
			return err
		}
		normalizedTags = append(normalizedTags, n)
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	if _, exists := t.store.TagPlaylists[name]; exists {
		return fmt.Errorf("playlist '%s' already exists", name)
	}

	t.store.TagPlaylists[name] = &TagPlaylist{
		Tags:      normalizedTags,
		MatchMode: matchMode,
		CreatedAt: time.Now(),
	}
	t.savePending.Store(true)
	return nil
}

// DeletePlaylist removes a playlist by name.
func (t *TagsManager) DeletePlaylist(name string) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if _, exists := t.store.TagPlaylists[name]; !exists {
		return fmt.Errorf("playlist '%s' not found", name)
	}

	delete(t.store.TagPlaylists, name)
	t.savePending.Store(true)
	return nil
}

// GetPlaylist returns a copy of a playlist by name.
func (t *TagsManager) GetPlaylist(name string) *TagPlaylist {
	t.mu.RLock()
	defer t.mu.RUnlock()

	if p, ok := t.store.TagPlaylists[name]; ok {
		return &TagPlaylist{
			Tags:      append([]string(nil), p.Tags...),
			MatchMode: p.MatchMode,
			CreatedAt: p.CreatedAt,
		}
	}
	return nil
}

// GetAllPlaylists returns copies of all playlists.
func (t *TagsManager) GetAllPlaylists() map[string]*TagPlaylist {
	t.mu.RLock()
	defer t.mu.RUnlock()

	result := make(map[string]*TagPlaylist)
	for name, p := range t.store.TagPlaylists {
		result[name] = &TagPlaylist{
			Tags:      append([]string(nil), p.Tags...),
			MatchMode: p.MatchMode,
			CreatedAt: p.CreatedAt,
		}
	}
	return result
}

// GetPlaylistStations returns station UUIDs matching a playlist's tag criteria.
func (t *TagsManager) GetPlaylistStations(name string) []string {
	t.mu.RLock()
	p, ok := t.store.TagPlaylists[name]
	if !ok {
		t.mu.RUnlock()
		return nil
	}
	tags := append([]string(nil), p.Tags...)
	matchMode := p.MatchMode
	t.mu.RUnlock()

	return t.GetStationsByTags(tags, matchMode == "all")
}

// addToAllTags appends a tag to AllTags if not already present and re-sorts.
// Caller must hold t.mu (write lock).
func (t *TagsManager) addToAllTags(tag string) {
	for _, tg := range t.store.AllTags {
		if tg == tag {
			return
		}
	}
	t.store.AllTags = append(t.store.AllTags, tag)
	slices.Sort(t.store.AllTags)
}
