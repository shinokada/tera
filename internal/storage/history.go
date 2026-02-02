package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// SearchHistoryItem represents a single search history entry
type SearchHistoryItem struct {
	SearchType string    `json:"search_type"` // "tag", "name", "language", "country", "state", "advanced"
	Query      string    `json:"query"`
	Timestamp  time.Time `json:"timestamp"`
}

// SearchHistoryStore represents the search history storage
type SearchHistoryStore struct {
	MaxSize      int                 `json:"max_size"`
	SearchItems  []SearchHistoryItem `json:"search_items"`  // For Search Stations
	LuckyQueries []string            `json:"lucky_queries"` // For I Feel Lucky
	LastUpdated  time.Time           `json:"last_updated"`
}

// DefaultMaxHistorySize is the default number of history items to keep
const DefaultMaxHistorySize = 10

// NewSearchHistoryStore creates a new search history store with defaults
func NewSearchHistoryStore() *SearchHistoryStore {
	return &SearchHistoryStore{
		MaxSize:      DefaultMaxHistorySize,
		SearchItems:  []SearchHistoryItem{},
		LuckyQueries: []string{},
		LastUpdated:  time.Now(),
	}
}

// LoadSearchHistory loads the search history from disk
func (s *Storage) LoadSearchHistory(ctx context.Context) (*SearchHistoryStore, error) {
	historyPath := filepath.Join(s.favoritePath, "search-history.json")

	// If file doesn't exist, return new store
	if _, err := os.Stat(historyPath); os.IsNotExist(err) {
		return NewSearchHistoryStore(), nil
	}

	data, err := os.ReadFile(historyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read search history: %w", err)
	}

	var store SearchHistoryStore
	if err := json.Unmarshal(data, &store); err != nil {
		return nil, fmt.Errorf("failed to parse search history: %w", err)
	}

	// Ensure slices are initialized
	if store.SearchItems == nil {
		store.SearchItems = []SearchHistoryItem{}
	}
	if store.LuckyQueries == nil {
		store.LuckyQueries = []string{}
	}

	// Ensure max size is valid
	if store.MaxSize <= 0 {
		store.MaxSize = DefaultMaxHistorySize
	}

	return &store, nil
}

// SaveSearchHistory saves the search history to disk
func (s *Storage) SaveSearchHistory(ctx context.Context, store *SearchHistoryStore) error {
	historyPath := filepath.Join(s.favoritePath, "search-history.json")

	store.LastUpdated = time.Now()

	data, err := json.MarshalIndent(store, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal search history: %w", err)
	}

	if err := os.WriteFile(historyPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write search history: %w", err)
	}

	return nil
}

// AddSearchItem adds a search item to history (for Search Stations)
// If the same type+query exists, it moves to top
// If history is full, removes oldest
func (s *Storage) AddSearchItem(ctx context.Context, searchType, query string) error {
	store, err := s.LoadSearchHistory(ctx)
	if err != nil {
		return err
	}

	// Create new item
	newItem := SearchHistoryItem{
		SearchType: searchType,
		Query:      query,
		Timestamp:  time.Now(),
	}

	// Check if already exists (same type and query)
	existingIndex := -1
	for i, item := range store.SearchItems {
		if item.SearchType == searchType && item.Query == query {
			existingIndex = i
			break
		}
	}

	// If exists, remove it (we'll add to front)
	if existingIndex >= 0 {
		store.SearchItems = append(store.SearchItems[:existingIndex], store.SearchItems[existingIndex+1:]...)
	}

	// Add to front
	store.SearchItems = append([]SearchHistoryItem{newItem}, store.SearchItems...)

	// Trim to max size
	if len(store.SearchItems) > store.MaxSize {
		store.SearchItems = store.SearchItems[:store.MaxSize]
	}

	return s.SaveSearchHistory(ctx, store)
}

// AddLuckyQuery adds a query to lucky history
// If the query exists, it moves to top
// If history is full, removes oldest
func (s *Storage) AddLuckyQuery(ctx context.Context, query string) error {
	store, err := s.LoadSearchHistory(ctx)
	if err != nil {
		return err
	}

	// Check if already exists
	existingIndex := -1
	for i, q := range store.LuckyQueries {
		if q == query {
			existingIndex = i
			break
		}
	}

	// If exists, remove it (we'll add to front)
	if existingIndex >= 0 {
		store.LuckyQueries = append(store.LuckyQueries[:existingIndex], store.LuckyQueries[existingIndex+1:]...)
	}

	// Add to front
	store.LuckyQueries = append([]string{query}, store.LuckyQueries...)

	// Trim to max size
	if len(store.LuckyQueries) > store.MaxSize {
		store.LuckyQueries = store.LuckyQueries[:store.MaxSize]
	}

	return s.SaveSearchHistory(ctx, store)
}

// UpdateHistorySize updates the max history size
// If new size is smaller, trims excess from end
func (s *Storage) UpdateHistorySize(ctx context.Context, newSize int) error {
	if newSize <= 0 {
		return fmt.Errorf("history size must be positive")
	}

	store, err := s.LoadSearchHistory(ctx)
	if err != nil {
		return err
	}

	store.MaxSize = newSize

	// Trim if necessary
	if len(store.SearchItems) > newSize {
		store.SearchItems = store.SearchItems[:newSize]
	}
	if len(store.LuckyQueries) > newSize {
		store.LuckyQueries = store.LuckyQueries[:newSize]
	}

	return s.SaveSearchHistory(ctx, store)
}

// ClearSearchHistory clears all history but keeps the max size setting
func (s *Storage) ClearSearchHistory(ctx context.Context) error {
	store, err := s.LoadSearchHistory(ctx)
	if err != nil {
		return err
	}

	store.SearchItems = []SearchHistoryItem{}
	store.LuckyQueries = []string{}

	return s.SaveSearchHistory(ctx, store)
}
