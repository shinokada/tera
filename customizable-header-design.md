# Customizable TERA Header - Implementation Design

## Current Implementation Analysis

Looking at `styles.go`, TERA currently renders a **hardcoded header**:

```go
// wrapPageWithHeader wraps content with TERA header at the top
func wrapPageWithHeader(content string) string {
    header := lipgloss.NewStyle().
        Width(50).
        Align(lipgloss.Center).
        Foreground(colorBlue()).
        Bold(true).
        PaddingTop(1).
        Render("TERA")  // ← HARDCODED!
    
    // ... rest of code
}
```

This header appears on **every page** via:
- `RenderPage()`
- `RenderPageWithBottomHelp()`

## ✅ **YES - Users Can Replace This!**

This is actually **EASIER** than I initially thought because:
1. ✅ Single source of truth: `wrapPageWithHeader()`
2. ✅ Already centralized rendering
3. ✅ Just need to make it configurable

## Implementation Design

### 1. Updated Configuration Schema

```yaml
# ~/.config/tera/appearance_config.yaml

appearance:
  header:
    # What to display
    mode: "default"  # "default" (TERA), "text", "ascii", "none"
    
    # For mode: "text" - simple text replacement
    custom_text: "My Radio"
    
    # For mode: "ascii" - User-provided ASCII art
    ascii_art: |
      ╔═══════════╗
      ║   RADIO   ║
      ╚═══════════╝
    
    # Display settings
    alignment: "center"  # "left", "center", "right"
    width: 50
    color: "auto"  # "auto" (use theme blue) or specific color
    bold: true
    padding_top: 1
    padding_bottom: 0

# Backwards compatibility: if file doesn't exist, use "TERA"
```

### 2. Storage Model

```go
// internal/storage/appearance_config.go

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
    Alignment   string `yaml:"alignment"`
    Width       int    `yaml:"width"`
    Color       string `yaml:"color"`
    Bold        bool   `yaml:"bold"`
    PaddingTop  int    `yaml:"padding_top"`
    PaddingBottom int  `yaml:"padding_bottom"`
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
```

### 3. Updated Header Renderer

```go
// internal/ui/header.go - NEW FILE

package ui

import (
    "strings"
    
    "github.com/charmbracelet/lipgloss"
    "github.com/shinokada/tera/internal/storage"
)

// HeaderRenderer handles rendering the app header based on configuration
type HeaderRenderer struct {
    config storage.AppearanceConfig
}

// NewHeaderRenderer creates a new header renderer with current config
func NewHeaderRenderer() *HeaderRenderer {
    config, err := storage.LoadAppearanceConfig()
    if err != nil {
        config = storage.DefaultAppearanceConfig()
    }
    
    return &HeaderRenderer{
        config: config,
    }
}

// Render generates the header content based on configuration
func (h *HeaderRenderer) Render() string {
    switch h.config.Header.Mode {
    case storage.HeaderModeNone:
        return "" // No header
        
    case storage.HeaderModeText:
        return h.renderText()
        
    case storage.HeaderModeASCII:
        return h.renderASCII()
        
    default: // HeaderModeDefault
        return h.renderDefault()
    }
}

// renderDefault renders the default "TERA" header
func (h *HeaderRenderer) renderDefault() string {
    style := h.createBaseStyle()
    return style.Render("TERA")
}

// renderText renders custom text header
func (h *HeaderRenderer) renderText() string {
    if h.config.Header.CustomText == "" {
        return h.renderDefault() // Fallback
    }
    
    style := h.createBaseStyle()
    return style.Render(h.config.Header.CustomText)
}

// renderASCII renders ASCII art header
func (h *HeaderRenderer) renderASCII() string {
    if h.config.Header.AsciiArt == "" {
        return h.renderDefault() // Fallback if no ASCII art provided
    }
    
    return h.styleASCII(h.config.Header.AsciiArt)
}

// createBaseStyle creates the base style for text headers
func (h *HeaderRenderer) createBaseStyle() lipgloss.Style {
    style := lipgloss.NewStyle().
        Width(h.config.Header.Width).
        PaddingTop(h.config.Header.PaddingTop).
        PaddingBottom(h.config.Header.PaddingBottom)
    
    // Alignment
    switch h.config.Header.Alignment {
    case "left":
        style = style.Align(lipgloss.Left)
    case "right":
        style = style.Align(lipgloss.Right)
    default:
        style = style.Align(lipgloss.Center)
    }
    
    // Color
    if h.config.Header.Color == "auto" {
        style = style.Foreground(colorBlue())
    } else {
        style = style.Foreground(lipgloss.Color(h.config.Header.Color))
    }
    
    // Bold
    if h.config.Header.Bold {
        style = style.Bold(true)
    }
    
    return style
}

// styleASCII applies styling to ASCII art
func (h *HeaderRenderer) styleASCII(art string) string {
    // Create style for each line
    lineStyle := lipgloss.NewStyle().
        Width(h.config.Header.Width)
    
    // Alignment
    switch h.config.Header.Alignment {
    case "left":
        lineStyle = lineStyle.Align(lipgloss.Left)
    case "right":
        lineStyle = lineStyle.Align(lipgloss.Right)
    default:
        lineStyle = lineStyle.Align(lipgloss.Center)
    }
    
    // Color
    if h.config.Header.Color == "auto" {
        lineStyle = lineStyle.Foreground(colorBlue())
    } else {
        lineStyle = lineStyle.Foreground(lipgloss.Color(h.config.Header.Color))
    }
    
    // Split into lines and style each
    lines := strings.Split(art, "\n")
    var result strings.Builder
    
    // Top padding
    for i := 0; i < h.config.Header.PaddingTop; i++ {
        result.WriteString("\n")
    }
    
    // Styled content
    for _, line := range lines {
        result.WriteString(lineStyle.Render(line))
        result.WriteString("\n")
    }
    
    // Bottom padding
    for i := 0; i < h.config.Header.PaddingBottom; i++ {
        result.WriteString("\n")
    }
    
    return result.String()
}

// Reload reloads the configuration (call after config changes)
func (h *HeaderRenderer) Reload() error {
    config, err := storage.LoadAppearanceConfig()
    if err != nil {
        return err
    }
    h.config = config
    return nil
}
```

