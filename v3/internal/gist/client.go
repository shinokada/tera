package gist

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const defaultBaseURL = "https://api.github.com"

// Client handles communication with the GitHub Gist API
type Client struct {
	token      string
	baseURL    string
	httpClient *http.Client
}

// NewClient creates a new GitHub Gist API client
func NewClient(token string) *Client {
	return &Client{
		token:   token,
		baseURL: defaultBaseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Gist represents a GitHub Gist
type Gist struct {
	ID          string              `json:"id"`
	URL         string              `json:"html_url"`
	Description string              `json:"description"`
	Public      bool                `json:"public"`
	CreatedAt   time.Time           `json:"created_at"`
	UpdatedAt   time.Time           `json:"updated_at"`
	Files       map[string]GistFile `json:"files,omitempty"`
}

// GistFile represents a file within a Gist
type GistFile struct {
	Filename string `json:"filename,omitempty"`
	Content  string `json:"content,omitempty"`
	RawURL   string `json:"raw_url,omitempty"`
}

// CreateGist creates a new gist with the provided files
// If public is true, the gist will be publicly visible; otherwise it will be secret
func (c *Client) CreateGist(description string, files map[string]string, public bool) (*Gist, error) {
	gistFiles := make(map[string]GistFile)
	for filename, content := range files {
		gistFiles[filename] = GistFile{
			Content: content,
		}
	}

	payload := struct {
		Description string              `json:"description"`
		Public      bool                `json:"public"`
		Files       map[string]GistFile `json:"files"`
	}{
		Description: description,
		Public:      public,
		Files:       gistFiles,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal gist payload: %w", err)
	}

	req, err := http.NewRequest("POST", c.baseURL+"/gists", bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	var gist Gist
	if err := c.do(req, &gist); err != nil {
		return nil, err
	}

	return &gist, nil
}

// ListGists lists all gists for the authenticated user
func (c *Client) ListGists() ([]*Gist, error) {
	req, err := http.NewRequest("GET", c.baseURL+"/gists", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	var gists []*Gist
	if err := c.do(req, &gists); err != nil {
		return nil, err
	}

	return gists, nil
}

// UpdateGist updates the description of an existing gist
func (c *Client) UpdateGist(gistID, description string) error {
	payload := struct {
		Description string `json:"description"`
	}{
		Description: description,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal update payload: %w", err)
	}

	req, err := http.NewRequest("PATCH", fmt.Sprintf("%s/gists/%s", c.baseURL, gistID), bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	return c.do(req, nil)
}

// UpdateGistFiles updates or replaces the files of an existing gist.
// Each key in files is the filename; the value is the new content.
// Passing an empty string as content deletes that file from the gist.
func (c *Client) UpdateGistFiles(gistID string, files map[string]string) error {
	gistFiles := make(map[string]GistFile)
	for filename, content := range files {
		gistFiles[filename] = GistFile{Content: content}
	}

	payload := struct {
		Files map[string]GistFile `json:"files"`
	}{
		Files: gistFiles,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal update payload: %w", err)
	}

	req, err := http.NewRequest("PATCH", fmt.Sprintf("%s/gists/%s", c.baseURL, gistID), bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	return c.do(req, nil)
}

// DeleteGist deletes a gist
func (c *Client) DeleteGist(gistID string) error {
	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/gists/%s", c.baseURL, gistID), nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	return c.do(req, nil)
}

// GetGist retrieves a specific gist
func (c *Client) GetGist(gistID string) (*Gist, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/gists/%s", c.baseURL, gistID), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	var gist Gist
	if err := c.do(req, &gist); err != nil {
		return nil, err
	}

	return &gist, nil
}

// ValidateToken checks if the token is valid and returns the username
func (c *Client) ValidateToken() (string, error) {
	req, err := http.NewRequest("GET", c.baseURL+"/user", nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	var user struct {
		Login string `json:"login"`
	}
	if err := c.do(req, &user); err != nil {
		return "", err
	}

	return user.Login, nil
}

// GetGistPublic fetches a public gist without requiring authentication
// This is useful for importing gists shared by other users
func GetGistPublic(gistID string) (*Gist, error) {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/gists/%s", defaultBaseURL, gistID), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "tera-radio-player")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == 404 {
		return nil, fmt.Errorf("gist not found (may be private or invalid ID)")
	}

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<10))
		return nil, fmt.Errorf("API error: %s (status: %d, body: %s)", resp.Status, resp.StatusCode, string(body))
	}

	var gist Gist
	if err := json.NewDecoder(io.LimitReader(resp.Body, 10<<20)).Decode(&gist); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &gist, nil
}

// ParseGistURL extracts the gist ID from various URL formats
// Supported formats:
// - https://gist.github.com/username/gist_id
// - https://gist.githubusercontent.com/username/gist_id/...
// - gist_id (raw ID)
func ParseGistURL(input string) (string, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return "", fmt.Errorf("empty input")
	}

	// Check if it's a URL
	if strings.HasPrefix(input, "https://gist.github.com/") ||
		strings.HasPrefix(input, "http://gist.github.com/") ||
		strings.HasPrefix(input, "https://gist.githubusercontent.com/") {
		// Parse the URL to extract path
		parts := strings.Split(input, "/")
		// URL format: https://gist.github.com/username/gist_id[/...]
		// We need at least 5 parts: https:, "", gist.github.com, username, gist_id
		if len(parts) >= 5 {
			gistID := parts[4]
			// Remove any query parameters or fragments
			if idx := strings.Index(gistID, "?"); idx != -1 {
				gistID = gistID[:idx]
			}
			if idx := strings.Index(gistID, "#"); idx != -1 {
				gistID = gistID[:idx]
			}
			if gistID != "" {
				return gistID, nil
			}
		}
		return "", fmt.Errorf("invalid gist URL format")
	}

	// Assume it's a raw gist ID - validate it looks like a hex string
	// GitHub gist IDs are 32-character hex strings
	if len(input) >= 20 && len(input) <= 40 {
		for _, c := range input {
			isDigit := c >= '0' && c <= '9'
			isLowerHex := c >= 'a' && c <= 'f'
			isUpperHex := c >= 'A' && c <= 'F'
			if !isDigit && !isLowerHex && !isUpperHex {
				return "", fmt.Errorf("invalid gist ID format")
			}
		}
		return input, nil
	}

	return "", fmt.Errorf("invalid gist URL or ID")
}

// do executes the request and decodes the response.
func (c *Client) do(req *http.Request, v interface{}) error {
	req.Header.Set("Authorization", "token "+c.token)
	req.Header.Set("User-Agent", "tera-radio-player")
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode >= 400 {
		// Limit error body to 1 KB — enough for diagnostics without risking
		// memory exhaustion or inadvertently logging token material from proxies.
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<10))
		return fmt.Errorf("API error: %s (status: %d, body: %s)", resp.Status, resp.StatusCode, string(body))
	}

	if v != nil {
		// Limit response to 10 MB to prevent memory exhaustion.
		if err := json.NewDecoder(io.LimitReader(resp.Body, 10<<20)).Decode(v); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}
	}

	return nil
}
