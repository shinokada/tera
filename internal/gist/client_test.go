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
		json.NewEncoder(w).Encode(resp)
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
		json.NewEncoder(w).Encode(resp)
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