### 4. Update styles.go

```go
// internal/ui/styles.go - MODIFIED

// Global header renderer instance (initialized in app.go)
var globalHeaderRenderer *HeaderRenderer

// InitializeHeaderRenderer initializes the global header renderer
func InitializeHeaderRenderer() {
    globalHeaderRenderer = NewHeaderRenderer()
}

// wrapPageWithHeader wraps content with header at the top and applies consistent padding
func wrapPageWithHeader(content string) string {
    var b strings.Builder
    
    // Render header using configuration
    if globalHeaderRenderer != nil {
        header := globalHeaderRenderer.Render()
        if header != "" {
            b.WriteString(header)
            // Only add newline if header is not empty
            if header != "" && !strings.HasSuffix(header, "\n") {
                b.WriteString("\n")
            }
        }
    } else {
        // Fallback to default if renderer not initialized
        header := lipgloss.NewStyle().
            Width(50).
            Align(lipgloss.Center).
            Foreground(colorBlue()).
            Bold(true).
            PaddingTop(1).
            Render("TERA")
        b.WriteString(header)
        b.WriteString("\n")
    }
    
    b.WriteString(content)
    return docStyleNoTopPadding().Render(b.String())
}
```

### 5. Initialize in app.go

```go
// internal/ui/app.go - MODIFIED

func NewModel(favoritePath string) Model {
    // ... existing code ...
    
    // Initialize header renderer
    InitializeHeaderRenderer()
    
    // ... rest of existing code ...
}
```

### 6. Settings UI Design

