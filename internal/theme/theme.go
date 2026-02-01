package theme

import (
	"os"
	"path/filepath"
	"strconv"
	"sync"

	"github.com/charmbracelet/lipgloss"
	xdgdirs "github.com/go-music-players/xdg-dirs"
	"gopkg.in/yaml.v3"
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

// GetConfigPath returns the theme config file path
func GetConfigPath() (string, error) {
	dirs, err := xdgdirs.New("tera")
	if err != nil {
		return "", err
	}
	return filepath.Join(dirs.Config, "theme.yaml"), nil
}

// Load loads the theme from config file, or returns default if not found
func Load() (*Theme, error) {
	mu.Lock()
	defer mu.Unlock()

	configPath, err := GetConfigPath()
	if err != nil {
		theme := DefaultTheme()
		current = &theme
		return current, nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Config doesn't exist, create default
			theme := DefaultTheme()
			current = &theme
			// Save default config for user reference
			_ = saveInternal(&theme)
			return current, nil
		}
		return nil, err
	}

	var theme Theme
	if err := yaml.Unmarshal(data, &theme); err != nil {
		return nil, err
	}

	current = &theme
	return current, nil
}

// Save saves the theme to config file
func Save(theme *Theme) error {
	mu.Lock()
	defer mu.Unlock()
	return saveInternal(theme)
}

// saveInternal saves theme without lock (caller must hold lock)
func saveInternal(theme *Theme) error {
	configPath, err := GetConfigPath()
	if err != nil {
		return err
	}

	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}

	// Generate YAML with color reference comments
	content := generateThemeYAML(theme)

	return os.WriteFile(configPath, []byte(content), 0644)
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
