# Critical Bug Fixes for TERA Search

## Issues Found

1. **Multiple stations playing simultaneously** - Player not stopped when starting new station
2. **Station info shows wrong content** - Not showing selected station properly
3. **Panic on player.Wait()** - Race condition when stopping player
4. **Player goroutines not cleaned up** - Zombie processes

## Fixes Applied

### 1. Player Management (`internal/player/mpv.go`)

**Problem:** Race condition in monitor() goroutine - Wait() called on nil cmd
**Fix:** Add nil checks and proper synchronization

### 2. Search Screen (`internal/ui/search.go`)

**Problem:** Player not stopped before playing new station
**Fix:** 
- Stop player when entering station info
- Stop player when going back to results
- Stop player when exiting to main menu

### 3. State Cleanup

**Problem:** Old state not cleared when changing screens
**Fix:** Clear selectedStation and stop player on state transitions

## Implementation

The fixes ensure:
1. Only one station plays at a time
2. Player is always stopped before starting new one
3. Clean state transitions
4. No zombie processes
5. Proper error handling
