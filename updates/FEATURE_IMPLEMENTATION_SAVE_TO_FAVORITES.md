# Feature Implementation: Save Last Played Station to Quick Play Favorites

**Date:** January 19, 2026  
**Feature:** Post-play prompt to add stations to Quick Play Favorites  
**Status:** ✅ Implemented

---

## Overview

Users can now easily save stations they just listened to into Quick Play Favorites directly after playing them, without navigating back through menus.

---

## User Experience

### Before This Feature
```text
1. User plays station from "Play from My List"
2. Station plays
3. User presses 'q' to quit
4. Back to Main Menu
5. User thinks "I liked that station!"
6. User must navigate: List → Find original list → Find station → Save
   (Tedious and easy to forget!)
```

### After This Feature
```text
1. User plays station from "Play from My List"
2. Station plays
3. User presses 'q' to quit
4. PROMPT APPEARS:
   
   Did you enjoy this station?
   
   Station: Jazz FM 91.1
   From list: jazz-stations
   
   1) ⭐ Add to Quick Play Favorites
   2) Return to Main Menu

5. User selects option 1
6. "✓ Added to Quick Play Favorites!"
7. Station now appears in Main Menu
```

---

## How It Works

### Trigger Points

The prompt appears after playing a station from:

1. ✅ **Play from My List** - Shows list name
2. ✅ **Search Results** - Shows "Search Results" as source
3. ✅ **I Feel Lucky** - Shows "Search Results" as source
4. ❌ **Quick Play Favorites** - No prompt (already in favorites)

### Smart Behavior

- **Duplicate Detection**: Won't add station if already in My-favorites.json
- **Non-intrusive**: Can quickly skip with ESC or option 2
- **Contextual**: Shows station name and source list
- **Consistent**: Uses fzf like rest of TERA

---

## Implementation Details

### Files Modified

1. **lib/lib.sh** (+76 lines)
   - Modified `_play()` to accept station data and list name
   - Added `_prompt_save_to_favorites()` function
   - Added `_add_to_quick_favorites()` function
   - Modified `_play_favorite_station()` to skip prompt for favorites

2. **lib/play.sh** (1 line)
   - Updated `fn_play()` to pass station data to `_play()`

3. **lib/search.sh** (simplified)
   - Updated `_search_play()` to pass station data to `_play()`
   - Removed old "save station" prompt (now handled by `_play()`)

4. **lib/lucky.sh** (no changes needed)
   - Already uses `_search_play()` which now passes data

### New Functions

#### `_prompt_save_to_favorites(station_data, list_name)`
**Purpose:** Display prompt asking user if they want to save station

**Parameters:**
- `station_data` - Full JSON object of the station
- `list_name` - Name of the source list (optional)

**Behavior:**
- Clears screen and shows station info
- Displays fzf menu with two options
- Calls `_add_to_quick_favorites()` if user selects option 1
- Returns to menu if user cancels or selects option 2

#### `_add_to_quick_favorites(station_data)`
**Purpose:** Add station to My-favorites.json with duplicate checking

**Parameters:**
- `station_data` - Full JSON object of the station

**Features:**
- Creates My-favorites.json if it doesn't exist
- Checks for duplicates using `stationuuid`
- Shows appropriate message if already exists
- Atomic file operation (tmp file + mv)
- 2-second confirmation message

### Function Signatures Changed

#### `_play(url, station_data, list_name)`
**Before:**
```bash
_play "$URL"
```

**After:**
```bash
_play "$URL" "$STATION_DATA" "$LIST_NAME"
```

**Parameters:**
- `url` - Stream URL (required)
- `station_data` - Full JSON object (optional - if empty, no prompt)
- `list_name` - Source list name (optional - for display only)

---

## Code Examples

### Playing from a List
```bash
# In lib/play.sh
_play "$URL_RESOLVED" "$STATION_DATA" "$LIST"
```

### Playing from Search
```bash
# In lib/search.sh
_play "$URL_RESOLVED" "$STATION_DATA" "Search Results"
```

### Playing from Quick Play (no prompt)
```bash
# In lib/lib.sh - _play_favorite_station()
_play "$URL_RESOLVED" "" ""  # Empty strings = no prompt
```

---

## User Guide

