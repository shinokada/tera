# TERA v3 Migration Guide

This guide helps you upgrade from TERA v2 to v3.

## TL;DR - Quick Summary

✅ **Migration is automatic** - Just upgrade and run TERA  
✅ **Your data is safe** - Favorites and blocklist untouched  
✅ **Old configs backed up** - Can rollback if needed  
✅ **Zero downtime** - Upgrade takes seconds  

## Table of Contents

- [What's New in v3](#whats-new-in-v3)
- [Migration Process](#migration-process)
- [What Gets Migrated](#what-gets-migrated)
- [What Stays The Same](#what-stays-the-same)
- [File Location Changes](#file-location-changes)
- [Troubleshooting](#troubleshooting)
- [Rollback Instructions](#rollback-instructions)
- [Manual Migration](#manual-migration)

---

## What's New in v3

### Unified Configuration System

The biggest change in v3 is configuration consolidation. Instead of managing multiple YAML files, everything is now in one place.

**Before (v2):**
```text
~/.config/tera/
├── theme.yaml              # Theme colors and padding
├── appearance_config.yaml  # Header customization
├── connection_config.yaml  # Network settings
├── shuffle.yaml            # Shuffle mode settings
├── blocklist.json
├── voted_stations.json
├── tokens/
│   └── github_token
└── favorites/
    └── *.json
```

**After (v3):**
```text
~/.config/tera/
├── config.yaml              # 🆕 All settings in one file!
├── data/                    # 🆕 User data directory
│   ├── blocklist.json       # (moved)
│   ├── voted_stations.json  # (moved)
│   ├── favorites/           # (moved)
│   │   └── *.json
│   └── cache/
│       ├── gist_metadata.json
│       └── search-history.json
└── .v2-backup-20250209-143025/  # 🆕 Automatic backup
    ├── theme.yaml
    ├── appearance_config.yaml
    ├── connection_config.yaml
    └── shuffle.yaml
```

### New Features

- **Unified Config**: One `config.yaml` file for all settings
- **Better Organization**: Config vs user data clearly separated
- **Automatic Migration**: Seamless upgrade from v2
- **Config Validation**: Auto-correct invalid values
- **New CLI Commands**: Manage config from command line
- **Secure Tokens**: Optional OS keychain storage for GitHub tokens

---

## Migration Process

### Automatic Migration (Recommended)

Migration happens automatically when you first run TERA v3:

```sh
# 1. Upgrade TERA
brew upgrade shinokada/tera/tera  # or your preferred method

# 2. Run TERA
tera

# You'll see:
🔄 Migrated TERA configuration from v2 to v3
✓ Config unified → ~/.config/tera/config.yaml
✓ Old configs backed up with timestamp
```

That's it! TERA is now running v3 with all your settings preserved.

### What Happens During Migration

1. **Detection**: TERA checks if `config.yaml` exists
2. **Backup**: Old v2 configs backed up to `.v2-backup-YYYYMMDD-HHMMSS/`
3. **Conversion**: Settings converted to unified format
4. **Validation**: Values checked and corrected if needed
5. **Organization**: User data moved to `data/` directory
6. **Cleanup**: Old config files removed (backup kept)
7. **Token Migration**: GitHub token moved to keychain (optional)

### Migration Timeline

- **Detection**: < 1 second
- **Backup**: < 1 second
- **Conversion**: < 1 second
- **Total Time**: ~2-3 seconds

---

## What Gets Migrated

### ✅ Configuration Settings

All your v2 settings are preserved:

| v2 File                  | →   | v3 Location                   | What's Migrated                               |
| ------------------------ | --- | ----------------------------- | --------------------------------------------- |
| `theme.yaml`             | →   | `config.yaml` (ui.theme)      | Colors, padding, theme name                   |
| `appearance_config.yaml` | →   | `config.yaml` (ui.appearance) | Header mode, alignment, width, color, padding |
| `connection_config.yaml` | →   | `config.yaml` (network)       | Auto-reconnect, delay, buffer size            |
| `shuffle.yaml`           | →   | `config.yaml` (shuffle)       | Auto-advance, interval, history settings      |
| `tokens/github_token`    | →   | OS Keychain (optional)        | GitHub personal access token                  |

### ✅ Your Data (Stays Unchanged)

These are moved to `data/` but otherwise untouched:

- **Favorites**: All your station lists
- **Blocklist**: Blocked stations
- **Voting History**: Stations you've voted for
- **Gist Metadata**: Your gist backup history
- **Search History**: Recent searches

---

## What Stays The Same

### No Changes Needed For:

- ✅ **Using Favorites** - Works exactly as before
- ✅ **Station Blocklist** - Same functionality
- ✅ **Gist Sync** - Same workflow
- ✅ **Keyboard Shortcuts** - All unchanged
- ✅ **Playback Controls** - Identical behavior
- ✅ **Search Functions** - Same as v2
- ✅ **Theme Selection** - Same themes available
- ✅ **File Locations** - Same OS-specific directories

### Backwards Compatibility

v3 is **100% backwards compatible** with v2 user data:
- Old favorites files work without modification
- Blocklist format unchanged
- Voting history preserved
- No breaking changes to user workflows

---

## File Location Changes

### macOS Example

```text
Before (v2):
~/Library/Application Support/tera/
├── theme.yaml
├── appearance_config.yaml
├── connection_config.yaml
├── shuffle.yaml
├── blocklist.json              ← At root level
├── voted_stations.json         ← At root level
├── tokens/
│   └── github_token
└── favorites/                  ← At root level
    ├── My-favorites.json
    └── Jazz.json

After (v3):
~/Library/Application Support/tera/
├── config.yaml                 ← 🆕 Unified config
├── data/                       ← 🆕 Data directory
│   ├── blocklist.json          ← Moved here
│   ├── voted_stations.json     ← Moved here
│   ├── favorites/              ← Moved here
│   │   ├── My-favorites.json
│   │   └── Jazz.json
│   └── cache/
│       └── gist_metadata.json
└── .v2-backup-20250209-143025/ ← 🆕 Backup
    ├── theme.yaml
    ├── appearance_config.yaml
    ├── connection_config.yaml
    └── shuffle.yaml
```

### Linux

- **Location**: `~/.config/tera/` (same as v2)
- **Structure**: Same as macOS example above

### Windows

- **Location**: `%APPDATA%\tera\` (same as v2)
- **Structure**: Same as macOS example above

---

## Unified Config Structure

### Example `config.yaml`

```yaml
version: "3.0"

player:
  default_volume: 100
  buffer_size_mb: 50

ui:
  theme:
    name: "default"
    colors:
      primary: "6"      # Cyan
      secondary: "12"   # Bright Blue
      highlight: "3"    # Yellow
      error: "9"        # Bright Red
      success: "2"      # Green
      muted: "8"        # Gray
      text: "7"         # White
    padding:
      page_horizontal: 2
      page_vertical: 1
      list_item_left: 2
      box_horizontal: 2
      box_vertical: 1
      
  appearance:
    header_mode: "default"      # text, ascii, none
    header_align: "center"      # left, center, right
    header_width: 50
    custom_text: ""
    ascii_art: ""
    header_color: "auto"
    header_bold: true
    padding_top: 1
    padding_bottom: 0
    
  default_list: "My-favorites"

network:
  auto_reconnect: true
  reconnect_delay: 5
  buffer_size_mb: 50

shuffle:
  auto_advance: false
  interval_minutes: 5
  remember_history: true
  max_history: 5
```

### How to Edit

```sh
# Show config location
tera config path

# Open in your editor
vi ~/.config/tera/config.yaml  # Linux
vi ~/Library/Application\ Support/tera/config.yaml  # macOS
notepad %APPDATA%\tera\config.yaml  # Windows

# Validate after editing
tera config validate

# Reset to defaults if needed
tera config reset
```

---

## Secure Token Storage (Optional)

v3 can optionally store your GitHub token in the OS keychain instead of a file.

### Benefits

- ✅ **More Secure**: OS handles encryption
- ✅ **Standard Practice**: Same as browsers, Docker, Git
- ✅ **Cross-Platform**: Works on macOS, Linux, Windows
- ✅ **Automatic Migration**: Token moved from file if found

### How It Works

**During Migration:**
1. TERA finds `tokens/github_token`
2. Reads the token
3. Stores it in OS keychain
4. Removes the token file
5. Removes `tokens/` directory

**Manual Setup (if migration failed):**
1. Go to Settings → GitHub Token
2. Press `e` to edit
3. Paste your token
4. Press Enter to save
5. Token stored in keychain

### Fallback Options

If keychain storage fails, you can:
- Use environment variable: `export TERA_GITHUB_TOKEN=ghp_xxxxx`
- Keep using file-based storage (deprecated but supported)

---

## Troubleshooting

### Migration Didn't Run

**Symptoms:**
- Still seeing old v2 file structure
- No `config.yaml` created
- No migration message shown

**Solution:**
```sh
# Check migration status
tera config migrate

# Force migration
tera config migrate --force

# Or reset and reconfigure
tera config reset
```

### Config Validation Errors

**Symptoms:**
- Error message about invalid config values
- TERA won't start
- Settings not loading

**Solution:**
```sh
# Check what's wrong
tera config validate

# Reset to defaults
tera config reset

# Then reconfigure through Settings UI
tera
# → Settings → Theme/Appearance/etc.
```

### Missing Favorites

**Symptoms:**
- Can't find your station lists
- "No favorite lists found" error

**Cause:** Migration may have failed to move favorites

**Solution:**
```sh
# Check if favorites still in old location
ls ~/.config/tera/favorites/  # Linux
ls ~/Library/Application\ Support/tera/favorites/  # macOS

# If they're there, manually move to data/
mkdir -p ~/.config/tera/data/
mv ~/.config/tera/favorites ~/.config/tera/data/
```

### GitHub Token Not Working

**Symptoms:**
- "Token not configured" error
- Gist operations fail
- Token prompt appears

**Solution 1 - UI:**
```text
1. Open TERA
2. Go to Settings → GitHub Token
3. Press 'e' to edit
4. Paste token
5. Press Enter to save
```

**Solution 2 - Environment Variable:**
```sh
# Add to your shell profile (~/.bashrc, ~/.zshrc, etc.)
export TERA_GITHUB_TOKEN=ghp_your_token_here

# Or run temporarily
export TERA_GITHUB_TOKEN=ghp_your_token_here
tera
```

### Rollback Needed

See [Rollback Instructions](#rollback-instructions) below.

---

## Rollback Instructions

If you need to rollback to v2:

### Step 1: Downgrade TERA

```sh
# Homebrew
brew uninstall tera
brew install shinokada/tera/tera@2

# Go
go install github.com/shinokada/tera/v2/cmd/tera@latest

# Manual
# Download v2.x.x from releases page
```

### Step 2: Restore v2 Configs

```sh
# Find your backup
ls ~/.config/tera/.v2-backup-*  # Linux
ls ~/Library/Application\ Support/tera/.v2-backup-*  # macOS

# Copy configs back
cp ~/.config/tera/.v2-backup-YYYYMMDD-HHMMSS/* ~/.config/tera/

# Move data back to root
mv ~/.config/tera/data/blocklist.json ~/.config/tera/
mv ~/.config/tera/data/voted_stations.json ~/.config/tera/
mv ~/.config/tera/data/favorites ~/.config/tera/

# Remove v3 files
rm ~/.config/tera/config.yaml
rm -rf ~/.config/tera/data/
```

### Step 3: Verify

```sh
tera --version  # Should show v2.x.x
tera            # Should start normally with old config
```

---

## Manual Migration

If automatic migration fails, you can migrate manually:

### Step 1: Create config.yaml
Replace `~/.config` with your OS specific, e.g. macOS `~/Library/Application Support`, Windows `C:\Users\<username>\AppData\Roaming
`, etc.

```sh
# Create new config
cat > ~/.config/tera/config.yaml << 'EOF'
version: "3.0"

player:
  default_volume: 100
  buffer_size_mb: 50

ui:
  theme:
    name: "default"
    colors:
      primary: "6"
      secondary: "12"
      highlight: "3"
      error: "9"
      success: "2"
      muted: "8"
      text: "7"
    padding:
      page_horizontal: 2
      page_vertical: 1
      list_item_left: 2
      box_horizontal: 2
      box_vertical: 1
  appearance:
    header_mode: "default"
    header_align: "center"
    header_width: 50
    custom_text: ""
    ascii_art: ""
    header_color: "auto"
    header_bold: true
    padding_top: 1
    padding_bottom: 0
  default_list: "My-favorites"

network:
  auto_reconnect: true
  reconnect_delay: 5
  buffer_size_mb: 50

shuffle:
  auto_advance: false
  interval_minutes: 5
  remember_history: true
  max_history: 5
EOF
```

### Step 2: Copy Your Old Settings

If you had customized settings, manually copy them:

```yaml
# From theme.yaml → ui.theme section
# From appearance_config.yaml → ui.appearance section
# From connection_config.yaml → network section
# From shuffle.yaml → shuffle section
```

### Step 3: Organize Data Directory

```sh
# Create data directory
mkdir -p ~/.config/tera/data/

# Create cache directory
mkdir -p ~/.config/tera/data/cache/

# Move user data
mv ~/.config/tera/favorites ~/.config/tera/data/
mv ~/.config/tera/blocklist.json ~/.config/tera/data/
mv ~/.config/tera/voted_stations.json ~/.config/tera/data/
mv ~/.config/tera/gist_metadata.json ~/.config/tera/data/cache/
```

### Step 4: Backup Old Configs

```sh
# Create backup directory
mkdir -p ~/.config/tera/.v2-backup-manual/

# Move old configs
mv ~/.config/tera/theme.yaml ~/.config/tera/.v2-backup-manual/
mv ~/.config/tera/appearance_config.yaml ~/.config/tera/.v2-backup-manual/
mv ~/.config/tera/connection_config.yaml ~/.config/tera/.v2-backup-manual/
mv ~/.config/tera/shuffle.yaml ~/.config/tera/.v2-backup-manual/
```

### Step 5: Verify

```sh
# Check config is valid
tera config validate

# Start TERA
tera
```

---

## Getting Help

If you encounter issues during migration:

1. **Check this guide** - Most issues covered above
2. **Validate config** - Run `tera config validate`
3. **Check logs** - TERA shows errors on startup
4. **Reset if needed** - Run `tera config reset`
5. **Open an issue** - [GitHub Issues](https://github.com/shinokada/tera/issues)

When opening an issue, include:
- TERA version: `tera --version`
- OS and version
- Config file: `cat ~/.config/tera/config.yaml`
- Error messages
- Steps to reproduce

---

## Summary

### Migration Checklist

- ✅ Backup created automatically
- ✅ Config unified into `config.yaml`
- ✅ User data moved to `data/` directory
- ✅ GitHub token migrated to keychain (optional)
- ✅ Old configs removed (backup kept)
- ✅ Favorites and blocklist work as before
- ✅ Can rollback if needed

### Benefits of v3

- 🎯 **Simpler**: One config file instead of four
- 📁 **Organized**: Config vs data clearly separated
- 🔒 **Secure**: Optional keychain token storage
- ✅ **Validated**: Auto-correct invalid values
- 🛠️ **Manageable**: New CLI commands for config
- ⚡ **Faster**: Reduced file I/O operations

### Need More Help?

- 📖 [README.md](../README.md) - Full documentation
- 📝 [CHANGELOG.md](../CHANGELOG.md) - What changed
- 🐛 [Issues](https://github.com/shinokada/tera/issues) - Report problems
- 💬 [Discussions](https://github.com/shinokada/tera/discussions) - Ask questions

---

**Welcome to TERA v3!** 🎉
