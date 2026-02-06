# Gist Management Guide

Complete guide to managing your favorite radio station lists with GitHub Gists.

---

## Accessing Gist Management

**Main Menu → 5) Gist Management**

Quick select: Press `5` from the main menu

---

## Gist Management Menu

```
TERA Gist Menu

1) Create a gist
2) My Gists
3) Recover favorites
4) Update a gist
5) Delete a gist
6) Token Management
```

**Navigation:**
- `↑↓` or `jk` - Navigate menu items
- `Enter` - Select highlighted item  
- `1-6` - Quick select by number
- `Esc` - Back to Main Menu
- `Ctrl+C` - Quit application

---

## Features Overview

### 1) Create a Gist

Upload all your favorite lists to GitHub as a gist.

**Process:**

1. **Choose visibility:**
   - `1) Secret gist` - Only you can see it (recommended)
   - `2) Public gist` - Anyone with the link can see it

2. **Enter gist name/description:**
   - Default: `TERA Radio Favorites - 2026-01-29 15:04:05`
   - Press Enter to use default
   - Or type custom name and press Enter

3. **TERA uploads your files:**
   - All `.json` files from `favorites/` in your config directory are uploaded
   - Gist is created on GitHub
   - Metadata saved locally to `gist_metadata.json` in your config directory

**Result:**
```
✓ Gist created (secret)! https://gist.github.com/username/abc123def456
```

**What gets uploaded:**
- All your favorite list files (e.g., `My-favorites.json`, `Rock.json`, etc.)
- Only `.json` files from your favorites folder
- Each file becomes a file in the gist

---

### 2) My Gists

View and manage all your saved gists.

**Display format:**
```
TERA Radio Favorites - 2026-01-29 15:04:05
2026-01-29 15:04

Terminal radio favorite lists
2026-01-28 09:30

My Favorite Stations
2026-01-27 14:15
```

**Actions:**
- Navigate with `↑↓` or `jk`
- Press `Enter` to open selected gist in your browser
- Press `Esc` to go back to Gist Menu

**Note:** Only gists created with this version of TERA appear here. Older gists can still be recovered using their URL (see "Recover favorites").

---

### 3) Recover Favorites

Download and restore favorite lists from any gist.

**Two recovery options:**

**Option A - From your saved gists:**
```
Select a Gist

TERA Radio Favorites - 2026-01-29 15:04:05
2026-01-29 15:04

Terminal radio favorite lists  
2026-01-28 09:30

[Use ↑↓ to select, Enter to recover]
```

**Option B - From any gist URL:**
```
Select a Gist

[Paste any gist URL and press Enter]
https://gist.github.com/username/xyz789
```

**What happens:**
1. TERA creates backup of existing files in `favorites/.backup/` in your config directory
2. Backup files named: `filename.20260129-150405.bak`
3. Downloads all files from the gist
4. Overwrites existing files in your favorites folder
5. Your old versions are safe in `.backup/` folder

**Success message:**
```
✓ Favorites restored successfully! (backups saved in .backup folder)
```

**Important notes:**
- Only `.json` files are restored
- Filenames are validated to prevent directory traversal attacks
- Old backups are kept (not overwritten) - manage them manually if needed
- You can recover from any public or secret gist you have access to

---

### 4) Update a Gist

Change the description/name of an existing gist.

**Process:**

1. **Select gist to update:**
   ```
   TERA Radio Favorites - 2026-01-29 15:04:05
   > Terminal radio favorite lists
   My Favorite Stations
   ```

2. **Enter new description:**
   ```
   Current: Terminal radio favorite lists
   
   New Description:
   [Type new description here]
   ```

3. **Press Enter to save**

**Result:**
```
✓ Gist updated!
```

**Notes:**
- Only description is updated (files remain unchanged)
- To update files, create a new gist
- Updates both GitHub and local metadata
- Changes are visible immediately in "My Gists"

---

### 5) Delete a Gist

Permanently remove a gist from GitHub.

**Process:**

1. **Select gist to delete:**
   ```
   TERA Radio Favorites - 2026-01-29 15:04:05
   Terminal radio favorite lists
   > Old backup from last week
   ```