### How to Use

1. **Play any station** from:
   - Play from My List
   - Search results
   - I Feel Lucky

2. **Listen to the station**
   - Press `q` to quit when done

3. **Save prompt appears** (if not already in favorites)
   ```text
   Did you enjoy this station?
   
   Station: [Station Name]
   From list: [List Name]
   
   1) ⭐ Add to Quick Play Favorites
   2) Return to Main Menu
   ```

4. **Choose your action:**
   - Press `1` or `↓` then `Enter` - Add to favorites
   - Press `2` or `Enter` - Skip and return to menu
   - Press `ESC` - Skip and return to menu

5. **Confirmation:**
   - If added: "✓ Added to Quick Play Favorites!"
   - If duplicate: "⭐ Station is already in Quick Play Favorites!"

### Quick Play Favorites Menu

Stations added to Quick Play Favorites appear in the Main Menu:

```text
TERA MAIN MENU

1) Play from my list
2) Search radio stations
...

--- Quick Play Favorites ---
10) ▶ Jazz FM 91.1
11) ▶ Classical WQXR
12) ▶ Rock Radio Paradise
```

You can now play them with a single number selection!

---

## Edge Cases Handled

### 1. My-favorites.json Doesn't Exist
- ✅ Automatically created with `[]`
- ✅ Station added successfully

### 2. Duplicate Station
- ✅ Detected by `stationuuid`
- ✅ Friendly message: "Already in Quick Play Favorites!"
- ✅ No error, just skip

### 3. User Cancels Prompt
- ✅ ESC key works
- ✅ Returns to menu cleanly
- ✅ No error messages

### 4. Playing from Quick Play Favorites
- ✅ No prompt shown (already a favorite)
- ✅ Smooth experience

### 5. Invalid Station Data
- ✅ Prompt simply doesn't appear
- ✅ No crash or error

---

## Testing Checklist

### Manual Testing

- [x] Play from "Play from My List" → Prompt appears
- [x] Select "Add to favorites" → Station added
- [x] Play same station again → "Already in favorites" message
- [x] Press ESC on prompt → Returns to menu
- [x] Select "Return to Main Menu" → Returns to menu
- [x] Play from Search → Prompt appears with "Search Results"
- [x] Play from "I Feel Lucky" → Prompt appears
- [x] Play from Quick Play Favorites → No prompt (correct)
- [x] Added station appears in Main Menu
- [x] Quick Play station works from Main Menu
- [x] My-favorites.json created if doesn't exist
- [x] Duplicate detection works correctly

### Edge Case Testing

- [x] Delete My-favorites.json, then add station → File created
- [x] Add multiple different stations → All added
- [x] Try to add same station 3 times → Duplicate message each time
- [x] Play station, kill mpv with Ctrl+C → Prompt still works
- [x] Empty station data → No prompt, no crash

---

## Performance Impact

**Minimal:**
- Only 1 extra step after playing (optional)
- Duplicate check is fast (jq query on small JSON)
- No impact on play performance
- Atomic file operations

---

## Future Enhancements

Possible improvements:

1. **Add to Different List**
   - Option 3) "Add to different list"
   - Shows list picker

2. **Rating System**
   - Rate station 1-5 stars
   - Store in metadata

3. **Play Count Tracking**
   - Track how many times played
   - Show popular stations

4. **Quick Actions**
   - Share station URL
   - Copy station info
   - Report broken stream

---

## Backward Compatibility

✅ **Fully backward compatible**

- Old callers of `_play()` with 1 argument still work
- Optional parameters gracefully ignored if not provided
- No changes to data formats
- No migration needed

---

## Summary

### What Changed
- Added post-play prompt for saving stations
- 2 new functions: `_prompt_save_to_favorites()`, `_add_to_quick_favorites()`
- Updated `_play()` signature to accept station data
- Updated all play contexts to pass data

### User Benefits
- ✅ Save stations immediately after listening
- ✅ No menu navigation needed
- ✅ Duplicate-safe
- ✅ Quick and intuitive

### Developer Benefits
- ✅ Reusable functions
- ✅ Clean separation of concerns
- ✅ Easy to extend
- ✅ Well-documented

---

**Status: Complete and Production Ready! 🎉**
