# TERA Navigation Guide

Complete reference for navigating TERA's interface efficiently.

---

## Navigation Philosophy

TERA uses **two complementary navigation systems** for the best experience:

1. **Interactive Menus (fzf)** - Visual menus with arrow keys
2. **Text Prompts** - Keyboard shortcuts when typing

This dual system gives you:
- ⚡ **Speed** - Arrow keys for quick menu navigation
- 🎯 **Flexibility** - Type shortcuts when entering data
- 🛡️ **Safety** - ESC and empty inputs won't cause problems
- 🔄 **Consistency** - Same patterns throughout the app

---

## Quick Reference

### Interactive Menus (fzf)

When you see a menu with numbered options:

| Key            | Action                        |
| -------------- | ----------------------------- |
| ↑ / ↓          | Navigate options              |
| Enter          | Select highlighted option     |
| ESC            | Cancel and go back            |
| Type text      | Fuzzy search/filter           |
| `0) Main Menu` | Select to return to main menu |

**Examples:**
- Main Menu
- Search Menu
- List Menu
- Search Submenu
- Station selection

### Text Input Prompts

When you see a prompt asking you to type:

| Input          | Action                                                        |
| -------------- | ------------------------------------------------------------- |
| `0` or `back`  | Go back to previous menu                                      |
| `00` or `main` | Return to Main Menu                                           |
| Empty + Enter  | Context-dependent: goes back in search, shows error in lists  |
| Type content   | Enter your data                                               |

**Examples:**
- "Type a name to search:"
- "Type a new list name:"
- "Type a list name to delete:"

---

## Navigation Commands Summary

| Context                  | Want to... | Do this                            |
| ------------------------ | ---------- | ---------------------------------- |
| fzf Menu                 | Go back    | Press ESC or select "0) Main Menu" |
| fzf Menu                 | Main menu  | Select "0) Main Menu"              |
| Text Prompt (Search)     | Go back    | Type `0` or press Enter (empty)    |
| Text Prompt (List Mgmt)  | Go back    | Type `0` (empty shows error)       |
| Text Prompt              | Main menu  | Type `00`                          |
| View/Info Page           | Continue   | Press Enter                        |
| Anywhere                 | Force quit | Ctrl+C (emergency)                 |

---

## Detailed Examples

### Example 1: Searching for Stations

```text
TERA - Search by Name

Type '0' to go back to Search Menu, '00' for Main Menu
Type a name to search: jazz      ← Type your search
```

**Navigation options:**
```text
Type a name to search: 0         ← Go back to Search Menu
Type a name to search: 00        ← Go to Main Menu
Type a name to search: [Enter]   ← Go back to Search Menu (empty)
```

### Example 2: Creating a List

```text
TERA - Create New List

My lists:
My-favorites
rock-stations

Type '0' to go back, '00' for main menu
Type a new list name: chill      ← Create new list
```

**Navigation options:**
```text
Type a new list name: 0          ← Cancel, go back to List Menu
Type a new list name: 00         ← Cancel, go to Main Menu
Type a new list name: [Enter]    ← Error (empty name not allowed)
```

### Example 3: Using fzf Menus

```text
TERA SEARCH MENU

  0) Main Menu         ← Use ↑↓ arrows to navigate
  1) Tag              ← Press Enter when highlighted
  2) Name             ← Press ESC to cancel
  3) Language
  4) Country code
  5) State
  6) Advanced search
  7) Exit

Choose an option (arrow keys to navigate): _
```

**Navigation:**
- ↑↓ to move highlight
- Enter to select option
- ESC to go back
- Type to fuzzy search (e.g., type "tag" to jump to Tag option)

### Example 4: Deleting a List

```text
TERA - Delete List

My lists: 
jazz
rock

Type '0' to go back, '00' for main menu
Type a list name to delete: rock    ← Delete rock list
```

**Navigation options:**
```text
Type a list name to delete: 0       ← Cancel deletion
Type a list name to delete: 00      ← Cancel, go to Main Menu
Type a list name to delete: [Enter] ← Error (empty not allowed)
```

---

## List Management Navigation

Special navigation rules when managing lists:

### Create List
- **First prompt**: Select "Create new list" from menu (arrow keys)
- **Second prompt**: Type list name or `0` to cancel

### Delete List
- **First prompt**: Select "Delete list" from menu (arrow keys)
- **Second prompt**: Type list name or `0` to cancel
- **Protection**: Cannot delete "My-favorites"

### Edit/Rename List
- **First prompt**: Type list name to edit or `0` to cancel
- **Second prompt**: Type new name or `0` to cancel
- **Protection**: Cannot rename "My-favorites"

### Common Workflows

**Creating multiple lists:**
```text
Main Menu → List Menu → Create → Type name → (created)
                ↑                                    |
                └────────────────────────────────────┘
Automatically returns to List Menu, ready for more
```

**Quick cancel:**
```text
List Menu → Create → Type '0' → Back to List Menu (no changes)
List Menu → Create → Type '00' → Main Menu (no changes)
```

**Safe editing:**
```text
List Menu → Edit → Type list name → Type '0' → Canceled (no changes)
```

---

## Protected Operations

TERA protects you from common mistakes:

### ❌ Cannot Delete "My Favorites"
```text
Type a list name to delete: My-favorites
❌ Cannot delete My-favorites list!
```

### ❌ Cannot Rename "My Favorites"
```text
Type a list name to edit: My-favorites
❌ Cannot rename My-favorites list!
```

### ❌ Cannot Create Duplicate Lists
```text
Type a new list name: jazz
❌ List 'jazz' already exists!
```

