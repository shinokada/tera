package gist

import (
	"os"
	"testing"

	"github.com/zalando/go-keyring"
)

func TestKeychainTokenStorage(t *testing.T) {
	// Setup temp directory to isolate config paths
	tmp := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmp)
	t.Setenv("HOME", tmp)
	t.Setenv("APPDATA", tmp) // Windows

	// Cleanup before and after tests
	defer func() {
		_ = keyring.Delete(keychainService, keychainUser)
	}()
	// Initial cleanup
	_ = keyring.Delete(keychainService, keychainUser)

	t.Run("SaveAndLoadFromKeychain", func(t *testing.T) {
		testToken := "test_token_123456789"

		// Save token
		err := SaveToken(testToken)
		if err != nil {
			t.Skipf("Keychain not available: %v", err)
		}

		// Load token
		loaded, err := LoadToken()
		if err != nil {
			t.Fatalf("Failed to load token: %v", err)
		}

		if loaded != testToken {
			t.Errorf("Expected token %s, got %s", testToken, loaded)
		}

		// Cleanup
		_, _ = DeleteToken()
	})

	t.Run("EnvironmentVariableTakesPrecedence", func(t *testing.T) {
		envToken := "ghp_env_token"
		keychainToken := "ghp_keychain_token"

		// Clean environment first
		_ = os.Unsetenv(envVarName)

		// Set keychain token
		err := SaveToken(keychainToken)
		if err != nil {
			t.Skipf("Keychain not available: %v", err)
		}

		// Set environment variable
		if err := os.Setenv(envVarName, envToken); err != nil {
			t.Fatalf("Failed to set environment variable: %v", err)
		}
		defer func() {
			_ = os.Unsetenv(envVarName)
		}()

		// Load should return environment variable
		loaded, err := LoadToken()
		if err != nil {
			t.Fatalf("Failed to load token: %v", err)
		}

		if loaded != envToken {
			t.Errorf("Expected env token %s, got %s", envToken, loaded)
		}

		// Cleanup
		_ = os.Unsetenv(envVarName)
		_, _ = DeleteToken()
	})

	t.Run("GetTokenSource", func(t *testing.T) {
		// Clean state
		_, _ = DeleteToken()
		_ = os.Unsetenv(envVarName)

		// No token
		source, err := GetTokenSource()
		if err != nil {
			t.Fatalf("GetTokenSource failed: %v", err)
		}
		if source != SourceNone {
			t.Errorf("Expected SourceNone, got %v", source)
		}

		// Keychain token
		err = SaveToken("ghp_test")
		if err != nil {
			t.Skipf("Keychain not available: %v", err)
		}

		source, _ = GetTokenSource()
		if source != SourceKeychain {
			t.Skipf("Keychain backend unavailable; token saved to %v", source)
		}

		// Environment variable takes precedence
		if err := os.Setenv(envVarName, "ghp_env"); err != nil {
			t.Fatalf("Failed to set environment variable: %v", err)
		}
		defer func() {
			_ = os.Unsetenv(envVarName)
		}()

		source, _ = GetTokenSource()
		if source != SourceEnvironment {
			t.Errorf("Expected SourceEnvironment, got %v", source)
		}

		// Cleanup
		_ = os.Unsetenv(envVarName)
		_, _ = DeleteToken()
	})

	t.Run("DeleteToken", func(t *testing.T) {
		testToken := "ghp_delete_test"

		// Save token
		err := SaveToken(testToken)
		if err != nil {
			t.Skipf("Keychain not available: %v", err)
		}

		// Delete token
		_, err = DeleteToken()
		if err != nil {
			t.Fatalf("Failed to delete token: %v", err)
		}

		// Verify deleted from keychain
		_, err = keyring.Get(keychainService, keychainUser)
		if err == nil {
			t.Error("Token should be deleted from keychain")
		}
	})

	t.Run("HasTokenWithKeychain", func(t *testing.T) {
		// Clean state
		_, _ = DeleteToken()
		_ = os.Unsetenv(envVarName)

		// Should not have token
		if HasToken() {
			t.Error("Should not have token")
		}

		// Save token
		err := SaveToken("ghp_test")
		if err != nil {
			t.Skipf("Keychain not available: %v", err)
		}

		// Should have token
		if !HasToken() {
			t.Error("Should have token")
		}

		// Cleanup
		_, _ = DeleteToken()
	})
}

func TestTokenMigration(t *testing.T) {
	// Setup temp directory for file token
	tmpDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmpDir)
	t.Setenv("HOME", tmpDir)
	t.Setenv("APPDATA", tmpDir) // Windows

	// Cleanup keychain
	defer func() {
		_ = keyring.Delete(keychainService, keychainUser)
	}()
	_ = keyring.Delete(keychainService, keychainUser)

	t.Run("MigrateFileToKeychain", func(t *testing.T) {
		// Create a file-based token
		fileToken := "ghp_file_token_12345"
		err := saveFileToken(fileToken)
		if err != nil {
			t.Fatalf("Failed to create file token: %v", err)
		}

		// Verify file token exists
		loaded, err := loadFileToken()
		if err != nil || loaded != fileToken {
			t.Fatalf("File token not created properly: %v", err)
		}

		// Migrate to keychain
		result, err := MigrateFileTokenToKeychain()
		if err != nil {
			t.Skipf("Keychain not available for migration: %v", err)
		}

		if !result.Migrated {
			t.Error("Expected migration to occur")
		}

		// Verify token is now in keychain
		keychainToken, err := keyring.Get(keychainService, keychainUser)
		if err != nil {
			t.Fatalf("Token should be in keychain after migration: %v", err)
		}
		if keychainToken != fileToken {
			t.Errorf("Expected keychain token %s, got %s", fileToken, keychainToken)
		}

		// Verify file token is deleted
		fileTokenAfter, _ := loadFileToken()
		if fileTokenAfter != "" {
			t.Error("File token should be deleted after migration")
		}

		// Cleanup
		_, _ = DeleteToken()
	})

	t.Run("NoMigrationIfAlreadyInKeychain", func(t *testing.T) {
		// Put token in keychain
		err := SaveToken("ghp_already_in_keychain")
		if err != nil {
			t.Skipf("Keychain not available: %v", err)
		}

		// Verify token actually landed in keychain (not file fallback)
		source, _ := GetTokenSource()
		if source != SourceKeychain {
			_, _ = DeleteToken()
			t.Skipf("Keychain backend unavailable; token saved to %v", source)
		}

		// Try to migrate
		result, err := MigrateFileTokenToKeychain()
		if err != nil {
			t.Fatalf("Migration check failed: %v", err)
		}

		if result.Migrated {
			t.Error("Should not migrate if token already in keychain")
		}

		// Cleanup
		_, _ = DeleteToken()
	})
}
