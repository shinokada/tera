# TERA Go Implementation Plan

## Project Setup

### Phase 0: Repository & Environment Setup
**Duration:** 1-2 days

#### Tasks:
1. **Initialize Go Module**
   ```bash
   cd golang
   go mod init github.com/shinokada/tera
   ```

2. **Setup Project Structure**
   ```
   tera/
   ├── cmd/tera/
   │   └── main.go
   ├── internal/
   │   ├── api/
   │   ├── player/
   │   ├── storage/
   │   ├── gist/
   │   └── ui/
   ├── pkg/
   │   └── utils/
   ├── testdata/
   ├── docs/
   ├── .gitignore
   ├── Makefile
   ├── go.mod
   └── README.md
   ```

3. **Install Core Dependencies**
   ```bash
   go get github.com/charmbracelet/bubbletea
   go get github.com/charmbracelet/bubbles
   go get github.com/charmbracelet/lipgloss
   go get github.com/google/go-github/v57/github
   go get github.com/zalando/go-keyring
   go get github.com/gofrs/flock
   ```

4. **Setup Development Tools**
   - golangci-lint for linting
   - Create Makefile for common tasks
   - Setup testing framework

#### Deliverables:
- ✅ Working Go project structure
- ✅ Dependencies installed
- ✅ Basic Makefile with build/test targets
- ✅ README with development setup instructions

---

## Phase 1: Foundation - Core Data Models & Storage

**Duration:** 3-4 days

### 1.1 Data Models
**File:** `internal/api/models.go`

Create core data structures:
```go
type Station struct {
    StationUUID  string `json:"stationuuid"`
    Name         string `json:"name"`
    URLResolved  string `json:"url_resolved"`
    Tags         string `json:"tags"`
    Country      string `json:"country"`
    Votes        int    `json:"votes"`
    Codec        string `json:"codec"`
    Bitrate      int    `json:"bitrate"`
}

type FavoritesList struct {
    Name     string    `json:"name"`
    Stations []Station `json:"stations"`
}

type Config struct {
    FavoritePath   string
    CachePath      string
    LastPlayed     *Station
}
```

**Tests:**
- JSON marshaling/unmarshaling
- Validation logic

### 1.2 Storage Layer
**File:** `internal/storage/favorites.go`

Implement local file-based storage:
```go
type FavoritesStorage struct {
    basePath string
}

func (s *FavoritesStorage) LoadList(name string) (*FavoritesList, error)
func (s *FavoritesStorage) SaveList(list *FavoritesList) error
func (s *FavoritesStorage) GetAllLists() ([]string, error)
func (s *FavoritesStorage) AddStation(listName string, station Station) error
func (s *FavoritesStorage) RemoveStation(listName, stationUUID string) error
```

**Tests:**
- File operations
- Concurrent access
- Error handling
- Migration from bash format

#### Deliverables:
- ✅ Core data models
- ✅ Storage layer with full CRUD
- ✅ Unit tests (>80% coverage)
- ✅ Backward compatibility with bash version

---

## Phase 2: API Client - Radio Browser Integration

**Duration:** 3-4 days

### 2.1 HTTP Client
**File:** `internal/api/client.go`

```go
type RadioBrowserClient struct {
    httpClient *http.Client
    baseURL    string
}

func NewRadioBrowserClient() *RadioBrowserClient
func (c *RadioBrowserClient) SearchByTag(tag string) ([]Station, error)
func (c *RadioBrowserClient) SearchByName(name string) ([]Station, error)
func (c *RadioBrowserClient) SearchByCountry(code string) ([]Station, error)
func (c *RadioBrowserClient) SearchAdvanced(params SearchParams) ([]Station, error)
```

### 2.2 Search Implementation
**File:** `internal/api/search.go`

Features:
- POST requests with form data
- Error handling with retries
- Response parsing
- Rate limiting

**Tests:**
- Mock HTTP responses
- Test all search types
- Error scenarios
- Timeout handling

#### Deliverables:
- ✅ Working API client
- ✅ All search methods implemented
- ✅ Retry logic with exponential backoff
- ✅ Integration tests with test server
- ✅ API client documentation

