package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/shinokada/tera/v3/internal/api"
)

// StationRating represents a user's rating for a station
type StationRating struct {
	Rating    int       `json:"rating"`     // 1-5 stars
	RatedAt   time.Time `json:"rated_at"`   // First rated
	UpdatedAt time.Time `json:"updated_at"` // Last updated
}

// RatingsCachedStation stores essential station info for display in Top Rated
type RatingsCachedStation struct {
	Name     string `json:"name"`
	URL      string `json:"url"`
	Country  string `json:"country,omitempty"`
	Language string `json:"language,omitempty"`
	Tags     string `json:"tags,omitempty"`
	Codec    string `json:"codec,omitempty"`
	Bitrate  int    `json:"bitrate,omitempty"`
	Votes    int    `json:"votes,omitempty"`
}

// RatingsStore holds all station ratings (no mutex - protected by manager)
type RatingsStore struct {
	Ratings      map[string]*StationRating        `json:"ratings"`
	StationCache map[string]*RatingsCachedStation `json:"station_cache,omitempty"`
	Version      int                              `json:"version"`
}

// StationWithRating combines station info with its rating for display
type StationWithRating struct {
	Station api.Station
	Rating  *StationRating
}

// RatingsManager manages station star ratings
type RatingsManager struct {
	dataPath    string
	store       *RatingsStore
	mu          sync.RWMutex
	saveMu      sync.Mutex // serializes concurrent Save() calls
	savePending atomic.Bool
	ctx         context.Context
	cancel      context.CancelFunc
	wg          sync.WaitGroup
}

// NewRatingsManager creates a new ratings manager.
// The error return is always nil in the current implementation: load failures
// are treated as a non-fatal "start fresh" condition. The signature is kept
// for forward-compatibility in case future callers need to detect init errors.
func NewRatingsManager(dataPath string) (*RatingsManager, error) {
	ctx, cancel := context.WithCancel(context.Background())
	r := &RatingsManager{
		dataPath: dataPath,
		store: &RatingsStore{
			Ratings:      make(map[string]*StationRating),
			StationCache: make(map[string]*RatingsCachedStation),
			Version:      1,
		},
		ctx:    ctx,
		cancel: cancel,
	}

	// Load existing ratings
	if err := r.Load(); err != nil {
		// If file doesn't exist or is corrupted, start fresh
		// Log warning but don't fail
		_ = err // Silent failure - start with empty store
	}

	// Start background save goroutine
	r.wg.Add(1)
	go r.saveLoop()

	return r, nil
}

// getRatingsFilePath returns the full path to the ratings file
func (r *RatingsManager) getRatingsFilePath() string {
	return filepath.Join(r.dataPath, "station_ratings.json")
}

// Load loads ratings from disk
func (r *RatingsManager) Load() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	filePath := r.getRatingsFilePath()
	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			// File doesn't exist yet, use empty store
			return nil
		}
		return fmt.Errorf("failed to read ratings file: %w", err)
	}

	var store RatingsStore
	if err := json.Unmarshal(data, &store); err != nil {
		return fmt.Errorf("failed to parse ratings file: %w", err)
	}

	// Ensure maps are initialized
	if store.Ratings == nil {
		store.Ratings = make(map[string]*StationRating)
	}
	if store.StationCache == nil {
		store.StationCache = make(map[string]*RatingsCachedStation)
	}

	r.store = &store
	return nil
}

// Save saves ratings to disk
func (r *RatingsManager) Save() error {
	r.saveMu.Lock()
	defer r.saveMu.Unlock()

	r.mu.RLock()
	data, err := json.MarshalIndent(r.store, "", "  ")
	r.mu.RUnlock()

	if err != nil {
		return fmt.Errorf("failed to marshal ratings: %w", err)
	}

	filePath := r.getRatingsFilePath()

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		return fmt.Errorf("failed to create ratings directory: %w", err)
	}

	// Write to a temp file then rename for crash-safety: a direct write would
	// leave a truncated/corrupt file if the process dies mid-write.
	tmpPath := filePath + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0644); err != nil {
		_ = os.Remove(tmpPath) // best-effort cleanup of partial write
		return fmt.Errorf("failed to write ratings temp file: %w", err)
	}
	if err := os.Rename(tmpPath, filePath); err != nil {
		_ = os.Remove(tmpPath) // best-effort cleanup
		return fmt.Errorf("failed to rename ratings file: %w", err)
	}

	return nil
}

