# TERA Go Technical Approach

## Architecture Overview

TERA Go will follow a clean architecture pattern with clear separation of concerns:

```text
tera/
├── cmd/tera/           # Application entry point
│   └── main.go
├── internal/
│   ├── api/            # Radio Browser API client
│   │   ├── client.go
│   │   ├── models.go
│   │   └── search.go
│   ├── player/         # MPV player integration
│   │   ├── player.go
│   │   └── controller.go
│   ├── storage/        # Local storage & favorites
│   │   ├── favorites.go
│   │   ├── lists.go
│   │   └── config.go
│   ├── gist/           # GitHub Gist integration
│   │   ├── client.go
│   │   ├── sync.go
│   │   └── token.go
│   └── ui/             # Bubble Tea TUI
│       ├── app.go
│       ├── menu.go
│       ├── search.go
│       ├── play.go
│       ├── list.go
│       ├── gist.go
│       └── components/
│           ├── stationlist.go
│           ├── stationinfo.go
│           └── help.go
├── pkg/                # Public reusable packages
│   └── utils/
│       ├── colors.go
│       └── validation.go
├── go.mod
└── go.sum
```

## Core Technology Stack

### UI Framework: Bubble Tea + Bubbles + Lipgloss
- **Bubble Tea**: The Elm-architecture TUI framework for building interactive terminal UIs
- **Bubbles**: Pre-built TUI components (lists, text inputs, spinners, etc.)
- **Lipgloss**: Styling library for beautiful terminal output

### HTTP Client: Standard library + retries
- Use `net/http` for Radio Browser API
- Add exponential backoff for retries
- Connection pooling for performance

### JSON Handling: encoding/json
- Native Go JSON marshaling/unmarshaling
- Faster than external dependencies
- Type-safe station data structures

### Process Management: os/exec
- Control MPV player subprocess
- Handle graceful shutdown
- Pipe management for player output

### Storage: File-based JSON
- Continue using JSON files for favorites (backward compatible)
- Path: `~/.config/tera/favorite/`
- Clean file locking for concurrent access

### GitHub Integration: go-github
- Use official GitHub API client
- Secure token storage in keychain/credential manager
- CRUD operations for gists

## Key Design Patterns

### 1. Model-View-Update (Bubble Tea's Core Pattern)

```go
type Model struct {
    currentView   View
    stations      []Station
    selectedIndex int
    player        *Player
    // ... other state
}

func (m Model) Init() tea.Cmd {
    return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    // Handle messages (keyboard, API responses, etc.)
}

func (m Model) View() string {
    // Render the current state
}
```

### 2. Command Pattern for Async Operations
All I/O and long-running operations return `tea.Cmd`:
- API searches
- File I/O
- MPV player control
- Gist operations

### 3. Message-Based Communication
Use custom messages for events:

```go
type SearchCompleteMsg struct {
    Stations []Station
    Error    error
}

type PlayerStateMsg struct {
    Playing bool
    Station *Station
}

type GistSyncMsg struct {
    Success bool
    Error   error
}
```

### 4. Interface-Based Dependencies
Enable testing and mocking:

```go
type APIClient interface {
    SearchByTag(tag string) ([]Station, error)
    SearchByName(name string) ([]Station, error)
    // ... other methods
}

type Storage interface {
    LoadFavorites() ([]List, error)
    SaveFavorites(lists []List) error
    // ... other methods
}

type Player interface {
    Play(url string) error
    Stop() error
    IsPlaying() bool
}
```

## Data Models

### Station (Radio Browser API Response)
```go
type Station struct {
    StationUUID  string `json:"stationuuid"`
    Name         string `json:"name"`
    URL          string `json:"url"`
    URLResolved  string `json:"url_resolved"`
    Homepage     string `json:"homepage"`
    Favicon      string `json:"favicon"`
    Tags         string `json:"tags"`
    Country      string `json:"country"`
    CountryCode  string `json:"countrycode"`
    State        string `json:"state"`
    Language     string `json:"language"`
    Votes        int    `json:"votes"`
    Codec        string `json:"codec"`
    Bitrate      int    `json:"bitrate"`
}
```

