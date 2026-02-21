# Changelog

All notable changes to TERA will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [3.4.0] - 2026-02-20

### Added
- **Custom Tags** - Tag any station with personal labels (e.g., `late night coding`, `gym workout`, `chill vibes`)
- **Tag input** - Press `t` while playing to add a tag; press `T` to open the full Manage Tags dialog
- **Manage Tags dialog** - Toggle tags on/off with Space, add new tags inline, and save in one keystroke
- **Browse by Tag** - New menu option (5) to view all your tags and the stations under each one
  - Select a tag to see its stations; press `d` to delete a tag from all stations
- **Tag Playlists** - New menu option (6) to create dynamic playlists driven by tag combinations
  - Three-step creation wizard: name → select tags → choose match mode (any/all)
  - Edit and delete playlists; play all matching stations in sequence
- **Tag pills in station lists** - Tagged stations show `[tag]` pills inline in every list view (Favorites, Search, Most Played, Top Rated, Browse by Tag, Tag Playlists)
- **Live pill refresh** - Tag pills update immediately when you add or remove tags without reloading the list
- **Tag autocomplete** - The tag input field suggests matching tags as you type; Tab to accept, ↑↓ to navigate
- **Keyboard shortcuts** across all playback screens: `t` add tag, `T` manage tags
- All tags stored locally in `station_tags.json`; nothing is transmitted externally

### Changed
- Main menu items renumbered to accommodate Browse by Tag (5) and Tag Playlists (6)
  - Manage Lists → 7, Block List → 8, I Feel Lucky → 9, Gist Management → 0, Settings → `-` (hyphen key)

## [3.3.0] - 2026-02-18

### Added
- **Star Ratings** - Rate any station from 1-5 stars with `r` then `1-5` keys
- **Top Rated** - New menu option (5) to browse your highest-rated stations
- **Rating display** - Stars shown in search results, favorites, and playing screen
- **Filter by rating** - Filter Top Rated list by minimum star rating
- **Sort options** - Sort by rating (high/low), recently rated
- **Quick rating** - Press `r` then `1-5` to rate while playing
- **Remove rating** - Press `r` then `0` to clear a station's rating
- Station caching for Top Rated and Most Played to display full station info

### Changed
- Main menu reordered: Top Rated is now option 5, I Feel Lucky moved to 6
- Star rendering with spacing for better display in Warp terminal (`★ ★ ★ ★ ☆`)
- Top Rated page uses bottom-aligned footer for consistent UI

### Fixed
- Most Played and Top Rated now properly display station names from cached data
- Empty lines in Most Played list when station names were missing
- Track metadata display filtering out URL-like strings

## [3.0.1] - 2026-02-09

### Fixed
- Complete v2 to v3 data migration now properly moves user files to `data/` subdirectory
- `voted_stations.json` and `blocklist.json` now correctly migrate to `data/` directory
- `favorites/` directory now correctly migrates to `data/favorites/`
- `search-history.json` now correctly migrates to `data/cache/` directory
- Added `data/cache/` directory creation to ensure proper file organization
- Old v2 config files (`appearance_config.yaml`, etc.) now properly removed after migration

### Changed
- Directory structure now strictly follows v3 specification:
  - Config: `config.yaml` (root of tera directory)
  - User data: `data/blocklist.json`, `data/voted_stations.json`
  - Favorites: `data/favorites/*.json`
  - Cache: `data/cache/search-history.json`

## [3.0.0] - 2026-02-09

### Added
- Unified configuration system - all settings now in single `config.yaml` file
- Automatic migration from v2 to v3 configuration on first run
- New organized directory structure with `data/` subdirectory for user content
- Cache directory for temporary data (`data/cache/`)

### Changed
- Merged `theme.yaml`, `appearance_config.yaml`, `connection_config.yaml`, and `shuffle.yaml` into single `config.yaml`
- User data (favorites, blocklist, voted stations) moved to `data/` subdirectory
- Configuration file located via `os.UserConfigDir()` for proper cross-platform support

### Migration
- V2 configs automatically backed up to `.v2-backup-TIMESTAMP/` directory
- V2 config files removed after successful migration and backup
- User data preserved during migration
