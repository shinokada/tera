# Search Results UI Improvements

## Issues Fixed

### 1. ✅ Added Empty Line at Top
**Before:**
```text
Search Results (308 stations)
101 SMOOTH JAZZ • ...
```

**After:**
```text
(empty line for breathing room)
Search Results (308 stations)
101 SMOOTH JAZZ • ...
```

**Change:**
```go
// In View() for searchStateResults
s.WriteString("\n")  // Add empty line at top
s.WriteString(m.resultsList.View())
```

### 2. ✅ Increased List Height to Use More Screen Space

**Problem:** Only 2 stations visible in fullscreen terminal

**Root Cause:** 
- List created with `m.height - 10` (too conservative)
- WindowSizeMsg used `msg.Height - 8` (also too conservative)

**Solution:**
Changed both to use `height - 4` to maximize visible stations:

```go
// In searchResultsMsg handler
listHeight := m.height - 4  // Was: m.height - 10
if listHeight < 5 {
    listHeight = 5
}
m.resultsList = list.New(m.resultsItems, delegate, m.width, listHeight)

// In tea.WindowSizeMsg handler
listHeight := msg.Height - 4  // Was: msg.Height - 8
```

**Reasoning:**
- List component includes its own title
- List component includes its own help text  
- List component includes status bar
- We only add 3 extra lines:
  1. Empty line at top
  2. Empty line after list
  3. Our custom footer ("Enter) Play | /) Filter | Esc) Back")
- So `height - 4` leaves perfect room

### Expected Result

**Before (with height - 10):**
- Terminal height: 50 lines
- List height: 40 lines
- ~2-3 stations visible (due to spacing)

**After (with height - 4):**
- Terminal height: 50 lines
- List height: 46 lines  
- ~10-15 stations visible (much better!)

**Fullscreen Terminal Example:**
- Terminal height: 60 lines
- List height: 56 lines
- ~15-20 stations visible

## Files Modified
- `internal/ui/search.go`
  - Added empty line at top of results view
  - Changed list height from `height - 10` to `height - 4`
  - Changed WindowSizeMsg calculation from `height - 8` to `height - 4`

## Benefits
1. ✅ Better visual breathing room at the top
2. ✅ Maximum screen real estate usage
3. ✅ More stations visible without scrolling
4. ✅ Better UX for browsing large result sets
5. ✅ Consistent with modern terminal app design

## Testing
- [x] Test with various terminal sizes (80x24, 120x40, fullscreen)
- [x] Verify list scrolls properly when results exceed visible area
- [x] Check that footer remains visible
- [x] Test filter functionality still works
- [x] Verify no visual glitches during window resize
