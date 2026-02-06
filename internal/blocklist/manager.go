package blocklist

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/shinokada/tera/internal/api"
)

// Manager handles blocklist operations with thread-safe access
type Manager struct {
	blocklistPath string
	blockedMap    map[string]BlockedStation // UUID -> BlockedStation for fast lookup
	blockRules    []BlockRule               // Active block rules
	mu            sync.RWMutex              // Protects concurrent access
	lastBlock     *BlockedStation           // Last blocked station for undo feature
}

// NewManager creates a new blocklist manager
// blocklistPath should be the full path to blocklist.json
func NewManager(blocklistPath string) *Manager {
	return &Manager{
		blocklistPath: blocklistPath,
		blockedMap:    make(map[string]BlockedStation),
	}
}

// Load reads the blocklist from disk
func (m *Manager) Load(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// If file doesn't exist, start with empty blocklist
	if _, err := os.Stat(m.blocklistPath); os.IsNotExist(err) {
		m.blockedMap = make(map[string]BlockedStation)
		return nil
	}

	data, err := os.ReadFile(m.blocklistPath)
	if err != nil {
		return fmt.Errorf("failed to read blocklist: %w", err)
	}

	var blocklist Blocklist
	if err := json.Unmarshal(data, &blocklist); err != nil {
		return fmt.Errorf("failed to parse blocklist: %w", err)
	}

	// Convert to map for fast lookups
	m.blockedMap = make(map[string]BlockedStation, len(blocklist.BlockedStations))
	for _, station := range blocklist.BlockedStations {
		m.blockedMap[station.StationUUID] = station
	}

	// Load block rules
	if blocklist.BlockRules != nil {
		m.blockRules = blocklist.BlockRules
	} else {
		m.blockRules = []BlockRule{}
	}

	return nil
}

// Save writes the blocklist to disk
func (m *Manager) Save(ctx context.Context) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Ensure directory exists
	dir := filepath.Dir(m.blocklistPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create blocklist directory: %w", err)
	}

	// Convert map to slice
	stations := make([]BlockedStation, 0, len(m.blockedMap))
	for _, station := range m.blockedMap {
		stations = append(stations, station)
	}

	blocklist := Blocklist{
		Version:         "1.0",
		BlockedStations: stations,
		BlockRules:      m.blockRules,
	}

	data, err := json.MarshalIndent(blocklist, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal blocklist: %w", err)
	}

	if err := os.WriteFile(m.blocklistPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write blocklist: %w", err)
	}

	return nil
}

// Block adds a station to the blocklist
// Returns a message (with optional warning) and error
func (m *Manager) Block(ctx context.Context, station *api.Station) (string, error) {
	if station == nil {
		return "", fmt.Errorf("station cannot be nil")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if already blocked
	if _, exists := m.blockedMap[station.StationUUID]; exists {
		return "", ErrStationAlreadyBlocked
	}

	// Create blocked station entry
	blocked := BlockedStation{
		StationUUID: station.StationUUID,
		Name:        station.Name,
		Tags:        station.Tags,
		Country:     station.Country,
		CountryCode: station.CountryCode,
		State:       station.State,
		Language:    station.Language,
		Codec:       station.Codec,
		Bitrate:     station.Bitrate,
		BlockedAt:   time.Now(),
	}

	// Add to map
	m.blockedMap[station.StationUUID] = blocked

	// Save last block for undo
	m.lastBlock = &blocked

	// Save to disk
	if err := m.save(); err != nil {
		// Rollback on error
		delete(m.blockedMap, station.StationUUID)
		m.lastBlock = nil
		return "", err
	}

	// Generate message with optional warning
	count := len(m.blockedMap)
	msg := fmt.Sprintf("🚫 Blocked: %s", station.TrimName())

	switch count {
	case BlockWarningThreshold:
		msg += fmt.Sprintf("\n⚠️ You've blocked %d stations. Consider using Block Rules in future updates.", count)
	case BlockLargeThreshold:
		msg += fmt.Sprintf("\n⚠️ Large blocklist (%d stations). Export recommended.", count)
	}

	return msg, nil
}

// Unblock removes a station from the blocklist by UUID
func (m *Manager) Unblock(ctx context.Context, stationUUID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if blocked
	if _, exists := m.blockedMap[stationUUID]; !exists {
		return ErrStationNotBlocked
	}

	// Remove from map
	delete(m.blockedMap, stationUUID)

	// Save to disk
	return m.save()
}

// IsBlocked checks if a station is blocked
func (m *Manager) IsBlocked(stationUUID string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	_, exists := m.blockedMap[stationUUID]
	return exists
}

// GetAll returns all blocked stations as a slice
// Sorted by blocked_at (most recent first)
func (m *Manager) GetAll() []BlockedStation {
	m.mu.RLock()
	defer m.mu.RUnlock()

	stations := make([]BlockedStation, 0, len(m.blockedMap))
	for _, station := range m.blockedMap {
		stations = append(stations, station)
	}

	// Sort by blocked_at descending (most recent first)
	// Simple bubble sort since we don't expect huge lists
	for i := 0; i < len(stations)-1; i++ {
		for j := i + 1; j < len(stations); j++ {
			if stations[i].BlockedAt.Before(stations[j].BlockedAt) {
				stations[i], stations[j] = stations[j], stations[i]
			}
		}
	}

	return stations
}

// Count returns the number of blocked stations
func (m *Manager) Count() int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return len(m.blockedMap)
}

