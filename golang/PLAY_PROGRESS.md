# Play Screen Development Progress

## ✅ Completed: Step 1 - List Selection

**Date**: January 23, 2026

### What Works
- Navigate to Play screen from main menu (press `1`)
- Load all favorite lists from `~/.config/tera/favorites/`
- Display lists with arrow key navigation
- Select a list with Enter
- Return to main menu with Esc or `0`
- Graceful error handling when no lists exist

### Files
- `internal/ui/play.go` - Core implementation
- `internal/ui/play_test.go` - Test suite
- `internal/ui/styles.go` - Shared styles
- `internal/ui/app.go` - Integration with main app

### How to Test

```bash
# Run unit tests
go test ./internal/ui/play_test.go -v

# Build and run
go build -o tera cmd/tera/main.go
./tera
# Press 1 to enter Play screen

# Create test data
mkdir -p ~/.config/tera/favorites
echo '[]' > ~/.config/tera/favorites/My-favorites.json
echo '[]' > ~/.config/tera/favorites/Jazz.json
```

## 🚧 Next: Step 2 - Station Selection

### Goals
1. Load stations from selected list JSON file
2. Display stations with fzf-style interface
3. Add filtering capability (`/` key)
4. Sort stations alphabetically (case-insensitive)
5. Handle empty lists gracefully

### Implementation Plan

#### 1. Load Stations
```go
func (m PlayModel) loadStations(listName string) tea.Cmd {
    // Read JSON file
    // Parse into []Station
    // Sort alphabetically
    // Return stationsLoadedMsg
}
```

#### 2. Station Display
- Use `bubbles/list` with custom delegate
- Enable filtering
- Show station name and basic info

#### 3. Station Item
```go
type stationListItem struct {
    station api.Station
}
// Implement list.Item interface
```

#### 4. Update Handler
```go
func (m PlayModel) updateStationSelection(msg tea.KeyMsg) {
    switch msg.String() {
    case "esc", "0": // Back to list selection
    case "/": // Start filtering
    case "enter": // Select station
    }
}
```

### Files to Create/Modify
- `internal/ui/play.go` - Add station selection logic
- `internal/ui/play_test.go` - Add station tests
- `internal/storage/favorites.go` - Use existing LoadList function

### Testing Strategy
1. Create test JSON with sample stations
2. Test loading and parsing
3. Test alphabetical sorting
4. Test filtering
5. Test navigation (back to lists, select station)

### Acceptance Criteria
- [ ] Can select a list and see its stations
- [ ] Stations are sorted alphabetically
- [ ] Can filter stations by typing `/` + text
- [ ] Can go back to list selection with Esc/0
- [ ] Shows helpful message for empty lists
- [ ] Can select a station with Enter
- [ ] All tests pass

## 🔮 Future: Step 3 - Playback

After station selection, implement:
1. Integrate MPV player
2. Show station info before playing
3. Handle playback start/stop
4. Add save to Quick Favorites (press `s`)
5. Check for duplicates before saving

## Development Notes

### Current Architecture
```
App (ui/app.go)
  └─ PlayModel (ui/play.go)
      ├─ playStateListSelection (DONE)
      ├─ playStateStationSelection (TODO)
      └─ playStatePlaying (TODO)
```

### Message Flow
```
navigateMsg{screenPlay} 
  → App.Update 
  → Initialize PlayModel
  → PlayModel.Init()
  → loadLists()
  → listsLoadedMsg
  → Update list display
```

### Data Flow
```
JSON files → getAvailableLists() → []string
  → playListItem → list.Model → View
```

## Questions?

Refer to:
- `golang/spec-docs/flow-charts.md` - Full specification
- `golang/PLAY_SCREEN_STEP1.md` - Step 1 details
- `golang/GETTING_STARTED.md` - Getting started guide
