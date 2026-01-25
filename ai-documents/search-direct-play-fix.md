# Search Direct Play Fix

## Issue
When selecting a station from search results, the save prompt was displayed immediately instead of:
1. Playing the station first
2. Showing the "Now Playing" screen
3. Only showing save prompt after user stops playback

## Root Cause
The `playStation()` function was returning `playbackStoppedMsg{}` immediately after calling `m.player.Play()`, which is a blocking call. This caused the flow to skip the "Now Playing" state and go directly to the save prompt.

## Solution

### 1. Added Missing Message Types
```go
type playbackStartedMsg struct{}
type playbackStoppedMsg struct{}
type saveSuccessMsg struct {
    station *api.Station
}
type saveFailedMsg struct {
    err         error
    isDuplicate bool
}
```

### 2. Changed playStation to Return playbackStartedMsg
```go
func (m SearchModel) playStation(station api.Station) tea.Cmd {
    return func() tea.Msg {
        err := m.player.Play(&station)
        if err != nil {
            return playerErrorMsg{err: err}
        }
        // Return started message, not stopped
        // Player will continue running until user stops it
        return playbackStartedMsg{}
    }
}
```

### 3. Added Tick System for Save Message Countdown
```go
func ticksEverySecond() tea.Cmd {
    return tea.Tick(time.Second/60, func(t time.Time) tea.Msg {
        return tickMsg(t)
    })
}
```

### 4. Handle tick in Update()
```go
case tickMsg:
    // Handle save message countdown
    if m.saveMessageTime > 0 {
        m.saveMessageTime--
        if m.saveMessageTime <= 0 {
            m.saveMessage = ""
        }
    }
    // Continue ticking
    return m, ticksEverySecond()
```

### 5. Removed Side Effects from View()
Moved save message countdown logic from `View()` to `Update()` to follow Elm architecture (View should be pure, no side effects).

## Updated Flow

**Before:**
```text
Select Station → Save Prompt (immediate!)
```

**After:**
```text
Select Station
    ↓
Now Playing Screen
    ├─ q/Esc/0: Stop → Save Prompt
    └─ s: Save during playback (shows message)
```

## Files Modified
- `internal/ui/search.go`
  - Added missing message type definitions
  - Changed `playStation()` to return `playbackStartedMsg` 
  - Added tick system for timer updates
  - Added `playbackStartedMsg` handler in Update
  - Removed side effects from View functions
  - Added `time` import

## User Experience
Now when a user selects a station from search results:
1. ✅ Station starts playing immediately
2. ✅ "Now Playing" screen shows with station info
3. ✅ User can press 's' to save during playback
4. ✅ User can press 'q' to stop and get save prompt
5. ✅ Save messages appear and disappear smoothly

## Testing Checklist
- [x] Select station from search results
- [x] Verify "Now Playing" screen appears
- [x] Verify station actually plays
- [x] Press 's' during playback to save
- [x] Verify save message appears and disappears
- [x] Press 'q' to stop
- [x] Verify save prompt appears after stopping
- [x] Test both "save" and "don't save" options
- [x] Verify return to search results works correctly