---

## Phase 3: MPV Player Integration

**Duration:** 2-3 days

### 3.1 Player Controller
**File:** `internal/player/player.go`

```go
type MPVPlayer struct {
    cmd     *exec.Cmd
    playing bool
    station *Station
    mu      sync.Mutex
}

func NewMPVPlayer() *MPVPlayer
func (p *MPVPlayer) Play(url string) error
func (p *MPVPlayer) Stop() error
func (p *MPVPlayer) IsPlaying() bool
func (p *MPVPlayer) GetCurrentStation() *Station
```

### 3.2 Process Management
**File:** `internal/player/controller.go`

Features:
- Start/stop MPV process
- Signal handling (SIGTERM, SIGINT)
- Output capture for debugging
- Cleanup on exit

**Tests:**
- Mock MPV process
- State transitions
- Error handling
- Cleanup verification

#### Deliverables:
- ✅ MPV player controller
- ✅ Graceful start/stop
- ✅ Unit tests
- ✅ Signal handling

---

## Phase 4: Basic TUI - Main Menu & Navigation

**Duration:** 5-6 days

### 4.1 Application Shell
**File:** `internal/ui/app.go`

```go
type App struct {
    currentScreen Screen
    state         *AppState
    keyMap        KeyMap
}

func (a *App) Init() tea.Cmd
func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd)
func (a *App) View() string
```

### 4.2 Main Menu Screen
**File:** `internal/ui/menu.go`

Features:
- Menu with fzf-style selection
- Quick play favorites (items 10+)
- Keyboard shortcuts (1-6, 0)
- Dynamic menu based on favorites

### 4.3 Navigation System
**File:** `internal/ui/navigation.go`

Implement screen transitions:
- Main Menu → Play
- Main Menu → Search
- Main Menu → List Management
- Main Menu → Gist
- Back navigation stack

### 4.4 Styling
**File:** `internal/ui/styles.go`

Use Lipgloss to create:
- Color scheme matching bash version (cyan, yellow, red, green)
- Title styles
- List item styles
- Error/success message styles

**Tests:**
- Screen transitions
- Key handling
- State updates

#### Deliverables:
- ✅ Working main menu
- ✅ Navigation between screens
- ✅ Consistent styling
- ✅ Help system

---

## Phase 5: Play & Search Screens

**Duration:** 6-7 days

### 5.1 Play Screen
**File:** `internal/ui/play.go`

Features:
- List favorites
- Filter/search within list
- Play selected station
- Show station info
- Save to quick favorites

```go
type PlayScreen struct {
    lists        []string
    selectedList string
    stations     list.Model
    player       *MPVPlayer
}
```

### 5.2 Search Menu
**File:** `internal/ui/search.go`

Implement search types:
- Tag search
- Name search
- Language search
- Country code search
- State search
- Advanced search

### 5.3 Search Results Screen
**File:** `internal/ui/search_results.go`

Features:
- Display search results in list
- Preview station info
- Play station
- Save to favorites
- Show loading indicator during search

### 5.4 Station Info Component
**File:** `internal/ui/components/stationinfo.go`

Reusable component to display:
- Station name
- Tags
- Country
- Votes
- Codec
- Bitrate

**Tests:**
- List filtering
- Search query handling
- Player integration
- Save to favorites

#### Deliverables:
- ✅ Play from favorites
- ✅ Search functionality
- ✅ Station info display
- ✅ Save to favorites

---

## Phase 6: List Management

**Duration:** 4-5 days

### 6.1 List Menu Screen
**File:** `internal/ui/list.go`

CRUD operations:
- Create new list
- Delete list
- Rename list
- Show all lists

### 6.2 Station Management
**File:** `internal/ui/station_management.go`

Features:
- Add station to list
- Remove station from list
- Move station between lists
- Prevent duplicates (by UUID)

### 6.3 Input Forms
**File:** `internal/ui/components/input.go`

Use bubbles textinput for:
- List name input
- Search queries
- Confirmations

**Tests:**
- List CRUD operations
- Duplicate detection
- Input validation

