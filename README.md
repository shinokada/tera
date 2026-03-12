# Tera: Terminal Radio
[REPO](https://github.com/shinokada/tera)

A terminal-based internet radio player powered by [Radio Browser](https://www.radio-browser.info/).

## Features

- 🔍 **Search** - Find stations by name, tag, language, country, or state
- 🎲 **I Feel Lucky** - Random station discovery by keyword
- 💾 **Favorites** - Organize stations into custom lists with duplicate detection
- ⭐ **Star Ratings** - Rate stations 1-5 stars and browse your top-rated collection
- 🏷️ **Custom Tags** - Tag stations with personal labels and build dynamic playlists
- ⚡ **Quick Play** - Direct playback from main menu (shortcuts 10-99+)
- 🕐 **Recently Played** - Last N stations shown below Quick Play Favorites in the main menu
- 🔊 **Playback Control** - Play/pause with persistent status, adjust volume, and mute during playback
- 🚫 **Block List** - Block unwanted stations from appearing in searches and auto-play
- ☁️ **Sync & Backup** - Export/restore local zip backups and sync all data via GitHub Gists
- 🗳️ **Voting** - Support your favorite stations on Radio Browser
- 🎨 **Themes** - Choose from predefined themes or customize via YAML config
- 💤 **Sleep Timer** - Set a timer to stop playback automatically
- 📊 **Most Played** - View your listening history sorted by play count, last played, or first played
- 🔄 **Update Checker** - Get notified when a new version is available
- ⌨️ **Keyboard-driven** - Full navigation without a mouse
- ❓ **Context Help** - Press `?` anytime to see available keyboard shortcuts
- 🖥️ **Command-Line Play** - Play stations directly from the terminal without opening the TUI

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
# 3) Most Played         - Your listening statistics
# 4) Top Rated           - Browse your highest-rated stations
# 5) Browse by Tag       - Browse stations by your custom tags
# 6) Tag Playlists       - Dynamic playlists from tag combinations
# 7) Manage Lists        - Create/edit/delete favorite lists
# 8) Block List          - Manage blocked stations
# 9) I Feel Lucky        - Random station by keyword
# 0) Sync & Backup       - Backup/restore data locally or via GitHub
# -) Settings            - Configure TERA

# Quick Play (from main menu):
# Type 10-99+ to instantly play stations from "My-favorites"

# Need help? Press ? anytime to see keyboard shortcuts!
```

## Command-Line Play

Play stations directly from the terminal without opening the TUI — useful for shell scripts, startup routines, or timed listening sessions.

```sh
tera play <source> [args] [--duration <duration>]
```

### Sources

| Full form | Short alias | Args | Description |
| --- | --- | --- | --- |
| `favorites` | `fav` | `[list-name] [n]` | Play nth station from a favorites list |
| `recent` | `rec` | `[n]` | Play the nth most recently played station |
| `top-rated` | `top` | `[n]` | Play the nth highest-rated station |
| `most-played` | `most` | `[n]` | Play the nth most-played station |
| `lucky` | — | `<keyword ...>` | Play a random station matching keyword(s) |

`[list-name]` defaults to `My-favorites`. `[n]` defaults to `1` (first item, 1-based).

### Examples

```sh
# Play the first station from My-favorites
tera play fav

# Play the 3rd station from the jazz list
tera play fav jazz 3

# Play the most recently played station
tera play rec

# Play the highest-rated station
tera play top

# Play the most-played station
tera play most

# Play a random station matching a keyword
tera play lucky ambient

# Multi-word keywords work too
tera play lucky smooth jazz

# Stop automatically after 30 minutes
tera play fav --duration 30m

# Play for 1 hour then stop
tera play lucky ambient --duration 1h
```

### Duration

The optional `--duration` flag accepts Go duration format: `30s`, `10m`, `1h`, `1h30m`. Without it, playback continues until `Ctrl+C`.

### Status Line

A single line is printed when playback starts:

```text
▶ Playing: Jazz FM  [jazz · item 1 of 12]  (Ctrl+C to stop)
▶ Playing: Jazz FM  [jazz · item 1 of 12]  (stops in 30m · Ctrl+C to stop early)
```

### Notes

- Requires `mpv` to be installed (same as the TUI)
- CLI play sessions update Recently Played and Most Played history in the TUI
- `lucky` requires network access to query Radio Browser; all other sources are local
- Run `tera play --help` for full usage

---

## Main Features

### Play from Favorites

Browse and play stations from your organized lists. Navigate with `↑↓` or `jk`, press `Enter` to play. Press `/` to filter stations by name.

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

Block unwanted stations to prevent them from appearing in shuffle mode and, by default, in search results.

**How to Block:**
- While playing any station, press `b` to block it instantly
- Press `u` within 5 seconds to undo (in case of accidental block)
- Works in Search, I Feel Lucky, and Play from Favorites

**Block List Management:**
From main menu, select "8. Block List" to:
- **1. View Blocked Stations** — List all blocked stations; press `u` to unblock, `c` to clear all
- **2. Manage Block Rules** — Block entire countries, languages, or tags at once
- **3. Import/Export Blocklist** — Backup and restore your blocklist
- **4. Search Visibility** — Control whether blocked stations appear in search results

**Search Visibility (default: hidden):**
By default, blocked stations are completely hidden from search results. To change this:
1. Go to **Block List → 4. Search Visibility**
2. Press `y` to show blocked stations in search (marked with 🚫)
3. Press `n` to hide them again (default)

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
- Press `0` to return to the main menu, or any other key to exit TERA

**Keyboard Shortcuts:**

| Screen  | Key | Action                         |
| ------- | --- | ------------------------------ |
| Playing | `Z` | Open sleep timer dialog        |
| Playing | `+` | Extend running timer by 15 min |

### I Feel Lucky

Enter a keyword (genre, mood, style) and TERA finds a random matching station. Perfect for music discovery!

**Input Focus**: The page has two focusable areas — `Genre/keyword` (default) and `Choose an option` (history navigation). Use `Tab` to toggle between them:
- `▶ Genre/keyword:` is highlighted when active — type freely, including keywords starting with digits (e.g. `80s`, `2pac`)
- `▶ Choose an option:` is highlighted when active — use `↑↓`/`jk` to navigate history, or `1-N` shortcuts to pick a recent search

**Shuffle Mode**: Enable shuffle mode to automatically cycle through multiple stations matching your keyword:
- Press `Ctrl+T` to toggle shuffle on/off
- Stations play in random order without repeats
- Optional auto-advance timer (configurable)
- Navigate backward through recently played stations
- Configure shuffle behavior in Settings → Shuffle Settings

See [Shuffle Mode](#shuffle-mode) for detailed features.

### Settings

Access app configuration from the main menu (Settings: `-`):

- **Theme / Colors** - Switch between predefined themes or customize colors
- **Appearance** - Customize header display (text, ASCII art, alignment, colors, padding)
- **Connection Settings** - Auto-reconnect and buffering for unstable networks (4G/GPRS)
- **Shuffle Settings** - Configure shuffle mode behavior (auto-advance, history size)
- **History** - Search history and Recently Played display settings (size, display rows, reset)
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

### Quick Play & Recently Played

The main menu shows two instant-access sections below the regular menu:

```text
Main Menu & Quick Play

Choose an option:

  1. Play from Favorites
  ...
  -. Settings

─── Quick Play Favorites ───
  10. Jazz FM • UK • MP3 192kbps
  11. BBC Radio 6 Music • UK • AAC 128kbps

─── Recently Played ───
  12. WBGO Jazz 88.3 • United States • 3 minutes ago
  13. FIP • France • 1 hour ago
  14. Radio Swiss Jazz • Switzerland • Yesterday

Type 10-14 to play instantly!
```

**Quick Play Favorites:** Stations from "My-favorites.json" with shortcuts starting at 10.

**Recently Played:** Your last N stations (default 5), shown in most-recently-played order. Shortcuts continue from where Quick Play Favorites end.

**How it works:**
- Type the shortcut number and press Enter, or navigate with `↑↓` and press Enter
- Station plays immediately via the shared player
- Press `Esc` to stop playback
- The `▶` indicator marks the currently playing station in both sections

**Configure Recently Played:**
1. Press `-` from the main menu → Settings
2. Select **5. History → 2. Play History**
3. Available options:
   - **Toggle Show** — enable or disable the section entirely
   - **History Size** — how many stations to track (1–20, default 5)
   - **Display Rows** — cap the number of rows shown at once (1–10; `0` = fill available space)
   - **Reset All Play Stats** — clears play counts, Most Played, and Recently Played

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

| Key      | Action                                         |
| -------- | ---------------------------------------------- |
| `Ctrl+T` | Toggle shuffle mode (in input screen)          |
| `Tab`    | Switch focus: Genre/keyword ↔ Choose an option |
| `n`      | Next shuffle station (manual skip)             |
| `[`      | Previous station (from history)                |
| `b`      | Block current station                          |
| `u`      | Undo block (5 sec window)                      |
| `p`      | Pause/resume auto-advance timer                |
| `h`      | Stop shuffle, keep playing current             |
| `f`      | Save to My-favorites                           |
| `s`      | Save to another list                           |
| `v`      | Vote for station                               |
| `r`      | Rate station (then 1-5 / 0)                    |
| `Esc`    | Stop shuffle and return to input               |

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

### Configuration File

Shuffle settings are stored in the config directory as `shuffle.yaml`:

```yaml
shuffle:
  auto_advance: true           # Auto-advance enabled
  interval_minutes: 5          # 5 minutes per station
  remember_history: true       # Track history
  max_history: 5               # Remember last 5 stations
```

File location:
- Linux: `~/.config/tera/shuffle.yaml`
- macOS: `~/Library/Application Support/tera/shuffle.yaml`
- Windows: `%APPDATA%\tera\shuffle.yaml`

You can edit this file directly or use the Settings menu.

## Keyboard Shortcuts

### Global Navigation

| Key         | Action        |
| ----------- | ------------- |
| `↑↓` / `jk` | Navigate      |
| `g` / `G`   | Top / End     |
| `Enter`     | Select / Play |
| `Esc`       | Back / Stop   |
| `0`         | Main Menu     |
| `?`         | Help          |
| `Ctrl+C`    | Quit          |

### Main Menu

| Key      | Action                                         |
| -------- | ---------------------------------------------- |
| `0`      | Sync & Backup                                  |
| `1-9`    | Quick select menu item                         |
| `10-99+` | Quick play from My-favorites / Recently Played |
| `-`      | Settings                                       |

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

### Favorites Station List

| Key | Action                  |
| --- | ----------------------- |
| `/` | Filter stations by name |
| `d` | Delete station          |

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

## Sync & Backup

Back up and sync your data locally or across devices using zip archives and GitHub Gists.

### Export Backup (zip)

Save a local copy of your data with no GitHub account required.

1. From the main menu press `0` → **Sync & Backup**
2. Select **7. Export backup (zip)**
3. Choose which categories to include (favorites, ratings, tags, etc.)
4. Confirm the save path (default: `~/tera-backup-YYYY-MM-DD.zip`)

### Restore from Backup (zip)

1. Select **8. Restore from backup (zip)**
2. Enter the path to your zip file
3. Choose which categories to restore
4. Confirm — you will be warned before any existing files are overwritten

### Sync to Gist

Push all selected data to a dedicated secret GitHub Gist (`tera-data-backup`).

**Quick Setup:**
1. Go to **0) Sync & Backup → Token Management**
2. Create a GitHub Personal Access Token with `gist` scope
3. Paste the token in TERA
4. Select **9. Sync all data to Gist** and choose categories

### Restore from Gist

1. Select **10. Restore all data from Gist**
2. TERA fetches the `tera-data-backup` Gist and shows available categories
3. Choose what to restore — you will be warned before overwriting

### Sync Categories

| Category                | Default |
| ----------------------- | ------- |
| Favorites (playlists)   | ✅ on    |
| Settings (config.yaml)  | ✅ on    |
| Ratings & votes         | ✅ on    |
| Blocklist               | ✅ on    |
| Station metadata & tags | ✅ on    |
| Search history          | ❌ off   |

Category selections are saved in `sync_prefs.json` and reused on the next run.

**Documentation:**
- [Gist Setup Guide](docs/GIST_SETUP.md) - Token setup and security
- [Gist Management Guide](docs/GIST_CRUD_GUIDE.md) - Complete feature guide

## File Locations

TERA stores its configuration files in the OS-standard config directory:

| Operating System | Location                              |
| ---------------- | ------------------------------------- |
| **Linux**        | `~/.config/tera/`                     |
| **macOS**        | `~/Library/Application Support/tera/` |
| **Windows**      | `%APPDATA%\tera\`                     |

### Configuration Files

```text
tera/
├── config.yaml             # Unified configuration (all settings)
├── data/
│   ├── blocklist.json          # Blocked radio stations
│   ├── voted_stations.json     # Voting history
│   ├── station_metadata.json   # Play count & listening history
│   ├── station_ratings.json    # Star ratings
│   ├── station_tags.json       # Custom tags and tag playlists
│   ├── favorites/
│   │   ├── My-favorites.json   # Quick play list (main menu 10+)
│   │   ├── Rock.json
│   │   └── Jazz.json
│   └── cache/
│       ├── gist_metadata.json
│       └── search-history.json
└── .v2-backup-YYYYMMDD-HHMMSS/ # Automatic v2 config backup
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
go run ./cmd/tera/
```

### Test
```sh
go test ./... -v
```

### Build
```sh
go build -o tera ./cmd/tera/
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
