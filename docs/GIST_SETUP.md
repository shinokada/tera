# Setting Up GitHub Gist Integration

TERA uses GitHub Gists to backup and restore your favorite radio station lists. To use this feature, you need to set up a GitHub Personal Access Token.

---

## Quick Start

1. **Launch TERA** and select **5) Gist Management** from the main menu
2. **Select 6) Token Management** from the Gist Menu
3. **Choose 1) Setup/Change Token**
4. Create a GitHub token (see below) and paste it into TERA
5. Done! You can now use all Gist features

---

## Creating Your GitHub Token

1. Go to [GitHub Token Settings](https://github.com/settings/tokens)
2. Click **"Generate new token (classic)"**
3. Give it a descriptive name like **"TERA Gist Access"**
4. Select **only** the `gist` scope (do not select any other permissions)
5. Click **"Generate token"**
6. **Copy the token immediately** (you won't be able to see it again!)

---

## Token Management Menu

Access path: **Main Menu → 5) Gist Management → 6) Token Management**

The Token Management menu provides four options:

### 1) Setup/Change Token

**First-time setup:**
- Paste your GitHub Personal Access Token
- TERA validates the token with GitHub API
- Token is saved securely to `~/.config/tera/tokens/github_token`
- File permissions automatically set to `600` (owner read/write only)

**Updating existing token:**
- Enter your new token
- Old token is automatically replaced
- New token is validated before saving

### 2) View Current Token

- Displays your token in masked format: `ghp_...xyz`
- Shows first 11 and last 4 characters only
- Full token is never displayed for security
- If no token exists, displays "No token configured"

### 3) Validate Token

- Tests your token against the GitHub API
- Verifies token has correct permissions
- Displays your GitHub username if valid
- Shows error message if token is invalid or expired

### 4) Delete Token

- Removes your stored token from the system
- Requires typing 'yes' to confirm
- Token file is securely deleted
- You'll need to setup a new token to use Gist features again

---

## Token Storage & Security

### Storage Location
```
~/.config/tera/
├── tokens/
│   └── github_token      # Your GitHub token (permissions: 600)
└── gist_metadata.json    # Your gist list (permissions: 644)
```

### Security Features

**File Permissions:**
- Token directory: `700` (owner full control only)
- Token file: `600` (owner read/write only)
- Nobody else on your system can read your token

**Display Protection:**
- Token displayed as `ghp_...xyz` in all menus
- Full token never shown in plain text
- Protected against shoulder surfing

**Validation:**
- Token verified with GitHub API before saving
- Invalid tokens are rejected immediately
- You'll see your GitHub username on successful validation

**Minimal Permissions:**
- Token has only `gist` scope
- Cannot access your repositories
- Cannot access other GitHub features
- Limited to Gist operations only

---

## Navigation Tips

**Keyboard shortcuts:**
- `↑↓` or `jk` - Navigate menu items
- `Enter` - Select highlighted item
- `1-4` - Quick select by number
- `Esc` - Go back to previous menu
- `Ctrl+C` - Quit application

**Menu hierarchy:**
```
Main Menu
  └─ 5) Gist Management
       └─ 6) Token Management
            ├─ 1) Setup/Change Token
            ├─ 2) View Current Token
            ├─ 3) Validate Token
            └─ 4) Delete Token
```

---

## Using Gist Features

Once your token is configured, you can:

- **Create a gist** - Upload your favorites to GitHub (choose secret or public)
- **My Gists** - View and open your saved gists in browser
- **Recover favorites** - Download and restore lists from any gist
- **Update a gist** - Change gist descriptions
- **Delete a gist** - Remove gists from GitHub

All these features are available in the Gist Management menu.

---

## Troubleshooting

### "No token configured!" when trying to create a gist
**Solution:** Go to Token Management and setup your token first

### "Invalid token" error during setup
**Possible causes:**
- Token copied incorrectly (check for extra spaces)
- Token missing `gist` scope (recreate with correct permission)
- Token already revoked on GitHub (create a new one)

### "Token validation failed" error
**Possible causes:**
- No internet connection
- GitHub API is down (rare)
- Token has been revoked on GitHub
- Token expired (GitHub tokens don't expire by default, but can be set to)

### "Authentication errors when creating/updating gists"
**Solution:**
1. Validate your token: Token Management → 3) Validate Token
2. If validation fails, delete and recreate your token
3. Make sure token has `gist` scope on GitHub

### Can't find the token file
The token is stored at: `~/.config/tera/tokens/github_token`

Check it exists:
```bash
ls -la ~/.config/tera/tokens/
```

You should see:
```
-rw------- 1 user user    40 Jan 29 10:30 github_token
```

---

## Token Rotation (Best Practice)

To safely rotate your token:

1. **Generate a new token** on GitHub (keep the old one active)
2. **Go to Token Management → Setup/Change Token** in TERA
3. **Paste the new token** (TERA validates and saves it)
4. **Test the new token** using Validate Token option
5. **Revoke the old token** on GitHub (now that new one works)

This ensures you always have a working token during the transition.

---

## Privacy & Security Best Practices

✅ **DO:**
- Keep your token secret (treat it like a password)
- Use minimal permissions (only `gist` scope)
- Rotate tokens every 6-12 months
- Delete tokens you're not using
- Review active tokens on GitHub regularly

❌ **DON'T:**
- Share your token with anyone
- Add extra scopes/permissions "just in case"
- Store token in public repos or shared files
- Use the same token for multiple applications

---

## Installation-Specific Notes

**For all installation methods** (homebrew, .deb package, manual install):
- Token must be set up via TERA's menu interface
- Token location is always `~/.config/tera/tokens/github_token`
- Works identically regardless of how TERA was installed

---

## Next Steps

Once your token is configured:

1. **Create your first gist** - Upload your current favorites
2. **Try recovering** - Download your favorites back to test it works
3. **Read [GIST_CRUD_GUIDE.md](GIST_CRUD_GUIDE.md)** - Learn all Gist features
4. **Check [NAVIGATION_GUIDE.md](NAVIGATION_GUIDE.md)** - Master TERA navigation
