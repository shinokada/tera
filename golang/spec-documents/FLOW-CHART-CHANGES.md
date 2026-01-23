# Flow Chart Updates - Change Summary

## Date: 2025-01-24

## Questions Addressed

### 1. Save Prompt After QuickPlay
**Issue:** Application Overview and Main Menu showed save prompts after QuickPlay, but QuickPlay stations are already in My-favorites.json.

**Resolution:** ✅ REMOVED save prompts
- QuickPlay items (10-19) play stations directly from My-favorites.json
- No need to save again to favorites
- Clean flow: Play → Stop → Return to Main Menu

**Changed Documents:**
- Application Overview diagram
- Main Menu Screen flow chart

---

### 2. fzf-style Display Usage
**Issue:** Flow chart showed fzf-style for all lists, but should we use it for small lists?

**Resolution:** ✅ CONTEXT-AWARE DISPLAY
- **Simple Arrow Navigation** for:
  - Favorite lists (3-10 items)
  - Menu options
  - Gist lists (1-10 items)
  - Any list < 15 items
- **fzf-style with Filtering** for:
  - Radio station search results (100s-1000s)
  - Stations within lists (10-100)
  - Any list > 15 items or needing quick filter

**Rationale:**
- Don't overcomplicate navigation for small, known sets
- Provide powerful filtering where actually needed
- Better performance (fzf has overhead)
- Consistent UX within content types

**Changed Documents:**
- Play Screen (lists = arrow nav, stations = fzf)
- Delete Station Screen (lists = arrow nav, stations = fzf)
- Search Results Screen (always fzf - many results)

---

### 3. Two Save Points in Play Screen
**Issue:** Play Screen had TWO ways to save:
1. During playback: Press 's' to save
2. After playback: Save prompt

This was confusing and redundant.

**Resolution:** ✅ CONTEXT-AWARE SAVE BEHAVIOR

**New Logic:**
```
if station already in Quick Favorites (My-favorites.json):
    - Hide 's' key option during playback
    - No save prompt after playback
    - Just play and stop cleanly
else if station from another list:
    - Show 's' key option during playback
    - No save prompt after playback
    - User can manually promote to Quick Favorites if desired
```

**Benefits:**
1. **No redundant prompts** - Only one save method (during playback)
2. **Smart defaults** - Don't offer to save what's already saved
3. **Clear UX** - Users know they can press 's' to promote to Quick Favorites
4. **Consistent with other screens**:
   - Search Results: Prompt after (NEW discovery)
   - Lucky: Prompt after (NEW discovery)  
   - Play Screen: Optional during (promotion/curation)

**Changed Documents:**
- Play Screen flow chart (major rewrite)
- Added duplicate checking by StationUUID
- Context-aware UI rendering

---

## Implementation Checklist

### Updated Flow Charts
- [x] Application Overview - Removed QuickPlay save
- [x] Main Menu - Removed QuickPlay save  
- [x] Play Screen - Full rewrite with context-aware logic
- [x] Delete Station - Specify arrow nav for lists, fzf for stations
- [x] Search Results - Confirm fzf-style usage
- [x] Added UI Display Guidelines Summary

### Documents to Update

#### 1. API_SPEC.md
```go
// Add to Storage interface:
IsStationInList(listName string, stationUUID string) (bool, error)

// Helper for UI logic
func (s *Storage) IsInQuickFavorites(stationUUID string) bool {
    exists, _ := s.IsStationInList("My-favorites", stationUUID)
    return exists
}
```

#### 2. keyboard-shortcuts-guide.md
Add note in Play Screen section:
```
's' key - Save to Quick Favorites (only shown if station not already in Quick Favorites)
```

#### 3. technical-approach.md
Update Play Screen UI section:
```go
type PlayScreen struct {
    // ...
    selectedListName string     // Track which list is being viewed
    isInQuickFavorites bool    // Cache this check
}

func (p *PlayScreen) shouldShowSaveOption() bool {
    // Only show 's' key if:
    // 1. Not already in Quick Favorites
    // 2. User is viewing a different list
    return !p.isInQuickFavorites && p.selectedListName != "My-favorites"
}
```

#### 4. implementation-plan.md
Update Phase 5 (Play Screen):
- Add duplicate checking by StationUUID
- Add context-aware save UI rendering
- Test scenarios:
  - Play from My-favorites → No save option
  - Play from other list + not in Quick Favorites → Show save option
  - Play from other list + already in Quick Favorites → No save option

---

## Testing Scenarios

### Scenario 1: QuickPlay from Main Menu
```
1. User presses "10" (first quick favorite)
2. Station plays
3. User presses 'q'
4. Returns to Main Menu
5. ✅ NO save prompt
```

### Scenario 2: Play from My-favorites List
```
1. Navigate to Play Screen (option 1)
2. Select "My-favorites"
3. Select a station
4. Station plays
5. ✅ NO 's' key shown in help
6. User presses 'q'
7. Returns to list selection
8. ✅ NO save prompt
```

