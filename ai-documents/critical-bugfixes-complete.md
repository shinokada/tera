# Critical Bug Fixes - Complete Summary

## Emergency: Multiple Stations Playing

### Immediate Action Taken
```bash
killall mpv  # or pkill -9 mpv
```

## Root Causes Identified

### 1. Player Not Stopped Before New Station
**Problem:** When playing a new station, the old player process wasn't stopped, causing multiple mpv instances to run simultaneously.

**Locations:**
- `executeStationAction()` - Starting playback without stopping previous
- State transitions - Not cleaning up player

### 2. Race Condition in Player Monitor
**Problem:** The `monitor()` goroutine could call `Wait()` on a nil or already-stopped process, causing panic.

**Error:**
```text
panic: runtime error: invalid memory address or nil pointer dereference
[signal SIGSEGV: segmentation violation code=0x2 addr=0xa0 pc=0x100be3cd8]
```

### 3. State Not Cleared on Navigation
**Problem:** `selectedStation` wasn't cleared when navigating away, causing confusion about what's displayed.

## Fixes Applied

### 1. Player Management (`internal/player/mpv.go`)

**Before:**
```go
func (p *MPVPlayer) monitor() {
    if p.cmd == nil {
        return
    }
    go func() {
        done <- p.cmd.Wait()  // CRASH: p.cmd could become nil
    }()
}
```

**After:**
```go
func (p *MPVPlayer) monitor() {
    p.mu.Lock()
    cmd := p.cmd  // Capture cmd while locked
    p.mu.Unlock()
    
    if cmd == nil || cmd.Process == nil {
        return
    }
    
    go func() {
        if cmd != nil && cmd.Process != nil {
            done <- cmd.Wait()  // SAFE: using captured cmd
        } else {
            done <- nil
        }
    }()
}
```

**Benefits:**
- No more nil pointer dereference
- Safe concurrent access
- Proper process cleanup

### 2. Search Screen State Management

#### executeStationAction()
**Added:**
```go
case 0: // Play station
    // Stop any currently playing station first
    if m.player != nil && m.player.IsPlaying() {
        m.player.Stop()
    }
    m.state = searchStatePlaying
    return m, m.playStation(*m.selectedStation)

case 2: // Back to results
    // Stop player when going back
    if m.player != nil && m.player.IsPlaying() {
        m.player.Stop()
    }
    m.selectedStation = nil  // Clear selection
    m.state = searchStateResults
    return m, nil
```

#### handleMenuInput()
**Added:**
```go
if msg.String() == "0" || msg.String() == "esc" {
    // Stop any playing station when exiting
    if m.player != nil && m.player.IsPlaying() {
        m.player.Stop()
    }
    m.selectedStation = nil
    return m, func() tea.Msg { return backToMainMsg{} }
}
```

#### handleResultsInput()
**Added:**
```go
case "esc":
    // Stop any playing station when going back
    if m.player != nil && m.player.IsPlaying() {
        m.player.Stop()
    }
    m.selectedStation = nil
    m.state = searchStateMenu
    return m, nil
```

#### handleStationInfoInput()
**Added:**
```go
if msg.String() == "0" {
    // Stop any playing station
    if m.player != nil && m.player.IsPlaying() {
        m.player.Stop()
    }
    m.selectedStation = nil
    return m, func() tea.Msg { return backToMainMsg{} }
}

if msg.String() == "esc" || msg.String() == "3" {
    // Stop player when going back
    if m.player != nil && m.player.IsPlaying() {
        m.player.Stop()
    }
    m.selectedStation = nil
    m.state = searchStateResults
    return m, nil
}
```

## Changes Summary

### Files Modified
1. `internal/player/mpv.go` - Fixed race condition
2. `internal/ui/search.go` - Added player cleanup on all state transitions

### Lines Changed
- **mpv.go**: 10 lines (monitor function)
- **search.go**: 40+ lines (state cleanup)

## Testing Required

### 1. Stop All Running Stations
```bash
# First, kill any running mpv instances
killall mpv
# or
pkill -9 mpv

# Check no mpv is running
ps aux | grep mpv
```