#### Deliverables:
- ✅ Full list management
- ✅ Station organization
- ✅ Input validation
- ✅ Error handling

---

## Phase 7: Gist Integration

**Duration:** 5-6 days

### 7.1 GitHub Client
**File:** `internal/gist/client.go`

```go
type GistClient struct {
    client *github.Client
    token  string
}

func (g *GistClient) CreateGist(files map[string]string) (*github.Gist, error)
func (g *GistClient) UpdateGist(gistID string, files map[string]string) error
func (g *GistClient) FetchGist(gistID string) (map[string]string, error)
func (g *GistClient) DeleteGist(gistID string) error
func (g *GistClient) ListGists() ([]*github.Gist, error)
```

### 7.2 Token Management
**File:** `internal/gist/token.go`

Secure token storage:
- Save to keyring (macOS/Linux/Windows)
- Fallback to encrypted file
- Token validation
- Masked display

### 7.3 Gist Sync
**File:** `internal/gist/sync.go`

Features:
- Upload all favorites to gist
- Download from gist
- Conflict resolution
- Progress indication

### 7.4 Gist UI
**File:** `internal/ui/gist.go`

Menu for:
- Create gist
- Update gist
- Delete gist
- List gists
- Recover from gist
- Token management

**Tests:**
- Mock GitHub API
- Token encryption
- Sync logic
- Error scenarios

#### Deliverables:
- ✅ Gist CRUD operations
- ✅ Secure token storage
- ✅ Backup/restore functionality
- ✅ Token management UI

---

## Phase 8: Polish & UX Improvements

**Duration:** 4-5 days

### 8.1 Loading States
Add spinners and progress bars:
- Searching stations
- Creating gist
- Loading favorites

### 8.2 Error Handling
Improve error messages:
- Network errors
- API errors
- File errors
- Player errors

### 8.3 Help System
**File:** `internal/ui/help.go`

Context-sensitive help:
- Show key bindings
- Explain current screen
- Tips and shortcuts

### 8.4 Configuration
**File:** `internal/storage/config.go`

Persistent settings:
- Last played station
- Default search type
- UI preferences
- MPV player path

### 8.5 Performance
- Optimize list rendering
- Cache search results
- Lazy load large lists
- Debounce search input

#### Deliverables:
- ✅ Smooth user experience
- ✅ Clear error messages
- ✅ Help documentation
- ✅ Performance optimizations

---

## Phase 9: Testing & Quality Assurance

**Duration:** 5-6 days

### 9.1 Unit Tests
- All packages >80% coverage
- Edge cases
- Error paths
- Concurrent access

### 9.2 Integration Tests
Test full workflows:
- Search → Play → Save
- Create list → Add stations → Play
- Backup to gist → Recover
- Delete station → Verify

### 9.3 Manual Testing
Test on different platforms:
- macOS (Intel & Apple Silicon)
- Linux (Ubuntu, Arch)
- Windows (if supported)

### 9.4 Regression Testing
- Compare with bash version
- Verify backward compatibility
- Test migration path

### 9.5 Performance Testing
- Startup time
- Search speed
- Memory usage
- Large lists (1000+ stations)

#### Deliverables:
- ✅ Comprehensive test suite
- ✅ Platform compatibility verified
- ✅ Performance benchmarks
- ✅ Bug fixes

---

## Phase 10: Documentation & Release

**Duration:** 3-4 days

### 10.1 User Documentation
- README with installation
- User guide
- Keyboard shortcuts
- Configuration options
- Troubleshooting

### 10.2 Developer Documentation
- Architecture overview
- API documentation
- Contributing guide
- Code style guide

### 10.3 Release Process
- Version tagging
- Changelog
- Build artifacts (binaries)
- GitHub release
- Homebrew formula

### 10.4 Migration Guide
Help users transition from bash:
- Installation steps
- Data migration
- Feature comparison
- Breaking changes (if any)

#### Deliverables:
- ✅ Complete documentation
- ✅ Release artifacts
- ✅ Homebrew formula
- ✅ Migration guide

---

## Timeline Summary

