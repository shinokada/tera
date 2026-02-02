# Search History - Quick Reference Card

## 📦 What Was Created

```
✅ /internal/storage/history.go              - Storage layer (complete)
✅ /internal/storage/history_test.go         - Tests (complete)
✅ /internal/ui/SEARCH_PATCH.md             - Patch guide for search.go
✅ /internal/ui/LUCKY_PATCH.md              - Patch guide for lucky.go
✅ /SEARCH_HISTORY_IMPLEMENTATION.md        - Detailed implementation spec
✅ /IMPLEMENTATION_GUIDE.md                  - Complete implementation guide
```

## 🎯 Quick Apply Guide

### Step 1: Search Screen (`/internal/ui/search.go`)
```go
// 1. Add executeHistorySearch() function at end
// 2. Add renderSearchMenu() function at end
// 3. In handleMenuInput(): Add history quick-select (10+)
// 4. In performSearch(): Add save to history
// 5. In View(): Replace searchStateMenu case
```

### Step 2: Lucky Screen (`/internal/ui/lucky.go`)
```go
// 1. Add searchHistory field to LuckyModel struct
// 2. In NewLuckyModel(): Load history
// 3. In updateInput(): Add quick-select (1-10)
// 4. In searchAndPickRandom(): Save to history
// 5. Replace viewInput() function completely
```

### Step 3: Settings (`/internal/ui/settings.go`)
```go
// 1. Add "Search History" menu item (#2)
// 2. Add settingsStateHistory state
// 3. Add searchHistory field
// 4. Load in NewSettingsModel()
// 5. Add renderHistorySettings() function
// 6. Add handleHistoryInput() function
```

## 🎨 UI Examples

### Search Menu
```
1-6:  Search Types (existing)
10+:  History Items (NEW)
      "10. tag: jazz"
      "11. name: BBC"
```

### Lucky Input
```
1-10: History Items (NEW)
      "1. meditation"
      "2. classical"
```

### Settings > Search History
```
1: Increase (+5)
2: Decrease (-5)
3: Reset (10)
4: Clear All
5: Back
```

## 🔑 Key Functions

### Storage API
```go
// Load history
history, err := store.LoadSearchHistory(ctx)

// Add search item (with dedup)
err := store.AddSearchItem(ctx, "tag", "jazz")

// Add lucky query (with dedup)
err := store.AddLuckyQuery(ctx, "meditation")

// Update size (with auto-trim)
err := store.UpdateHistorySize(ctx, 15)

// Clear all (preserves size)
err := store.ClearSearchHistory(ctx)
```

## 🧪 Test Checklist

Search:
- [ ] Type "10" after seeing history → Runs search
- [ ] Duplicate search → Moves to top
- [ ] 15 searches → Only keeps 10

Lucky:
- [ ] Type "1" after seeing history → Runs search
- [ ] Duplicate → Moves to top

Settings:
- [ ] Increase size → Updates immediately
- [ ] Decrease size → Trims history
- [ ] Clear → Removes all items

## 📝 Code Snippets

### Check if history has items:
```go
if m.searchHistory != nil && len(m.searchHistory.SearchItems) > 0 {
    // Show history
}
```

### Execute history search:
```go
return m.executeHistorySearch(item.SearchType, item.Query)
```

### Save to history:
```go
store := storage.NewStorage(m.favoritePath)
_ = store.AddSearchItem(ctx, "tag", "jazz")
```

## 🎓 Remember

1. Search history includes **type** (tag/name/etc)
2. Lucky history is **query only**
3. Numbers: Search=10+, Lucky=1-10
4. Duplicates → Move to top
5. History survives restarts
6. Background saves (no blocking)

---

**All files are ready!** Just apply the patches from the `.md` guides.
