# Test Fixes & Navigation Updates

## Issues Fixed

### 1. List Management Menu Display
**Problem:** Menu not showing options (1-4) properly, insufficient height
**Fix:** Updated `WindowSizeMsg` handler to use minimum height of 12 lines for menu display

### 2. Play from Favorites Empty Display  
**Problem:** Not showing the list of favorites
**Fix:** List model initialization issue - now properly initializes with window size

### 3. Keyboard Navigation - Removed "0" Key
**Problem:** Using both "0" and "Esc" for back navigation (against spec)
**Fix:** Removed all "0" key handlers except in main menu (Exit option)
- Only "Esc" goes back
- Only "q" quits
- Main menu "0" still works for Exit

## Files Modified

1. `internal/ui/list.go` - Removed "0" from `handleMenuInput`, updated help text
2. `internal/ui/play.go` - Removed "0" from navigation handlers, updated help text  
3. `internal/ui/search.go` - Removed "0" from navigation (keeping "Esc" only)
4. `internal/ui/app.go` - Fixed window resize handling for list management menu
5. Test files updated to match new behavior

## Updated Keyboard Shortcuts

Per `keyboard-shortcuts-guide.md`:
- **Esc** - Go back/Cancel (primary back navigation)
- **q** - Exit TERA (quit application)
- **0** - Only in main menu for Exit option
- **Ctrl+C** - Force quit

All navigation now consistent across screens.
