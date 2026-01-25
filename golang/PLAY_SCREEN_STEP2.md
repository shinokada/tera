# Play Screen Implementation - Step 2: Station Selection

## What We Built

This is the second step in implementing the Play Screen. We've added the ability to browse and filter stations from a selected favorite list.

## Files Created/Modified

### Modified Files
1. **`internal/ui/play.go`** - Added station selection logic
   - `stationListItem` - List item for stations with description
   - `loadStations()` - Loads stations from JSON file
   - `getStationsFromList()` - Reads and sorts stations alphabetically
   - `updateStationSelection()` - Handles station selection input
   - `viewStationSelection()` - Renders station list view
   - `noStationsView()` - Handles empty lists gracefully

### New Files
2. **`internal/ui/play_station_test.go`** - Comprehensive tests
   - Tests for station list items
   - Tests for loading and sorting stations
   - Tests for navigation (back to lists)
   - Tests for empty lists
   - Tests for view rendering

## Features Implemented

✅ **Station Loading**
- Loads stations from selected list's JSON file
- Automatically sorts alphabetically (case-insensitive)
- Handles empty lists gracefully
- Error handling for missing/corrupt files

✅ **Station Display**
- fzf-style filtering with `/` key
- Shows station name, country, codec, and bitrate
- Pagination for large lists
- Status bar shows filter status

✅ **Navigation**
- Arrow keys to navigate stations
- `/` to start filtering (fzf-style)
- `Enter` to select station (placeholder for now)
- `Esc` or `0` to go back to list selection
- Proper state cleanup when navigating back

✅ **User Experience**
- Clear visual hierarchy
- Helpful messages for empty lists
- Smooth transitions between states
- Consistent keyboard shortcuts

## Testing

Run the tests:
```bash
cd /Users/shinichiokada/Terminal-Tools/tera
go test ./internal/ui -v -run "Play|Station"
```

Or use the script:
```bash
chmod +x run_play_step2_tests.sh
./run_play_step2_tests.sh
```

## How to Try It

1. **Build the app:**
```bash
go build -o tera cmd/tera/main.go
```

2. **Create test data with stations:**
```bash
mkdir -p ~/.config/tera/favorites
cat > ~/.config/tera/favorites/Jazz.json << 'EOF'
[
  {
    "stationuuid": "1",
    "name": "Smooth Jazz Florida",
    "url_resolved": "http://example.com/stream1",
    "country": "USA",
    "codec": "MP3",
    "bitrate": 128,
    "tags": "jazz, smooth",
    "votes": 100
  },
  {
    "stationuuid": "2",
    "name": "Jazz FM London",
    "url_resolved": "http://example.com/stream2",
    "country": "UK",
    "codec": "AAC",
    "bitrate": 192,
    "tags": "jazz, uk",
    "votes": 250
  },
  {
    "stationuuid": "3",
    "name": "Classic Jazz Radio",
    "url_resolved": "http://example.com/stream3",
    "country": "USA",
    "codec": "MP3",
    "bitrate": 160,
    "tags": "jazz, classic",
    "votes": 180
  }
]
EOF
```

3. **Run the app:**
```bash
./tera
```

4. **Try it out:**
   - Press `1` to enter Play screen
   - Select "Jazz" list with Enter
   - See stations sorted alphabetically:
     - Classic Jazz Radio
     - Jazz FM London
     - Smooth Jazz Florida
   - Press `/` and type "smooth" to filter
   - Press `Esc` to clear filter
   - Press `0` or `Esc` to go back to list selection

## Code Architecture

### State Machine
```text
playStateListSelection → playStateStationSelection → playStatePlaying
         ↑                        ↓                         ↓
         └────────────────────────┴─────────────────────────┘
```

### Data Flow
```text
Select List → loadStations() → getStationsFromList()
    ↓
Read JSON file → Parse stations → Sort alphabetically
    ↓
stationsLoadedMsg → Update stationListModel
    ↓
Display with filtering enabled
```

### Station List Item
```go
type stationListItem struct {
    station api.Station
}

// Title: Station name (trimmed)
// Description: "USA • MP3 128kbps"
// FilterValue: Full station name for searching
```

## Key Features

### Alphabetical Sorting
Stations are sorted case-insensitively by name:
```go
sort.Slice(stations, func(i, j int) bool {
    return strings.ToLower(stations[i].TrimName()) < 
           strings.ToLower(stations[j].TrimName())
})
```

### fzf-style Filtering
Built-in with bubbles/list:
```go
m.stationListModel.SetFilteringEnabled(true)
// User presses '/' to start filtering
// Fuzzy matching on station names
```

### Navigation Cleanup
When returning to list selection, we clean up:
```go
m.state = playStateListSelection
m.stations = nil
m.stationItems = nil
m.stationListModel = list.Model{}
```

## Next Steps

### Step 3: Playback (Coming Next)
- Integrate MPV player
- Show station info before playing
- Handle playback start/stop
- Add save to Quick Favorites (press `s`)
- Check for duplicates before saving
- Handle playback errors

## Test Coverage

✅ **Unit Tests**
- Station list item interface
- Station loading and parsing
- Alphabetical sorting
- Empty list handling
- File not found errors

✅ **Integration Tests**
- State transitions
- Navigation (Esc, 0 keys)
- Message handling
- View rendering

✅ **Edge Cases**
- Empty lists
- Missing files
- Invalid JSON
- Unicode in station names

## Questions?

Refer to:
- `golang/spec-docs/flow-charts.md` - Play Screen flow chart
- `golang/PLAY_PROGRESS.md` - Development roadmap
- `golang/GETTING_STARTED.md` - Getting started guide