```
┌──────────────────────────────────────────────────────────────┐
│ ⚙️  Settings > Appearance                                    │
├──────────────────────────────────────────────────────────────┤
│                                                               │
│ ┌─ Header Settings ───────────────────────────────────────┐  │
│ │                                                          │  │
│ │ Header Display Mode:                                    │  │
│ │   ● Default ("TERA")                                    │  │
│ │   ○ Custom Text                                         │  │
│ │   ○ ASCII Art Banner                                    │  │
│ │   ○ None (No header)                                    │  │
│ │                                                          │  │
│ │ ─────────────────────────────────────────────────────── │  │
│ │                                                          │  │
│ │ [If Custom Text selected]                               │  │
│ │                                                          │  │
│ │ Custom Text: [My Radio Station____________]            │  │
│ │              (max 100 characters)                       │  │
│ │                                                          │  │
│ │ Preview:                                                │  │
│ │         My Radio Station                                │  │
│ │                                                          │  │
│ │ ─────────────────────────────────────────────────────── │  │
│ │                                                          │  │
│ │ [If ASCII Art Banner selected]                          │  │
│ │                                                          │  │
│ │ ┌─ Paste ASCII art (max 15 lines): ──────────────────┐ │  │
│ │ │  ____      _    ____ ___ ___                        │ │  │
│ │ │ |  _ \    / \  |  _ \_ _/ _ \                       │ │  │
│ │ │ | |_) |  / _ \ | | | | | | | |                      │ │  │
│ │ │ |  _ <  / ___ \| |_| | | |_| |                      │ │  │
│ │ │ |_| \_\/_/   \_\____/___\___/                       │ │  │
│ │ │ _                                                   │ │  │
│ │ └──────────────────────────────────────────────────────┘ │  │
│ │                                                          │  │
│ │ Preview:                                                │  │
│ │ ┌────────────────────────────────────────────────────┐  │  │
│ │ │      ____      _    ____ ___ ___                   │  │  │
│ │ │     |  _ \    / \  |  _ \_ _/ _ \                  │  │  │
│ │ │     | |_) |  / _ \ | | | | | | | |                 │  │  │
│ │ │     |  _ <  / ___ \| |_| | | |_| |                 │  │  │
│ │ │     |_| \_\/_/   \_\____/___\___/                  │  │  │
│ │ └────────────────────────────────────────────────────┘  │  │
│ │                                                          │  │
│ │ ─────────────────────────────────────────────────────── │  │
│ │                                                          │  │
│ │ Display Options (all modes):                            │  │
│ │   Alignment: [Center ▼]  (Left/Center/Right)           │  │
│ │   Color:     [Auto ▼]    (Auto/Custom)                 │  │
│ │   Bold:      [✓]                                        │  │
│ │                                                          │  │
│ └──────────────────────────────────────────────────────────┘  │
│                                                               │
│ [Save & Apply]  [Preview Changes]  [Reset to Default]        │
│                                                               │
├──────────────────────────────────────────────────────────────┤
│ ↑↓/jk: Navigate • Tab: Next field • Space: Toggle           │
│ Enter: Edit • Esc: Back • Ctrl+C: Quit                       │
└──────────────────────────────────────────────────────────────┘
```

### 7. ASCII Art Guidelines

**For Users Creating Custom ASCII Art:**

1. **Maximum Size**: 15 lines maximum
2. **Width**: Should fit within terminal width (typically 50-80 chars)
3. **Character Set**: Use standard ASCII characters for best compatibility
4. **Tools**: Users can create ASCII art using:
   - Online tools (patorjk.com/software/taag/)
   - External tools like `figlet` (user installs separately and copies output)
   - Hand-crafted art
   - Box-drawing characters for borders

**Common Box-Drawing Characters:**
```
╔═══╗  ┌───┐  ╭───╮  ┏━━━┓
║   ║  │   │  │   │  ┃   ┃
╚═══╝  └───┘  ╰───╯  ┗━━━┛
```

**Example Using External figlet:**
```bash
# User runs this externally
$ figlet -f slant "RADIO"

# Copies the output and pastes into TERA settings:
      ____  ___    ____ ____ ____
     / __ \/   |  / __ \  _/ __ \
    / /_/ / /| | / / / // // / / /
   / _, _/ ___ |/ /_/ // // /_/ / 
  /_/ |_/_/  |_/_____/___/\____/
```

### 8. Example Configurations

#### Example 1: Simple Custom Text
```yaml
appearance:
  header:
    mode: "text"
    custom_text: "🎵 John's Radio Station 🎵"
    alignment: "center"
    width: 50
    color: "auto"
    bold: true
    padding_top: 1
```

Result:
```
    🎵 John's Radio Station 🎵
```

#### Example 2: Custom ASCII Art (Simple)
```yaml
appearance:
  header:
    mode: "ascii"
    ascii_art: |
      ╔══════════════════════╗
      ║  🎵  MY STATION  🎵  ║
      ╚══════════════════════╝
    alignment: "center"
    width: 50
    color: "99"  # Purple
```

Result:
```
    ╔══════════════════════╗
    ║  🎵  MY STATION  🎵  ║
    ╚══════════════════════╝
```

#### Example 3: Custom ASCII Art (Fancy)
```yaml
appearance:
  header:
    mode: "ascii"
    ascii_art: |
       ____  ___    ____ ____ ____
      / __ \/   |  / __ \  _/ __ \
     / /_/ / /| | / / / // // / / /
    / _, _/ ___ |/ /_/ // // /_/ / 
   /_/ |_/_/  |_/_____/___/\____/
    alignment: "center"
    width: 60
    color: "auto"
```

Result:
```
       ____  ___    ____ ____ ____
      / __ \/   |  / __ \  _/ __ \
     / /_/ / /| | / / / // // / / /
    / _, _/ ___ |/ /_/ // // /_/ / 
   /_/ |_/_/  |_/_____/___/\____/
```

#### Example 4: No Header
```yaml
appearance:
  header:
    mode: "none"
```

Result: (Header completely removed, content starts immediately)

### 9. Implementation Phases

