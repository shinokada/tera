package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// SleepTimerConfig holds the user's persisted sleep timer preferences.
type SleepTimerConfig struct {
	LastDurationMinutes int `json:"last_duration_minutes"`
	Version             int `json:"version"`
}

// sleepTimerFilePath returns the full path to the sleep timer config file.
func sleepTimerFilePath(dataPath string) string {
	return filepath.Join(dataPath, "sleep_timer.json")
}

// LoadSleepTimerConfig reads the persisted sleep timer config from disk.
// If the file does not exist, a zero-value config is returned without error.
func LoadSleepTimerConfig(dataPath string) (*SleepTimerConfig, error) {
	path := sleepTimerFilePath(dataPath)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &SleepTimerConfig{Version: 1}, nil
		}
		return nil, fmt.Errorf("failed to read sleep timer config: %w", err)
	}

	var cfg SleepTimerConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		// Corrupted file — return defaults
		return &SleepTimerConfig{Version: 1}, nil
	}
	return &cfg, nil
}

// SaveSleepTimerConfig writes the config to disk using an atomic rename.
func SaveSleepTimerConfig(dataPath string, cfg *SleepTimerConfig) error {
	if cfg == nil {
		return fmt.Errorf("config must not be nil")
	}
	cfg.Version = 1

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal sleep timer config: %w", err)
	}

	if err := os.MkdirAll(dataPath, 0755); err != nil {
		return fmt.Errorf("failed to create data directory: %w", err)
	}

	path := sleepTimerFilePath(dataPath)
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0644); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("failed to write sleep timer config: %w", err)
	}
	if err := os.Rename(tmp, path); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("failed to save sleep timer config: %w", err)
	}
	return nil
}
