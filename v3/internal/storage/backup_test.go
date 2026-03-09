package storage

import (
	"archive/zip"
	"os"
	"path/filepath"
	"testing"
)

// setupBackupDir creates a temporary tera config directory populated with
// fixture files that match the expected on-disk layout.
func setupBackupDir(t *testing.T) (configDir string, bm *BackupManager) {
	t.Helper()
	configDir = t.TempDir()

	dirs := []string{
		filepath.Join(configDir, "data", "favorites"),
		filepath.Join(configDir, "data", "cache"),
	}
	for _, d := range dirs {
		if err := os.MkdirAll(d, 0755); err != nil {
			t.Fatalf("mkdir %s: %v", d, err)
		}
	}

	files := map[string]string{
		"config.yaml":                          "version: 3.0\n",
		"data/blocklist.json":                  `{"stations":[]}`,
		"data/voted_stations.json":             `{"stations":[]}`,
		"data/station_ratings.json":            `{}`,
		"data/station_tags.json":               `{}`,
		"data/station_metadata.json":           `{}`,
		"data/cache/search-history.json":       `{"search_items":[]}`,
		"data/favorites/Jazz.json":             `[]`,
		"data/favorites/Pops.json":             `[]`,
	}
	for rel, content := range files {
		path := filepath.Join(configDir, filepath.FromSlash(rel))
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("write fixture %s: %v", rel, err)
		}
	}

	bm = &BackupManager{configDir: configDir}
	return configDir, bm
}

// zipContains returns the set of filenames inside a zip archive.
func zipContains(t *testing.T, zipPath string) map[string]struct{} {
	t.Helper()
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		t.Fatalf("open zip: %v", err)
	}
	defer func() { _ = r.Close() }()
	names := make(map[string]struct{})
	for _, f := range r.File {
		names[f.Name] = struct{}{}
	}
	return names
}

func TestExport_AllCategories(t *testing.T) {
	_, bm := setupBackupDir(t)
	dest := filepath.Join(t.TempDir(), "backup.zip")

	if err := bm.Export(dest, DefaultSyncPrefs()); err != nil {
		t.Fatalf("Export: %v", err)
	}

	got := zipContains(t, dest)
	want := []string{
		"config.yaml",
		"data/blocklist.json",
		"data/voted_stations.json",
		"data/station_ratings.json",
		"data/station_tags.json",
		"data/station_metadata.json",
		"data/favorites/Jazz.json",
		"data/favorites/Pops.json",
	}
	for _, name := range want {
		if _, ok := got[name]; !ok {
			t.Errorf("expected %s in archive, not found", name)
		}
	}
	// Search history should be absent (default prefs have it off)
	if _, ok := got["data/cache/search-history.json"]; ok {
		t.Error("search-history.json should not be in archive when SearchHistory=false")
	}
}

func TestExport_SelectiveCategories(t *testing.T) {
	_, bm := setupBackupDir(t)
	dest := filepath.Join(t.TempDir(), "backup.zip")

	prefs := SyncPrefs{Favorites: true} // only favorites
	if err := bm.Export(dest, prefs); err != nil {
		t.Fatalf("Export: %v", err)
	}

	got := zipContains(t, dest)

	// Favorites should be present
	for _, name := range []string{"data/favorites/Jazz.json", "data/favorites/Pops.json"} {
		if _, ok := got[name]; !ok {
			t.Errorf("expected %s in archive", name)
		}
	}
	// Everything else should be absent
	for _, name := range []string{"config.yaml", "data/blocklist.json", "data/station_ratings.json"} {
		if _, ok := got[name]; ok {
			t.Errorf("unexpected %s in archive", name)
		}
	}
}

