package theme

import (
	"os"
	"path/filepath"
	"strconv"
	"sync"

	"github.com/charmbracelet/lipgloss"
	"github.com/shinokada/tera/v3/internal/config"
)

// Theme holds all theme configuration
type Theme struct {
	Colors  ColorConfig   `yaml:"colors"`
	Padding PaddingConfig `yaml:"padding"`
}

// ColorConfig defines all color settings
type ColorConfig struct {
	Primary   string `yaml:"primary"`   // Titles, borders (default: 6/Cyan)
	Secondary string `yaml:"secondary"` // TERA header (default: 12/Bright Blue)
	Highlight string `yaml:"highlight"` // Selected items (default: 3/Yellow)
	Error     string `yaml:"error"`     // Error messages (default: 9/Bright Red)
	Success   string `yaml:"success"`   // Success messages (default: 2/Green)
	Muted     string `yaml:"muted"`     // Help text, subtle (default: 8/Bright Black/Gray)
	Text      string `yaml:"text"`      // Default text (default: 7/White)
}

// PaddingConfig defines padding settings
type PaddingConfig struct {
	PageHorizontal int `yaml:"page_horizontal"`
	PageVertical   int `yaml:"page_vertical"`
	ListItemLeft   int `yaml:"list_item_left"`
	BoxHorizontal  int `yaml:"box_horizontal"`
	BoxVertical    int `yaml:"box_vertical"`
}

// DefaultTheme returns the default theme configuration
func DefaultTheme() Theme {
	return Theme{
		Colors: ColorConfig{
			Primary:   "6",  // Cyan
			Secondary: "12", // Bright Blue
			Highlight: "3",  // Yellow
			Error:     "9",  // Bright Red
			Success:   "2",  // Green
			Muted:     "8",  // Bright Black (Gray)
			Text:      "7",  // White
		},
		Padding: PaddingConfig{
			PageHorizontal: 2,
			PageVertical:   1,
			ListItemLeft:   2,
			BoxHorizontal:  2,
			BoxVertical:    1,
		},
	}
}

// Global theme instance
var (
	current *Theme
	mu      sync.RWMutex
)

// GetConfigDir returns the theme config directory path
func GetConfigDir() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, "tera"), nil
}

// GetConfigPath returns the theme config file path
// In v3, this points to the unified config.yaml
func GetConfigPath() (string, error) {
	return config.GetConfigPath()
}

// LoadFromUnifiedConfig loads theme from the unified v3 config
func LoadFromUnifiedConfig() (*Theme, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}

	theme := &Theme{
		Colors: ColorConfig{
			Primary:   cfg.UI.Theme.Colors["primary"],
			Secondary: cfg.UI.Theme.Colors["secondary"],
			Highlight: cfg.UI.Theme.Colors["highlight"],
			Error:     cfg.UI.Theme.Colors["error"],
			Success:   cfg.UI.Theme.Colors["success"],
			Muted:     cfg.UI.Theme.Colors["muted"],
			Text:      cfg.UI.Theme.Colors["text"],
		},
		Padding: PaddingConfig{
			PageHorizontal: cfg.UI.Theme.Padding.PageHorizontal,
			PageVertical:   cfg.UI.Theme.Padding.PageVertical,
			ListItemLeft:   cfg.UI.Theme.Padding.ListItemLeft,
			BoxHorizontal:  cfg.UI.Theme.Padding.BoxHorizontal,
			BoxVertical:    cfg.UI.Theme.Padding.BoxVertical,
		},
	}

	return theme, nil
}

// Load loads the theme from unified config, or returns default if not found
func Load() (*Theme, error) {
	mu.Lock()
	defer mu.Unlock()

	theme, err := LoadFromUnifiedConfig()
	if err != nil {
		// If unified config fails, return default
		defaultTheme := DefaultTheme()
		current = &defaultTheme
		return current, nil
	}

	current = theme
	return current, nil
}

// Save saves the theme to unified config
func Save(theme *Theme) error {
	mu.Lock()
	defer mu.Unlock()
	return saveInternal(theme)
}

// saveInternal saves theme without lock (caller must hold lock)
func saveInternal(theme *Theme) error {
	// Load current unified config
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	// Update theme portion
	cfg.UI.Theme.Colors["primary"] = theme.Colors.Primary
	cfg.UI.Theme.Colors["secondary"] = theme.Colors.Secondary
	cfg.UI.Theme.Colors["highlight"] = theme.Colors.Highlight
	cfg.UI.Theme.Colors["error"] = theme.Colors.Error
	cfg.UI.Theme.Colors["success"] = theme.Colors.Success
	cfg.UI.Theme.Colors["muted"] = theme.Colors.Muted
	cfg.UI.Theme.Colors["text"] = theme.Colors.Text

	cfg.UI.Theme.Padding.PageHorizontal = theme.Padding.PageHorizontal
	cfg.UI.Theme.Padding.PageVertical = theme.Padding.PageVertical
	cfg.UI.Theme.Padding.ListItemLeft = theme.Padding.ListItemLeft
	cfg.UI.Theme.Padding.BoxHorizontal = theme.Padding.BoxHorizontal
	cfg.UI.Theme.Padding.BoxVertical = theme.Padding.BoxVertical

	// Save unified config
	return config.Save(cfg)
}

