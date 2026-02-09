package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// V2Theme represents the old v2 theme structure
type V2Theme struct {
	Colors struct {
		Primary   string `yaml:"primary"`
		Secondary string `yaml:"secondary"`
		Highlight string `yaml:"highlight"`
		Error     string `yaml:"error"`
		Success   string `yaml:"success"`
		Muted     string `yaml:"muted"`
		Text      string `yaml:"text"`
	} `yaml:"colors"`
	Padding struct {
		PageHorizontal int `yaml:"page_horizontal"`
		PageVertical   int `yaml:"page_vertical"`
		ListItemLeft   int `yaml:"list_item_left"`
		BoxHorizontal  int `yaml:"box_horizontal"`
		BoxVertical    int `yaml:"box_vertical"`
	} `yaml:"padding"`
}

// V2AppearanceConfig represents the old v2 appearance structure
type V2AppearanceConfig struct {
	Header struct {
		Mode          string `yaml:"mode"`
		CustomText    string `yaml:"custom_text"`
		AsciiArt      string `yaml:"ascii_art"`
		Alignment     string `yaml:"alignment"`
		Width         int    `yaml:"width"`
		Color         string `yaml:"color"`
		Bold          bool   `yaml:"bold"`
		PaddingTop    int    `yaml:"padding_top"`
		PaddingBottom int    `yaml:"padding_bottom"`
	} `yaml:"header"`
}

// V2ConnectionConfig represents the old v2 connection structure
type V2ConnectionConfig struct {
	AutoReconnect  bool `yaml:"auto_reconnect"`
	ReconnectDelay int  `yaml:"reconnect_delay"`
	StreamBufferMB int  `yaml:"stream_buffer_mb"`
}

// V2ShuffleConfig represents the old v2 shuffle structure
type V2ShuffleConfig struct {
	Shuffle struct {
		AutoAdvance     bool `yaml:"auto_advance"`
		IntervalMinutes int  `yaml:"interval_minutes"`
		RememberHistory bool `yaml:"remember_history"`
		MaxHistory      int  `yaml:"max_history"`
	} `yaml:"shuffle"`
}

// MigrateFromV2 migrates configuration from v2 to v3
func MigrateFromV2(v2ConfigDir string) (*Config, error) {
	cfg := DefaultConfig()

	// Read v2 theme.yaml
	themePath := filepath.Join(v2ConfigDir, "theme.yaml")
	if theme, err := readV2Theme(themePath); err == nil {
		cfg.UI.Theme.Colors = theme
	}

	// Read v2 theme padding
	if padding, err := readV2ThemePadding(themePath); err == nil {
		cfg.UI.Theme.Padding = padding
	}

	// Read v2 appearance_config.yaml
	appearancePath := filepath.Join(v2ConfigDir, "appearance_config.yaml")
	if appearance, err := readV2Appearance(appearancePath); err == nil {
		cfg.UI.Appearance = appearance
	}

	// Read v2 connection_config.yaml
	connectionPath := filepath.Join(v2ConfigDir, "connection_config.yaml")
	if network, err := readV2Connection(connectionPath); err == nil {
		cfg.Network = network
	}

	// Read v2 shuffle.yaml
	shufflePath := filepath.Join(v2ConfigDir, "shuffle.yaml")
	if shuffle, err := readV2Shuffle(shufflePath); err == nil {
		cfg.Shuffle = shuffle
	}

	return &cfg, nil
}

// readV2Theme reads theme colors from v2 theme.yaml
func readV2Theme(path string) (map[string]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var v2Theme V2Theme
	if err := yaml.Unmarshal(data, &v2Theme); err != nil {
		return nil, err
	}

	colors := make(map[string]string)
	colors["primary"] = v2Theme.Colors.Primary
	colors["secondary"] = v2Theme.Colors.Secondary
	colors["highlight"] = v2Theme.Colors.Highlight
	colors["error"] = v2Theme.Colors.Error
	colors["success"] = v2Theme.Colors.Success
	colors["muted"] = v2Theme.Colors.Muted
	colors["text"] = v2Theme.Colors.Text

	return colors, nil
}

// readV2ThemePadding reads padding from v2 theme.yaml
func readV2ThemePadding(path string) (PaddingConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return PaddingConfig{}, err
	}

	var v2Theme V2Theme
	if err := yaml.Unmarshal(data, &v2Theme); err != nil {
		return PaddingConfig{}, err
	}

	padding := PaddingConfig{
		PageHorizontal: v2Theme.Padding.PageHorizontal,
		PageVertical:   v2Theme.Padding.PageVertical,
		ListItemLeft:   v2Theme.Padding.ListItemLeft,
		BoxHorizontal:  v2Theme.Padding.BoxHorizontal,
		BoxVertical:    v2Theme.Padding.BoxVertical,
	}

	return padding, nil
}

