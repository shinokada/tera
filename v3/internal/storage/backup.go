package storage

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// BackupManager handles zip-based export and restore of user data.
type BackupManager struct {
	configDir string
}

// NewBackupManager creates a BackupManager rooted at os.UserConfigDir()/tera.
func NewBackupManager() (*BackupManager, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get config directory: %w", err)
	}
	return &BackupManager{configDir: filepath.Join(configDir, "tera")}, nil
}

// DefaultBackupPath returns a sensible default export path using today's date.
// Example: ~/tera-backup-2026-03-09.zip
func DefaultBackupPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	filename := fmt.Sprintf("tera-backup-%s.zip", time.Now().Format("2006-01-02"))
	return filepath.Join(home, filename), nil
}

// ResolveBackupPath resolves the user-supplied path to a .zip file path.
// If path is a directory (ends with separator or exists as a dir), the default
// filename is appended automatically.
func ResolveBackupPath(path string) (string, error) {
	// Expand ~ to home directory
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("failed to expand ~: %w", err)
		}
		path = filepath.Join(home, path[2:])
	}

	// If it already looks like a directory (trailing slash or existing dir), append filename
	info, statErr := os.Stat(path)
	isDir := (statErr == nil && info.IsDir()) || strings.HasSuffix(path, string(os.PathSeparator))
	if isDir {
		filename := fmt.Sprintf("tera-backup-%s.zip", time.Now().Format("2006-01-02"))
		path = filepath.Join(path, filename)
	}

	return path, nil
}

// categoryFiles returns the list of relative paths (relative to configDir) for a given category.
func (b *BackupManager) categoryFiles(prefs SyncPrefs) []string {
	var files []string

	if prefs.Settings {
		files = append(files, "config.yaml")
	}
	if prefs.Favorites {
		// Add all *.json files from data/favorites/
		favDir := filepath.Join(b.configDir, "data", "favorites")
		entries, err := os.ReadDir(favDir)
		if err == nil {
			for _, e := range entries {
				if !e.IsDir() && strings.HasSuffix(e.Name(), ".json") {
					files = append(files, filepath.Join("data", "favorites", e.Name()))
				}
			}
		}
	}
	if prefs.RatingsVotes {
		files = append(files,
			filepath.Join("data", "station_ratings.json"),
			filepath.Join("data", "voted_stations.json"),
		)
	}
	if prefs.Blocklist {
		files = append(files, filepath.Join("data", "blocklist.json"))
	}
	if prefs.MetadataTags {
		files = append(files,
			filepath.Join("data", "station_metadata.json"),
			filepath.Join("data", "station_tags.json"),
		)
	}
	if prefs.SearchHistory {
		files = append(files, filepath.Join("data", "cache", "search-history.json"))
	}

	return files
}

// Export creates a zip archive at destPath containing the files selected by prefs.
// Only files that actually exist on disk are included; missing files are silently skipped.
func (b *BackupManager) Export(destPath string, prefs SyncPrefs) error {
	resolved, err := ResolveBackupPath(destPath)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(resolved), 0755); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	zipFile, err := os.Create(resolved)
	if err != nil {
		return fmt.Errorf("failed to create zip file: %w", err)
	}
	defer func() { _ = zipFile.Close() }()

	w := zip.NewWriter(zipFile)
	defer func() { _ = w.Close() }()

	for _, relPath := range b.categoryFiles(prefs) {
		absPath := filepath.Join(b.configDir, relPath)
		if _, err := os.Stat(absPath); os.IsNotExist(err) {
			continue // skip missing files silently
		}

		if err := addFileToZip(w, absPath, relPath); err != nil {
			return fmt.Errorf("failed to add %s to archive: %w", relPath, err)
		}
	}

	return nil
}

// ConflictingFiles returns the list of relative paths that would be overwritten
// by a restore of the given zip using the given prefs. Used to populate the
// overwrite warning UI before committing a restore.
func (b *BackupManager) ConflictingFiles(srcPath string, prefs SyncPrefs) ([]string, error) {
	r, err := zip.OpenReader(srcPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open zip archive: %w", err)
	}
	defer func() { _ = r.Close() }()

	var conflicts []string
	for _, f := range r.File {
		slashName := filepath.ToSlash(f.Name)
		if !zipEntryWanted(slashName, prefs) {
			continue
		}
		absPath := filepath.Join(b.configDir, filepath.FromSlash(slashName))
		if _, err := os.Stat(absPath); err == nil {
			conflicts = append(conflicts, slashName)
		}
	}
	return conflicts, nil
}

