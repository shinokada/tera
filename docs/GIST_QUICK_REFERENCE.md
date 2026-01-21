# TERA Gist Features - Quick Reference

## Gist Menu Access
```text
Main Menu → 6) Gist
```

---

## Menu Options

### 1) Create a Gist
**What it does:** Backs up all your favorite lists to GitHub  
**Result:** Private gist + saved locally

**Steps:**
1. Select `1) Create a gist`
2. Wait for upload
3. Gist opens in browser automatically
4. Metadata saved to `~/.config/tera/gist_metadata.json`

---

### 2) My Gists
**What it does:** Shows all your saved gists  
**Result:** List with creation dates

**Actions:**
- Type number → Opens gist in browser
- Type `0` → Back to menu
- Press Enter → Back to menu

**Example:**
```text
 1) Terminal radio favorite lists | 2026-01-19 10:30
 2) Terminal radio favorite lists | 2026-01-18 15:45
```

---

### 3) Recover Favorites
**What it does:** Import lists from any gist  
**Two ways:**

#### A) Select from your saved gists
```text
Your saved gists:
 1) Gist from Jan 19
 2) Gist from Jan 18
 
Enter: 1          ← Type the number
```

#### B) Enter any gist URL
```text
Enter: https://gist.github.com/user/abc123
```

**Result:** All `.json` files downloaded to your favorites folder

---

### 4) Delete a Gist
**What it does:** Removes gist from GitHub and local list  
**Important:** Requires confirmation!

**Steps:**
1. Select gist number
2. Type `yes` to confirm
3. Gist deleted from GitHub
4. Removed from local list

**Note:** Your local lists are NOT deleted

---

## Quick Commands

| Want to...           | Do this                 |
| -------------------- | ----------------------- |
| Backup lists         | Create a gist           |
| See all backups      | My Gists                |
| Restore from backup  | Recover (select number) |
| Import from friend   | Recover (enter URL)     |
| Remove old backup    | Delete a gist           |
| Open gist in browser | My Gists → type number  |

---

## Navigation

**All screens support:**
- `0` = Go back to previous menu
- `00` = Return to Main Menu
- `ESC` = Cancel (in menus)
- Empty + Enter = Go back

---

## File Locations

```text
~/.config/tera/
├── gist_metadata.json     ← Your gist list
├── favorite/
│   ├── My-favorites.json  ← Your lists
│   ├── jazz.json
│   └── rock.json
```

---

## Common Workflows

### Backup Your Lists
```text
Main Menu → Gist → Create a gist
```

### Share with a Friend
```text
1. Main Menu → Gist → My Gists
2. Type number to open in browser
3. Copy URL from browser
4. Share URL with friend
5. Friend uses: Recover → paste URL
```

### Sync Two Computers
```text
Computer A: Create a gist
Computer B: Recover from that gist URL
```

### Clean Up Old Gists
```text
Main Menu → Gist → My Gists
(note which ones to delete)
Main Menu → Gist → Delete a gist
```

---

## Tips

💡 **Quick Recovery:** Just type the number instead of copy/pasting URLs

💡 **Multiple Backups:** Create gists at different times, keep versions

💡 **Safety:** Gists are private by default, only visible with URL

💡 **Local Lists Safe:** Deleting a gist doesn't delete your local lists

---

## Troubleshooting

### "Failed to create gist"
→ Check your GitHub token is valid and has 'gist' scope

### "Failed to clone gist"
→ Check the URL is correct and you have internet

### No gists showing in "My Gists"
→ Gists created before this update won't show (still work with URLs)

---

## Need Help?

See full documentation:
- `docs/GIST_CRUD_GUIDE.md` - Complete guide
- `docs/GIST_SETUP.md` - GitHub token setup
- `docs/NAVIGATION_GUIDE.md` - Navigation help

---

**Quick Test:**
1. Create a gist
2. Check "My Gists" (should see it)
3. Try selecting it in "Recover" 
4. Delete it (if you want)

That's it! 🎵
