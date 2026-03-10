# TERA v3 Technical Roadmap

This document contains technical implementation details for v3 development.

## Architecture (v3.0.0)

```
v3/
├── cmd/tera/main.go           # CLI entry point
├── internal/
│   ├── api/                   # Radio Browser API client
│   ├── blocklist/             # Station blocking
│   ├── config/                # NEW: Unified configuration
│   ├── credentials/           # Secure token storage
│   ├── gist/                  # GitHub Gist sync
│   ├── player/                # MPV integration
│   ├── shuffle/               # Shuffle mode manager
│   ├── storage/               # Config & favorites
│   ├── theme/                 # Theme system
│   └── ui/                    # Bubble Tea interface
└── pkg/utils/                 # Public utilities
```

---

## Unified Configuration System

### Current Problem
Config scattered across multiple files:
- `theme.yaml`
- `appearance_config.yaml`
- `connection_config.yaml`
- `shuffle.yaml`

**Note:** `blocklist.json`, `voted_stations.json`, and `favorites/*.json` are **user data**, not configuration. They should remain separate from system config.

### Solution Architecture

**New structure:**
```go
// v3/internal/config/config.go
package config

type Config struct {
    Version string `yaml:"version"`
    Player  PlayerConfig `yaml:"player"`
    UI      UIConfig `yaml:"ui"`
    Network NetworkConfig `yaml:"network"`
    Shuffle ShuffleConfig `yaml:"shuffle"`
}

type PlayerConfig struct {
    DefaultVolume int `yaml:"default_volume"`
    BufferSizeMB  int `yaml:"buffer_size_mb"`
}

type UIConfig struct {
    Theme          ThemeConfig      `yaml:"theme"`
    Appearance     AppearanceConfig `yaml:"appearance"`
    DefaultList    string          `yaml:"default_list"`
}

type ThemeConfig struct {
    Name    string            `yaml:"name"`
    Colors  map[string]string `yaml:"colors"`
    Padding PaddingConfig     `yaml:"padding"`
}

type AppearanceConfig struct {
    HeaderMode  string `yaml:"header_mode"`  // text, ascii, none
    HeaderAlign string `yaml:"header_align"` // left, center, right
    HeaderWidth int    `yaml:"header_width"`
}

type NetworkConfig struct {
    AutoReconnect  bool `yaml:"auto_reconnect"`
    ReconnectDelay int  `yaml:"reconnect_delay"`
    BufferSizeMB   int  `yaml:"buffer_size_mb"`
}

type ShuffleConfig struct {
    AutoAdvance      bool `yaml:"auto_advance"`
    IntervalMinutes  int  `yaml:"interval_minutes"`
    RememberHistory  bool `yaml:"remember_history"`
    MaxHistory       int  `yaml:"max_history"`
}
```

**New file structure:**
```yaml
# config.yaml
version: "3.0"

player:
  default_volume: 80
  buffer_size_mb: 50

ui:
  theme:
    name: "default"
    colors:
      primary: "#00FFFF"
      highlight: "#FFFF00"
    padding:
      list_item_left: 2
      
  appearance:
    header_mode: "text"
    header_align: "center"
    header_width: 60
    
  default_list: "My-favorites"

network:
  auto_reconnect: true
  reconnect_delay: 5
  buffer_size_mb: 50

shuffle:
  auto_advance: true
  interval_minutes: 5
  remember_history: true
  max_history: 7
```

### Implementation Steps

**1: Config package**
1. Create `v3/internal/config/` package
2. Define `Config` struct hierarchy
3. Implement `Load()` and `Save()` functions
4. Write unit tests