### Favorites List
```go
type FavoritesList struct {
    Name     string    `json:"name"`
    Stations []Station `json:"stations"`
    Updated  time.Time `json:"updated"`
}
```

### Configuration
```go
type Config struct {
    FavoritePath string
    CachePath    string
    PlayerPath   string  // MPV binary path
    LastPlayed   *Station
    GistID       string
}
```

## State Management

### Application State
```go
type AppState struct {
    // Navigation
    CurrentScreen Screen
    PreviousScreen Screen
    
    // Data
    SearchResults []Station
    FavoriteLists []FavoritesList
    QuickFavorites []Station  // My-favorites.json
    
    // Player
    Player        *MPVPlayer
    NowPlaying    *Station
    
    // UI State
    SearchQuery   string
    SelectedIndex int
    Loading       bool
    ErrorMsg      string
    
    // Config
    Config        *Config
}
```

### Screen Types (State Machine)
```go
type Screen int

const (
    ScreenMainMenu Screen = iota
    ScreenPlay
    ScreenSearch
    ScreenSearchResults
    ScreenList
    ScreenGist
    ScreenStationInfo
)
```

## API Client Design

### Radio Browser API Client
```go
type RadioBrowserClient struct {
    httpClient *http.Client
    baseURL    string
}

func NewRadioBrowserClient() *RadioBrowserClient {
    return &RadioBrowserClient{
        httpClient: &http.Client{Timeout: 10 * time.Second},
        baseURL:    "https://de1.api.radio-browser.info/json/stations",
    }
}

func (c *RadioBrowserClient) SearchByTag(tag string) ([]Station, error) {
    // POST to /search with tag parameter
}

func (c *RadioBrowserClient) SearchByName(name string) ([]Station, error) {
    // POST to /search with name parameter
}

// Advanced search with multiple criteria
func (c *RadioBrowserClient) Search(params SearchParams) ([]Station, error) {
    // Support multiple filters
}
```

### Error Handling Strategy
```go
type APIError struct {
    StatusCode int
    Message    string
    Err        error
}

func (e *APIError) Error() string {
    return fmt.Sprintf("API error (status %d): %s", e.StatusCode, e.Message)
}

// Retry logic with exponential backoff
func (c *RadioBrowserClient) doWithRetry(req *http.Request, maxRetries int) (*http.Response, error) {
    // Implement exponential backoff
}
```

## MPV Player Integration

### Player Controller
```go
type MPVPlayer struct {
    cmd      *exec.Cmd
    stdin    io.WriteCloser
    stdout   io.ReadCloser
    playing  bool
    station  *Station
    mu       sync.Mutex
}

func NewMPVPlayer() *MPVPlayer {
    return &MPVPlayer{}
}

func (p *MPVPlayer) Play(url string) error {
    p.mu.Lock()
    defer p.mu.Unlock()
    
    // Stop any current playback
    if p.playing {
        p.stop()
    }
    
    // Start new MPV process
    p.cmd = exec.Command("mpv", "--no-video", "--really-quiet", url)
    // ... setup pipes and error handling
}

func (p *MPVPlayer) Stop() error {
    // Gracefully terminate MPV
}

func (p *MPVPlayer) IsPlaying() bool {
    p.mu.Lock()
    defer p.mu.Unlock()
    return p.playing
}
```

### Player State Management
Use channels to communicate player events back to the main UI:

```go
type PlayerEvent int

const (
    PlayerStarted PlayerEvent = iota
    PlayerStopped
    PlayerError
)

type PlayerMsg struct {
    Event   PlayerEvent
    Station *Station
    Error   error
}
```

## Storage Layer

### Favorites Management
```go
type FavoritesStorage struct {
    basePath string
}

func (s *FavoritesStorage) LoadList(name string) (*FavoritesList, error) {
    path := filepath.Join(s.basePath, name+".json")
    data, err := os.ReadFile(path)
    // ... unmarshal and return
}

func (s *FavoritesStorage) SaveList(list *FavoritesList) error {
    // Atomic write with temp file + rename
    // Ensures data integrity
}

func (s *FavoritesStorage) AddStation(listName string, station Station) error {
    // Load, append, save
    // Check for duplicates by StationUUID
}

func (s *FavoritesStorage) RemoveStation(listName string, stationUUID string) error {
    // Load, filter, save
}
```