### ❌ Cannot Use Empty Names
```text
Type a new list name: 
❌ List name cannot be empty.
```

### ✅ Duplicate Station Detection
```text
Saving station to: Jazz Collection
⚠️ This station is already in your Jazz Collection list!
Press Enter to continue...
```

---

## Special Features

### Quick Play Favorites

Your "My Favorites" list shows on main menu:

```text
TERA MAIN MENU

1) Play from my list
2) Search radio stations
...

--- Quick Play Favorites ---
10) ▶ BBC World Service    ← Just select number to play
11) ▶ Jazz FM
12) ▶ Classical KDFC
```

**Navigation:**
- Use arrow keys to select
- Press Enter to play immediately
- Press ESC to cancel

### Search Results

After searching, results appear in an interactive list:

```text
TERA - Search Results

  << Main Menu >>      ← Special option to go back
  1. Jazz Radio 24/7
  2. Smooth Jazz Global
  3. Jazz FM London
  ...

> _                    ← Type to filter results
```

**Navigation:**
- ↑↓ to browse results
- Type to filter/search
- Enter to select and play
- ESC to return to search menu
- Select "<< Main Menu >>" to go to main menu

---

## Navigation Tips

### Getting Around Fast

1. **From anywhere to Main Menu**: Type `00` in text prompts
2. **Go back one step**: Press ESC or type `0`
3. **Emergency exit**: Ctrl+C (but try ESC first!)
4. **Made a mistake?**: Just ESC or `0` — nothing is saved until confirmed

### Can't Find Something?

1. Type `00` to return to Main Menu
2. Navigate to the section you need
3. All features accessible from Main Menu

### Keyboard Shortcuts

**In any fzf menu:**
- Start typing to fuzzy search
- Ctrl+N / Ctrl+P = Next/Previous (alternative to arrows)
- Ctrl+C = Force quit (emergency)

**In text prompts:**
- `0` = Back
- `00` = Main Menu
- `back` = Back (alternative)
- `main` = Main Menu (alternative)

### Visual Cues

Look for these prompts:
- 🟡 Yellow text = Navigation hints
- 🔵 Cyan text = Menu headers
- 🟢 Green text = Success messages
- 🔴 Red text = Error messages
- ▶ Symbol = Playable favorite station

---

## First-Time User Guide

When you first install TERA:

1. ✅ Configuration directory auto-created at `~/.config/tera/favorite/`
2. ✅ "My-favorites.json" created with sample stations
3. ✅ All navigation works immediately
4. ✅ No setup required

**Try these first steps:**

```bash
# 1. Launch TERA
tera

# 2. Play a sample station
Select "10) ▶ [Station Name]" from Quick Play Favorites

# 3. Search for your own
Select "2) Search radio stations"
Select "1) Tag"
Type: jazz
Use arrows to select, Enter to play

# 4. Save to favorites
After playing, choose "Yes" to save
Select "My Favorites" from list
```

**Navigation practice:**
- Try pressing ESC in different menus
- Type `0` in various prompts
- Select "0) Main Menu" from menus
- All these are safe—you can't break anything!

---

## Troubleshooting Navigation

### Menu not responding to arrow keys?

**Check fzf installation:**
```bash
which fzf
# Should show: /usr/local/bin/fzf or similar
```

**If missing:**
```bash
brew install fzf              # macOS
sudo apt install fzf          # Debian/Ubuntu
sudo pacman -S fzf           # Arch Linux
```

### Typed `0` but nothing happened?

- Make sure you pressed **Enter** after typing `0`
- Check you're in a **text prompt**, not an fzf menu
- In fzf menus, press **ESC** instead

### Want to exit immediately?

**Proper way:**
1. Press ESC until you reach Main Menu
2. Select "0) Exit"

**Emergency way:**
- Press Ctrl+C (force quit)
- Note: This may leave player running—use proper exit when possible

### Fuzzy search not working?

- Make sure you're in an **fzf menu** (not a text prompt)
- fzf menus show highlighted options and a search cursor
- Text prompts show "Type..." and wait for Enter

---

## Keyboard Reference Card

Print this for quick reference:

```text
┌─────────────────────────────────────────────┐
│         TERA NAVIGATION REFERENCE           │
├─────────────────────────────────────────────┤
│ FZF MENUS                                   │
│  ↑↓         Navigate                        │
│  Enter      Select                          │
│  ESC        Cancel/Back                     │
│  Type       Fuzzy search                    │
│                                             │
│ TEXT PROMPTS                                │
│  0          Back to previous menu           │
│  00         Back to Main Menu               │
│  Empty      Back to previous menu           │
│  back       Back (alternative)              │
│  main       Main Menu (alternative)         │
│                                             │
│ PLAYER CONTROLS                             │
│  q/Space    Pause/Quit                      │
│  9/0        Volume                          │
│  m          Mute                            │
│                                             │
│ EMERGENCY                                   │
│  Ctrl+C     Force quit                      │
└─────────────────────────────────────────────┘
```

---

## Summary

**Remember:**
- 🎯 Interactive menus → Arrow keys and ESC
- ⌨️ Text prompts → Type `0`, `00`, or Enter
- 🛡️ When in doubt → ESC or `0` takes you back safely
- 🚀 Practice makes perfect → Try different paths!

**Navigation is designed to be intuitive:**
- You can't accidentally delete or modify anything
- All destructive actions require confirmation
- ESC and `0` are always safe
- Main Menu is always accessible

Happy listening! 🎵

---

**See also:**
- [Main README](README.md) - Overview and installation
- [Favorites Guide](FAVORITES.md) - Quick play favorites setup
- [Gist Setup](GIST_SETUP.md) - GitHub integration