**2: Migration logic**
```go
// v3/internal/config/migrate.go
package config

func MigrateFromV2(v2ConfigDir string) (*Config, error) {
    cfg := DefaultConfig()
    
    // Read old theme.yaml
    if theme, err := readV2Theme(v2ConfigDir); err == nil {
        cfg.UI.Theme = theme
    }
    
    // Read old appearance_config.yaml
    if appearance, err := readV2Appearance(v2ConfigDir); err == nil {
        cfg.UI.Appearance = appearance
    }
    
    // Read old connection_config.yaml
    if network, err := readV2Connection(v2ConfigDir); err == nil {
        cfg.Network = network
    }
    
    // Read old shuffle.yaml
    if shuffle, err := readV2Shuffle(v2ConfigDir); err == nil {
        cfg.Shuffle = shuffle
    }
    
    // Backup old files
    backupV2Configs(v2ConfigDir)
    
    return cfg, nil
}
```

**3: Integration**
1. Update `internal/storage/config.go` to use new config package
2. Update all packages that read config (theme, player, ui)
3. Add migration check in `cmd/tera/main.go`

**4: Testing & Release**
1. Test migration on all platforms
2. Update documentation
3. Tag v3.0.0

### Files to Create/Modify

**New files:**
- `v3/internal/config/config.go`
- `v3/internal/config/loader.go`
- `v3/internal/config/migrate.go`
- `v3/internal/config/config_test.go`
- `v3/internal/config/migrate_test.go`

**Modified files:**
- `v3/internal/storage/config.go` (use new config)
- `v3/internal/theme/theme.go` (read from unified config)
- `v3/internal/ui/app.go` (check migration on startup)
- `v3/cmd/tera/main.go` (auto-migrate check)

---

## User Data Organization

### Distinction: Config vs. User Data

**System Configuration** (config.yaml):
- How the application behaves
- Default settings
- Theme and appearance preferences
- Network settings

**User Data** (separate files):
- What the user has done
- User-created content
- Usage history
- Caching and state

### New Directory Structure

```
~/.config/tera/              # Config directory (os.UserConfigDir())
├── config.yaml              # Unified system configuration
└── data/                    # User data directory
    ├── blocklist.json       # User-blocked stations
    ├── voted_stations.json  # User voting history
    ├── favorites/           # User playlists
    │   ├── Blues.json
    │   ├── Jazz.json
    │   └── My-favorites.json
    └── cache/               # Temporary data
        ├── gist_metadata.json
        └── search-history.json
```

### Migration from v2

Migration should be done automatically on first run.

```go
// v3/cmd/tera/main.go
func main() {
    // 1. Detect if migration is needed
    configPath := filepath.Join(os.UserConfigDir(), "tera", "config.yaml")
    
    if !fileExists(configPath) {
        // No v3 config exists - check for v2 config
        v2ConfigDir := filepath.Join(os.UserConfigDir(), "tera")
        if hasV2Config(v2ConfigDir) {
            // Auto-migrate with user notification
            fmt.Println("🔄 Migrating from Tera v2 to v3...")
            
            if err := migrateFromV2(v2ConfigDir); err != nil {
                fmt.Fprintf(os.Stderr, "Migration failed: %v\n", err)
                fmt.Println("Your v2 config has been backed up.")
                fmt.Println("Please report this issue: https://github.com/...")
                os.Exit(1)
            }
            
            fmt.Println("✓ Migration complete!")
            fmt.Println("  - Config unified → ~/.config/tera/config.yaml")
            fmt.Println("  - User data → ~/.config/tera/data/")
            fmt.Println("  - GitHub token → OS keychain")
            fmt.Println("")
        }
    }
    
    // 2. Continue normal startup
    app.Run()
}

func migrateFromV2(v2ConfigDir string) error {
    // Create backup first
    backupDir := v2ConfigDir + ".v2-backup-" + time.Now().Format("20060102-150405")
    if err := copyDir(v2ConfigDir, backupDir); err != nil {
        return fmt.Errorf("backup failed: %w", err)
    }
    
    // Migrate config files
    cfg, err := config.MigrateFromV2(v2ConfigDir)
    if err != nil {
        return fmt.Errorf("config migration failed: %w", err)
    }
    
    // Migrate user data
    if err := storage.MigrateDataFromV2(v2ConfigDir); err != nil {
        return fmt.Errorf("data migration failed: %w", err)
    }
    
    // Migrate GitHub token to keychain
    if err := credentials.MigrateFromFile(v2ConfigDir); err != nil {
        // Non-fatal - user can set token later in Settings
        fmt.Printf("⚠️  Could not migrate GitHub token: %v\n", err)
        fmt.Println("   You can set it later in Settings > GitHub Token")
    }
    
    // Save new config
    if err := cfg.Save(); err != nil {
        return fmt.Errorf("save config failed: %w", err)
    }
    
    // Clean up old config files (optional - keep backup)
    // removeOldV2Files(v2ConfigDir)
    
    return nil
}
```