func TestExport_SkipsMissingFiles(t *testing.T) {
	_, bm := setupBackupDir(t)

	// Remove blocklist so it's missing on disk
	_ = os.Remove(filepath.Join(bm.configDir, "data", "blocklist.json"))

	dest := filepath.Join(t.TempDir(), "backup.zip")
	prefs := SyncPrefs{Blocklist: true, Settings: true}
	if err := bm.Export(dest, prefs); err != nil {
		t.Fatalf("Export should not fail on missing file: %v", err)
	}

	got := zipContains(t, dest)
	if _, ok := got["data/blocklist.json"]; ok {
		t.Error("missing blocklist.json should have been silently skipped")
	}
	if _, ok := got["config.yaml"]; !ok {
		t.Error("config.yaml should be in archive")
	}
}

func TestListArchiveCategories(t *testing.T) {
	_, bm := setupBackupDir(t)
	dest := filepath.Join(t.TempDir(), "backup.zip")

	all := DefaultSyncPrefs()
	all.SearchHistory = true
	if err := bm.Export(dest, all); err != nil {
		t.Fatalf("Export: %v", err)
	}

	prefs, err := bm.ListArchiveCategories(dest)
	if err != nil {
		t.Fatalf("ListArchiveCategories: %v", err)
	}

	if !prefs.Favorites {
		t.Error("expected Favorites=true")
	}
	if !prefs.Settings {
		t.Error("expected Settings=true")
	}
	if !prefs.RatingsVotes {
		t.Error("expected RatingsVotes=true")
	}
	if !prefs.Blocklist {
		t.Error("expected Blocklist=true")
	}
	if !prefs.MetadataTags {
		t.Error("expected MetadataTags=true")
	}
	if !prefs.SearchHistory {
		t.Error("expected SearchHistory=true")
	}
}

func TestListArchiveCategories_Selective(t *testing.T) {
	_, bm := setupBackupDir(t)
	dest := filepath.Join(t.TempDir(), "backup.zip")

	prefs := SyncPrefs{Favorites: true, Blocklist: true}
	if err := bm.Export(dest, prefs); err != nil {
		t.Fatalf("Export: %v", err)
	}

	got, err := bm.ListArchiveCategories(dest)
	if err != nil {
		t.Fatalf("ListArchiveCategories: %v", err)
	}

	if !got.Favorites {
		t.Error("expected Favorites=true")
	}
	if !got.Blocklist {
		t.Error("expected Blocklist=true")
	}
	if got.Settings || got.RatingsVotes || got.MetadataTags || got.SearchHistory {
		t.Errorf("expected only Favorites and Blocklist, got %+v", got)
	}
}

func TestRestore_AllCategories(t *testing.T) {
	_, bm := setupBackupDir(t)
	dest := filepath.Join(t.TempDir(), "backup.zip")

	if err := bm.Export(dest, DefaultSyncPrefs()); err != nil {
		t.Fatalf("Export: %v", err)
	}

	// Restore into a fresh empty configDir
	restoreDir := t.TempDir()
	restoreBM := &BackupManager{configDir: restoreDir}

	if err := restoreBM.Restore(dest, DefaultSyncPrefs(), false); err != nil {
		t.Fatalf("Restore: %v", err)
	}

	for _, rel := range []string{
		"config.yaml",
		"data/blocklist.json",
		"data/voted_stations.json",
		"data/station_ratings.json",
		"data/station_tags.json",
		"data/station_metadata.json",
		"data/favorites/Jazz.json",
		"data/favorites/Pops.json",
	} {
		path := filepath.Join(restoreDir, filepath.FromSlash(rel))
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("expected restored file %s to exist", rel)
		}
	}
}

