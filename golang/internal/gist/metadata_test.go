package gist

import (
	"testing"
	"time"
)

func TestMetadataCRUD(t *testing.T) {
	// Setup temp home
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	// Verify init (empty)
	count, err := GetGistCount()
	if err != nil {
		t.Fatalf("GetGistCount failed: %v", err)
	}
	if count != 0 {
		t.Errorf("Expected 0 gists, got %d", count)
	}

	// Create
	meta := &GistMetadata{
		ID:          "1",
		Description: "Desc 1",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	if err := SaveMetadata(meta); err != nil {
		t.Fatalf("SaveMetadata failed: %v", err)
	}

	// Read
	gists, err := GetAllGists()
	if err != nil {
		t.Fatalf("GetAllGists failed: %v", err)
	}
	if len(gists) != 1 {
		t.Errorf("Expected 1 gist, got %d", len(gists))
	}
	if gists[0].ID != "1" {
		t.Errorf("Expected ID '1', got '%s'", gists[0].ID)
	}

	// Update
	if err := UpdateMetadata("1", "Updated Desc"); err != nil {
		t.Fatalf("UpdateMetadata failed: %v", err)
	}
	updated, err := GetGistByID("1")
	if err != nil {
		t.Fatalf("GetGistByID failed: %v", err)
	}
	if updated.Description != "Updated Desc" {
		t.Errorf("Expected description 'Updated Desc', got '%s'", updated.Description)
	}

	// Delete
	if err := DeleteMetadata("1"); err != nil {
		t.Fatalf("DeleteMetadata failed: %v", err)
	}
	count, err = GetGistCount()
	if err != nil {
		t.Fatalf("GetGistCount failed: %v", err)
	}
	if count != 0 {
		t.Errorf("Expected 0 gists, got %d", count)
	}
}
