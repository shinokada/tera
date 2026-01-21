# Quick Guide: Updating Gists in TERA

## What Can You Update?

Currently, you can update:
- **Gist Description** - Change the title/description of your gist

## How to Update a Gist

### Step-by-Step

1. **Open TERA**
   ```bash
   ./tera
   ```

2. **Navigate to Gist Menu**
   - Select "Gist menu" from main menu

3. **Select "Update a gist"**
   ```text
   TERA GIST MENU
   
   You have 3 saved gist(s)
   
   0) Main Menu
   1) Create a gist
   2) My Gists
   3) Recover favorites from a gist
   4) Update a gist          ← Select this
   5) Delete a gist
   6) Exit
   ```

4. **Choose a Gist**
   ```text
   Your gists:
   Description                                        | Created
   --------------------------------------------------------------------------------
    1) Terminal radio favorite lists                 | 2026-01-19 10:30
    2) My awesome radio stations                      | 2026-01-18 15:20
    3) Rock classics collection                       | 2026-01-17 09:45
   
   Enter gist number to update: 1
   ```

5. **Enter New Description**
   ```text
   Current description: Terminal radio favorite lists
   
   Enter new description: My Updated Radio Favorites 2026
   ```

6. **Confirmation**
   ```text
   Updating gist on GitHub...
   ✓ Gist updated successfully!
   
   New description: My Updated Radio Favorites 2026
   
   Press Enter to continue...
   ```

## Quick Tips

### Cancel Anytime
- Press `Enter` without typing = Cancel
- Type `0` = Return to Gist Menu

### Best Practices
1. **Be Descriptive** - Use clear, meaningful descriptions
2. **Add Dates** - Include year/date in description for versioning
3. **Use Tags** - Add keywords like "jazz", "rock", "favorites"

### Common Update Scenarios

**Scenario 1: Add Date Version**
```
Old: Terminal radio favorites
New: Terminal radio favorites - Jan 2026
```

**Scenario 2: Add Content Description**
```
Old: My lists
New: Jazz & Classical stations - curated 2026
```

**Scenario 3: Project/Context Update**
```
Old: Radio stations
New: [TERA] Production radio stations for work
```

## What Gets Updated

| Component                  | Updated? | Where?                            |
| -------------------------- | -------- | --------------------------------- |
| Gist description on GitHub | ✅ Yes    | GitHub.com                        |
| Local metadata             | ✅ Yes    | ~/.config/tera/gist_metadata.json |
| Gist files (JSON content)  | ❌ No     | Not yet implemented               |
| Gist URL                   | ❌ No     | URLs are permanent                |
| Created date               | ❌ No     | Original creation date preserved  |

## Error Messages

### "Invalid choice"
**Problem:** Entered number out of range
**Solution:** Enter a number from the displayed list (1, 2, 3, etc.)

### "Failed to update gist on GitHub"
**Problem:** API request failed
**Possible causes:**
- Invalid GitHub token
- Token expired
- No internet connection
- Gist was deleted on GitHub

**Solution:**
1. Verify internet connection
2. Try listing your gists first (`My Gists`)

## Technical Details

### API Endpoint Used
```
PATCH https://api.github.com/gists/{gist_id}
```

### What Happens Behind the Scenes
1. Validates GitHub token exists
2. Displays your saved gists from local metadata
3. You select a gist
4. Sends PATCH request to GitHub API
5. Updates local metadata file
6. Shows confirmation

### Files Modified
- **Remote:** GitHub gist (via API)
- **Local:** `~/.config/tera/gist_metadata.json`

## Keyboard Shortcuts

| Key             | Action                |
| --------------- | --------------------- |
| `0`             | Back to Gist Menu     |
| `Enter` (empty) | Cancel operation      |
| `1-9`           | Select gist by number |

## Related Commands

```bash
# View all gists
Gist Menu → My Gists

# Create new gist
Gist Menu → Create a gist

# Delete gist
Gist Menu → Delete a gist

# Recover from gist
Gist Menu → Recover favorites from a gist
```

## Examples

### Example 1: Simple Update
```text
Before: Radio stations
After:  Radio stations - favorites 2026
```

### Example 2: Categorize
```text
Before: My lists
After:  [Jazz] Favorite jazz stations
```

### Example 3: Version Control
```text
Before: Station collection
After:  Station collection v2.1 - Updated Jan 2026
```

## FAQ

**Q: Can I update the gist files themselves?**
A: Not yet. Currently only description updates are supported. File updates coming in future version.

**Q: What if I made a typo in the new description?**
A: Just run update again! You can update as many times as you want.

**Q: Will this affect my GitHub URL?**
A: No! The gist URL remains the same. Only the description changes.

**Q: Can I update gists I didn't create in TERA?**
A: If they're in your local metadata, yes. If not, you'll need to add them via "Recover favorites from a gist" first.

**Q: Is the update instant?**
A: Yes! Changes appear immediately on GitHub and in your local metadata.

## Troubleshooting

### Update doesn't appear on GitHub
1. Wait 10-30 seconds and refresh
2. Check internet connection
3. Verify token has `gist` scope
4. Check GitHub status page

### Can't find my gist in the list
1. Go to "My Gists" first to see all saved gists
2. If missing, use "Recover favorites from a gist"
3. Check `~/.config/tera/gist_metadata.json`

### "Update cancelled" but I didn't cancel
- This happens when you press Enter without typing anything
- This is intentional - empty input = cancel

---

**Need more help?** Check the full documentation in `docs/GIST_CRUD_GUIDE.md`