2. **Confirm deletion:**
   ```
   Are you sure you want to delete this gist?
   Old backup from last week
   
   Type 'yes' to confirm:
   [Type here]
   ```

3. **Type `yes` and press Enter**

**Result:**
```
✓ Gist deleted!
```

**Important:**
- Gist is permanently deleted from GitHub (cannot be undone)
- Local metadata is removed
- Your local favorite files are NOT affected
- If you need the files, recover from gist before deleting
- Typing anything other than 'yes' cancels the deletion

---

### 6) Token Management

Manage your GitHub Personal Access Token.

See [GIST_SETUP.md](GIST_SETUP.md) for detailed token management documentation.

**Quick access:**
- Setup token: `6` → `1`
- View token: `6` → `2`
- Validate token: `6` → `3`
- Delete token: `6` → `4`

---

## Common Use Cases

### Backup Your Favorites

```
Main Menu → 5) Gist Management → 1) Create a gist
[Choose secret gist → Enter name → Done]
```

**When to backup:**
- After adding many new stations
- Before major changes to your lists
- Weekly/monthly routine backup
- Before reinstalling your system

---

### Share Favorites with Friend

1. Create a **public gist** (or secret if sharing URL privately)
2. Go to **2) My Gists**
3. Select the gist (opens in browser)
4. Copy and share the URL
5. Friend recovers from your gist URL in their TERA

**Privacy note:** Anyone with a secret gist URL can view it. Only create public gists if you're okay with anyone finding them.

---

### Sync Between Devices

**Device A (source):**
```
Create a gist → Note the URL
```

**Device B (destination):**
```
Recover favorites → Paste gist URL
```

**Tips:**
- Create a new backup before syncing
- Use the same gist URL on all devices
- Update the gist when you make changes
- Remember: last recovery overwrites (backups are saved)

---

### Restore After Reinstall

1. Install TERA on new system
2. Setup GitHub token (if needed)
3. Go to Recover favorites
4. Paste your gist URL
5. All favorites restored

---

### Regular Cleanup

```
My Gists → [Review old backups] → Delete outdated ones
```

**Good practice:**
- Keep last 2-3 backups
- Delete very old backups
- Update gist description with version/date info

---

### Storage Location (Config Directory)
```
tera/
├── gist_metadata.json         # List of your saved gists
├── tokens/
│   └── github_token          # Your GitHub token
└── favorites/
    ├── .backup/              # Automatic backups
    │   ├── My-favorites.json.20260129-150405.bak
    │   └── Rock.json.20260129-150405.bak
    ├── My-favorites.json     # Your lists
    ├── Rock.json
    └── Jazz.json
```

---

## Understanding Gist Metadata

**Location:** `~/.config/tera/gist_metadata.json`

**Format:**
```json
[
  {
    "id": "abc123def456",
    "url": "https://gist.github.com/username/abc123def456",
    "description": "TERA Radio Favorites - 2026-01-29 15:04:05",
    "created_at": "2026-01-29T15:04:05Z",
    "updated_at": "2026-01-29T15:04:05Z"
  }
]
```

