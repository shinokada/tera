# Shuffle Mode

## 🎨 Shuffle Mode UI Design

### Settings Menu - Updated**

```text

                         TERA
  ⚙️  Settings

    > 1. Theme / Colors
      2. Shuffle Settings
      3. Search History
      4. Check for Updates
      5. About TERA

↑↓/jk: Navigate • Enter: Select • 1-5: Shortcut • Esc/0: Back • Ctrl+C: Quit
```

---

### New Shuffle Settings Page**

```

                         TERA

  ⚙️  Settings > Shuffle Settings

Current Settings:

  Auto-advance:       Disabled
  Auto-advance interval: 5 minutes
  Remember history:   Enabled (Last 5 stations)

    > 1. Toggle Auto-advance (Off/On)
      2. Set Auto-advance Interval (1, 3, 5, 10, 15 min)
      3. Toggle History (Off/On)
      4. Set History Size (3, 5, 7, 10)
      5. Reset to Defaults
      6. Back to Settings

✓ Shuffle settings saved automatically

↑↓/jk: Navigate • Enter: Select • 1-6: Shortcut • Esc: Back • 0: Main Menu
```

**Key features:**
- Shows current configuration at the top
- Each setting has a toggle or cycle option
- Settings save automatically (persisted to config file)
- Simple numbered menu like other settings pages

---

### Shuffle Settings - Sub-page for Interval Selection**

When you press `2` (Set Auto-advance Interval):

```text

                         TERA

  ⚙️  Settings > Shuffle Settings > Auto-advance Interval

Select auto-advance interval:

  Current: 5 minutes


    > 1. 1 minute
      2. 3 minutes
      3. 5 minutes  ← Current
      4. 10 minutes
      5. 15 minutes
      6. Back

Auto-advance will skip to next shuffle station after this interval

↑↓/jk: Navigate • Enter: Select • 1-6: Shortcut • Esc: Back • 0: Main Menu
```

---

### Shuffle Settings - Sub-page for History Size**

When you press `4` (Set History Size):

```text

                         TERA

  ⚙️  Settings > Shuffle Settings > History Size

Select shuffle history size:

  Current: 5 stations

    > 1. 3 stations (Minimal)
      2. 5 stations (Default)  ← Current
      3. 7 stations
      4. 10 stations (Maximum)
      5. Back


History allows you to go back to previous shuffle stations

↑↓/jk: Navigate • Enter: Select • 1-5: Shortcut • Esc: Back • 0: Main Menu
```

---

### I Feel Lucky Input - With Shuffle Toggle**

```text
🎲 I Feel Lucky

Choose an option:

Type a genre of music: rock, classical, jazz, pop, country, hip, heavy, blues, soul.
Or type a keyword like: meditation, relax, mozart, Beatles, etc.

Use only one word.

Genre/keyword: [jazz_____________]

Shuffle mode: [ ] Off (press 't' to enable)

─── Recent Searches ───
  1. jazz
  2. rock
  3. meditation


↑↓/jk: Navigate • Enter: Search • t: Toggle shuffle • 1-3: Quick search • Esc: Back
```

**When shuffle is enabled:**
```
Shuffle mode: [✓] On  (press 't' to disable)
              Auto-advance in 5 min • History: 5 stations
```

---

### Now Playing with Shuffle Active**

```
🎵 Now Playing (🔀 Shuffle: jazz)

Station: Smooth Jazz 24/7
Country: United States
Codec: AAC • Bitrate: 128 kbps
Tags: jazz, smooth jazz, relaxing
Homepage: https://smoothjazz247.com

▶ Playing...

🔀 Shuffle Active • Next in: 4:23
   Station 3 of session
   
─── Shuffle History ───
  ← Jazz FM London
  ← WBGO Jazz 88.3
  → Smooth Jazz 24/7  ← Current

f: Fav • s: List • v: Vote • n: Next • b: Prev • p: Pause timer • h: Stop shuffle • ?: Help
```

**Key UI elements:**
- `🔀` icon shows shuffle is active
- Timer countdown (if auto-advance enabled)
- Station counter
- Last 3 stations shown (configurable)
- All controls accessible

---

### Shuffle Mode Without Auto-advance**

```
🎵 Now Playing (🔀 Shuffle: jazz)

Station: Smooth Jazz 24/7
Country: United States

▶ Playing...

🔀 Shuffle Active (Manual mode)
   Station 3 of session
   
─── Shuffle History ───
  ← Jazz FM London
  ← WBGO Jazz 88.3
  → Smooth Jazz 24/7  ← Current

f: Fav • s: List • v: Vote • n: Next shuffle • b: Previous • h: Stop shuffle • ?: Help
```

---

### Configuration File Structure**

The settings would be saved to `~/.config/tera/shuffle.yaml`:

```yaml
# Shuffle Mode Settings
shuffle:
  # Auto-advance to next station after interval
  auto_advance: false
  
  # Interval in minutes before auto-advancing
  # Valid values: 1, 3, 5, 10, 15
  interval_minutes: 5
  
  # Remember shuffle history for back navigation
  remember_history: true
  
  # Number of stations to keep in history
  # Valid values: 3, 5, 7, 10
  max_history: 5
  
  # Persist shuffle state across sessions (future enhancement)
  persist_session: false
```

---

### Summary of Changes to Settings Page

**New menu item:**
```go
// In settings.go - add to menuItems
components.NewMenuItem("Shuffle Settings", "Configure shuffle mode behavior", "2"),
```

