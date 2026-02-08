package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSearchByName(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`[
			{
				"stationuuid": "test-uuid-1",
				"name": "Test Jazz Station",
				"url_resolved": "http://example.com/stream",
				"tags": "jazz,smooth",
				"country": "United States",
				"countrycode": "US",
				"state": "California",
				"language": "english",
				"votes": 100,
				"codec": "MP3",
				"bitrate": 128
			}
		]`))
	}))
	defer server.Close()

	// Override baseURL for testing
	oldBaseURL := baseURL
	baseURL = server.URL
	defer func() { baseURL = oldBaseURL }()

	client := NewClient()
	ctx := context.Background()

	stations, err := client.SearchByName(ctx, "Jazz")
	if err != nil {
		t.Fatalf("SearchByName failed: %v", err)
	}

	if len(stations) != 1 {
		t.Errorf("Expected 1 station, got %d", len(stations))
	}

	if stations[0].Name != "Test Jazz Station" {
		t.Errorf("Expected station name 'Test Jazz Station', got %s", stations[0].Name)
	}
}

func TestSearchByLanguage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check form data
		if err := r.ParseForm(); err != nil {
			t.Fatalf("Failed to parse form: %v", err)
		}

		language := r.FormValue("language")
		if language != "english" {
			t.Errorf("Expected language 'english', got '%s'", language)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`[
			{
				"stationuuid": "test-uuid-1",
				"name": "English Station",
				"url_resolved": "http://example.com/stream",
				"language": "english",
				"votes": 50
			}
		]`))
	}))
	defer server.Close()

	oldBaseURL := baseURL
	baseURL = server.URL
	defer func() { baseURL = oldBaseURL }()

	client := NewClient()
	ctx := context.Background()

	stations, err := client.SearchByLanguage(ctx, "english")
	if err != nil {
		t.Fatalf("SearchByLanguage failed: %v", err)
	}

	if len(stations) != 1 {
		t.Errorf("Expected 1 station, got %d", len(stations))
	}
}

func TestSearchByCountry(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			t.Fatalf("Failed to parse form: %v", err)
		}

		country := r.FormValue("country")
		if country != "US" {
			t.Errorf("Expected country 'US', got '%s'", country)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`[
			{
				"stationuuid": "test-uuid-1",
				"name": "US Station",
				"url_resolved": "http://example.com/stream",
				"country": "United States",
				"countrycode": "US",
				"votes": 75
			}
		]`))
	}))
	defer server.Close()

	oldBaseURL := baseURL
	baseURL = server.URL
	defer func() { baseURL = oldBaseURL }()

	client := NewClient()
	ctx := context.Background()

	stations, err := client.SearchByCountry(ctx, "US")
	if err != nil {
		t.Fatalf("SearchByCountry failed: %v", err)
	}

	if len(stations) != 1 {
		t.Errorf("Expected 1 station, got %d", len(stations))
	}
}

func TestSearchByState(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			t.Fatalf("Failed to parse form: %v", err)
		}

		state := r.FormValue("state")
		if state != "California" {
			t.Errorf("Expected state 'California', got '%s'", state)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`[
			{
				"stationuuid": "test-uuid-1",
				"name": "California Station",
				"url_resolved": "http://example.com/stream",
				"state": "California",
				"votes": 60
			}
		]`))
	}))
	defer server.Close()

	oldBaseURL := baseURL
	baseURL = server.URL
	defer func() { baseURL = oldBaseURL }()

	client := NewClient()
	ctx := context.Background()

	stations, err := client.SearchByState(ctx, "California")
	if err != nil {
		t.Fatalf("SearchByState failed: %v", err)
	}

	if len(stations) != 1 {
		t.Errorf("Expected 1 station, got %d", len(stations))
	}
}

