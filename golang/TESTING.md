# TERA Go Testing Strategy

## Test Coverage Goals

- **Unit Tests:** >80% code coverage
- **Integration Tests:** All major workflows
- **Manual Tests:** Cross-platform verification

## Test Organization

```
tera/
├── internal/
│   ├── api/
│   │   ├── client.go
│   │   ├── client_test.go      # Unit tests
│   │   └── integration_test.go  # API integration
│   ├── storage/
│   │   ├── favorites.go
│   │   └── favorites_test.go
│   └── player/
│       ├── mpv.go
│       └── mpv_test.go
└── testdata/
    ├── stations.json
    └── config.json
```

## Unit Testing

### API Client Tests

```go
// internal/api/client_test.go
package api

import (
    "context"
    "net/http"
    "net/http/httptest"
    "testing"
)

func TestClient_SearchByTag(t *testing.T) {
    // Mock server
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if r.Method != "POST" {
            t.Errorf("Expected POST, got %s", r.Method)
        }
        
        w.WriteHeader(http.StatusOK)
        w.Write([]byte(`[
            {
                "stationuuid": "test-1",
                "name": "Jazz FM",
                "url_resolved": "http://example.com",
                "tags": "jazz",
                "votes": 100
            }
        ]`))
    }))
    defer server.Close()
    
    // Create client with test server
    client := NewClient()
    client.baseURL = server.URL
    
    // Test search
    stations, err := client.SearchByTag(context.Background(), "jazz")
    if err != nil {
        t.Fatalf("SearchByTag failed: %v", err)
    }
    
    if len(stations) != 1 {
        t.Errorf("Expected 1 station, got %d", len(stations))
    }
    
    if stations[0].Name != "Jazz FM" {
        t.Errorf("Expected 'Jazz FM', got '%s'", stations[0].Name)
    }
}

func TestClient_SearchByTag_NetworkError(t *testing.T) {
    client := NewClient()
    client.baseURL = "http://invalid-url-12345.test"
    
    _, err := client.SearchByTag(context.Background(), "jazz")
    if err == nil {
        t.Error("Expected error for invalid URL")
    }
}

func TestClient_SearchByTag_Timeout(t *testing.T) {
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        time.Sleep(2 * time.Second) // Longer than client timeout
    }))
    defer server.Close()
    
    client := NewClient()
    client.httpClient.Timeout = 100 * time.Millisecond
    client.baseURL = server.URL
    
    ctx := context.Background()
    _, err := client.SearchByTag(ctx, "jazz")
    
    if err == nil {
        t.Error("Expected timeout error")
    }
}
```

### Storage Tests

```go
// internal/storage/favorites_test.go
package storage

import (
    "context"
    "encoding/json"
    "os"
    "path/filepath"
    "testing"
    
    "github.com/shinokada/tera/internal/api"
)

func TestStorage_LoadList(t *testing.T) {
    tmpDir := t.TempDir()
    
    // Create test data
    stations := []api.Station{
        {StationUUID: "test-1", Name: "Test 1"},
        {StationUUID: "test-2", Name: "Test 2"},
    }
    data, _ := json.Marshal(stations)
    
    listPath := filepath.Join(tmpDir, "mylist.json")
    os.WriteFile(listPath, data, 0644)
    
    // Test load
    store := NewStorage(tmpDir)
    list, err := store.LoadList(context.Background(), "mylist")
    
    if err != nil {
        t.Fatalf("LoadList failed: %v", err)
    }
    
    if list.Name != "mylist" {
        t.Errorf("Expected name 'mylist', got '%s'", list.Name)
    }
    
    if len(list.Stations) != 2 {
        t.Errorf("Expected 2 stations, got %d", len(list.Stations))
    }
}

func TestStorage_SaveList(t *testing.T) {
    tmpDir := t.TempDir()
    store := NewStorage(tmpDir)
    
    list := &FavoritesList{
        Name: "newlist",
        Stations: []api.Station{
            {StationUUID: "test-1", Name: "Test Station"},
        },
    }
    
    err := store.SaveList(context.Background(), list)
    if err != nil {
        t.Fatalf("SaveList failed: %v", err)
    }
    
    // Verify file exists
    path := filepath.Join(tmpDir, "newlist.json")
    if _, err := os.Stat(path); os.IsNotExist(err) {
        t.Error("File was not created")
    }
    
    // Verify content
    data, _ := os.ReadFile(path)
    var stations []api.Station
    json.Unmarshal(data, &stations)
    
    if len(stations) != 1 {
        t.Errorf("Expected 1 station in file, got %d", len(stations))
    }
}

func TestStorage_AddStation_Duplicate(t *testing.T) {
    tmpDir := t.TempDir()
    store := NewStorage(tmpDir)
    
    station := api.Station{StationUUID: "test-1", Name: "Test"}
    
    // Create list with station
    list := &FavoritesList{
        Name:     "test",
        Stations: []api.Station{station},
    }
    store.SaveList(context.Background(), list)
    
    // Try to add duplicate
    err := store.AddStation(context.Background(), "test", station)
    
    if err == nil {
        t.Error("Expected error for duplicate station")
    }
}
```

