# GitHub Token Management Guide

## Overview

TERA now provides a complete token management system with a user-friendly menu interface. 

## Token Storage System

### Storage Location

Tokens are stored in a secure directory:
```text
~/.config/tera/tokens/github_token
```

### Security Features

- **File Permissions:** `600` (owner read/write only)
- **Location:** Hidden in user's home config directory
- **Format:** Plain text (no encryption - relies on file permissions)
- **Works with:** All installation methods (brew, .deb, /awesome, source)

### File Structure

```text
~/.config/tera/
├── favorite/           # Your favorite lists
├── tokens/             # Token storage
│   └── github_token    # Your GitHub token (600 permissions)
└── gist_metadata.json  # Gist metadata
```

## Using Token Management

### Accessing Token Management

1. Launch TERA:
   ```bash
   tera
   ```

2. Select **`6) Gist`** from main menu

3. Select **`1) Token Management`**

### Menu Options

#### 1) Setup/Change Token
**Purpose:** Add or update your GitHub token

**Process:**
1. Opens setup wizard
2. Prompts for token (input is hidden for security)
3. Validates token format
4. Tests with GitHub API
5. Shows GitHub username upon success
6. Saves token securely

**When to use:**
- First-time setup
- Changing to different GitHub account
- Updating expired token

#### 2) View Current Token
**Purpose:** Check your current token status

**Shows:**
- Masked token: `ghp_VaVnzkTqrBY...UPxM`
- Associated GitHub username
- Validation status (valid/expired)

**When to use:**
- Verify which token is active
- Check if token is still working
- Before rotating tokens

#### 3) Validate Token
**Purpose:** Test if your token works

**Tests:**
- Makes API call to GitHub
- Checks token validity
- Confirms gist scope
- Shows success/failure reason

**When to use:**
- Troubleshoot authentication errors
- Verify token after update
- Check if token expired or revoked

#### 4) Delete Token
**Purpose:** Remove your stored token

**Process:**
1. Shows current masked token
2. Requires `yes` confirmation
3. Permanently deletes token file
4. Unsets GITHUB_TOKEN environment variable

**When to use:**
- Switching to different machine
- Security concern (token compromised)
- Uninstalling TERA
- Switching GitHub accounts

## Token Workflow Examples

### First-Time Setup

```text
TERA MAIN MENU
  6) Gist
    1) Token Management
      1) Setup/Change Token
        → Prompts for GitHub token
        → Validates format
        → Tests with API
        → Shows: "✓ Token is valid! GitHub user: username"
        → Saves to ~/.config/tera/tokens/github_token
```

### Validating After Update

```text
Token Management
  3) Validate Token
    → Tests current token
    → Shows: "✓ Token is VALID!"
    → OR: "✗ Token is INVALID or EXPIRED!"
```

### Rotating Token (Security Best Practice)

```text
1. On GitHub Settings:
   - Revoke old token
   - Generate new token
   
2. In TERA:
   Token Management
     1) Setup/Change Token
       → Prompts: "Replace existing token?"
       → Paste new token
       → Validates and saves
       
3. Back on GitHub:
   - Confirm old token is revoked
```

### Switching GitHub Accounts

```text
Token Management
  4) Delete Token
    → Confirms deletion
    → Removes current token
    
  1) Setup/Change Token
    → Paste new account's token
    → Validates
    → Saves new token
```

## Security Best Practices

### Storing Tokens Safely

✓ **Do:**
- Use TERA's Token Management menu
- Create tokens with minimal scope (`gist` only)
- Rotate tokens periodically
- Validate tokens regularly
- Keep your machine secure

✗ **Don't:**
- Share your token with others
- Store token in plain text files you share
- Commit token to git repositories
- Use token with unnecessary scopes
- Ignore validation errors

### If Token Is Compromised

1. **Immediately on GitHub:**
   - Go to Settings → Developer settings → Personal access tokens
   - Find the compromised token
   - Click "Delete"

