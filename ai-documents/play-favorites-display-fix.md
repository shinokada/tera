# Play from Favorites Display Fix

## Problem
When selecting "Play from Favorites", the screen showed pagination (1/2) but no actual list items were visible. Only the title and help text were shown.

## Root Causes

### 1. Timing Issue
- Screen initializes with `width=0, height=0`
- `Init()` immediately loads lists
- `listsLoadedMsg` arrives and tries to create list model with `width=0`
- List with zero dimensions doesn't display properly
- Later, `WindowSizeMsg` arrives with actual dimensions but items already hidden

### 2. Missing Initialization
- No code to create the favorites directory on first run
- No code to create default `My-favorites.json` file
- App would fail if directory didn't exist

## Solutions Implemented

### 1. Deferred List Model Initialization
**Changed:** Don't create list model until we have proper dimensions

```go
case listsLoadedMsg:
    m.lists = msg.lists
    m.listItems = make([]list.Item, len(msg.lists))
    // ... populate items ...
    
    // Only initialize if we have dimensions
    if m.width > 0 && m.height > 0 {
        m.initializeListModel()
    }
```

**Added:** Check in view render and initialize on-demand
```go
func (m PlayModel) viewListSelection() string {
    // If we have lists but no model, initialize now
    if len(m.lists) > 0 && m.listModel.Items() == nil {
        if m.width > 0 && m.height > 0 {
            m2 := m
            m2.initializeListModel()
            return m2.viewListSelection()
        }
        return "Loading..."
    }
    // ... rest of view ...
}
```

### 2. Auto-Create Favorites Directory and Default List
**Added to `NewPlayModel()`:**
```go
// Ensure favorites directory exists
if err := os.MkdirAll(favoritePath, 0755); err != nil {
    fmt.Fprintf(os.Stderr, "Warning: failed to create favorites directory: %v\n", err)
}

// Ensure My-favorites.json exists
store := storage.NewStorage(favoritePath)
if _, err := store.LoadList(context.Background(), "My-favorites"); err != nil {
    // Create empty My-favorites list
    emptyList := &storage.FavoritesList{
        Name:     "My-favorites",
        Stations: []api.Station{},
    }
    if err := store.SaveList(context.Background(), emptyList); err != nil {
        fmt.Fprintf(os.Stderr, "Warning: failed to create My-favorites: %v\n", err)
    }
}
```

### 3. Extract List Model Creation
**Added helper methods:**
- `initializeListModel()` - Creates list model with proper dimensions
- `initializeStationListModel()` - Creates station list model with proper dimensions

This eliminates code duplication and ensures consistent initialization.

## Result
✅ Lists now display properly when "Play from Favorites" is selected
✅ Favorites directory is automatically created on first run
✅ `My-favorites.json` is automatically created if missing
✅ List model initializes with correct dimensions
✅ Display works correctly regardless of message timing

## Files Modified
- `internal/ui/play.go`
  - Added directory/file initialization in `NewPlayModel()`
  - Deferred list model creation until dimensions available
  - Added `initializeListModel()` and `initializeStationListModel()` helpers
  - Added dimension check in view functions
