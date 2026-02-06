package gist

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const metadataFileName = "gist_metadata.json"

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

// GetAllGists retrieves all stored gist metadata
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

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write metadata file: %w", err)
	}

	return nil
}

// SaveMetadata adds or updates a gist metadata
func SaveMetadata(metadata *GistMetadata) error {
	gists, err := GetAllGists()
	if err != nil {
		return err
	}

	// Check if exists, update if so
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

// GetGistByID returns a specific gist metadata
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

// UpdateMetadata updates the description and timestamp of a gist
func UpdateMetadata(id, description string) error {
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

// DeleteMetadata removes a gist from metadata
func DeleteMetadata(id string) error {
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

// GetGistCount returns the number of saved gists
func GetGistCount() (int, error) {
	gists, err := GetAllGists()
	if err != nil {
		return 0, err
	}
	return len(gists), nil
}
