# v3.1.0 - Station Metadata: Play Count & Last Played

## Feature Overview

Track user listening behavior by recording:
- **Play count**: How many times each station has been played
- **Last played timestamp**: When the station was last played
- **Most Played view**: A new sortable list showing listening statistics

## Goals

1. **Non-intrusive tracking** - Automatic, invisible to user during playback
2. **Privacy-first** - All data stored locally, never transmitted
3. **Useful insights** - Help users discover their favorites
4. **Consistent UX** - Follow existing Tera UI patterns

---

## Data Model

### Storage Structure

Extend the existing data directory structure:

```
~/.config/tera/data/
├── favorites/
│   ├── My-favorites.json
│   └── [other-lists].json
├── cache/
│   └── search-history.json
├── blocklist.json
├── voted_stations.json
└── station_metadata.json       # NEW: Play statistics
```

### Station Metadata Schema

**File**: `~/.config/tera/data/station_metadata.json`

```json
{
  "stations": {
    "station-uuid-1": {
      "play_count": 42,
      "last_played": "2026-02-10T14:30:00Z",
      "first_played": "2025-12-01T10:00:00Z",
      "total_duration_seconds": 12600
    },
    "station-uuid-2": {
      "play_count": 15,
      "last_played": "2026-02-09T20:15:00Z",
      "first_played": "2026-01-15T18:00:00Z",
      "total_duration_seconds": 4500
    }
  },
  "version": 1
}
```

### Go Struct

```go
// StationMetadata tracks listening statistics for a station
type StationMetadata struct {
    PlayCount            int       `json:"play_count"`
    LastPlayed           time.Time `json:"last_played"`
    FirstPlayed          time.Time `json:"first_played"`
    TotalDurationSeconds int64     `json:"total_duration_seconds"`
}

// MetadataStore holds all station metadata
type MetadataStore struct {
    Stations map[string]*StationMetadata `json:"stations"`
    Version  int                         `json:"version"`
    mu       sync.RWMutex                `json:"-"`
}
```

---

## Implementation Details

### 1. Core Metadata Manager

**Location**: `v3/internal/storage/station_metadata.go`

```go
package storage

type MetadataManager struct {
    dataPath string
    store    *MetadataStore
    mu       sync.RWMutex
}

// Core operations
func NewMetadataManager(dataPath string) (*MetadataManager, error)
func (m *MetadataManager) Load() error
func (m *MetadataManager) Save() error
func (m *MetadataManager) RecordPlay(stationUUID string) error
func (m *MetadataManager) GetMetadata(stationUUID string) *StationMetadata
func (m *MetadataManager) GetTopPlayed(limit int) []StationWithMetadata
func (m *MetadataManager) GetRecentlyPlayed(limit int) []StationWithMetadata
```

