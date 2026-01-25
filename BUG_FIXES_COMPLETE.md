# Bug Fixes Summary

## Issues Fixed

### 1. ✅ Station Continues Playing After Quit
**Problem:** Pressing 'q' to quit tera doesn't stop the MPV player process.

**Root Cause:** The `player.Stop()` method wasn't called when quitting the application.

**Fix:** Added proper cleanup in `app.go`:
- Stop player on `Ctrl+C` (global quit)
- Stop player when pressing 'q' or '0' from main menu
- Stop both `playScreen.player` and `searchScreen.player` to handle all cases

**Files Changed:**
- `internal/ui/app.go`

---

### 2. ✅ Search Menu Screen Height Too Short
**Problem:** Search menu only shows 1-2 options; need to scroll to see all 6 options.

**Root Cause:** List height was hardcoded to small values (12 lines) instead of using dynamic terminal height.

**Fix:** 
- Changed initial menu height from 12 to 20 lines
- Added dynamic height calculation based on terminal size
- Calculate usable height as `terminalHeight - 8` (leaving room for title, help, padding)
- Set minimum height of 5 lines for very small terminals
- Update list sizes on window resize events

**Formula:**
```go
listHeight := msg.Height - 8
if listHeight < 5 {
    listHeight = 5  // Minimum
}
```

**Files Changed:**
- `internal/ui/search.go`

---

### 3. ✅ Missing Save Prompt After Search Play
**Problem:** After playing a station from search results and pressing 'q', no save prompt appears.

**Expected Behavior:** According to `flow-charts.md` section 4, search results should show a save prompt after playback stops.

**Fix:**
- Added new state `searchStateSavePrompt`
- Modified `handlePlaybackStopped()` to check if station is already in Quick Favorites
- If already saved: Show message and return to results
- If not saved: Show save prompt
- Added `handleSavePrompt()` to handle user input (y/1 for yes, n/2/Esc for no)
- Added `renderSavePrompt()` to display the save dialog

**New Flow:**
```text
Play Station → Press q → Check if in Quick Favorites
    ├─ Yes → Show "Already saved" message → Back to results
    └─ No  → Show save prompt → User choice
              ├─ Yes → Save to Quick Favorites → Back to results
              └─ No  → Back to results
```

**Files Changed:**
- `internal/ui/search.go`

---

### 4. ✅ Filter Count Not Updating
**Problem:** When filtering search results with '/', the result count doesn't update.

**Root Cause:** Status bar wasn't enabled on the results list.

**Fix:**
- Added `SetShowStatusBar(true)` when creating results list
- The bubbles list component automatically shows "x/y items" when filtering is enabled and status bar is visible

**Files Changed:**
- `internal/ui/search.go`

---

### 5. ✅ Play Screen Height Too Short
**Problem:** Similar to search menu - play screen lists don't use full terminal height.

**Fix:**
- Applied same dynamic height calculation as search screen
- Calculate usable height when lists are loaded
- Update on window resize
- Minimum height of 5 lines

**Files Changed:**
- `internal/ui/play.go`

---

## Testing Checklist

### Test 1: Stop Player on Quit
- [ ] Start tera
- [ ] Go to Play or Search
- [ ] Play a station
- [ ] Press 'q' to quit
- [ ] Verify: MPV process stops (no audio continues)

### Test 2: Search Menu Height
- [ ] Start tera
- [ ] Press '2' for Search
- [ ] Verify: All 6 search options visible without scrolling
  1. Search by Tag
  2. Search by Name  
  3. Search by Language
  4. Search by Country Code
  5. Search by State
  6. Advanced Search

### Test 3: Save Prompt After Search Play
- [ ] Search for stations
- [ ] Select and play a station
- [ ] Press 'q' to stop
- [ ] Verify: Save prompt appears (if not already in Quick Favorites)
- [ ] Press '1' or 'y' to save
- [ ] Verify: Station added to Quick Favorites
- [ ] Play same station again and press 'q'
- [ ] Verify: "Already in Quick Favorites" message (no prompt)

### Test 4: Filter Count Updates
- [ ] Search for stations (get multiple results)
- [ ] Press '/' to activate filter
- [ ] Type some characters
- [ ] Verify: Status bar shows "x/y items" and updates as you type

### Test 5: Play Screen Height
- [ ] Go to Play from Favorites
- [ ] Verify: List of favorites uses full height
- [ ] Select a list
- [ ] Verify: Station list uses full height with filter enabled

### Test 6: Window Resize
- [ ] On each screen (main, search, play, results)
- [ ] Resize terminal window
- [ ] Verify: Lists adapt to new size

---

## Code Changes Summary

### `internal/ui/app.go`
```go
// Added player cleanup on quit
case "ctrl+c":
    if a.screen == screenPlay && a.playScreen.player != nil {
        a.playScreen.player.Stop()
    } else if a.screen == screenSearch && a.searchScreen.player != nil {
        a.searchScreen.player.Stop()
    }
    return a, tea.Quit
```

### `internal/ui/search.go`
```go
// Added new state
const (
    // ...
    searchStateSavePrompt  // NEW
)

// Dynamic height calculation
listHeight := msg.Height - 8
if listHeight < 5 {
    listHeight = 5
}

// Save prompt after playback
func (m SearchModel) handlePlaybackStopped() (tea.Model, tea.Cmd) {
    if isDuplicate {
        // Show message
    } else {
        // Show save prompt
        m.state = searchStateSavePrompt
    }
}

// Status bar for filter count
m.resultsList.SetShowStatusBar(true)
```

### `internal/ui/play.go`
```go
// Same dynamic height calculation as search
listHeight := msg.Height - 10
if listHeight < 5 {
    listHeight = 5
}
```

---

## Verification Commands

Build and run:
```bash
make clean
make build
./tera
```

Check for MPV processes after quit:
```bash
ps aux | grep mpv
# Should return nothing (no running MPV processes)
```

---

## Notes

- All fixes are backward compatible
- No breaking changes to existing functionality
- Terminal resize is handled gracefully
- Minimum heights prevent issues on very small terminals
- Save prompt only appears when needed (new stations)
- Duplicate checking by StationUUID (as per spec)