**What it's for:**
- Tracks gists created by TERA
- Powers "My Gists" feature
- Stored locally (not uploaded to GitHub)
- Updated when you create/update/delete gists
- Safe to delete if it becomes corrupted (you'll lose "My Gists" list but can still recover by URL)

---

## Troubleshooting

### "No token configured!"

**Solution:** Setup your GitHub token first
```
Gist Menu → 6) Token Management → 1) Setup/Change Token
```

See [GIST_SETUP.md](GIST_SETUP.md) for detailed token setup instructions.

---

### "Failed to create gist" / "Authentication error"

**Possible causes:**
- Token is invalid or expired
- Token missing `gist` scope
- No internet connection

**Solution:**
1. Validate your token: `6) Token Management → 3) Validate Token`
2. If validation fails, setup a new token
3. Check your internet connection

---

### "No gists found - create a gist first"

**This appears when:**
- You haven't created any gists with this version of TERA yet
- `gist_metadata.json` is empty or missing

**Solution:** Create your first gist, or recover from an existing gist URL

---

### "No favorite lists found"

**This means:**
- Your `favorites` folder is empty
- No `.json` files to upload

**Solution:**
1. Add stations to your favorites first
2. Check that files exist: `ls ~/.config/tera/favorites/*.json`
3. Or recover favorites from an existing gist

---

### Can't recover from old gist

**If gist doesn't appear in "My Gists":**
- Old gists (created before metadata feature) won't appear
- Solution: Paste the gist URL directly in "Recover favorites"

**If gist URL doesn't work:**
- Check URL is correct (copy from GitHub directly)
- Verify you have internet connection
- For secret gists, make sure URL is complete
- Check if gist was deleted on GitHub

---

### Backup files accumulating

**Location:** `~/.config/tera/favorites/.backup/`

**These are safe to delete manually:**
```bash
# View backups
ls -lh ~/.config/tera/favorites/.backup/

# Delete old backups (older than 30 days)
find ~/.config/tera/favorites/.backup/ -name "*.bak" -mtime +30 -delete

# Delete all backups
rm ~/.config/tera/favorites/.backup/*.bak
```

**TERA does not auto-delete backups** - manage them manually when needed.

---

### Metadata file corrupted

**Symptoms:**
- "My Gists" shows errors
- Can't list gists

**Solution - Reset metadata:**
```bash
# Backup first (optional)
cp ~/.config/tera/gist_metadata.json ~/.config/tera/gist_metadata.json.backup

# Reset to empty
echo "[]" > ~/.config/tera/gist_metadata.json
```

**Note:** This only affects "My Gists" display. Your actual gists on GitHub are safe, and you can still recover from them using their URLs.

---

## Security & Privacy

### Secret vs Public Gists

**Secret gists:**
- Not searchable on GitHub
- Only accessible via direct URL
- Still visible to anyone with the URL
- Recommended for personal backups

**Public gists:**
- Searchable on GitHub
- Listed on your GitHub profile
- Anyone can find and view them
- Good for sharing with community

**Important:** Both types can be viewed by anyone with the URL. Secret just means "not listed publicly."

---

### Token Security

**Your token allows:**
- Creating gists under your GitHub account
- Viewing your gists
- Updating your gists
- Deleting your gists

**Your token does NOT allow:**
- Access to your repositories
- Access to other GitHub features
- Anything beyond gist operations (we use minimal `gist` scope)

**Best practices:**
- Never share your token
- Rotate token every 6-12 months
- Revoke immediately if compromised
- Use unique tokens per application

See [GIST_SETUP.md](GIST_SETUP.md) for detailed security information.

---

## FAQ

**Q: Can I update files in an existing gist?**  
A: No. Create a new gist with updated files, then delete the old one if needed.

**Q: Will deleting a gist delete my local files?**  
A: No. Deleting a gist only removes it from GitHub. Your local favorites are safe.

**Q: Can I recover from someone else's gist?**  
A: Yes! If they share the URL, you can recover any public or secret gist.

**Q: How many gists can I create?**  
A: GitHub has no documented limit. TERA has no limit. Create as many as you need.

**Q: What happens if I lose my gist URL?**  
A: If it's in "My Gists," you can open it there. Otherwise, check your GitHub profile → Gists section on github.com.

**Q: Can I organize gists into folders?**  
A: No, GitHub Gists don't support folders. Use descriptive names/descriptions instead.

**Q: Why don't my old gists show in "My Gists"?**  
A: The metadata tracking feature is new. Old gists can still be recovered using their URL.

**Q: Can I use gists without a GitHub account?**  
A: No, you need a GitHub account and personal access token to create gists. However, you can recover/view public gists from URLs without authentication.

---

## Quick Reference

| Task | Steps |
|------|-------|
| **Backup** | `5` → `1` → Choose visibility → Enter → Done |
| **Restore** | `5` → `3` → Select gist or paste URL → Done |
| **View** | `5` → `2` → Select gist → Opens in browser |
| **Rename** | `5` → `4` → Select gist → Enter new name → Done |
| **Delete** | `5` → `5` → Select gist → Type 'yes' → Done |
| **Token** | `5` → `6` → [Setup/View/Validate/Delete] |

---

**Related Documentation:**
- [GIST_SETUP.md](GIST_SETUP.md) - Token setup and security
- [NAVIGATION_GUIDE.md](NAVIGATION_GUIDE.md) - Keyboard navigation
- [README.md](README.md) - Getting started with TERA
