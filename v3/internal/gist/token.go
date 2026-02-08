package gist

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	tokenDirName  = "tera/tokens"
	tokenFileName = "github_token"
)

// getTokenPath returns the full path to the token file
func getTokenPath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user config directory: %w", err)
	}
	return filepath.Join(configDir, tokenDirName, tokenFileName), nil
}

// SaveToken saves the token to a secure file
func SaveToken(token string) error {
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

// LoadToken loads the token from the secure file
func LoadToken() (string, error) {
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

// HasToken checks if a token exists
func HasToken() bool {
	token, err := LoadToken()
	return err == nil && token != ""
}

// DeleteToken removes the stored token
func DeleteToken() error {
	tokenPath, err := getTokenPath()
	if err != nil {
		return err
	}

	if err := os.Remove(tokenPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete token file: %w", err)
	}

	return nil
}

// ValidateTokenWithClient verifies the token using the GitHub API
func ValidateTokenWithClient(token string) (string, error) {
	client := NewClient(token)
	return client.ValidateToken()
}

// GetMaskedToken returns a masked version of the token for display
func GetMaskedToken(token string) string {
	if len(token) <= 15 {
		return "***********"
	}
	return fmt.Sprintf("%s...%s", token[:11], token[len(token)-4:])
}
