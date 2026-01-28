# Tera

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
- Go 1.21+ (for building from source)

## Installation

```sh
go install github.com/shinokada/tera/cmd/tera@latest
```

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