#### Phase 1: Foundation (Day 1-2)
- [ ] Create `appearance_config.go` with full schema
- [ ] Create `header.go` with HeaderRenderer
- [ ] Update `styles.go` to use HeaderRenderer
- [ ] Initialize in `app.go`
- [ ] Test with default mode (should work exactly as before)

#### Phase 2: Custom Text Mode (Day 2-3)
- [ ] Add text mode to settings UI
- [ ] Implement text input handling
- [ ] Add live preview
- [ ] Test and refine

#### Phase 3: Custom ASCII Art (Day 3-4)
- [ ] Add multi-line text input for custom art
- [ ] Implement line count validation
- [ ] Add custom art preview
- [ ] Test edge cases

#### Phase 4: Polish & Testing (Day 4-5)
- [ ] Add alignment options
- [ ] Add color customization
- [ ] Implement "None" mode
- [ ] Write comprehensive tests
- [ ] Update documentation

### 10. User Workflows

#### Workflow 1: Replace "TERA" with Custom Text
1. Main Menu → Settings → Appearance
2. Header Display Mode: Select "Custom Text"
3. Enter: "🎵 My Radio 🎵"
4. Save & Apply
5. Return to main menu
6. See custom text instead of "TERA" ✓

#### Workflow 2: Use Custom ASCII Art
1. Create ASCII art externally (using online tools or figlet)
2. Main Menu → Settings → Appearance
3. Header Display Mode: Select "ASCII Art Banner"
4. Paste custom art in text box
5. Preview updates live
6. Adjust alignment if needed
7. Save & Apply
8. Custom art appears everywhere ✓

#### Workflow 3: Remove Header Entirely
1. Main Menu → Settings → Appearance
2. Header Display Mode: Select "None"
3. Save & Apply
4. Header disappears, more screen space! ✓

### 11. Backwards Compatibility

**Perfect!** This design is 100% backwards compatible:

- ✅ If `appearance_config.yaml` doesn't exist → uses default "TERA"
- ✅ If config is corrupted → falls back to default
- ✅ Existing users see no change until they configure it
- ✅ No breaking changes to any existing code
- ✅ No external dependencies required

### 12. Benefits

✅ **Complete Customization**: Users can replace "TERA" with anything
✅ **Multiple Options**: Text, ASCII, or remove completely
✅ **Easy Implementation**: Single source of truth in `wrapPageWithHeader()`
✅ **Global Effect**: Header change applies to ALL screens automatically
✅ **Backwards Compatible**: Works with existing installations
✅ **Flexible**: From minimal (text) to bold (ASCII art)
✅ **Settings Location**: Logical place in Settings > Appearance
✅ **No External Dependencies**: Users paste their own ASCII art

### 13. Potential Concerns & Solutions

#### Concern 1: Creating ASCII art
**Solution**: 
- Document online tools (patorjk.com/software/taag/)
- Users can use `figlet` externally if they want
- Provide examples of box-drawing characters
- No dependency on external tools

#### Concern 2: User enters very long text/ASCII
**Solution**:
- Validation limits (100 chars for text, 15 lines for ASCII)
- Preview shows how it will look
- Automatic truncation with warning

#### Concern 3: ASCII art breaks layout
**Solution**:
- Max height validation
- Width constraints
- Preview before applying
- Easy reset to default

#### Concern 4: Performance with ASCII rendering
**Solution**:
- Header generated once, cached
- Reload only when config changes
- No external tool execution needed

### 14. File Structure

```
internal/
├── ui/
│   ├── appearance_settings.go      # NEW: Settings UI
│   ├── header.go                   # NEW: HeaderRenderer
│   ├── styles.go                   # MODIFIED: Use HeaderRenderer
│   └── app.go                      # MODIFIED: Initialize renderer
└── storage/
    └── appearance_config.go        # NEW: Configuration
```

## Final Recommendation

### ✅ **Absolutely YES - This is Perfect!**

**Why it's excellent**:

1. **Easy to Implement**: Single point of change (`wrapPageWithHeader`)
2. **Global Effect**: One config affects all screens automatically
3. **User Freedom**: Complete control - replace, customize, or remove
4. **Progressive Options**:
   - Level 1: Simple text replacement
   - Level 2: Custom ASCII art (paste from external source)
   - Level 3: No header at all
5. **Settings is Perfect**: Natural location for this feature
6. **No Dependencies**: Users create ASCII art externally and paste it in

**This is actually BETTER than my initial proposal** because:
- Simpler implementation (one central point vs multiple components)
- More consistent (header changes everywhere automatically)
- Easier to maintain (single source of truth)
- No external tool dependencies

**Estimated Time**: ~3-4 days for complete implementation with all modes (simplified without figlet integration)

**Go for it!** 🚀