### 2. Build and Test
```bash
# Rebuild
go build -o tera ./cmd/tera/

# Run
./tera
```

### 3. Test Scenarios

#### Test 1: Single Station Play
1. Search for a station
2. Select and play
3. **Verify:** Only one station plays
4. Stop with q
5. **Verify:** Station stops

#### Test 2: Switch Stations
1. Search and play station A
2. Go back (Esc)
3. Select station B
4. Play station B
5. **Verify:** Station A stopped automatically
6. **Verify:** Only station B playing

#### Test 3: Multiple Switches
1. Play station A
2. Go back
3. Play station B
4. Go back
5. Play station C
6. **Verify:** Only C is playing
7. **Check processes:** `ps aux | grep mpv` shows only 1 mpv

#### Test 4: Navigation While Playing
1. Start playing a station
2. Press Esc (go back to results)
3. **Verify:** Playback stops
4. **Verify:** No zombie processes

#### Test 5: Exit While Playing
1. Start playing a station
2. Press 0 (return to main menu)
3. **Verify:** Playback stops
4. **Check processes:** No mpv running

#### Test 6: Station Info Display
1. Search for stations
2. Select a station
3. **Verify:** Shows ONLY that station's info
4. **Verify:** Info matches selected station
5. Navigate menu with arrows
6. **Verify:** Menu works correctly

### 4. Check for Zombie Processes
```bash
# While testing, run in another terminal:
watch 'ps aux | grep mpv | grep -v grep'

# Should show:
# - 0 processes when not playing
# - 1 process when playing
# - NEVER multiple processes
```

### 5. Check for Panic
```bash
# Test should complete without panics
# No "SIGSEGV" errors
# No "nil pointer dereference" errors
```

## Expected Behavior After Fix

### ✅ Correct Behavior
- Only one station plays at any time
- Previous station stops automatically when playing new one
- Station stops when navigating away
- Station stops when exiting
- No panic errors
- No zombie processes
- Station info shows correct station

### ❌ Previous Behavior
- Multiple stations played simultaneously
- Stations kept playing after navigation
- Panic on player stop
- Zombie mpv processes
- Confusing station info display

## Verification Checklist

- [ ] No mpv processes running before test
- [ ] Build succeeds without errors
- [ ] Single station plays correctly
- [ ] Switching stations stops previous
- [ ] Navigation stops playback
- [ ] No panic errors during test
- [ ] `ps aux | grep mpv` shows max 1 process
- [ ] Station info displays correctly
- [ ] Arrow keys work in station menu
- [ ] All state transitions clean up properly

## If Problems Persist

1. **Check Process Count:**
   ```bash
   ps aux | grep mpv | wc -l
   ```
   Should be 0 or 1, never more.

2. **Force Kill All:**
   ```bash
   killall -9 mpv
   pkill -9 mpv
   ```

3. **Rebuild Clean:**
   ```bash
   go clean
   go build -o tera ./cmd/tera/
   ```

4. **Check mpv Version:**
   ```bash
   mpv --version
   ```

5. **Test mpv Directly:**
   ```bash
   mpv --no-video --no-terminal http://example.com/stream.mp3
   ```

## Prevention

These fixes ensure:
1. **One Source of Truth:** Only one player instance
2. **Clean Transitions:** Always stop before starting
3. **Safe Concurrency:** Proper locking and nil checks
4. **Clear State:** selectedStation cleared on navigation
5. **No Zombies:** Proper process cleanup

## Documentation

The fixes maintain:
- ✅ All navigation features
- ✅ Arrow key support
- ✅ Number shortcuts
- ✅ Visual feedback
- ✅ Error handling

And add:
- ✅ Proper player lifecycle management
- ✅ Safe concurrent access
- ✅ Clean state transitions
- ✅ No zombie processes

## Next Steps

1. Test thoroughly with above scenarios
2. Monitor for any remaining issues
3. Add unit tests for player lifecycle
4. Document player state management
