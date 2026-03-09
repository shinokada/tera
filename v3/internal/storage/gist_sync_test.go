package storage

import (
	"path/filepath"
	"testing"
)

func TestGistFilename(t *testing.T) {
	cases := []struct {
		relPath string
		want    string
	}{
		{"config.yaml", "config.yaml"},
		{filepath.Join("data", "blocklist.json"), "blocklist.json"},
		{filepath.Join("data", "voted_stations.json"), "voted_stations.json"},
		{filepath.Join("data", "station_ratings.json"), "ratings.json"},
		{filepath.Join("data", "station_tags.json"), "tags.json"},
		{filepath.Join("data", "station_metadata.json"), "metadata.json"},
		{filepath.Join("data", "favorites", SystemFileSearchHistory), "search-history.json"},
		{filepath.Join("data", "favorites", "Jazz.json"), "fav--Jazz.json"},
		{filepath.Join("data", "favorites", "My-80s-Rock-list.json"), "fav--My-80s-Rock-list.json"},
		{filepath.Join("data", "favorites", "Bossa-nova.json"), "fav--Bossa-nova.json"},
	}

	for _, tc := range cases {
		got := gistFilename(tc.relPath)
		if got != tc.want {
			t.Errorf("gistFilename(%q) = %q, want %q", tc.relPath, got, tc.want)
		}
	}
}

func TestGistFilenameToRelPath(t *testing.T) {
	cases := []struct {
		name string
		want string
	}{
		{"config.yaml", "config.yaml"},
		{"blocklist.json", filepath.Join("data", "blocklist.json")},
		{"voted_stations.json", filepath.Join("data", "voted_stations.json")},
		{"ratings.json", filepath.Join("data", "station_ratings.json")},
		{"tags.json", filepath.Join("data", "station_tags.json")},
		{"metadata.json", filepath.Join("data", "station_metadata.json")},
		{"search-history.json", filepath.Join("data", "favorites", SystemFileSearchHistory)},
		{"fav--Jazz.json", filepath.Join("data", "favorites", "Jazz.json")},
		{"fav--Bossa-nova.json", filepath.Join("data", "favorites", "Bossa-nova.json")},
		{"unknown.json", ""},
	}

	for _, tc := range cases {
		got := gistFilenameToRelPath(tc.name)
		if got != tc.want {
			t.Errorf("gistFilenameToRelPath(%q) = %q, want %q", tc.name, got, tc.want)
		}
	}
}

func TestGistFilenameRoundTrip(t *testing.T) {
	relPaths := []string{
		"config.yaml",
		filepath.Join("data", "blocklist.json"),
		filepath.Join("data", "voted_stations.json"),
		filepath.Join("data", "station_ratings.json"),
		filepath.Join("data", "station_tags.json"),
		filepath.Join("data", "station_metadata.json"),
		filepath.Join("data", "favorites", SystemFileSearchHistory),
		filepath.Join("data", "favorites", "Jazz.json"),
		filepath.Join("data", "favorites", "Smooth-Jazz.json"),
	}

	for _, rel := range relPaths {
		gistName := gistFilename(rel)
		roundTripped := gistFilenameToRelPath(gistName)
		if roundTripped != rel {
			t.Errorf("round-trip failed for %q: gistFilename=%q, relPath=%q", rel, gistName, roundTripped)
		}
	}
}


