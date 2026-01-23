# TERA Go API Specification

## Radio Browser API Client

### Base Configuration

```go
const (
    BaseURL = "https://de1.api.radio-browser.info/json/stations"
    Timeout = 10 * time.Second
)
```

### Client Interface

```go
type RadioBrowserClient interface {
    SearchByTag(ctx context.Context, tag string) ([]Station, error)
    SearchByName(ctx context.Context, name string) ([]Station, error)
    SearchByLanguage(ctx context.Context, lang string) ([]Station, error)
    SearchByCountry(ctx context.Context, code string) ([]Station, error)
    SearchByState(ctx context.Context, state string) ([]Station, error)
    SearchAdvanced(ctx context.Context, params SearchParams) ([]Station, error)
}
```

### Search Methods

#### SearchByTag
```go
func (c *Client) SearchByTag(ctx context.Context, tag string) ([]Station, error)
```
**POST** `/search`  
**Body:** `tag=<value>`  
**Returns:** Array of Station objects

#### SearchByName
```go
func (c *Client) SearchByName(ctx context.Context, name string) ([]Station, error)
```
**POST** `/search`  
**Body:** `name=<value>`  
**Returns:** Array of Station objects

#### SearchByLanguage
```go
func (c *Client) SearchByLanguage(ctx context.Context, lang string) ([]Station, error)
```
**POST** `/search`  
**Body:** `language=<value>`  
**Returns:** Array of Station objects

#### SearchByCountry
```go
func (c *Client) SearchByCountry(ctx context.Context, code string) ([]Station, error)
```
**POST** `/search`  
**Body:** `countrycode=<value>`  
**Returns:** Array of Station objects

#### SearchByState
```go
func (c *Client) SearchByState(ctx context.Context, state string) ([]Station, error)
```
**POST** `/search`  
**Body:** `state=<value>`  
**Returns:** Array of Station objects

#### SearchAdvanced
```go
func (c *Client) SearchAdvanced(ctx context.Context, params SearchParams) ([]Station, error)
```
**POST** `/search`  
**Body:** Multiple form fields  
**Returns:** Array of Station objects

**SearchParams:**
```go
type SearchParams struct {
    Tag         string
    Name        string
    Language    string
    CountryCode string
    State       string
    Codec       string
    BitrateMin  int
    BitrateMax  int
}
```

---

## Storage Interface

```go
type Storage interface {
    // Lists
    GetAllLists(ctx context.Context) ([]string, error)
    LoadList(ctx context.Context, name string) (*FavoritesList, error)
    SaveList(ctx context.Context, list *FavoritesList) error
    DeleteList(ctx context.Context, name string) error
    RenameList(ctx context.Context, oldName, newName string) error
    
    // Stations
    AddStation(ctx context.Context, listName string, station Station) error
    RemoveStation(ctx context.Context, listName string, stationUUID string) error
    StationExists(ctx context.Context, listName string, stationUUID string) (bool, error)
    
    // Config
    LoadConfig(ctx context.Context) (*Config, error)
    SaveConfig(ctx context.Context, config *Config) error
}
```

### File Operations

**Base Path:** `~/.config/tera/favorite/`  
**Format:** JSON  
**Encoding:** UTF-8

#### List File Structure
```json
[
  {
    "stationuuid": "...",
    "name": "Station Name",
    "url_resolved": "https://...",
    "tags": "jazz,smooth",
    "country": "United States",
    "votes": 1234,
    "codec": "MP3",
    "bitrate": 128
  }
]
```

---

## Player Interface

```go
type Player interface {
    Play(ctx context.Context, url string, station *Station) error
    Stop(ctx context.Context) error
    IsPlaying() bool
    CurrentStation() *Station
    Wait() error
}
```

### MPV Integration

#### Command Format
```bash
mpv --no-video --really-quiet <stream_url>
```

#### Process Management
- Start MPV as subprocess
- Capture stdout/stderr for logging
- Handle SIGTERM for graceful shutdown
- Track process state

---

## Gist Client Interface

```go
type GistClient interface {
    Create(ctx context.Context, files map[string]string, description string) (*GistInfo, error)
    Update(ctx context.Context, gistID string, description string) error
    Delete(ctx context.Context, gistID string) error
    Get(ctx context.Context, gistID string) (map[string]string, error)
    List(ctx context.Context) ([]GistInfo, error)
}
```

