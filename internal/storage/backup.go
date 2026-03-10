package storage

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path"
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
	// Expand ~ to home directory (handle both Unix ~/... and Windows ~\...)
	if path == "~" || strings.HasPrefix(path, "~/") || strings.HasPrefix(path, "~\\") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("failed to expand ~: %w", err)
		}
		if path == "~" {
			path = home
		} else {
			path = filepath.Join(home, path[2:])
		}
	}

	// If it already looks like a directory (trailing slash or existing dir), append filename
	info, statErr := os.Stat(path)
	isDir := (statErr == nil && info.IsDir()) ||
		strings.HasSuffix(path, "/") ||
		strings.HasSuffix(path, "\\")
	if isDir {
		filename := fmt.Sprintf("tera-backup-%s.zip", time.Now().Format("2006-01-02"))
		path = filepath.Join(path, filename)
	}

	return path, nil
}

// categoryFiles returns the list of relative paths (relative to configDir) for a given category.
// Returns an error if the favorites directory exists but cannot be read (e.g. permission denied).
func (b *BackupManager) categoryFiles(prefs SyncPrefs) ([]string, error) {
	var files []string

	if prefs.Settings {
		files = append(files, "config.yaml")
	}
	if prefs.Favorites || prefs.SearchHistory {
		// Add *.json files from data/favorites/, routing search-history.json
		// to SearchHistory and all others to Favorites.
		favDir := filepath.Join(b.configDir, "data", "favorites")
		entries, err := os.ReadDir(favDir)
		if err != nil {
			if !os.IsNotExist(err) {
				return nil, fmt.Errorf("failed to read favorites directory: %w", err)
			}
		} else {
			for _, e := range entries {
				if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
					continue
				}
				relPath := filepath.Join("data", "favorites", e.Name())
				if e.Name() == SystemFileSearchHistory {
					if prefs.SearchHistory {
						files = append(files, relPath)
					}
					continue
				}
				if prefs.Favorites {
					files = append(files, relPath)
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

	return files, nil
}

// Export creates a zip archive at destPath containing the files selected by prefs.
// Only files that actually exist on disk are included; missing files are silently skipped.
// Uses a temp-file-and-rename strategy so a failed export never corrupts an
// existing backup at the destination path.
func (b *BackupManager) Export(destPath string, prefs SyncPrefs) (err error) {
	resolved, err := ResolveBackupPath(destPath)
	if err != nil {
		return err
	}

	destDir := filepath.Dir(resolved)
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	// Write to a temp file in the same directory so the rename is atomic.
	tmpFile, err := os.CreateTemp(destDir, ".tera-backup-*.zip.tmp")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpName := tmpFile.Name()

	// On any error path, remove the temp file.
	defer func() {
		if err != nil {
			_ = os.Remove(tmpName)
		}
	}()

	w := zip.NewWriter(tmpFile)
	writeErr := func() error {
		categoryFiles, err := b.categoryFiles(prefs)
		if err != nil {
			return err
		}
		for _, relPath := range categoryFiles {
			absPath := filepath.Join(b.configDir, relPath)
			if _, statErr := os.Stat(absPath); os.IsNotExist(statErr) {
				continue // skip missing files silently
			}
			if addErr := addFileToZip(w, absPath, relPath); addErr != nil {
				return fmt.Errorf("failed to add %s to archive: %w", relPath, addErr)
			}
		}
		return nil
	}()
	if writeErr != nil {
		_ = w.Close()
		_ = tmpFile.Close()
		return writeErr
	}
	if err = w.Close(); err != nil {
		_ = tmpFile.Close()
		return fmt.Errorf("failed to finalize zip archive: %w", err)
	}
	if err = tmpFile.Close(); err != nil {
		return fmt.Errorf("failed to close temp file: %w", err)
	}

	// Atomically replace destination.
	if err = os.Rename(tmpName, resolved); err != nil {
		return fmt.Errorf("failed to finalize backup file: %w", err)
	}
	return nil
}

// cleanArchiveEntryName normalises a raw ZIP entry name to a clean forward-slash
// path and rejects anything that is absolute or starts with a traversal segment.
// This must be called before zipEntryWanted so that category matching always
// sees the canonical form of the name, preventing crafted names like
// "data/favorites/../../config.yaml" from matching the wrong category.
func cleanArchiveEntryName(name string) (string, error) {
	cleaned := path.Clean(filepath.ToSlash(name))
	if cleaned == "." || cleaned == ".." ||
		strings.HasPrefix(cleaned, "../") ||
		strings.HasPrefix(cleaned, "/") {
		return "", fmt.Errorf("invalid archive entry %q", name)
	}
	return cleaned, nil
}

// archiveEntryPath safely resolves a zip entry name relative to baseDir.
// It rejects absolute paths and any traversal that would escape baseDir (zip-slip).
// Callers should pre-clean the name with cleanArchiveEntryName first.
func archiveEntryPath(baseDir, slashName string) (string, error) {
	cleaned := filepath.Clean(filepath.FromSlash(slashName))
	if filepath.IsAbs(cleaned) {
		return "", fmt.Errorf("invalid archive entry %q", slashName)
	}
	destPath := filepath.Join(baseDir, cleaned)
	rel, err := filepath.Rel(baseDir, destPath)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
		return "", fmt.Errorf("archive entry escapes config dir: %q", slashName)
	}
	return destPath, nil
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
		slashName, err := cleanArchiveEntryName(f.Name)
		if err != nil {
			return nil, err
		}
		if f.FileInfo().IsDir() {
			continue
		}
		if !zipEntryWanted(slashName, prefs) {
			continue
		}
		absPath, err := archiveEntryPath(b.configDir, slashName)
		if err != nil {
			return nil, err
		}
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
		slashName, err := cleanArchiveEntryName(f.Name)
		if err != nil {
			return err
		}
		if f.FileInfo().IsDir() {
			continue
		}
		if !zipEntryWanted(slashName, prefs) {
			continue
		}
		// Build dest using archiveEntryPath to guard against zip-slip.
		destPath, err := archiveEntryPath(b.configDir, slashName)
		if err != nil {
			return err
		}
		if err := extractFileFromZip(f, destPath); err != nil {
			return fmt.Errorf("failed to restore %s: %w", slashName, err)
		}
	}

	return nil
}

// zipEntryWanted reports whether a zip entry (named with forward slashes)
// belongs to one of the selected categories.
// slashName must already be cleaned via cleanArchiveEntryName.
func zipEntryWanted(slashName string, prefs SyncPrefs) bool {
	switch {
	case slashName == "config.yaml":
		return prefs.Settings
	case slashName == "data/favorites/"+SystemFileSearchHistory:
		return prefs.SearchHistory
	case strings.HasPrefix(slashName, "data/favorites/"):
		return prefs.Favorites
	case slashName == "data/station_ratings.json" || slashName == "data/voted_stations.json":
		return prefs.RatingsVotes
	case slashName == "data/blocklist.json":
		return prefs.Blocklist
	case slashName == "data/station_metadata.json" || slashName == "data/station_tags.json":
		return prefs.MetadataTags
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
		name, err := cleanArchiveEntryName(f.Name)
		if err != nil {
			return SyncPrefs{}, err
		}
		switch {
		case name == "config.yaml":
			prefs.Settings = true
		case name == "data/favorites/"+SystemFileSearchHistory:
			prefs.SearchHistory = true
		case strings.HasPrefix(name, "data/favorites/"):
			prefs.Favorites = true
		case name == "data/station_ratings.json" || name == "data/voted_stations.json":
			prefs.RatingsVotes = true
		case name == "data/blocklist.json":
			prefs.Blocklist = true
		case name == "data/station_metadata.json" || name == "data/station_tags.json":
			prefs.MetadataTags = true
		}
	}
	return prefs, nil
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
func extractFileFromZip(f *zip.File, destPath string) (err error) {
	if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	rc, err := f.Open()
	if err != nil {
		return fmt.Errorf("failed to open zip entry: %w", err)
	}
	defer func() { _ = rc.Close() }()

	out, err := os.OpenFile(destPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer func() {
		if closeErr := out.Close(); err == nil && closeErr != nil {
			err = fmt.Errorf("failed to close file: %w", closeErr)
		}
	}()

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
