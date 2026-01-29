# Tera: Terminal Radio

A terminal-based internet radio player powered by [Radio Browser](https://www.radio-browser.info/).

## Features

- 🔍 **Search** - Find stations by name, tag, language, country, or state
- 🎲 **I Feel Lucky** - Random station discovery by keyword
- 💾 **Favorites** - Organize stations into custom lists
- ☁️ **Gist Sync** - Backup and restore favorites via GitHub Gists
- 🗳️ **Voting** - Support your favorite stations
- 🎨 **Themes** - Customizable colors via YAML config
- ⌨️ **Keyboard-driven** - Full navigation without a mouse

## Requirements

- [mpv](https://mpv.io/) - Media player for audio playback

## Installation

### Homebrew (macOS/Linux)

```sh
brew install shinokada/tera/tera
// if you already installed
brew upgrade shinokada/tera/tera
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

## Usage

```sh
tera              # Start the interactive player
tera --help       # Show help
tera --version    # Show version
```

## Theme Configuration

Customize colors and padding by editing `~/.config/tera/theme.yaml`:

```sh
tera theme path   # Show config file location
tera theme reset  # Reset to defaults
```

The config file includes an ANSI color reference (0-15 standard colors, 16-255 extended). Example:

```yaml
colors:
  primary: "6"      # Cyan
  highlight: "3"    # Yellow
  error: "9"        # Bright Red
```

## Keyboard Shortcuts

| Key         | Action                    |
| ----------- | ------------------------- |
| `↑↓` / `jk` | Navigate                  |
| `Enter`     | Select / Play             |
| `Esc`       | Back                      |
| `0`         | Main menu                 |
| `f`         | Save to Quick Favorites   |
| `s`         | Save to list              |
| `n`         | New list (in save dialog) |
| `v`         | Vote for station          |
| `d`         | Delete                    |
| `Ctrl+C`    | Quit                      |

## Search Guide

The **Search Radio Stations** menu offers multiple ways to find stations:

### Search Types

| Option                     | Description                                    | Example Query                       |
| -------------------------- | ---------------------------------------------- | ----------------------------------- |
| **Search by Tag**          | Find stations by genre/style tags              | `jazz`, `rock`, `news`, `classical` |
| **Search by Name**         | Find stations by their name                    | `BBC`, `NPR`, `KEXP`                |
| **Search by Language**     | Find stations broadcasting in a language       | `english`, `spanish`, `japanese`    |
| **Search by Country Code** | Find stations from a specific country          | `US`, `UK`, `FR`, `JP`              |
| **Search by State**        | Find stations from a state/region              | `California`, `Texas`, `Bavaria`    |
| **Advanced Search**        | Search both name AND tag fields simultaneously | `smooth jazz`, `classic rock`       |

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

In the results list:
- Use `↑↓` or `jk` to navigate
- Press `Enter` to view station details and play
- Type `/` to filter results locally

## Gist Sync

Backup and sync your favorite lists across devices using GitHub Gists. Setup via `Gist` menu or see [Gist Setup Guide](bash/docs/GIST_SETUP.md).

## Troubleshooting

**No sound?**
- Ensure `mpv` is installed: `mpv --version`
- Check your system audio settings

**Station won't play?**
- Some streams may be temporarily offline
- Try a different station

**Stop stuck playback:**
```sh
pkill mpv
```

## Development

### Stop all running stations

```sh
killall mpv
# or
pkill -9 mpv
```

### Test

```sh
go test ./... -v
```

## Contributing

Contributions are welcome! Please open an issue or submit a pull request.

## License

[MIT](LICENSE) © Shinichi Okada