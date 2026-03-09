package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const syncPrefsFileName = "sync_prefs.json"

// SyncPrefs holds the user's persisted checklist selections for backup/sync.
// Each field maps to a category of files that can be included in an export or sync.
type SyncPrefs struct {
	Favorites     bool `json:"favorites"`
	Settings      bool `json:"settings"`
	RatingsVotes  bool `json:"ratings_votes"`
	Blocklist     bool `json:"blocklist"`
	MetadataTags  bool `json:"metadata_tags"`
	SearchHistory bool `json:"search_history"`
}

// DefaultSyncPrefs returns sensible defaults.
// Search history is off by default as it is ephemeral and not worth syncing for most users.
func DefaultSyncPrefs() SyncPrefs {
	return SyncPrefs{
		Favorites:     true,
		Settings:      true,
		RatingsVotes:  true,
		Blocklist:     true,
		MetadataTags:  true,
		SearchHistory: false,
	}
}

// getSyncPrefsPath returns the full path to the sync_prefs.json file.
func getSyncPrefsPath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("failed to get config directory: %w", err)
	}
	return filepath.Join(configDir, "tera", syncPrefsFileName), nil
}

// LoadSyncPrefs loads sync preferences from disk.
// Returns defaults on any error so the caller can always proceed.
func LoadSyncPrefs() (SyncPrefs, error) {
	path, err := getSyncPrefsPath()
	if err != nil {
		return DefaultSyncPrefs(), err
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return DefaultSyncPrefs(), nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return DefaultSyncPrefs(), fmt.Errorf("failed to read sync prefs: %w", err)
	}

	if len(data) == 0 {
		return DefaultSyncPrefs(), nil
	}

	var prefs SyncPrefs
	if err := json.Unmarshal(data, &prefs); err != nil {
		return DefaultSyncPrefs(), fmt.Errorf("failed to parse sync prefs: %w", err)
	}

	return prefs, nil
}

// SaveSyncPrefs persists sync preferences to disk using an atomic write.
func SaveSyncPrefs(prefs SyncPrefs) error {
	path, err := getSyncPrefsPath()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := json.MarshalIndent(prefs, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal sync prefs: %w", err)
	}

	return atomicWriteFile(path, data, 0600)
}
