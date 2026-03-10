package storage

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/shinokada/tera/v3/internal/gist"
)

const (
	// BackupGistDescription is the fixed description used to identify the
	// dedicated backup Gist across push, pull, and restore operations.
	BackupGistDescription = "tera-data-backup"
	backupGistHTTPTimeout = 30 * time.Second
	// backupGistMarkerFile is a sentinel file written into every backup Gist.
	// FindBackupGist requires its presence to accept a description-matched Gist,
	// preventing false positives from unrelated user Gists with the same description.
	backupGistMarkerFile = "tera-manifest.json"
)

// GistSyncManager handles Gist-based push and pull of user data.
// It maintains a single dedicated secret Gist identified by the description
// "tera-data-backup", separate from the per-playlist Gists managed elsewhere.
type GistSyncManager struct {
	client    *gist.Client
	configDir string
}

// NewGistSyncManager creates a GistSyncManager using the provided Gist client.
func NewGistSyncManager(client *gist.Client) (*GistSyncManager, error) {
	if client == nil {
		return nil, fmt.Errorf("gist client is required")
	}
	configDir, err := os.UserConfigDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get config directory: %w", err)
	}
	return &GistSyncManager{
		client:    client,
		configDir: filepath.Join(configDir, "tera"),
	}, nil
}

// gistFilename maps a config-relative file path to its Gist filename.
// GitHub Gist filenames cannot contain "/", so we encode the directory
// structure using a prefix convention.
//
//	config.yaml                       → config.yaml
//	data/blocklist.json               → blocklist.json
//	data/voted_stations.json          → voted_stations.json
//	data/station_ratings.json         → ratings.json
//	data/station_tags.json            → tags.json
//	data/station_metadata.json        → metadata.json
//	data/favorites/Jazz.json               → fav--Jazz.json
//	data/favorites/search-history.json    → search-history.json
func gistFilename(relPath string) string {
	slashPath := filepath.ToSlash(relPath)
	switch slashPath {
	case "config.yaml":
		return "config.yaml"
	case "data/blocklist.json":
		return "blocklist.json"
	case "data/voted_stations.json":
		return "voted_stations.json"
	case "data/station_ratings.json":
		return "ratings.json"
	case "data/station_tags.json":
		return "tags.json"
	case "data/station_metadata.json":
		return "metadata.json"
	case "data/favorites/" + SystemFileSearchHistory:
		return "search-history.json"
	}
	// data/favorites/Jazz.json → fav--Jazz.json
	if strings.HasPrefix(slashPath, "data/favorites/") {
		base := filepath.Base(relPath)
		return "fav--" + base
	}
	// Fallback: replace slashes with "--"
	return strings.ReplaceAll(slashPath, "/", "--")
}

// gistFilenameToRelPath is the inverse of gistFilename.
// Returns "" for unrecognised filenames.
func gistFilenameToRelPath(name string) string {
	switch name {
	case "config.yaml":
		return "config.yaml"
	case "blocklist.json":
		return filepath.Join("data", "blocklist.json")
	case "voted_stations.json":
		return filepath.Join("data", "voted_stations.json")
	case "ratings.json":
		return filepath.Join("data", "station_ratings.json")
	case "tags.json":
		return filepath.Join("data", "station_tags.json")
	case "metadata.json":
		return filepath.Join("data", "station_metadata.json")
	case "search-history.json":
		return filepath.Join("data", "favorites", SystemFileSearchHistory)
	}
	if strings.HasPrefix(name, "fav--") {
		base := strings.TrimPrefix(name, "fav--")
		// Reject empty, dot-segments, or any name that encodes a path separator.
		// A crafted filename like "fav--../../config.yaml" would otherwise escape
		// the config tree when joined against m.configDir.
		if base == "" || base == "." || base == ".." || base != filepath.Base(base) {
			return ""
		}
		return filepath.Join("data", "favorites", base)
	}
	return ""
}