### Player Tests

```go
// internal/player/mpv_test.go
package player

import (
    "context"
    "testing"
    "time"
)

func TestMPVPlayer_Play(t *testing.T) {
    // This test requires mpv to be installed
    if testing.Short() {
        t.Skip("Skipping player test in short mode")
    }
    
    player := NewMPVPlayer()
    
    // Use a short test stream
    testURL := "http://example.com/test.mp3"
    
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    
    err := player.Play(ctx, testURL, nil)
    
    // We expect an error since it's a fake URL
    // But it tests that the player starts
    if err == nil {
        t.Error("Expected error for invalid stream")
    }
}

func TestMPVPlayer_Stop(t *testing.T) {
    player := NewMPVPlayer()
    
    // Should handle stop on non-playing player
    err := player.Stop(context.Background())
    if err != nil {
        t.Errorf("Stop on idle player should not error: %v", err)
    }
}
```

## Integration Testing

### Full Workflow Tests

```go
// tests/integration/search_play_save_test.go
//go:build integration

package integration

import (
    "context"
    "testing"
    
    "github.com/shinokada/tera/internal/api"
    "github.com/shinokada/tera/internal/storage"
)

func TestSearchPlaySaveWorkflow(t *testing.T) {
    // Setup
    tmpDir := t.TempDir()
    store := storage.NewStorage(tmpDir)
    client := api.NewClient()
    
    ctx := context.Background()
    
    // 1. Search for stations
    stations, err := client.SearchByTag(ctx, "jazz")
    if err != nil {
        t.Fatalf("Search failed: %v", err)
    }
    
    if len(stations) == 0 {
        t.Skip("No stations found")
    }
    
    // 2. Save to list
    list := &storage.FavoritesList{
        Name:     "integration-test",
        Stations: []api.Station{stations[0]},
    }
    
    err = store.SaveList(ctx, list)
    if err != nil {
        t.Fatalf("SaveList failed: %v", err)
    }
    
    // 3. Load list back
    loaded, err := store.LoadList(ctx, "integration-test")
    if err != nil {
        t.Fatalf("LoadList failed: %v", err)
    }
    
    // 4. Verify
    if len(loaded.Stations) != 1 {
        t.Errorf("Expected 1 station, got %d", len(loaded.Stations))
    }
    
    if loaded.Stations[0].StationUUID != stations[0].StationUUID {
        t.Error("Station UUID mismatch")
    }
}
```

### Run Integration Tests

```bash
# Run all tests
go test ./...

# Run only unit tests
go test -short ./...

# Run integration tests
go test -tags=integration ./tests/integration/...
```

## Table-Driven Tests

```go
func TestStation_TrimName(t *testing.T) {
    tests := []struct {
        name  string
        input string
        want  string
    }{
        {
            name:  "leading spaces",
            input: "  Jazz FM",
            want:  "Jazz FM",
        },
        {
            name:  "trailing spaces",
            input: "Jazz FM  ",
            want:  "Jazz FM",
        },
        {
            name:  "both sides",
            input: "  Jazz FM  ",
            want:  "Jazz FM",
        },
        {
            name:  "no spaces",
            input: "Jazz FM",
            want:  "Jazz FM",
        },
        {
            name:  "internal spaces",
            input: "Jazz  FM",
            want:  "Jazz  FM",
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            station := api.Station{Name: tt.input}
            got := station.TrimName()
            
            if got != tt.want {
                t.Errorf("TrimName() = %q, want %q", got, tt.want)
            }
        })
    }
}
```

## Mocking

### HTTP Client Mock

```go
type MockHTTPClient struct {
    DoFunc func(req *http.Request) (*http.Response, error)
}

func (m *MockHTTPClient) Do(req *http.Request) (*http.Response, error) {
    return m.DoFunc(req)
}

func TestWithMock(t *testing.T) {
    mockClient := &MockHTTPClient{
        DoFunc: func(req *http.Request) (*http.Response, error) {
            return &http.Response{
                StatusCode: 200,
                Body:       io.NopCloser(strings.NewReader(`[]`)),
            }, nil
        },
    }
    
    // Use mock in tests
}
```

### Storage Mock

