# Gist Management Guide

Managing your favorite radio station lists with GitHub Gists.

---

## Setup

### GitHub Token Required

Create `.env` in TERA directory:

```bash
cp .env.example .env
# Add your token: GITHUB_TOKEN="your_token_here"
```

**Get token:** https://github.com/settings/tokens (select 'gist' scope)

---

## Gist Menu

Access: Main Menu → `6) Gist`

```
TERA GIST MENU

You have 3 saved gist(s)

1) Create a gist
2) My Gists
3) Recover favorites from a gist
4) Delete a gist
5) Exit
```

---

## Features

### 1) Create a Gist

Backs up all your lists to a private GitHub Gist.

**Process:**
- Uploads all `.json` files from favorites folder
- Saves metadata locally (ID, URL, timestamp)
- Opens in browser automatically

**Result:**
```
✓ Successfully created a secret Gist!
Gist URL: https://gist.github.com/username/abc123
```

---

### 2) My Gists

Lists all your saved gists with creation dates.

**Display:**
```
 1) Terminal radio favorite lists | 2026-01-19 10:30
 2) Terminal radio favorite lists | 2026-01-18 15:45
 3) Terminal radio favorite lists | 2026-01-17 09:20
```

**Actions:**
- Type number → Opens gist in browser
- Type `0` or Enter → Back to menu

---

### 3) Recover Favorites

Import lists from any gist.

**Two Options:**

**A) Select from saved gists:**
```
Your saved gists:
 1) Gist from Jan 19 (2026-01-19 10:30)
 2) Gist from Jan 18 (2026-01-18 15:45)

Enter gist number or URL: 1
```

**B) Enter any URL:**
```
Enter gist number or URL: https://gist.github.com/user/xyz789
```

**Result:** All `.json` files downloaded to your favorites folder

---

### 4) Delete a Gist

Removes gist from GitHub and local metadata.

**Process:**
1. Select gist number from list
2. Type `yes` to confirm
3. Deleted from GitHub
4. Removed from local tracking

**Important:** Your local lists are NOT deleted

---

## Use Cases

### Backup
```
After adding stations → Create a gist
```

### Share
```
My Gists → Type number → Copy URL → Share
```

### Sync Devices
```
Device A: Create a gist
Device B: Recover from that gist
```

### Cleanup
```
Delete a gist → Remove old backups
```

---

## Navigation

All screens support:
- `0` = Back to previous menu
- `00` = Main Menu
- ESC = Cancel (in menus)
- Empty + Enter = Back

---

## File Locations

```
~/.config/tera/
├── gist_metadata.json     # Your gist list
└── favorite/
    └── *.json            # Your lists
```

---

## Troubleshooting

### "GitHub token not found"
→ Create `.env` file with your token

### "Failed to create gist"
→ Check token is valid and has 'gist' scope

### "Failed to clone gist"
→ Check URL is correct and internet connection

### No gists in "My Gists"
→ Only shows gists created after this update

---

## Privacy & Security

**Gists are private by default:**
- Not searchable on GitHub
- Only accessible with direct URL
- Share URL only with trusted people

**Token Security:**
- Never share your token
- Use minimal scopes (only 'gist')
- Rotate every 6-12 months

---

## Advanced

### Metadata Format

Stored at `~/.config/tera/gist_metadata.json`:

```json
[
  {
    "id": "abc123",
    "url": "https://gist.github.com/user/abc123",
    "description": "Terminal radio favorite lists",
    "created_at": "2026-01-19T10:30:00Z",
    "updated_at": "2026-01-19T10:30:00Z"
  }
]
```

### Manual Metadata Cleanup

If metadata becomes corrupted:

```bash
# Backup first
cp ~/.config/tera/gist_metadata.json ~/.config/tera/gist_metadata.json.backup

# Reset
echo "[]" > ~/.config/tera/gist_metadata.json
```

---

## FAQ

**Q: Can I update an existing gist?**  
A: Not yet. Create a new gist and delete the old one.

**Q: Will old gists still work?**  
A: Yes! Enter the URL manually in "Recover favorites"

**Q: Can friends without GitHub use my gists?**  
A: Yes, gists are viewable by anyone with the URL

**Q: What happens if I delete a gist?**  
A: Only the GitHub gist is deleted. Your local lists are safe.

---

**Quick Reference:** See [GIST_QUICK_REFERENCE.md](GIST_QUICK_REFERENCE.md)  
**Setup Help:** See [GIST_SETUP.md](GIST_SETUP.md)  
**Navigation:** See [NAVIGATION_GUIDE.md](NAVIGATION_GUIDE.md)
