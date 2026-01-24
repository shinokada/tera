# Search Screen Bug Fixes

## Issues Fixed

### 1. API Test Failures - `baseURL` Assignment Error
**Problem**: Tests attempted to reassign `const baseURL`, causing compilation errors.

**Fix**: Changed `baseURL` from `const` to `var` in `internal/api/client.go`
```go
// Before
const baseURL = "https://de1.api.radio-browser.info/json/stations"

// After
var baseURL = "https://de1.api.radio-browser.info/json/stations"
```

### 2. Duplicate Helper Functions
**Problem**: Functions `contains()` and `findSubstring()` were declared in both `search_test.go` and `play_station_test.go`, causing redeclaration errors.

**Fix**: Removed duplicate helper functions from `internal/ui/search_test.go`, keeping only the versions in `play_station_test.go` which are shared across tests.

### 3. Wrong Player API Signature
**Problem**: `m.player.Play()` was called with `(string, string)` but expects `(*api.Station)`

**Fix**: Updated `playStation()` method in `internal/ui/search.go`
```go
// Before
err := m.player.Play(station.URLResolved, station.Name)

// After
err := m.player.Play(&station)
```

### 4. Non-existent Player Method
**Problem**: Called `m.player.Wait()` which doesn't exist in `MPVPlayer` API

**Fix**: Removed the `Wait()` call since the player runs asynchronously and playback monitoring is handled by the player's internal goroutine.

### 5. Missing Style Definitions
**Problem**: `boldStyle` and `subtleStyle` were undefined in `internal/ui/search.go`

**Fix**: Added missing style definitions to `internal/ui/styles.go`
```go
boldStyle = lipgloss.NewStyle().Bold(true)
subtleStyle = lipgloss.NewStyle().Foreground(colorGray)
```

### 6. Non-existent Error Type
**Problem**: Test referenced `api.ErrEmptyResponse` which doesn't exist

**Fix**: Replaced with generic error in `internal/ui/search_test.go`
```go
// Before
msg := searchErrorMsg{err: api.ErrEmptyResponse}

// After
msg := searchErrorMsg{err: fmt.Errorf("search failed")}
```

### 7. Wrong List Item Type
**Problem**: Used `tea.ListItem` which doesn't exist

**Fix**: Changed to correct `list.Item` type from bubbles/list package and added import
```go
import "github.com/charmbracelet/bubbles/list"

model.resultsItems = []list.Item{
    stationListItem{station: stations[0]},
}
```

### 8. Text Input Not Receiving Key Events
**Problem**: Text input component wasn't receiving keypress events because `handleTextInput()` only checked for special keys and didn't pass other keys to the text input.

**Fix**: Added default case in `handleTextInput()` to pass unhandled keys to text input
```go
default:
    var cmd tea.Cmd
    m.textInput, cmd = m.textInput.Update(msg)
    return m, cmd
```

Also removed duplicate text input update logic at the end of `Update()` function.

### 9. Station Selection Test Failure
**Problem**: Test was manually setting `resultsItems` and `resultsList` fields, but `resultsList` wasn't properly initialized, causing nil pointer dereference.

**Fix**: Changed test to use `searchResultsMsg` to properly initialize the list model through the normal update flow.

### 10. ANSI Escape Codes in Shell Script Output
**Problem**: Shell script displayed literal escape codes like `\033[1;33m` instead of colors.

**Fix**: Removed `echo -e` and color variable references, using plain echo with simple text formatting.

## Files Modified
- `internal/api/client.go` - Changed baseURL to var
- `internal/ui/search.go` - Fixed player API calls, text input handling
- `internal/ui/search_test.go` - Removed duplicates, fixed imports, error references, and station selection test
- `internal/ui/styles.go` - Added missing style definitions
- `run_search_tests.sh` - Removed ANSI escape codes for cleaner output

## Test Status
All tests now compile and pass successfully.