// SetRating sets the rating for a station (1-5 stars) and caches station info
func (r *RatingsManager) SetRating(station *api.Station, rating int) error {
	if station == nil {
		return fmt.Errorf("station cannot be nil")
	}
	if rating < 1 || rating > 5 {
		return fmt.Errorf("rating must be between 1 and 5, got %d", rating)
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	stationUUID := station.StationUUID
	now := time.Now()
	existing := r.store.Ratings[stationUUID]

	if existing == nil {
		r.store.Ratings[stationUUID] = &StationRating{
			Rating:    rating,
			RatedAt:   now,
			UpdatedAt: now,
		}
	} else {
		existing.Rating = rating
		existing.UpdatedAt = now
	}

	// Cache station info for later display
	r.store.StationCache[stationUUID] = &RatingsCachedStation{
		Name:     station.Name,
		URL:      station.URLResolved,
		Country:  station.Country,
		Language: station.Language,
		Tags:     station.Tags,
		Codec:    station.Codec,
		Bitrate:  station.Bitrate,
		Votes:    station.Votes,
	}

	r.savePending.Store(true)
	return nil
}

// RemoveRating removes the rating for a station
func (r *RatingsManager) RemoveRating(stationUUID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.store.Ratings[stationUUID]; !exists {
		return nil // Nothing to remove
	}

	delete(r.store.Ratings, stationUUID)
	delete(r.store.StationCache, stationUUID)
	r.savePending.Store(true)
	return nil
}

// GetRating returns the rating for a station, or nil if not rated
func (r *RatingsManager) GetRating(stationUUID string) *StationRating {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if rating, exists := r.store.Ratings[stationUUID]; exists {
		// Return a copy to prevent external modification
		ratingCopy := *rating
		return &ratingCopy
	}
	return nil
}

// sortedRatingsLocked collects all rated station entries, sorts them by less, and
// truncates to at most limit results (0 = no limit). Must be called with RLock held.
// Station info is populated from the cache if available.
func (r *RatingsManager) sortedRatingsLocked(less func(a, b StationWithRating) bool, limit int) []StationWithRating {
	result := make([]StationWithRating, 0, len(r.store.Ratings))
	for uuid, rating := range r.store.Ratings {
		ratingCopy := *rating
		station := api.Station{StationUUID: uuid}

		// Populate station info from cache if available
		if cached, ok := r.store.StationCache[uuid]; ok {
			station.Name = cached.Name
			station.URLResolved = cached.URL
			station.Country = cached.Country
			station.Language = cached.Language
			station.Tags = cached.Tags
			station.Codec = cached.Codec
			station.Bitrate = cached.Bitrate
			station.Votes = cached.Votes
		}

		result = append(result, StationWithRating{
			Station: station,
			Rating:  &ratingCopy,
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

// GetTopRated returns stations sorted by rating (highest first)
func (r *RatingsManager) GetTopRated(limit int) []StationWithRating {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.sortedRatingsLocked(func(a, b StationWithRating) bool {
		return a.Rating.Rating > b.Rating.Rating
	}, limit)
}

// GetByMinRating returns stations with at least the specified rating,
// sorted by rating (highest first)
func (r *RatingsManager) GetByMinRating(minRating int) []StationWithRating {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]StationWithRating, 0)
	for uuid, rating := range r.store.Ratings {
		if rating.Rating >= minRating {
			ratingCopy := *rating
			station := api.Station{StationUUID: uuid}

			// Populate station info from cache if available
			if cached, ok := r.store.StationCache[uuid]; ok {
				station.Name = cached.Name
				station.URLResolved = cached.URL
				station.Country = cached.Country
				station.Language = cached.Language
				station.Tags = cached.Tags
				station.Codec = cached.Codec
				station.Bitrate = cached.Bitrate
				station.Votes = cached.Votes
			}

			result = append(result, StationWithRating{
				Station: station,
				Rating:  &ratingCopy,
			})
		}
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].Rating.Rating != result[j].Rating.Rating {
			return result[i].Rating.Rating > result[j].Rating.Rating
		}
		return result[i].Station.StationUUID < result[j].Station.StationUUID
	})
	return result
}

// GetRecentlyRated returns stations sorted by when they were last rated (most recent first)
func (r *RatingsManager) GetRecentlyRated(limit int) []StationWithRating {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.sortedRatingsLocked(func(a, b StationWithRating) bool {
		return a.Rating.UpdatedAt.After(b.Rating.UpdatedAt)
	}, limit)
}

// GetAllRated returns all rated stations sorted alphabetically by UUID
func (r *RatingsManager) GetAllRated() []StationWithRating {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.sortedRatingsLocked(func(a, b StationWithRating) bool {
		return a.Station.StationUUID < b.Station.StationUUID
	}, 0)
}

// GetTotalRated returns the count of rated stations
func (r *RatingsManager) GetTotalRated() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.store.Ratings)
}

