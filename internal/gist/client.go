package gist

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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

// do plays the request and decodes the response
func (c *Client) do(req *http.Request, v interface{}) error {
	req.Header.Set("Authorization", "token "+c.token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API error: %s (status: %d, body: %s)", resp.Status, resp.StatusCode, string(body))
	}

	if v != nil {
		if err := json.NewDecoder(resp.Body).Decode(v); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}
	}

	return nil
}
