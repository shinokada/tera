# TERA Volume Control & Footer Navigation Implementation

## Summary of Changes

### 1. Core Volume Support ✅

**File: `internal/api/models.go`**
- Added `Volume int` field to `Station` struct
- Volume is stored per-station (0-100, where 0 means use default)

**File: `internal/player/mpv.go`**
- Added volume field to `MPVPlayer`
- Implemented `GetVolume()`, `SetVolume()`, `IncreaseVolume()`, `DecreaseVolume()` methods
- Modified `Play()` to use station-specific volume or player default
- Volume is passed to mpv via `--volume` flag

### 2. Help System ✅

**File: `internal/ui/components/help.go`** (NEW)
- Created reusable help overlay component
- `HelpModel` with sections and items
- Built-in help content creators:
  - `CreateMainMenuHelp()` - for main menu
  - `CreatePlayingHelp()` - for playing screens
- Centered overlay with nice borders
- Press any key to close

### 3. Message Types ✅

**File: `internal/ui/messages.go`**
- Added `volumeChangedMsg` for volume updates
- Added `showHelpMsg` and `hideHelpMsg` for help overlay

### 4. Main Menu Updates ✅

**File: `internal/ui/app.go`**
- Added help model to `App` struct
- Added volume display (temporary, 2-second timeout)
- Implemented `?` key for help
- Implemented `/` and `*` keys for volume control (when playing)
- Context-sensitive footer:
  - **Not playing:** `↑↓/jk: Navigate • Enter: Select • 1-6: Menu • 10+: Quick Play • ?: Help`
  - **Playing:** `↑↓/jk: Navigate • Enter: Select • /*: Volume • m: Mute • Esc: Stop • ?: Help`
- Added `saveStationVolume()` to persist volume changes to favorites
- Volume display shows temporarily when adjusted
- Help overlay appears on `?` key

## What Still Needs to be Done

### 1. Play Screen (Favorites) - `internal/ui/play.go`

**Required Changes:**
- [ ] Add help model to `PlayModel` struct
- [ ] Add volume display fields
- [ ] Handle `?`, `/`, `*` keys in Update()
- [ ] Update footer based on playing state:
  - **List selection:** `↑↓/jk: Navigate • Enter: Select • Esc: Back • ?: Help`
  - **Playing:** `f: Favorites • v: Vote • 0: Main Menu • ?: Help`
- [ ] Save volume when adjusting during playback
- [ ] Remove the "Save Station?" confirmation page (already planned in footer-navigation.md)

### 2. Search Screen - `internal/ui/search.go`

**Required Changes:**
- [ ] Add help model to `SearchModel` struct
- [ ] Add volume display fields
- [ ] Handle `?`, `/`, `*` keys
- [ ] Update footer:
  - **Playing:** `f: Save to Favorites • s: Save to list • v: Vote • ?: Help`
- [ ] Save volume when adjusting

### 3. I Feel Lucky Screen - `internal/ui/lucky.go`

**Required Changes:**
- [ ] Add help model to `LuckyModel` struct
- [ ] Add volume display fields
- [ ] Handle `?`, `/`, `*` keys
- [ ] Update footer:
  - **Playing:** `f: Save to Favorites • s: Save to list • v: Vote • ?: Help`
- [ ] Save volume when adjusting
- [ ] Make `Esc` stop playback and return to main menu (no save prompt)

### 4. Mute Functionality (All Screens)

**Note:** mpv's `m` key for mute doesn't work in `--no-terminal` mode. We have two options:

**Option A: Restart playback with volume 0**
- When user presses `m`, restart the stream with `--volume=0`
- Store previous volume to restore when unmuting
- Display "Muted" message

**Option B: Remove mute feature**
- Since we can't control mpv at runtime, just remove the `m: Mute` from help
- Only provide volume up/down controls

**Recommendation:** Go with Option B for simplicity, or implement Option A if mute is important.

## Implementation Pattern

For each screen that needs updates, follow this pattern:

```go
type ScreenModel struct {
    // ... existing fields ...
    helpModel           components.HelpModel
    volumeDisplay       string
    volumeDisplayFrames int
}

// In Init()
func NewScreenModel(...) ScreenModel {
    return ScreenModel{
        // ... existing init ...
        helpModel: components.NewHelpModel(components.CreatePlayingHelp()),
    }
}

// In Update()
case tea.KeyMsg:
    // If help is visible, let it handle the key
    if m.helpModel.IsVisible() {
        var cmd tea.Cmd
        m.helpModel, cmd = m.helpModel.Update(msg)
        return m, cmd
    }

    switch msg.String() {
    case "?":
        m.helpModel.Show()
        return m, nil
    case "/":
        if m.player != nil && m.player.IsPlaying() {
            newVol := m.player.DecreaseVolume(5)
            m.volumeDisplay = fmt.Sprintf("Volume: %d%%", newVol)
            m.volumeDisplayFrames = 2
            // Save station volume
            if m.currentStation != nil {
                m.currentStation.Volume = newVol
                m.saveStationVolume(m.currentStation)
            }
            return m, tickEverySecond()
        }
    case "*":
        // Similar to "/" but IncreaseVolume
    }

case tickMsg:
    if m.volumeDisplayFrames > 0 {
        m.volumeDisplayFrames--
        if m.volumeDisplayFrames == 0 {
            m.volumeDisplay = ""
        }
        return m, tickEverySecond()
    }

// In View()
// Add volume display
if m.volumeDisplay != "" {
    content.WriteString(highlightStyle().Render(m.volumeDisplay))
    content.WriteString("\n")
}

// Overlay help if visible
if m.helpModel.IsVisible() {
    return m.helpModel.View()
}
```

## Testing Checklist

Once all screens are updated:

- [ ] Test volume controls on main menu quick play
- [ ] Test volume persistence (adjust volume, restart app, play same station)
- [ ] Test help overlay on all screens
- [ ] Verify context-sensitive footers appear correctly
- [ ] Test that `Esc` stops playback and returns to appropriate screen
- [ ] Verify no "Save Station?" prompts appear
- [ ] Test volume display shows for 2 seconds and disappears
- [ ] Verify per-station volume is saved correctly

## Next Steps

Would you like me to:
1. Implement these changes for the Play screen first?
2. Implement for all screens at once?
3. Focus on a specific screen?
4. Add the mute functionality (with stream restart)?

