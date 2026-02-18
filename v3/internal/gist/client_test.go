package gist

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCreateGist(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected method POST, got %s", r.Method)
		}
		if r.URL.Path != "/gists" {
			t.Errorf("Expected path /gists, got %s", r.URL.Path)
		}

		var payload struct {
			Description string              `json:"description"`
			Public      bool                `json:"public"`
			Files       map[string]GistFile `json:"files"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("Failed to decode request body: %v", err)
		}

		if payload.Description != "Test Gist" {
			t.Errorf("Expected description 'Test Gist', got '%s'", payload.Description)
		}
		if payload.Public {
			t.Error("Expected public to be false")
		}
		if _, ok := payload.Files["test.txt"]; !ok {
			t.Error("Expected file 'test.txt' in payload")
		}

		// Response
		resp := Gist{
			ID:          "test-id",
			URL:         "http://example.com/gist",
			Description: "Test Gist",
			Files: map[string]GistFile{
				"test.txt": {Content: "content"},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer ts.Close()

	client := NewClient("token")
	client.baseURL = ts.URL

	files := map[string]string{
		"test.txt": "content",
	}

	gist, err := client.CreateGist("Test Gist", files, false)
	if err != nil {
		t.Fatalf("CreateGist failed: %v", err)
	}

	if gist.ID != "test-id" {
		t.Errorf("Expected ID 'test-id', got '%s'", gist.ID)
	}
}

func TestListGists(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("Expected method GET, got %s", r.Method)
		}
		resp := []*Gist{
			{ID: "1", Description: "Gist 1"},
			{ID: "2", Description: "Gist 2"},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer ts.Close()

	client := NewClient("token")
	client.baseURL = ts.URL

	gists, err := client.ListGists()
	if err != nil {
		t.Fatalf("ListGists failed: %v", err)
	}

	if len(gists) != 2 {
		t.Errorf("Expected 2 gists, got %d", len(gists))
	}
}

func TestParseGistURL(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{
			name:    "full github URL",
			input:   "https://gist.github.com/username/abc123def456789012345678901234567890",
			want:    "abc123def456789012345678901234567890",
			wantErr: false,
		},
		{
			name:    "full github URL with trailing path",
			input:   "https://gist.github.com/username/abc123def456789012345678901234567890/raw",
			want:    "abc123def456789012345678901234567890",
			wantErr: false,
		},
		{
			name:    "http URL",
			input:   "http://gist.github.com/user/abc123def456789012345678901234567890",
			want:    "abc123def456789012345678901234567890",
			wantErr: false,
		},
		{
			name:    "githubusercontent URL",
			input:   "https://gist.githubusercontent.com/user/abc123def456789012345678901234567890/raw",
			want:    "abc123def456789012345678901234567890",
			wantErr: false,
		},
		{
			name:    "raw gist ID 32 chars",
			input:   "abc123def456789012345678901234ab",
			want:    "abc123def456789012345678901234ab",
			wantErr: false,
		},
		{
			name:    "URL with query params",
			input:   "https://gist.github.com/user/abc123def456789012345678901234567890?file=test.txt",
			want:    "abc123def456789012345678901234567890",
			wantErr: false,
		},
		{
			name:    "URL with fragment",
			input:   "https://gist.github.com/user/abc123def456789012345678901234567890#file-test-txt",
			want:    "abc123def456789012345678901234567890",
			wantErr: false,
		},
		{
			name:    "empty input",
			input:   "",
			want:    "",
			wantErr: true,
		},
		{
			name:    "whitespace only",
			input:   "   ",
			want:    "",
			wantErr: true,
		},
		{
			name:    "invalid short ID",
			input:   "abc123",
			want:    "",
			wantErr: true,
		},
		{
			name:    "invalid characters in ID",
			input:   "abc123xyz456789012345678901234ab",
			want:    "",
			wantErr: true,
		},
		{
			name:    "invalid URL format",
			input:   "https://gist.github.com/",
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseGistURL(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseGistURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ParseGistURL() = %v, want %v", got, tt.want)
			}
		})
	}
}