func TestBuildFormValues(t *testing.T) {
	client := NewClient()

	tests := []struct {
		name     string
		params   SearchParams
		expected map[string]string
	}{
		{
			name: "Basic tag search",
			params: SearchParams{
				Tag: "jazz",
			},
			expected: map[string]string{
				"tag":        "jazz",
				"order":      "votes",
				"limit":      "100",
				"hidebroken": "true",
			},
		},
		{
			name: "Search with custom limit",
			params: SearchParams{
				Name:  "BBC",
				Limit: 50,
			},
			expected: map[string]string{
				"name":       "BBC",
				"order":      "votes",
				"limit":      "50",
				"hidebroken": "true",
			},
		},
		{
			name: "Search with offset",
			params: SearchParams{
				Tag:    "rock",
				Offset: 100,
			},
			expected: map[string]string{
				"tag":        "rock",
				"order":      "votes",
				"limit":      "100",
				"offset":     "100",
				"hidebroken": "true",
			},
		},
		{
			name: "Search with reverse order",
			params: SearchParams{
				Language: "spanish",
				Reverse:  true,
			},
			expected: map[string]string{
				"language":   "spanish",
				"order":      "votes",
				"reverse":    "true",
				"limit":      "100",
				"hidebroken": "true",
			},
		},
		{
			name: "Advanced search with multiple params",
			params: SearchParams{
				Name:    "Radio",
				Tag:     "news",
				Country: "US",
				Order:   "bitrate",
			},
			expected: map[string]string{
				"name":       "Radio",
				"tag":        "news",
				"country":    "US",
				"order":      "bitrate",
				"limit":      "100",
				"hidebroken": "true",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formValues := client.buildFormValues(tt.params)

			for key, expectedValue := range tt.expected {
				actualValue := formValues.Get(key)
				if actualValue != expectedValue {
					t.Errorf("For param %s, expected %s, got %s", key, expectedValue, actualValue)
				}
			}
		})
	}
}

func TestSearchAdvanced(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			t.Fatalf("Failed to parse form: %v", err)
		}

		// Check that multiple parameters are present
		name := r.FormValue("name")
		tag := r.FormValue("tag")

		if name == "" || tag == "" {
			t.Error("Expected both name and tag parameters in advanced search")
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`[
			{
				"stationuuid": "test-uuid-1",
				"name": "Test Station",
				"url_resolved": "http://example.com/stream",
				"tags": "test",
				"votes": 10,
				"codec": "MP3",
				"bitrate": 128
			}
		]`))
	}))
	defer server.Close()

	oldBaseURL := baseURL
	baseURL = server.URL
	defer func() { baseURL = oldBaseURL }()

	client := NewClient()
	ctx := context.Background()

	params := SearchParams{
		Name:     "Test",
		Tag:      "news",
		Language: "english",
		Country:  "US",
	}

	stations, err := client.SearchAdvanced(ctx, params)
	if err != nil {
		t.Fatalf("SearchAdvanced failed: %v", err)
	}

	if len(stations) != 1 {
		t.Errorf("Expected 1 station, got %d", len(stations))
	}

	if stations[0].Name != "Test Station" {
		t.Errorf("Expected station name 'Test Station', got %s", stations[0].Name)
	}
}

func TestSearchErrorHandling(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	oldBaseURL := baseURL
	baseURL = server.URL
	defer func() { baseURL = oldBaseURL }()

	client := NewClient()
	ctx := context.Background()

	_, err := client.SearchByName(ctx, "test")
	if err == nil {
		t.Error("Expected error for server error, got nil")
	}
}

func TestSearchTrimming(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			t.Fatalf("Failed to parse form: %v", err)
		}

		// Check that query params are trimmed
		tag := r.FormValue("tag")
		if strings.Contains(tag, " ") {
			t.Errorf("Tag should be trimmed, got '%s'", tag)
		}
		if tag != "jazz" {
			t.Errorf("Expected trimmed tag 'jazz', got '%s'", tag)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`[]`))
	}))
	defer server.Close()

	oldBaseURL := baseURL
	baseURL = server.URL
	defer func() { baseURL = oldBaseURL }()

	client := NewClient()
	ctx := context.Background()

	// Test with whitespace
	params := SearchParams{
		Tag: "  jazz  ",
	}
	_, err := client.Search(ctx, params)
	if err != nil {
		t.Fatalf("Search with whitespace failed: %v", err)
	}
}

func TestSearchWithContext(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate slow response
		// In real scenario, this would be cancelled by context
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`[]`))
	}))
	defer server.Close()

	oldBaseURL := baseURL
	baseURL = server.URL
	defer func() { baseURL = oldBaseURL }()

	client := NewClient()
	ctx := context.Background()

	_, err := client.SearchByTag(ctx, "jazz")
	if err != nil {
		t.Fatalf("SearchByTag with context failed: %v", err)
	}
}