// FindBackupGist returns the single tera-data-backup Gist, or nil if none exists.
// It returns an error if more than one qualifying Gist is found, since Push/Pull
// cannot safely choose between them.
//
// A Gist qualifies only if its description matches BackupGistDescription AND it
// contains the backupGistMarkerFile sentinel. This prevents a false-positive match
// on an unrelated user Gist that happens to share the same description.
func (m *GistSyncManager) FindBackupGist() (*gist.Gist, error) {
	gists, err := m.client.ListGists()
	if err != nil {
		return nil, fmt.Errorf("failed to list gists: %w", err)
	}
	var matches []*gist.Gist
	for _, g := range gists {
		if g.Description != BackupGistDescription {
			continue
		}
		// ListGists returns only summary data; fetch the full Gist to inspect files.
		full, err := m.client.GetGist(g.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch gist %s: %w", g.ID, err)
		}
		if _, hasMarker := full.Files[backupGistMarkerFile]; hasMarker {
			matches = append(matches, full)
		}
	}
	switch len(matches) {
	case 0:
		return nil, nil
	case 1:
		return matches[0], nil
	default:
		return nil, fmt.Errorf(
			"found %d backup gists with description %q; delete duplicates and keep one",
			len(matches), BackupGistDescription,
		)
	}
}

// AvailableCategories inspects the backup Gist and returns which categories are present.
// Returns a zero-value SyncPrefs (all false) if no backup Gist exists yet.
func (m *GistSyncManager) AvailableCategories() (SyncPrefs, error) {
	// FindBackupGist already fetches the full Gist (including file list).
	g, err := m.FindBackupGist()
	if err != nil {
		return SyncPrefs{}, err
	}
	if g == nil {
		return SyncPrefs{}, nil
	}

	var prefs SyncPrefs
	for name := range g.Files {
		relPath := gistFilenameToRelPath(name)
		switch {
		case relPath == "config.yaml":
			prefs.Settings = true
		case relPath == filepath.Join("data", "favorites", SystemFileSearchHistory):
			prefs.SearchHistory = true
		case strings.HasPrefix(filepath.ToSlash(relPath), "data/favorites/"):
			prefs.Favorites = true
		case relPath == filepath.Join("data", "station_ratings.json") ||
			relPath == filepath.Join("data", "voted_stations.json"):
			prefs.RatingsVotes = true
		case relPath == filepath.Join("data", "blocklist.json"):
			prefs.Blocklist = true
		case relPath == filepath.Join("data", "station_metadata.json") ||
			relPath == filepath.Join("data", "station_tags.json"):
			prefs.MetadataTags = true
		}
	}
	return prefs, nil
}

// syncPrefForRelPath reports whether the given config-relative path belongs to
// one of the selected sync categories. It reuses zipEntryWanted by converting
// the OS-native path to a slash-separated one.
func syncPrefForRelPath(relPath string, prefs SyncPrefs) bool {
	return zipEntryWanted(filepath.ToSlash(relPath), prefs)
}

// Push uploads selected files to the dedicated backup Gist.
// If no backup Gist exists it is created (secret); otherwise its files are updated.
// Files that exist locally are pushed; files that are in-scope in the Gist but
// no longer exist locally are deleted (sent as null per the GitHub API).
func (m *GistSyncManager) Push(prefs SyncPrefs) error {
	bm := &BackupManager{configDir: m.configDir}
	relPaths, err := bm.categoryFiles(prefs)
	if err != nil {
		return fmt.Errorf("failed to collect category files: %w", err)
	}

	// Collect files to upsert. Missing files are silently skipped;
	// empty files are pushed as empty strings (distinct from deletion).
	present := make(map[string]*string)
	for _, relPath := range relPaths {
		absPath := filepath.Join(m.configDir, relPath)
		data, err := os.ReadFile(absPath)
		if err != nil {
			if os.IsNotExist(err) {
				continue // skip missing files silently
			}
			return fmt.Errorf("failed to read %s: %w", relPath, err)
		}
		content := string(data)
		present[gistFilename(relPath)] = &content
	}

	// Always include the marker file so FindBackupGist can identify this Gist.
	marker := `{"app":"tera"}`
	present[backupGistMarkerFile] = &marker

	existing, err := m.FindBackupGist()
	if err != nil {
		return err
	}

	if existing == nil {
		if len(present) == 1 { // only the marker
			return fmt.Errorf("no files found to push for the selected categories")
		}
		_, err = m.client.CreateGist(BackupGistDescription, present, false)
		return err
	}

	// Add nil tombstones for in-scope files that no longer exist locally.
	// FindBackupGist already fetched the full Gist, so use it directly.
	for name := range existing.Files {
		relPath := gistFilenameToRelPath(name)
		if relPath == "" || !syncPrefForRelPath(relPath, prefs) {
			continue
		}
		if _, inPresent := present[name]; !inPresent {
			present[name] = nil // nil → delete from remote
		}
	}

	return m.client.UpdateGistFiles(existing.ID, present)
}

