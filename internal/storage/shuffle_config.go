package storage

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// ShuffleConfigPath returns the path to the shuffle config file
func ShuffleConfigPath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("failed to get config directory: %w", err)
	}

	teraDir := filepath.Join(configDir, "tera")
	if err := os.MkdirAll(teraDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create config directory: %w", err)
	}

	return filepath.Join(teraDir, "shuffle.yaml"), nil
}

// LoadShuffleConfig loads shuffle configuration from disk
func LoadShuffleConfig() (ShuffleConfig, error) {
	configPath, err := ShuffleConfigPath()
	if err != nil {
		return DefaultShuffleConfig(), err
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			// File doesn't exist, return defaults
			return DefaultShuffleConfig(), nil
		}
		return DefaultShuffleConfig(), fmt.Errorf("failed to read shuffle config: %w", err)
	}

	var config ShuffleConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return DefaultShuffleConfig(), fmt.Errorf("failed to parse shuffle config: %w", err)
	}

	// Validate and fix any invalid values
	config = validateShuffleConfig(config)

	return config, nil
}

// SaveShuffleConfig saves shuffle configuration to disk
func SaveShuffleConfig(config ShuffleConfig) error {
	configPath, err := ShuffleConfigPath()
	if err != nil {
		return err
	}

	// Validate before saving
	config = validateShuffleConfig(config)

	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal shuffle config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write shuffle config: %w", err)
	}

	return nil
}

// validateShuffleConfig ensures all config values are within valid ranges
func validateShuffleConfig(config ShuffleConfig) ShuffleConfig {
	// Valid interval values: 1, 3, 5, 10, 15
	validIntervals := []int{1, 3, 5, 10, 15}
	if !contains(validIntervals, config.IntervalMinutes) {
		config.IntervalMinutes = 5 // Default
	}

	// Valid history sizes: 3, 5, 7, 10
	validHistorySizes := []int{3, 5, 7, 10}
	if !contains(validHistorySizes, config.MaxHistory) {
		config.MaxHistory = 5 // Default
	}

	return config
}

// contains checks if an int slice contains a value
func contains(slice []int, val int) bool {
	for _, item := range slice {
		if item == val {
			return true
		}
	}
	return false
}
