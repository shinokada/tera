# Keyboard Navigation and Feature Updates

## Summary of Changes

### 1. Arrow Key Navigation Issue

**Current Behavior**: Main menu and search type menu only respond to number keys (1, 2, 3, etc.)

**Expected Behavior**: Should support arrow keys (↑/↓) and vim keys (j/k) for navigation with visual highlighting

**Status**: ⚠️ **Needs Implementation**

The current menus are simple text-based menus. To support arrow key navigation, we need to:
- Convert them to list-based selection models
- Add visual highlighting for the current selection
- Keep number shortcuts as alternatives for power users

**Recommendation**: Implement in next phase. This requires converting the menu system to use the bubbles list component.

### 2. Missing Keyboard Shortcuts

After reviewing `golang/spec-docs/keyboard-shortcuts-guide.md`, here are the missing features:

**High Priority** (Spec says should exist):
- ✗ Arrow key navigation in menus (↑/↓, j/k)
- ✗ Help system (`?` key) - Not implemented anywhere
- ✗ Info preview (`i` key) - Not in search results or play screen
- ✓ Delete from Play Screen (`d` key) - Needs to be moved from main menu

**Medium Priority**:
- ✓ Filter in lists (`/` key) - Already works (bubbles list default)
- ✓ Save during playback (`s` key) - Partially implemented in play.go
- ⚠️ Quick jump to first/last (`g`/`G`) - May work via bubbles list
- ⚠️ Page up/down (Ctrl+u/Ctrl+d) - May work via bubbles list

**Implementation Plan**:
1. Move delete functionality to Play Screen (this session)
2. Add arrow key navigation to menus (next session)
3. Implement help system (future)
4. Add info preview overlay (future)

### 3. Delete Station Location ✓ FIXED

**Original Design**: Delete Station was main menu option #4

**Updated Design**: Delete moved to Play Screen with `d` key

**Rationale**:
- More intuitive - delete while viewing stations
- Faster workflow - no need to go to separate screen
- Consistent with browsing experience
- Main menu becomes simpler

**Flow Chart Updates Made**:
- ✓ Removed "DeleteStation" from Application Overview diagram
- ✓ Removed option 4 from Main Menu flow
- ✓ Renumbered options: 4=Lucky, 5=Gist (was 5=Lucky, 6=Gist)
- ✓ Added delete flow to Play Screen with `d` key
- ✓ Added info preview flow to Play Screen with `i` key  
- ✓ Removed entire "Delete Station Screen" section (was section 6)
- ✓ Renumbered all subsequent sections

**New Main Menu Options**:
```text
1. Play from Favorites
2. Search Stations
3. Manage Lists
4. I Feel Lucky
5. Gist Management
0/q. Exit
10-19. Quick Play Favorites
```

## Updated Flow Charts

### Main Menu Changes
- Removed option 4 (Delete Station)
- Option 4 is now "I Feel Lucky" (was 5)
- Option 5 is now "Gist Management" (was 6)

### Play Screen New Features

**New keyboard shortcuts added to flow**:
- `d` - Delete current station (with Y/N confirmation)
- `i` - Show info preview overlay

**Delete Flow**:
```text
StationInput → d → ConfirmDelete? → Yes → DeleteStation → SaveJSON → Success
                                  → No → ShowStations
```

**Info Preview Flow**:
```text
StationInput → i → InfoPreview → Esc/i → ShowStations
```

## Next Steps

### This Session - Documentation Only ✓ DONE
- ✓ Updated flow-charts.md
- ✓ Created missing-features-plan.md
- ✓ Created this summary document

### Next Session - Implementation
1. **Add Delete to Play Screen**
   - Add `d` key handler in `updateStationSelection()`
   - Show Y/N confirmation dialog
   - Remove station from list
   - Save updated JSON
   - Show success/error message

2. **Convert Menus to List Navigation**
   - Create list models for main menu
   - Create list model for search type menu
   - Add highlighting/cursor
   - Maintain number shortcuts as alternatives

3. **Add Info Preview**
   - Create info overlay component
   - Add `i` key handler in Play Screen
   - Add `i` key handler in Search Results
   - Show station details without leaving list

### Future - Nice to Have
1. Help system (`?` key)
2. Complete save during playback for all screens
3. Quick navigation keys (g/G, Ctrl+u/d)
4. Verify page up/down works properly

## Files That Need Updates

### For Delete Feature:
- `internal/ui/play.go` - Add delete handler
- May need confirmation dialog component

### For Menu Navigation:
- `internal/ui/app.go` - Convert main menu to list
- `internal/ui/search.go` - Convert search type menu to list

### For Info Preview:
- `internal/ui/components/info_overlay.go` - New component
- `internal/ui/play.go` - Add info overlay state
- `internal/ui/search.go` - Add info overlay state

## Key Decisions Made

1. ✓ Delete station belongs in Play Screen, not main menu
2. ✓ Use simple Y/N confirmation (not "type yes")
3. ✓ Keep number shortcuts even after adding arrow navigation
4. ⏳ Defer help system and info preview to future implementation
5. ⏳ Focus on delete functionality first (highest value)

## Documentation Status

- ✓ flow-charts.md updated
- ✓ Implementation plan created
- ✓ This summary created
- ⏳ keyboard-shortcuts-guide.md needs updating after implementation
- ⏳ README needs updating after implementation
