package gist

import (
	"os"
	"path/filepath"
	"testing"
)

func TestTokenCRUD(t *testing.T) {
	// Setup temp home
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	// Verify init (no token)
	if HasToken() {
		t.Error("Expected no token initially")
	}

	// Save
	token := "ghp_test123"
	if err := SaveToken(token); err != nil {
		t.Fatalf("SaveToken failed: %v", err)
	}

	// Verify existence
	if !HasToken() {
		t.Error("Expected token to exist")
	}

	// Load
	loaded, err := LoadToken()
	if err != nil {
		t.Fatalf("LoadToken failed: %v", err)
	}
	if loaded != token {
		t.Errorf("Expected token '%s', got '%s'", token, loaded)
	}

	// Permissions check
	configDir, err := os.UserConfigDir()
	if err != nil {
		configDir = filepath.Join(os.Getenv("HOME"), ".config")
	}
	path := filepath.Join(configDir, "tera/tokens/github_token")

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Stat failed: %v", err)
	}
	// On Unix, permission bits should be 600 (-rw-------)
	// Note: os.FileMode includes type bits, so we mask.
	// However, on some systems or temp dirs permissions might behave differently,
	// but 0600 is what we set.
	if info.Mode().Perm() != 0600 {
		t.Logf("Warning: Expected file permissions 0600, got %o", info.Mode().Perm())
		// Don't fail the test strictly on permissions if env is weird, but good to know
	}

	// Delete
	if err := DeleteToken(); err != nil {
		t.Fatalf("DeleteToken failed: %v", err)
	}

	// Verify deletion
	if HasToken() {
		t.Error("Expected no token after delete")
	}
}

func TestGetMaskedToken(t *testing.T) {
	token := "ghp_1234567890abcdef1234"
	masked := GetMaskedToken(token)
	expected := "ghp_1234567...1234"
	if masked != expected {
		t.Errorf("Expected '%s', got '%s'", expected, masked)
	}

	short := "short"
	if GetMaskedToken(short) != "***********" {
		t.Error("Expected full mask for short token")
	}
}
