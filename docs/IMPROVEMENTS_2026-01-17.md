# TERA Improvements Summary - January 17, 2026

## 1. Bug Fixes

### lib/gistlib.sh - Undefined Variable
**Issue**: `script_dir` variable was undefined, causing incomplete error messages.
**Fix**: Changed to use simpler path without variables.
```bash
# Before
yellowprint "Or run: cp ${script_dir}/.env.example ${script_dir}/.env"

# After  
yellowprint "Or run: cp .env.example .env"
```

### lib/gistlib.sh - Missing return statement
**Issue**: Success path in `create_gist()` was missing `return` after `gist_menu` call.
**Fix**: Added `return` for consistency and to prevent potential bugs if code is added later.
```bash
# Before
read -p "Press Enter to return to menu..."
gist_menu

# After
read -p "Press Enter to return to menu..."
gist_menu
return
```

### lib/gistlib.sh - Working directory not restored
**Issue**: `cd "$FAVORITE_PATH"` changed directory but never restored it, potentially affecting subsequent operations.
**Fix**: Used `pushd`/`popd` to preserve and restore working directory.
```bash
# Before
cd "$FAVORITE_PATH" || {
    redprint "Error: Could not access favorites directory."
    sleep 2
    gist_menu
    return
}
# ... operations ...
gist_menu

# After
pushd "$FAVORITE_PATH" > /dev/null || {
    redprint "Error: Could not access favorites directory."
    sleep 2
    gist_menu
    return
}
# ... operations ...
popd > /dev/null
gist_menu
```

**Important**: Also added `popd` before early return when gist directory determination fails.

## 2. Test Improvements

### test_integration.bats - Ordering Verification
**Issue**: Tests only checked if `clear` and heading existed, not their order.
**Fix**: Added line number comparison to verify `clear` comes before heading.

### test_integration.bats - Test Name Alignment
**Issue**: Test name said "No double" but assertion checked for ">= 1".
**Fix**: Renamed to "Main Menu entry exists in menus" to match actual test.

### test_menu_structure.bats - Submenu Verification
**Issue**: Test didn't verify Main Menu was in the submenu function body.
**Fix**: Scoped search to the `search_submenu()` function using awk.

### test_search.bats - Function-Scoped Tests
**Issue**: Tests searched entire file instead of specific functions.
**Fix**: Used awk to scope grep to `_wget_simple_search()` and `_wget_search()` functions.

## 3. GitHub Actions Updates

### .github/workflows/test.yml
**Issue**: Using deprecated `actions/checkout@v3` with Node.js 16.
**Fix**: Upgraded to `actions/checkout@v4` (Node.js 20).
**Removed**: Redundant `continue-on-error: false`.

## 4. Documentation Improvements

### docs/GIST_SETUP.md - Style Enhancement
**Issue**: Word "only" appeared twice in close proximity.
**Fix**: Rephrased for better readability.
```markdown
# Before
- ℹ️ The token is only stored locally on your machine
- ℹ️ Only the `gist` scope is needed

# After
- ℹ️ The token is stored locally on your machine
- ℹ️ The `gist` scope is the only permission needed
```

## 5. New Feature: Duplicate Detection

### lib/search.sh - Prevent Duplicate Stations
**Feature**: Added duplicate detection when saving stations to lists.

**Implementation**:
```bash
# Check if station already exists in the list (by stationuuid)
STATION_UUID=$(jq -r '.stationuuid' "$TEMP_FILE")
EXISTING_STATION=$(jq --arg uuid "$STATION_UUID" '.[] | select(.stationuuid == $uuid)' "$FAVORITE_FULL" 2>/dev/null)

if [ -n "$EXISTING_STATION" ]; then
    yellowprint "This station is already in your $DISPLAY_NAME list!"
    read -p "Press Enter to continue..."
    return
fi
```

**Benefits**:
- Prevents duplicate entries in favorite lists
- Uses station UUID for accurate detection
- Provides clear user feedback
- Maintains data integrity

### tests/test_duplicates.bats - New Test Suite
**Added**: Comprehensive tests for duplicate detection feature.

Tests verify:
- UUID extraction from station data
- Duplicate checking logic
- Warning message display
- Early return on duplicate
- Correct operation order

### docs/CHANGELOG.md - New Documentation
**Added**: Changelog documenting the duplicate detection feature.

## Summary of Files Modified

### Bug Fixes (1 file)
- `lib/gistlib.sh` - Fixed undefined variable, added missing return, and restored working directory with pushd/popd

### Test Improvements (5 files)
- `tests/test_integration.bats`
- `tests/test_menu_structure.bats`
- `tests/test_search.bats`
- `tests/test_duplicates.bats` (NEW)
- `tests/test_gist_improvements.bats` (NEW)

### GitHub Actions (1 file)
- `.github/workflows/test.yml`

### Documentation (2 files)
- `docs/GIST_SETUP.md`
- `docs/CHANGELOG.md` (NEW)

### Features (1 file)
- `lib/search.sh`

## Testing

All improvements have corresponding tests:
```bash
cd tests
bats test_duplicates.bats          # New duplicate detection tests
bats test_gist_improvements.bats   # New gist function tests
bats test_integration.bats         # Updated ordering tests
bats test_menu_structure.bats      # Updated submenu tests
bats test_search.bats              # Updated function-scoped tests
```

## Impact

✅ **Code Quality**: Fixed undefined variable bug and added proper cleanup
✅ **Function Consistency**: All gist_menu calls now properly return
✅ **Directory Safety**: Working directory is preserved and restored
✅ **Test Coverage**: Improved test accuracy and scope
✅ **CI/CD**: Updated to latest GitHub Actions
✅ **Documentation**: Clearer, more polished docs
✅ **User Experience**: No more duplicate stations in lists
✅ **Maintainability**: Better tests mean easier future changes

## Next Steps

Potential future enhancements:
- Add option to merge/deduplicate existing lists
- Support duplicate detection across multiple lists
- Add "similar station" detection (fuzzy matching)
- Provide option to update existing station details