### File Locking
```go
import "github.com/gofrs/flock"

func (s *FavoritesStorage) withLock(path string, fn func() error) error {
    lock := flock.New(path + ".lock")
    if err := lock.Lock(); err != nil {
        return err
    }
    defer lock.Unlock()
    
    return fn()
}
```

## GitHub Gist Integration

### Gist Client
```go
type GistClient struct {
    client *github.Client
    token  string
}

func NewGistClient(token string) *GistClient {
    ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
    tc := oauth2.NewClient(context.Background(), ts)
    
    return &GistClient{
        client: github.NewClient(tc),
        token:  token,
    }
}

func (g *GistClient) CreateGist(files map[string]string) (*github.Gist, error) {
    // Create gist with all favorite list files
}

func (g *GistClient) UpdateGist(gistID string, files map[string]string) error {
    // Update existing gist
}

func (g *GistClient) FetchGist(gistID string) (map[string]string, error) {
    // Download and parse gist files
}
```

### Token Management
Use secure storage for GitHub tokens:

```go
import "github.com/zalando/go-keyring"

const serviceName = "tera"
const keyringKey = "github-token"

func SaveToken(token string) error {
    return keyring.Set(serviceName, keyringKey, token)
}

func LoadToken() (string, error) {
    return keyring.Get(serviceName, keyringKey)
}

func DeleteToken() error {
    return keyring.Delete(serviceName, keyringKey)
}
```

Fallback to encrypted file storage on systems without keyring support.

## UI Components with Bubble Tea

### Main Menu
```go
type MainMenu struct {
    list     list.Model
    stations []Station  // Quick favorites
}

func (m MainMenu) Init() tea.Cmd {
    return loadQuickFavorites
}

func (m MainMenu) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.String() {
        case "1":
            return m, switchToPlayScreen
        case "2":
            return m, switchToSearchScreen
        // ... other menu options
        }
    case QuickFavoritesLoaded:
        m.stations = msg.Stations
        // Update list with quick play options
    }
    
    return m, nil
}

func (m MainMenu) View() string {
    return lipgloss.JoinVertical(
        lipgloss.Left,
        titleStyle.Render("TERA MAIN MENU"),
        m.list.View(),
    )
}
```

### Search Screen
```go
type SearchScreen struct {
    textInput    textinput.Model
    searchType   SearchType  // Tag, Name, Country, etc.
    loading      bool
    results      []Station
    resultList   list.Model
}

// Implement Init, Update, View
```

### Station List Component
```go
type StationList struct {
    list      list.Model
    stations  []Station
    selected  int
}

// Custom list item rendering with station info
type stationDelegate struct{}

func (d stationDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
    station := item.(Station)
    
    str := fmt.Sprintf(
        "%s\n  %s • %s • %d kbps",
        station.Name,
        station.Country,
        station.Tags,
        station.Bitrate,
    )
    
    fmt.Fprint(w, str)
}
```

### Progress Indicators
```go
import "github.com/charmbracelet/bubbles/spinner"

type LoadingModel struct {
    spinner spinner.Model
    message string
}

func NewLoadingModel(msg string) LoadingModel {
    s := spinner.New()
    s.Spinner = spinner.Dot
    
    return LoadingModel{
        spinner: s,
        message: msg,
    }
}
```

## Keyboard Navigation System

### Global Key Bindings
```go
type KeyMap struct {
    Up       key.Binding
    Down     key.Binding
    PageUp   key.Binding
    PageDown key.Binding
    Home     key.Binding
    End      key.Binding
    Enter    key.Binding
    Back     key.Binding
    Quit     key.Binding
    Help     key.Binding
    Search   key.Binding
}

var DefaultKeyMap = KeyMap{
    Up: key.NewBinding(
        key.WithKeys("up", "k"),
        key.WithHelp("↑/k", "move up"),
    ),
    Down: key.NewBinding(
        key.WithKeys("down", "j"),
        key.WithHelp("↓/j", "move down"),
    ),
    // ... other bindings
}
```

