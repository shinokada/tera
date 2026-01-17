# TERA Changelog

## [Unreleased]

### Added
- **Duplicate Detection**: When saving a radio station to a list, TERA now checks if the station already exists (by `stationuuid`). If a duplicate is found, it displays a friendly warning message instead of adding it again.

### Features
- Prevents duplicate entries in favorite lists
- Uses station UUID for accurate duplicate detection
- Provides clear user feedback when duplicate is detected
- Maintains data integrity in station lists

### Technical Details
The duplicate check in `_save_station_to_list()` function:
1. Extracts the station's UUID from search results
2. Queries the target list for any existing station with the same UUID
3. If found, displays: "This station is already in your [list name] list!"
4. Returns early without adding the duplicate
5. Otherwise, proceeds with normal save operation

### Benefits
- **Cleaner Lists**: No duplicate stations cluttering your favorites
- **Data Integrity**: Maintains unique stations per list
- **Better UX**: Clear feedback when attempting to save duplicates
- **Storage Efficiency**: Reduces redundant data in JSON files

### Example Usage

```bash
# Search and find a station
tera -> Search -> Tag -> jazz

# Try to save to "Jazz Collection"
# If you already saved it before, you'll see:
# "This station is already in your Jazz Collection list!"
# Press Enter to continue...

# Otherwise, it saves successfully:
# "Successfully saved the station to your Jazz Collection list."
```

## Previous Updates

See `docs/README_UPDATES.md` for historical changes.
