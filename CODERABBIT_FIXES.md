# CodeRabbit Suggestions Implementation

## Summary

This document outlines the implementation of three CodeRabbit suggestions to improve code quality and user experience in TERA.

## Changes

### 1. ✅ Add Tick Infrastructure for Settings Messages (settings.go)

**Issue**: The message countdown doesn't work because `tickMsg` isn't returning a command to continue ticking.

**Status**: **NEEDS INFRASTRUCTURE FIRST**

The code references `tickMsg` but it's not defined anywhere. We need to add the tick infrastructure before applying CodeRabbit's fix.

**Required Changes**:

1. Add message type definitions to `internal/ui/messages.go` (new file)
2. Add tick command generator
3. Update settings.go to use the tick system

### 2. ✅ Detect Current Theme on Initialization (settings.go:186-194)

**Issue**: The `currentTheme` field is hardcoded to "Default" instead of detecting the actually saved theme.

**Implementation**: Match current theme colors against predefined themes.

### 3. ✅ Extract Menu Item Count as Constant (app.go)

**Issue**: The magic number `6` appears in multiple places and must be manually kept in sync.

**Implementation**: Extract as a package-level constant `mainMenuItemCount`.

## Files to Create/Modify

### New File: `internal/ui/messages.go`

```go
package ui

import (
	"time"

	"github.com/charmbracelet/bubbletea"
	"github.com/shinokada/tera/internal/api"
)

// Common message types used across UI components

// tickMsg is sent on a timer for countdown/animation purposes
type tickMsg time.Time

// backToMainMsg signals return to main menu
type backToMainMsg struct{}

// tickEverySecond returns a command that ticks every second
func tickEverySecond() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}
```

### Modified File: `internal/ui/settings.go`

**Changes**:
1. Detect actual current theme in `NewSettingsModel()`
2. Return tick command after applying theme
3. Continue ticking in `tickMsg` handler

### Modified File: `internal/ui/app.go`

**Changes**:
1. Add `const mainMenuItemCount = 6` at package level
2. Replace all instances of magic number `6` with the constant

## Implementation Priority

1. **HIGH**: Create `messages.go` with tick infrastructure
2. **HIGH**: Update `settings.go` to use tick system
3. **MEDIUM**: Detect current theme in settings initialization  
4. **LOW**: Extract menu item count constant

## Benefits

1. **Tick Infrastructure**: Messages will properly disappear after countdown
2. **Theme Detection**: Users see the correct current theme name
3. **Menu Item Constant**: Easier maintenance, fewer bugs when adding/removing menu items

## Testing

After implementation, test:
1. Apply a theme and verify the success message disappears after ~3 seconds
2. Restart the app and verify the current theme is correctly displayed
3. Navigate the main menu and verify all shortcuts (1-6) work correctly