### Context-Specific Keys
Each screen can have additional key bindings:

```go
// In search results screen
func (m SearchResults) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.String() {
        case "p":
            // Play selected station
        case "s":
            // Save to favorites
        case "i":
            // Show station info
        }
    }
}
```

## Performance Optimizations

### 1. Connection Pooling
```go
var defaultClient = &http.Client{
    Transport: &http.Transport{
        MaxIdleConns:        10,
        MaxIdleConnsPerHost: 10,
        IdleConnTimeout:     30 * time.Second,
    },
    Timeout: 10 * time.Second,
}
```

### 2. Caching Search Results
```go
type SearchCache struct {
    cache map[string]cachedResult
    mu    sync.RWMutex
    ttl   time.Duration
}

type cachedResult struct {
    stations  []Station
    timestamp time.Time
}

func (c *SearchCache) Get(query string) ([]Station, bool) {
    c.mu.RLock()
    defer c.mu.RUnlock()
    
    result, ok := c.cache[query]
    if !ok || time.Since(result.timestamp) > c.ttl {
        return nil, false
    }
    
    return result.stations, true
}
```

### 3. Lazy Loading for Large Lists
Only render visible items in the viewport.

### 4. Debouncing Search Input
```go
func debounce(d time.Duration) func(func()) tea.Cmd {
    return func(fn func()) tea.Cmd {
        return tea.Tick(d, func(time.Time) tea.Msg {
            fn()
            return nil
        })
    }
}
```

## Error Handling Strategy

### User-Facing Errors
```go
type UserError struct {
    Title   string
    Message string
    Hint    string  // Suggestion for resolution
}

func (e UserError) Error() string {
    return e.Message
}
```

### Error Screen Component
```go
type ErrorScreen struct {
    err     UserError
    canRetry bool
}

func (e ErrorScreen) View() string {
    return lipgloss.JoinVertical(
        lipgloss.Left,
        errorTitleStyle.Render(e.err.Title),
        errorMessageStyle.Render(e.err.Message),
        hintStyle.Render(e.err.Hint),
    )
}
```

## Testing Strategy

### Unit Tests
- Test each component in isolation
- Mock API client, storage, and player
- Table-driven tests for business logic

### Integration Tests
- Test full user flows
- Use test doubles for external dependencies
- Verify state transitions

### Example Test Structure
```go
func TestSearchByTag(t *testing.T) {
    tests := []struct {
        name     string
        tag      string
        want     []Station
        wantErr  bool
    }{
        {
            name: "valid tag",
            tag:  "jazz",
            want: []Station{/* ... */},
            wantErr: false,
        },
        // ... more cases
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            client := NewRadioBrowserClient()
            got, err := client.SearchByTag(tt.tag)
            // ... assertions
        })
    }
}
```

## Migration Strategy

### Backward Compatibility
- Read existing JSON favorite files
- Support old file format
- Graceful migration to new format if needed

### Data Migration
```go
func MigrateFromBash() error {
    // Check for old ~/.config/tera directory
    // Convert old favorite files to new format
    // Preserve user data
}
```

## Build and Distribution

### Cross-Compilation
```bash
# macOS
GOOS=darwin GOARCH=amd64 go build -o tera-darwin-amd64
GOOS=darwin GOARCH=arm64 go build -o tera-darwin-arm64

# Linux
GOOS=linux GOARCH=amd64 go build -o tera-linux-amd64

# Windows
GOOS=windows GOARCH=amd64 go build -o tera-windows-amd64.exe
```

### Release Artifacts
- Single binary (no runtime dependencies except mpv)
- GitHub releases with checksums
- Homebrew formula for easy installation

### Dependencies
Only external dependencies users need:
- `mpv` - audio player (already required)
- Internet connection for Radio Browser API

All Go dependencies are bundled in the binary.