```go
type MockStorage struct {
    LoadListFunc func(ctx context.Context, name string) (*FavoritesList, error)
    SaveListFunc func(ctx context.Context, list *FavoritesList) error
}

func (m *MockStorage) LoadList(ctx context.Context, name string) (*FavoritesList, error) {
    return m.LoadListFunc(ctx, name)
}

func (m *MockStorage) SaveList(ctx context.Context, list *FavoritesList) error {
    return m.SaveListFunc(ctx, list)
}
```

## Test Coverage

### Generate Coverage Report

```bash
# Run tests with coverage
go test -coverprofile=coverage.out ./...

# View in browser
go tool cover -html=coverage.out

# Show coverage by function
go tool cover -func=coverage.out
```

### Coverage Goals

- **Critical paths:** 100% (API, Storage operations)
- **Business logic:** >90%
- **UI code:** >70%
- **Overall:** >80%

## Continuous Integration

### GitHub Actions Workflow

```yaml
# .github/workflows/test.yml
name: Tests

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      
      - name: Install dependencies
        run: |
          sudo apt-get update
          sudo apt-get install -y mpv
      
      - name: Run tests
        run: go test -v -race -coverprofile=coverage.out ./...
      
      - name: Upload coverage
        uses: codecov/codecov-action@v3
        with:
          file: ./coverage.out
```

## Manual Testing Checklist

### Main Menu
- [ ] All menu options display correctly
- [ ] Quick favorites (10-19) work
- [ ] Number keys navigate correctly
- [ ] Esc/q returns to previous screen

### Play Screen
- [ ] Lists load and display
- [ ] Stations sorted alphabetically
- [ ] Filter works in real-time
- [ ] Play starts mpv correctly
- [ ] Save prompt appears after playback
- [ ] Adding to quick favorites works
- [ ] Duplicate detection works

### Search
- [ ] All search types work
- [ ] Results display correctly
- [ ] Filter works on results
- [ ] Station info displays
- [ ] Play from results works
- [ ] Save to list works
- [ ] Duplicate prevention works

### List Management
- [ ] Create list works
- [ ] Delete list works (with protection for My-favorites)
- [ ] Rename list works (with protection)
- [ ] Show all lists works
- [ ] Input validation works
- [ ] Navigation shortcuts (0, 00) work

### Gist Integration
- [ ] Token setup works
- [ ] Token validation works
- [ ] Create gist uploads all files
- [ ] My gists displays correctly
- [ ] Update gist description works
- [ ] Delete gist works (with confirmation)
- [ ] Recover from gist clones correctly
- [ ] Error messages are helpful

### Error Handling
- [ ] Network errors display helpful messages
- [ ] File errors show clear reasons
- [ ] API errors suggest fixes
- [ ] Missing mpv shows installation help
- [ ] Invalid tokens show setup instructions

## Performance Testing

### Benchmarks

```go
func BenchmarkStation_TrimName(b *testing.B) {
    station := api.Station{Name: "  Test Station  "}
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _ = station.TrimName()
    }
}

func BenchmarkStorage_LoadList(b *testing.B) {
    tmpDir := b.TempDir()
    store := NewStorage(tmpDir)
    
    // Setup test data
    list := &FavoritesList{
        Name:     "bench",
        Stations: make([]api.Station, 100),
    }
    store.SaveList(context.Background(), list)
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, _ = store.LoadList(context.Background(), "bench")
    }
}
```

### Run Benchmarks

```bash
go test -bench=. -benchmem ./...
```

## Test Data

### testdata/stations.json

```json
[
  {
    "stationuuid": "test-uuid-1",
    "name": "Test Jazz Station",
    "url_resolved": "http://example.com/stream1",
    "tags": "jazz,smooth",
    "country": "United States",
    "votes": 100,
    "codec": "MP3",
    "bitrate": 128
  },
  {
    "stationuuid": "test-uuid-2",
    "name": "Test Rock Station",
    "url_resolved": "http://example.com/stream2",
    "tags": "rock,classic",
    "country": "United Kingdom",
    "votes": 200,
    "codec": "AAC",
    "bitrate": 192
  }
]
```

## Testing Best Practices

1. **Isolate tests** - Each test should be independent
2. **Use table-driven tests** - For testing multiple inputs
3. **Test edge cases** - Empty inputs, large data, etc.
4. **Mock external dependencies** - HTTP, file system, time
5. **Clean up** - Use t.TempDir() and defer for cleanup
6. **Clear test names** - Describe what is being tested
7. **Test errors** - Not just happy paths
8. **Use contexts** - For cancellation and timeouts
9. **Avoid sleep** - Use channels or contexts instead
10. **Document why** - Comment unusual test scenarios
