# TERA Go API Specification

This document defines the public interfaces and external integrations for the TERA terminal radio application, written in Go. It reflects the current Radio Browser API behavior and aligns with idiomatic Go design.

---

## Radio Browser API Client
Each external service defines its own base URL and timeout constants to avoid ambiguity and allow independent tuning.

### Base Configuration

```go
const (
    RadioBrowserBaseURL = "https://de1.api.radio-browser.info"
    RadioBrowserTimeout = 10 * time.Second

    // Timeouts
    RadioBrowserTimeout = 10 * time.Second
    PlayerTimeout       = 5 * time.Second
)
```

All Radio Browser endpoints are built relative to `RadioBrowserBaseURL`.

---

### Client Interface

```go
type RadioBrowserClient interface {
    // Search
    SearchByTag(ctx context.Context, tag string) ([]Station, error)
    SearchByName(ctx context.Context, name string) ([]Station, error)
    SearchByLanguage(ctx context.Context, language string) ([]Station, error)
    SearchByCountry(ctx context.Context, countryCode string) ([]Station, error)
    SearchByState(ctx context.Context, state string) ([]Station, error)
    SearchAdvanced(ctx context.Context, params SearchParams) ([]Station, error)

    // Lookup
    GetByUUID(ctx context.Context, uuid string) (*Station, error)
    GetByURL(ctx context.Context, streamURL string) (*Station, error)

    // Listing
    ListAll(ctx context.Context) ([]Station, error)
}
```

---

### Search Endpoints

All search operations use **HTTP GET** and pass parameters via the query string.

**Base Path:** `/json/stations/search`

#### SearchByTag

```go
func (c *Client) SearchByTag(ctx context.Context, tag string) ([]Station, error)
```

**Request:**

```
GET /json/stations/search?tag=<value>
```

---

#### SearchByName

```go
func (c *Client) SearchByName(ctx context.Context, name string) ([]Station, error)
```

**Request:**

```
GET /json/stations/search?name=<value>
```

---

#### SearchByLanguage

```go
func (c *Client) SearchByLanguage(ctx context.Context, language string) ([]Station, error)
```

**Request:**

```
GET /json/stations/search?language=<value>
```

---

#### SearchByCountry

```go
func (c *Client) SearchByCountry(ctx context.Context, countryCode string) ([]Station, error)
```

**Request:**

```
GET /json/stations/search?countrycode=<value>
```

---

#### SearchByState

```go
func (c *Client) SearchByState(ctx context.Context, state string) ([]Station, error)
```

**Request:**

```
GET /json/stations/search?state=<value>
```

---

#### SearchAdvanced

```go
func (c *Client) SearchAdvanced(ctx context.Context, params SearchParams) ([]Station, error)
```

**Request:**