// Reset resets the theme to default values
func Reset() error {
	mu.Lock()
	defer mu.Unlock()

	theme := DefaultTheme()
	current = &theme
	return saveInternal(&theme)
}

// Current returns the current theme (loads if not already loaded)
func Current() *Theme {
	mu.RLock()
	if current != nil {
		defer mu.RUnlock()
		return current
	}
	mu.RUnlock()

	theme, _ := Load()
	return theme
}

// generateThemeYAML creates a YAML string with helpful comments
// This is now deprecated in v3 as theme is part of unified config
// Kept for backward compatibility with theme.yaml export if needed
func generateThemeYAML(theme *Theme) string {
	return `# TERA Theme Configuration
# Edit this file to customize the appearance of TERA
#
# ANSI Color Reference (256-color palette):
# ─────────────────────────────────────────
# Standard colors (0-7):
#   0: Black         1: Red           2: Green         3: Yellow
#   4: Blue          5: Magenta       6: Cyan          7: White
#
# Bright colors (8-15):
#   8: Bright Black (Gray)    9: Bright Red      10: Bright Green
#  11: Bright Yellow         12: Bright Blue     13: Bright Magenta
#  14: Bright Cyan           15: Bright White
#
# Extended colors (16-255): See https://en.wikipedia.org/wiki/ANSI_escape_code#8-bit
#
# You can also use hex colors: "#FF5733" or "#F53"

colors:
  # Main UI elements (titles, borders)
  primary: "` + theme.Colors.Primary + `"      # Default: 6 (Cyan)
  
  # TERA header text
  secondary: "` + theme.Colors.Secondary + `"   # Default: 12 (Bright Blue)
  
  # Selected/highlighted items
  highlight: "` + theme.Colors.Highlight + `"    # Default: 3 (Yellow)
  
  # Error messages
  error: "` + theme.Colors.Error + `"        # Default: 9 (Bright Red)
  
  # Success messages
  success: "` + theme.Colors.Success + `"      # Default: 2 (Green)
  
  # Help text, secondary info
  muted: "` + theme.Colors.Muted + `"        # Default: 8 (Bright Black/Gray)
  
  # Default text color
  text: "` + theme.Colors.Text + `"         # Default: 7 (White)

padding:
  # Horizontal padding for main content area
  page_horizontal: ` + strconv.Itoa(theme.Padding.PageHorizontal) + `
  
  # Vertical padding for main content area
  page_vertical: ` + strconv.Itoa(theme.Padding.PageVertical) + `
  
  # Left padding for list items
  list_item_left: ` + strconv.Itoa(theme.Padding.ListItemLeft) + `
  
  # Horizontal padding inside boxes
  box_horizontal: ` + strconv.Itoa(theme.Padding.BoxHorizontal) + `
  
  # Vertical padding inside boxes
  box_vertical: ` + strconv.Itoa(theme.Padding.BoxVertical) + `
`
}

// ExportLegacyThemeFile exports current theme as standalone theme.yaml
// This is for users who want to share or backup their theme separately
func ExportLegacyThemeFile(outputPath string) error {
	theme := Current()
	if theme == nil {
		theme = &Theme{}
		*theme = DefaultTheme()
	}

	content := generateThemeYAML(theme)
	return os.WriteFile(outputPath, []byte(content), 0644)
}

// Color helper methods for lipgloss integration
func (t *Theme) PrimaryColor() lipgloss.Color {
	return lipgloss.Color(t.Colors.Primary)
}

func (t *Theme) SecondaryColor() lipgloss.Color {
	return lipgloss.Color(t.Colors.Secondary)
}

func (t *Theme) HighlightColor() lipgloss.Color {
	return lipgloss.Color(t.Colors.Highlight)
}

func (t *Theme) ErrorColor() lipgloss.Color {
	return lipgloss.Color(t.Colors.Error)
}

func (t *Theme) SuccessColor() lipgloss.Color {
	return lipgloss.Color(t.Colors.Success)
}

func (t *Theme) MutedColor() lipgloss.Color {
	return lipgloss.Color(t.Colors.Muted)
}

func (t *Theme) TextColor() lipgloss.Color {
	return lipgloss.Color(t.Colors.Text)
}