### Scenario 3: Play from Other List (Not in Quick Favorites)
```
1. Navigate to Play Screen
2. Select "Jazz" list
3. Select station "WBGO Jazz"
4. Check: Station not in My-favorites
5. ✅ 's' key shown in help: "Press 's' to add to Quick Favorites"
6. User presses 's'
7. ✅ Show success: "Added to Quick Favorites!"
8. Continue playing
9. User presses 'q'
10. ✅ NO save prompt (already saved)
```

### Scenario 4: Play from Other List (Already in Quick Favorites)
```
1. Navigate to Play Screen
2. Select "Classical" list  
3. Select station "WQXR" (already in Quick Favorites)
4. Check: Station exists in My-favorites
5. ✅ NO 's' key shown
6. Station plays normally
7. User presses 'q'
8. ✅ NO save prompt
```

### Scenario 5: Search and Play (NEW Station)
```
1. Search for "BBC Radio"
2. Select station
3. Play station
4. User presses 'q'
5. ✅ Show prompt: "Add to Quick Favorites?"
6. User selects "Yes"
7. ✅ Station added to My-favorites.json
```

### Scenario 6: Lucky (NEW Station)
```
1. Select "I Feel Lucky" (option 5)
2. Random station plays
3. User enjoys it
4. User presses 's' during playback
5. ✅ Added to Quick Favorites
6. User presses 'q'
7. ✅ NO additional save prompt (already saved)
```

---

## Code Patterns

### Checking if Station is in Quick Favorites
```go
func isInQuickFavorites(storage *Storage, stationUUID string) bool {
    favorites, err := storage.LoadList("My-favorites")
    if err != nil {
        return false
    }
    
    for _, station := range favorites.Stations {
        if station.StationUUID == stationUUID {
            return true
        }
    }
    return false
}
```

### Context-Aware UI Rendering
```go
func renderPlayingInfo(station *Station, listName string, inQuickFav bool) string {
    var b strings.Builder
    
    b.WriteString(stationInfoBox(station))
    b.WriteString("\n\n")
    b.WriteString("Press 'q' to stop\n")
    
    // Only show save option if appropriate
    if !inQuickFav && listName != "My-favorites" {
        b.WriteString("Press 's' to add to Quick Favorites\n")
    }
    
    return b.String()
}
```

### Save During Playback Handler
```go
func (m PlayScreen) handleSaveKeyPress() (tea.Model, tea.Cmd) {
    // Only process if save is allowed
    if m.isInQuickFavorites {
        return m, showMessage("Already in Quick Favorites!")
    }
    
    if m.selectedListName == "My-favorites" {
        return m, showMessage("Already playing from Quick Favorites!")
    }
    
    // Proceed with save
    return m, saveToQuickFavorites(m.currentStation)
}
```

---

## Benefits Summary

### User Experience
- ✅ **No redundant prompts** - Users aren't asked to save what's already saved
- ✅ **Clear intent** - 's' key is for "promotion" to Quick Favorites
- ✅ **Consistent patterns** - Search/Lucky prompt after (discovery), Play allows during (curation)
- ✅ **Less interruption** - Can press 'q' and immediately return to browsing

### Code Quality
- ✅ **Clear separation of concerns** - Different save contexts have different UX
- ✅ **Easy to test** - Clear boolean conditions for save visibility
- ✅ **Maintainable** - Logic is explicit and well-documented

### Performance
- ✅ **One duplicate check** - Check once before rendering UI
- ✅ **No unnecessary prompts** - Skip prompt logic when not needed
- ✅ **Efficient fzf usage** - Only use for lists that benefit from filtering

---

## Migration from Bash

The Bash version has this logic:
```bash
_play_favorite_station() {
    # Pass empty strings to skip the prompt
    _play "$URL_RESOLVED" "" ""
}
```

This is now formally specified in the flow charts and will be implemented consistently in Go with:
- Explicit checking of StationUUID
- Context-aware UI rendering
- Clear boolean flags for save eligibility

---

## Next Steps

1. **Update other documents** (API_SPEC.md, keyboard-shortcuts-guide.md, technical-approach.md)
2. **Review with team** - Ensure this UX makes sense
3. **Implement in Phase 5** - Add context-aware logic to Play Screen
4. **Add unit tests** - Test all scenarios above
5. **User testing** - Validate the improved UX feels natural

---

## Questions Resolved

✅ **Q1:** Should QuickPlay have save prompt?  
**A:** No - stations are already in My-favorites.json

✅ **Q2:** Should we use fzf-style for small lists?  
**A:** No - use simple arrow navigation for < 15 items

✅ **Q3:** Should we have two save points?  
**A:** No - context-aware single save point during playback only

---

## Document Status

- [x] flow-charts-UPDATED.md created with all changes
- [ ] Update original flow-charts.md (or replace with updated version)
- [ ] Update API_SPEC.md with IsStationInList method
- [ ] Update keyboard-shortcuts-guide.md with context notes
- [ ] Update technical-approach.md with implementation patterns
- [ ] Update implementation-plan.md with Phase 5 details
