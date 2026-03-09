package storage

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/shinokada/tera/v3/internal/gist"
)

const (
	backupGistDescription = "tera-data-backup"
	backupGistHTTPTimeout = 30 * time.Second
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
		return filepath.Join("data", "favorites", base)
	}
	return ""
}

// FindBackupGist returns the existing tera-data-backup Gist, or nil if none exists.
func (m *GistSyncManager) FindBackupGist() (*gist.Gist, error) {
	gists, err := m.client.ListGists()
	if err != nil {
		return nil, fmt.Errorf("failed to list gists: %w", err)
	}
	for _, g := range gists {
		if g.Description == backupGistDescription {
			return g, nil
		}
	}
	return nil, nil
}

// AvailableCategories inspects the backup Gist and returns which categories are present.
// Returns a zero-value SyncPrefs (all false) if no backup Gist exists yet.
func (m *GistSyncManager) AvailableCategories() (SyncPrefs, error) {
	g, err := m.FindBackupGist()
	if err != nil {
		return SyncPrefs{}, err
	}
	if g == nil {
		return SyncPrefs{}, nil
	}

	// Re-fetch the full Gist to get file list (ListGists omits file contents)
	full, err := m.client.GetGist(g.ID)
	if err != nil {
		return SyncPrefs{}, fmt.Errorf("failed to fetch backup gist: %w", err)
	}

	var prefs SyncPrefs
	for name := range full.Files {
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

	// present holds files to upsert; skipped tracks names explicitly excluded
	// due to being empty or missing, so we never tombstone them accidentally.
	present := make(map[string]string)
	skipped := make(map[string]struct{})
	for _, relPath := range relPaths {
		absPath := filepath.Join(m.configDir, relPath)
		data, err := os.ReadFile(absPath)
		if err != nil {
			if os.IsNotExist(err) {
				skipped[gistFilename(relPath)] = struct{}{}
				continue // skip missing files silently
			}
			return fmt.Errorf("failed to read %s: %w", relPath, err)
		}
		// Guard against empty reads: UpdateGistFiles interprets "" as a deletion
		// tombstone. All valid tera data files contain at least "{}" or "[]".
		if len(data) == 0 {
			skipped[gistFilename(relPath)] = struct{}{}
			continue
		}
		present[gistFilename(relPath)] = string(data)
	}

	existing, err := m.FindBackupGist()
	if err != nil {
		return err
	}

	if existing == nil {
		if len(present) == 0 {
			return fmt.Errorf("no files found to push for the selected categories")
		}
		_, err = m.client.CreateGist(backupGistDescription, present, false)
		return err
	}

	// Add tombstones (empty string → null → delete) only for in-scope files
	// that no longer exist locally. Files skipped due to being empty or missing
	// are left untouched in the remote Gist.
	full, err := m.client.GetGist(existing.ID)
	if err != nil {
		return fmt.Errorf("failed to fetch backup gist: %w", err)
	}
	for name := range full.Files {
		relPath := gistFilenameToRelPath(name)
		if relPath == "" || !syncPrefForRelPath(relPath, prefs) {
			continue
		}
		if _, inPresent := present[name]; inPresent {
			continue
		}
		if _, wasSkipped := skipped[name]; wasSkipped {
			continue // empty/missing locally — preserve remote copy
		}
		present[name] = "" // tombstone: file was intentionally removed
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
		return fmt.Errorf("no backup Gist found (description: %q); push first to create one", backupGistDescription)
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
