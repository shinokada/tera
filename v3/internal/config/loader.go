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

	// Parse YAML
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Validate and fix any issues
	if err := cfg.Validate(); err != nil {
		// Log validation errors but continue with corrected values
		fmt.Fprintf(os.Stderr, "Config validation warnings: %v\n", err)
	}

	return &cfg, nil
}

// Save saves the configuration to disk
func Save(cfg *Config) error {
	if cfg == nil {
		return fmt.Errorf("config cannot be nil")
	}

	// Validate before saving
	if err := cfg.Validate(); err != nil {
		// Validation errors have already corrected the values
		// Log warnings but proceed with save
		fmt.Fprintf(os.Stderr, "Config validation warnings: %v\n", err)
	}

	configPath, err := GetConfigPath()
	if err != nil {
		return fmt.Errorf("failed to get config path: %w", err)
	}

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

	// Write to file
	if err := os.WriteFile(configPath, content, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
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