```go
// v3/internal/storage/migrate.go
func MigrateDataFromV2(v2ConfigDir string) error {
    v3DataDir := filepath.Join(os.UserConfigDir(), "tera", "data")
    
    // Migrate user data (not config)
    filesToMove := map[string]string{
        "blocklist.json":      "blocklist.json",
        "voted_stations.json": "voted_stations.json",
        "favorites":           "favorites",
        "gist_metadata.json":  "cache/gist_metadata.json",
    }
    
    for oldFile, newFile := range filesToMove {
        oldPath := filepath.Join(v2ConfigDir, oldFile)
        newPath := filepath.Join(v3DataDir, newFile)
        if err := moveIfExists(oldPath, newPath); err != nil {
            return err
        }
    }
    
    return nil
}
```

### Testing Migration
```sh
# Hidden command for testing/debugging
tera debug migrate-check

# Output:
# V2 Config detected:
#   ✓ theme.yaml
#   ✓ appearance_config.yaml
#   ✓ shuffle.yaml
#   ✓ blocklist.json (37 stations)
#   ✓ favorites/ (5 playlists)
#   ✓ tokens/github_token
# 
# Migration would:
#   - Unified config → config.yaml
#   - Move user data → data/
#   - Migrate token → keychain
```

---

## Secure Credential Storage

### Current Problem
`tokens/github_token` stored as plain text file - insecure and platform-specific.

### Solution: OS Keychain Integration

Use platform-native secure storage:
- **macOS**: Keychain
- **Linux**: Secret Service (gnome-keyring, KWallet)
- **Windows**: Credential Manager

### Implementation

**Add dependency:**
```bash
go get github.com/zalando/go-keyring
```

**New package:**
```go
// v3/internal/credentials/credentials.go
package credentials

import (
    "fmt"
    "os"
    "path/filepath"
    "strings"
    "github.com/zalando/go-keyring"
)

const (
    serviceName = "tera"
    tokenKey    = "github_token"
)

// SetGitHubToken stores the GitHub token securely
func SetGitHubToken(token string) error {
    return keyring.Set(serviceName, tokenKey, token)
}

// GetGitHubToken retrieves the GitHub token
// Falls back to TERA_GITHUB_TOKEN env var for headless environments
func GetGitHubToken() (string, error) {
    // Try OS keychain first
    token, err := keyring.Get(serviceName, tokenKey)
    if err == nil {
        return token, nil
    }
    
    // Fallback to environment variable (for CI/CD, headless servers)
    if envToken := os.Getenv("TERA_GITHUB_TOKEN"); envToken != "" {
        return envToken, nil
    }
    
    if err == keyring.ErrNotFound {
        return "", fmt.Errorf("github token not configured. Run: tera config set-token")
    }
    
    return "", fmt.Errorf("failed to retrieve github token: %w", err)
}

// DeleteGitHubToken removes the GitHub token
func DeleteGitHubToken() error {
    return keyring.Delete(serviceName, tokenKey)
}

// MigrateFromFile migrates token from v2 file storage to keychain
func MigrateFromFile(v2ConfigDir string) error {
    oldPath := filepath.Join(v2ConfigDir, "tokens", "github_token")
    
    data, err := os.ReadFile(oldPath)
    if err != nil {
        if os.IsNotExist(err) {
            return nil // No token to migrate
        }
        return fmt.Errorf("failed to read old token: %w", err)
    }
    
    token := strings.TrimSpace(string(data))
    if token == "" {
        return nil
    }
    
    // Store in keychain
    if err := SetGitHubToken(token); err != nil {
        return fmt.Errorf("failed to store token in keychain: %w", err)
    }
    
    // Remove old file
    os.Remove(oldPath)
    
    // Remove tokens/ directory if empty
    tokensDir := filepath.Dir(oldPath)
    if isEmpty, _ := isDirEmpty(tokensDir); isEmpty {
        os.Remove(tokensDir)
    }
    
    fmt.Println("✓ Migrated GitHub token to secure storage")
    return nil
}

func isDirEmpty(dir string) (bool, error) {
    entries, err := os.ReadDir(dir)
    if err != nil {
        return false, err
    }
    return len(entries) == 0, nil
}
```

