package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type HeaderMode string

const (
	HeaderModeDefault HeaderMode = "default" // Show "TERA"
	HeaderModeText    HeaderMode = "text"    // Custom text
	HeaderModeASCII   HeaderMode = "ascii"   // ASCII art
	HeaderModeNone    HeaderMode = "none"    // No header
)

type HeaderConfig struct {
	// Display mode
	Mode HeaderMode `yaml:"mode"`

	// Simple text mode
	CustomText string `yaml:"custom_text"`

	// User-provided ASCII art
	AsciiArt string `yaml:"ascii_art"`

	// Display settings
	Alignment     string `yaml:"alignment"`
	Width         int    `yaml:"width"`
	Color         string `yaml:"color"`
	Bold          bool   `yaml:"bold"`
	PaddingTop    int    `yaml:"padding_top"`
	PaddingBottom int    `yaml:"padding_bottom"`
}

type AppearanceConfig struct {
	Header HeaderConfig `yaml:"header"`
}

func DefaultAppearanceConfig() AppearanceConfig {
	return AppearanceConfig{
		Header: HeaderConfig{
			Mode:          HeaderModeDefault,
			CustomText:    "",
			AsciiArt:      "",
			Alignment:     "center",
			Width:         50,
			Color:         "auto",
			Bold:          true,
			PaddingTop:    1,
			PaddingBottom: 0,
		},
	}
}

func (c *AppearanceConfig) Validate() error {
	h := &c.Header

	// Validate mode
	switch h.Mode {
	case HeaderModeDefault, HeaderModeText, HeaderModeASCII, HeaderModeNone:
		// Valid
	default:
		h.Mode = HeaderModeDefault
	}

	// Validate custom text length
	if len(h.CustomText) > 100 {
		h.CustomText = h.CustomText[:100]
	}

	// Validate ASCII art line count
	if h.Mode == HeaderModeASCII && h.AsciiArt != "" {
		lines := strings.Split(h.AsciiArt, "\n")
		if len(lines) > 15 {
			return fmt.Errorf("ASCII art exceeds 15 lines")
		}
	}

	// Validate width
	if h.Width < 10 {
		h.Width = 10
	}
	if h.Width > 120 {
		h.Width = 120
	}

	// Validate alignment
	switch h.Alignment {
	case "left", "center", "right":
		// Valid
	default:
		h.Alignment = "center"
	}

	return nil
}

func LoadAppearanceConfig() (AppearanceConfig, error) {
	configPath, err := GetAppearanceConfigPath()
	if err != nil {
		return DefaultAppearanceConfig(), err
	}

	// If file doesn't exist, return defaults (backwards compatible)
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return DefaultAppearanceConfig(), nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return DefaultAppearanceConfig(), fmt.Errorf("failed to read appearance config: %w", err)
	}

	var config AppearanceConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return DefaultAppearanceConfig(), fmt.Errorf("failed to parse appearance config: %w", err)
	}

	if err := config.Validate(); err != nil {
		return DefaultAppearanceConfig(), err
	}

	return config, nil
}

func SaveAppearanceConfig(config AppearanceConfig) error {
	if err := config.Validate(); err != nil {
		return err
	}

	configPath, err := GetAppearanceConfigPath()
	if err != nil {
		return err
	}

	// Ensure config directory exists
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal appearance config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write appearance config: %w", err)
	}

	return nil
}

func GetAppearanceConfigPath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("failed to get config directory: %w", err)
	}

	return filepath.Join(configDir, "tera", "appearance_config.yaml"), nil
}
