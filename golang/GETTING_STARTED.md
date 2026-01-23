# TERA Go - Developer Quick Start

## Prerequisites

- Go 1.21+
- mpv (audio player)
- git
- Make (optional)

## Project Setup

```bash
# Create project structure
cd golang
go mod init github.com/shinokada/tera

# Install dependencies
go get github.com/charmbracelet/bubbletea@latest
go get github.com/charmbracelet/bubbles@latest
go get github.com/charmbracelet/lipgloss@latest
go get github.com/google/go-github/v57/github@latest
go get golang.org/x/oauth2@latest
go get github.com/zalando/go-keyring@latest
go get github.com/gofrs/flock@latest
```

## Directory Structure

```
tera/
├── cmd/tera/
│   └── main.go              # Entry point
├── internal/
│   ├── api/
│   │   ├── client.go        # HTTP client
│   │   ├── models.go        # Data structures
│   │   └── search.go        # Search methods
│   ├── player/
│   │   └── mpv.go           # MPV controller
│   ├── storage/
│   │   ├── favorites.go     # List management
│   │   ├── config.go        # Config handling
│   │   └── lock.go          # File locking
│   ├── gist/
│   │   ├── client.go        # GitHub API
│   │   ├── token.go         # Token management
│   │   └── metadata.go      # Local tracking
│   └── ui/
│       ├── app.go           # Main app model
│       ├── menu.go          # Main menu
│       ├── play.go          # Play screen
│       ├── search.go        # Search screens
│       ├── list.go          # List management
│       ├── gist.go          # Gist screens
│       ├── styles.go        # Lipgloss styles
│       └── components/      # Reusable components
│           ├── list.go
│           ├── stationinfo.go
│           └── help.go
├── pkg/
│   └── utils/
│       └── paths.go         # Path utilities
├── go.mod
├── go.sum
├── Makefile
└── README.md
```

## Phase 1: Data Models (Start Here)

### Step 1: Create internal/api/models.go

```go
package api

import "time"

// Station represents a radio station from Radio Browser API
type Station struct {
    StationUUID  string `json:"stationuuid"`
    Name         string `json:"name"`
    URLResolved  string `json:"url_resolved"`
    Tags         string `json:"tags"`
    Country      string `json:"country"`
    CountryCode  string `json:"countrycode"`
    State        string `json:"state"`
    Language     string `json:"language"`
    Votes        int    `json:"votes"`
    Codec        string `json:"codec"`
    Bitrate      int    `json:"bitrate"`
}

// TrimName returns station name with whitespace trimmed
func (s *Station) TrimName() string {
    return strings.TrimSpace(s.Name)
}
```

### Step 2: Create internal/storage/models.go

```go
package storage

import "github.com/shinokada/tera/internal/api"

// FavoritesList represents a collection of favorite stations
type FavoritesList struct {
    Name     string        `json:"-"`
    Stations []api.Station `json:"stations"`
}

// Config represents application configuration
type Config struct {
    FavoritePath string       `json:"favorite_path"`
    CachePath    string       `json:"cache_path"`
    LastPlayed   *api.Station `json:"last_played,omitempty"`
}
```

### Step 3: Write Tests

```go
// internal/api/models_test.go
package api

import (
    "encoding/json"
    "testing"
)

func TestStation_Unmarshal(t *testing.T) {
    jsonData := `{
        "stationuuid": "test-123",
        "name": "  Jazz FM  ",
        "url_resolved": "http://example.com",
        "votes": 100,
        "codec": "MP3",
        "bitrate": 128
    }`
    
    var station Station
    err := json.Unmarshal([]byte(jsonData), &station)
    if err != nil {
        t.Fatalf("Failed to unmarshal: %v", err)
    }
    
    if station.StationUUID != "test-123" {
        t.Errorf("Expected UUID test-123, got %s", station.StationUUID)
    }
    
    if station.TrimName() != "Jazz FM" {
        t.Errorf("Expected trimmed name 'Jazz FM', got '%s'", station.TrimName())
    }
}
```

## Phase 2: Storage Layer

### Step 1: Create internal/storage/favorites.go

```go
package storage

import (
    "context"
    "encoding/json"
    "os"
    "path/filepath"
)

type Storage struct {
    favoritePath string
}

func NewStorage(favoritePath string) *Storage {
    return &Storage{favoritePath: favoritePath}
}

func (s *Storage) LoadList(ctx context.Context, name string) (*FavoritesList, error) {
    path := filepath.Join(s.favoritePath, name+".json")
    
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, err
    }
    
    var stations []api.Station
    if err := json.Unmarshal(data, &stations); err != nil {
        return nil, err
    }
    
    return &FavoritesList{
        Name:     name,
        Stations: stations,
    }, nil
}
```

### Step 2: Write Storage Tests

