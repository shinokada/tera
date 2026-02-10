# Changelog

All notable changes to TERA will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [3.1.0]

- OS keychain token storage with file-based fallback
- Environment variable support (TERA_GITHUB_TOKEN)
- Automatic token migration from file to keychain
- migrate-token CLI subcommand
- Token source detection (GetTokenSource)

## [3.0.0] - 2026-02-09

### 🎉 Major Changes

#### Unified Configuration System
The biggest change in v3 is the consolidation of all configuration into a single `config.yaml` file. This makes TERA easier to configure and maintain.

**Before (v2):**
```text
~/.config/tera/
├── theme.yaml
├── appearance_config.yaml
├── connection_config.yaml
├── shuffle.yaml
└── ...
```

**After (v3):**
```text
~/.config/tera/
├── config.yaml              # All settings in one file
└── data/                    # User data separated
    ├── blocklist.json
    ├── voted_stations.json
    └── favorites/
```

### ✨ Added

#### Configuration Management
- **Unified Config System**: All settings now in one `config.yaml` file
  - Player settings (volume, buffer)
  - UI settings (theme, appearance)
  - Network settings (auto-reconnect, buffer)
  - Shuffle settings (auto-advance, history)
- **Automatic Migration**: v2 configs automatically converted on first run
- **Config Validation**: Built-in validation with auto-correction for invalid values
- **Config Backup**: Old v2 configs automatically backed up with timestamp

#### New CLI Commands
```sh
# Config management
tera config path       # Show config file location
tera config reset      # Reset all settings to defaults
tera config validate   # Check config for errors
tera config migrate    # Check migration status / force migrate

# Theme management
tera theme reset       # Reset theme to defaults
tera theme path        # Show config file location
tera theme edit        # Show where to edit theme
tera theme export      # Export theme as standalone file
```

#### Secure Credential Storage (Optional)
- **OS Keychain Integration**: GitHub tokens can now be stored in OS keychain
  - macOS: Keychain Access
  - Linux: Secret Service (gnome-keyring, KWallet)
  - Windows: Credential Manager
- **Automatic Token Migration**: Old file-based tokens migrated on first run
- **Environment Variable Fallback**: For CI/CD and headless environments
- **Token Management UI**: New Settings → GitHub Token for easy management

#### File Organization
- **Clear Data Separation**: Config vs user data now clearly separated
- **New Data Directory**: All user data organized under `data/` subdirectory
  - `data/blocklist.json`
  - `data/voted_stations.json`
  - `data/favorites/`
  - `data/cache/`
- **Backwards Compatible**: Favorites and blocklist work exactly as before

### 🔧 Changed

#### Configuration Structure
- **Theme Config**: Now under `ui.theme` in unified config
- **Appearance Config**: Now under `ui.appearance` in unified config
- **Network Config**: Now under `network` in unified config
- **Shuffle Config**: Now under `shuffle` in unified config
- **File Locations**: User data moved to `data/` subdirectory

#### Storage Package
- **Adapter Functions**: New helper functions for accessing unified config
- **Backward Compatible**: Existing code continues to work during transition
- **Migration Check**: Automatic v2 config detection and migration

#### Theme System
- **Unified Access**: Theme now reads from `config.yaml`
- **Export Function**: Can still export standalone `theme.yaml` if needed
- **No Behavior Changes**: All theme features work exactly as before

### 🐛 Fixed

- **Header Spacing**: Fixed missing blank line between TERA header and page title
- **Type Safety**: Fixed HeaderMode type conversion in config adapter
- **Migration Safety**: Robust error handling during v2 to v3 migration

### 📝 Documentation

- **Migration Guide**: Comprehensive guide for v2 → v3 upgrade
- **Updated README**: Reflects v3 changes and new features
- **API Changes**: Internal storage API updated for unified config
- **File Locations**: Updated documentation for new file structure

### 🔄 Migration Notes

**For Users:**
- Migration is **automatic** on first run of v3
- Your favorites and user data are **never touched**
- Old v2 configs are **backed up** with timestamp
- If migration fails, you can **retry** with `tera config migrate`
- You can **reset** to defaults with `tera config reset`

**What Gets Migrated:**
- ✅ Theme settings (colors, padding)
- ✅ Appearance settings (header customization)
- ✅ Connection settings (auto-reconnect, buffer)
- ✅ Shuffle settings (auto-advance, history)
- ✅ GitHub token (to keychain, optional)

**What Stays The Same:**
- ✅ Favorites lists (unchanged)
- ✅ Blocklist (unchanged)
- ✅ Voting history (unchanged)
- ✅ All file locations (same OS-specific directories)

**Rollback:**
If you need to rollback to v2, your backed-up configs are in:
```text
~/.config/tera/.v2-backup-YYYYMMDD-HHMMSS/
```

### 🚀 Performance

- **Lazy Config Loading**: Config only loaded when needed
- **In-Memory Caching**: Parsed config cached for better performance
- **Reduced File I/O**: One file instead of four reduces disk operations

### 🔒 Security

- **Secure Token Storage**: OS keychain integration for GitHub tokens
- **Automatic Encryption**: OS handles encryption and security
- **File-Based Fallback**: Still supports file-based tokens (deprecated)
- **Environment Variables**: Headless environments can use `TERA_GITHUB_TOKEN`

### ⚠️ Breaking Changes

None! v3 is fully backward compatible:
- Existing favorites work without changes
- Blocklist works without changes
- User data locations unchanged
- Automatic migration handles all config changes
- Old v2 configs backed up, not deleted

### 📦 Dependencies

No new required dependencies. Optional features:
- `github.com/zalando/go-keyring` - For secure token storage (optional)

### 🙏 Acknowledgments

Thanks to all users who provided feedback on configuration management and helped shape v3!

---

## [2.x.x] - Previous Releases

For v2 release history, see the [v2 branch](https://github.com/shinokada/tera/tree/v2).

### Key v2 Features
- Multi-file configuration system
- Theme customization via YAML
- Appearance settings for header
- Connection settings for unstable networks
- Shuffle mode with auto-advance
- Block list functionality
- Gist sync for backup
- Voting support
- Quick play from main menu

---

## Semantic Versioning

TERA follows [Semantic Versioning](https://semver.org/):

- **MAJOR** version for incompatible API changes
- **MINOR** version for backwards-compatible functionality additions
- **PATCH** version for backwards-compatible bug fixes

### Version Number Format: MAJOR.MINOR.PATCH

Examples:
- `3.0.0` - Major release with configuration overhaul (fully backward compatible)
- `3.1.0` - Minor release with new features
- `3.0.1` - Patch release with bug fixes

---

## Upgrade Paths

### From v2 to v3
- ✅ **Automatic migration** on first run
- ✅ **Zero downtime** - just upgrade and run
- ✅ **Rollback available** - old configs backed up

### From v1 to v3
- ⚠️ Must upgrade to v2 first, then to v3
- See v2 migration guide for v1 → v2 upgrade
- Then follow automatic v2 → v3 migration

---

[3.0.0]: https://github.com/shinokada/tera/releases/tag/v3.0.0
