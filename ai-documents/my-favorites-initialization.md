# My-Favorites Initialization Fix

## Summary
Ensured that the `My-favorites.json` list is created at application startup, not just when entering the Play screen.

## Changes Made

### 1. `/Users/shinichiokada/Terminal-Tools/tera/internal/ui/app.go`

#### Added imports:
- `context` - for context.Background()
- `github.com/shinokada/tera/internal/storage` - for storage operations

#### Updated `NewApp()`:
- Added directory creation for favorites path at startup
- Added call to `ensureMyFavorites()` after menu initialization

#### Added new method `ensureMyFavorites()`:
```go
// ensureMyFavorites ensures My-favorites.json exists at startup
func (a *App) ensureMyFavorites() {
	store := storage.NewStorage(a.favoritePath)
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
}
```

## Behavior

### Before:
- Favorites directory was only created when entering Play screen
- My-favorites.json was only created when entering Play screen
- Users might see errors if they tried to access favorites-related features before visiting the Play screen

### After:
- Favorites directory is created at application startup
- My-favorites.json is created at application startup (if it doesn't exist)
- All favorites-related features can work immediately without needing to visit the Play screen first
- Ensures a consistent state for new installations

## Benefits

1. **Better first-run experience**: New users have the required directory structure immediately
2. **Prevents errors**: No need to wait until Play screen visit to initialize
3. **Consistent state**: App has a known good state from startup
4. **Fail-fast**: Any issues with directory/file creation are detected at startup, not later

## Testing

To test:
1. Delete the favorites directory: `rm -rf ~/.config/tera/favorites`
2. Start the app
3. Verify that `~/.config/tera/favorites/My-favorites.json` is created
4. The file should contain an empty stations array: `{"name":"My-favorites","stations":[]}`
