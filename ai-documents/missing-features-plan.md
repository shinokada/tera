# Implementation Plan for Missing Features

## Priority 1: Critical UX Issues

### 1. Arrow Key Navigation in Menus ✓ NEEDED
**Current**: Main menu and search menu use number keys only  
**Expected**: Arrow keys (↑/↓) and vim keys (j/k) should work

**Implementation**:
- Convert main menu to list-based selection with highlighted items
- Convert search menu to list-based selection  
- Keep number shortcuts as alternatives
- Add cursor/highlight to show current selection

**Files to modify**:
- `internal/ui/app.go` - Convert main menu to list model
- `internal/ui/search.go` - Convert search type menu to list model

### 2. Delete Station from Play Screen ✓ NEEDED
**Current**: Delete is a separate main menu option (#4)
**Expected**: Delete station while browsing in Play Screen

**Implementation**:
- Add `d` key in Play Screen station selection
- Show confirmation dialog
- Remove station from current list
- Update JSON file
- Refresh station list

**Files to modify**:
- `internal/ui/play.go` - Add delete functionality in `updateStationSelection()`

## Priority 2: Important Features

### 3. Help System (`?` key) ✓ NEEDED
**Status**: Not implemented  
**Expected**: Context-sensitive help on every screen

**Implementation**:
- Create help overlay/modal component
- Add help text for each screen
- Toggle with `?` key
- Show available keys and their functions

**Files to create/modify**:
- `internal/ui/components/help.go` - Help overlay component
- Add help state to each screen model

### 4. Info Preview (`i` key) ✓ NEEDED
**Current**: Must select station to see details  
**Expected**: Quick info overlay without leaving list

**Implementation**:
- Show station info overlay on `i` press
- Keep list in background
- Close with Esc or i again

**Files to modify**:
- `internal/ui/play.go` - Add info overlay state
- `internal/ui/search.go` - Add info preview in results

### 5. Quick Navigation (g/G, PgUp/PgDn) 
**Status**: Partially works (bubbles list has some support)  
**Expected**: Full implementation with Ctrl+u/Ctrl+d

**Implementation**:
- Verify bubbles list default behavior
- Add Ctrl+u/Ctrl+d if not present
- Ensure g/G work for first/last

## Priority 3: Nice to Have

### 6. Enhanced Save During Playback
**Status**: Partially implemented  
**Expected**: Save with `s` key during playback in all screens

**Implementation**:
- Ensure `s` key works in Play Screen (already done)
- Ensure `s` key works in Search Screen playback
- Ensure `s` key works in Lucky Screen playback
- Add duplicate checking

**Files to verify**:
- `internal/ui/play.go` - Verify save during playback
- `internal/ui/search.go` - Add save during playback
- Create Lucky screen with save support

## Flow Chart Updates Needed

### Update: flow-charts.md

**Remove from Main Menu**:
- Option 4: Delete Station

**Add to Play Screen**:
- `d` key: Delete selected station (with confirmation)

**Updated Main Menu Options**:
```text
1. Play from Favorites
2. Search Stations  
3. Manage Lists
4. I Feel Lucky (moved up from 5)
5. Gist Management (moved up from 6)
0/q. Exit
10-19. Quick Play Favorites
```

**Updated Play Screen Flow**:
```mermaid
StationInput -->|d| ConfirmDelete[Confirm Delete?]
ConfirmDelete -->|Yes| DeleteStation[Remove from List]
ConfirmDelete -->|No| ShowStations
DeleteStation --> UpdateJSON[Save Updated List]
UpdateJSON --> ShowStations
```

## Implementation Order

1. **Phase 1** (This session):
   - Add arrow key navigation to main menu
   - Add arrow key navigation to search menu
   - Move delete functionality to Play Screen
   - Update flow charts

2. **Phase 2** (Next session):
   - Implement help system
   - Add info preview (i key)
   - Complete save during playback for all screens

3. **Phase 3** (Future):
   - Implement Lucky screen
   - Add quick navigation keys
   - Polish and testing

## Questions to Confirm

1. ✓ Delete station should be in Play Screen with `d` key?
2. ✓ Main menu should support arrow key selection with highlighting?
3. ✓ Search type menu should also support arrow keys?
4. Should we keep number shortcuts as alternatives? (Recommended: Yes)
5. Should delete require typing "yes" or just Y/N? (Recommend: Just Y/N for speed)