// readV2Appearance reads appearance config from v2
func readV2Appearance(path string) (AppearanceConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return AppearanceConfig{}, err
	}

	var v2App V2AppearanceConfig
	if err := yaml.Unmarshal(data, &v2App); err != nil {
		return AppearanceConfig{}, err
	}

	appearance := AppearanceConfig{
		HeaderMode:    v2App.Header.Mode,
		HeaderAlign:   v2App.Header.Alignment,
		HeaderWidth:   v2App.Header.Width,
		CustomText:    v2App.Header.CustomText,
		AsciiArt:      v2App.Header.AsciiArt,
		HeaderColor:   v2App.Header.Color,
		HeaderBold:    v2App.Header.Bold,
		PaddingTop:    v2App.Header.PaddingTop,
		PaddingBottom: v2App.Header.PaddingBottom,
	}

	return appearance, nil
}

// readV2Connection reads connection config from v2
func readV2Connection(path string) (NetworkConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return NetworkConfig{}, err
	}

	var v2Conn V2ConnectionConfig
	if err := yaml.Unmarshal(data, &v2Conn); err != nil {
		return NetworkConfig{}, err
	}

	network := NetworkConfig{
		AutoReconnect:  v2Conn.AutoReconnect,
		ReconnectDelay: v2Conn.ReconnectDelay,
		BufferSizeMB:   v2Conn.StreamBufferMB,
	}

	return network, nil
}

// readV2Shuffle reads shuffle config from v2
func readV2Shuffle(path string) (ShuffleConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return ShuffleConfig{}, err
	}

	var v2Shuffle V2ShuffleConfig
	if err := yaml.Unmarshal(data, &v2Shuffle); err != nil {
		return ShuffleConfig{}, err
	}

	shuffle := ShuffleConfig{
		AutoAdvance:     v2Shuffle.Shuffle.AutoAdvance,
		IntervalMinutes: v2Shuffle.Shuffle.IntervalMinutes,
		RememberHistory: v2Shuffle.Shuffle.RememberHistory,
		MaxHistory:      v2Shuffle.Shuffle.MaxHistory,
	}

	return shuffle, nil
}

// HasV2Config checks if v2 config files exist
func HasV2Config(v2ConfigDir string) bool {
	// Check for at least one v2 config file
	v2Files := []string{
		"theme.yaml",
		"appearance_config.yaml",
		"connection_config.yaml",
		"shuffle.yaml",
	}

	for _, file := range v2Files {
		path := filepath.Join(v2ConfigDir, file)
		if _, err := os.Stat(path); err == nil {
			return true
		}
	}

	return false
}

// BackupV2Configs creates a backup of v2 config files
func BackupV2Configs(v2ConfigDir string) error {
	timestamp := time.Now().Format("20060102-150405")
	backupDir := filepath.Join(v2ConfigDir, fmt.Sprintf(".v2-backup-%s", timestamp))

	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return fmt.Errorf("failed to create backup directory: %w", err)
	}

	v2Files := []string{
		"theme.yaml",
		"appearance_config.yaml",
		"connection_config.yaml",
		"shuffle.yaml",
	}

	backedUp := false
	for _, file := range v2Files {
		srcPath := filepath.Join(v2ConfigDir, file)
		dstPath := filepath.Join(backupDir, file)

		data, err := os.ReadFile(srcPath)
		if err != nil {
			if os.IsNotExist(err) {
				continue // File doesn't exist, skip
			}
			return fmt.Errorf("failed to read %s: %w", file, err)
		}

		if err := os.WriteFile(dstPath, data, 0644); err != nil {
			return fmt.Errorf("failed to backup %s: %w", file, err)
		}
		backedUp = true
	}

	if !backedUp {
		// No files were backed up, remove empty backup directory
		if err := os.Remove(backupDir); err != nil {
			// Ignore error when removing empty backup dir
			_ = err
		}
		return fmt.Errorf("no v2 config files found to backup")
	}

	return nil
}

// RemoveV2Configs removes old v2 config files (after successful migration)
func RemoveV2Configs(v2ConfigDir string) error {
	v2Files := []string{
		"theme.yaml",
		"appearance_config.yaml",
		"connection_config.yaml",
		"shuffle.yaml",
	}

	var errs []string
	for _, file := range v2Files {
		path := filepath.Join(v2ConfigDir, file)
		if err := os.Remove(path); err != nil {
			if !os.IsNotExist(err) {
				errs = append(errs, fmt.Sprintf("%s: %v", file, err))
			}
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("failed to remove some v2 config files: %s", strings.Join(errs, "; "))
	}

	return nil
}

// DetectV2Config returns information about v2 config files
func DetectV2Config(v2ConfigDir string) map[string]bool {
	v2Files := []string{
		"theme.yaml",
		"appearance_config.yaml",
		"connection_config.yaml",
		"shuffle.yaml",
	}

	detected := make(map[string]bool)
	for _, file := range v2Files {
		path := filepath.Join(v2ConfigDir, file)
		_, err := os.Stat(path)
		detected[file] = (err == nil)
	}

	return detected
}
