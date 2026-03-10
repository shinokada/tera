package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const (
	configFileName = "config.yaml"
	teraDir        = "tera"
)

// userConfigDirFunc is the function used to get the user config directory
// It can be overridden in tests
var userConfigDirFunc = os.UserConfigDir

// GetConfigDir returns the path to the config directory
func GetConfigDir() (string, error) {
	configDir, err := userConfigDirFunc()
	if err != nil {
		return "", fmt.Errorf("failed to get user config directory: %w", err)
	}
	return filepath.Join(configDir, teraDir), nil
}

// GetConfigPath returns the full path to the config file
func GetConfigPath() (string, error) {
	configDir, err := GetConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, configFileName), nil
}

// Load loads the configuration from disk
// If the config file doesn't exist, it creates a new one with defaults
func Load() (*Config, error) {
	configPath, err := GetConfigPath()
	if err != nil {
		return nil, fmt.Errorf("failed to get config path: %w", err)
	}

	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// Config doesn't exist, create with defaults
		cfg := DefaultConfig()
		if err := Save(&cfg); err != nil {
			return nil, fmt.Errorf("failed to create default config: %w", err)
		}
		return &cfg, nil
	}

	// Read config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse YAML — seed from defaults so fields absent in legacy configs
	// (e.g. play_history introduced in v3.7) get proper default values
	// instead of Go zero-values which Validate() cannot fully recover.
	cfg := DefaultConfig()
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Validate and fix any issues.
	// If validation corrects anything, save the fixed config back to disk
	// so the warning is only ever printed once (on first run after upgrade).
	if err := cfg.Validate(); err != nil {
		fmt.Fprintf(os.Stderr, "Config validation warnings: %v\n", err)
		// Best-effort save — ignore error so a read-only filesystem doesn't
		// prevent startup, but do silence the duplicate save-time warning by
		// calling Save directly rather than going through the public helper.
		_ = saveRaw(configPath, &cfg)
	}

	return &cfg, nil
}

// Save validates cfg, then saves it to disk.
func Save(cfg *Config) error {
	if cfg == nil {
		return fmt.Errorf("config cannot be nil")
	}

	// Validate before saving; corrections are applied in-place.
	// Warnings are only printed here when Save is called directly with an
	// out-of-range value — the Load path handles its own warning.
	if err := cfg.Validate(); err != nil {
		fmt.Fprintf(os.Stderr, "Config validation warnings: %v\n", err)
	}

	configPath, err := GetConfigPath()
	if err != nil {
		return fmt.Errorf("failed to get config path: %w", err)
	}
	return saveRaw(configPath, cfg)
}

// saveRaw writes cfg to configPath without re-running Validate.
// Used by Load to persist corrected values without double-printing warnings.
func saveRaw(configPath string, cfg *Config) error {
	// Ensure config directory exists
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Marshal to YAML
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Add header comment
	header := generateConfigHeader()
	content := []byte(header + string(data))

	// Write atomically via a temp file in the same directory, then rename.
	// This prevents a crash or short write from leaving a malformed config.yaml
	// that would break the next launch.
	tmpPath := configPath + ".tmp"
	if err := os.WriteFile(tmpPath, content, 0644); err != nil {
		_ = os.Remove(tmpPath) // best-effort cleanup
		return fmt.Errorf("failed to write config temp file: %w", err)
	}
	if err := os.Rename(tmpPath, configPath); err != nil {
		_ = os.Remove(tmpPath) // best-effort cleanup
		return fmt.Errorf("failed to rename config file: %w", err)
	}
	return nil
}

// generateConfigHeader creates a helpful header comment for the config file
func generateConfigHeader() string {
	return `# TERA v3 Configuration File
# This file contains all application settings in one place
#
# To reset to defaults, delete this file and restart TERA
#
# Documentation: https://github.com/shinokada/tera
#

`
}

// Exists checks if the config file exists
func Exists() (bool, error) {
	configPath, err := GetConfigPath()
	if err != nil {
		return false, err
	}

	_, err = os.Stat(configPath)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// Reset resets the configuration to default values
func Reset() error {
	cfg := DefaultConfig()
	return Save(&cfg)
}

// Backup creates a backup of the current config file
func Backup(suffix string) error {
	configPath, err := GetConfigPath()
	if err != nil {
		return fmt.Errorf("failed to get config path: %w", err)
	}

	// Check if config exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return fmt.Errorf("config file does not exist")
	}

	// Read current config
	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	// Create backup path
	backupPath := configPath + "." + suffix
	if err := os.WriteFile(backupPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write backup file: %w", err)
	}

	return nil
}
