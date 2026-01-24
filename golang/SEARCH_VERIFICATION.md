# Search Screen Verification Guide

This guide helps you verify that the Search Screen implementation is working correctly.

## Pre-flight Checks

### 1. Dependencies Check
```bash
# Ensure all dependencies are present
go mod tidy
go mod download

# Check for any missing imports
go build ./...
```

### 2. Run Tests
```bash
# Make test script executable
chmod +x run_search_tests.sh

# Run all search tests
./run_search_tests.sh
```

Expected output:
```
================================
TERA Search Screen Test Suite
================================

Running API Search Tests...
----------------------------
✓ API Search Tests Passed

Running UI Search Tests...
----------------------------
✓ UI Search Tests Passed

================================
Test Summary
================================
Total Test Suites: 2
Passed: 2
Failed: 0

All search tests passed! ✓
```

## Manual Testing Checklist

### Search Menu Testing

#### Test 1: Navigate to Search Screen
1. Build and run: `go run cmd/tera/main.go`
2. From main menu, press `2`
3. ✅ Should see "🔍 Search Radio Stations" title
4. ✅ Should see options 1-6 for different search types
5. ✅ Should see "0/Esc) Back to Main Menu" help text

#### Test 2: Back Navigation from Menu
1. From search menu, press `0`
2. ✅ Should return to main menu
3. Return to search (press `2`)
4. Press `Esc`
5. ✅ Should return to main menu

### Text Input Testing

#### Test 3: Search by Tag
1. From search menu, press `1`
2. ✅ Should see "Search by Tag" label
3. ✅ Placeholder should suggest "jazz, rock, news"
4. Type `jazz` and press `Enter`
5. ✅ Should show loading spinner
6. ✅ After 1-3 seconds, should display results

#### Test 4: Empty Query Validation
1. From search menu, press `1`
2. Press `Enter` without typing anything
3. ✅ Should remain on input screen (not search)
4. ✅ No API call should be made

#### Test 5: Input Navigation
1. From search menu, press `2`
2. Type some text
3. Press `0`
4. ✅ Should return to search menu
5. ✅ Text should be cleared
6. Press `2` again, type text
7. Press `Esc`
8. ✅ Should return to main menu

### Search Results Testing

#### Test 6: Browse Results
1. Search for "jazz" (or any popular tag)
2. ✅ Should see results list with station names
3. ✅ Each station should show country and codec info
4. ✅ Title should show count: "Search Results (X stations)"
5. Use arrow keys to navigate
6. ✅ Selection should move up/down
7. ✅ Help text should show filtering option

#### Test 7: Filter Results
1. With results displayed, press `/`
2. ✅ Filter prompt should appear
3. Type part of a station name
4. ✅ List should filter in real-time
5. Press `Esc` to clear filter
6. ✅ Full list should return

#### Test 8: No Results Handling
1. Search for something obscure: "xyzqwerty12345"
2. ✅ Should show "No Results" message
3. ✅ Should show friendly help text
4. Press any key
5. ✅ Should return to search menu

### Station Info Testing

#### Test 9: View Station Details
1. Search for stations and select one with `Enter`
2. ✅ Should show "📻 Station Information"
3. ✅ Should display: Name, Tags, Country, Language, Votes, Codec, Bitrate
4. ✅ Should show menu: 1) Play, 2) Save, 3) Back
5. ✅ Should show navigation help

#### Test 10: Station Info Navigation
1. From station info, press `3`
2. ✅ Should return to search results
3. Select station again
4. Press `0`
5. ✅ Should return to main menu
6. Return and search again
7. Select station, press `Esc`
8. ✅ Should return to results

### Playback Testing

#### Test 11: Play Station
1. Select a station from results
2. Press `1` to play
3. ✅ Should show "🎵 Now Playing"
4. ✅ Should display station details
5. ✅ Should show controls: "q/Esc/0) Stop | s) Save"
6. Wait a few seconds
7. ✅ Should hear audio (if mpv is installed)

#### Test 12: Stop Playback
1. While playing, press `q`
2. ✅ Playback should stop
3. ✅ Should return to search results

#### Test 13: Save During Playback
1. Play a station
2. Press `s`
3. ✅ Should show success message if new station
4. ✅ Should show "Already in Quick Favorites" if duplicate
5. Message should appear briefly then fade

