# Shuffle Mode Testing Checklist

## Quick Start Test

```bash
cd /Users/shinichiokada/Terminal-Tools/tera

# Run tests
go test ./internal/ui -v -run "TestLuckyShuffle"

# Build and run
go build -o tera cmd/tera/main.go
./tera
```

## Manual Testing Checklist

### 1. Toggle Shuffle Mode
- [ ] Go to "I Feel Lucky" (option 4)
- [ ] Press `t` - verify shuffle indicator shows `[✓] On`
- [ ] Press `t` again - verify shuffle indicator shows `[ ] Off`
- [ ] With shuffle on, verify auto-advance info appears (if enabled)

### 2. Start Shuffle Search
- [ ] Enable shuffle mode with `t`
- [ ] Enter keyword: `jazz`
- [ ] Press Enter
- [ ] Verify search completes
- [ ] Verify first station starts playing
- [ ] Verify shuffle UI displays:
  - 🔀 Shuffle Active indicator
  - Timer countdown (if auto-advance enabled)
  - Station counter (e.g., "Station 1 of session")
  - Shuffle History section (empty initially)

### 3. Navigation Controls
- [ ] Press `n` - verify next station plays
- [ ] Verify history updates with previous station
- [ ] Press `b` - verify returns to previous station
- [ ] Press `n` multiple times - verify no duplicate stations
- [ ] Verify history shows last 3 stations (or configured max)

### 4. Timer Controls (if auto-advance enabled)
- [ ] Verify countdown timer displays (e.g., "Next in: 4:52")
- [ ] Press `p` - verify "⏸ Auto-advance paused" message
- [ ] Verify timer stops counting down
- [ ] Press `p` again - verify "▶ Auto-advance resumed" message
- [ ] Verify timer resumes countdown
- [ ] Wait for timer to reach 0:00
- [ ] Verify auto-advance to next station

### 5. Stop Shuffle
- [ ] Press `h` while shuffling
- [ ] Verify message: "Shuffle stopped - continuing with current station"
- [ ] Verify state changes to normal playback
- [ ] Verify station continues playing
- [ ] Verify shuffle controls no longer available

### 6. Exit Shuffle
- [ ] Start shuffle mode again
- [ ] Press `Esc`
- [ ] Verify playback stops
- [ ] Verify returns to I Feel Lucky input screen
- [ ] Verify shuffle mode is disabled
- [ ] Verify search history is reloaded

### 7. Return to Main Menu
- [ ] Start shuffle mode
- [ ] Press `0`
- [ ] Verify playback stops
- [ ] Verify returns to main menu
- [ ] Verify shuffle session ends

### 8. Save Features During Shuffle
- [ ] Start shuffle mode
- [ ] Press `f` while station is playing
- [ ] Verify station saves to My-favorites
- [ ] Verify success message appears
- [ ] Press `s` while station is playing
- [ ] Verify list selection appears
- [ ] Select a list and save
- [ ] Verify success message

### 9. Vote During Shuffle
- [ ] Press `v` while station is playing
- [ ] Verify vote success message
- [ ] Verify shuffle continues normally

### 10. Volume Controls During Shuffle
- [ ] Press `*` - verify volume increases
- [ ] Verify volume message appears
- [ ] Press `/` - verify volume decreases
- [ ] Press `m` - verify mute toggles
- [ ] Verify all controls work during shuffle

### 11. Configure Shuffle Settings
- [ ] Go to Settings (option 6)
- [ ] Select "Shuffle Settings" (option 2)
- [ ] Verify current settings display
- [ ] Toggle auto-advance (option 1)
- [ ] Verify setting changes
- [ ] Change interval (option 2)
- [ ] Select different interval
- [ ] Verify setting saves
- [ ] Change history size (option 4)
- [ ] Select different size
- [ ] Verify setting saves
- [ ] Press Esc to return to settings menu

### 12. Configuration File
- [ ] Verify `~/.config/tera/shuffle.yaml` exists
- [ ] Open file and verify content:
  ```yaml
  shuffle:
    auto_advance: true
    interval_minutes: 5
    remember_history: true
    max_history: 5
  ```
- [ ] Edit file manually
- [ ] Restart TERA
- [ ] Verify settings applied from file

### 13. Edge Cases
- [ ] Search for keyword with 0 results
  - Verify error message
  - Verify returns to input
- [ ] Search for keyword with 1 result
  - Verify single station plays
  - Verify `n` shows error (no more stations)
- [ ] Search for keyword with many results
  - Verify random selection works
  - Verify no duplicates until all played
- [ ] Test with auto-advance disabled
  - Verify no timer shown
  - Verify "Manual mode" indicator
  - Verify `n` and `b` still work

### 14. Help Screen
- [ ] Press `?` during shuffle playback
- [ ] Verify help screen displays
- [ ] Verify all shortcuts listed
- [ ] Press `?` or `Esc` to close

## Expected Output Examples

### Input Screen with Shuffle On
```
🎲 I Feel Lucky

Genre/keyword: jazz

Shuffle mode: [✓] On  (press 't' to disable)
              Auto-advance in 5 min • History: 5 stations

─── Recent Searches ───
  1. rock
  2. classical
```

### Shuffle Playback
```
🎵 Now Playing (🔀 Shuffle: jazz)

Station: Smooth Jazz 24/7
Country: United States
Codec: AAC • Bitrate: 128 kbps

▶ Playing...

🔀 Shuffle Active • Next in: 4:23
   Station 3 of session
   
─── Shuffle History ───
  ← Jazz FM London
  ← WBGO Jazz 88.3
  → Smooth Jazz 24/7  ← Current
```

### Shuffle Playback (Manual Mode)
```
🔀 Shuffle Active • Manual mode
   Station 5 of session
```

### Shuffle Playback (Timer Paused)
```
🔀 Shuffle Active • ⏸ Timer paused
   Station 2 of session
```

## Test Results

Date tested: _______________

✅ All basic functionality works
✅ Navigation controls work correctly  
✅ Timer controls work correctly
✅ Settings save and load properly
✅ Edge cases handled gracefully
✅ UI displays correctly
✅ No crashes or errors

Tested by: _______________

Notes:
_____________________________________
_____________________________________
_____________________________________
