# Search Screen Implementation - Complete

This document describes the completed Search Screen implementation for TERA in Go.

## Overview

The Search Screen allows users to discover new radio stations by searching the Radio Browser API using various criteria such as tags, station name, language, country, state, or advanced multi-criteria searches.

## Implementation Files

### 1. API Layer (`internal/api/search.go`)
**Purpose**: Handles all Radio Browser API search operations

**Key Features**:
- Multiple search types (Tag, Name, Language, Country, State, Advanced)
- Flexible search parameters with ordering and pagination
- Default sorting by votes (highest first)
- Automatic hiding of broken stations
- Query parameter trimming and sanitization

**Main Functions**:
```go
func (c *Client) Search(params SearchParams) ([]Station, error)
func (c *Client) SearchByTag(tag string) ([]Station, error)
func (c *Client) SearchByName(name string) ([]Station, error)
func (c *Client) SearchByLanguage(language string) ([]Station, error)
func (c *Client) SearchByCountry(country string) ([]Station, error)
func (c *Client) SearchByState(state string) ([]Station, error)
func (c *Client) SearchAdvanced(params SearchParams) ([]Station, error)
```

### 2. UI Layer (`internal/ui/search.go`)
**Purpose**: Implements the Bubble Tea UI for the search screen

**States**:
- `searchStateMenu`: Main search menu selection
- `searchStateInput`: Text input for search query
- `searchStateLoading`: API call in progress
- `searchStateResults`: Display search results
- `searchStateStationInfo`: Station details and action menu
- `searchStatePlaying`: Playback active

**Key Features**:
- Clean state machine architecture
- fzf-style filtering of search results
- Real-time duplicate checking against Quick Favorites
- Save to Quick Favorites during or after playback
- Consistent navigation (0, 00, Esc)
- Loading spinner during API calls
- Error handling with user-friendly messages

### 3. App Integration (`internal/ui/app.go`)
**Purpose**: Integrates Search Screen into the main app

**Changes**:
- Added `searchScreen` to App struct
- Added `screenSearch` to Screen enum
- Implemented navigation to/from Search Screen
- Added API client initialization
- Updated main menu to show "Search Stations" option

## Search Flow

### Flow Chart Implementation

Based on `golang/spec-docs/flow-charts.md`, the implementation follows this flow:

1. **Search Menu** → User selects search type (1-6)
2. **Text Input** → User enters search query
   - Navigation: `0` = back to menu, `00`/`Esc` = main menu
3. **Loading** → API call with spinner feedback
4. **Results** → fzf-style list with filtering
   - Empty results: Show friendly message
   - Results found: Display with sorting by votes
5. **Station Info** → Show details and action menu
   - 1: Play station
   - 2: Save to Quick Favorites
   - 3: Back to results
6. **Playing** → Active playback with controls
   - `q`/`Esc`/`0`: Stop playback
   - `s`: Save to Quick Favorites (during playback)

### Navigation Shortcuts

As specified in `keyboard-shortcuts-guide.md`:
- `0`: Back to previous screen
- `00`: Return to main menu
- `Esc`: Cancel/Back
- `/`: Filter results (in results view)
- `Enter`: Select/Confirm

## Key Design Decisions

### 1. Duplicate Prevention
The implementation loads Quick Favorites (My-favorites.json) on initialization and checks for duplicates by `StationUUID` before saving. This prevents users from accidentally saving the same station multiple times.

### 2. Save Prompt After Playback
Following the specification, search results show a save prompt after playback ends. This is different from Play Screen behavior where Quick Favorites don't show save prompts.

### 3. Default Sort Order
Results are sorted by votes (descending) to show the most popular/reliable stations first. This matches the specification's requirement for "Order by votes, reverse=true".

### 4. Error Handling
All API errors are caught and displayed to the user with helpful context, then return to the search menu rather than crashing the app.

### 5. Text Input State
The text input field is properly focused when entering input state and cleared after search, providing a clean UX.

## Tests

