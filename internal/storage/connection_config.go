package storage

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const connectionConfigFile = "connection_config.yaml"

// GetConnectionConfigPath returns the path to the connection config file
func GetConnectionConfigPath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("failed to get config directory: %w", err)
	}

	return filepath.Join(configDir, "tera", connectionConfigFile), nil
}

// LoadConnectionConfig loads connection configuration from file
func LoadConnectionConfig() (ConnectionConfig, error) {
	configPath, err := GetConnectionConfigPath()
	if err != nil {
		return DefaultConnectionConfig(), err
	}

	// If file doesn't exist, return defaults
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return DefaultConnectionConfig(), nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return DefaultConnectionConfig(), fmt.Errorf("failed to read connection config: %w", err)
	}

	var config ConnectionConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return DefaultConnectionConfig(), fmt.Errorf("failed to parse connection config: %w", err)
	}

	// Validate and apply bounds
	config = validateConnectionConfig(config)

	return config, nil
}

// SaveConnectionConfig saves connection configuration to file
func SaveConnectionConfig(config ConnectionConfig) error {
	configPath, err := GetConnectionConfigPath()
	if err != nil {
		return err
	}

	// Ensure config directory exists
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Validate before saving
	config = validateConnectionConfig(config)

	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal connection config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write connection config: %w", err)
	}

	return nil
}

// validateConnectionConfig ensures all config values are within valid ranges
func validateConnectionConfig(config ConnectionConfig) ConnectionConfig {
	// Reconnect delay: 1-30 seconds
	if config.ReconnectDelay < 1 {
		config.ReconnectDelay = 1
	}
	if config.ReconnectDelay > 30 {
		config.ReconnectDelay = 30
	}

	// Stream buffer: 0 (no buffering) or 10-200 MB
	if config.StreamBufferMB != 0 && config.StreamBufferMB < 10 {
		config.StreamBufferMB = 10
	}
	if config.StreamBufferMB > 200 {
		config.StreamBufferMB = 200
	}

	return config
}