**Settings UI Integration:**
```go
// v3/internal/ui/settings.go
package ui

type TokenSettingsModel struct {
    tokenInput    textinput.Model
    mode          string // "view", "edit", "confirm"
    showToken     bool
    currentToken  string
    errorMessage  string
    successMessage string
}

func (m *TokenSettingsModel) View() string {
    switch m.mode {
    case "view":
        return m.viewMode()
    case "edit":
        return m.editMode()
    case "confirm":
        return m.confirmMode()
    }
    return ""
}

func (m *TokenSettingsModel) viewMode() string {
    var token string
    if m.showToken {
        token = m.currentToken
    } else {
        token = strings.Repeat("•", min(len(m.currentToken), 20))
    }
    
    status := "❌ No token configured"
    if m.currentToken != "" {
        status = "✓ Token configured"
    }
    
    return fmt.Sprintf(`
  Settings > GitHub Token


GitHub Token: %s   [%s]
                                     
Current Status: %s

Commands:
  e: Edit token
  d: Delete token
  s: Show/Hide token
  Esc: Back to Settings

`, 
        token,
        ternary(m.showToken, "Hide", "Show"),
        status,
    )
}

func (m *TokenSettingsModel) editMode() string {
    return fmt.Sprintf(`
  Settings > GitHub Token > Edit


Enter GitHub Token:
%s

Commands:
  Enter: Save token
  Ctrl+U: Clear input
  Esc: Cancel

`,
        m.tokenInput.View(),
    )
}

func (m *TokenSettingsModel) confirmMode() string {
    return fmt.Sprintf(`
  Confirm Token


Token: %s

Save this token to secure storage?

  y: Yes, save token
  n: No, go back and edit

`,
        m.tokenInput.Value(),
    )
}

func (m *TokenSettingsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch m.mode {
        case "view":
            return m.handleViewKeys(msg)
        case "edit":
            return m.handleEditKeys(msg)
        case "confirm":
            return m.handleConfirmKeys(msg)
        }
    }
    return m, nil
}

func (m *TokenSettingsModel) handleViewKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
    switch msg.String() {
    case "e":
        m.mode = "edit"
        m.tokenInput.SetValue(m.currentToken)
        m.tokenInput.Focus()
    case "d":
        return m, m.deleteToken()
    case "s":
        m.showToken = !m.showToken
    case "esc":
        return m, navigateToSettings
    }
    return m, nil
}

func (m *TokenSettingsModel) handleEditKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
    switch msg.String() {
    case "enter":
        m.mode = "confirm"
        return m, nil
    case "ctrl+u":
        m.tokenInput.SetValue("")
    case "esc":
        m.mode = "view"
        return m, nil
    }
    
    var cmd tea.Cmd
    m.tokenInput, cmd = m.tokenInput.Update(msg)
    return m, cmd
}

func (m *TokenSettingsModel) handleConfirmKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
    switch msg.String() {
    case "y":
        return m, m.saveToken()
    case "n":
        m.mode = "edit"
        return m, nil
    }
    return m, nil
}

func (m *TokenSettingsModel) saveToken() tea.Cmd {
    return func() tea.Msg {
        token := m.tokenInput.Value()
        if err := credentials.SetGitHubToken(token); err != nil {
            return tokenErrorMsg{err: err}
        }
        return tokenSuccessMsg{message: "✓ GitHub token saved securely"}
    }
}

func (m *TokenSettingsModel) deleteToken() tea.Cmd {
    return func() tea.Msg {
        if err := credentials.DeleteGitHubToken(); err != nil {
            return tokenErrorMsg{err: err}
        }
        return tokenSuccessMsg{message: "✓ GitHub token deleted"}
    }
}
```

