# Keyboard Shortcut Standardization - Complete Update

## Changes Made

### Removed Shortcuts
- ❌ `0` - Go back one level (replaced by `Esc`)
- ❌ `00` - Go to main menu (no longer needed)

### Standardized Shortcuts  
- ✅ `Esc` - Go back one level (industry standard)
- ✅ `q` - Quit application (industry standard)
- ✅ `Ctrl+C` - Force quit (already existed)

## Benefits

1. **Industry Standard**: Matches vim, less, htop, top, man pages
2. **Better Ergonomics**: Corner keys vs. top row numbers
3. **Simpler Mental Model**: One key back, one key quit
4. **More Discoverable**: Users expect these keys
5. **No Conflicts**: Works consistently everywhere

## Files Modified

### Code Files

#### internal/ui/search.go
- `handleMenuInput()`: Changed `0/esc` → `esc` for back, added `q` for quit
- `handleTextInput()`: Changed `0` → `esc` for back, `00/esc` → `q` for quit  
- `handleResultsInput()`: Added `q` for quit
- `handleStationInfoInput()`: Changed `0` → `q` for quit
- `handlePlayerUpdate()`: Separated `q` (stop & save prompt) from `esc` (just stop)
- `handleSavePrompt()`: Added `q` for quit
- Updated all View() help text

#### internal/ui/play.go
- `updateListSelection()`: Changed `esc/0` → `esc` for back, added `q` for quit
- `updateStationSelection()`: Changed `esc/0` → `esc` for back, added `q` for quit
- `updatePlaying()`: Removed `0` from `q/esc/0`
- Updated all help text in views

### Updated Help Text

**Search Menu:**
```text
Before: ↑↓/jk: Navigate • Enter: Select • 1-6: Quick select • 0/Esc: Back
After:  ↑↓/jk: Navigate • Enter: Select • 1-6: Quick select • Esc: Back • q: Quit
```

**Search Input:**
```text
Before: Enter) Search  |  0) Back  |  00/Esc) Main Menu
After:  Enter) Search  |  Esc) Back  |  q) Quit
```

**Search Results:**
```text
Before: Enter) Play  |  Esc) Back  |  q) Quit
After:  (No change - was already correct)
```

**Now Playing:**
```text
Before: q/Esc/0) Stop  |  s) Save to Quick Favorites
After:  q) Stop & Save Prompt  |  Esc) Stop & Back  |  s) Save to Quick Favorites
```

**Station Info:**
```text
Before: ↑↓/jk: Navigate • Enter: Select • 1-3: Quick select • 0: Main • Esc: Back
After:  ↑↓/jk: Navigate • Enter: Select • 1-3: Quick select • Esc: Back • q: Quit
```

**Save Prompt:**
```text
Before: y/1: Yes • n/2/Esc: No
After:  y/1: Yes • n/2/Esc: No • q: Quit
```

**Play Screen - List Selection:**
```text
Before: ↑/↓: navigate • enter: select • esc/0: back to menu
After:  ↑/↓: navigate • enter: select • esc: back • q: quit
```

**Play Screen - Station Selection:**
```text
Before: ↑/↓: navigate • /: filter • enter: play • esc/0: back
After:  ↑/↓: navigate • /: filter • enter: play • esc: back • q: quit
```

**Play Screen - Playing:**
```text
Before: q/esc/0: stop • s: save to favorites
After:  q/esc: stop • s: save to favorites
```

## Behavior Changes

### During Playback (Search)
**Before:**
- `q`, `esc`, or `0` all stopped playback and showed save prompt

**After:**
- `q` - Stops playback and shows save prompt ✅
- `Esc` - Stops playback and goes back WITHOUT save prompt ✅

**Rationale:** Gives users two options - `q` for careful exit with save prompt, `Esc` for quick exit

### Navigation
**Before:**
- `0` - Back one level
- `00` - Jump to main menu
- `Esc` - Back one level (duplicate of `0`)

**After:**
- `Esc` - Back one level (standard)
- `q` - Quit application (standard)

**Rationale:** Simpler, more standard, no need for multi-key sequences

## Testing Checklist

- [x] Search menu: `Esc` goes back, `q` quits
- [x] Search input: `Esc` goes back, `q` quits
- [x] Search results: `Esc` goes back, `q` quits
- [x] Now playing: `q` stops & prompts, `Esc` stops & goes back
- [x] Save prompt: `q` quits app
- [x] Play screen lists: `Esc` goes back, `q` quits
- [x] Play screen stations: `Esc` goes back, `q` quits
- [x] Play screen playing: `q` or `Esc` stops
- [x] All help text updated
- [x] No `0` or `00` shortcuts remain

## Migration Notes

Users familiar with `0`/`00` will need to learn:
- Use `Esc` instead of `0` for back
- Use `q` for quit (no more `00`)

This is a **breaking change** but worth it for:
- Industry standard compliance
- Better UX
- Simpler mental model
