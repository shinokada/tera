# Phase 2 Implementation Instructions

## Quick Overview

Phase 2 adds a graphical Settings UI for customizing the header. After implementation, users can:
- Select header mode (default/text/ascii/none) from a menu
- Enter custom text with a text input widget
- Paste ASCII art with a multi-line editor
- Adjust alignment, width, colors, padding
- Preview changes before saving
- Save and see changes immediately

## Current Issue Fix

**Problem**: Your config file at `~/.config/tera/appearance_config.yaml` is not being applied because the header renderer loads config once at startup and never reloads it.

**Quick Fix**: The implementation below includes a `Reload()` call that fixes this issue.

## Files Summary

| File | Action | Purpose |
|------|--------|---------|
| `internal/ui/appearance_settings.go` | **CREATE** | New UI for appearance settings |
| `internal/ui/components/help.go` | **MODIFY** | Add help for appearance screen |
| `internal/ui/app.go` | **MODIFY** | Add new screen and routing |
| `internal/ui/settings.go` | **MODIFY** | Add "Appearance" menu item |

## Detailed Instructions

### 1. Create appearance_settings.go

The complete file is available at `/home/claude/appearance_settings.go`. You need to copy it to:

```bash
cp /path/to/appearance_settings.go /Users/shinichiokada/Terminal-Tools/tera/internal/ui/appearance_settings.go
```

This file contains:
- `AppearanceSettingsModel` struct and state machine
- Input widgets for text, ASCII art, width, color
- List selectors for mode and alignment
- Preview functionality
- Save/reset functionality with config reload

### 2. Update components/help.go

Add this function at the end of the file:

```go
// CreateAppearanceHelp creates help sections for appearance settings
func CreateAppearanceHelp() []HelpSection {
	return []HelpSection{
		{
			Title: "Navigation",
			Items: []HelpItem{
				{"↑↓/jk", "Navigate menu"},
				{"Enter", "Select/Edit"},
				{"Esc/q", "Back"},
				{"?", "Toggle help"},
			},
		},
		{
			Title: "Header Modes",
			Items: []HelpItem{
				{"default", "Show 'TERA'"},
				{"text", "Custom text"},
				{"ascii", "ASCII art"},
				{"none", "No header"},
			},
		},
		{
			Title: "Actions",
			Items: []HelpItem{
				{"Save", "Apply changes"},
				{"Preview", "See result"},
				{"Reset", "Restore defaults"},
			},
		},
	}
}
```

### 3. Update app.go

#### 3a. Add Screen Constant

Find around line 18-26 where Screen constants are defined:

```go
const (
	screenMainMenu Screen = iota
	screenPlay
	screenSearch
	screenList
	screenLucky
	screenGist
	screenSettings
	screenShuffleSettings
	screenConnectionSettings
	screenAppearanceSettings  // ADD THIS LINE
)
```

#### 3b. Add Field to App Struct

Find around line 31-66:

```go
type App struct {
	// ... existing fields ...
	shuffleSettingsScreen    ShuffleSettingsModel
	connectionSettingsScreen ConnectionSettingsModel
	appearanceSettingsScreen AppearanceSettingsModel  // ADD THIS LINE
	apiClient                *api.Client
	// ... rest of fields ...
}
```

#### 3c. Update Update() Function - navigateMsg Handler

Find the `case navigateMsg:` section in the `Update()` function and add:

```go
case navigateMsg:
	app.screen = msg.screen
	switch msg.screen {
	// ... existing cases ...
	case screenAppearanceSettings:
		if app.appearanceSettingsScreen.width == 0 {
			app.appearanceSettingsScreen = NewAppearanceSettingsModel()
			app.appearanceSettingsScreen.width = app.width
			app.appearanceSettingsScreen.height = app.height
		}
		return app, app.appearanceSettingsScreen.Init()
	}
```

#### 3d. Update Update() Function - Screen Switch

Find the main screen switch statement and add:

```go
switch app.screen {
// ... existing cases ...
case screenAppearanceSettings:
	newModel, cmd := app.appearanceSettingsScreen.Update(msg)
	app.appearanceSettingsScreen = newModel
	return app, cmd
}
```

#### 3e. Update View() Function

Find the View() function's switch statement and add:

```go
func (app App) View() string {
	switch app.screen {
	// ... existing cases ...
	case screenAppearanceSettings:
		return app.appearanceSettingsScreen.View()
	}
}
```

### 4. Update settings.go

#### 4a. Add Menu Item

Find `NewSettingsModel()` function where items are created (around line 213):

```go
items := []list.Item{
	listItem{title: "Theme", description: "Change color theme"},
	listItem{title: "Appearance", description: "Customize header"}, // ADD THIS
	listItem{title: "Shuffle Settings", description: "Configure tag shuffle"},
	listItem{title: "Connection Settings", description: "Manage API timeouts"},
	// ... rest of items ...
}
```

#### 4b. Handle Menu Selection

Find `handleMenuSelection()` function and add:

