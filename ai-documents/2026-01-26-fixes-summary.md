# TERA Fixes - January 26, 2026

## Issues Fixed

### 1. Test Failure (TestGetAvailableLists)
Fixed incorrect test expectations. My-favorites.json is created at app startup, not in getAvailableLists().

### 2. Search Results Empty Lines
Added `delegate.SetSpacing(0)` to remove spacing between station items.

### 3. Now Playing Navigation Keys
Changed key mappings:
- `1` - Stop & Save Prompt (was `q`)
- `Esc` - Back without save (was stop & back)
- `f` - Save to Quick Favorites (was `s`)
- `s` - Save to list (TODO)
- `q` - Quit app

### 4. Empty Line at Top of Now Playing
Added `\n` at start of viewPlaying() for better spacing.

### 5. Main Menu Navigation (m key)
Added `m` key to return directly to main menu from any depth > 1.
Implemented in search and list management screens.

### 6. Documentation Updates
Updated keyboard-shortcuts-guide.md with all new key mappings and navigation patterns.

## TODO Features

- Volume control with arrow keys (requires MPV IPC)
- Save to custom list feature (s key implementation)
- Investigate Manage Lists display issue (not reproduced)

## Modified Files

- internal/ui/play_test.go
- internal/ui/search.go
- internal/ui/play.go
- internal/ui/list.go
- golang/spec-docs/keyboard-shortcuts-guide.md
