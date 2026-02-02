# Search History Feature - Complete Implementation Guide

## 🎯 Overview

This feature adds search history to both "Search Stations" and "I Feel Lucky" screens, allowing users to quickly rerun recent searches without retyping. The history size is configurable in Settings.

## ✅ Completed Work

### 1. Storage Layer
- **File**: `/internal/storage/history.go` ✅ Created
- **Tests**: `/internal/storage/history_test.go` ✅ Created
- **Features**:
  - Store up to N recent searches (configurable, default 10)
  - Automatic deduplication (moves duplicate to top)
  - Separate storage for Search and Lucky queries
  - Persistent storage in `~/.config/tera/search-history.json`

## 📝 Remaining Work

You need to apply the patches to these files:

### 2. Search Screen
- **File**: `/internal/ui/search.go`
- **Patch**: `/internal/ui/SEARCH_PATCH.md`
- **Changes**:
  - Added `searchHistory` field to struct ✅
  - Load history on init ✅
  - Handle number keys (10+) for quick search from history
  - Save searches to history automatically
  - Display history in menu view

### 3. Lucky Screen
- **File**: `/internal/ui/lucky.go`
- **Patch**: `/internal/ui/LUCKY_PATCH.md`
- **Changes**:
  - Add `searchHistory` field to struct
  - Load history on init
  - Handle number keys (1-10) for quick search from history
  - Save searches to history automatically
  - Display history in input view

### 4. Settings Screen
- **File**: `/internal/ui/settings.go`
- **Patch**: Create based on SEARCH_HISTORY_IMPLEMENTATION.md
- **Changes**:
  - Add "Search History" menu item
  - Create history settings view
  - Implement increase/decrease/reset/clear actions

## 🎨 UI Reference

### Search Stations with History
```
                         TERA                                                     
                          
      🔍 Search Radio Stations
    > 1. Search by Tag
      2. Search by Name
      3. Search by Language
      4. Search by Country Code
      5. Search by State
      6. Advanced Search

     ─── Recent Searches ───
     10. tag: smooth
     11. tag: jazz
     12. language: english
     13. name: BBC Radio
     14. country: US
     15. state: California
     16. advanced: classical piano
     17. tag: rock
     18. name: NPR
     19. tag: news

  ↑↓/jk: Navigate • Enter: Select • 1-6: Search Type • 10+: Quick Search • Esc: Back
```

### I Feel Lucky with History
```
                         TERA                                                                                             
  I Feel Lucky

  Type a genre of music: rock, classical, jazz, pop, country, hip, heavy, blues, soul.  
  Or type a keyword like: meditation, relax, mozart, Beatles, etc.

  Genre/keyword: > rock, jazz, classical, meditation... 

     ─── Recent Searches ───
     1. jazz
     2. meditation
     3. classical
     4. rock
     5. beatles
     6. blues
     7. piano
     8. 80s
     9. soul
    10. ambient

  Enter: Search • 1-10: Quick search • Esc: Back • Ctrl+C: Quit
```

### Settings > Search History
```
                         TERA                                                
  ⚙️  Settings > Search History

  Current History Size: 10 searches
  (Number of recent searches to keep) 
                                     
    > 1. Increase (+5)      [Will become: 15]
      2. Decrease (-5)      [Will become: 5]
      3. Reset to Default   [Will become: 10]
      4. Clear History      [Removes all saved searches]
      5. Back to Settings
                                                                                
  ↑↓/jk: Navigate • Enter/1-5: Select • Esc: Back • Ctrl+C: Quit
```

## 🚀 How to Apply Patches

### Option 1: Manual Application
1. Open each file mentioned in the patches
2. Find the locations described in each patch
3. Add/modify the code as shown
4. Save and test

### Option 2: Guided Steps

