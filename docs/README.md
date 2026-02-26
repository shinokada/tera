# Tera: Terminal Radio

A terminal-based internet radio player powered by [Radio Browser](https://www.radio-browser.info/).

## Features

- 🔍 **Search** - Find stations by name, tag, language, country, or state
- 🎲 **I Feel Lucky** - Random station discovery by keyword
- 💾 **Favorites** - Organize stations into custom lists with duplicate detection
- ⭐ **Star Ratings** - Rate stations 1-5 stars and browse your top-rated collection
- 🏷️ **Custom Tags** - Tag stations with personal labels and build dynamic playlists
- ⚡ **Quick Play** - Direct playback from main menu (shortcuts 10-99+)
- 🔊 **Playback Control** - Play/pause with persistent status, adjust volume, and mute during playback
- 🚫 **Block List** - Block unwanted stations from appearing in searches and auto-play
- ☁️ **Gist Sync** - Backup and restore favorites via GitHub Gists
- 🗳️ **Voting** - Support your favorite stations on Radio Browser
- 📊 **Most Played** - View your listening history sorted by play count, last played, or first played
- 🎨 **Themes** - Choose from predefined themes or customize via unified config
- 💤 **Sleep Timer** - Set a timer to stop playback automatically
- 🔄 **Update Checker** - Get notified when a new version is available
- ⌨️ **Keyboard-driven** - Full navigation without a mouse
- ❓ **Context Help** - Press `?` anytime to see available keyboard shortcuts

## What's New in v3.0.0

🎉 **Unified Configuration System** - All settings now in one `config.yaml` file!

- **Simpler Configuration** - One file instead of multiple YAML files
- **Automatic Migration** - v2 configs automatically converted on first run
- **Secure Token Storage** - GitHub tokens now stored in OS keychain (optional)
- **Better Organization** - Clear separation of config vs user data
- **Easier to Edit** - All settings in one place with validation

See [MIGRATION.md](/docs/MIGRATION.md) for upgrade details and [CHANGELOG.md](CHANGELOG.md) for full changes.

## Requirements

- [mpv](https://mpv.io/) - Media player for audio playback

## Installation

### Homebrew (macOS/Linux)
```sh
# Install or update
brew update && brew upgrade
brew install shinokada/tera/tera
# Upgrade existing installation
brew upgrade shinokada/tera/tera
```

### Golang
```sh
go install github.com/shinokada/tera/v3/cmd/tera@latest
```

### Debian/Ubuntu
```sh
sudo dpkg -i tera_3.x.x_linux_amd64.deb
sudo apt-get install -f  # Install mpv dependency if needed
```

### Fedora/RHEL
```sh
sudo rpm -i tera_3.x.x_linux_amd64.rpm
```

### Windows Scoop
```sh
scoop bucket add shinokada https://github.com/shinokada/scoop-bucket
scoop install tera
```

### Windows Winget
```sh
winget install Shinokada.Tera
# Upgrade existing installation
winget upgrade Shinokada.Tera
```

### Manual Download

Download the latest binary for your platform from the [releases page](https://github.com/shinokada/tera/releases):

| Platform | Architecture  | File                             |
| -------- | ------------- | -------------------------------- |
| macOS    | Intel         | `tera_3.x.x_darwin_amd64.tar.gz` |
| macOS    | Apple Silicon | `tera_3.x.x_darwin_arm64.tar.gz` |
| Linux    | x86_64        | `tera_3.x.x_linux_amd64.tar.gz`  |
| Linux    | ARM64         | `tera_3.x.x_linux_arm64.tar.gz`  |
| Windows  | x86_64        | `tera_3.x.x_windows_amd64.zip`   |
| Windows  | ARM64         | `tera_3.x.x_windows_arm64.zip`   |

#### macOS/Linux

```sh
# Download and extract (example for macOS Apple Silicon)
tar -xzf tera_3.x.x_darwin_arm64.tar.gz

# Move to a directory in your PATH
sudo mv tera /usr/local/bin/
```

#### Windows

1. Download the `.zip` file for your architecture
2. Extract the archive
3. Add the extracted directory to your PATH or move `tera.exe` to a directory already in your PATH

## Upgrading from v2 to v3

**Good news!** Migration is automatic. When you first run TERA v3:

1. ✅ Your v2 config files are automatically detected
2. ✅ Settings are migrated to the new unified `config.yaml`
3. ✅ Your favorites and data remain untouched
4. ✅ Old config files are backed up with timestamp
5. ✅ You're ready to go!

See the [Migration Guide](/docs/MIGRATION.md) for details.

## Quick Start

```sh
# Start TERA
tera

# Main Menu Options:
# 1) Play from Favorites - Browse your saved lists
# 2) Search Stations     - Find new stations
# 3) Most Played         - Your listening statistics
# 4) Top Rated           - Browse your highest-rated stations
# 5) Browse by Tag       - Browse stations by your custom tags
# 6) Tag Playlists       - Dynamic playlists from tag combinations
# 7) Manage Lists        - Create/edit/delete favorite lists
# 8) Block List          - Manage blocked stations
# 9) I Feel Lucky        - Random station by keyword
# 0) Gist Management     - Backup/restore via GitHub
# -) Settings            - Configure TERA

# Quick Play (from main menu):
# Type 10-99+ to instantly play stations from "My-favorites"

# Need help? Press ? anytime to see keyboard shortcuts!
```

## Configuration Management (v3)

TERA v3 introduces new command-line tools for managing your configuration:

### Config Commands

```sh
# View config file location
tera config path

# Reset all settings to defaults
tera config reset

# Validate config file
tera config validate

# Check migration status
tera config migrate
```

### Theme Commands

```sh
# Reset theme to defaults
tera theme reset

# Show config file location
tera theme path

# Show where to edit theme
tera theme edit

# Export theme as standalone file
tera theme export
```

## Main Features

### Play from Favorites

Browse and play stations from your organized lists. Navigate with `↑↓` or `jk`, press `Enter` to play.

### Star Ratings

Rate your favorite stations from 1-5 stars to build your personal collection of top stations.

**How to Rate:**
- While playing any station, press `r` then `1-5` to rate
- Press `r` then `0` to remove a rating
- Press `r` then any other key (or `Esc`) to cancel without changing the rating
- Stars appear in search results, favorites, and the playing screen

**Top-Rated View:**
From main menu, select "4. Top Rated" to:
- Browse all your rated stations sorted by rating
- Filter by minimum star rating (press `f`)
- Sort by rating high/low or recently rated (press `s`)
- Play stations directly from the list

**Keyboard Shortcuts:**

| Screen    | Key            | Action           |
| --------- | -------------- | ---------------- |
| Playing   | `r` then `1-5` | Rate station     |
| Playing   | `r` then `0`   | Remove rating    |
| Top Rated | `f`            | Cycle filter     |
| Top Rated | `s`            | Cycle sort order |

**Storage Location:**
- Linux: `~/.config/tera/data/station_ratings.json`
- macOS: `~/Library/Application Support/tera/data/station_ratings.json`
- Windows: `%APPDATA%\tera\data\station_ratings.json`

### Custom Tags

Organize stations with your own personal labels. Tags are stored locally and never transmitted.

**How to Tag:**
- While playing any station, press `t` to add a single tag
- Press `T` (shift+t) to open the **Manage Tags** dialog and toggle multiple tags at once
- The tag input has autocomplete — start typing and press Tab to complete from existing tags

**Browse by Tag (menu option 5):**
- See all your tags and how many stations each one covers
- Select a tag to browse and play the matching stations
- Press `d` on a tag to remove it from every station at once

**Tag Playlists (menu option 6):**
- Create named playlists that dynamically pull in stations matching a tag combination
- Choose **any** (OR) or **all** (AND) matching
- Edit or delete playlists at any time; the station list updates automatically

**Tag pills in lists:**
Tagged stations show `[tag]` pills inline in every list view — Favorites, Search, Most Played, and Top Rated — so you can see your labels at a glance.

**Keyboard Shortcuts:**

| Screen        | Key | Action                       |
| ------------- | --- | ---------------------------- |
| Playing       | `t` | Add a tag (quick input)      |
| Playing       | `T` | Open Manage Tags dialog      |
| Browse by Tag | `d` | Delete tag from all stations |

**Tag rules:**
- Lowercase only (normalized automatically)
- Up to 50 characters; alphanumeric, spaces, hyphens, underscores
- Up to 20 tags per station

**Storage Location:**
- Linux: `~/.config/tera/data/station_tags.json`
- macOS: `~/Library/Application Support/tera/data/station_tags.json`
- Windows: `%APPDATA%\tera\data\station_tags.json`

### Most Played

Track your listening history and rediscover your favorite stations.

**What's tracked** (stored locally, never transmitted):
- Play count per station
- Last played timestamp
- First played timestamp
- Total listening duration

**Sort options** (press `s` to cycle):
- **Play Count** (default) — Most played stations first
- **Last Played** — Most recently heard first
- **First Played** — Oldest discoveries first

**Key bindings**:

| Key         | Action                |
| ----------- | --------------------- |
| `↑↓` / `jk` | Navigate              |
| `Enter`     | Play selected station |
| `s`         | Cycle sort order      |
| `f`         | Add to favorites      |
| `?`         | Help                  |
| `Esc` / `m` | Back to main menu     |

All data is stored in `station_metadata.json` in your data directory. Delete this file at any time to clear your history.

### Search Stations

Six search methods to find stations:
- **By Tag** - Genre/style (jazz, rock, classical)
- **By Name** - Station name (BBC, NPR, KEXP)
- **By Language** - Broadcasting language
- **By Country** - Country code (US, UK, JP)
- **By State** - Region/state name
- **Advanced** - Search both name and tags

See [Search Guide](#search-guide) below for details.

### Manage Lists

Create, rename, and delete your favorite lists. Stations can be:
- Saved to multiple lists
- Moved between lists
- Deleted from lists

**Duplicate Detection**: TERA automatically prevents adding the same station twice to any list.

### Block List

Block unwanted stations to prevent them from appearing in search results and shuffle mode.

**How to Block:**
- While playing any station, press `b` to block it instantly
- Press `u` within 5 seconds to undo (in case of accidental block)
- Works in Search, I Feel Lucky, and Play from Favorites

**Block List Management:**
From main menu, select "8. Block List" to:
- View all blocked stations with details (country, language, codec)
- Press `u` to unblock a selected station
- Press `c` to clear entire block list (with confirmation)

**Keyboard Shortcuts:**

| Screen     | Key | Action                    |
| ---------- | --- | ------------------------- |
| Playing    | `b` | Block current station     |
| Playing    | `u` | Undo block (5 sec window) |
| Block List | `u` | Unblock selected station  |
| Block List | `c` | Clear all blocks          |

**Storage Location:**
- Linux: `~/.config/tera/data/blocklist.json`
- macOS: `~/Library/Application Support/tera/data/blocklist.json`
- Windows: `%APPDATA%\tera\data\blocklist.json`

### Sleep Timer

Set a timer to automatically stop playback — useful for falling asleep to radio.

**How to Use:**
- While playing any station, press `Z` to open the sleep timer dialog
- Choose a preset duration (15, 30, 45, 60, or 90 minutes) or enter a custom value
- Press `+` while the timer is running to extend it by 15 minutes
- When the timer expires, playback stops and a session summary is shown

**Session Summary:**
- Lists every station played during the timer session
- Shows total listening duration vs. the duration you set
- Press `0` to return to the main menu or `q`/`Esc` to quit

**Keyboard Shortcuts:**

| Screen  | Key | Action                         |
| ------- | --- | ------------------------------ |
| Playing | `Z` | Open sleep timer dialog        |
| Playing | `+` | Extend running timer by 15 min |

### I Feel Lucky

Enter a keyword (genre, mood, style) and TERA finds a random matching station. Perfect for music discovery!

**Shuffle Mode**: Enable shuffle mode to automatically cycle through multiple stations matching your keyword:
- Press `Ctrl+T` to toggle shuffle on/off
- Stations play in random order without repeats
- Optional auto-advance timer (configurable)
- Navigate backward through recently played stations
- Configure shuffle behavior in Settings → Shuffle Settings

See [Shuffle Mode](#shuffle-mode) for detailed features.

### Settings

Access app configuration from the main menu (option `-`):

- **Theme / Colors** - Switch between predefined themes or customize colors
- **Appearance** - Customize header display (text, ASCII art, alignment, colors, padding)
- **Connection Settings** - Auto-reconnect and buffering for unstable networks (4G/GPRS)
- **Shuffle Settings** - Configure shuffle mode behavior (auto-advance, history size)
- **Search History** - View and clear your search history
- **Check for Updates** - View current version and check for new releases
- **About TERA** - See version, installation method, and update command

The Settings menu automatically detects how you installed TERA (Homebrew, Go, Scoop, Winget, etc.) and shows the appropriate update command.

### Appearance Settings

Customize how the TERA header appears at the top of the application:

**Header Modes:**
- **Default** - Show "TERA" text (default)
- **Text** - Display custom text
- **ASCII** - Show custom ASCII art (max 15 lines)
- **None** - Hide header completely

**Customization Options:**
- **Alignment** - Left, center, or right
- **Width** - Header width (10-120 characters)
- **Color** - Auto, hex code (#FF0000), or ANSI code (0-255)
- **Bold** - Enable/disable bold text
- **Padding** - Top and bottom spacing (0-10 lines)

**Tips:**
- Preview changes before saving
- Use [TAAG](https://patorjk.com/software/taag/) or `figlet` to generate ASCII art
- All settings stored in unified `config.yaml` (see [File Locations](#file-locations-v3))

### Connection Settings

For users on unstable networks (mobile data, GPRS, 4G), configure automatic reconnection:

- **Auto-reconnect** - Automatically retry when stream drops (default: enabled)
- **Reconnect delay** - Wait time between attempts: 1-30 seconds (default: 5s)
- **Stream buffer** - Cache size to handle brief signal drops: 10-200 MB (default: 50MB)

Settings stored in unified `config.yaml` (see [File Locations](#file-locations-v3)).

### Quick Play from Main Menu

The main menu shows your "My-favorites" list with shortcuts 10-99+. Type the number to play instantly:

```
Main Menu & Quick Play

Choose an option:

  1. Play from Favorites
  2. Search Stations
  3. Most Played
  4. Top Rated
  5. Browse by Tag
  6. Tag Playlists
  7. Manage Lists
  8. Block List
  9. I Feel Lucky
  0. Gist Management
  -. Settings

─── Quick Play Favorites ───
  10. Jazz FM • UK • MP3 192kbps
  11. BBC Radio 6 Music • UK • AAC 128kbps
  12. KEXP 90.3 FM • US • AAC 128kbps

Type 10-12 to play instantly!
```

**How it works:**
- Stations from "My-favorites.json" appear with shortcuts 10+
- Type the number (e.g., `11`) and press Enter
- Station plays immediately
- Press `Esc` to stop playback

## Theme Configuration

### In-App Theme Selection

The easiest way to change themes is through the Settings menu:

1. Press `-` from the main menu to open Settings
2. Select "Theme / Colors"
3. Choose from predefined themes:
   - **Default** - Cyan and blue tones
   - **Ocean** - Deep blue theme
   - **Forest** - Green nature theme
   - **Sunset** - Warm orange and red
   - **Purple Haze** - Purple and magenta
   - **Monochrome** - Classic black and white
   - **Dracula** - Popular dark theme
   - **Nord** - Arctic, north-bluish

### Manual Configuration

You can also customize colors and padding by editing the unified config file:

```sh
tera config path  # Show config file location
tera theme edit   # Show where to edit theme
tera theme reset  # Reset to defaults
```

The config file includes an ANSI color reference (0-15 standard colors, 16-255 extended colors). Example:

```yaml
ui:
  theme:
    colors:
      primary: "6"      # Cyan
      highlight: "3"    # Yellow
      error: "9"        # Bright Red
    padding:
      list_item_left: 2
```

## Update Checker

TERA automatically checks for new versions on startup. When an update is available:

- A yellow **⬆ Update** indicator appears in the main menu footer
- Go to **Settings → Check for Updates** for details and update instructions

### Checking for Updates

1. Press `-` from the main menu to open Settings
2. Select "Check for Updates" (option 2)
3. View:
   - Your current version
   - Latest available version
   - Link to release notes
   - Installation method (automatically detected)
   - Specific update command for your installation

TERA detects how you installed it and provides the correct update command. For example:
- If installed via Homebrew: Shows `brew upgrade shinokada/tera/tera`
- If installed via Go: Shows `go install github.com/shinokada/tera/v3/cmd/tera@latest`
- If installed via Scoop: Shows `scoop update tera`
- If installed via Winget: Shows `winget upgrade tera`

### Update Commands

| Installation Method | Update Command                                                            |
| ------------------- | ------------------------------------------------------------------------- |
| Homebrew            | `brew upgrade shinokada/tera/tera`                                        |
| Go install          | `go install github.com/shinokada/tera/v3/cmd/tera@latest`                 |
| Scoop               | `scoop update tera`                                                       |
| Winget              | `winget upgrade tera`                                                     |
| APT/DEB             | `sudo apt update && sudo apt install --only-upgrade tera`                 |
| RPM/DNF             | `sudo dnf upgrade tera`                                                   |
| Manual              | Download from [releases page](https://github.com/shinokada/tera/releases) |

## Shuffle Mode

Shuffle mode is an enhanced version of "I Feel Lucky" that lets you explore multiple stations matching your search keyword without manually searching each time.

### How It Works

1. Navigate to **I Feel Lucky** from the main menu (option 9)
2. Press `Ctrl+T` to toggle shuffle mode on
3. Enter your keyword (e.g., "jazz", "rock", "meditation")
4. Press Enter to start shuffle mode

### Features

**Automatic Station Discovery**
- TERA finds all stations matching your keyword
- Plays them in random order without repeats
- No duplicates until all stations have been played

**Auto-Advance Timer** (Optional)
- Automatically skip to the next station after a set interval
- Configurable intervals: 1, 3, 5, 10, or 15 minutes
- Pause/resume timer with `p` key
- Disable for manual control

**Station History**
- Keep track of recently played stations
- Navigate backward with `b` key
- Configurable history size: 3, 5, 7, or 10 stations
- See last few stations in the shuffle history display

**Seamless Playback**
- All standard playback controls work (volume, mute, favorites, voting)
- Save any station to your favorites while shuffling
- Stop shuffle but keep playing current station with `h`

### Shuffle Keyboard Shortcuts

| Key   | Action                                |
| ----- | ------------------------------------- |
| `Ctrl+T` | Toggle shuffle mode (in input screen) |
| `n`   | Next shuffle station (manual skip)    |
| `[`   | Previous station (from history)       |
| `b`   | Block current station                 |
| `u`   | Undo block (5 sec window)             |
| `p`   | Pause/resume auto-advance timer       |
| `h`   | Stop shuffle, keep playing current    |
| `f`   | Save to My-favorites                  |
| `s`   | Save to another list                  |
| `v`   | Vote for station                      |
| `Esc` | Stop shuffle and return to input      |

### Shuffle Settings

Configure shuffle behavior in **Settings → Shuffle Settings**:

**Auto-advance**
- Enable/disable automatic station switching
- When disabled, you manually control station changes with `n`/`[`

**Auto-advance Interval**
- Set how long each station plays before auto-advancing
- Options: 1, 3, 5, 10, or 15 minutes
- Default: 5 minutes

**Remember History**
- Enable/disable station history tracking
- When disabled, you cannot go back to previous stations

**History Size**
- Number of previous stations to remember
- Options: 3, 5, 7, or 10 stations
- Default: 5 stations

### Example Shuffle Session

```text
🎵 Now Playing (🔀 Shuffle: jazz)

Station: Smooth Jazz 24/7
Country: United States
Codec: AAC • Bitrate: 128 kbps

▶ Playing...

🔀 Shuffle Active • Next in: 4:23
   Station 3 of session
   
─── Shuffle History ───
  ← Jazz FM London
  ← WBGO Jazz 88.3
  → Smooth Jazz 24/7  ← Current

Space: Pause • n: Next • [: Prev • f: Fav • b: Block • p: Pause timer • h: Stop shuffle • 0: Main Menu • ?: Help
```

### Configuration

Shuffle settings are stored in the unified `config.yaml`:

```yaml
shuffle:
  auto_advance: true           # Auto-advance enabled
  interval_minutes: 5          # 5 minutes per station
  remember_history: true       # Track history
  max_history: 5               # Remember last 5 stations
```

You can edit this file directly or use the Settings menu.

## Keyboard Shortcuts

### Global Navigation

| Key         | Action        |
| ----------- | ------------- |
| `↑↓` / `jk` | Navigate      |
| `Enter`     | Select / Play |
| `Esc`       | Back / Stop   |
| `0`         | Main Menu     |
| `?`         | Help          |
| `Ctrl+C`    | Quit          |

### Main Menu

| Key      | Action                       |
| -------- | ---------------------------- |
| `0`      | Gist Management              |
| `1-9`    | Quick select menu item       |
| `10-99+` | Quick play from My-favorites |
| `-`      | Settings                     |

### Playback Controls

| Key     | Action            |
| ------- | ----------------- |
| `Space` | Pause / Resume    |
| `*`     | Volume up (+5%)   |
| `/`     | Volume down (-5%) |
| `m`     | Toggle mute       |
| `r`     | Rate station      |
| `b`     | Block station     |
| `u`     | Undo block (5s)   |
| `Z`     | Sleep timer       |
| `+`     | Extend timer      |

### Playing/Browsing Stations

| Key | Action               |
| --- | -------------------- |
| `f` | Save to My-favorites |
| `s` | Save to another list |
| `v` | Vote for station     |
| `t` | Add tag              |
| `T` | Manage tags          |

> **Tip:** Press `?` while playing to see all available shortcuts for the current screen in a help overlay.

### List Management

| Key | Action                |
| --- | --------------------- |
| `n` | New list (in dialogs) |
| `d` | Delete item           |

## Search Guide

The **Search Stations** menu offers multiple ways to find stations:

### Search Types

| Option                     | Description                              | Example Query                       |
| -------------------------- | ---------------------------------------- | ----------------------------------- |
| **Search by Tag**          | Find stations by genre/style tags        | `jazz`, `rock`, `news`, `classical` |
| **Search by Name**         | Find stations by their name              | `BBC`, `NPR`, `KEXP`                |
| **Search by Language**     | Find stations broadcasting in a language | `english`, `spanish`, `japanese`    |
| **Search by Country Code** | Find stations from a specific country    | `US`, `UK`, `FR`, `JP`              |
| **Search by State**        | Find stations from a state/region        | `California`, `Texas`, `Bavaria`    |
| **Advanced Search**        | Search both name AND tag fields          | `smooth jazz`, `classic rock`       |

### Query Format

- **Single words work**: `jazz`, `rock`, `news`
- **Multi-word phrases work**: `classic rock`, `smooth jazz`, `talk radio`
- **Partial matching**: Searching `BBC` finds "BBC Radio 1", "BBC World Service", etc.
- **Case insensitive**: `Jazz`, `JAZZ`, and `jazz` all work the same

### When to Use Advanced Search

Use **Advanced Search** when:
- You're not sure if your term is a station name or a genre tag
- You want broader results across multiple fields
- You're exploring and want maximum discovery

**Features:**
- **Country**: Enter a 2-letter code (e.g., "US") for Country Code search, or a full name (e.g., "Japan") for Country Name search.
- **Bitrate**: Press 1, 2, or 3 to filter by quality. Press the same number again to unselect.
- **Language**: Case-insensitive (e.g., "English" becomes "english").

For example, searching `jazz` in Advanced Search finds:
- Stations with "jazz" in their name (e.g., "Jazz FM")
- Stations tagged with "jazz" as a genre

### Search Results

Results are sorted by **votes** (most popular first) and limited to 100 stations. Broken/offline stations are automatically filtered out.

**In the results:**
- Navigate with `↑↓` or `jk`
- Press `Enter` to view station details and play
- Press `f` to add to My-favorites
- Press `s` to add to another list
- Press `v` to vote for the station

## Gist Sync

Backup and sync your favorite lists across devices using GitHub Gists.

**Quick Setup:**
1. Go to: Main Menu → 0) Gist Management → 6) Token Management
2. Create a GitHub Personal Access Token (with `gist` scope only)
3. Paste token in TERA
4. Create your first gist backup!

**Features:**
- Create secret or public gists
- View your gist history
- Recover favorites from any gist URL
- Update gist descriptions
- Delete old backups

**Documentation:**
- [Gist Setup Guide](/docs/GIST_SETUP.md) - Token setup and security
- [Gist Management Guide](/docs/GIST_CRUD_GUIDE.md) - Complete feature guide

## File Locations (v3)

TERA v3 organizes files more clearly with unified config and separate user data:

| Operating System | Location                              |
| ---------------- | ------------------------------------- |
| **Linux**        | `~/.config/tera/`                     |
| **macOS**        | `~/Library/Application Support/tera/` |
| **Windows**      | `%APPDATA%\tera\`                     |

### v3 File Structure

```text
tera/
├── config.yaml             # 🆕 Unified configuration (all settings)
├── data/                   # 🆕 User data directory
│   ├── blocklist.json      # Blocked radio stations
│   ├── voted_stations.json # Voting history
│   ├── station_metadata.json # 🆕 Play count & listening history
│   ├── station_ratings.json  # Star ratings
│   ├── station_tags.json     # Custom tags and tag playlists
│   ├── favorites/          # Your station lists
│   │   ├── My-favorites.json
│   │   ├── Rock.json
│   │   └── Jazz.json
│   └── cache/              # Temporary data
│       ├── gist_metadata.json
│       └── search-history.json
└── .v2-backup-YYYYMMDD-HHMMSS/  # 🆕 Automatic v2 config backup
```

**What changed from v2:**
- ✅ One `config.yaml` instead of multiple YAML files
- ✅ User data organized under `data/` directory
- ✅ Automatic backup of old v2 configs
- ✅ GitHub token optionally stored in OS keychain

**Environment Variable Override:**
You can set a custom favorites directory:
```sh
export TERA_FAVORITE_PATH="/path/to/your/favorites"
```

## Troubleshooting

### No sound?
- Ensure `mpv` is installed: `mpv --version`
- Check your system audio settings
- Try playing a test stream: `mpv https://stream.example.com`

### Station won't play?
- Some streams may be temporarily offline
- Try another station
- Check if the station works in a web browser

### Stop stuck playback
```sh
pkill mpv
```

### Multiple stations playing at once
TERA should prevent this, but if it happens:
```sh
killall mpv
# or on Linux
pkill -9 mpv
```

### Can't find My-favorites.json
TERA creates it automatically at first launch. Check the favorites directory in your OS-specific config location (see [File Locations](#file-locations-v3)).

If missing, restart TERA and it will be created.

### Config migration issues
If automatic migration fails:

```sh
# Check migration status
tera config migrate

# Reset to defaults if needed
tera config reset
```

Your favorites and user data are never touched during migration.

## Development

### Requirements
- Go 1.21+
- mpv

### Run from source
```sh
git clone https://github.com/shinokada/tera.git
cd tera/v3
go run cmd/tera/main.go
```

### Test
```sh
go test ./... -v
```

### Build
```sh
cd v3
go build -o tera cmd/tera/main.go
```

## Contributing

Contributions are welcome! Please:
1. Open an issue to discuss proposed changes
2. Fork the repository
3. Create a feature branch
4. Submit a pull request

## License

[MIT](LICENSE) © Shinichi Okada

## Links

- [GitHub Repository](https://github.com/shinokada/tera)
- [Issue Tracker](https://github.com/shinokada/tera/issues)
- [Radio Browser](https://www.radio-browser.info/) - Station database
- [Migration Guide](v3/docs/MIGRATION.md) - Upgrading from v2 to v3
- [Changelog](CHANGELOG.md) - Version history