| Phase | Duration | Description |
|-------|----------|-------------|
| 0 | 1-2 days | Project setup |
| 1 | 3-4 days | Data models & storage |
| 2 | 3-4 days | API client |
| 3 | 2-3 days | MPV player |
| 4 | 5-6 days | Basic TUI |
| 5 | 6-7 days | Play & search |
| 6 | 4-5 days | List management |
| 7 | 5-6 days | Gist integration |
| 8 | 4-5 days | Polish & UX |
| 9 | 5-6 days | Testing & QA |
| 10 | 3-4 days | Documentation & release |
| **Total** | **41-52 days** | **~2 months** |

---

## Development Workflow

### Daily Workflow
1. **Start of day:**
   - Pull latest changes
   - Review task list
   - Run tests

2. **Development:**
   - Work on one phase at a time
   - Write tests first (TDD when possible)
   - Commit frequently with clear messages

3. **End of day:**
   - Run full test suite
   - Update documentation
   - Push changes
   - Plan next day's tasks

### Code Review Checklist
- [ ] Tests pass
- [ ] Code formatted with `gofmt`
- [ ] No linter warnings
- [ ] Documentation updated
- [ ] Error handling present
- [ ] Performance considered

---

## Milestones

### Milestone 1: Core Foundation (End of Phase 3)
- ✅ Data models
- ✅ Storage layer
- ✅ API client
- ✅ Player integration
- **Deliverable:** CLI tool that can search and play stations

### Milestone 2: Basic TUI (End of Phase 4)
- ✅ Main menu
- ✅ Navigation
- ✅ Styling
- **Deliverable:** Interactive TUI with basic navigation

### Milestone 3: Feature Complete (End of Phase 7)
- ✅ All screens implemented
- ✅ Full functionality
- ✅ Gist integration
- **Deliverable:** Feature parity with bash version

### Milestone 4: Production Ready (End of Phase 10)
- ✅ All tests passing
- ✅ Documentation complete
- ✅ Performance optimized
- ✅ Release artifacts ready
- **Deliverable:** v1.0.0 release

---

## Risk Mitigation

### Technical Risks

**Risk 1: Bubble Tea Learning Curve**
- **Mitigation:** Build simple prototypes first
- **Fallback:** Use simpler terminal library (termui, tview)

**Risk 2: MPV Integration Issues**
- **Mitigation:** Test early on all platforms
- **Fallback:** Support alternative players (mplayer, vlc)

**Risk 3: GitHub Keyring Access**
- **Mitigation:** Implement file-based fallback
- **Fallback:** Encrypted local storage

**Risk 4: Performance with Large Lists**
- **Mitigation:** Implement pagination early
- **Fallback:** Limit list sizes, use filtering

### Schedule Risks

**Risk 1: Underestimated Complexity**
- **Mitigation:** Build MVPs for each phase
- **Buffer:** 25% time buffer included in estimates

**Risk 2: Platform-Specific Issues**
- **Mitigation:** Test on target platforms early
- **Fallback:** Focus on Linux/macOS first

---

## Success Criteria

### Functional Requirements
- [ ] All bash features implemented
- [ ] Backward compatible with existing data
- [ ] Works on Linux and macOS
- [ ] Single binary distribution
- [ ] No runtime dependencies (except mpv)

### Performance Requirements
- [ ] Startup < 100ms
- [ ] Search results < 2s
- [ ] Smooth 60fps UI
- [ ] Memory usage < 50MB

### Quality Requirements
- [ ] Test coverage > 80%
- [ ] No critical bugs
- [ ] All documentation complete
- [ ] Clean, maintainable code

### User Experience
- [ ] Intuitive navigation
- [ ] Clear error messages
- [ ] Responsive keyboard controls
- [ ] Consistent styling

---

## Post-Release

### v1.1 Features (Future)
- Playlist support
- Recording streams
- Station recommendations
- Dark/light themes
- Custom key bindings
- Search history
- Recently played

### v2.0 Features (Future)
- Web interface
- Mobile app (via SSH)
- Plugin system
- Collaborative playlists
- Social features
