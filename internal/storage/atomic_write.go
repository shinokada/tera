package storage

import (
	"fmt"
	"os"
	"path/filepath"
)

// atomicWriteFile writes data to path atomically using a write-to-temp-then-rename
// strategy. This prevents a mid-write process death from leaving a truncated or
// partially-written file at the destination, which could corrupt user data.
//
// The temp file is created in the same directory as the destination so that
// os.Rename is always an intra-filesystem move (guaranteed atomic on POSIX).
func atomicWriteFile(path string, data []byte, perm os.FileMode) error {
	dir := filepath.Dir(path)

	// Create the temp file in the same directory as the destination so that
	// os.Rename is always an intra-filesystem move.
	tmp, err := os.CreateTemp(dir, ".tera-write-*")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpPath := tmp.Name()

	// Always clean up the temp file on failure.
	success := false
	defer func() {
		if !success {
			_ = os.Remove(tmpPath)
		}
	}()

	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("failed to write temp file: %w", err)
	}
	if err := tmp.Sync(); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("failed to sync temp file: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("failed to close temp file: %w", err)
	}
	if err := os.Chmod(tmpPath, perm); err != nil {
		return fmt.Errorf("failed to set permissions on temp file: %w", err)
	}
	if err := os.Rename(tmpPath, path); err != nil {
		return fmt.Errorf("failed to rename temp file to destination: %w", err)
	}

	// Sync the parent directory so the rename is durable. Some platforms
	// (macOS, certain Linux configurations) return EINVAL for Sync on a
	// directory; we ignore that error since the rename itself already succeeded.
	if dirHandle, err := os.Open(dir); err == nil {
		_ = dirHandle.Sync() // best-effort; ignore EINVAL on some platforms
		_ = dirHandle.Close()
	}

	success = true
	return nil
}
