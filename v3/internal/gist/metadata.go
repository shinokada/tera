package gist

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"
)

const metadataFileName = "gist_metadata.json"

// metadataMu serialises all read-modify-write operations on gist_metadata.json.
// Without this, concurrent tea.Cmd goroutines (create, update, delete) can
// race on the GetAllGists → saveAllGists pair and silently clobber each other.
var metadataMu sync.Mutex

// GistMetadata represents the local metadata for a saved gist
type GistMetadata struct {
	ID          string    `json:"id"`
	URL         string    `json:"url"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// getMetadataPath returns the full path to the metadata file
func getMetadataPath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user config directory: %w", err)
	}
	// Using the same config directory as tokens
	return filepath.Join(configDir, "tera", metadataFileName), nil
}

// GetAllGists retrieves all stored gist metadata.
// Callers that need a consistent read-modify-write must hold metadataMu.
func GetAllGists() ([]*GistMetadata, error) {
	path, err := getMetadataPath()
	if err != nil {
		return nil, err
	}

	content, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return []*GistMetadata{}, nil
		}
		return nil, fmt.Errorf("failed to read metadata file: %w", err)
	}

	var gists []*GistMetadata
	if len(content) == 0 {
		return gists, nil
	}

	if err := json.Unmarshal(content, &gists); err != nil {
		return nil, fmt.Errorf("failed to parse metadata file: %w", err)
	}

	return gists, nil
}

// saveAllGists saves the list of gists to the metadata file
func saveAllGists(gists []*GistMetadata) error {
	path, err := getMetadataPath()
	if err != nil {
		return err
	}

	// Ensure directory exists (0700 for security - may contain sensitive tokens)
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := json.MarshalIndent(gists, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	// Atomic write: temp file + fsync + rename, matching the pattern used for
	// other persistent stores. Uses 0600 since the file contains Gist IDs that
	// grant access to private gists.
	if err := atomicWriteMetadata(path, data); err != nil {
		return fmt.Errorf("failed to write metadata file: %w", err)
	}

	return nil
}

// atomicWriteMetadata writes data to path via temp-file + fsync + rename
// so a crash mid-write cannot corrupt the destination file.
func atomicWriteMetadata(path string, data []byte) error {
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, ".gist-metadata-*.tmp")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpName := tmp.Name()

	var writeErr error
	defer func() {
		_ = tmp.Close()
		if writeErr != nil {
			_ = os.Remove(tmpName)
		}
	}()

	if _, writeErr = tmp.Write(data); writeErr != nil {
		return fmt.Errorf("failed to write temp file: %w", writeErr)
	}
	if writeErr = tmp.Sync(); writeErr != nil {
		return fmt.Errorf("failed to sync temp file: %w", writeErr)
	}
	if writeErr = tmp.Close(); writeErr != nil {
		return fmt.Errorf("failed to close temp file: %w", writeErr)
	}
	if writeErr = os.Chmod(tmpName, 0600); writeErr != nil {
		return fmt.Errorf("failed to chmod temp file: %w", writeErr)
	}
	// os.Rename is atomic on POSIX but on Windows it can fail if the destination
	// already exists (golang/go#8914). A proper atomic replacement on Windows
	// requires MoveFileExW(MOVEFILE_REPLACE_EXISTING) via syscall. For now use a
	// non-atomic fallback on Windows to avoid persistent save failures.
	writeErr = os.Rename(tmpName, path)
	if writeErr != nil {
		if runtime.GOOS == "windows" {
			writeErr = os.WriteFile(path, data, 0600)
			_ = os.Remove(tmpName)
		} else {
			return fmt.Errorf("failed to rename temp file: %w", writeErr)
		}
	}
	if writeErr != nil {
		return fmt.Errorf("failed to write metadata file: %w", writeErr)
	}
	return nil
}

// SaveMetadata adds or updates a gist metadata.
func SaveMetadata(metadata *GistMetadata) error {
	metadataMu.Lock()
	defer metadataMu.Unlock()

	gists, err := GetAllGists()
	if err != nil {
		return err
	}

	found := false
	for i, g := range gists {
		if g.ID == metadata.ID {
			gists[i] = metadata
			found = true
			break
		}
	}
	if !found {
		gists = append(gists, metadata)
	}

	return saveAllGists(gists)
}

// GetGistByID returns a specific gist metadata.
// This is a read-only lookup so it does not need the write mutex.
func GetGistByID(id string) (*GistMetadata, error) {
	gists, err := GetAllGists()
	if err != nil {
		return nil, err
	}
	for _, g := range gists {
		if g.ID == id {
			return g, nil
		}
	}
	return nil, fmt.Errorf("gist not found: %s", id)
}

// UpdateMetadata updates the description and timestamp of a gist.
func UpdateMetadata(id, description string) error {
	metadataMu.Lock()
	defer metadataMu.Unlock()

	gists, err := GetAllGists()
	if err != nil {
		return err
	}

	found := false
	for _, g := range gists {
		if g.ID == id {
			g.Description = description
			g.UpdatedAt = time.Now()
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("gist not found: %s", id)
	}

	return saveAllGists(gists)
}

// DeleteMetadata removes a gist from metadata.
func DeleteMetadata(id string) error {
	metadataMu.Lock()
	defer metadataMu.Unlock()

	gists, err := GetAllGists()
	if err != nil {
		return err
	}

	newGists := make([]*GistMetadata, 0, len(gists))
	for _, g := range gists {
		if g.ID != id {
			newGists = append(newGists, g)
		}
	}
	if len(newGists) == len(gists) {
		return fmt.Errorf("gist not found: %s", id)
	}

	return saveAllGists(newGists)
}

// GetGistCount returns the number of saved gists.
func GetGistCount() (int, error) {
	gists, err := GetAllGists()
	if err != nil {
		return 0, err
	}
	return len(gists), nil
}