### Save to Quick Favorites Testing

#### Test 14: Save from Station Info
1. Select a station from results
2. Press `2` to save
3. ✅ Should show success message
4. Try saving the same station again
5. ✅ Should show "Already in Quick Favorites"

#### Test 15: Verify Save
1. Save a new station
2. Return to main menu (press `0`)
3. Press `1` for "Play from Favorites"
4. Select "My-favorites" list
5. ✅ Should see the saved station in the list

### Different Search Types Testing

#### Test 16: Search by Name
1. From search menu, press `2`
2. Type `BBC` and press `Enter`
3. ✅ Should find BBC stations
4. ✅ Results should be sorted by votes

#### Test 17: Search by Language
1. From search menu, press `3`
2. Type `english` and press `Enter`
3. ✅ Should find English language stations

#### Test 18: Search by Country
1. From search menu, press `4`
2. Type `US` and press `Enter`
3. ✅ Should find US stations

#### Test 19: Search by State
1. From search menu, press `5`
2. Type `California` and press `Enter`
3. ✅ Should find California stations

#### Test 20: Advanced Search
1. From search menu, press `6`
2. Type a general query
3. ✅ Should search across multiple fields
4. ✅ Results should be relevant

### Error Handling Testing

#### Test 21: Network Error (Optional)
1. Disconnect internet
2. Try to search
3. ✅ Should show error message
4. ✅ Should return to search menu
5. ✅ Error should be displayed clearly

#### Test 22: Invalid API Response (requires mock)
This is covered by automated tests.

## Performance Checks

### Response Times
- ✅ Search should complete in < 3 seconds
- ✅ UI should remain responsive during search
- ✅ No noticeable lag when navigating

### Memory Usage
```bash
# Monitor memory while using search
go run cmd/tera/main.go &
PID=$!
while kill -0 $PID 2>/dev/null; do
    ps -o rss= -p $PID
    sleep 5
done
```

Expected: Memory should stay < 50MB for normal usage

## Integration Testing

### Test 23: Full User Journey
1. Start app
2. Press `2` for Search
3. Press `1` for Tag search
4. Type `jazz`
5. Select a station
6. Press `1` to play
7. Listen for a few seconds
8. Press `s` to save
9. Press `q` to stop
10. Press `0` to return to main
11. Press `1` for Play
12. Select "My-favorites"
13. ✅ Saved station should be present
14. ✅ Can play the saved station

### Test 24: Multiple Searches
1. Search for `jazz`
2. Browse results
3. Press `Esc` to return to menu
4. Press `2` for Name search
5. Search for `BBC`
6. ✅ New results should replace old ones
7. ✅ No mixing of results

### Test 25: Screen Resize (Terminal)
1. Start app and navigate to search results
2. Resize terminal window
3. ✅ UI should adapt to new size
4. ✅ No text overflow or truncation issues

## Troubleshooting

### Tests Failing
```bash
# Check Go version (need 1.21+)
go version

# Re-download dependencies
go clean -modcache
go mod download

# Run tests with verbose output
go test -v ./internal/api
go test -v ./internal/ui
```

### No Audio During Playback
- Check if mpv is installed: `which mpv`
- Test mpv directly: `mpv <some_radio_url>`
- Check audio output settings

### Slow API Responses
- Check internet connection
- Radio Browser API might be slow/busy
- Try different search queries

### Build Errors
```bash
# Check for syntax errors
go vet ./...

# Format code
go fmt ./...

# Check imports
go mod tidy
```

## Verification Checklist

Before marking as complete, verify:

- [x] All automated tests pass
- [x] Can navigate to search screen from main menu
- [x] All 6 search types work
- [x] Results display correctly
- [x] Can play stations from results
- [x] Can save to Quick Favorites
- [x] Duplicate prevention works
- [x] Navigation shortcuts work (0, 00, Esc)
- [x] Filtering works (/)
- [x] Error messages display properly
- [x] UI is responsive and doesn't hang
- [x] No memory leaks during extended use
- [x] Integration with Play Screen works
- [x] Help text is clear and accurate

## Success Criteria

✅ **All automated tests pass**
✅ **All manual tests pass**
✅ **No crashes or panics**
✅ **Clean error handling**
✅ **Intuitive user experience**
✅ **Matches specification**

If all checks pass, the Search Screen is production-ready! 🎉
