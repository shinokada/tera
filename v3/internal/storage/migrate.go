package storage

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// MigrateDataFromV2 migrates user data from v2 structure to v3 structure
// v2 structure: all files at root of tera config dir
// v3 structure: data/ subdirectory with favorites/, cache/, and root-level data files
func MigrateDataFromV2(v2ConfigDir string) error {
	// Define v3 data directory
	v3DataDir := filepath.Join(v2ConfigDir, "data")

	// Ensure v3 data directory structure exists
	if err := os.MkdirAll(v3DataDir, 0755); err != nil {
		return fmt.Errorf("failed to create data directory: %w", err)
	}

	// Create cache subdirectory
	cacheDir := filepath.Join(v3DataDir, "cache")
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return fmt.Errorf("failed to create cache directory: %w", err)
	}

	// Map of files to migrate: oldPath -> newPath
	filesToMove := map[string]string{
		"blocklist.json":      filepath.Join(v3DataDir, "blocklist.json"),
		"voted_stations.json": filepath.Join(v3DataDir, "voted_stations.json"),
	}

	// Migrate individual files
	for oldFile, newPath := range filesToMove {
		oldPath := filepath.Join(v2ConfigDir, oldFile)
		if err := moveFileIfExists(oldPath, newPath); err != nil {
			return fmt.Errorf("failed to move %s: %w", oldFile, err)
		}
	}

	// Migrate favorites directory
	oldFavoritesDir := filepath.Join(v2ConfigDir, "favorites")
	newFavoritesDir := filepath.Join(v3DataDir, "favorites")
	if err := moveDirIfExists(oldFavoritesDir, newFavoritesDir); err != nil {
		return fmt.Errorf("failed to move favorites: %w", err)
	}

	// Move search-history.json if it exists (could be in favorites/ or root)
	searchHistoryPaths := []string{
		filepath.Join(v2ConfigDir, "favorites", "search-history.json"),
		filepath.Join(v2ConfigDir, "search-history.json"),
		filepath.Join(newFavoritesDir, "search-history.json"), // Also check if it got moved with favorites
	}
	for _, oldPath := range searchHistoryPaths {
		newPath := filepath.Join(cacheDir, "search-history.json")
		if err := moveFileIfExists(oldPath, newPath); err == nil {
			// Successfully moved, stop trying other paths
			break
		}
		// Don't fail migration if search history move fails - just continue to next path
	}

	return nil
}

// moveFileIfExists moves a file from src to dst if src exists
// If dst already exists, it will be overwritten
func moveFileIfExists(src, dst string) error {
	// Check if source exists
	if _, err := os.Stat(src); os.IsNotExist(err) {
		return nil // Source doesn't exist, nothing to do
	}

	// Check if destination already exists
	if _, err := os.Stat(dst); err == nil {
		// Destination exists, remove it first
		if err := os.Remove(dst); err != nil {
			return fmt.Errorf("failed to remove existing file %s: %w", dst, err)
		}
	}

	// Ensure destination directory exists
	dstDir := filepath.Dir(dst)
	if err := os.MkdirAll(dstDir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dstDir, err)
	}

	// Move the file
	if err := os.Rename(src, dst); err != nil {
		// If rename fails (possibly across filesystems), try copy + delete
		if err := copyFile(src, dst); err != nil {
			return fmt.Errorf("failed to copy file: %w", err)
		}
		if err := os.Remove(src); err != nil {
			return fmt.Errorf("failed to remove source file: %w", err)
		}
	}

	return nil
}

// moveDirIfExists moves a directory from src to dst if src exists
// If dst already exists, files are merged (src files overwrite dst files)
func moveDirIfExists(src, dst string) error {
	// Check if source exists
	srcInfo, err := os.Stat(src)
	if os.IsNotExist(err) {
		return nil // Source doesn't exist, nothing to do
	}
	if err != nil {
		return fmt.Errorf("failed to stat source directory: %w", err)
	}
	if !srcInfo.IsDir() {
		return fmt.Errorf("source is not a directory: %s", src)
	}

	// Ensure destination directory exists
	if err := os.MkdirAll(dst, 0755); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	// Read source directory
	entries, err := os.ReadDir(src)
	if err != nil {
		return fmt.Errorf("failed to read source directory: %w", err)
	}

	// Move each file
	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			// Recursively move subdirectories
			if err := moveDirIfExists(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			// Move file
			if err := moveFileIfExists(srcPath, dstPath); err != nil {
				return err
			}
		}
	}

	// Remove source directory if it's empty
	if isEmpty, _ := isDirEmpty(src); isEmpty {
		if err := os.Remove(src); err != nil {
			// Log warning but don't fail
			fmt.Fprintf(os.Stderr, "Warning: could not remove old directory %s: %v\n", src, err)
		}
	}

	return nil
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := srcFile.Close(); closeErr != nil && err == nil {
			err = closeErr
		}
	}()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := dstFile.Close(); closeErr != nil && err == nil {
			err = closeErr
		}
	}()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return err
	}

	// Sync to ensure data is written
	return dstFile.Sync()
}

// isDirEmpty checks if a directory is empty
func isDirEmpty(dir string) (bool, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return false, err
	}
	return len(entries) == 0, nil
}
