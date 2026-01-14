# TERA Navigation Guide

## Two Navigation Systems

TERA uses two complementary navigation systems for the best user experience:

### 1. Interactive Menus (fzf)
**When you see a menu with arrow keys:**
- ⬆️⬇️ Use arrow keys to navigate options
- ↩️ Press Enter to select
- ⎋ Press ESC to cancel/go back
- Look for "0) Main Menu" option at the top

**Examples:**
- Main Menu
- Search Menu
- List Menu
- Search Submenu
- Station selection screens

### 2. Text Input Prompts
**When you see a prompt asking you to type:**
- Type `0` or `back` - Go back to previous menu
- Type `00` or `main` - Return to Main Menu
- Press Enter (empty) - Go back to previous menu

**Examples:**
- "Type a name to search:"
- "Type a new list name:"
- "Type a list name to delete:"

---

## Quick Navigation Reference

| Context | Action | Navigation |
|---------|--------|------------|
| fzf Menu | Want to go back | Press ESC or select "0) Main Menu" |
| fzf Menu | Want main menu | Select "0) Main Menu" |
| Text Prompt | Want to go back | Type `0` or press Enter |
| Text Prompt | Want main menu | Type `00` |
| View Page | Continue | Press Enter |

---

## Examples

### Example 1: Searching for Stations
```
TERA - Search by Name

Type '0' to go back to Search Menu, '00' for Main Menu
Type a name to search: jazz      ← Type your search
  or
Type a name to search: 0         ← Go back to Search Menu
  or
Type a name to search: 00        ← Go to Main Menu
  or
Type a name to search: [Enter]   ← Go back to Search Menu
```

### Example 2: Creating a List
```
TERA - Create New List

My lists:
My-favorites
rock-stations

Type '0' to go back, '00' for main menu
Type a new list name: chill      ← Type new list name
  or
Type a new list name: 0          ← Go back to List Menu
  or
Type a new list name: 00         ← Go to Main Menu
```

### Example 3: Using fzf Menus
```
TERA SEARCH MENU

  0) Main Menu         ← Use arrows to select
  1) Tag              ← Press Enter when highlighted
  2) Name             ← Press ESC to cancel
  3) Language
  4) Country code
  5) State
  6) Advanced search
  7) Exit

Choose an option (arrow keys to navigate):
```

---

## Special Navigation Tips

### Quick Exit
From any menu:
- Select "Exit" option, or
- Press ESC repeatedly to go back, or
- Type `00` in text prompts to get to Main Menu, then select Exit

### Can't Find Something?
1. Type `00` to return to Main Menu
2. Navigate to the section you need
3. All major sections are accessible from Main Menu

### Made a Mistake?
- In fzf menus: Press ESC to cancel
- In text prompts: Type `0` to go back without changes
- Empty inputs are safe - they just return to previous menu

---

## Navigation Philosophy

TERA's dual navigation system gives you:
- **Speed**: Arrow keys and fzf for quick menu navigation
- **Flexibility**: Type `0`/`00` when entering data
- **Safety**: ESC and empty inputs won't cause problems
- **Consistency**: Same patterns throughout the app

**Remember:**
- Interactive menus → Use arrow keys and ESC
- Text prompts → Use 0, 00, or Enter
- When in doubt → ESC or 0 will take you back safely

---

## First-Time Users

When you first install TERA:
1. ✅ `~/.config/tera/favorite/` directory is auto-created
2. ✅ `My-favorites.json` is auto-created with examples
3. ✅ All navigation options work immediately
4. ✅ No setup required - just start using!

Try these first steps:
1. Open TERA - you'll see the Main Menu
2. Select "Quick Play Favorites" to try the included examples
3. Select "Search" to find your own stations
4. Experiment with navigation - you can't break anything!

---

## Troubleshooting Navigation

**Menu not responding to arrow keys?**
- Make sure fzf is installed: `which fzf`
- TERA checks dependencies on startup

**Typed `0` but nothing happened?**
- Make sure you pressed Enter after typing
- Check if you're in a text prompt (not an fzf menu)

**Want to exit immediately?**
- From Main Menu: Select "Exit"
- From anywhere: Press ESC until you reach Main Menu, then select "Exit"
- Or use Ctrl+C (emergency exit)

---

Happy listening! 🎵