**New state:**
```go
const (
    settingsStateMenu settingsState = iota
    settingsStateTheme
    settingsStateShuffle        // NEW
    settingsStateHistory
    settingsStateUpdates
    settingsStateAbout
)
```

**Benefits of this approach:**
1. ✅ **Centralized configuration** - all settings in one place
2. ✅ **Persistent** - settings saved to disk, survive app restart
3. ✅ **Discoverable** - users know where to find shuffle settings
4. ✅ **Consistent UX** - matches existing settings pattern
5. ✅ **Flexible** - easy to add more shuffle options later

---

### **Keyboard Shortcuts Summary**

| Context | Key | Action |
|---------|-----|--------|
| I Feel Lucky Input | `t` | Toggle shuffle mode on/off |
| Shuffle Playing | `h` | Stop shuffle & keep playing |
| Shuffle Playing | `n` | Next shuffle station |
| Shuffle Playing | `b` | Previous shuffle station (from history) |
| Shuffle Playing | `p` | Pause/resume auto-advance timer |
| Settings Menu | `2` | Open Shuffle Settings |

---

Adjust:
- The number of interval options
- History size limits
- Default settings
- The layout of any screens

## Progress

Perfect! I'll implement the shuffle mode feature step by step. Let me start:

## Step 1: Create the shuffle configuration modelsNow let's create the shuffle config file handler:## Step 2: Create the shuffle manager## Step 3: Create the shuffle settings UI## Step 4: Integrate shuffle settings into the main settings menu

### Summary of Shuffle Mode Implementation (Part 1/2)

I've successfully implemented the **foundation and settings** for the shuffle mode feature. Here's what has been completed:

#### ✅ Completed Components:

1. **Configuration System** (`internal/storage/`)
   - Added `ShuffleConfig` struct to `models.go` with fields for:
     - Auto-advance (on/off)
     - Interval minutes (1, 3, 5, 10, 15)
     - Remember history (on/off)
     - Max history size (3, 5, 7, 10)
   - Created `shuffle_config.go` with load/save functions
   - Config stored at `~/.config/tera/shuffle.yaml`

2. **Shuffle Manager** (`internal/shuffle/manager.go`)
   - Core shuffle logic with:
     - Random station selection (no repeats until exhausted)
     - History management (configurable size)
     - Auto-advance timer with pause/resume
     - Session tracking (station count, keyword)
   - Timer integration for countdown display

3. **Shuffle Settings UI** (`internal/ui/shuffle_settings.go`)
   - Complete settings page with 3 sub-screens:
     - Main menu (toggle settings, navigate to sub-pages)
     - Interval selection (1/3/5/10/15 minutes)
     - History size selection (3/5/7/10 stations)
   - Auto-save functionality
   - Current settings display
   - Keyboard shortcuts (1-6 for menu, navigation keys)

4. **Integration with Settings Menu** (`internal/ui/settings.go`)
   - Added "Shuffle Settings" as option #2
   - Updated all keyboard shortcuts (now 1-5 instead of 1-4)
   - Proper navigation flow to shuffle settings page

5. **App-Level Integration** (`internal/ui/app.go`)
   - Added `screenShuffleSettings` screen type
   - Added `shuffleSettingsScreen` field to App struct
   - Implemented navigation handling
   - View rendering integration

### 🚧 Still To Implement (Part 2):

1. **Lucky Screen Integration**
   - Add shuffle toggle ('t' key) in the I Feel Lucky input screen
   - Show shuffle status indicator
   - Pass stations to shuffle manager

2. **Shuffle Playback Mode**
   - New playback state for shuffle mode
   - Display current station with shuffle indicator (🔀)
   - Show countdown timer (when auto-advance enabled)
   - Display shuffle history
   - Keyboard controls:
     - `n` - Next shuffle station
     - `b` - Previous station (from history)
     - `h` - Stop shuffle (keep playing current)
     - `p` - Pause/resume timer
     - `f`, `s`, `v` - Favorite, save, vote (existing)

3. **Testing & Refinement**
   - Test all settings save/load
   - Test shuffle flow end-to-end
   - Handle edge cases (empty stations, stream failures)

The foundation is solid and ready for the playback integration. Would you like me to continue with Part 2 (integrating shuffle into the Lucky screen and adding the playback mode)?

Great! Let me continue with Part 2 - integrating shuffle into the Lucky screen and adding the playback mode.

## Part 2: Lucky Screen Integration & Shuffle Playback

### Summary - Almost Complete!

We've successfully implemented most of the shuffle mode. Here's what we've added to `lucky.go`:

✅ New state: `luckyStateShufflePlaying`
✅ Shuffle fields in LuckyModel
✅ New message types for shuffle
✅ Loaded shuffle config in NewLuckyModel
✅ Updated Update() to handle shuffle messages and timer
✅ Added 't' key toggle in updateInput
✅ Updated all search triggers to use shuffle when enabled

### What's left to add to lucky.go:

1. Add the three new methods at the end of the file (before reloadSearchHistory):
   - `updateShufflePlaying` 
   - `searchForShuffle`
   - `shuffleTimerTick`
   - `viewShufflePlaying`

2. Update the View() method to handle `luckyStateShufflePlaying`

3. Update `viewInput()` to show shuffle toggle status

Let me provide these final pieces in a compact format. Would you like me to:

A) Provide the complete updated `lucky.go` file 
B) Provide just the code snippets to add
C) Move on to tests and README updates (the shuffle implementation will work once the methods from the helper file are added)

What would you prefer?