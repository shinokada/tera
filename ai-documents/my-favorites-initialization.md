# My-Favorites Initialization Fix - Updated

## Summary
Ensured that the `My-favorites.json` list is created at application startup with proper error handling to prevent data loss.

## Changes Made

### 1. `/Users/shinichiokada/Terminal-Tools/tera/internal/ui/app.go`

#### Added imports:
- `context` - for context.Background()
- `github.com/shinokada/tera/internal/storage` - for storage operations

#### Updated `NewApp()`:
- Added directory creation for favorites path at startup
- Added call to `ensureMyFavorites()` after menu initialization

#### Added new method `ensureMyFavorites()` with proper error handling:
```go
// ensureMyFavorites ensures My-favorites.json exists at startup
func (a *App) ensureMyFavorites() {
	store := storage.NewStorage(a.favoritePath)
	if _, err := store.LoadList(context.Background(), "My-favorites"); err != nil {
		// Only create if file doesn't exist, not on other errors
		if !os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "Warning: failed to load My-favorites: %v\n", err)
			return
		}
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

### 2. `/Users/shinichiokada/Terminal-Tools/tera/internal/ui/play.go`

#### Removed redundant initialization code from `NewPlayModel()`:
- Removed directory creation (now handled in app.go)
- Removed My-favorites.json creation (now handled in app.go)
- Added comment explaining that initialization happens at app startup
- Simplified the function to just return the PlayModel struct

## Key Improvement: Proper Error Handling

The critical improvement suggested by CodeRabbit was to use `os.IsNotExist(err)` to distinguish between different error types:

### Before (UNSAFE):
```go
if _, err := store.LoadList(...); err != nil {
    // Creates empty list for ANY error - could overwrite corrupted data!
    createEmptyList()
}
```

### After (SAFE):
```go
if _, err := store.LoadList(...); err != nil {
    if !os.IsNotExist(err) {
        // Real error (permissions, corruption, etc.) - don't overwrite!
        logError(err)
        return
    }
    // File doesn't exist - safe to create
    createEmptyList()
}
```

## Error Scenarios Now Handled Correctly

### ✅ File doesn't exist (first run):
- Creates empty My-favorites.json
- User can start adding favorites

### ✅ Permission denied:
- Logs warning
- Doesn't overwrite anything
- User can fix permissions

### ✅ Corrupted JSON:
- Logs warning
- Doesn't overwrite (user data preserved)
- User can manually fix or restore from backup

### ✅ Disk read error:
- Logs warning
- Doesn't hide the problem
- User knows something is wrong

## Benefits

1. **Prevents data loss**: Won't overwrite corrupted files or files with permission issues
2. **Better diagnostics**: Users see real error messages instead of silent failures
3. **Single point of initialization**: All initialization happens at app startup in app.go
4. **No redundancy**: Removed duplicate initialization code from play.go
5. **Fail-fast**: Issues are detected and reported immediately at startup

## Testing

To test various scenarios:

1. **First run (normal case)**:
   ```bash
   rm -rf ~/.config/tera/favorites
   ./tera
   # Should create My-favorites.json successfully
   ```

2. **Permission denied**:
   ```bash
   mkdir -p ~/.config/tera/favorites
   chmod 000 ~/.config/tera/favorites
   ./tera
   # Should show warning, not create file
   chmod 755 ~/.config/tera/favorites  # restore
   ```

3. **Corrupted file**:
   ```bash
   echo "invalid json" > ~/.config/tera/favorites/My-favorites.json
   ./tera
   # Should show warning, not overwrite
   ```

4. **Normal operation**:
   ```bash
   # With existing My-favorites.json
   ./tera
   # Should load existing file, not recreate
   ```

## Credit
This improvement was suggested by CodeRabbit AI code review, correctly identifying that the original implementation could hide real errors and potentially cause data loss.