// ConflictingGistFiles returns the relative paths of files that would be
// overwritten by a Pull of the given Gist using the given prefs.
// It drives the check from the Gist's own file list so it works correctly on
// a fresh machine where no local favorites exist yet to enumerate.
func (m *GistSyncManager) ConflictingGistFiles(g *gist.Gist, prefs SyncPrefs) ([]string, error) {
	if g == nil {
		return nil, fmt.Errorf("gist is required")
	}
	// Ensure we have the full file list (ListGists omits file details).
	if len(g.Files) == 0 {
		var err error
		g, err = m.client.GetGist(g.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch backup gist: %w", err)
		}
	}

	var conflicts []string
	for name := range g.Files {
		relPath := gistFilenameToRelPath(name)
		if relPath == "" || !syncPrefForRelPath(relPath, prefs) {
			continue
		}
		absPath := filepath.Join(m.configDir, relPath)
		if _, err := os.Stat(absPath); err == nil {
			conflicts = append(conflicts, relPath)
		}
	}
	sort.Strings(conflicts)
	return conflicts, nil
}

// Pull downloads selected files from the dedicated backup Gist.
// When force is false and files already exist, it returns RestoreConflictError.
// Pass force=true to overwrite without checking.
func (m *GistSyncManager) Pull(prefs SyncPrefs, force bool) error {
	g, err := m.FindBackupGist()
	if err != nil {
		return err
	}
	if g == nil {
		return fmt.Errorf("no backup Gist found (description: %q); push first to create one", BackupGistDescription)
	}

	// Re-fetch full Gist to get file contents and raw URLs
	full, err := m.client.GetGist(g.ID)
	if err != nil {
		return fmt.Errorf("failed to fetch backup gist: %w", err)
	}

	if !force {
		conflicts, err := m.ConflictingGistFiles(full, prefs)
		if err != nil {
			return err
		}
		if len(conflicts) > 0 {
			return &RestoreConflictError{Paths: conflicts}
		}
	}

	// Drive wanted from the Gist's own file list so restores work on a fresh
	// machine where categoryFiles would return nothing for missing favorites.
	wanted := make(map[string]string) // gist filename → rel path
	for name := range full.Files {
		relPath := gistFilenameToRelPath(name)
		if relPath == "" || !syncPrefForRelPath(relPath, prefs) {
			continue
		}
		wanted[name] = relPath
	}

	httpClient := &http.Client{Timeout: backupGistHTTPTimeout}

	// Fetch all content first; only write to disk once everything is ready
	// so a mid-restore failure doesn't leave a partially updated config.
	staged := make(map[string][]byte, len(wanted))
	for name, gistFile := range full.Files {
		relPath, ok := wanted[name]
		if !ok {
			continue
		}
		content, err := fetchRawContent(httpClient, gistFile.RawURL, gistFile.Content)
		if err != nil {
			return fmt.Errorf("failed to fetch %s: %w", name, err)
		}
		staged[relPath] = []byte(content)
	}

	for relPath, content := range staged {
		destPath := filepath.Join(m.configDir, relPath)
		if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
			return fmt.Errorf("failed to create directory for %s: %w", relPath, err)
		}
		if err := os.WriteFile(destPath, content, 0600); err != nil {
			return fmt.Errorf("failed to write %s: %w", relPath, err)
		}
	}

	return nil
}

// fetchRawContent returns the file content. If rawURL is non-empty it fetches
// from there (avoids the 1 MB in-payload limit); otherwise falls back to the
// inline content string already returned by the API.
func fetchRawContent(client *http.Client, rawURL, inlineContent string) (string, error) {
	if rawURL == "" {
		return inlineContent, nil
	}

	resp, err := client.Get(rawURL)
	if err != nil {
		return "", fmt.Errorf("HTTP GET failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("HTTP error %d fetching raw content", resp.StatusCode)
	}

	// Limit to 10 MiB; read one extra byte so we can detect a truncated response.
	const maxBytes = 10 << 20
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxBytes+1))
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}
	if len(body) > maxBytes {
		return "", fmt.Errorf("raw content exceeds %d bytes", maxBytes)
	}
	return string(body), nil
}