// Restore extracts a zip archive at srcPath, restoring files selected by prefs.
// When force is false and files already exist, it returns RestoreConflictError
// listing the conflicting paths. Pass force=true to overwrite without checking.
func (b *BackupManager) Restore(srcPath string, prefs SyncPrefs, force bool) error {
	if !force {
		conflicts, err := b.ConflictingFiles(srcPath, prefs)
		if err != nil {
			return err
		}
		if len(conflicts) > 0 {
			return &RestoreConflictError{Paths: conflicts}
		}
	}

	r, err := zip.OpenReader(srcPath)
	if err != nil {
		return fmt.Errorf("failed to open zip archive: %w", err)
	}
	defer func() { _ = r.Close() }()

	// Walk every entry in the zip and restore those whose category is selected.
	// We derive the category from the entry name directly rather than calling
	// categoryFiles on the (possibly empty) restore target, which would miss
	// favorites because the target dir contains no files yet to enumerate.
	for _, f := range r.File {
		slashName := filepath.ToSlash(f.Name)
		if !zipEntryWanted(slashName, prefs) {
			continue
		}
		// Build dest using FromSlash so subdirectories are created correctly on
		// all platforms.
		destPath := filepath.Join(b.configDir, filepath.FromSlash(slashName))
		if err := extractFileFromZip(f, destPath); err != nil {
			return fmt.Errorf("failed to restore %s: %w", slashName, err)
		}
	}

	return nil
}

// zipEntryWanted reports whether a zip entry (named with forward slashes)
// belongs to one of the selected categories.
func zipEntryWanted(slashName string, prefs SyncPrefs) bool {
	switch {
	case slashName == "config.yaml":
		return prefs.Settings
	case strings.HasPrefix(slashName, "data/favorites/"):
		return prefs.Favorites
	case slashName == "data/station_ratings.json" || slashName == "data/voted_stations.json":
		return prefs.RatingsVotes
	case slashName == "data/blocklist.json":
		return prefs.Blocklist
	case slashName == "data/station_metadata.json" || slashName == "data/station_tags.json":
		return prefs.MetadataTags
	case slashName == "data/cache/search-history.json":
		return prefs.SearchHistory
	}
	return false
}

// ListArchiveCategories inspects a zip archive and returns which categories are present.
func (b *BackupManager) ListArchiveCategories(srcPath string) (SyncPrefs, error) {
	r, err := zip.OpenReader(srcPath)
	if err != nil {
		return SyncPrefs{}, fmt.Errorf("failed to open zip archive: %w", err)
	}
	defer func() { _ = r.Close() }()

	var prefs SyncPrefs
	for _, f := range r.File {
		name := filepath.ToSlash(f.Name)
		switch {
		case name == "config.yaml":
			prefs.Settings = true
		case strings.HasPrefix(name, "data/favorites/"):
			prefs.Favorites = true
		case name == "data/station_ratings.json" || name == "data/voted_stations.json":
			prefs.RatingsVotes = true
		case name == "data/blocklist.json":
			prefs.Blocklist = true
		case name == "data/station_metadata.json" || name == "data/station_tags.json":
			prefs.MetadataTags = true
		case name == "data/cache/search-history.json":
			prefs.SearchHistory = true
		}
	}
	return prefs, nil
}

// intersectPrefs returns a SyncPrefs where a category is true only if it is
// true in both a and b. Used to restrict a restore to what is both available
// in the archive and requested by the user.
func intersectPrefs(a, b SyncPrefs) SyncPrefs {
	return SyncPrefs{
		Favorites:     a.Favorites && b.Favorites,
		Settings:      a.Settings && b.Settings,
		RatingsVotes:  a.RatingsVotes && b.RatingsVotes,
		Blocklist:     a.Blocklist && b.Blocklist,
		MetadataTags:  a.MetadataTags && b.MetadataTags,
		SearchHistory: a.SearchHistory && b.SearchHistory,
	}
}

// addFileToZip adds the file at absPath to the zip writer using relPath as the
// entry name inside the archive.
func addFileToZip(w *zip.Writer, absPath, relPath string) error {
	f, err := os.Open(absPath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer func() { _ = f.Close() }()

	// Use forward slashes in the zip entry name regardless of OS
	entry, err := w.Create(filepath.ToSlash(relPath))
	if err != nil {
		return fmt.Errorf("failed to create zip entry: %w", err)
	}

	if _, err := io.Copy(entry, f); err != nil {
		return fmt.Errorf("failed to write zip entry: %w", err)
	}
	return nil
}

// extractFileFromZip extracts a single zip.File entry to destPath,
// creating parent directories as needed.
func extractFileFromZip(f *zip.File, destPath string) error {
	if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	rc, err := f.Open()
	if err != nil {
		return fmt.Errorf("failed to open zip entry: %w", err)
	}
	defer func() { _ = rc.Close() }()

	out, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer func() { _ = out.Close() }()

	if _, err := io.Copy(out, rc); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}
	return nil
}

// RestoreConflictError is returned by Restore when force=false and one or more
// destination files already exist. The UI uses Paths to populate the overwrite
// warning before asking the user to confirm.
type RestoreConflictError struct {
	Paths []string
}

// Error implements the error interface.
func (e *RestoreConflictError) Error() string {
	return fmt.Sprintf("restore would overwrite %d existing file(s): %s",
		len(e.Paths), strings.Join(e.Paths, ", "))
}