2. **In TERA:**
   - Run Token Management
   - Select "Delete Token"
   - Generate new token
   - Set up new token in TERA

3. **Verification:**
   - Validate new token
   - Check your GitHub account for unauthorized gists

### Token Rotation Schedule

**Recommended:**
- Rotate tokens every 90 days
- Immediately after security incidents
- When switching machines
- When sharing machine access

**How to Rotate:**
```text
1. Generate new token on GitHub
2. TERA → Token Management → Setup/Change Token
3. Paste new token
4. Revoke old token on GitHub
5. Validate new token in TERA
```

## Troubleshooting

### "Token validation failed"

**Causes and Solutions:**
```text
1. Token too short
   → Copy full token from GitHub (should be ~40+ chars)

2. Token format wrong
   → Should start with "ghp_"
   → Copy directly from GitHub, don't edit

3. Missing 'gist' scope
   → Go to GitHub settings
   → Edit token
   → Add 'gist' scope
   → Click "Update token"

4. Token expired
   → Generate new token on GitHub
   → Update in TERA's Token Management

5. Network issue
   → Check internet connection
   → Try again later
```

### "Token not found" errors when using Gist features

**Causes and Solutions:**
```text
1. Token was deleted
   → Run Token Management → Setup Token

2. Token file permissions wrong
   → Check: ls -la ~/.config/tera/tokens/github_token
   → Should show: -rw------- (600)
   → Reset: chmod 600 ~/.config/tera/tokens/github_token

3. Wrong token location
   → Token should be at: ~/.config/tera/tokens/github_token
   → Check: cat ~/.config/tera/tokens/github_token

4. Directory doesn't exist
   → Create: mkdir -p ~/.config/tera/tokens
   → Then setup token again
```

### Token works, but Gist creation fails

**Troubleshooting:**
```text
1. Validate token first:
   → Token Management → Validate Token
   → Should show "✓ Token is VALID!"

2. Check GitHub API status:
   → Visit https://www.githubstatus.com/
   → Ensure API is operational

3. Verify favorite lists exist:
   → Main menu → 3) List
   → Create lists if needed

4. Check internet connection:
   → Try: ping github.com
   → Or: curl -I https://api.github.com
```

## Integration with Gist Operations

Once token is configured, you can use all Gist features:

- **Create a gist:** Backup favorites automatically
- **My Gists:** View all created gists
- **Recover from gist:** Restore backups
- **Update a gist:** Change descriptions
- **Delete a gist:** Remove old backups

### Gist Operations that Need Token

| Operation    | Requires Token | Can Proceed Without? |
| ------------ | -------------- | -------------------- |
| Create gist  | ✓ Yes          | No                   |
| View gists   | ✓ Yes          | No                   |
| Recover gist | ✓ Yes          | No                   |
| Update gist  | ✓ Yes          | No                   |
| Delete gist  | ✓ Yes          | No                   |

## Environment Variables

TERA exports the following after loading token:

```bash
GITHUB_TOKEN=ghp_YourActualTokenHere123456789
```

This is automatically set when TERA starts and can be used by other scripts in your shell session.

## FAQ

**Q: Is my token encrypted?**
A: No, tokens are stored in plain text. Security relies on file permissions (600) and protecting your machine.

**Q: Can I use the same token on multiple machines?**
A: Yes, tokens work on any machine. For security, consider having separate tokens per machine.

**Q: What happens if I delete the token file manually?**
A: TERA will act as if no token is configured. Use Token Management to set it up again.

**Q: Can I have multiple tokens?**
A: Currently, TERA stores one token. Create multiple on GitHub, then switch in TERA when needed.

**Q: Does token expire?**
A: GitHub tokens don't auto-expire, but GitHub may revoke them. Validate regularly.

**Q: How do I backup my token?**
A: Copy the token string from GitHub settings. Don't backup token files.

**Q: What if I forget my token?**
A: You can't retrieve it from TERA. Generate a new one on GitHub and set it up again.
