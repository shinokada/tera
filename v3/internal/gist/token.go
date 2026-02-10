package gist

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/zalando/go-keyring"
)

const (
	tokenDirName    = "tera/tokens"
	tokenFileName   = "github_token"
	keychainService = "tera"
	keychainUser    = "github-token"
	envVarName      = "TERA_GITHUB_TOKEN"
)

// TokenSource indicates where the token was loaded from
type TokenSource string

const (
	SourceKeychain    TokenSource = "keychain"
	SourceEnvironment TokenSource = "environment"
	SourceFile        TokenSource = "file"
	SourceNone        TokenSource = "none"
)

// MigrationResult contains the result of a token migration operation
type MigrationResult struct {
	Migrated bool
	Warning  error // Non-fatal warning (e.g., couldn't delete old file)
}

// getTokenPath returns the full path to the token file (legacy)
func getTokenPath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user config directory: %w", err)
	}
	return filepath.Join(configDir, tokenDirName, tokenFileName), nil
}

// SaveToken saves the token to the OS keychain
// Falls back to file storage if keychain is unavailable
// Trims whitespace and validates token before saving
func SaveToken(token string) error {
	token = strings.TrimSpace(token)
	if token == "" {
		return fmt.Errorf("cannot save empty token")
	}

	// Try keychain first
	err := keyring.Set(keychainService, keychainUser, token)
	if err == nil {
		// Successfully saved to keychain
		// Clean up old file-based token if it exists
		_ = deleteFileToken()
		return nil
	}

	// Keychain failed, fall back to file storage
	return saveFileToken(token)
}

// LoadToken loads the token from available sources in priority order:
// 1. Environment variable (TERA_GITHUB_TOKEN)
// 2. OS Keychain
// 3. File storage (legacy)
func LoadToken() (string, error) {
	// 1. Check environment variable first (highest priority)
	if envToken := os.Getenv(envVarName); envToken != "" {
		token := strings.TrimSpace(envToken)
		if token != "" {
			return token, nil
		}
	}

	// 2. Try keychain
	token, err := keyring.Get(keychainService, keychainUser)
	if err == nil && strings.TrimSpace(token) != "" {
		return strings.TrimSpace(token), nil
	}

	// 3. Fall back to file storage (legacy)
	return loadFileToken()
}

// GetTokenSource returns where the token is currently stored/available
func GetTokenSource() (TokenSource, error) {
	// Check environment variable
	if envToken := strings.TrimSpace(os.Getenv(envVarName)); envToken != "" {
		return SourceEnvironment, nil
	}

	// Check keychain
	keychainToken, err := keyring.Get(keychainService, keychainUser)
	if err == nil && strings.TrimSpace(keychainToken) != "" {
		return SourceKeychain, nil
	}

	// Check file
	token, err := loadFileToken()
	if err == nil && token != "" {
		return SourceFile, nil
	}

	return SourceNone, nil
}

// HasToken checks if a token exists in any source
func HasToken() bool {
	token, err := LoadToken()
	return err == nil && token != ""
}

// DeleteToken removes the stored token from keychain and file
// Does not remove environment variable (user must unset it themselves)
// Returns the active token source after deletion (for user awareness)
func DeleteToken() (activeSource TokenSource, err error) {
	var errs []error

	// Delete from keychain
	delErr := keyring.Delete(keychainService, keychainUser)
	if delErr != nil && !errors.Is(delErr, keyring.ErrNotFound) {
		errs = append(errs, fmt.Errorf("keychain: %w", delErr))
	}

	// Delete from file
	if delErr := deleteFileToken(); delErr != nil {
		errs = append(errs, fmt.Errorf("file: %w", delErr))
	}

	// Check what's still active (e.g., env var)
	activeSource, _ = GetTokenSource()

	return activeSource, errors.Join(errs...)
}

// MigrateFileTokenToKeychain migrates token from file storage to keychain
// Returns MigrationResult with migration status and any non-fatal warnings
func MigrateFileTokenToKeychain() (*MigrationResult, error) {
	result := &MigrationResult{
		Migrated: false,
		Warning:  nil,
	}

	// Check if token exists in keychain already
	existingToken, err := keyring.Get(keychainService, keychainUser)
	if err == nil && strings.TrimSpace(existingToken) != "" {
		// Token already in keychain, no migration needed
		return result, nil
	}

	// Check if file token exists
	fileToken, err := loadFileToken()
	if err != nil || fileToken == "" {
		// No file token to migrate
		return result, nil
	}

	// Migrate to keychain
	err = keyring.Set(keychainService, keychainUser, fileToken)
	if err != nil {
		// Keychain not available, keep using file
		return result, fmt.Errorf("keychain unavailable, keeping file storage: %w", err)
	}

	result.Migrated = true

	// Successfully migrated, delete file token
	if err := deleteFileToken(); err != nil {
		// Token is in keychain now, but couldn't delete file
		// This is not critical, set as warning
		result.Warning = fmt.Errorf("token migrated to keychain but could not delete file: %w", err)
	}

	return result, nil
}

// File-based token functions (legacy support)

func saveFileToken(token string) error {
	tokenPath, err := getTokenPath()
	if err != nil {
		return err
	}

	// Create directory with 700 permissions
	dir := filepath.Dir(tokenPath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("failed to create token directory: %w", err)
	}

	// Save token with 600 permissions
	if err := os.WriteFile(tokenPath, []byte(token), 0600); err != nil {
		return fmt.Errorf("failed to write token file: %w", err)
	}

	return nil
}

func loadFileToken() (string, error) {
	tokenPath, err := getTokenPath()
	if err != nil {
		return "", err
	}

	content, err := os.ReadFile(tokenPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil // No token exists, not an error per se
		}
		return "", fmt.Errorf("failed to read token file: %w", err)
	}

	token := strings.TrimSpace(string(content))
	if token == "" {
		return "", fmt.Errorf("token file is empty")
	}

	return token, nil
}

func deleteFileToken() error {
	tokenPath, err := getTokenPath()
	if err != nil {
		return err
	}

	if err := os.Remove(tokenPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete token file: %w", err)
	}

	// Also try to remove the tokens directory if it's empty
	tokenDir := filepath.Dir(tokenPath)
	_ = os.Remove(tokenDir) // Ignore error, will fail if not empty

	return nil
}

// ValidateTokenWithClient verifies the token using the GitHub API
func ValidateTokenWithClient(token string) (string, error) {
	client := NewClient(token)
	return client.ValidateToken()
}

// GetMaskedToken returns a masked version of the token for display
// Shows 4 prefix + 4 suffix characters to balance security and usability
// Tokens 12 characters or shorter are completely masked for security
func GetMaskedToken(token string) string {
	if len(token) <= 12 {
		return "************"
	}
	return fmt.Sprintf("%s...%s", token[:4], token[len(token)-4:])
}
