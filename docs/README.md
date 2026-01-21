# TERA Documentation

**Version 0.7.1+** - Terminal Radio Player

---

## Quick Links

- 🚀 [Quick Start](#quick-start)
- 📚 [User Guides](#user-guides)
- 🔧 [Installation](#installation)
- ⚙️ [Configuration](#configuration)

---

## Quick Start

```bash
# Install TERA
awesome install shinokada/tera

# Launch
tera

# Navigate with arrow keys
# Play stations from "My Favorites"
# Search for new stations
```

**First time?** TERA comes ready to use with sample stations.

---

## User Guides

### Essential Guides

- **[Navigation Guide](NAVIGATION_GUIDE.md)** - How to navigate TERA (arrow keys, shortcuts, menus)
- **[List Navigation Guide](LIST_NAVIGATION_GUIDE.md)** - Managing your station lists
- **[Favorites Guide](FAVORITES.md)** - Setting up quick-play favorites

### Feature Guides

- **[Gist Setup](GIST_SETUP.md)** - Backup and share lists via GitHub
- **[Token Management](TOKEN_MANAGEMENT.md)** - Manage GitHub tokens securely
- **[Gist CRUD Guide](GIST_CRUD_GUIDE.md)** - Complete gist management (create, view, update, delete)
- **[Update Gist Quick Guide](UPDATE_GIST_QUICK_GUIDE.md)** - How to update gist descriptions
- **[Gist Quick Reference](GIST_QUICK_REFERENCE.md)** - One-page gist cheatsheet

### Reference

- **[Changelog](CHANGELOG.md)** - Version history and updates

---

## Installation

### Quick Install

```bash
# Awesome Package Manager
awesome install shinokada/tera

# Homebrew
brew tap shinokada/tera && brew install tera

# Debian/Ubuntu
# Download from releases, then:
sudo apt install ./tera_*.deb
```

### Requirements

- mpv (media player)
- jq (JSON processor)
- fzf (fuzzy finder)
- wget (downloader)
- git (optional, for gist features)

---

## Configuration

### File Locations

```text
~/.config/tera/
├── favorite/
│   ├── My-favorites.json    # Your favorites
│   └── *.json              # Custom lists
├── tokens/
│   └── github_token         # GitHub token (secure storage)
└── gist_metadata.json       # Saved gists (auto-created)

~/.cache/tera/
└── radio_searches.json      # Search cache
```

### GitHub Token Setup (For Gist Features)

**Interactive Setup (Recommended):**
1. Run TERA
2. Select `6) Gist`
3. Select `1) Token Management`
4. Choose `1) Setup/Change Token`
5. Follow the prompts

Token is stored securely in `~/.config/tera/tokens/github_token`

See [Token Management Guide](TOKEN_MANAGEMENT.md) for complete details.

---

## Features Overview

### Core Features

- **50,000+ Stations** - Global radio via Radio Browser API
- **Arrow Key Navigation** - Intuitive fzf-powered interface
- **Quick Play** - Top 10 favorites on main menu
- **Smart Search** - By tag, name, language, country, state
- **List Management** - Create, rename, delete custom lists
- **Duplicate Detection** - Prevents saving stations twice
- **I Feel Lucky** - Random station discovery

### Gist Features

- **Create Gists** - Backup all lists to GitHub
- **My Gists** - View all your saved gists
- **Update Gists** - Change gist descriptions
- **Delete Gists** - Remove old backups
- **Quick Recovery** - Select gist by number or URL
- **Auto-Tracking** - Metadata saved locally

---

## Navigation Quick Reference

| Context                      | Action       | Key/Command                          |
| ---------------------------- | ------------ | ------------------------------------ |
| Any menu                     | Navigate     | Arrow keys ↑↓                        |
| Any menu                     | Select       | Enter                                |
| Any menu                     | Go back      | ESC                                  |
| Text prompt (search)         | Go back      | `0` or Empty + Enter                 |
| Text prompt (list mgmt)      | Go back      | `0` (empty shows error)              |
| Text prompt                  | Main menu    | `00`                                 |
| Playing                      | Pause/Resume | Space                                |
| Playing                      | Quit         | `q`                                  |
| Playing                      | Volume +/-   | `9` / `0`                            |

**Full details:** [Navigation Guide](NAVIGATION_GUIDE.md)

**Full details:** [Navigation Guide](NAVIGATION_GUIDE.md)

## Common Tasks

### Add a Station

1. Main Menu → `2) Search`
2. Choose search type
3. Select station → Play
4. Save to "My Favorites" or custom list

### Create a List

1. Main Menu → `3) List`
2. Select "Create new list"
3. Enter list name
4. Add stations via search

### Backup Lists

1. Main Menu → `6) Gist`
2. `1) Token Management` - Set up your GitHub token (first time only)
3. `2) Create a gist` - Backup your lists
4. Check `3) My Gists` to see your backups

### Manage Your Token

1. Main Menu → `6) Gist` → `1) Token Management`
2. Options:
   - `1) Setup/Change Token` - Add or update token
   - `2) View Current Token` - See masked token
   - `3) Validate Token` - Test if token works
   - `4) Delete Token` - Remove token securely

### Update Gist Description

1. Main Menu → `6) Gist`
2. Select `4) Update a gist`
3. Choose gist and enter new description

### Restore Lists

1. Main Menu → `6) Gist`
2. Select `3) Recover favorites`
3. Choose gist by number or enter URL

---

## Troubleshooting

### Station Won't Play

```bash
# Test mpv
mpv https://live.musopen.org:8085/streamvbr0
```

### Navigation Not Working

```bash
# Check fzf
which fzf

# Update if needed
brew upgrade fzf  # macOS
```

### Gist Issues

- **No token set up:** Go to `6) Gist` → `1) Token Management` → `1) Setup Token`
- **Token validation failed:** Run `3) Validate Token` to check status
- **Failed to create gist:** Verify token is valid, check internet connection
- **Can't recover:** Ensure gist URL is correct and token has access

**Full troubleshooting:** [Token Management](TOKEN_MANAGEMENT.md) | [Gist CRUD Guide](GIST_CRUD_GUIDE.md)

---

## Getting Help

- 📖 **Guides:** See list above
- 🐛 **Issues:** [GitHub Issues](https://github.com/shinokada/tera/issues)
- 💬 **Discussions:** [GitHub Discussions](https://github.com/shinokada/tera/discussions)
- 🌐 **Website:** [tera.codewithshin.com](https://tera.codewithshin.com/)

---

## License & Author

**License:** MIT  
**Author:** Shinichi Okada ([@shinokada](https://github.com/shinokada))

---

**Made with ❤️ for terminal music lovers** 🎵