**User Experience:**

1. **Interactive Users (TUI):**
   - Navigate to Settings > GitHub Token
   - Press `e` to edit
   - Enter token (visible by default, can verify correctness)
   - Press `Enter` to proceed to confirmation
   - Press `y` to save to OS keychain
   - Token is automatically saved to secure storage

2. **Headless Environments (CI/CD, servers):**
   ```bash
   export TERA_GITHUB_TOKEN=ghp_xxxxx
   tera sync
   ```


### Benefits

✅ More secure than plain text files  
✅ Cross-platform (macOS/Linux/Windows using `os.UserConfigDir()` principle)  
✅ Standard practice (same as browsers, Docker, Git)  
✅ No `tokens/` directory needed  
✅ Automatic encryption by OS  
✅ Environment variable fallback for headless systems  

### Migration Timeline

**v3.0.0:**
1. Auto-migrate token from `tokens/github_token` to keychain on first run
2. Keep reading from file as fallback (deprecated, warning shown)
3. Document new token management via Settings UI

**v3.1.0:**
1. Remove file fallback completely
2. Only support keychain + environment variable

---

## v3.1.0: New features

### (fzf like) Sort function in Favorites list
### Recent search with station name


## v3.1.0: Station Metadata

### New Fields

```go
// v3/internal/api/models.go
type Station struct {
    // Existing fields...
    StationUUID string
    Name        string
    URLResolved string
    
    // NEW: Play statistics
    PlayCount  int       `json:"play_count,omitempty"`
    LastPlayed time.Time `json:"last_played,omitempty"`
    FirstPlayed time.Time `json:"first_played,omitempty"`
}
```

### Storage

Favorites files remain JSON, just with extra fields:
```json
{
  "stations": [
    {
      "stationuuid": "abc123",
      "name": "Jazz FM",
      "play_count": 42,
      "last_played": "2026-05-15T14:30:00Z",
      "first_played": "2026-03-01T09:15:00Z"
    }
  ]
}
```

### UI Updates

**New menu item:**
```
1. Play from Favorites
2. Most Played          ← NEW
3. Recently Played      ← NEW
4. Search Stations
```

**Implementation:**
```go
// v3/internal/ui/statistics.go
func (m *Model) showMostPlayed() tea.Cmd {
    stations := m.storage.GetAllStations()
    sort.Slice(stations, func(i, j int) bool {
        return stations[i].PlayCount > stations[j].PlayCount
    })
    
    return m.showStationList(stations[:10], "Top 10 Most Played")
}
```

### Files to Create/Modify

**Modified files:**
- `v3/internal/api/models.go` (add new fields)
- `v3/internal/storage/favorites.go` (update on play)
- `v3/internal/ui/menu.go` (new menu items)

**New files:**
- `v3/internal/ui/statistics.go`

---

## v3.2.0: User Ratings

### Schema

```go
type Station struct {
    // Existing fields...
    
    // NEW: User rating (1-5 stars, nil = not rated)
    UserRating *int `json:"user_rating,omitempty"`
}

// Helper methods
func (s *Station) SetRating(stars int) error {
    if stars < 1 || stars > 5 {
        return errors.New("rating must be 1-5")
    }
    s.UserRating = &stars
    return nil
}

func (s *Station) GetRating() int {
    if s.UserRating == nil {
        return 0  // Not rated
    }
    return *s.UserRating
}
```

### UI

**While playing:**
```
🎵 Now Playing

Station: Jazz FM
Rating: ★★★★☆ (4/5)          ← Show current rating

Press 1-5 to rate this station
Press 0 to clear rating
```

**In lists:**
```
Favorites:
  1. Jazz FM ★★★★★
  2. BBC Radio ★★★★☆
  3. KEXP ★★★☆☆
  4. Classical FM (not rated)
```

