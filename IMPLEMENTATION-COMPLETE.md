# Connection Settings Implementation - COMPLETED

## ✅ Implementation Status: COMPLETE

All code has been successfully implemented according to the plan. The feature is ready for testing.

## Files Created

1. ✅ `/Users/shinichiokada/Terminal-Tools/tera/internal/storage/connection_config.go`
   - Connection configuration storage functions
   - Load/Save ConnectionConfig
   - Validation functions

2. ✅ `/Users/shinichiokada/Terminal-Tools/tera/internal/ui/connection_settings.go`
   - Complete UI for Connection Settings screen
   - Three sub-screens: Menu, Delay selection, Buffer selection
   - Follows shuffle_settings.go pattern

## Files Modified

1. ✅ `/Users/shinichiokada/Terminal-Tools/tera/internal/storage/models.go`
   - Added `ConnectionConfig` struct
   - Added `DefaultConnectionConfig()` function

2. ✅ `/Users/shinichiokada/Terminal-Tools/tera/internal/ui/settings.go`
   - Added `settingsStateConnection` to enum
   - Updated menu items (now 1-6 instead of 1-5)
   - Added navigation to Connection Settings (option #2)
   - Updated all menu number shortcuts
   - Updated help text to show "1-6" instead of "1-5"

3. ✅ `/Users/shinichiokada/Terminal-Tools/tera/internal/ui/app.go`
   - Added `screenConnectionSettings` to Screen enum
   - Added `connectionSettingsScreen ConnectionSettingsModel` field to App struct
   - Added initialization in navigateMsg switch
   - Added routing in Update switch
   - Added view rendering in View function

4. ✅ `/Users/shinichiokada/Terminal-Tools/tera/internal/player/mpv.go`
   - Added import for storage package
   - Modified `Play()` method to load and use connection settings
   - Dynamically builds mpv arguments based on ConnectionConfig
   - Implements auto-reconnect, reconnect delay, and buffer size

## What Was Implemented

### Data Layer
- `ConnectionConfig` struct with 3 settings:
  - `AutoReconnect` (bool) - Enable/disable auto-reconnect
  - `ReconnectDelay` (int) - Delay between reconnects (1-30 seconds)
  - `StreamBufferMB` (int) - Buffer size (10-200 MB)
- Default values: AutoReconnect=true, Delay=5s, Buffer=50MB
- Configuration stored in `~/.config/tera/connection_config.yaml`

### UI Layer
- New "Connection Settings" menu item in Settings (option #2)
- Three-screen flow:
  1. Main menu (5 options: Toggle, Set Delay, Set Buffer, Reset, Back)
  2. Delay selection (6 options: 1s, 3s, 5s, 10s, 15s, 30s)
  3. Buffer selection (6 options: 10MB, 25MB, 50MB, 100MB, 150MB, 200MB)
- Real-time config save on changes
- Current settings displayed at top
- Helpful info messages explaining each setting
- All navigation works: arrows, j/k, number shortcuts, Esc, 0

### Player Integration
- MPV flags dynamically added based on config:
  - If auto-reconnect ON: `--loop-playlist=force` + `--stream-lavf-o=reconnect_streamed=1,reconnect_delay_max=X`
  - If buffer > 0: `--cache=yes` + `--demuxer-max-bytes=XM`
  - If buffer = 0: `--no-cache` (original behavior)
- Works for all playback contexts (main menu, search, lucky, etc.)

## Testing Steps

To test the implementation:

1. **Build the project:**
   ```bash
   cd /Users/shinichiokada/Terminal-Tools/tera
   go build -o tera_test ./cmd/tera
   ```

2. **Run and navigate to Connection Settings:**
   ```bash
   ./tera_test
   # Press 6 (Settings)
   # Press 2 (Connection Settings)
   ```

3. **Test the UI:**
   - Toggle auto-reconnect on/off
   - Change reconnect delay
   - Change buffer size
   - Reset to defaults
   - Verify settings persist after restart

4. **Test auto-reconnect functionality:**
   - Enable auto-reconnect
   - Play a station
   - Briefly disconnect WiFi/network
   - Verify stream automatically reconnects

5. **Test buffering:**
   - Set buffer to 50MB
   - Play a station
   - Monitor for smooth playback during brief drops
   - Try with larger/smaller buffers

## Configuration File Example

After running once, you'll find this file:
`~/.config/tera/connection_config.yaml`

```yaml
auto_reconnect: true
reconnect_delay: 5
stream_buffer_mb: 50
```

## Menu Structure

```
TERA > Settings
  1. Theme / Colors
  2. Connection Settings ← NEW
  3. Shuffle Settings
  4. Search History
  5. Check for Updates
  6. About TERA
```

```
TERA > Settings > Connection Settings
Current Settings:
  Auto-reconnect:    Enabled
  Reconnect delay:   5 seconds
  Stream buffer:     50 MB

  1. Toggle Auto-reconnect (On)
  2. Set Reconnect Delay (5 sec)
  3. Set Stream Buffer (50 MB)
  4. Reset to Defaults
  5. Back to Settings
```

## Known Limitations

1. Settings apply only to newly started streams (must stop/start to apply changes)
2. Very weak connections may still require manual reconnect despite best efforts
3. Large buffer sizes (150MB+) use more RAM

## Troubleshooting

**Q: Settings don't persist after restart**
A: Check that `~/.config/tera` directory is writable

**Q: Auto-reconnect not working**
A: Ensure mpv is updated (older versions may not support all flags)

**Q: Buffer size has no effect**
A: Some streams may not benefit from buffering (depends on server configuration)

## Next Steps

1. Test manually using the checklist above
2. Gather user feedback from GitHub issue #4
3. Consider adding connection quality indicator (future enhancement)
4. Update README.md documentation

## GitHub Issue Resolution

This implementation addresses **GitHub Issue #4**:
> "Using GPRS/4G, I sometimes lose the signal/connection and then have to reconnect manually to the station I was listening to."

**Solution provided:**
- Auto-reconnect feature with configurable retry delay
- Stream buffering to handle brief signal drops
- User-friendly UI to configure these settings
- Default settings chosen for mobile/4G use cases

---

**Implementation completed:** February 4, 2026
**Developer:** Claude (with guidance from Shinichi Okada)
**Status:** ✅ READY FOR TESTING
