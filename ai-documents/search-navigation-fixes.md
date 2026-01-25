# Search Screen Navigation Fixes

## Issues Fixed

### 1. Search Results Not Displaying
**Problem:** After searching, the results list was not visible.

**Root Cause:** The results list was being created with initial width/height values that might be zero or too small.

**Solution:**
- Added default width (80) and height (24) values in `NewSearchModel()`
- Improved window size handling to update list sizes based on current state
- Added checks to ensure list has items before setting size

### 2. No Arrow Key Navigation in Station Info Menu
**Problem:** The station info submenu only supported number shortcuts (1-3).

**Root Cause:** The station info screen used simple string matching instead of list-based navigation.

**Solution:**
- Added `stationInfoMenu` field to `SearchModel` using the menu component
- Created menu items for the three actions (Play, Save, Back)
- Refactored `handleStationInfoInput()` to use `HandleMenuKey()`
- Added `executeStationAction()` helper function
- Updated `renderStationInfo()` to display the menu list

## Changes Made

### 1. SearchModel Structure
```go
type SearchModel struct {
    // ... existing fields ...
    stationInfoMenu  list.Model  // NEW: Station info submenu navigation
    width            int          // UPDATED: Default 80
    height           int          // UPDATED: Default 24
}
```

### 2. NewSearchModel()
**Added:**
- Station info menu initialization
- Default width/height values
- Menu items for station actions

### 3. Window Size Handling
**Improved:**
```go
case tea.WindowSizeMsg:
    m.width = msg.Width
    m.height = msg.Height
    
    // Update list sizes based on current state
    switch m.state {
    case searchStateResults:
        if m.resultsList.Items() != nil && len(m.resultsList.Items()) > 0 {
            m.resultsList.SetSize(msg.Width, msg.Height-10)
        }
    case searchStateMenu:
        m.menuList.SetSize(msg.Width, msg.Height-8)
    case searchStateStationInfo:
        m.stationInfoMenu.SetSize(msg.Width-4, 8)
    }
```

### 4. Station Info Input Handling
**Before:**
```go
switch msg.String() {
case "1": // Play
case "2": // Save
case "3": // Back
}
```

**After:**
```go
// Handle menu navigation and selection
newList, selected := components.HandleMenuKey(msg, m.stationInfoMenu)
m.stationInfoMenu = newList

if selected >= 0 {
    return m.executeStationAction(selected)
}
```

### 5. Station Info Rendering
**Before:**
```text
What would you like to do?

1) Play this station
2) Save to Quick Favorites
3) Back to search results

0) Main Menu  |  Esc) Back
```

**After:**
```text
What would you like to do?

> 1. Play this station
  2. Save to Quick Favorites
  3. Back to search results

↑↓/jk: Navigate • Enter: Select • 1-3: Quick select • 0: Main • Esc: Back
```

## Features Added

### Station Info Menu Navigation

**Arrow Keys:**
- `↑` / `k` - Move up
- `↓` / `j` - Move down
- `Enter` - Select action

**Shortcuts:**
- `1` - Play station (quick select)
- `2` - Save to favorites (quick select)
- `3` - Back to results (quick select)
- `0` - Return to main menu
- `Esc` - Back to results

**Visual Feedback:**
- Highlighted selected item (with `>` indicator)
- Consistent styling with other menus
- Clear keyboard shortcuts help text

## User Experience Improvements

### Before
1. **Search Results:** List might not display properly
2. **Station Info:** Number-only navigation
3. **No Visual Feedback:** Unclear which option was selected

### After
1. **Search Results:** Properly sized and visible list
2. **Station Info:** Multiple navigation methods (arrows, vim keys, numbers)
3. **Visual Feedback:** Clear highlighting of selected option
4. **Consistent UX:** Same navigation pattern as other menus

## Testing Checklist

- [x] Search results display correctly after search
- [x] Results list is properly sized
- [x] Arrow keys work in station info menu (↑↓)
- [x] Vim keys work in station info menu (j/k)
- [x] Number shortcuts work (1-3)
- [x] Enter key selects highlighted option
- [x] Esc returns to results
- [x] 0 returns to main menu
- [x] Visual highlighting works
- [x] Window resize updates list sizes

## Files Modified

1. `internal/ui/search.go`
   - Added `stationInfoMenu` field
   - Updated `NewSearchModel()` with defaults and menu
   - Improved window size handling
   - Refactored `handleStationInfoInput()`
   - Added `executeStationAction()`
   - Updated `renderStationInfo()`

## Consistency with Other Screens

The station info menu now follows the same patterns as:
- Main menu (arrow keys + number shortcuts)
- Search menu (arrow keys + number shortcuts)
- Play screen list selection

This creates a consistent user experience throughout the application.

## Next Steps

Consider adding similar arrow key navigation to:
1. List management screens
2. Gist management screens
3. Any other menu-based interfaces

All menus should use the reusable `components.MenuItem` and `components.HandleMenuKey()` for consistency.
