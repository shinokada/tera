# Station Name Improvements - January 17, 2026

## Overview
Implemented improvements to station name handling in TERA to address whitespace issues and improve usability through alphabetical sorting.

## Problems Identified

1. **Extra Whitespace**: Station names from the API sometimes contained leading/trailing whitespace (e.g., "  SmoothJazz.com 64k aac+" instead of "SmoothJazz.com 64k aac+")
2. **Random Order**: Stations were displayed in the order they were added (FIFO), making it difficult to find stations in longer lists

## Solutions Implemented

### 1. Whitespace Trimming

**When Saving Stations** (`lib/search.sh`):
- Modified `_save_station_to_list()` function to trim station names before saving to JSON files
- Uses jq's `gsub("^\\s+|\\s+$";"")` to remove leading and trailing whitespace

**When Displaying Stations** (`lib/lib.sh`):
- Modified `_station_list()` function to trim station names when reading from JSON
- Provides a safety net in case any untrimmed names exist in older data

**In Helper Scripts**:
- `list_favorites.sh`: Trims station names when displaying favorite stations
- `remove_favorite.sh`: Trims station names when listing stations for removal

### 2. Alphabetical Sorting

**Station Lists** (`lib/lib.sh`):
- Modified `_station_list()` function to sort stations alphabetically using `sort -f` (case-insensitive)
- Example: "BBC Radio", "Jazz FM", "SmoothJazz.com" instead of random order

**Helper Scripts**:
- `list_favorites.sh`: Sorts stations by name using jq's `sort_by(.value.name | ascii_downcase)`
- `remove_favorite.sh`: Sorts stations alphabetically in the list display

### 3. Station Selection Logic Updates

Since stations are now sorted alphabetically, the display index no longer matches the JSON array index. Updated all station selection functions to use **name-based lookup** instead of **index-based lookup**:

**Play Function** (`lib/play.sh`):
- Changed from: `jq -r ".[$ANS-1] |.url_resolved"`
- Changed to: Finding station by matching trimmed name using jq filter
- Benefits: Correctly finds and plays stations regardless of their position in the JSON array

**Delete Function** (`lib/delete_station.sh`):
- Changed from: `jq "del(.[${ANS}-1])"`
- Changed to: `jq --arg name "$STATION_TO_DELETE" 'del(.[] | select(.name | gsub("^\\s+|\\s+$";"") == $name))'`
- Benefits: Correctly identifies and deletes stations by name

## Files Modified

1. **lib/lib.sh**
   - Updated `_station_list()` to trim and sort station names

2. **lib/search.sh**
   - Updated `_save_station_to_list()` to trim station names before saving

3. **lib/play.sh**
   - Updated station selection logic to use name-based lookup
   - Inline display of station info (removed dependency on `_info_select_radio_play`)

4. **lib/delete_station.sh**
   - Updated station deletion to use name-based lookup

5. **list_favorites.sh**
   - Added sorting and trimming for display

6. **remove_favorite.sh**
   - Added sorting and trimming for display

## Technical Details

### jq Patterns Used

**Trimming whitespace:**
```bash
.name | gsub("^\\s+|\\s+$";"")
```

**Alphabetical sorting (case-insensitive):**
```bash
sort_by(.value.name | ascii_downcase)
```

**Name-based station lookup:**
```bash
jq --arg name "$STATION_NAME" '.[] | select(.name | gsub("^\\s+|\\s+$";"") == $name)'
```

**Name-based deletion:**
```bash
jq --arg name "$STATION_NAME" 'del(.[] | select(.name | gsub("^\\s+|\\s+$";"") == $name))'
```

## Benefits

1. **Cleaner Display**: No more extra spaces in station names
2. **Better Organization**: Alphabetical ordering makes finding stations much easier
3. **Improved UX**: Users can quickly locate their favorite stations in long lists
4. **Backward Compatible**: Works with existing JSON files containing untrimmed names
5. **Future-Proof**: New stations are automatically trimmed when saved

## Testing Recommendations

1. Test saving new stations - verify names are trimmed
2. Test playing stations from lists - verify alphabetical order
3. Test deleting stations - verify correct station is deleted
4. Test with existing favorites - verify backward compatibility
5. Test with station names that have unusual whitespace patterns

## User Impact

**Positive Changes:**
- Stations now appear in alphabetical order (case-insensitive)
- No more confusing whitespace in station names
- Easier to find specific stations in long lists

**No Breaking Changes:**
- Existing favorites continue to work
- All functionality preserved
- Index-based selection replaced with more reliable name-based selection

## Example Before/After

### Before:
```
1) SmoothJazz.com 64k aac+
2)   BBC Radio 1
3) Jazz FM
4)   Classical Music
```

### After:
```
1) BBC Radio 1
2) Classical Music
3) Jazz FM
4) SmoothJazz.com 64k aac+
```

## Notes

- The sorting is case-insensitive, so "BBC Radio" and "abc station" will sort together properly
- Whitespace trimming applies to leading and trailing spaces only (internal spaces are preserved)
- The changes are transparent to users - they just see better-organized lists