#### For search.go:
```bash
# Open the file
code /Users/shinichiokada/Terminal-Tools/tera/internal/ui/search.go

# Apply changes from SEARCH_PATCH.md in this order:
1. Add the two new functions at the end (executeHistorySearch, renderSearchMenu)
2. Modify handleMenuInput() to add history quick-select
3. Modify performSearch() to save to history
4. Modify View() searchStateMenu case
```

#### For lucky.go:
```bash
# Open the file
code /Users/shinichiokada/Terminal-Tools/tera/internal/ui/lucky.go

# Apply changes from LUCKY_PATCH.md in this order:
1. Add searchHistory field to struct
2. Load history in NewLuckyModel()
3. Modify updateInput() for quick-select
4. Modify searchAndPickRandom() to save history
5. Replace viewInput() function
```

#### For settings.go:
```bash
# Open the file
code /Users/shinichiokada/Terminal-Tools/tera/internal/ui/settings.go

# Apply changes from SEARCH_HISTORY_IMPLEMENTATION.md Section "Settings Screen Updates"
1. Add "Search History" menu item
2. Add settingsStateHistory to enum
3. Add searchHistory field
4. Load history in constructor
5. Handle menu selection for history
6. Add renderHistorySettings() function
7. Add handleHistoryInput() function
```

## 🧪 Testing Plan

After applying all patches, test these scenarios:

### Search Stations:
1. ✅ Perform a search (e.g., "jazz" by tag)
2. ✅ Go back to menu - should see "10. tag: jazz" in history
3. ✅ Press "10" - should immediately search for jazz again
4. ✅ Perform different search types - each should appear with correct prefix
5. ✅ Repeat same search - should move to top (not duplicate)
6. ✅ Perform 15 searches - should only keep last 10 (or current MaxSize)

### I Feel Lucky:
1. ✅ Search for "meditation"
2. ✅ Return to screen - should see "1. meditation" in history
3. ✅ Press "1" - should immediately search again
4. ✅ Add more searches - should appear in order
5. ✅ Duplicate search should move to top

### Settings:
1. ✅ Navigate to Settings > Search History
2. ✅ Increase size - should update and show [Will become: X]
3. ✅ Decrease size - should trim excess history items
4. ✅ Reset - should return to 10
5. ✅ Clear - should remove all items but keep size setting

## 📁 File Structure
```
tera/
├── internal/
│   ├── storage/
│   │   ├── history.go          ✅ Created
│   │   └── history_test.go     ✅ Created
│   └── ui/
│       ├── search.go           🚧 Needs patches
│       ├── lucky.go            🚧 Needs patches
│       ├── settings.go         🚧 Needs patches
│       ├── SEARCH_PATCH.md     ✅ Created (guide)
│       └── LUCKY_PATCH.md      ✅ Created (guide)
└── SEARCH_HISTORY_IMPLEMENTATION.md  ✅ Created (full spec)
```

## 💡 Key Design Decisions

1. **Storage Format**: JSON file for easy inspection and editing
2. **Deduplication**: Moves to top instead of ignoring
3. **Separate Lists**: Search and Lucky have separate histories
4. **Number Ranges**: Search uses 10+ (to avoid conflict with menu 1-6), Lucky uses 1-10
5. **Default Size**: 10 items (configurable)
6. **Persistence**: Survives app restarts
7. **No Duplicates**: Same type+query moves to top
8. **Auto-trim**: When size decreased, oldest items removed

## 🎓 Learning Points

- Search history includes **search type** (tag, name, etc.) so it can replay correctly
- Lucky history is just the query string (always searches by tag)
- History is loaded once at startup and kept in memory
- Saves happen in background goroutines to avoid blocking UI
- Settings changes are immediate and persistent

## ✨ Future Enhancements (Not in this PR)

- Search suggestions based on frequency
- Ability to pin favorite searches
- Export/import history
- Search across history
- Clear individual history items

---

**Ready to implement!** Apply the patches in order and test thoroughly. The storage layer is complete and tested, so the integration should be straightforward.
