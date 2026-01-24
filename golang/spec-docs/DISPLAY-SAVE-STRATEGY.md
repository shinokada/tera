# Display & Save Strategy Quick Reference

## UI Display Strategy

| Screen | Lists | Stations/Results | Rationale |
|--------|-------|------------------|-----------|
| **Main Menu** | N/A | Simple (10 quick fav) | Known items, direct access |
| **Play Screen** | Arrow nav | fzf-style | Few lists, many stations |
| **Search Results** | N/A | fzf-style | 100s-1000s of results |
| **Delete Station** | Arrow nav | fzf-style | Few lists, many stations |
| **List Management** | Simple menu | N/A | Few operations |
| **Gist Menu** | Simple menu | N/A | 6 fixed options |
| **My Gists** | Arrow nav | N/A | 1-10 gists typically |

### Rules of Thumb
- **< 15 items + user knows names** → Simple arrow navigation
- **> 15 items OR need filtering** → fzf-style
- **100s-1000s items** → Always fzf-style

---

## Save to Favorites Strategy

| Screen | Context | When to Save | How to Save | Prompt After? |
|--------|---------|--------------|-------------|---------------|
| **QuickPlay (10-19)** | Already in My-favorites | Never | N/A | ❌ No |
| **Play from My-favorites** | Already in My-favorites | Never | N/A | ❌ No |
| **Play from other list** | Curation/Promotion | Optional during | Press 's' key | ❌ No |
| **Search Results** | NEW discovery | After playback | User chooses | ✅ Yes |
| **Lucky** | NEW discovery | After playback | User chooses | ✅ Yes |

### Save Logic Decision Tree

```
Is station from QuickPlay (10-19)?
  └─> YES: Never save (already in My-favorites)
  
Is station from "My-favorites" list?
  └─> YES: Never save (already in My-favorites)
  
Is station from another favorites list?
  └─> Check: Already in My-favorites?
      ├─> YES: No save option
      └─> NO: Show 's' key during playback
  
Is station from Search or Lucky?
  └─> Prompt after playback stops
```

---

## Check: Should Save Option Be Shown?

```go
func shouldShowSaveOption(station Station, listName string) bool {
    // Never save from My-favorites or QuickPlay
    if listName == "My-favorites" || listName == "" {
        return false
    }
    
    // Check if already in Quick Favorites
    if isInQuickFavorites(station.StationUUID) {
        return false
    }
    
    // Show save option
    return true
}
```

---

## User-Facing Messages

### When Save Not Shown
```
[No message - 's' key just not mentioned in help]
```

### When User Presses 's' (Already Saved)
```
⭐ Already in Quick Favorites!
```

### When User Presses 's' (Success)
```
✓ Added to Quick Favorites!
You can access it from the Main Menu.
```

### Save Prompt After Playback (Search/Lucky)
```
Did you enjoy this station?

1) ⭐ Add to Quick Favorites
2) Return to [Previous Screen]
```

---

## Implementation Checklist

### For Each Screen That Plays Stations

- [ ] Determine context (QuickPlay, My-favorites, Other list, Search, Lucky)
- [ ] Check if station is in Quick Favorites
- [ ] Render appropriate UI (show or hide 's' key)
- [ ] Handle 's' key press with duplicate checking
- [ ] Show appropriate prompt after playback (if applicable)

### Duplicate Checking

```go
// Always check by StationUUID, never by name
func isInQuickFavorites(stationUUID string) bool {
    favorites, _ := storage.LoadList("My-favorites")
    for _, station := range favorites.Stations {
        if station.StationUUID == stationUUID {
            return true
        }
    }
    return false
}
```

---

## Test Matrix

| Test Case | List | Already in Quick? | Show 's'? | Prompt After? |
|-----------|------|-------------------|-----------|---------------|
| QuickPlay #10 | My-favorites | Yes | ❌ | ❌ |
| Play My-favorites → Station A | My-favorites | Yes | ❌ | ❌ |
| Play Jazz → Station B | Jazz | No | ✅ | ❌ |
| Play Jazz → Station C | Jazz | Yes | ❌ | ❌ |
| Search → Play Station D | (none) | No | N/A | ✅ |
| Search → Play Station E | (none) | Yes | N/A | ✅ (with "already saved") |
| Lucky → Station F | (none) | No | N/A | ✅ |
| Lucky → Station G | (none) | Yes | N/A | ✅ (with "already saved") |

---

## Common Scenarios

### Scenario: User Building Quick Favorites
```
1. Search for "jazz"
2. Find cool station
3. Play it
4. Like it → Save prompt appears
5. Choose "Add to Quick Favorites"
6. Station saved to My-favorites.json
7. Now appears in Main Menu as #10-19
```

### Scenario: User Curating from Existing Lists
```
1. Go to Play Screen
2. Select "Classical" list
3. Browse stations
4. Play "WQXR"
5. Really like it → Press 's' during playback
6. Added to Quick Favorites
7. Continue browsing Classical list
```

### Scenario: User Playing Quick Favorites
```
1. At Main Menu
2. Press "10" (first quick favorite)
3. Station plays
4. Press 'q' to stop
5. Back to Main Menu
6. No interruptions, no prompts
```

---

## Key Principles

1. **Never ask to save what's already saved** - Check by StationUUID
2. **Context matters** - Discovery vs. Curation have different UX
3. **One save method per context** - Either 's' key OR prompt after, never both
4. **Fail gracefully** - If check fails, don't show save option
5. **Clear feedback** - Always confirm save or explain why not

---

## FAQ

**Q: Why not always show the save option?**
A: Redundant prompts are annoying. If it's already in Quick Favorites, there's no point.

**Q: Why not prompt after playback in Play Screen?**
A: Play Screen is for curating existing favorites. User can press 's' if they want to promote. No interruption needed.

**Q: Why DO we prompt in Search and Lucky?**
A: These are discovery contexts. User just found something new. We want to help them save it.

**Q: What if someone plays from My-favorites and presses 's'?**
A: The 's' key won't be shown in the help text, but if somehow pressed, show: "Already in Quick Favorites!"

**Q: Can I save to other lists besides Quick Favorites?**
A: Yes, in Search Results there's a "Save to List" option that lets you choose any list.

---

## Visual Guide

```
Main Menu
├── QuickPlay (10-19)
│   └── Plays from My-favorites.json
│       └── No save (already saved)
│
├── Play Screen (1)
│   ├── My-favorites list
│   │   └── No save (already saved)
│   └── Other lists
│       ├── Station in Quick? → No save
│       └── Station not in Quick? → Show 's' key
│
├── Search (2)
│   └── Search Results
│       └── Play station
│           └── Prompt after playback
│
└── Lucky (5)
    └── Random station
        └── Prompt after playback
```

---

This quick reference should help during implementation. Keep it handy!