```go
// internal/storage/favorites_test.go
func TestStorage_LoadList(t *testing.T) {
    // Create temp directory
    tmpDir := t.TempDir()
    
    // Write test file
    testData := []api.Station{
        {StationUUID: "test-1", Name: "Test Station"},
    }
    data, _ := json.Marshal(testData)
    os.WriteFile(filepath.Join(tmpDir, "test.json"), data, 0644)
    
    // Test load
    store := NewStorage(tmpDir)
    list, err := store.LoadList(context.Background(), "test")
    
    if err != nil {
        t.Fatalf("LoadList failed: %v", err)
    }
    
    if len(list.Stations) != 1 {
        t.Errorf("Expected 1 station, got %d", len(list.Stations))
    }
}
```

## Phase 3: API Client

### Step 1: Create internal/api/client.go

```go
package api

import (
    "bytes"
    "context"
    "encoding/json"
    "net/http"
    "net/url"
    "time"
)

const baseURL = "https://de1.api.radio-browser.info/json/stations"

type Client struct {
    httpClient *http.Client
}

func NewClient() *Client {
    return &Client{
        httpClient: &http.Client{
            Timeout: 10 * time.Second,
        },
    }
}

func (c *Client) SearchByTag(ctx context.Context, tag string) ([]Station, error) {
    form := url.Values{}
    form.Add("tag", tag)
    
    return c.doSearch(ctx, form)
}

func (c *Client) doSearch(ctx context.Context, form url.Values) ([]Station, error) {
    req, err := http.NewRequestWithContext(
        ctx,
        "POST",
        baseURL+"/search",
        bytes.NewBufferString(form.Encode()),
    )
    if err != nil {
        return nil, err
    }
    
    req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
    
    resp, err := c.httpClient.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    var stations []Station
    if err := json.NewDecoder(resp.Body).Decode(&stations); err != nil {
        return nil, err
    }
    
    return stations, nil
}
```

## Phase 4: Basic Bubble Tea App

### Step 1: Create cmd/tera/main.go

```go
package main

import (
    "fmt"
    "os"
    
    tea "github.com/charmbracelet/bubbletea"
    "github.com/shinokada/tera/internal/ui"
)

func main() {
    p := tea.NewProgram(ui.NewApp(), tea.WithAltScreen())
    
    if _, err := p.Run(); err != nil {
        fmt.Printf("Error: %v\n", err)
        os.Exit(1)
    }
}
```

### Step 2: Create internal/ui/app.go

```go
package ui

import (
    tea "github.com/charmbracelet/bubbletea"
)

type App struct {
    screen Screen
}

type Screen int

const (
    ScreenMainMenu Screen = iota
)

func NewApp() App {
    return App{
        screen: ScreenMainMenu,
    }
}

func (a App) Init() tea.Cmd {
    return nil
}

func (a App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.String() {
        case "q", "ctrl+c":
            return a, tea.Quit
        }
    }
    return a, nil
}

func (a App) View() string {
    return "TERA - Press q to quit\n"
}
```

### Step 3: Test Run

```bash
go run cmd/tera/main.go
```

## Development Workflow

### Build
```bash
go build -o tera cmd/tera/main.go
```

### Run Tests
```bash
go test ./...
```

### Run with Coverage
```bash
go test -cover ./...
```

### Format Code
```bash
go fmt ./...
```

### Lint
```bash
golangci-lint run
```

## Makefile

```makefile
.PHONY: build test clean run

build:
	go build -o tera cmd/tera/main.go

test:
	go test -v ./...

coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

run:
	go run cmd/tera/main.go

clean:
	rm -f tera coverage.out

install:
	go install cmd/tera/main.go
```

## Next Steps

1. ✅ Complete Phase 1 (Data Models)
2. ✅ Complete Phase 2 (Storage)
3. ✅ Complete Phase 3 (API Client)
4. ✅ Complete Phase 4 (Basic UI)
5. → Continue with Play Screen
6. → Add Search functionality
7. → Implement List management
8. → Add Gist integration

## Testing Strategy

### Unit Tests
- Test each function in isolation
- Mock external dependencies
- Table-driven tests

### Integration Tests
- Test full workflows
- Use temp directories
- Mock HTTP responses

### Example Test Pattern

```go
func TestFeature(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    string
        wantErr bool
    }{
        {
            name:    "valid input",
            input:   "test",
            want:    "expected",
            wantErr: false,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := Feature(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("got error %v, wantErr %v", err, tt.wantErr)
            }
            if got != tt.want {
                t.Errorf("got %v, want %v", got, tt.want)
            }
        })
    }
}
```

## Common Issues

### Import Paths
Use full import paths:
```go
import "github.com/shinokada/tera/internal/api"
```

### Context
Always pass context:
```go
func (c *Client) Search(ctx context.Context, query string) error
```

### Error Handling
Always handle errors:
```go
data, err := os.ReadFile(path)
if err != nil {
    return fmt.Errorf("read file: %w", err)
}
```

## Resources

- [Bubble Tea Tutorial](https://github.com/charmbracelet/bubbletea/tree/master/tutorials)
- [Bubbles Components](https://github.com/charmbracelet/bubbles)
- [Lipgloss Examples](https://github.com/charmbracelet/lipgloss)
- [Go Testing](https://go.dev/doc/tutorial/add-a-test)