---

## v3.4.0: Custom Tags

### Schema

```go
type Station struct {
    // Existing fields...
    
    // NEW: User-defined tags
    CustomTags []string `json:"custom_tags,omitempty"`
}
```

### Storage

```json
{
  "stationuuid": "abc123",
  "name": "Jazz FM",
  "custom_tags": ["workout", "coding", "focus"]
}
```

### UI

**Tag management:**
```
Station: Jazz FM
Tags: #workout #coding #focus

Commands:
  t: Add tag
  d: Remove tag
  f: Filter by tag
```

---

## v4.0.0: Library/SDK Mode

### Package Structure

```
v4/
├── tera.go              # Public API entry point
├── client.go            # Main client
├── search.go            # Search functions
├── favorites.go         # Favorites management
├── player.go            # Playback control
├── errors.go            # Public error types
├── options.go           # Client options
├── internal/            # Private implementation
│   ├── ui/             # CLI interface
│   ├── api/
│   └── storage/
└── cmd/
    └── tera/
        └── main.go      # CLI that uses public API
```

### Public API

```go
// v4/tera.go
package tera

// Client is the main TERA client
type Client struct {
    cfg    *Config
    api    *api.Client
    player *player.Player
    store  *storage.Storage
}

// New creates a new TERA client
func New(opts ...Option) (*Client, error)

// Search searches for radio stations
func (c *Client) Search(ctx context.Context, query SearchQuery) ([]Station, error)

// Play plays a radio station
func (c *Client) Play(station Station) error

// Stop stops playback
func (c *Client) Stop() error

// Favorites returns the favorites manager
func (c *Client) Favorites() *FavoritesManager

// Close cleans up resources
func (c *Client) Close() error
```

### Options Pattern

```go
// v4/options.go
type Option func(*Client) error

func WithConfigDir(dir string) Option {
    return func(c *Client) error {
        c.cfg.ConfigDir = dir
        return nil
    }
}

func WithVolume(vol int) Option {
    return func(c *Client) error {
        c.player.SetVolume(vol)
        return nil
    }
}

// Usage:
client, err := tera.New(
    tera.WithConfigDir("/custom/path"),
    tera.WithVolume(80),
)
```

---

## Testing Strategy

### Unit Tests
- All packages should have `_test.go` files
- Aim for >70% coverage
- Use table-driven tests

### Integration Tests
- Test config migration end-to-end
- Test player integration with mpv
- Test API integration with Radio Browser

### Manual Testing Checklist
- [ ] Install on Linux
- [ ] Install on macOS (Intel)
- [ ] Install on macOS (ARM)
- [ ] Install on Windows
- [ ] Test migration from v2
- [ ] Test fresh install
- [ ] Test all menu options
- [ ] Test playback

---

## Performance Considerations

### v3.0: Config Loading
- Lazy load config (only when needed)
- Cache parsed config in memory
- Validate on load, not on every access

### v3.2: Statistics
- Update play count async (don't block playback)
- Batch writes to disk (every 5 minutes or on exit)
- Index by station UUID for fast lookups

### v4.0: Library Mode
- Make all operations context-aware
- Support graceful cancellation
- Thread-safe operations

---

## Dependencies

Current dependencies (keep minimal):
- `github.com/charmbracelet/bubbletea` - TUI framework
- `github.com/charmbracelet/bubbles` - TUI components
- `github.com/charmbracelet/lipgloss` - Styling
- `gopkg.in/yaml.v3` - YAML parsing
- `golang.org/x/text` - Text processing

New dependencies for v3:
- `github.com/zalando/go-keyring` - Secure credential storage

Consider adding for v4:
- `golang.org/x/sync/errgroup` - Concurrent operations
- `github.com/stretchr/testify` - Testing utilities

---

## Documentation

Each release should include:
- Updated README.md
- CHANGELOG.md entry
- Migration guide (if breaking changes)
- API documentation (for v4+)

---

**Last Updated:** February 2026  
**In Development:** v3.0.0 (Unified Config + Secure Credentials)