// Clear removes all blocked stations
func (m *Manager) Clear(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.blockedMap = make(map[string]BlockedStation)
	m.lastBlock = nil

	return m.save()
}

// GetLastBlocked returns the last blocked station (for undo feature)
func (m *Manager) GetLastBlocked() *BlockedStation {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.lastBlock
}

// UndoLastBlock removes the last blocked station if called within time window
// Returns true if undo was successful, false if no recent block to undo
func (m *Manager) UndoLastBlock(ctx context.Context) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.lastBlock == nil {
		return false, nil
	}

	// Remove from map
	delete(m.blockedMap, m.lastBlock.StationUUID)
	m.lastBlock = nil

	// Save to disk
	if err := m.save(); err != nil {
		return false, err
	}

	return true, nil
}

// save is an internal helper that saves without locking (caller must hold lock)
func (m *Manager) save() error {
	// Ensure directory exists
	dir := filepath.Dir(m.blocklistPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create blocklist directory: %w", err)
	}

	// Convert map to slice
	stations := make([]BlockedStation, 0, len(m.blockedMap))
	for _, station := range m.blockedMap {
		stations = append(stations, station)
	}

	blocklist := Blocklist{
		Version:         "1.0",
		BlockedStations: stations,
		BlockRules:      m.blockRules,
	}

	data, err := json.MarshalIndent(blocklist, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal blocklist: %w", err)
	}

	if err := os.WriteFile(m.blocklistPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write blocklist: %w", err)
	}

	return nil
}

// AddBlockRule adds a new blocking rule
func (m *Manager) AddBlockRule(ctx context.Context, ruleType BlockRuleType, value string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if rule already exists
	for _, rule := range m.blockRules {
		if rule.Type == ruleType && strings.EqualFold(rule.Value, value) {
			return fmt.Errorf("rule already exists: %s", rule.String())
		}
	}

	// Add new rule
	newRule := BlockRule{
		Type:  ruleType,
		Value: value,
	}
	m.blockRules = append(m.blockRules, newRule)

	// Save to disk
	return m.save()
}

// RemoveBlockRule removes a blocking rule
func (m *Manager) RemoveBlockRule(ctx context.Context, ruleType BlockRuleType, value string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Find and remove rule
	for i, rule := range m.blockRules {
		if rule.Type == ruleType && strings.EqualFold(rule.Value, value) {
			// Remove rule from slice
			m.blockRules = append(m.blockRules[:i], m.blockRules[i+1:]...)
			// Save to disk
			return m.save()
		}
	}

	return fmt.Errorf("rule not found")
}

// GetBlockRules returns all active block rules
func (m *Manager) GetBlockRules() []BlockRule {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Return a copy to prevent external modification
	rules := make([]BlockRule, len(m.blockRules))
	copy(rules, m.blockRules)
	return rules
}

// IsBlockedByRule checks if a station is blocked by any rule
func (m *Manager) IsBlockedByRule(station *api.Station) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, rule := range m.blockRules {
		if rule.Matches(station) {
			return true
		}
	}
	return false
}

// IsBlockedByAny checks if a station is blocked (either individually or by rule)
func (m *Manager) IsBlockedByAny(station *api.Station) bool {
	if station == nil {
		return false
	}

	// Check individual block first
	if m.IsBlocked(station.StationUUID) {
		return true
	}

	// Check rules
	return m.IsBlockedByRule(station)
}
