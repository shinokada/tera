package config

import (
	"errors"
	"fmt"
	"strings"
)

const (
	// ConfigVersion is the current config schema version
	ConfigVersion = "3.0"
)

// Config represents the unified application configuration
type Config struct {
	Version string        `yaml:"version"`
	Player  PlayerConfig  `yaml:"player"`
	UI      UIConfig      `yaml:"ui"`
	Network NetworkConfig `yaml:"network"`
	Shuffle ShuffleConfig `yaml:"shuffle"`
}

// PlayerConfig represents player settings
type PlayerConfig struct {
	DefaultVolume int `yaml:"default_volume"` // 0-100
	BufferSizeMB  int `yaml:"buffer_size_mb"` // Buffer size in megabytes
}

// UIConfig represents user interface settings
type UIConfig struct {
	Theme       ThemeConfig      `yaml:"theme"`
	Appearance  AppearanceConfig `yaml:"appearance"`
	DefaultList string           `yaml:"default_list"` // Default favorites list to load
}

// ThemeConfig represents theme/color settings
type ThemeConfig struct {
	Name    string            `yaml:"name"`
	Colors  map[string]string `yaml:"colors"`
	Padding PaddingConfig     `yaml:"padding"`
}

// PaddingConfig represents padding settings
type PaddingConfig struct {
	PageHorizontal int `yaml:"page_horizontal"`
	PageVertical   int `yaml:"page_vertical"`
	ListItemLeft   int `yaml:"list_item_left"`
	BoxHorizontal  int `yaml:"box_horizontal"`
	BoxVertical    int `yaml:"box_vertical"`
}

// AppearanceConfig represents appearance settings (header, layout)
type AppearanceConfig struct {
	HeaderMode    string `yaml:"header_mode"`    // text, ascii, default, none
	HeaderAlign   string `yaml:"header_align"`   // left, center, right
	HeaderWidth   int    `yaml:"header_width"`   // Width constraint
	CustomText    string `yaml:"custom_text"`    // For text mode
	AsciiArt      string `yaml:"ascii_art"`      // For ascii mode
	HeaderColor   string `yaml:"header_color"`   // Color for header
	HeaderBold    bool   `yaml:"header_bold"`    // Bold header text
	PaddingTop    int    `yaml:"padding_top"`    // Padding above header
	PaddingBottom int    `yaml:"padding_bottom"` // Padding below header
}

// NetworkConfig represents network/connection settings
type NetworkConfig struct {
	AutoReconnect  bool `yaml:"auto_reconnect"`
	ReconnectDelay int  `yaml:"reconnect_delay"` // Seconds between reconnect attempts
	BufferSizeMB   int  `yaml:"buffer_size_mb"`  // Stream buffer size
}

// ShuffleConfig represents shuffle mode settings
type ShuffleConfig struct {
	AutoAdvance     bool `yaml:"auto_advance"`
	IntervalMinutes int  `yaml:"interval_minutes"` // Minutes between auto-advance
	RememberHistory bool `yaml:"remember_history"`
	MaxHistory      int  `yaml:"max_history"` // Number of stations to remember
}

// DefaultConfig returns a new Config with sensible defaults
func DefaultConfig() Config {
	return Config{
		Version: ConfigVersion,
		Player: PlayerConfig{
			DefaultVolume: 100,
			BufferSizeMB:  50,
		},
		UI: UIConfig{
			Theme: ThemeConfig{
				Name: "default",
				Colors: map[string]string{
					"primary":   "6",  // Cyan
					"secondary": "12", // Bright Blue
					"highlight": "3",  // Yellow
					"error":     "9",  // Bright Red
					"success":   "2",  // Green
					"muted":     "8",  // Bright Black (Gray)
					"text":      "7",  // White
				},
				Padding: PaddingConfig{
					PageHorizontal: 2,
					PageVertical:   1,
					ListItemLeft:   2,
					BoxHorizontal:  2,
					BoxVertical:    1,
				},
			},
			Appearance: AppearanceConfig{
				HeaderMode:    "default",
				HeaderAlign:   "center",
				HeaderWidth:   50,
				CustomText:    "",
				AsciiArt:      "",
				HeaderColor:   "auto",
				HeaderBold:    true,
				PaddingTop:    1,
				PaddingBottom: 0,
			},
			DefaultList: "My-favorites",
		},
		Network: NetworkConfig{
			AutoReconnect:  true,
			ReconnectDelay: 5,
			BufferSizeMB:   50,
		},
		Shuffle: ShuffleConfig{
			AutoAdvance:     false,
			IntervalMinutes: 5,
			RememberHistory: true,
			MaxHistory:      5,
		},
	}
}

