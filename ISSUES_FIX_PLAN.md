# Issues Fix Plan

## Issues Identified

### 1. Station Continues Playing After Quit
**Problem:** Pressing 'q' to quit tera doesn't stop the MPV player.
**Location:** `internal/ui/app.go` - main app Update handler
**Fix:** Ensure `player.Stop()` is called in all quit scenarios

### 2. Search Results Screen Height Too Short
**Problem:** Only shows 1-2 items, need to scroll to see more
**Location:** Multiple places where list height is set
**Fix:** Use dynamic height calculation based on terminal size
- `SearchModel.Init()`: Set initial height
- `SearchModel.Update()` for `tea.WindowSizeMsg`: Update on resize
- Use `height - 10` for content, leaving room for title/help

### 3. Missing Save Prompt After Search Play
**Problem:** After playing from search results and quitting, no save prompt appears
**Spec:** flow-charts.md section 4 shows save prompt should appear
**Location:** `internal/ui/search.go` - `handlePlaybackStopped()`
**Fix:** Implement save prompt state and logic after playback

### 4. Filter Count Not Updating
**Problem:** When filtering in search results, count doesn't update
**Location:** `internal/ui/search.go` - results list view
**Fix:** The bubbles list handles this automatically, may be a display issue
- Ensure `SetFilteringEnabled(true)` is set
- Check if status bar is shown

### 5. Search Menu Screen Height Too Short  
**Problem:** Only shows 1 option, same as issue #2
**Location:** `SearchModel.NewSearchModel()` and window size handling
**Fix:** Same as #2 - dynamic height calculation

## Implementation Order

1. **Fix #1 (Critical)** - Stop player on quit
2. **Fix #2 & #5 together** - Dynamic height for all lists
3. **Fix #3** - Add save prompt after search playback
4. **Fix #4** - Verify filter functionality

## Changes Needed

### File: `internal/ui/app.go`
```go
// In Update method, handle quit
case tea.KeyMsg:
    if msg.String() == "ctrl+c" || msg.String() == "q" {
        // Stop any playing stations
        if m.currentScreen == screenPlay {
            if playModel, ok := m.screens[screenPlay].(PlayModel); ok {
                playModel.player.Stop()
            }
        } else if m.currentScreen == screenSearch {
            if searchModel, ok := m.screens[screenSearch].(SearchModel); ok {
                searchModel.player.Stop()
            }
        }
        return m, tea.Quit
    }
```

### File: `internal/ui/search.go`

#### Dynamic Height
```go
// In NewSearchModel()
menuList := components.CreateMenu(menuItems, "🔍 Search Radio Stations", 50, 20)
// Will be updated in first WindowSizeMsg

// In Update for WindowSizeMsg
case tea.WindowSizeMsg:
    m.width = msg.Width
    m.height = msg.Height
    
    // Calculate usable height (leaving room for title, help, etc.)
    listHeight := msg.Height - 8  // Title + help + padding
    
    switch m.state {
    case searchStateMenu:
        m.menuList.SetSize(msg.Width-4, listHeight)
    case searchStateResults:
        m.resultsList.SetSize(msg.Width-4, listHeight)
    case searchStateStationInfo:
        m.stationInfoMenu.SetSize(msg.Width-4, 10)
    }
```

#### Save Prompt
```go
// Add new state
const (
    searchStateMenu searchState = iota
    searchStateInput
    searchStateLoading
    searchStateResults
    searchStateStationInfo
    searchStatePlaying
    searchStateSavePrompt  // NEW
)

// In handlePlaybackStopped()
func (m SearchModel) handlePlaybackStopped() (tea.Model, tea.Cmd) {
    // Check if station is already in Quick Favorites
    isDuplicate := false
    for _, s := range m.quickFavorites {
        if s.StationUUID == m.selectedStation.StationUUID {
            isDuplicate = true
            break
        }
    }
    
    if isDuplicate {
        m.saveMessage = "Already in Quick Favorites"
        m.saveMessageTime = 150
        m.state = searchStateResults
        return m, nil
    }
    
    // Show save prompt
    m.state = searchStateSavePrompt
    return m, nil
}

// Add handleSavePrompt method
func (m SearchModel) handleSavePrompt(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
    switch msg.String() {
    case "y", "1":
        // Save to Quick Favorites
        m.state = searchStateResults
        return m, m.saveToQuickFavorites(*m.selectedStation)
    case "n", "2", "esc":
        // Don't save, go back to results
        m.state = searchStateResults
        m.selectedStation = nil
        return m, nil
    }
    return m, nil
}

// In View()
case searchStateSavePrompt:
    return m.renderSavePrompt()

// Add renderSavePrompt method
func (m SearchModel) renderSavePrompt() string {
    var s strings.Builder
    
    s.WriteString(titleStyle.Render("💾 Save Station?"))
    s.WriteString("\n\n")
    
    if m.selectedStation != nil {
        s.WriteString(fmt.Sprintf("Do you want to add '%s' to Quick Favorites?\n\n", 
            m.selectedStation.TrimName()))
    }
    
    s.WriteString("1) ⭐ Yes - Add to Quick Favorites\n")
    s.WriteString("2) No - Return to search results\n\n")
    s.WriteString(subtleStyle.Render("y/1: Yes • n/2/Esc: No"))
    
    return s.String()
}
```

#### Filter Count Display
```go
// In NewSearchModel() when creating resultsList
m.resultsList.SetShowStatusBar(true)  // Ensure status bar is shown
m.resultsList.SetFilteringEnabled(true)

// The status bar automatically shows "x/y items" when filtering
```

### File: `internal/ui/play.go`
Same dynamic height fixes as search.go

## Testing Plan

1. Test quit with 'q' while playing - should stop station
2. Test search menu with small terminal - should show all options
3. Test search results with many items - should show appropriate number
4. Test filter in search results - count should update
5. Test play station from search then quit - should show save prompt
6. Test save prompt yes/no options
7. Test resize terminal while on different screens
