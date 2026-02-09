# Tera: Terminal Radio

A terminal-based internet radio player powered by [Radio Browser](https://www.radio-browser.info/).

## Features

- 🔍 **Search** - Find stations by name, tag, language, country, or state
- 🎲 **I Feel Lucky** - Random station discovery by keyword
- 💾 **Favorites** - Organize stations into custom lists with duplicate detection
- ⚡ **Quick Play** - Direct playback from main menu (shortcuts 10-99+)
- 🔊 **Playback Control** - Play/pause with persistent status, adjust volume, and mute during playback
- 🚫 **Block List** - Block unwanted stations from appearing in searches and auto-play
- ☁️ **Gist Sync** - Backup and restore favorites via GitHub Gists
- 🗳️ **Voting** - Support your favorite stations on Radio Browser
- 🎨 **Themes** - Choose from predefined themes or customize via YAML config
- 🔄 **Update Checker** - Get notified when a new version is available
- ⌨️ **Keyboard-driven** - Full navigation without a mouse
- ❓ **Context Help** - Press `?` anytime to see available keyboard shortcuts

## Requirements

- [mpv](https://mpv.io/) - Media player for audio playback

## Installation

### Homebrew (macOS/Linux)
```sh
# update and upgrade
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

## Quick Start

```sh
# Start TERA
tera

# Main Menu Options:
# 1) Play from Favorites - Browse your saved lists
# 2) Search Stations     - Find new stations
# 3) Manage Lists        - Create/edit/delete favorite lists
# 4) Block List          - Manage blocked stations
# 5) I Feel Lucky        - Random station by keyword
# 6) Gist Management     - Backup/restore via GitHub
# 7) Settings            - Configure TERA

# Quick Play (from main menu):
# Type 10-99+ to instantly play stations from "My-favorites"

# Need help? Press ? anytime to see keyboard shortcuts!
```

## Main Features

### Play from Favorites

Browse and play stations from your organized lists. Navigate with `↑↓` or `jk`, press `Enter` to play.

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
From main menu, select "4. Block List" to:
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
- Linux: `~/.config/tera/blocklist.json`
- macOS: `~/Library/Application Support/tera/blocklist.json`
- Windows: `%APPDATA%\tera\blocklist.json`

### I Feel Lucky

Enter a keyword (genre, mood, style) and TERA finds a random matching station. Perfect for music discovery!

**Shuffle Mode**: Enable shuffle mode to automatically cycle through multiple stations matching your keyword:
- Press `t` to toggle shuffle on/off
- Stations play in random order without repeats
- Optional auto-advance timer (configurable)
- Navigate backward through recently played stations
- Configure shuffle behavior in Settings → Shuffle Settings

See [Shuffle Mode](#shuffle-mode) for detailed features.

### Settings

Access app configuration from the main menu (option 7):

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
- **Padding** - Top and bottom spacing (0-5 lines)

**Tips:**
- Preview changes before saving
- Use [TAAG](https://patorjk.com/software/taag/) or `figlet` to generate ASCII art
- Settings stored in the config directory (see [File Locations](#file-locations))

### Connection Settings

For users on unstable networks (mobile data, GPRS, 4G), configure automatic reconnection:

- **Auto-reconnect** - Automatically retry when stream drops (default: enabled)
- **Reconnect delay** - Wait time between attempts: 1-30 seconds (default: 5s)
- **Stream buffer** - Cache size to handle brief signal drops: 10-200 MB (default: 50MB)

Settings stored in the config directory (see [File Locations](#file-locations)).

### Quick Play from Main Menu

The main menu shows your "My-favorites" list with shortcuts 10-99+. Type the number to play instantly:

```
Main Menu & Quick Play

Choose an option:

  1. Play from Favorites
  2. Search Stations
  3. Manage Lists
  4. Block List
  5. I Feel Lucky
  6. Gist Management
  7. Settings

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

1. Press `7` from the main menu to open Settings
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

You can also customize colors and padding by editing the theme config file:

```sh
tera theme path   # Show config file location
tera theme reset  # Reset to defaults
```

The config file includes an ANSI color reference (0-15 standard colors, 16-255 extended colors). Example:

```yaml
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

1. Press `7` from the main menu to open Settings
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

1. Navigate to **I Feel Lucky** from the main menu (option 5)
2. Press `t` to toggle shuffle mode on
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
| `t`   | Toggle shuffle mode (in input screen) |
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

f: Fav • s: List • v: Vote • n: Next • [: Prev • b: Block • p: Pause timer • h: Stop shuffle
```

### Configuration File

Shuffle settings are stored in the config directory as `shuffle.yaml`:

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
| `1-7`    | Quick select menu item       |
| `10-99+` | Quick play from My-favorites |

### Playback Controls

| Key     | Action            |
| ------- | ----------------- |
| `Space` | Pause / Resume    |
| `*`     | Volume up (+5%)   |
| `/`     | Volume down (-5%) |
| `m`     | Toggle mute       |
| `b`     | Block station     |
| `u`     | Undo block (5s)   |

### Playing/Browsing Stations

| Key | Action               |
| --- | -------------------- |
| `f` | Save to My-favorites |
| `s` | Save to another list |
| `v` | Vote for station     |

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
1. Go to: Main Menu → 6) Gist Management → 6) Token Management
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
- [Gist Setup Guide](GIST_SETUP.md) - Token setup and security
- [Gist Management Guide](GIST_CRUD_GUIDE.md) - Complete feature guide

## File Locations

TERA stores its configuration files in the OS-standard config directory:

| Operating System | Location                              |
| ---------------- | ------------------------------------- |
| **Linux**        | `~/.config/tera/`                     |
| **macOS**        | `~/Library/Application Support/tera/` |
| **Windows**      | `%APPDATA%\tera\`                     |

### Configuration Files

```
tera/
├── theme.yaml              # Color and padding customization
├── appearance_config.yaml  # Header customization (text, ASCII art, etc.)
├── connection_config.yaml  # Auto-reconnect and buffering settings
├── shuffle.yaml            # Shuffle mode settings
├── blocklist.json          # Blocked radio stations
├── voted_stations.json     # Voting history (prevents duplicate votes)
├── gist_metadata.json      # Your gist history
├── tokens/
│   └── github_token        # GitHub Personal Access Token
└── favorites/
    ├── My-favorites.json   # Quick play list (main menu 10+)
    ├── Rock.json           # Your custom lists
    └── Jazz.json
```

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
TERA creates it automatically at first launch. Check the favorites directory in your OS-specific config location (see [File Locations](#file-locations)).

If missing, restart TERA and it will be created.

## Development

### Requirements
- Go 1.21+
- mpv

### Run from source
```sh
git clone https://github.com/shinokada/tera.git
cd tera
go run cmd/tera/main.go
```

### Test
```sh
go test ./... -v
```

### Build
```sh
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