func TestRestore_SelectiveCategories(t *testing.T) {
	_, bm := setupBackupDir(t)
	dest := filepath.Join(t.TempDir(), "backup.zip")

	// Archive contains everything
	all := DefaultSyncPrefs()
	all.SearchHistory = true
	if err := bm.Export(dest, all); err != nil {
		t.Fatalf("Export: %v", err)
	}

	// Restore only favorites into a fresh dir
	restoreDir := t.TempDir()
	restoreBM := &BackupManager{configDir: restoreDir}
	prefs := SyncPrefs{Favorites: true}
	if err := restoreBM.Restore(dest, prefs, false); err != nil {
		t.Fatalf("Restore: %v", err)
	}

	// Favorites should be present
	for _, rel := range []string{"data/favorites/Jazz.json", "data/favorites/Pops.json"} {
		path := filepath.Join(restoreDir, filepath.FromSlash(rel))
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("expected %s to exist after selective restore", rel)
		}
	}
	// config.yaml should NOT be present
	if _, err := os.Stat(filepath.Join(restoreDir, "config.yaml")); err == nil {
		t.Error("config.yaml should not have been restored when Settings=false")
	}
}

func TestRestore_OverwriteGuard(t *testing.T) {
	_, bm := setupBackupDir(t)
	dest := filepath.Join(t.TempDir(), "backup.zip")

	if err := bm.Export(dest, DefaultSyncPrefs()); err != nil {
		t.Fatalf("Export: %v", err)
	}

	// Restore into a dir that already has config.yaml
	restoreDir := t.TempDir()
	existingFile := filepath.Join(restoreDir, "config.yaml")
	if err := os.WriteFile(existingFile, []byte("existing"), 0644); err != nil {
		t.Fatalf("write existing: %v", err)
	}

	restoreBM := &BackupManager{configDir: restoreDir}
	err := restoreBM.Restore(dest, DefaultSyncPrefs(), false)
	if err == nil {
		t.Fatal("expected RestoreConflictError, got nil")
	}

	var conflictErr *RestoreConflictError
	if !isRestoreConflictError(err, &conflictErr) {
		t.Fatalf("expected *RestoreConflictError, got %T: %v", err, err)
	}
	if len(conflictErr.Paths) == 0 {
		t.Error("expected at least one conflicting path")
	}
}

func TestRestore_ForceOverwrite(t *testing.T) {
	_, bm := setupBackupDir(t)
	dest := filepath.Join(t.TempDir(), "backup.zip")

	if err := bm.Export(dest, SyncPrefs{Settings: true}); err != nil {
		t.Fatalf("Export: %v", err)
	}

	restoreDir := t.TempDir()
	// Pre-create the file with different content
	if err := os.WriteFile(filepath.Join(restoreDir, "config.yaml"), []byte("old"), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}

	restoreBM := &BackupManager{configDir: restoreDir}
	if err := restoreBM.Restore(dest, SyncPrefs{Settings: true}, true); err != nil {
		t.Fatalf("Restore with force=true: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(restoreDir, "config.yaml"))
	if err != nil {
		t.Fatalf("read restored file: %v", err)
	}
	if string(data) != "version: 3.0\n" {
		t.Errorf("expected restored content, got %q", string(data))
	}
}

func TestResolveBackupPath_DirectoryAppendsFilename(t *testing.T) {
	dir := t.TempDir()
	resolved, err := ResolveBackupPath(dir + string(os.PathSeparator))
	if err != nil {
		t.Fatalf("ResolveBackupPath: %v", err)
	}
	if filepath.Dir(resolved) != dir {
		t.Errorf("expected resolved path inside %s, got %s", dir, resolved)
	}
	if filepath.Ext(resolved) != ".zip" {
		t.Errorf("expected .zip extension, got %s", resolved)
	}
}

func TestResolveBackupPath_FilePathPassthrough(t *testing.T) {
	path := "/tmp/mybackup.zip"
	resolved, err := ResolveBackupPath(path)
	if err != nil {
		t.Fatalf("ResolveBackupPath: %v", err)
	}
	if resolved != path {
		t.Errorf("expected %s, got %s", path, resolved)
	}
}

// isRestoreConflictError is a helper to type-assert without importing errors package.
func isRestoreConflictError(err error, target **RestoreConflictError) bool {
	if ce, ok := err.(*RestoreConflictError); ok {
		*target = ce
		return true
	}
	return false
}