```go
func (m *SettingsModel) handleMenuSelection() (SettingsModel, tea.Cmd) {
	if item, ok := m.menuList.SelectedItem().(listItem); ok {
		switch item.title {
		case "Theme":
			m.state = settingsStateTheme
			return *m, nil
		case "Appearance":  // ADD THIS CASE
			return *m, navigateTo(screenAppearanceSettings)
		case "Shuffle Settings":
			return *m, navigateTo(screenShuffleSettings)
		// ... rest of cases ...
		}
	}
	return *m, nil
}
```

## Build and Test

```bash
cd /Users/shinichiokada/Terminal-Tools/tera

# Build
go build

# Test
./tera
```

## Usage Flow

1. **Launch TERA**: `./tera`
2. **Go to Settings**: Press `6` or navigate to "Settings"
3. **Select Appearance**: Choose "Appearance" from settings menu
4. **Configure Header**:
   - Select "Header Mode" to choose: default, text, ascii, or none
   - Select "Custom Text" to enter your text (e.g., "🎵 My Radio Station 🎵")
   - Select "Alignment" to choose: left, center, or right
   - Select "Width" to set header width (10-120)
   - Select "Color" to set color (auto, hex, or ANSI code)
   - Toggle "Bold" by pressing Enter
   - Adjust "Padding Top" and "Padding Bottom" by pressing Enter (cycles 0-5)
   - Select "Edit ASCII Art" to paste multi-line ASCII art
5. **Preview**: Select "Preview" to see how it looks
6. **Save**: Select "Save" to apply changes
7. **Verify**: Press Esc to go back - header should update immediately!

## Example Configurations

### Simple Text Header
```yaml
appearance:
  header:
    mode: "text"
    custom_text: "🎵 Radio"
    alignment: "center"
    width: 50
    color: "auto"
    bold: true
```

### ASCII Art Header
```yaml
appearance:
  header:
    mode: "ascii"
    ascii_art: |
       ____      _    ____ ___ ___  
      |  _ \    / \  |  _ \_ _/ _ \ 
      | |_) |  / _ \ | | | | | | | |
      |  _ <  / ___ \| |_| | | |_| |
      |_| \_\/_/   \_\____/___\___/ 
    alignment: "center"
    width: 60
```

### No Header (Maximum Space)
```yaml
appearance:
  header:
    mode: "none"
```

## Key Features

✅ **Live Preview** - See changes before saving
✅ **Validation** - Prevents invalid configurations
✅ **Instant Apply** - No restart needed
✅ **Reset Option** - Restore defaults easily
✅ **Help System** - Press `?` for help
✅ **Config Reload** - Fixes the current issue where changes don't apply

## Troubleshooting

### "Cannot find package"
**Cause**: Missing imports in appearance_settings.go
**Fix**: Make sure all imports are present:
```go
import (
	"fmt"
	"strings"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/shinokada/tera/internal/storage"
	"github.com/shinokada/tera/internal/ui/components"
)
```

### "undefined: CreateAppearanceHelp"
**Cause**: Function not added to help.go
**Fix**: Add the `CreateAppearanceHelp()` function to components/help.go

### Changes don't take effect
**Cause**: Old bug where config wasn't reloaded
**Fix**: The new code includes `globalHeaderRenderer.Reload()` after save

### Can't navigate to Appearance Settings
**Cause**: Missing screen constant or switch cases
**Fix**: Review steps 3 and 4 carefully

## What This Fixes

1. ✅ **Your Current Issue**: Config file changes now apply immediately via `Reload()`
2. ✅ **User Experience**: No more manual YAML editing
3. ✅ **Validation**: Prevents bad configurations
4. ✅ **Preview**: See changes before committing
5. ✅ **Discoverability**: Settings are in the app, not hidden in a file

## Next Steps After Implementation

1. Test all header modes (default, text, ascii, none)
2. Test with different terminal widths
3. Try various ASCII art styles
4. Test color options
5. Verify persistence (settings survive restart)
6. Test error handling (invalid inputs)

## Files Created Reference

All new code is provided in:
- `/home/claude/appearance_settings.go` - Complete UI implementation (copy to `internal/ui/`)
- `/Users/shinichiokada/Terminal-Tools/tera/help_patch.go` - Help function to add

## Implementation Checklist

- [ ] Copy appearance_settings.go to internal/ui/
- [ ] Add CreateAppearanceHelp() to components/help.go
- [ ] Add screenAppearanceSettings constant
- [ ] Add appearanceSettingsScreen field to App
- [ ] Update navigateMsg handler
- [ ] Update Update() screen switch
- [ ] Update View() screen switch  
- [ ] Add "Appearance" to settings menu
- [ ] Update handleMenuSelection()
- [ ] Build successfully
- [ ] Test navigation to Appearance
- [ ] Test all input modes
- [ ] Test save functionality
- [ ] Verify config reload works
- [ ] Test with your existing config file

Total estimated implementation time: 15-30 minutes
