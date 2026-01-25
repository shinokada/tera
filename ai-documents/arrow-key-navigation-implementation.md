# Arrow Key Navigation Implementation Plan

## Overview
Add list-based navigation using arrow keys (↑/↓ or j/k) to main menu and search menu, while maintaining number-based shortcuts.

## Current State
- **Main Menu**: Number-only navigation (1, 2, 3...)
- **Search Menu**: Number-only navigation (1-6 for search types)
- **Keyboard Guide**: States arrow keys should work in menus

## Target State
- **Main Menu**: Arrow keys + number shortcuts
- **Search Menu**: Arrow keys + number shortcuts
- Both maintain existing keyboard shortcuts for compatibility

## Implementation Approach

### 1. Main Menu
Convert from simple number selection to list-based navigation:

**Features:**
- Arrow keys (↑/↓) or vim keys (j/k) to navigate
- Enter to select highlighted item
- Number shortcuts (1-6, 10-19) still work
- Esc/q to quit

**Menu Items:**
1. Play from Favorites
2. Search Stations
3. Manage Lists (coming soon)
4. I Feel Lucky (coming soon)
5. Delete Station (coming soon)
6. Gist Management (coming soon)
7. Exit

**Quick Play Section:**
- 10-19: Direct play shortcuts for first 10 favorites
- Display as separate section below main menu
- Highlight on selection

### 2. Search Menu
Convert from number selection to list-based navigation:

**Features:**
- Arrow keys (↑/↓) or vim keys (j/k) to navigate
- Enter to select highlighted search type
- Number shortcuts (1-6) still work
- Esc/0 to return to main menu

**Menu Items:**
1. Search by Tag
2. Search by Name
3. Search by Language
4. Search by Country Code
5. Search by State
6. Advanced Search

## Technical Implementation

### Component Changes

#### 1. Use Bubble Tea List Component
```go
import "github.com/charmbracelet/bubbles/list"
```

**Benefits:**
- Built-in arrow key navigation
- Filtering support
- Customizable rendering
- Accessibility

#### 2. Main Menu Model
```go
type MainMenuModel struct {
    list          list.Model
    quickPlay     []api.Station
    width         int
    height        int
}

type menuItem struct {
    title       string
    desc        string
    shortcut    string
    action      Screen
}
```

#### 3. Search Menu Model
```go
type SearchMenuModel struct {
    list       list.Model
    width      int
    height     int
}

type searchMenuItem struct {
    title      string
    desc       string
    searchType api.SearchType
}
```

### Key Bindings

#### Main Menu
- `↑`, `k`: Previous item
- `↓`, `j`: Next item
- `g`, `Home`: First item
- `G`, `End`: Last item
- `Enter`: Select item
- `1-6`: Direct selection
- `10-19`: Quick play
- `0`, `q`: Exit

#### Search Menu
- `↑`, `k`: Previous item
- `↓`, `j`: Next item
- `Enter`: Select search type
- `1-6`: Direct selection
- `0`, `Esc`: Back to main menu

### Visual Design

#### Main Menu Layout
```text
╔════════════════════════════════════════╗
║   TERA - Terminal Radio                ║
╠════════════════════════════════════════╣
║ > 1. Play from Favorites               ║  <-- highlighted
║   2. Search Stations                   ║
║   3. Manage Lists (coming soon)        ║
║   4. I Feel Lucky (coming soon)        ║
║   5. Delete Station (coming soon)      ║
║   6. Gist Management (coming soon)     ║
║   0. Exit                              ║
╠════════════════════════════════════════╣
║   Quick Play Favorites                 ║
║   10. ▶ Jazz FM                        ║
║   11. ▶ BBC Radio 1                    ║
║   12. ▶ Classical KDFC                 ║
╠════════════════════════════════════════╣
║ ↑↓/jk: Navigate  Enter: Select        ║
║ 1-6: Quick select  q: Quit            ║
╚════════════════════════════════════════╝
```

#### Search Menu Layout
```text
╔════════════════════════════════════════╗
║   🔍 Search Radio Stations             ║
╠════════════════════════════════════════╣
║ > 1. Search by Tag                     ║  <-- highlighted
║   2. Search by Name                    ║
║   3. Search by Language                ║
║   4. Search by Country Code            ║
║   5. Search by State                   ║
║   6. Advanced Search                   ║
╠════════════════════════════════════════╣
║ ↑↓/jk: Navigate  Enter: Select        ║
║ 1-6: Quick select  0/Esc: Back        ║
╚════════════════════════════════════════╝
```

## File Changes

### New Files
- `internal/ui/components/menu.go` - Reusable menu component

### Modified Files
1. `internal/ui/app.go`
   - Add MainMenuModel
   - Convert viewMainMenu() to use list
   - Handle arrow key navigation
   - Maintain number shortcuts

2. `internal/ui/search.go`
   - Convert searchStateMenu to use list
   - Add handleMenuNavigation()
   - Maintain number shortcuts

### Backward Compatibility
- All existing number shortcuts remain functional
- Users can choose arrow keys or numbers
- No breaking changes to behavior

## Testing Plan

### Unit Tests
1. Main menu list creation
2. Search menu list creation
3. Arrow key handling
4. Number shortcut handling
5. Quick play shortcuts

### Integration Tests
1. Navigate main menu with arrows
2. Navigate search menu with arrows
3. Number shortcuts still work
4. Combination of both methods

### Manual Testing
1. Test all arrow key combinations
2. Test all vim key combinations (j/k)
3. Test all number shortcuts
4. Test Esc/q/0 navigation
5. Test quick play shortcuts (10-19)

## Implementation Steps

1. **Create menu component**
   - Build reusable menu list component
   - Add custom rendering
   - Handle shortcuts

2. **Update main menu**
   - Convert to list-based
   - Add quick play section
   - Integrate with app.go

3. **Update search menu**
   - Convert to list-based
   - Maintain search flow
   - Integrate with search.go

4. **Add tests**
   - Unit tests for components
   - Integration tests for navigation

5. **Update documentation**
   - Update keyboard shortcuts guide
   - Add navigation examples
   - Document both methods

## Benefits

### User Experience
- Consistent navigation across all menus
- Familiar vim-style keys
- Muscle memory from other TUI apps
- Accessibility improvements

### Code Quality
- Reusable menu component
- Standard Bubble Tea patterns
- Better separation of concerns
- Easier to test

### Maintainability
- Centralized menu logic
- Easier to add new menu items
- Standard component API
- Better documentation

## Migration Path

### Phase 1: Implement
- Add list components
- Maintain number shortcuts
- Update UI rendering

### Phase 2: Test
- Comprehensive testing
- User feedback
- Bug fixes

### Phase 3: Document
- Update guides
- Add examples
- Create migration notes
