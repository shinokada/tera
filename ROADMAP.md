# TERA Development Roadmap

This document outlines the development plan for TERA, a terminal-based internet radio player.

## Version Strategy

Starting with v3.0.0, TERA uses the **folder strategy** for versioning:
- v2.x lives in the `v2` branch (maintenance mode)
- v3.x+ lives in the `v3/` folder on the `main` branch

### Why the Change?

The folder strategy provides:
- All versions visible in one branch
- Easier development workflow
- Cleaner CI/CD setup
- Standard Go module practice for v3+

## Version Support Policy

| Version  | Status             | Support Level            | Timeline         |
| -------- | ------------------ | ------------------------ | ---------------- |
| **v3.x** | Active Development | New features + bug fixes | Now - Ongoing    |
| **v2.x** | Maintenance Mode   | Bug fixes only           | Until March 2026 |
| **v1.x** | Archived           | No support               | -                |

### What Goes Where?

**v2 Branch (Maintenance Mode)**
- ✅ Critical bug fixes
- ✅ Security patches
- ❌ New features
- ❌ Breaking changes

**v3 Folder (Active Development)**
- ✅ All bug fixes
- ✅ New features
- ✅ Breaking changes (with major version bump)
- ✅ Experimental features

## Release Timeline

### v3.0.x Series

**v3.0.0** - Migration Release ✅
- ✅ Create v3/ folder
- ✅ Update imports to v3
- Build unified config system
- Release v3.0.0 with unified config
- Auto-migration from v2

**v3.0.1-v3.0.x** - Bug Fixes
- Address any issues found in v3.0.0
- Backport critical fixes to v2 as needed

---

**Current state (v2):**
```
os.UserConfigDir()
├── theme.yaml
├── appearance_config.yaml
├── connection_config.yaml
├── shuffle.yaml
├── blocklist.json
└── tokens/
```

**New state (v3.x.x):**
```
os.UserConfigDir()
├── config.yaml          # Everything in one file!
├── favorites/
└── gist_metadata.json
```

---

### v3.1.x Series - Station Metadata 

**v3.1.0** - Play Count & Last Played
- Track how many times you've played each station
- Show when you last played a station
- "Most Played" list view

---

### v3.2.x Series - User Ratings

**v3.2.0** - Star Ratings
- Rate stations 1-5 stars
- Filter/sort by rating
- "Top Rated" list view

---

### v3.3.x Series - Custom Tags

**v3.3.0** - Personal Tags
- Add custom tags to stations (e.g., "workout", "coding", "relaxing")
- Filter stations by custom tags
- Tag-based playlists

---

### v4.0.x Series - Library/SDK Mode (Late 2026 or 2027)

**v4.0.0** - Public API 🚀
- Expose TERA as a Go library
- CLI still works exactly the same
- Enable embedding in other apps

**Public API:**
```go
package tera

type Client struct { ... }

func New(opts ...Option) (*Client, error)
func (c *Client) Search(ctx context.Context, query string) ([]Station, error)
func (c *Client) Play(station Station) error
func (c *Client) Favorites() *FavoritesManager
```

**Use cases:**
- Web dashboard for TERA
- Discord bots ("!play jazz")
- Home automation integration
- Mobile apps using TERA backend


---

### v5.0.x Series - Advanced Features

Ideas for future exploration:

**Playlist Engine**
- Time-based scheduling
- Auto-advance with transitions
- Fade between stations
- Queue management

**Plugin System**
- Discord Rich Presence
- Last.fm scrobbling
- Custom notifications
- Statistics tracking

**Multi-user/Cloud Sync**
- Beyond GitHub Gists
- Real-time sync
- Collaborative playlists

---

## Migration Guide for Users

### From v2 to v3

**Installation:**
```bash
# Uninstall v2 (optional)
go clean -i github.com/shinokada/tera/v2/cmd/tera

# Install v3
go install github.com/shinokada/tera/v3/cmd/tera@latest
```

**Data Migration:**
- All favorites, blocklists, and settings are automatically migrated
- v2 config files are preserved (not deleted)
- First run of v3 will create new config structure

**What Changes:**
- v3.0: Config file structure (auto-migrated)
- v3.1: New features only (no breaking changes)

---

## Contributing

Want to help? Here are areas we'd love contributions:

### High Priority
- Bug fixes for both v2 and v3
- Documentation improvements
- Testing on different platforms

### Medium Priority
- Feature requests for v3.x series
- UI/UX improvements
- Performance optimizations

### Future
- Plugin development (when v5 arrives)
- Web frontend (when v4 API is ready)

---

## Questions?

- **GitHub Issues:** https://github.com/shinokada/tera/issues
- **Discussions:** https://github.com/shinokada/tera/discussions

---

**Last Updated:** February 2026  
**Current Version:** v3.0.0  
**Next Release:** v3.1.0 (Unified Config)