```
GET /json/stations/search?<query parameters>
```

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
    Limit       int
    Offset      int
}
```

Only non-zero / non-empty fields are included in the query string.

---

### Lookup Endpoints

#### GetByUUID

```go
func (c *Client) GetByUUID(ctx context.Context, uuid string) (*Station, error)
```

**Request:**

```
GET /json/stations/byuuid/<uuid>
```

---

#### GetByURL

```go
func (c *Client) GetByURL(ctx context.Context, streamURL string) (*Station, error)
```

**Request:**

```
GET /json/stations/byurl/<url>
```

---

### List All Stations

```go
func (c *Client) ListAll(ctx context.Context) ([]Station, error)
```

**Request:**

```
GET /json/stations
```

---

## Station Model

```go
type Station struct {
    StationUUID string `json:"stationuuid"`
    Name        string `json:"name"`
    URL         string `json:"url"`
    URLResolved string `json:"url_resolved"`
    Tags        string `json:"tags"`
    Country     string `json:"country"`
    CountryCode string `json:"countrycode"`
    Language    string `json:"language"`
    Codec       string `json:"codec"`
    Bitrate     int    `json:"bitrate"`
    Votes       int    `json:"votes"`
}
```

---

## Storage Interface

Persistent storage is file-based and uses JSON encoding.

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

---

### File Layout

**Base Path:** `~/.config/tera/favorites/`

* One JSON file per favorites list
* UTF-8 encoding

#### Favorites List File Structure

```json
[
  {
    "stationuuid": "...",
    "name": "Station Name",
    "url_resolved": "https://...",
    "tags": "jazz,smooth",
    "country": "United States",
    "codec": "MP3",
    "bitrate": 128,
    "votes": 1234
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

---

### MPV Integration

#### Command

```bash
mpv --no-video --really-quiet <stream_url>
```

#### Behavior

* Start MPV as a subprocess
* Capture stdout/stderr for logging
* Handle SIGTERM for graceful shutdown
* Track process lifecycle and play state

---

## GitHub Gist Client

Used for backup and synchronization of favorites lists.

```go
type GistClient interface {
    Create(ctx context.Context, files map[string]string, description string) (*GistInfo, error)
    Update(ctx context.Context, gistID string, description string) error
    Delete(ctx context.Context, gistID string) error
    Get(ctx context.Context, gistID string) (map[string]string, error)
    List(ctx context.Context) ([]GistInfo, error)
}
```

---

### GitHub API Details

#### Create Gist

**POST** `https://api.github.com/gists`

Headers:

* `Authorization: Bearer <token>`
* `Accept: application/vnd.github+json`
* `X-GitHub-Api-Version: 2022-11-28`

```json
{
  "description": "Terminal radio favorite lists",
  "public": false,
  "files": {
    "favorites.json": {
      "content": "..."
    }
  }
}
```

---

#### Update Gist

**PATCH** `https://api.github.com/gists/<gist_id>`

```json
{
  "description": "New description"
}
```

---

#### Delete Gist

**DELETE** `https://api.github.com/gists/<gist_id>`

Returns `204 No Content` on success.

---

#### Get Gist

**GET** `https://api.github.com/gists/<gist_id>`

Returns a Gist object including file contents.

---

## Notes

* All network operations must honor context cancellation
* Clients should implement retry and backoff where appropriate
* API consumers must not rely on undocumented Radio Browser behavior


## Token Manager Interface
Manages GitHub authentication tokens used by the Gist client.
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
Token storage follows a secure, layered strategy.
**Priority:**
1. System keyring (macOS Keychain, Linux Secret Service, Windows Credential Manager)
2. Encrypted file fallback

**Keyring:**
- Service: `tera`
- Key: `github-token`

### File Fallback
- Path: `~/.config/tera/tokens/github_token`  
- Permissions: `0600` (owner read/write only)
Implementations must prefer the system keyring when available.

### Token Validation
Validate verifies the token against the GitHub API and returns token metadata.

- Returns an error if the token is expired, revoked, or invalid
- Must respect context cancellation
- Should extract scopes from response headers when available

#### Request
```sh
GET https://api.github.com/user
```
#### Headers
```text
Authorization: Bearer <token>
Accept: application/vnd.github+json
```

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
Logical grouping of favorite stations.
```go
type FavoritesList struct {
    Name     string    `json:"-"` // Derived from filename
    Stations []Station `json:"stations"`
}
```

**File Format:** `<name>.json` contains an array of `Station` objects.

### Config
```go
type Config struct {
    FavoritePath string   `json:"favorite_path"`
    CachePath    string   `json:"cache_path"`
    LastPlayed   *Station `json:"last_played,omitempty"`
    GistID       string   `json:"gist_id,omitempty"`
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
    GitHubAPIURL    = "https://api.github.com"
    
    // Timeouts
    PlayerTimeout  = 5 * time.Second
    
    // Limits
    MaxSearchResults = 1000
    MaxListSize      = 10000
    MaxTokenLength   = 200
    MinTokenLength   = 20
)
```
