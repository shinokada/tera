package gist

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/zalando/go-keyring"
)

func TestTokenCRUD(t *testing.T) {
	// Setup temp home
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	// Clean up keychain before test
	_ = keyring.Delete(keychainService, keychainUser)
	defer func() {
		_ = keyring.Delete(keychainService, keychainUser)
	}()

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

	// Check where token was saved (keychain or file)
	source, err := GetTokenSource()
	if err != nil {
		t.Fatalf("GetTokenSource failed: %v", err)
	}

	switch source {
	case SourceFile:
		// If saved to file (keychain not available), check permissions
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
		if info.Mode().Perm() != 0600 {
			t.Logf("Warning: Expected file permissions 0600, got %o", info.Mode().Perm())
		}
	case SourceKeychain:
		// Token is in keychain, verify it
		keychainToken, err := keyring.Get(keychainService, keychainUser)
		if err != nil {
			t.Fatalf("Token should be in keychain: %v", err)
		}
		if keychainToken != token {
			t.Errorf("Expected keychain token '%s', got '%s'", token, keychainToken)
		}
	}

	// Delete
	if _, err := DeleteToken(); err != nil {
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

func TestEnvironmentVariableToken(t *testing.T) {
	// Clean state
	_ = os.Unsetenv(envVarName)
	_ = keyring.Delete(keychainService, keychainUser)
	defer func() {
		_ = os.Unsetenv(envVarName)
		_ = keyring.Delete(keychainService, keychainUser)
	}()

	// Set environment variable
	envToken := "ghp_env_test_token"
	if err := os.Setenv(envVarName, envToken); err != nil {
		t.Fatalf("Failed to set environment variable: %v", err)
	}

	// Load should return environment variable
	loaded, err := LoadToken()
	if err != nil {
		t.Fatalf("Failed to load token from env: %v", err)
	}

	if loaded != envToken {
		t.Errorf("Expected env token '%s', got '%s'", envToken, loaded)
	}

	// HasToken should return true
	if !HasToken() {
		t.Error("HasToken should return true when env var is set")
	}

	// Source should be environment
	source, err := GetTokenSource()
	if err != nil {
		t.Fatalf("GetTokenSource failed: %v", err)
	}
	if source != SourceEnvironment {
		t.Errorf("Expected SourceEnvironment, got %v", source)
	}
}

func TestFileTokenFallback(t *testing.T) {
	// Setup temp home
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	// Clean environment and keychain
	_ = os.Unsetenv(envVarName)
	_ = keyring.Delete(keychainService, keychainUser)
	defer func() {
		_ = keyring.Delete(keychainService, keychainUser)
	}()

	// Create a file token directly
	fileToken := "ghp_file_test_token"
	err := saveFileToken(fileToken)
	if err != nil {
		t.Fatalf("Failed to create file token: %v", err)
	}

	// LoadToken should find the file token
	loaded, err := LoadToken()
	if err != nil {
		t.Fatalf("Failed to load file token: %v", err)
	}

	if loaded != fileToken {
		t.Errorf("Expected file token '%s', got '%s'", fileToken, loaded)
	}

	// Source should be file
	source, err := GetTokenSource()
	if err != nil {
		t.Fatalf("GetTokenSource failed: %v", err)
	}
	if source != SourceFile {
		t.Errorf("Expected SourceFile, got %v", source)
	}

	// Clean up
	_ = deleteFileToken()
}
