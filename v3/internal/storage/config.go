package storage

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/shinokada/tera/v3/internal/config"
)

// GetUnifiedConfig loads the unified v3 config
// This is the primary config function for v3
func GetUnifiedConfig() (*config.Config, error) {
	return config.Load()
}

// SaveUnifiedConfig saves the unified v3 config
func SaveUnifiedConfig(cfg *config.Config) error {
	return config.Save(cfg)
}

// Legacy adapter functions for backward compatibility
// These allow existing code to continue working while transitioning to unified config

// LoadAppearanceConfig loads appearance settings from unified config
func LoadAppearanceConfigFromUnified() (AppearanceConfig, error) {
	cfg, err := config.Load()
	if err != nil {
		return DefaultAppearanceConfig(), err
	}

	return AppearanceConfig{
		Header: HeaderConfig{
			Mode:          HeaderMode(cfg.UI.Appearance.HeaderMode),
			CustomText:    cfg.UI.Appearance.CustomText,
			AsciiArt:      cfg.UI.Appearance.AsciiArt,
			Alignment:     cfg.UI.Appearance.HeaderAlign,
			Width:         cfg.UI.Appearance.HeaderWidth,
			Color:         cfg.UI.Appearance.HeaderColor,
			Bold:          cfg.UI.Appearance.HeaderBold,
			PaddingTop:    cfg.UI.Appearance.PaddingTop,
			PaddingBottom: cfg.UI.Appearance.PaddingBottom,
		},
	}, nil
}

// SaveAppearanceConfigToUnified saves appearance settings to unified config
func SaveAppearanceConfigToUnified(appearance AppearanceConfig) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	cfg.UI.Appearance.HeaderMode = string(appearance.Header.Mode)
	cfg.UI.Appearance.CustomText = appearance.Header.CustomText
	cfg.UI.Appearance.AsciiArt = appearance.Header.AsciiArt
	cfg.UI.Appearance.HeaderAlign = appearance.Header.Alignment
	cfg.UI.Appearance.HeaderWidth = appearance.Header.Width
	cfg.UI.Appearance.HeaderColor = appearance.Header.Color
	cfg.UI.Appearance.HeaderBold = appearance.Header.Bold
	cfg.UI.Appearance.PaddingTop = appearance.Header.PaddingTop
	cfg.UI.Appearance.PaddingBottom = appearance.Header.PaddingBottom

	return config.Save(cfg)
}

// LoadConnectionConfigFromUnified loads connection settings from unified config
func LoadConnectionConfigFromUnified() (ConnectionConfig, error) {
	cfg, err := config.Load()
	if err != nil {
		return DefaultConnectionConfig(), err
	}

	return ConnectionConfig{
		AutoReconnect:  cfg.Network.AutoReconnect,
		ReconnectDelay: cfg.Network.ReconnectDelay,
		StreamBufferMB: cfg.Network.BufferSizeMB,
	}, nil
}

// SaveConnectionConfigToUnified saves connection settings to unified config
func SaveConnectionConfigToUnified(conn ConnectionConfig) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	cfg.Network.AutoReconnect = conn.AutoReconnect
	cfg.Network.ReconnectDelay = conn.ReconnectDelay
	cfg.Network.BufferSizeMB = conn.StreamBufferMB

	return config.Save(cfg)
}

// LoadShuffleConfigFromUnified loads shuffle settings from unified config
func LoadShuffleConfigFromUnified() (ShuffleConfig, error) {
	cfg, err := config.Load()
	if err != nil {
		return DefaultShuffleConfig(), err
	}

	return ShuffleConfig{
		AutoAdvance:     cfg.Shuffle.AutoAdvance,
		IntervalMinutes: cfg.Shuffle.IntervalMinutes,
		RememberHistory: cfg.Shuffle.RememberHistory,
		MaxHistory:      cfg.Shuffle.MaxHistory,
	}, nil
}

// SaveShuffleConfigToUnified saves shuffle settings to unified config
func SaveShuffleConfigToUnified(shuffle ShuffleConfig) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	cfg.Shuffle.AutoAdvance = shuffle.AutoAdvance
	cfg.Shuffle.IntervalMinutes = shuffle.IntervalMinutes
	cfg.Shuffle.RememberHistory = shuffle.RememberHistory
	cfg.Shuffle.MaxHistory = shuffle.MaxHistory

	return config.Save(cfg)
}

// CheckAndMigrateV2Config checks for v2 config and migrates if found
// Returns true if migration was performed, false otherwise
func CheckAndMigrateV2Config() (bool, error) {
	// Check if v3 config already exists
	exists, err := config.Exists()
	if err != nil {
		return false, err
	}
	if exists {
		// v3 config exists, no migration needed
		return false, nil
	}

	// Check for v2 config
	configDir, err := os.UserConfigDir()
	if err != nil {
		return false, err
	}
	v2ConfigDir := filepath.Join(configDir, "tera")

	if !config.HasV2Config(v2ConfigDir) {
		// No v2 config found, will use defaults
		return false, nil
	}

	// Migrate from v2 to v3
	cfg, err := config.MigrateFromV2(v2ConfigDir)
	if err != nil {
		return false, fmt.Errorf("migration failed: %w", err)
	}

	// Save migrated config
	if err := config.Save(cfg); err != nil {
		return false, fmt.Errorf("failed to save migrated config: %w", err)
	}

	// Backup v2 configs
	if err := config.BackupV2Configs(v2ConfigDir); err != nil {
		// Log warning but don't fail
		fmt.Fprintf(os.Stderr, "Warning: Could not backup v2 configs: %v\n", err)
	}

	// Remove old v2 config files (optional - can be skipped if backup succeeded)
	if err := config.RemoveV2Configs(v2ConfigDir); err != nil {
		// Log warning but don't fail
		fmt.Fprintf(os.Stderr, "Warning: Could not remove old v2 configs: %v\n", err)
	}

	return true, nil
}