// ClearAll removes all ratings (for testing or user request)
func (r *RatingsManager) ClearAll() error {
	r.mu.Lock()
	r.store.Ratings = make(map[string]*StationRating)
	r.store.StationCache = make(map[string]*RatingsCachedStation)
	r.mu.Unlock()

	// Save is called unconditionally, so no need to set savePending here.
	return r.Save()
}

// saveLoop runs in the background and saves periodically when changes are pending.
// The final save on shutdown is handled by Close() after wg.Wait() returns.
func (r *RatingsManager) saveLoop() {
	defer r.wg.Done()

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if r.savePending.CompareAndSwap(true, false) {
				if err := r.Save(); err != nil {
					// Re-set pending so next tick retries the save
					r.savePending.Store(true)
				}
			}
		case <-r.ctx.Done():
			return
		}
	}
}

// Close stops the background save goroutine and saves any pending changes.
// It returns any error from the final disk write so callers can log or handle it.
func (r *RatingsManager) Close() error {
	// Signal background goroutine to stop and wait for it to exit
	r.cancel()
	r.wg.Wait()

	// Perform the final save here (after the goroutine has stopped) so the
	// error is propagated to the caller instead of being silently discarded.
	if r.savePending.Load() {
		return r.Save()
	}
	return nil
}

// RenderStars returns a string of stars for the given rating
// e.g., rating 4 -> "★ ★ ★ ★ ☆"
func RenderStars(rating int, useUnicode bool) string {
	if rating < 0 {
		rating = 0
	}
	if rating > 5 {
		rating = 5
	}

	var filledStar, emptyStar string
	if useUnicode {
		filledStar = "★"
		emptyStar = "☆"
	} else {
		filledStar = "*"
		emptyStar = "-"
	}

	var parts []string
	for i := 0; i < rating; i++ {
		parts = append(parts, filledStar)
	}
	for i := rating; i < 5; i++ {
		parts = append(parts, emptyStar)
	}
	return strings.Join(parts, " ")
}

// RenderStarsCompact returns only filled stars (e.g., "★ ★ ★")
// Returns empty string for unrated (rating 0 or invalid)
func RenderStarsCompact(rating int, useUnicode bool) string {
	if rating < 1 || rating > 5 {
		return ""
	}

	var filledStar string
	if useUnicode {
		filledStar = "★"
	} else {
		filledStar = "*"
	}

	var parts []string
	for i := 0; i < rating; i++ {
		parts = append(parts, filledStar)
	}
	return strings.Join(parts, " ")
}

// FormatRatedAt formats the rated time as a human-readable relative string
func FormatRatedAt(t time.Time) string {
	return FormatLastPlayed(t) // Reuse the same formatting logic
}