// Validate validates the configuration and returns any errors
func (c *Config) Validate() error {
	var errs []string

	// Validate version
	if c.Version == "" {
		c.Version = ConfigVersion
	}

	// Validate Player config
	if err := c.Player.Validate(); err != nil {
		errs = append(errs, fmt.Sprintf("player: %v", err))
	}

	// Validate UI config
	if err := c.UI.Validate(); err != nil {
		errs = append(errs, fmt.Sprintf("ui: %v", err))
	}

	// Validate Network config
	if err := c.Network.Validate(); err != nil {
		errs = append(errs, fmt.Sprintf("network: %v", err))
	}

	// Validate Shuffle config
	if err := c.Shuffle.Validate(); err != nil {
		errs = append(errs, fmt.Sprintf("shuffle: %v", err))
	}

	if len(errs) > 0 {
		return errors.New(strings.Join(errs, "; "))
	}

	return nil
}

// Validate validates PlayerConfig
func (p *PlayerConfig) Validate() error {
	var errs []string

	// Validate volume (0-100)
	if p.DefaultVolume < 0 {
		p.DefaultVolume = 0
		errs = append(errs, "default_volume must be >= 0, set to 0")
	}
	if p.DefaultVolume > 100 {
		p.DefaultVolume = 100
		errs = append(errs, "default_volume must be <= 100, set to 100")
	}

	// Validate buffer size (0 or 10-200 MB)
	if p.BufferSizeMB != 0 {
		if p.BufferSizeMB < 10 {
			p.BufferSizeMB = 10
			errs = append(errs, "buffer_size_mb must be >= 10 or 0, set to 10")
		}
		if p.BufferSizeMB > 200 {
			p.BufferSizeMB = 200
			errs = append(errs, "buffer_size_mb must be <= 200, set to 200")
		}
	}

	if len(errs) > 0 {
		return errors.New(strings.Join(errs, "; "))
	}
	return nil
}

// Validate validates UIConfig
func (u *UIConfig) Validate() error {
	var errs []string

	// Validate theme
	if err := u.Theme.Validate(); err != nil {
		errs = append(errs, fmt.Sprintf("theme: %v", err))
	}

	// Validate appearance
	if err := u.Appearance.Validate(); err != nil {
		errs = append(errs, fmt.Sprintf("appearance: %v", err))
	}

	if len(errs) > 0 {
		return errors.New(strings.Join(errs, "; "))
	}
	return nil
}

// Validate validates ThemeConfig
func (t *ThemeConfig) Validate() error {
	var errs []string

	// Ensure theme name is not empty
	if t.Name == "" {
		t.Name = "default"
		errs = append(errs, "name cannot be empty, set to 'default'")
	}

	// Ensure required colors exist
	requiredColors := []string{"primary", "secondary", "highlight", "error", "success", "muted", "text"}
	if t.Colors == nil {
		t.Colors = make(map[string]string)
	}
	defaults := DefaultConfig().UI.Theme.Colors
	for _, key := range requiredColors {
		if _, exists := t.Colors[key]; !exists {
			t.Colors[key] = defaults[key]
			errs = append(errs, fmt.Sprintf("missing color '%s', set to default", key))
		}
	}

	// Validate padding
	if err := t.Padding.Validate(); err != nil {
		errs = append(errs, fmt.Sprintf("padding: %v", err))
	}

	if len(errs) > 0 {
		return errors.New(strings.Join(errs, "; "))
	}
	return nil
}

// Validate validates PaddingConfig
func (p *PaddingConfig) Validate() error {
	var errs []string

	// All padding values should be non-negative
	if p.PageHorizontal < 0 {
		p.PageHorizontal = 0
		errs = append(errs, "page_horizontal must be >= 0, set to 0")
	}
	if p.PageVertical < 0 {
		p.PageVertical = 0
		errs = append(errs, "page_vertical must be >= 0, set to 0")
	}
	if p.ListItemLeft < 0 {
		p.ListItemLeft = 0
		errs = append(errs, "list_item_left must be >= 0, set to 0")
	}
	if p.BoxHorizontal < 0 {
		p.BoxHorizontal = 0
		errs = append(errs, "box_horizontal must be >= 0, set to 0")
	}
	if p.BoxVertical < 0 {
		p.BoxVertical = 0
		errs = append(errs, "box_vertical must be >= 0, set to 0")
	}

	// Reasonable upper bounds
	maxPadding := 20
	if p.PageHorizontal > maxPadding {
		p.PageHorizontal = maxPadding
		errs = append(errs, fmt.Sprintf("page_horizontal exceeds max (%d), set to %d", maxPadding, maxPadding))
	}
	if p.PageVertical > maxPadding {
		p.PageVertical = maxPadding
		errs = append(errs, fmt.Sprintf("page_vertical exceeds max (%d), set to %d", maxPadding, maxPadding))
	}
	if p.ListItemLeft > maxPadding {
		p.ListItemLeft = maxPadding
		errs = append(errs, fmt.Sprintf("list_item_left exceeds max (%d), set to %d", maxPadding, maxPadding))
	}
	if p.BoxHorizontal > maxPadding {
		p.BoxHorizontal = maxPadding
		errs = append(errs, fmt.Sprintf("box_horizontal exceeds max (%d), set to %d", maxPadding, maxPadding))
	}
	if p.BoxVertical > maxPadding {
		p.BoxVertical = maxPadding
		errs = append(errs, fmt.Sprintf("box_vertical exceeds max (%d), set to %d", maxPadding, maxPadding))
	}

	if len(errs) > 0 {
		return errors.New(strings.Join(errs, "; "))
	}
	return nil
}