### GitHub API Methods

#### Create Gist
```go
func (c *GistClient) Create(ctx context.Context, files map[string]string, description string) (*GistInfo, error)
```
**POST** `https://api.github.com/gists`  
**Headers:**
- `Authorization: Bearer <token>`
- `Accept: application/vnd.github+json`
- `X-GitHub-Api-Version: 2022-11-28`

**Body:**
```json
{
  "description": "Terminal radio favorite lists",
  "public": false,
  "files": {
    "filename.json": {
      "content": "..."
    }
  }
}
```

#### Update Gist
**PATCH** `https://api.github.com/gists/<gist_id>`  
**Body:**
```json
{
  "description": "New description"
}
```

#### Delete Gist
**DELETE** `https://api.github.com/gists/<gist_id>`  
**Returns:** 204 No Content on success

#### Get Gist
**GET** `https://api.github.com/gists/<gist_id>`  
**Returns:** Gist object with files

---

## Token Manager Interface

```go
type TokenManager interface {
    Save(token string) error
    Load() (string, error)
    Delete() error
    Exists() bool
    Validate(ctx context.Context, token string) (*TokenInfo, error)
}
```

### Token Storage

**Priority:**
1. System keyring (macOS Keychain, Linux Secret Service, Windows Credential Manager)
2. Encrypted file fallback

**Keyring:**
- Service: `tera`
- Key: `github-token`

**File Fallback:** `~/.config/tera/tokens/github_token`  
**Permissions:** 600 (owner read/write only)

### Token Validation
```go
func (tm *TokenManager) Validate(ctx context.Context, token string) (*TokenInfo, error)
```
**GET** `https://api.github.com/user`  
**Headers:** `Authorization: Bearer <token>`

**TokenInfo:**
```go
type TokenInfo struct {
    Username string
    Valid    bool
    Scopes   []string
}
```

---

## Data Models

### Station
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

### FavoritesList
```go
type FavoritesList struct {
    Name     string    `json:"-"`          // From filename
    Stations []Station `json:"stations"`   // Station array
}
```

**File Format:** `<name>.json` contains array of Station objects directly

### Config
```go
type Config struct {
    FavoritePath string    `json:"favorite_path"`
    CachePath    string    `json:"cache_path"`
    LastPlayed   *Station  `json:"last_played,omitempty"`
    GistID       string    `json:"gist_id,omitempty"`
}
```

**File:** `~/.config/tera/config.json`

### GistInfo
```go
type GistInfo struct {
    ID          string    `json:"id"`
    URL         string    `json:"url"`
    Description string    `json:"description"`
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
}
```

### GistMetadata
```go
type GistMetadata struct {
    Gists []GistInfo `json:"gists"`
}
```

**File:** `~/.config/tera/gist_metadata.json`

---

## Error Types

```go
// API Errors
type APIError struct {
    StatusCode int
    Message    string
    Err        error
}

// Storage Errors
type StorageError struct {
    Operation string // "read", "write", "delete"
    Path      string
    Err       error
}

// Player Errors
type PlayerError struct {
    Reason string // "not_found", "failed_start", "stream_error"
    Err    error
}

// Validation Errors
type ValidationError struct {
    Field   string
    Message string
}
```

---

## Constants

```go
const (
    // Paths
    ConfigDir     = "~/.config/tera"
    FavoriteDir   = "~/.config/tera/favorite"
    CacheDir      = "~/.cache/tera"
    TokenDir      = "~/.config/tera/tokens"
    
    // Files
    ConfigFile       = "config.json"
    MetadataFile     = "gist_metadata.json"
    TokenFile        = "github_token"
    MyFavoritesFile  = "My-favorites.json"
    
    // API
    RadioBrowserURL = "https://de1.api.radio-browser.info/json/stations"
    GitHubAPIURL    = "https://api.github.com"
    
    // Timeouts
    HTTPTimeout    = 10 * time.Second
    PlayerTimeout  = 5 * time.Second
    
    // Limits
    MaxSearchResults = 1000
    MaxListSize      = 10000
    MaxTokenLength   = 200
    MinTokenLength   = 20
)
```
