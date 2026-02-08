package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"

	"gopkg.in/yaml.v3"
)

// GetShuffleConfigPath returns the path to the shuffle config file
func GetShuffleConfigPath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("failed to get config directory: %w", err)
	}

	return filepath.Join(configDir, "tera", "shuffle.yaml"), nil
}

// LoadShuffleConfig loads shuffle configuration from disk
func LoadShuffleConfig() (ShuffleConfig, error) {
	configPath, err := GetShuffleConfigPath()
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

	var wrapper struct {
		Shuffle ShuffleConfig `yaml:"shuffle"`
	}

	if err := yaml.Unmarshal(data, &wrapper); err != nil {
		return DefaultShuffleConfig(), fmt.Errorf("failed to parse shuffle config: %w", err)
	}

	// Validate and fix any invalid values
	config := validateShuffleConfig(wrapper.Shuffle)

	return config, nil
}

// SaveShuffleConfig saves shuffle configuration to disk
func SaveShuffleConfig(config ShuffleConfig) error {
	configPath, err := GetShuffleConfigPath()
	if err != nil {
		return err
	}

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Validate before saving
	config = validateShuffleConfig(config)

	wrapper := struct {
		Shuffle ShuffleConfig `yaml:"shuffle"`
	}{
		Shuffle: config,
	}

	data, err := yaml.Marshal(wrapper)
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
	if !slices.Contains(validIntervals, config.IntervalMinutes) {
		config.IntervalMinutes = 5 // Default
	}

	// Valid history sizes: 3, 5, 7, 10
	validHistorySizes := []int{3, 5, 7, 10}
	if !slices.Contains(validHistorySizes, config.MaxHistory) {
		config.MaxHistory = 5 // Default
	}

	return config
}
