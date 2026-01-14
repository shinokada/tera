# File Rename: sample.json → myfavorite.json

## Change Summary

Renamed the default favorite list file from `sample.json` to `myfavorite.json` for better clarity and user experience.

## Motivation

**Before:** `~/.config/tera/favorite/sample.json`
- "sample" is confusing and sounds temporary
- Doesn't convey that it's the user's actual favorite list
- Technical terminology not user-friendly

**After:** `~/.config/tera/favorite/myfavorite.json`
- "myfavorite" is clear and personal
- Immediately communicates purpose
- More intuitive for users

## Files Modified

### 1. `tera` (Main Program)
**Changes:**
- Line 10: `FAVORITE_FILE="myfavorite.json"` (was "sample.json")
- Lines 81-93: Added automatic migration from old filename
- Line 93: Updated comment to reference myfavorite.json

**Migration Logic:**
```bash
# Initialize myfavorite.json if it doesn't exist
if [ ! -f "$FAVORITE_FULL" ]; then
    # Check if old sample.json exists and migrate it
    if [ -f "${FAVORITE_PATH}/sample.json" ]; then
        mv "${FAVORITE_PATH}/sample.json" "$FAVORITE_FULL"
        greenprint "Migrated your favorites from sample.json to myfavorite.json"
        sleep 1
    else
        # Create new myfavorite.json from template
        touch "$FAVORITE_FULL"
        cp "${script_dir}/lib/sample.json" "$FAVORITE_FULL"
    fi
fi
```

### 2. `lib/lib.sh`
**Changes:**
- Line 190: Updated comment to reference myfavorite.json
- Line 191: `FAVORITE_STATIONS_FILE="${FAVORITE_PATH}/myfavorite.json"`

### 3. `lib/search.sh`
**Changes:**
- Line 23: Comment updated to reference myfavorite.json
- Line 24: Check changed from `"sample"` to `"myfavorite"`
- Line 53: Comment updated
- Line 55: Conversion changed from `"sample"` to `"myfavorite"`

### 4. `docs/README.md`
**Changes:**
- Updated file path in Saving Stations section
- Changed from `sample.json` to `myfavorite.json`

## Display Name Mapping

The system still displays "Favorite" to users, but internally uses "myfavorite":

| What User Sees | Actual Filename | Full Path |
|----------------|-----------------|-----------|
| "Favorite" | `myfavorite.json` | `~/.config/tera/favorite/myfavorite.json` |
| "MyJazz" | `MyJazz.json` | `~/.config/tera/favorite/MyJazz.json` |
| "Classical" | `Classical.json` | `~/.config/tera/favorite/Classical.json` |

## Migration Path

### For Existing Users

When users upgrade to this version:

1. **If they have `sample.json`:**
   - File is automatically renamed to `myfavorite.json` on first run
   - All stations are preserved
   - User sees: "Migrated your favorites from sample.json to myfavorite.json"
   - No data loss

2. **If they don't have any favorites yet:**
   - New `myfavorite.json` is created from template
   - Starts with the default sample stations

### For New Users

- First run creates `~/.config/tera/favorite/myfavorite.json`
- Populated with sample stations from `lib/sample.json`
- Clean, intuitive experience from the start

## Technical Details

### Constants Affected

```bash
# Before
FAVORITE_FILE="sample.json"
FAVORITE_FULL="${FAVORITE_PATH}/sample.json"

# After
FAVORITE_FILE="myfavorite.json"
FAVORITE_FULL="${FAVORITE_PATH}/myfavorite.json"
```

### Function Updates

All functions that referenced the file have been updated:
- `menu()` in `tera`
- `_play_favorite_station()` in `lib/lib.sh`
- `_save_station_to_list()` in `lib/search.sh`

### lib/sample.json

- Remains unchanged
- Still serves as the template for initializing favorites
- Located at: `lib/sample.json` in the repository
- Only used during initialization, not at runtime

## Testing

### Test Migration
```bash
# Create an old sample.json
mkdir -p ~/.config/tera/favorite
echo '[]' > ~/.config/tera/favorite/sample.json

# Run tera
tera

# Should see: "Migrated your favorites from sample.json to myfavorite.json"
# Verify file exists: ls ~/.config/tera/favorite/myfavorite.json
```

### Test New Installation
```bash
# Remove favorites directory
rm -rf ~/.config/tera/favorite

# Run tera
tera

# Verify myfavorite.json was created
ls ~/.config/tera/favorite/myfavorite.json
```

### Test Save Functionality
```bash
# Search and save a station to "Favorite"
# Check it was saved to myfavorite.json
cat ~/.config/tera/favorite/myfavorite.json
```

### Test Quick Play Favorites
```bash
# Verify saved stations appear in main menu
# Play a station from Quick Play Favorites
# Confirm it plays correctly
```

## Backward Compatibility

✅ **Fully backward compatible**
- Existing users' data is automatically migrated
- No manual intervention required
- Zero data loss
- Seamless transition

## Benefits

1. **Clarity:** "myfavorite" is self-explanatory
2. **Ownership:** "my" prefix makes it personal
3. **Professional:** Better naming convention
4. **Consistency:** Aligns with user expectations
5. **Migration:** Automatic, seamless upgrade path

## Impact on Other Files

The following files are **NOT affected** and remain unchanged:
- `lib/sample.json` - Template file
- `lib/delete_station.sh` - Works with any list name
- `lib/list.sh` - List management independent of filenames
- `lib/play.sh` - Plays from any list
- `lib/lucky.sh` - Doesn't interact with favorites
- `lib/gistlib.sh` - Works with any list

## Version Information

This change is part of TERA v0.6.0+ updates.

## Summary

| Aspect | Old | New |
|--------|-----|-----|
| Filename | `sample.json` | `myfavorite.json` |
| Full Path | `~/.config/tera/favorite/sample.json` | `~/.config/tera/favorite/myfavorite.json` |
| Display Name | "Favorite" (confusing) | "Favorite" (clear) |
| Migration | N/A | Automatic on upgrade |
| User Impact | Confusing terminology | Clear, intuitive |

**Result:** Better user experience with zero disruption for existing users! 🎉