**Key behaviors**:
- Thread-safe: Uses mutex for concurrent access
- Auto-save: Writes to disk after each play recorded
- Graceful degradation: If file is corrupted, start fresh (don't crash)
- Efficient: Keep in-memory, flush periodically

### 2. Integration with Player

**Location**: `v3/internal/player/mpv.go`

```go
// When station starts playing
func (p *Player) Play(station api.Station) error {
    // ... existing play logic ...
    
    // Record play in background
    go p.recordPlay(station)
    
    return nil
}

func (p *Player) recordPlay(station api.Station) {
    if p.metadataManager == nil {
        return
    }
    
    if err := p.metadataManager.RecordPlay(station.StationUUID); err != nil {
        // Log error but don't interrupt playback
        log.Printf("Failed to record play: %v", err)
    }
}
```

**Important**: 
- Non-blocking: Use goroutine
- Silent failure: Never interrupt playback
- Best effort: If save fails, continue anyway

### 3. UI Integration

#### A. Station List Enhancement

**Location**: `v3/internal/ui/components/stationinfo.go`

Add metadata display to station info panel:

```
┌─ Station Info ──────────────────────────────────────┐
│ Name: Jazz FM                                       │
│ Country: United States                              │
│ Codec: MP3 | Bitrate: 128 kbps                     │
│                                                     │
│ 🎵 Played 42 times                                 │
│ 🕐 Last played: 2 hours ago                        │
└─────────────────────────────────────────────────────┘
```

**Display rules**:
- Show play count if > 0
- Show "Last played: [relative time]" if played before
  - "5 minutes ago"
  - "2 hours ago"
  - "3 days ago"
  - "2 weeks ago"
  - "on Feb 5, 2026" (if > 30 days)
- Use subtle color (dim/gray) to not overwhelm
- Only show if data exists (don't show "0 times")

#### B. New "Most Played" View

**Location**: `v3/internal/ui/most_played.go` (new file)

Add new menu item in main menu:

```
┌─ TERA Main Menu ─────────────────────────────────────┐
│                                                      │
│  1. Search Stations                                  │
│  2. Browse Favorites                                 │
│  3. Most Played          ← NEW                      │
│  4. Manage Lists                                     │
│  5. I'm Feeling Lucky                                │
│  6. Settings                                         │
│                                                      │
└──────────────────────────────────────────────────────┘
```

**Most Played View Design**:

```
┌─ Most Played Stations ───────────────────────────────┐
│                                                      │
│ Sort by: ▼ Play Count    [Enter to change]         │
│                                                      │
│  1. Jazz FM (United States)                   42 ▶  │
│     Last played: 2 hours ago                         │
│                                                      │
│  2. Classical Radio (Germany)                 35 ▶  │
│     Last played: yesterday                           │
│                                                      │
│  3. Rock 101 (Canada)                        28 ▶  │
│     Last played: 3 days ago                          │
│                                                      │
│  ... (showing 1-20 of 47)                           │
│                                                      │
└──────────────────────────────────────────────────────┘
Help: ↑↓/jk: Navigate • Enter: Play • s: Sort • Esc: Back
```

**Sort options**:
1. **Play Count** (default) - Most played first
2. **Last Played** - Most recent first
3. **First Played** - Oldest discoveries first
4. **Station Name** - Alphabetical

**Key bindings**:
- `s` - Cycle through sort options
- `Enter` - Play selected station
- `f` - Add to favorites
- `i` - Show station info
- Numbers `1-9` - Quick select
- `Esc/m` - Back to main menu

#### C. Recently Played Section

Add to main menu as alternative view:

```
┌─ Recently Played ────────────────────────────────────┐
│                                                      │
│  Today                                               │
│  • Jazz FM (2 hours ago)                     [Play] │
│  • Rock 101 (5 hours ago)                    [Play] │
│                                                      │
│  Yesterday                                           │
│  • Classical Radio (yesterday at 8:30 PM)    [Play] │
│  • Blues Station (yesterday at 2:15 PM)      [Play] │
│                                                      │
│  This Week                                           │
│  • Country 95 (3 days ago)                   [Play] │
│  • Electronic Vibes (5 days ago)             [Play] │
│                                                      │
└──────────────────────────────────────────────────────┘
```

---

## UI/UX Consistency Checklist

✅ **Follow Existing Patterns**:

1. **List Navigation**
   - ↑↓ or j/k for navigation
   - Enter to select/play
   - Numbers 1-9 for quick select
   - Esc or m to go back

2. **Color Scheme**
   - Use existing theme colors (follow `v3/internal/theme/theme.go`)
   - Subtle metadata (dim/gray)
   - Highlight on selection (same as other lists)
   - Success messages in green
   - Errors in red

3. **Layout**
   - Title at top with box drawing characters
   - Help bar at bottom
   - Consistent padding
   - Use `RenderPage()` and `RenderPageWithBottomHelp()`

4. **Messages**
   - Success: "✓ Station metadata updated"
   - Error: "✗ Failed to load metadata"
   - Info: "ℹ No play history yet - start listening!"

5. **Loading States**
   - Show spinner if loading takes time
   - Graceful fallback if metadata unavailable

---

## Migration & Backwards Compatibility

### v3.0.x → v3.1.0 Migration

**Auto-creation**:
- On first run of v3.1.0, create empty `station_metadata.json`
- No user action required
- Existing users start with clean slate

**Fallback behavior**:
- If metadata file doesn't exist: treat all play counts as 0
- If file is corrupted: log warning, start fresh
- Missing fields: use sensible defaults

---

## Privacy & Data Considerations

### What We Track
✅ Station UUID, play count, timestamps
✅ Aggregate duration per station

### What We DON'T Track
❌ No IP addresses
❌ No personal identifiers
❌ No external transmission
❌ No listening content/patterns shared

### User Control
- All data stored in `~/.config/tera/data/`
- User can delete `station_metadata.json` to clear history
- Optional: Add "Clear Statistics" menu option in Settings

---

## Testing Strategy

### Unit Tests

```go
// v3/internal/storage/station_metadata_test.go

func TestRecordPlay(t *testing.T)
func TestGetTopPlayed(t *testing.T)
func TestConcurrentAccess(t *testing.T)
func TestCorruptedFile(t *testing.T)
func TestMigrationFromEmpty(t *testing.T)
```

### Manual Testing

1. **New user experience**
   - Install v3.1.0 fresh
   - Verify metadata file created
   - Play station, verify count increments

2. **Existing user upgrade**
   - Start with v3.0.x config
   - Upgrade to v3.1.0
   - Verify no disruption

3. **Edge cases**
   - Delete metadata file while running
   - Corrupt JSON manually
   - Play same station rapidly

4. **UI testing**
   - Test all sort options
   - Verify relative time display
   - Check with 0, 1, 100, 1000+ stations

---

## Performance Considerations

### Optimization

1. **In-Memory Cache**
   - Keep metadata in RAM
   - Only write to disk on changes
   - Debounce saves (max once per 5 seconds)

2. **Lazy Loading**
   - Load metadata only when needed
   - Don't block startup

3. **File Size Management**
   - Typical size: ~100 bytes per station
   - 1000 stations = ~100KB (negligible)
   - No cleanup needed for years

---

## Future Enhancements (v3.2+)

Ideas for later releases:

- **Listening Streaks**: "5 days in a row"
- **Time of Day Analysis**: "You usually listen to jazz in the morning"
- **Listening Time Graphs**: Visual charts of listening patterns
- **Export Statistics**: Export to CSV/JSON
- **Favorites Auto-Sort**: Sort favorites by play count
- **Recommendations**: "Similar stations to your top played"

---

## Implementation Checklist

### Phase 1: Core Infrastructure
- [ ] Create `station_metadata.go` with core structs
- [ ] Implement `MetadataManager` with CRUD operations
- [ ] Add unit tests for metadata operations
- [ ] Integrate with player (`mpv.go`)

### Phase 2: Basic UI
- [ ] Add play count/last played to station info panel
- [ ] Create "Most Played" menu item
- [ ] Implement basic list view with sort by play count
- [ ] Add help text and keybindings

### Phase 3: Enhanced Features
- [ ] Add multiple sort options (last played, first played, name)
- [ ] Implement "Recently Played" grouped view
- [ ] Add relative time formatting
- [ ] Polish UI colors and alignment

### Phase 4: Testing & Documentation
- [ ] Write comprehensive tests
- [ ] Update README.md with new feature
- [ ] Create CHANGELOG entry
- [ ] Manual testing on different platforms

---

## Summary

This feature brings **passive listening insights** to Tera without changing the core experience. Users can continue using Tera exactly as before, while gaining visibility into their listening habits through the new "Most Played" view.

Key principles:
- ✅ **Non-intrusive**: Automatic tracking, no user action required
- ✅ **Privacy-focused**: All data local, never transmitted
- ✅ **Consistent**: Follows existing Tera UI/UX patterns
- ✅ **Optional**: Metadata display only in dedicated views
- ✅ **Safe**: Never interrupts playback
