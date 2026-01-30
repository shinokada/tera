# Tera: Terminal Radio

A terminal-based internet radio player powered by [Radio Browser](https://www.radio-browser.info/).

## Features

- 🔍 **Search** - Find stations by name, tag, language, country, or state
- 🎲 **I Feel Lucky** - Random station discovery by keyword
- 💾 **Favorites** - Organize stations into custom lists with duplicate detection
- ⚡ **Quick Play** - Direct playback from main menu (shortcuts 10-99+)
- ☁️ **Gist Sync** - Backup and restore favorites via GitHub Gists
- 🗳️ **Voting** - Support your favorite stations on Radio Browser
- 🎨 **Themes** - Choose from predefined themes or customize via YAML config
- ⌨️ **Keyboard-driven** - Full navigation without a mouse

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

### Debian/Ubuntu
```sh
sudo dpkg -i tera_1.x.x_linux_amd64.deb
sudo apt-get install -f  # Install mpv dependency if needed
```

### Fedora/RHEL
```sh
sudo rpm -i tera_1.x.x_linux_amd64.rpm
```

### Windows Scoop
```sh
scoop bucket add shinokada https://github.com/shinokada/scoop-bucket
scoop install tera
```

### Manual Download

Download the latest binary for your platform from the [releases page](https://github.com/shinokada/tera/releases):

| Platform | Architecture  | File                             |
| -------- | ------------- | -------------------------------- |
| macOS    | Intel         | `tera_1.x.x_darwin_amd64.tar.gz` |
| macOS    | Apple Silicon | `tera_1.x.x_darwin_arm64.tar.gz` |
| Linux    | x86_64        | `tera_1.x.x_linux_amd64.tar.gz`  |
| Linux    | ARM64         | `tera_1.x.x_linux_arm64.tar.gz`  |
| Windows  | x86_64        | `tera_1.x.x_windows_amd64.zip`   |
| Windows  | ARM64         | `tera_1.x.x_windows_arm64.zip`   |

#### macOS/Linux

```sh
# Download and extract (example for macOS Apple Silicon)
tar -xzf tera_1.x.x_darwin_arm64.tar.gz

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
# 4) I Feel Lucky        - Random station by keyword
# 5) Gist Management     - Backup/restore via GitHub

# Quick Play (from main menu):
# Type 10-99+ to instantly play stations from "My-favorites"
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

### I Feel Lucky

Enter a keyword (genre, mood, style) and TERA finds a random matching station. Perfect for music discovery!

### Quick Play from Main Menu

The main menu shows your "My-favorites" list with shortcuts 10-99+. Type the number to play instantly:

```
Main Menu & Quick Play

Choose an option:

  1. Play from Favorites
  2. Search Stations
  3. Manage Lists
  4. I Feel Lucky
  5. Gist Management

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

1. Press `6` from the main menu to open Settings
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

You can also customize colors and padding by editing `~/.config/tera/theme.yaml`:

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

## Keyboard Shortcuts

### Global Navigation
| Key         | Action        |
| ----------- | ------------- |
| `↑↓` / `jk` | Navigate      |
| `Enter`     | Select / Play |
| `Esc`       | Back          |
| `Ctrl+C`    | Quit          |

### Main Menu
| Key      | Action                       |
| -------- | ---------------------------- |
| `1-6`    | Quick select menu item       |
| `10-99+` | Quick play from My-favorites |

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
1. Go to: Main Menu → 5) Gist Management → 6) Token Management
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

```
~/.config/tera/
├── theme.yaml              # Color and padding customization
├── gist_metadata.json      # Your gist history
├── tokens/
│   └── github_token        # GitHub Personal Access Token
└── favorites/
    ├── My-favorites.json   # Quick play list (main menu 10+)
    ├── Rock.json           # Your custom lists
    └── Jazz.json
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
TERA creates it automatically at first launch. Check:
```sh
ls -la ~/.config/tera/favorites/
```

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
