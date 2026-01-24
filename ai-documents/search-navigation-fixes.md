# Search Screen Display and Navigation Fixes

## Issues Fixed

### Issue 1: No Radio Stations Displayed in Search Results
**Problem**: After searching, the list showed "1000 items" but no station names were visible.

**Root Cause**: The `stationListItem` type was being used in `search.go` but was only defined in `play.go`, causing stations to not render properly.

**Fix**: The `stationListItem` type is already defined in `play.go` and is accessible to search.go since they're in the same package. The real issue was that the list wasn't being properly updated with navigation keys.

### Issue 2: Arrow Keys Not Working in Lists
**Problem**: Arrow keys (up/down) didn't work to navigate through search results or any other lists.

**Root Cause**: In `handleResultsInput()`, only "esc" and "enter" keys were explicitly handled, and no other keys were passed to the list component. This prevented arrow keys, page up/down, and other navigation keys from reaching the list model.

**Fix**: Added a `default` case to `handleResultsInput()` that passes all unhandled keys to the list component:

```go
default:
    // Pass all other keys (arrows, etc.) to the list for navigation
    var cmd tea.Cmd
    m.resultsList, cmd = m.resultsList.Update(msg)
    return m, cmd
```

Also removed the duplicate list update logic at the end of the main `Update()` function to avoid double-processing messages.

## Navigation Clarification

**Number-based menus**: The main menu and search type selection menu use number keys (1, 2, 3, etc.) by design. Arrow key navigation is not needed here.

**List-based selection**: The search results list and station lists in the Play screen now properly support:
- Arrow keys (↑/↓) for navigation  
- Page Up/Page Down for faster scrolling
- `/` for filtering/search
- `j`/`k` vim-style navigation
- `Enter` to select
- `Esc` to go back

## Files Modified
- `internal/ui/search.go` - Fixed `handleResultsInput()` to pass navigation keys to list component

## Testing
Test the following scenarios:
1. Search for stations (e.g., "jazz")
2. Use arrow keys to navigate the results
3. Press `/` to filter results
4. Press Enter to select a station
5. Verify the station details display correctly

All list navigation should now work as expected.
