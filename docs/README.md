# TERA - Terminal Radio

<p align="center">
<img width="600" src="https://raw.githubusercontent.com/shinokada/tera/main/images/tera-3.png" />
<br />
<a href="https://tera.codewithshin.com/">https://tera.codewithshin.com/</a>
</p>

**Version 0.7.0**

A modern, interactive radio player for your terminal with 50,000+ stations worldwide.

[![Website](https://img.shields.io/badge/website-tera.codewithshin.com-blue)](https://tera.codewithshin.com/)
[![License](https://img.shields.io/badge/license-MIT-green)](LICENSE)

---

## Overview

TERA (TErminal RAdio) is an interactive music radio player that brings the world's radio stations to your command line. Discover new music, manage favorite lists, and explore 35,000+ stations—all without leaving your terminal.

**Why TERA?**
- 🎵 Escape algorithmic recommendations and discover real variety
- 🚫 No ads, no subscriptions, completely free
- ⚡ Fast, keyboard-driven interface with arrow key navigation
- 📻 Access to global radio stations via Radio Browser API
- 💾 Save and organize your favorite stations

---

## Features

### Core Features
- **35,000+ Radio Stations** powered by [Radio Browser API](https://de1.api.radio-browser.info/)
- **Arrow Key Navigation** throughout the entire application
- **Quick Play Favorites** - Top 10 favorites accessible from main menu
- **Duplicate Detection** - Prevents saving the same station twice
- **Smart Search** - Find stations by tag, name, language, country, or state
- **List Management** - Create, edit, and delete custom station lists
- **Gist Integration** - Share your favorite lists via GitHub Gist
- **I Feel Lucky** - Random station discovery mode

### User Experience
- **Modern Interface** - fzf-powered fuzzy search and selection
- **Seamless Navigation** - Return to main menu from anywhere with `0` or ESC
- **Interactive Menus** - All menus support arrow keys and keyboard shortcuts
- **No Setup Required** - Works immediately after installation

---

## Requirements

- **Unix-like environment** (Linux, macOS, BSD)
- [mpv](https://mpv.io/) - Media player
- [jq](https://stedolan.github.io/jq/) - JSON processor
- [fzf](https://github.com/junegunn/fzf) - Fuzzy finder
- [wget](https://www.gnu.org/software/wget/) - Network downloader
- [git](https://git-scm.com/) - For Gist features (optional)

---

## Installation

### Using Awesome Package Manager

```bash
awesome install shinokada/tera
```

### Homebrew (macOS/Linux)

```bash
brew tap shinokada/tera
brew install tera
```

### Debian/Ubuntu

Download from [releases page](https://github.com/shinokada/tera/releases):

```bash
sudo apt install ./tera_0.7.0_all.deb
rm ./tera_0.7.0_all.deb
```

### Arch Linux

```bash
# Available on AUR
yay -S tera
```

See [aur.archlinux.org/packages/tera](https://aur.archlinux.org/packages/tera)

### Verify Installation

Test that mpv is working:

```bash
mpv https://live.musopen.org:8085/streamvbr0
```

If music plays, you're ready to go!

---

## Quick Start

### Launch TERA

```bash
tera
```

You'll see the main menu with options to:
1. Play from your lists
2. Search for stations
3. Manage lists
4. Delete stations
5. Try "I Feel Lucky" mode
6. Upload/recover lists via Gist

### First-Time Users

1. **Try the defaults**: TERA comes with sample stations in "My Favorites"
2. **Search for music**: Select `2) Search radio stations`
3. **Save favorites**: After playing, save stations to your lists
4. **Quick access**: Saved stations appear in main menu for instant playback

### Navigation Quick Reference

| Context     | Action        | Navigation       |
| ----------- | ------------- | ---------------- |
| Any menu    | Arrow keys ↑↓ | Navigate options |
| Any menu    | Enter         | Select option    |
| Any menu    | ESC           | Go back          |
| Text prompt | Type `0`      | Go back          |
| Text prompt | Type `00`     | Main menu        |
| Text prompt | Empty + Enter | Go back          |

**See [Navigation Guide](NAVIGATION_GUIDE.md) for complete details.**

---

## Usage

### Searching for Stations

```bash
# From main menu
2) Search radio stations

# Search by:
- Tag (jazz, rock, classical, etc.)
- Name (station name)
- Language
- Country code
- State

# After search:
- Use arrow keys to select
- Press Enter to play
- Save to "My Favorites" or custom lists
```

### Managing Lists

```bash
# From main menu
3) List (Create/Read/Update/Delete)

# Operations:
- Create new lists
- View list contents
- Rename lists
- Delete lists
- Protected: "My Favorites" cannot be deleted/renamed
```

**See [List Navigation Guide](LIST_NAVIGATION_GUIDE.md) for details.**

### Quick Play Favorites

Your "My Favorites" list appears on the main menu:

```
--- Quick Play Favorites ---
10) ▶ BBC World Service
11) ▶ Jazz FM
12) ▶ Classical KDFC
```

Just select the number to play instantly!

**See [Favorites Guide](FAVORITES.md) for advanced usage.**

### Gist Integration

Share and backup your station lists:

```bash
# From main menu
6) Gist

# Features:
1) Create a gist - Upload all your lists
2) Recover from gist - Download lists from URL
```

**Setup required**: See [Gist Setup Guide](GIST_SETUP.md)

### Player Controls

While playing:

| Key           | Action                      |
| ------------- | --------------------------- |
| `p` / `SPACE` | Pause/unpause               |
| `q`           | Stop and quit               |
| `9` / `0`     | Volume down/up              |
| `/` / `*`     | Volume down/up              |
| `m`           | Mute                        |
| `[` / `]`     | Decrease/increase speed 10% |
| `{` / `}`     | Halve/double speed          |

See [MPV manual](https://mpv.io/manual/master/) for all controls.

---

## Documentation

- **[Navigation Guide](NAVIGATION_GUIDE.md)** - Complete navigation reference
- **[List Navigation Guide](LIST_NAVIGATION_GUIDE.md)** - List management details
- **[Favorites Guide](FAVORITES.md)** - Quick play favorites setup
- **[Gist Setup](GIST_SETUP.md)** - GitHub Gist integration
- **[Changelog](CHANGELOG.md)** - Recent updates and features

---

## File Locations

```
~/.config/tera/          # Configuration directory
├── favorite/            # Your favorite lists
│   ├── myfavorites.json # "My Favorites" list
│   └── *.json          # Custom lists
└── gisturl             # Last gist URL (if used)

~/.cache/tera/           # Temporary files
└── radio_searches.json  # Search results cache
```

---

## Uninstallation

### Using Script

```bash
curl -s https://raw.githubusercontent.com/shinokada/tera/main/uninstall.sh | bash
```

### Manual Removal

```bash
# Remove executable
rm $(which tera)

# Remove configuration and cache
rm -rf ~/.config/tera
rm -rf ~/.cache/tera
```

---

## Options

```bash
tera              # Start TERA
tera --version    # Show version
tera -h           # Show help
tera --help       # Show help
```

---

## Tips & Tricks

1. **Fast Navigation**: Type `0` in any text prompt to go back
2. **Emergency Exit**: Use `Ctrl+C` to force quit
3. **No Duplicates**: TERA prevents saving the same station twice
4. **ESC is Safe**: Press ESC anywhere without fear—it just goes back
5. **Quick Access**: Add stations to "My Favorites" for main menu access
6. **Fuzzy Search**: In any list, start typing to filter results

---

## Troubleshooting

### Station Won't Play

```bash
# Test mpv directly
mpv https://live.musopen.org:8085/streamvbr0

# Check if station URL is valid
# Some stations may be offline or have changed URLs
```

### Navigation Not Working

```bash
# Ensure fzf is installed
which fzf

# Update fzf if needed
brew upgrade fzf  # macOS
```

### Gist Features Not Working

See [Gist Setup Guide](GIST_SETUP.md) for GitHub token configuration.

---

## Contributing

Contributions welcome! Please:
1. Fork the repository
2. Create a feature branch
3. Test your changes thoroughly
4. Submit a pull request

---

## Support

- 📖 **Documentation**: [tera.codewithshin.com](https://tera.codewithshin.com/)
- 🐛 **Issues**: [GitHub Issues](https://github.com/shinokada/tera/issues)
- 💬 **Discussions**: [GitHub Discussions](https://github.com/shinokada/tera/discussions)

---

## Acknowledgments

- **Radio Browser API** - Station database
- **MPV** - Media playback
- **fzf** - Fuzzy finding interface

---

## License

MIT License - See [LICENSE](LICENSE) file for details.

---

## Author

**Shinichi Okada** ([@shinokada](https://github.com/shinokada))

---

**Made with ❤️ for music lovers who live in the terminal**
