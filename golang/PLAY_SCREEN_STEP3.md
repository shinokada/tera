# Play Screen Implementation - Step 3: Playback

## What We Built

This is the final step in implementing the Play Screen. We've added MPV player integration and playback functionality.

## Files Created/Modified

### New Files
1. **`internal/player/mpv.go`** - MPV player controller
   - `MPVPlayer` struct with thread-safe operations
   - `Play()` - Start playing a station
   - `Stop()` - Stop playback
   - `IsPlaying()` - Check playback status
   - `GetCurrentStation()` - Get currently playing station
   - Process monitoring and cleanup

2. **`internal/player/mpv_test.go`** - Player tests
   - Tests for player lifecycle
   - Thread safety tests
   - State management tests

### Modified Files
3. **`internal/ui/play.go`** - Added playback logic
   - Integrated MPV player
   - `startPlayback()` - Initiates playback
   - `stopPlayback()` - Stops playback
   - `updatePlaying()` - Handles playback input
   - `viewPlaying()` - Renders playback view
   - `formatStationInfo()` - Formats station details
   - New message types for playback events

## Features Implemented

✅ **MPV Integration**
- Thread-safe player controller
- Automatic process management
- Graceful start/stop
- Error handling

✅ **Playback Controls**
- Press `Enter` on station to start playing
- Press `q`, `Esc`, or `0` to stop
- Real-time playback status display
- Automatic cleanup on errors

✅ **Station Info Display**
- Formatted station information box
- Shows: name, country, codec, bitrate, tags, votes
- Visual playback indicator (▶ Playing / ⏸ Stopped)

✅ **Error Handling**
- MPV not found error
- Stream connection errors
- Graceful fallback to station list

## Testing

Run the tests:
```bash
# Test MPV player
go test ./internal/player -v

# Test Play screen
go test ./internal/ui -v -run Play

# Build and run
go build -o tera cmd/tera/main.go
./tera
```

## How to Try It

### Prerequisites
Make sure MPV is installed:
```bash
# macOS
brew install mpv

# Linux
sudo apt install mpv  # Debian/Ubuntu
sudo dnf install mpv  # Fedora
```

### Test It Out
1. **Create test data** (if you haven't):
```bash
mkdir -p ~/.config/tera/favorites
cat > ~/.config/tera/favorites/Radio.json << 'EOF'
[
  {
    "stationuuid": "1",
    "name": "KEXP 90.3 FM",
    "url_resolved": "https://kexp-mp3-128.streamguys1.com/kexp128.mp3",
    "country": "USA",
    "codec": "MP3",
    "bitrate": 128,
    "tags": "indie, alternative",
    "votes": 5000
  }
]
EOF
```

2. **Run the app:**
```bash
./tera
```

3. **Try playback:**
   - Press `1` to enter Play screen
   - Select "Radio" list
   - Select "KEXP 90.3 FM" with Enter
   - **Listen!** 🎵
   - Press `q` to stop

## User Flow

```
Main Menu 
  → Play Screen (1)
    → Select List
      → Select Station
        → NOW PLAYING! ▶
          → Press q to stop
          → Press s to save (TODO Step 3.2)
```

## Current Playback View

```
╭───────────────────────────────────╮
│ Now Playing                        │
╰───────────────────────────────────╯

╭───────────────────────────────────╮
│ KEXP 90.3 FM                       │
│                                    │
│ Country: USA                       │
│ Codec: MP3 (128 kbps)             │
│ Tags: indie, alternative           │
│ Votes: 5000                        │
╰───────────────────────────────────╯

▶ Playing...

q/esc/0: stop • s: save to favorites
```

## Architecture

### MPV Player
```go
type MPVPlayer struct {
    cmd     *exec.Cmd      // MPV process
    playing bool           // State
    station *api.Station   // Current station
    mu      sync.Mutex     // Thread safety
    stopCh  chan struct{}  // Stop signal
}
```

### Playback Flow
```
Select Station 
  → startPlayback()
    → player.Play(station)
      → exec.Command("mpv", ...)
      → monitor() goroutine
        → playbackStartedMsg
          → View updates
            → User presses q
              → stopPlayback()
                → player.Stop()
                  → Kill process
                    → Back to station list
```

### Thread Safety
- All player operations are mutex-protected
- Safe concurrent access from UI and monitor goroutine
- Proper cleanup on stop

## Next: Step 3.2 - Save to Quick Favorites

To complete the Play Screen, we still need:
- [ ] Implement 's' key to save station
- [ ] Check if station already in Quick Favorites
- [ ] Add to My-favorites.json
- [ ] Duplicate checking by StationUUID
- [ ] Show save confirmation

This will match the spec from `flow-charts.md` where:
- Stations from Quick Favorites → No save option
- Stations from other lists → Can save during playback

## Questions?

Refer to:
- `golang/spec-docs/flow-charts.md` - Play Screen playback flow
- `golang/PLAY_PROGRESS.md` - Development roadmap
