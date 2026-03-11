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

	"github.com/shinokada/tera/v3/internal/config"
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
	dir, err := teraConfigDir()
	if err != nil {
		return nil, err
	}
	return &GistSyncManager{
		client:    client,
		configDir: dir,
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
		// Reject fav--search-history.json: search-history.json already has its
		// own canonical mapping and this alias would resolve to the same
		// destination, letting a crafted gist silently overwrite it with
		// arbitrary content categorised as a plain favorites file.
		if base == SystemFileSearchHistory {
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
		categorizePath(name, &prefs)
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
	// FindBackupGist already fetches the full Gist (including file contents and
	// raw URLs) to verify the marker file, so reuse it directly.
	g, err := m.FindBackupGist()
	if err != nil {
		return err
	}
	if g == nil {
		return fmt.Errorf("no backup Gist found (description: %q); push first to create one", BackupGistDescription)
	}
	return m.PullFromGist(g, prefs, force)
}

// PullFromGist downloads selected files from the given Gist directly,
// without requiring it to be the dedicated backup Gist.
// When force is false and files already exist, it returns RestoreConflictError.
// Pass force=true to overwrite without checking.
func (m *GistSyncManager) PullFromGist(g *gist.Gist, prefs SyncPrefs, force bool) error {
	if g == nil {
		return fmt.Errorf("gist is required")
	}
	if len(g.Files) == 0 {
		full, err := m.client.GetGist(g.ID)
		if err != nil {
			return fmt.Errorf("failed to fetch gist: %w", err)
		}
		g = full
	}
	if _, ok := g.Files[backupGistMarkerFile]; !ok {
		return fmt.Errorf("gist is not a tera backup (missing %s)", backupGistMarkerFile)
	}

	if !force {
		conflicts, err := m.ConflictingGistFiles(g, prefs)
		if err != nil {
			return err
		}
		if len(conflicts) > 0 {
			return &RestoreConflictError{Paths: conflicts}
		}
	}

	// Drive wanted from the Gist's own file list so restores work on a fresh
	// machine where categoryFiles would return nothing for missing favorites.
	httpClient := &http.Client{Timeout: backupGistHTTPTimeout}
	return stageAndWriteGistFiles(httpClient, g.Files, prefs, m.configDir)
}

// categorizePath updates prefs based on a single gist filename.
// It is the single source of truth for mapping gist filenames to SyncPrefs
// categories, shared by AvailableCategories, AvailableCategoriesFromGist, and
// AvailableCategoriesFromGistFiles.
func categorizePath(name string, prefs *SyncPrefs) {
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

// AvailableCategoriesFromGist inspects the given Gist and returns which
// categories are present, without requiring authentication or ownership.
func (m *GistSyncManager) AvailableCategoriesFromGist(g *gist.Gist) (SyncPrefs, error) {
	if g == nil {
		return SyncPrefs{}, fmt.Errorf("gist is required")
	}
	if len(g.Files) == 0 {
		full, err := m.client.GetGist(g.ID)
		if err != nil {
			return SyncPrefs{}, fmt.Errorf("failed to fetch gist: %w", err)
		}
		g = full
	}
	if _, ok := g.Files[backupGistMarkerFile]; !ok {
		return SyncPrefs{}, fmt.Errorf("gist is not a tera backup (missing %s)", backupGistMarkerFile)
	}
	var prefs SyncPrefs
	for name := range g.Files {
		categorizePath(name, &prefs)
	}
	return prefs, nil
}

// AvailableCategoriesFromGistFiles inspects a raw Gist file map and returns
// which tera data categories are present. This is a package-level helper used
// when no GistSyncManager is available (i.e. no token configured).
func AvailableCategoriesFromGistFiles(files map[string]gist.GistFile) SyncPrefs {
	if _, ok := files[backupGistMarkerFile]; !ok {
		return SyncPrefs{}
	}
	var prefs SyncPrefs
	for name := range files {
		categorizePath(name, &prefs)
	}
	return prefs
}

// teraConfigDir returns the tera configuration directory path.
// Delegates to config.GetConfigDir so that all parts of the app share a
// single source of truth for config-dir resolution.
func teraConfigDir() (string, error) {
	return config.GetConfigDir()
}

// ConflictingFilesForGist checks for existing local files that would be
// overwritten by a restore from the given Gist. Works without a GistSyncManager.
// The caller must ensure g.Files is populated (e.g. by fetching the full gist
// first); this function does not re-fetch because it has no authenticated client
// and a public re-fetch would fail for private gists.
func ConflictingFilesForGist(g *gist.Gist, prefs SyncPrefs) ([]string, error) {
	if g == nil {
		return nil, fmt.Errorf("gist is required")
	}
	if len(g.Files) == 0 {
		return nil, fmt.Errorf("gist files not populated; fetch the full gist before calling ConflictingFilesForGist")
	}
	baseDir, err := teraConfigDir()
	if err != nil {
		return nil, err
	}

	var conflicts []string
	for name := range g.Files {
		relPath := gistFilenameToRelPath(name)
		if relPath == "" || !syncPrefForRelPath(relPath, prefs) {
			continue
		}
		absPath := filepath.Join(baseDir, relPath)
		if _, err := os.Stat(absPath); err == nil {
			conflicts = append(conflicts, relPath)
		}
	}
	sort.Strings(conflicts)
	return conflicts, nil
}

// RestoreFromGistDirect downloads selected files from a Gist without requiring
// a GistSyncManager or token. Used when restoring from a public gist URL with
// no token configured. When force is false, returns RestoreConflictError if
// any local files would be overwritten.
// The caller must ensure g.Files is populated (e.g. by fetching the full gist
// first); this function does not re-fetch because it has no authenticated client
// and a public re-fetch would fail for private gists.
func RestoreFromGistDirect(g *gist.Gist, prefs SyncPrefs, force bool) error {
	if g == nil {
		return fmt.Errorf("gist is required")
	}
	if len(g.Files) == 0 {
		return fmt.Errorf("gist files not populated; fetch the full gist before calling RestoreFromGistDirect")
	}
	if _, ok := g.Files[backupGistMarkerFile]; !ok {
		return fmt.Errorf("gist is not a tera backup (missing %s)", backupGistMarkerFile)
	}
	if !force {
		conflicts, err := ConflictingFilesForGist(g, prefs)
		if err != nil {
			return err
		}
		if len(conflicts) > 0 {
			return &RestoreConflictError{Paths: conflicts}
		}
	}
	baseDir, err := teraConfigDir()
	if err != nil {
		return err
	}

	httpClient := &http.Client{Timeout: backupGistHTTPTimeout}
	return stageAndWriteGistFiles(httpClient, g.Files, prefs, baseDir)
}

// stageAndWriteGistFiles fetches Gist file content into memory, then writes
// each selected file atomically to baseDir. Fetching everything before any
// disk write ensures that a network error never leaves a partially-written
// config.
//
// Atomicity guarantee: each individual file is written atomically
// (temp-file + rename), so no single file is ever torn. The set of files as
// a whole is NOT written atomically — a failure mid-loop leaves
// already-written files on disk. This is intentional: the individually-atomic
// files are never corrupt, and the user can complete a partial restore by
// re-running with force=true (the overwrite-warning screen offers this).
func stageAndWriteGistFiles(
	httpClient *http.Client,
	files map[string]gist.GistFile,
	prefs SyncPrefs,
	baseDir string,
) error {
	staged := make(map[string][]byte)
	for name, gistFile := range files {
		relPath := gistFilenameToRelPath(name)
		if relPath == "" || !syncPrefForRelPath(relPath, prefs) {
			continue
		}
		content, err := fetchRawContent(httpClient, gistFile.RawURL, gistFile.Content)
		if err != nil {
			return fmt.Errorf("failed to fetch %s: %w", name, err)
		}
		staged[relPath] = []byte(content)
	}
	for relPath, content := range staged {
		destPath := filepath.Join(baseDir, relPath)
		if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
			return fmt.Errorf("failed to create directory for %s: %w", relPath, err)
		}
		if err := atomicWriteFile(destPath, content, 0600); err != nil {
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