### API Tests (`internal/api/search_test.go`)

**Test Coverage**:
- ✅ All search methods (Tag, Name, Language, Country, State, Advanced)
- ✅ Endpoint building logic
- ✅ Query parameter construction
- ✅ Error handling (server errors, invalid JSON)
- ✅ Query trimming/sanitization
- ✅ HTTP headers and request formatting

**Test Functions**:
```go
func TestSearch(t *testing.T)
func TestBuildSearchEndpoint(t *testing.T)
func TestBuildQueryParams(t *testing.T)
func TestSearchAdvanced(t *testing.T)
func TestSearchErrorHandling(t *testing.T)
func TestSearchTrimming(t *testing.T)
```

### UI Tests (`internal/ui/search_test.go`)

**Test Coverage**:
- ✅ Model initialization
- ✅ Search menu navigation (all 6 search types)
- ✅ Back navigation (0, Esc, 00)
- ✅ Text input handling
- ✅ Empty query validation
- ✅ Search results processing
- ✅ Error handling
- ✅ Station selection
- ✅ Search type labels
- ✅ Station info menu
- ✅ Window resize
- ✅ Quick Favorites loading
- ✅ Station details rendering

**Test Functions**:
```go
func TestSearchModelInit(t *testing.T)
func TestSearchMenuNavigation(t *testing.T)
func TestSearchBackNavigation(t *testing.T)
func TestSearchTextInput(t *testing.T)
func TestSearchEmptyQuery(t *testing.T)
func TestSearchResults(t *testing.T)
func TestSearchError(t *testing.T)
func TestSearchStationSelection(t *testing.T)
func TestSearchTypeLabels(t *testing.T)
func TestStationInfoMenu(t *testing.T)
func TestWindowResize(t *testing.T)
func TestQuickFavoritesLoading(t *testing.T)
func TestRenderStationDetails(t *testing.T)
```

## Running Tests

```bash
# Run all search-related tests
go test ./internal/api -v -run Search
go test ./internal/ui -v -run Search

# Run with coverage
go test ./internal/api -cover -v
go test ./internal/ui -cover -v

# Run all tests
go test ./... -v
```

## Usage

From the main menu:
1. Press `2` to enter Search screen
2. Select search type (1-6)
3. Enter search query
4. Browse results with arrow keys or filter with `/`
5. Press Enter to select a station
6. Choose action: Play (1), Save (2), or Back (3)
7. During playback: Press `s` to save to Quick Favorites
8. Navigate back with `0`, `00`, or `Esc`

## Future Enhancements

Potential improvements for future iterations:
- [ ] Save to any list (not just Quick Favorites)
- [ ] Advanced search UI with multiple input fields
- [ ] Search history
- [ ] Filter results by codec/bitrate
- [ ] Sort results by different criteria (bitrate, name, etc.)
- [ ] Pagination for large result sets (>100 stations)
- [ ] Recently played from search

## Specification Compliance

✅ All requirements from `flow-charts.md` Section 3 implemented
✅ All navigation shortcuts from `keyboard-shortcuts-guide.md` implemented
✅ API integration following `API_SPEC.md` patterns
✅ Consistent with Play Screen architecture
✅ Proper error handling and user feedback
✅ Test coverage for all major features

## Related Files

- Specification: `golang/spec-docs/flow-charts.md` (Section 3)
- Keyboard Guide: `golang/spec-docs/keyboard-shortcuts-guide.md`
- API Spec: `golang/API_SPEC.md`
- Models: `internal/api/models.go`
- Storage: `internal/storage/favorites.go`
- Player: `internal/player/mpv.go`

## Summary

The Search Screen implementation is **complete** and **fully tested**. It provides:
- 6 different search types covering all common use cases
- Clean, intuitive UI following TERA's design patterns
- Robust error handling
- Comprehensive test coverage (18 test functions)
- Full integration with the main app
- Consistent navigation and keyboard shortcuts
- Duplicate prevention for Quick Favorites

The implementation follows all specifications and is ready for use.
