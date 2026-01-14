# Quick Play Favorites Fix - January 13, 2026

## Issue Identified

The Quick Play Favorites feature on the main menu was reading from the wrong file location, causing newly saved stations to not appear.

### The Problem

**Before:**
- Quick Play Favorites read from: `${script_dir}/lib/favorite.json` (static file in repo)
- Users saving to "Favorite" list saved to: `~/.config/tera/favorite/sample.json` (user config)
- Result: Saved stations never appeared in Quick Play Favorites

**Root Cause:**
The system has two separate "favorite.json" files:
1. `lib/favorite.json` - A static sample file in the repository
2. `~/.config/tera/favorite/sample.json` - The actual user's Favorite list

When users saved stations to the "Favorite" list, they were saving to `sample.json` in their config directory, but the Quick Play Favorites was displaying stations from the static `lib/favorite.json` file.

## Solution

Updated both locations where the Quick Play Favorites reads stations to use the user's actual Favorite list.

### Files Modified

#### 1. `tera` (main menu function)
**Changed:**
```bash
# OLD - reading from static repository file
FAVORITE_STATIONS_FILE="${script_dir}/lib/favorite.json"

# NEW - reading from user's config directory
FAVORITE_STATIONS_FILE="${FAVORITE_PATH}/sample.json"
```

**Location:** Line ~93 in the `menu()` function

**Impact:** The main menu now displays stations from the user's actual Favorite list

#### 2. `lib/lib.sh` (_play_favorite_station function)
**Changed:**
```bash
# OLD - reading from static repository file
FAVORITE_STATIONS_FILE="${script_dir}/lib/favorite.json"

# NEW - reading from user's config directory
FAVORITE_STATIONS_FILE="${FAVORITE_PATH}/sample.json"
```

**Location:** Line ~188 in the `_play_favorite_station()` function

**Impact:** Clicking on Quick Play Favorites now plays stations from the user's list

## How It Works Now

### Complete Flow:

1. **User searches and finds a station** → Uses search menu
2. **User plays the station** → Tests if they like it
3. **User saves to "Favorite"** → Station saved to `~/.config/tera/favorite/sample.json`
4. **User returns to main menu** → Menu reads from `~/.config/tera/favorite/sample.json`
5. **Quick Play Favorites shows the new station** → ✅ Success!

### File Path Reference

| Variable | Value | Purpose |
|----------|-------|---------|
| `$FAVORITE_PATH` | `~/.config/tera/favorite` | User's favorite lists directory |
| `$FAVORITE_FILE` | `sample.json` | Default favorite list filename |
| `$FAVORITE_FULL` | `~/.config/tera/favorite/sample.json` | Full path to user's Favorite list |

### Display Name Mapping

| User Sees | Actual File | Location |
|-----------|-------------|----------|
| "Favorite" | `sample.json` | `~/.config/tera/favorite/sample.json` |
| Custom list names | `{name}.json` | `~/.config/tera/favorite/{name}.json` |

## Benefits

1. **Immediate Feedback**: Saved stations appear in Quick Play Favorites immediately
2. **Consistency**: Same file used for saving and displaying
3. **User Experience**: The feature now works as users expect
4. **Data Persistence**: User's favorites persist across sessions correctly

## Testing

To verify the fix works:

1. Start TERA: `tera`
2. Search for a station (option 2)
3. Play and save the station to "Favorite"
4. Return to main menu
5. Check "Quick Play Favorites" section
6. The newly saved station should appear in the list (if there are <10 stations total)

## Technical Notes

- The `lib/favorite.json` file is now only used as a sample/template
- On first run, it's copied to `~/.config/tera/favorite/sample.json`
- After that, `lib/favorite.json` is not used by the application
- All user interactions work with files in `~/.config/tera/favorite/`

## Backward Compatibility

This fix maintains backward compatibility:
- Existing users with stations in `~/.config/tera/favorite/sample.json` will see them immediately
- The change only affects where the code reads from, not the file structure
- No migration needed