// Validate validates AppearanceConfig
func (a *AppearanceConfig) Validate() error {
	var errs []string

	// Validate header mode
	validModes := map[string]bool{"default": true, "text": true, "ascii": true, "none": true}
	if !validModes[a.HeaderMode] {
		a.HeaderMode = "default"
		errs = append(errs, "invalid header_mode, set to 'default'")
	}

	// Validate header alignment
	validAligns := map[string]bool{"left": true, "center": true, "right": true}
	if !validAligns[a.HeaderAlign] {
		a.HeaderAlign = "center"
		errs = append(errs, "invalid header_align, set to 'center'")
	}

	// Validate header width
	if a.HeaderWidth < 10 {
		a.HeaderWidth = 10
		errs = append(errs, "header_width must be >= 10, set to 10")
	}
	if a.HeaderWidth > 120 {
		a.HeaderWidth = 120
		errs = append(errs, "header_width must be <= 120, set to 120")
	}

	// Validate custom text length
	if len(a.CustomText) > 100 {
		a.CustomText = a.CustomText[:100]
		errs = append(errs, "custom_text truncated to 100 characters")
	}

	// Validate ASCII art line count
	if a.HeaderMode == "ascii" && a.AsciiArt != "" {
		lines := strings.Split(a.AsciiArt, "\n")
		if len(lines) > 15 {
			a.AsciiArt = strings.Join(lines[:15], "\n")
			errs = append(errs, "ascii_art truncated to 15 lines")
		}
	}

	// Validate padding values
	if a.PaddingTop < 0 {
		a.PaddingTop = 0
		errs = append(errs, "padding_top must be >= 0, set to 0")
	}
	if a.PaddingTop > 10 {
		a.PaddingTop = 10
		errs = append(errs, "padding_top must be <= 10, set to 10")
	}
	if a.PaddingBottom < 0 {
		a.PaddingBottom = 0
		errs = append(errs, "padding_bottom must be >= 0, set to 0")
	}
	if a.PaddingBottom > 10 {
		a.PaddingBottom = 10
		errs = append(errs, "padding_bottom must be <= 10, set to 10")
	}

	if len(errs) > 0 {
		return errors.New(strings.Join(errs, "; "))
	}
	return nil
}

// Validate validates NetworkConfig
func (n *NetworkConfig) Validate() error {
	var errs []string

	// Validate reconnect delay (1-30 seconds)
	if n.ReconnectDelay < 1 {
		n.ReconnectDelay = 1
		errs = append(errs, "reconnect_delay must be >= 1, set to 1")
	}
	if n.ReconnectDelay > 30 {
		n.ReconnectDelay = 30
		errs = append(errs, "reconnect_delay must be <= 30, set to 30")
	}

	// Validate buffer size (0 or 10-200 MB)
	if n.BufferSizeMB != 0 {
		if n.BufferSizeMB < 10 {
			n.BufferSizeMB = 10
			errs = append(errs, "buffer_size_mb must be >= 10 or 0, set to 10")
		}
		if n.BufferSizeMB > 200 {
			n.BufferSizeMB = 200
			errs = append(errs, "buffer_size_mb must be <= 200, set to 200")
		}
	}

	if len(errs) > 0 {
		return errors.New(strings.Join(errs, "; "))
	}
	return nil
}

// Validate validates ShuffleConfig
func (s *ShuffleConfig) Validate() error {
	var errs []string

	// Validate interval (must be in valid set)
	validIntervals := map[int]bool{1: true, 3: true, 5: true, 10: true, 15: true}
	if !validIntervals[s.IntervalMinutes] {
		s.IntervalMinutes = 5
		errs = append(errs, "interval_minutes must be 1, 3, 5, 10, or 15, set to 5")
	}

	// Validate max history (must be in valid set)
	validHistory := map[int]bool{3: true, 5: true, 7: true, 10: true}
	if !validHistory[s.MaxHistory] {
		s.MaxHistory = 5
		errs = append(errs, "max_history must be 3, 5, 7, or 10, set to 5")
	}

	if len(errs) > 0 {
		return errors.New(strings.Join(errs, "; "))
	}
	return nil
}
