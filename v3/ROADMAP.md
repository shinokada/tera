# TERA v3 Technical Roadmap

This document contains technical implementation details for v3 development.

## Architecture (v3.0.0)

```
v3/
├── cmd/tera/main.go           # CLI entry point
├── internal/
│   ├── api/                   # Radio Browser API client
│   ├── blocklist/             # Station blocking
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
- `blocklist.json`

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
    Data    DataConfig `yaml:"data"`
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

type DataConfig struct {
    BlockedStations []string `yaml:"blocked_stations"`
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

data:
  blocked_stations: []
```

### Implementation Steps

**Week 1: Config package**
1. Create `v3/internal/config/` package
2. Define `Config` struct hierarchy
3. Implement `Load()` and `Save()` functions
4. Write unit tests

**Week 2: Migration logic**
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
    
    // Read old blocklist.json
    if blocked, err := readV2Blocklist(v2ConfigDir); err == nil {
        cfg.Data.BlockedStations = blocked
    }
    
    // Backup old files
    backupV2Configs(v2ConfigDir)
    
    return cfg, nil
}
```

**Week 3: Integration**
1. Update `internal/storage/config.go` to use new config package
2. Update all packages that read config (theme, player, ui)
3. Add migration check in `cmd/tera/main.go`

**Week 4: Testing & Release**
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
**In Development:** v3.0.0  (Unified Config)
