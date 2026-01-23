# Questions About Flow Charts - Resolved

## Date: January 24, 2026

Based on your analysis of the flow charts and bash implementation, I've addressed all three concerns. Here's what changed:

---

## 1. ✅ Save Prompt After QuickPlay - FIXED

### Issue
Application Overview and Main Menu showed save prompts after QuickPlay (menu items 10-19), but these stations are already in My-favorites.json.

### Resolution
**Removed save prompts from:**
- Application Overview state diagram
- Main Menu flow chart

### Rationale
Looking at the bash code, `_play_favorite_station()` correctly passes empty strings to skip the prompt. The flow charts now match this implementation.

```bash
# Bash implementation (correct):
_play_favorite_station() {
    # Pass empty string for station data and list name to skip the prompt
    _play "$URL_RESOLVED" "" ""
}
```

---

## 2. ✅ fzf-style Display Strategy - CLARIFIED

### Issue
Should we use fzf-style display for all lists in Play Screen, even when there are few items?

### Resolution
**Implemented context-aware display strategy:**

| Content Type | Display Method | Reason |
|--------------|---------------|--------|
| Favorite lists | Simple arrow navigation | Few items (3-10), user knows names |
| Stations within list | fzf-style with filter | Moderate items (10-100), need quick filter |
| Search results | fzf-style with filter | Many items (100s-1000s), essential |
| Gist lists | Simple arrow navigation | Few items (1-10) |
| Menu options | Simple arrow navigation | Fixed small set |

### Rationale
- Bash currently uses fzf for both, which works but is overkill for 3-5 lists
- Go version can be more nuanced
- Better performance (fzf has overhead for small lists)
- Simpler navigation where it makes sense
- Powerful filtering where it's needed

---

## 3. ✅ Two Save Points in Play Screen - REDESIGNED

### Issue
Play Screen had confusing save behavior:
1. During playback: Press 's' to save
2. After playback: Show save prompt

### Resolution
**Implemented context-aware save logic:**

```
IF station is already in Quick Favorites (My-favorites.json):
    - Hide 's' key option during playback
    - No save prompt after playback
    - Just play normally

ELSE IF station is from another list:
    - Show 's' key option during playback ("Press 's' to add to Quick Favorites")
    - No save prompt after playback
    - User can manually promote if desired
```

### New Flow Chart Logic

**Before (Confusing):**
```
Play Station
  └─> During: Press 's' to save
  └─> After: Show prompt to save
      └─> Result: Could save twice!
```

**After (Clear):**
```
Get Station
  └─> Check: Already in Quick Favorites?
      ├─> YES: Play normally (no save options)
      └─> NO: Play with 's' key option
          └─> User presses 's': Save immediately
          └─> User presses 'q': No prompt, just stop
```

### Comparison with Other Screens

| Screen | Save Context | Save Method | Rationale |
|--------|--------------|-------------|-----------|
| **Search Results** | NEW discovery | Prompt after playback | User just found this |
| **Lucky** | NEW discovery | Prompt after playback | Random find, might like it |
| **Play Screen** | Curation | Optional 's' during playback | Promoting existing favorites |
| **QuickPlay** | Already saved | No save option | Already in My-favorites |

---

## Files Created

1. **flow-charts-UPDATED.md**
   - Complete updated flow charts
   - All 14 screens
   - Changed sections clearly marked

2. **FLOW-CHART-CHANGES.md**
   - Detailed change summary
   - Implementation checklist
   - Testing scenarios
   - Code patterns
   - Next steps

3. **THIS FILE (questions/flow-chart-resolution.md)**
   - Quick reference summary
   - Decision rationale

---

## Implementation Impact

### API Changes Needed
```go
// Add to Storage interface
IsStationInList(listName string, stationUUID string) (bool, error)
```

### UI Changes Needed
```go
type PlayScreen struct {
    // ...
    isInQuickFavorites bool    // Check once, cache result
    selectedListName   string  // Track which list we're in
}

func (p *PlayScreen) shouldShowSaveOption() bool {
    return !p.isInQuickFavorites && p.selectedListName != "My-favorites"
}
```

### Testing Scenarios
- ✅ QuickPlay (10-19): No save prompt
- ✅ Play from My-favorites: No save option
- ✅ Play from other list (not in Quick): Show 's' key
- ✅ Play from other list (already in Quick): No save option
- ✅ Search results: Prompt after
- ✅ Lucky: Prompt after

---

## Benefits

### User Experience
- No redundant "do you want to save this?" prompts
- Clear single action: 's' key = "add to my quick favorites"
- Less interruption when browsing stations

### Code Quality
- Clear boolean logic for save eligibility
- Easy to test
- Matches user intent

### Performance
- One duplicate check per play
- fzf only where beneficial
- No unnecessary UI rendering

---

## Next Actions

1. Review the updated flow charts in `flow-charts-UPDATED.md`
2. Check if the logic makes sense for your use case
3. If approved, I can:
   - Replace original `flow-charts.md` with updated version
   - Update other spec documents (API_SPEC.md, keyboard-shortcuts-guide.md, etc.)
   - Add implementation notes to technical-approach.md

---

## Your Call

**Option A (Recommended):** Replace flow-charts.md with the updated version
- Clean slate
- All changes incorporated
- Ready for implementation

**Option B:** Keep both files for comparison
- Original: `flow-charts.md`
- Updated: `flow-charts-UPDATED.md`
- Can diff them later

**Option C:** Request specific changes
- Let me know what needs adjustment
- I'll refine further

---

## Summary

Your instincts were correct on all three points:

1. ✅ **QuickPlay save prompt** - Unnecessary, removed
2. ✅ **fzf for everything** - Overkill, now context-aware
3. ✅ **Two save points** - Confusing, now single context-aware save

The updated flow charts are cleaner, clearer, and match the actual bash implementation patterns while improving the UX for the Go version.
