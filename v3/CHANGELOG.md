# Changelog

All notable changes to TERA will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

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
