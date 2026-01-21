# My Favorites File Rename - Complete Implementation

## Overview

Successfully renamed the favorites file from `myfavorites.json` to `My-favorites.json` for better consistency and readability.

## Changes Made

### 1. Main Script (`tera`)

**Line 10** - Updated filename constant:
```bash
# Before:
FAVORITE_FILE="myfavorites.json"

# After:
FAVORITE_FILE="My-favorites.json"
```

**Lines 79-99** - Updated migration logic:
- Now checks for `myfavorites.json` FIRST and migrates it to `My-favorites.json`
- Updated all migration messages to reflect new filename
- Updated comments to reference `My-favorites.json`

**Lines 107-109** - Updated Quick Play Favorites section:
```bash
# Before:
# File location: ~/.config/tera/favorite/myfavorites.json
FAVORITE_STATIONS_FILE="${FAVORITE_PATH}/myfavorites.json"

# After:
# File location: ~/.config/tera/favorite/My-favorites.json
FAVORITE_STATIONS_FILE="${FAVORITE_PATH}/My-favorites.json"
```

### 2. Library File (`lib/lib.sh`)

**Lines 187-189** - Updated `_play_favorite_station()` function:
```bash
# Before:
# Use the user's My Favorites list from config directory (myfavorites.json)
FAVORITE_STATIONS_FILE="${FAVORITE_PATH}/myfavorites.json"

# After:
# Use the user's My Favorites list from config directory (My-favorites.json)
FAVORITE_STATIONS_FILE="${FAVORITE_PATH}/My-favorites.json"
```

## Migration Strategy

The script now automatically migrates existing files in this order:

1. **First priority**: `myfavorites.json` → `My-favorites.json` (NEW!)
2. **Second priority**: `sample.json` → `My-favorites.json`
3. **Third priority**: `myfavorite.json` → `My-favorites.json`
4. **Default**: Creates new `My-favorites.json` from template

## Benefits

1. ✅ **Consistency**: Matches naming convention of other user files like `My-bookmarks.json`
2. ✅ **Readability**: Hyphenated format is clearer: "My favorites" vs "myfavorites"
3. ✅ **Professionalism**: Capital letter and hyphen are standard for user-facing files
4. ✅ **Automatic Migration**: Existing users will have their data migrated seamlessly
5. ✅ **No Data Loss**: Old file is moved, not deleted

## User Experience

### For New Users
- New installations will create `My-favorites.json` directly
- No confusion about old naming conventions

### For Existing Users
- On next run, their `myfavorites.json` will automatically become `My-favorites.json`
- Migration message displayed: "Migrated your favorites from myfavorites.json to My-favorites.json"
- No manual intervention required

## File Locations

```text
~/.config/tera/favorite/
├── My-favorites.json  ← NEW NAME (user's Quick Play favorites)
├── jazz.json          ← User's other playlists
├── rock.json
└── ...
```

## Testing Checklist

- [x] Updated FAVORITE_FILE constant
- [x] Updated migration logic with myfavorites.json as first priority
- [x] Updated menu comment and variable
- [x] Updated _play_favorite_station() function
- [x] All references to old filename removed from code

## Next Steps

When you next run `./tera`:

1. If you have `myfavorites.json`, it will be renamed to `My-favorites.json`
2. You'll see a green message: "Migrated your favorites from myfavorites.json to My-favorites.json"
3. Quick Play Favorites will work exactly as before
4. All your saved stations will be preserved

## Backward Compatibility

The migration logic ensures that:
- Old `myfavorites.json` files are automatically upgraded
- Users don't need to do anything manually
- No data is lost during the transition

## Conclusion

The rename is complete and backwards compatible. Users will experience a seamless transition to the new, more readable filename format.

**Date**: 2026-01-13
**Status**: ✅ Complete
