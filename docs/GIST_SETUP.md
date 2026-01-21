# Setting Up GitHub Gist Integration

TERA uses GitHub Gists to backup and restore your favorite radio station lists. To use this feature, you need to set up a GitHub Personal Access Token.

## Token Management

TERA provides a complete Token Management system with Create, Read, Update, Delete (CRUD) operations.

### Accessing Token Management (Option 1 - Recommended)

1. **Run TERA:**
   ```bash
   tera
   ```

2. **Navigate to Gist Menu:**
   - Press `6) Gist` from the main menu

3. **Select Token Management:**
   - Choose `6) Token Management (CRUD)` from the Gist Menu
   - You'll see options:
     - `1) Setup Token` - Initial token configuration
     - `2) View Token` - Check current token status
     - `3) Validate Token` - Test token with GitHub API
     - `4) Delete Token` - Remove stored token
     - `0) Back to Gist Menu` - Return to previous menu

### Accessing Token Management (Option 2 - Direct Menu)

You can also access token management from the main menu if available in your version.

### Token Setup Workflow

**First Time Setup:**

1. **Run TERA:**
   ```bash
   tera
   ```

2. **Go to Gist Menu → Token Management (CRUD) → Setup Token**

3. **Create a GitHub Personal Access Token:**
   - Go to [GitHub Token Settings](https://github.com/settings/tokens)
   - Click "Generate new token (classic)"
   - Give it a name like "TERA Gist Access"
   - Select **only** the `gist` scope
   - Click "Generate token"
   - Copy the token (you won't be able to see it again!)

4. **Paste your token into TERA:**
   - TERA will securely store it in: `~/.config/tera/tokens/github_token`
   - File permissions are set to `600` (owner read/write only)
   - Token is automatically loaded when needed

**Update Existing Token:**

1. **Go to Gist Menu → Token Management (CRUD) → Setup Token**
2. **Enter your new token**
3. **Old token is automatically replaced**

**View Current Token:**

1. **Go to Gist Menu → Token Management (CRUD) → View Token**
2. **Token is displayed in masked format:** `ghp_...xyz`
3. **Full token is never shown in plain text**

**Test Token:**

1. **Go to Gist Menu → Token Management (CRUD) → Validate Token**
2. **TERA connects to GitHub API to verify the token works**
3. **You'll see a success or error message**

**Remove Token:**

1. **Go to Gist Menu → Token Management (CRUD) → Delete Token**
2. **Confirm deletion**
3. **Token file is securely removed**

## Token Storage Security

Your GitHub token is stored securely:

- **Location:** `~/.config/tera/tokens/github_token`
- **File Permissions:** 600 (owner read/write only)
- **Directory Permissions:** 700 (owner full control)
- **Display Format:** Masked as `ghp_...xyz` in menus
- **Validation:** API test before accepting token
- **Encryption:** Stored as plain text in secure location

## Token Management Features

Your GitHub token is used automatically by TERA to:

### Create a Gist
- Upload your favorite station lists
- Creates a private gist (only you can see it)
- Stores backup of your lists

### Recover Favorites from a Gist
- Download previously backed up lists
- Search your gists or paste a gist URL
- Restore your collections

## Security

✓ **Secure Storage:**
- Tokens are stored in `~/.config/tera/tokens/` with `600` permissions (user only)
- File is not tracked in git (already in `.gitignore`)
- Only you can read/write your token file

✓ **Minimal Permissions:**
- Token has only `gist` scope
- Cannot access private repos or other GitHub features
- Limited to Gist operations only

⚠️ **Best Practices:**
- Never share your GitHub token with anyone
- Token shown only as masked display (`ghp_...xyz`)
- You can revoke/rotate tokens anytime on GitHub
- If compromised, delete from TERA and revoke on GitHub

## Token File Location

```text
~/.config/tera/tokens/github_token
```

File structure:
- **Location:** User's home directory `~/.config/tera/`
- **Permissions:** `600` (owner read/write only)
- **Format:** Plain text containing just the token
- **Works with all installation methods:** brew, .deb, /awesome, or source

## Troubleshooting

### "Token validation failed"
- Make sure token is correctly copied (no extra spaces)
- Verify the token has `gist` scope enabled
- Check if token is expired - regenerate from GitHub
- Ensure you have internet connection

### "Authentication errors when creating/updating gists"
- Token may have expired - validate and update it
- Check token permissions - must have `gist` scope
- Verify file permissions: `ls -la ~/.config/tera/tokens/github_token` should show `600`

### "Token not found when running installed package"
- For brew/deb/awesome installations, token must be set up via TERA menu
- Run TERA and go to **Gist → Token Management → Setup Token**
- Token will be stored in `~/.config/tera/tokens/`

## Using Gist Features

Once your token is configured, you can:

1. **Create a gist**: Uploads your station lists to a secret GitHub Gist
2. **My Gists**: View and manage your created gists
3. **Recover from gist**: Download and restore lists from a Gist URL
4. **Update a gist**: Change gist descriptions
5. **Delete a gist**: Remove gists from GitHub

## Token Rotation

To safely update your token:

1. Go to **Token Management** → **Setup/Change Token**
2. Generate a new token on GitHub
3. Paste the new token in TERA
4. TERA validates and saves it
5. Revoke the old token on GitHub for security
