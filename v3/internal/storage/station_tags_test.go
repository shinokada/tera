package storage

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

// newTestTagsManager creates a TagsManager backed by a temp directory.
func newTestTagsManager(t *testing.T) (*TagsManager, string) {
	t.Helper()
	dir := t.TempDir()
	tm, err := NewTagsManager(dir)
	if err != nil {
		t.Fatalf("NewTagsManager: %v", err)
	}
	return tm, dir
}

// ---------------------------------------------------------------------------
// Tag operations
// ---------------------------------------------------------------------------

func TestAddTag(t *testing.T) {
	tm, _ := newTestTagsManager(t)
	if err := tm.AddTag("uuid-1", "chill vibes"); err != nil {
		t.Fatalf("AddTag: %v", err)
	}
	tags := tm.GetTags("uuid-1")
	if len(tags) != 1 || tags[0] != "chill vibes" {
		t.Errorf("expected [chill vibes], got %v", tags)
	}
}

func TestAddTagNormalization(t *testing.T) {
	tm, _ := newTestTagsManager(t)
	// Should normalize to lowercase and trimmed.
	if err := tm.AddTag("uuid-1", "  Gym Workout  "); err != nil {
		t.Fatalf("AddTag: %v", err)
	}
	tags := tm.GetTags("uuid-1")
	if len(tags) != 1 || tags[0] != "gym workout" {
		t.Errorf("expected [gym workout], got %v", tags)
	}
}

func TestAddTagValidation(t *testing.T) {
	tm, _ := newTestTagsManager(t)

	tests := []struct {
		tag     string
		wantErr bool
	}{
		{"", true},
		{"valid", false},
		{"valid tag", false},
		{"a", false},
		{"a b", false},
		// Exceeding 50 characters
		{"this-tag-is-way-too-long-and-should-be-rejected-because-it-exceeds-50", true},
		// Invalid characters
		{"invalid@tag", true},
		{"has/slash", true},
	}

	for _, tc := range tests {
		err := tm.AddTag("uuid-test", tc.tag)
		if tc.wantErr && err == nil {
			t.Errorf("AddTag(%q): expected error, got nil", tc.tag)
		}
		if !tc.wantErr && err != nil {
			t.Errorf("AddTag(%q): unexpected error: %v", tc.tag, err)
		}
	}
}

func TestAddTagIdempotent(t *testing.T) {
	tm, _ := newTestTagsManager(t)
	for i := 0; i < 3; i++ {
		if err := tm.AddTag("uuid-1", "chill"); err != nil {
			t.Fatalf("AddTag iteration %d: %v", i, err)
		}
	}
	if len(tm.GetTags("uuid-1")) != 1 {
		t.Errorf("expected 1 tag (deduplicated), got %d", len(tm.GetTags("uuid-1")))
	}
}

func TestRemoveTag(t *testing.T) {
	tm, _ := newTestTagsManager(t)
	_ = tm.AddTag("uuid-1", "chill")
	_ = tm.AddTag("uuid-1", "focus")
	if err := tm.RemoveTag("uuid-1", "chill"); err != nil {
		t.Fatalf("RemoveTag: %v", err)
	}
	tags := tm.GetTags("uuid-1")
	if len(tags) != 1 || tags[0] != "focus" {
		t.Errorf("expected [focus], got %v", tags)
	}
}

func TestSetTags(t *testing.T) {
	tm, _ := newTestTagsManager(t)
	_ = tm.AddTag("uuid-1", "old")
	if err := tm.SetTags("uuid-1", []string{"new1", "new2"}); err != nil {
		t.Fatalf("SetTags: %v", err)
	}
	tags := tm.GetTags("uuid-1")
	if len(tags) != 2 {
		t.Errorf("expected 2 tags, got %v", tags)
	}
}

func TestClearTags(t *testing.T) {
	tm, _ := newTestTagsManager(t)
	_ = tm.AddTag("uuid-1", "work")
	if err := tm.ClearTags("uuid-1"); err != nil {
		t.Fatalf("ClearTags: %v", err)
	}
	if tags := tm.GetTags("uuid-1"); len(tags) != 0 {
		t.Errorf("expected empty tags after clear, got %v", tags)
	}
}

func TestMaxTagsLimit(t *testing.T) {
	tm, _ := newTestTagsManager(t)
	for i := 0; i < 20; i++ {
		tag := string(rune('a'+i)) + "tag"
		if err := tm.AddTag("uuid-1", tag); err != nil {
			t.Fatalf("AddTag %d: %v", i, err)
		}
	}
	// 21st tag should fail.
	if err := tm.AddTag("uuid-1", "overflow"); err == nil {
		t.Error("expected error for 21st tag, got nil")
	}
}

// ---------------------------------------------------------------------------
// AllTags index
// ---------------------------------------------------------------------------

func TestAllTagsIndex(t *testing.T) {
	tm, _ := newTestTagsManager(t)
	_ = tm.AddTag("uuid-1", "chill")
	_ = tm.AddTag("uuid-2", "focus")
	_ = tm.AddTag("uuid-3", "chill") // duplicate across stations
	all := tm.GetAllTags()
	if len(all) != 2 {
		t.Errorf("expected 2 unique tags, got %v", all)
	}
}

// ---------------------------------------------------------------------------
// Query operations
// ---------------------------------------------------------------------------

func TestGetStationsByTag(t *testing.T) {
	tm, _ := newTestTagsManager(t)
	_ = tm.AddTag("uuid-1", "chill")
	_ = tm.AddTag("uuid-2", "focus")
	_ = tm.AddTag("uuid-3", "chill")
	uuids := tm.GetStationsByTag("chill")
	if len(uuids) != 2 {
		t.Errorf("expected 2 stations for 'chill', got %d", len(uuids))
	}
}

func TestGetStationsByTagsMatchAny(t *testing.T) {
	tm, _ := newTestTagsManager(t)
	_ = tm.AddTag("uuid-1", "chill")
	_ = tm.AddTag("uuid-2", "focus")
	_ = tm.AddTag("uuid-3", "workout")
	uuids := tm.GetStationsByTags([]string{"chill", "focus"}, false)
	if len(uuids) != 2 {
		t.Errorf("expected 2 stations (any), got %d", len(uuids))
	}
}

func TestGetStationsByTagsMatchAll(t *testing.T) {
	tm, _ := newTestTagsManager(t)
	_ = tm.AddTag("uuid-1", "chill")
	_ = tm.AddTag("uuid-1", "focus")
	_ = tm.AddTag("uuid-2", "chill")
	uuids := tm.GetStationsByTags([]string{"chill", "focus"}, true)
	if len(uuids) != 1 || uuids[0] != "uuid-1" {
		t.Errorf("expected [uuid-1] for matchAll, got %v", uuids)
	}
}

// ---------------------------------------------------------------------------
// Tag playlist operations
// ---------------------------------------------------------------------------

func TestCreatePlaylist(t *testing.T) {
	tm, _ := newTestTagsManager(t)
	_ = tm.AddTag("uuid-1", "chill")
	if err := tm.CreatePlaylist("Evening Chill", []string{"chill"}, "any"); err != nil {
		t.Fatalf("CreatePlaylist: %v", err)
	}
	p := tm.GetPlaylist("Evening Chill")
	if p == nil {
		t.Fatal("playlist not found after creation")
	}
	if p.MatchMode != "any" {
		t.Errorf("expected matchMode 'any', got %q", p.MatchMode)
	}
}

func TestCreatePlaylistDuplicate(t *testing.T) {
	tm, _ := newTestTagsManager(t)
	_ = tm.AddTag("uuid-1", "chill")
	_ = tm.CreatePlaylist("Dup", []string{"chill"}, "any")
	if err := tm.CreatePlaylist("Dup", []string{"chill"}, "any"); err == nil {
		t.Error("expected error creating duplicate playlist, got nil")
	}
}

func TestCreatePlaylistValidation(t *testing.T) {
	tm, _ := newTestTagsManager(t)
	tests := []struct {
		name      string
		tags      []string
		matchMode string
		wantErr   bool
	}{
		{"", []string{"chill"}, "any", true},          // empty name
		{"p1", []string{}, "any", true},               // no tags
		{"p2", []string{"chill"}, "invalid", true},    // bad matchMode
		{"p3", []string{"chill"}, "any", false},       // valid
		{"p4", []string{"chill"}, "all", false},       // valid matchAll
	}
	for _, tc := range tests {
		err := tm.CreatePlaylist(tc.name, tc.tags, tc.matchMode)
		if tc.wantErr && err == nil {
			t.Errorf("CreatePlaylist(%q): expected error, got nil", tc.name)
		}
		if !tc.wantErr && err != nil {
			t.Errorf("CreatePlaylist(%q): unexpected error: %v", tc.name, err)
		}
	}
}

func TestDeletePlaylist(t *testing.T) {
	tm, _ := newTestTagsManager(t)
	_ = tm.AddTag("uuid-1", "chill")
	_ = tm.CreatePlaylist("MyList", []string{"chill"}, "any")
	if err := tm.DeletePlaylist("MyList"); err != nil {
		t.Fatalf("DeletePlaylist: %v", err)
	}
	if p := tm.GetPlaylist("MyList"); p != nil {
		t.Error("playlist should be nil after deletion")
	}
}

func TestDeletePlaylistNotFound(t *testing.T) {
	tm, _ := newTestTagsManager(t)
	if err := tm.DeletePlaylist("nonexistent"); err == nil {
		t.Error("expected error deleting nonexistent playlist")
	}
}

func TestGetPlaylistStations(t *testing.T) {
	tm, _ := newTestTagsManager(t)
	_ = tm.AddTag("uuid-1", "chill")
	_ = tm.AddTag("uuid-2", "chill")
	_ = tm.AddTag("uuid-3", "focus")
	_ = tm.CreatePlaylist("ChillOnly", []string{"chill"}, "any")
	uuids := tm.GetPlaylistStations("ChillOnly")
	if len(uuids) != 2 {
		t.Errorf("expected 2 stations in playlist, got %d", len(uuids))
	}
}

func TestGetAllPlaylists(t *testing.T) {
	tm, _ := newTestTagsManager(t)
	_ = tm.AddTag("uuid-1", "chill")
	_ = tm.CreatePlaylist("A", []string{"chill"}, "any")
	_ = tm.CreatePlaylist("B", []string{"chill"}, "all")
	all := tm.GetAllPlaylists()
	if len(all) != 2 {
		t.Errorf("expected 2 playlists, got %d", len(all))
	}
}

// ---------------------------------------------------------------------------
// Persistence
// ---------------------------------------------------------------------------

func TestSaveAndLoad(t *testing.T) {
	dir := t.TempDir()
	tm1, err := NewTagsManager(dir)
	if err != nil {
		t.Fatalf("NewTagsManager: %v", err)
	}
	_ = tm1.AddTag("uuid-1", "chill")
	_ = tm1.CreatePlaylist("MyList", []string{"chill"}, "any")
	if err := tm1.Save(); err != nil {
		t.Fatalf("Save: %v", err)
	}

	tm2, err := NewTagsManager(dir)
	if err != nil {
		t.Fatalf("NewTagsManager (reload): %v", err)
	}
	tags := tm2.GetTags("uuid-1")
	if len(tags) != 1 || tags[0] != "chill" {
		t.Errorf("tags not persisted: %v", tags)
	}
	if p := tm2.GetPlaylist("MyList"); p == nil {
		t.Error("playlist not persisted")
	}
}

func TestCorruptedFile(t *testing.T) {
	dir := t.TempDir()
	// Write garbage JSON.
	if err := os.WriteFile(filepath.Join(dir, "station_tags.json"), []byte("{invalid json}"), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	// Should not return error — should start fresh.
	tm, err := NewTagsManager(dir)
	if err != nil {
		t.Fatalf("NewTagsManager with corrupt file: %v", err)
	}
	if tags := tm.GetTags("any-uuid"); len(tags) != 0 {
		t.Errorf("expected empty tags after corrupt file, got %v", tags)
	}
}

// ---------------------------------------------------------------------------
// Concurrency
// ---------------------------------------------------------------------------

func TestConcurrentAccess(t *testing.T) {
	tm, _ := newTestTagsManager(t)
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			uuid := "uuid-concurrent"
			tag := "tag"
			_ = tm.AddTag(uuid, tag)
			_ = tm.GetTags(uuid)
		}(i)
	}
	wg.Wait()
	// Should have exactly 1 tag (idempotent adds).
	if tags := tm.GetTags("uuid-concurrent"); len(tags) != 1 {
		t.Errorf("expected 1 tag after concurrent adds, got %d", len(tags))
	}
}

// ---------------------------------------------------------------------------
// Debounced save
// ---------------------------------------------------------------------------

func TestDebouncedSave(t *testing.T) {
	dir := t.TempDir()
	tm, err := NewTagsManager(dir)
	if err != nil {
		t.Fatalf("NewTagsManager: %v", err)
	}
	_ = tm.AddTag("uuid-1", "chill")
	// savePending should be set.
	if !tm.savePending.Load() {
		t.Error("savePending should be true after AddTag")
	}
	// Manually trigger save.
	if err := tm.Save(); err != nil {
		t.Fatalf("Save: %v", err)
	}
	// File should exist.
	if _, err := os.Stat(filepath.Join(dir, "station_tags.json")); err != nil {
		t.Errorf("station_tags.json not found after save: %v", err)
	}
}

// ---------------------------------------------------------------------------
// Edge cases
// ---------------------------------------------------------------------------

func TestGetTagsUnknownUUID(t *testing.T) {
	tm, _ := newTestTagsManager(t)
	tags := tm.GetTags("nonexistent")
	if tags == nil || len(tags) != 0 {
		t.Errorf("expected empty slice for unknown UUID, got %v", tags)
	}
}

func TestRemoveTagNonexistentStation(t *testing.T) {
	tm, _ := newTestTagsManager(t)
	// Should not error.
	if err := tm.RemoveTag("ghost-uuid", "some-tag"); err != nil {
		t.Errorf("expected nil error for unknown station, got %v", err)
	}
}

func TestSetTagsDeduplication(t *testing.T) {
	tm, _ := newTestTagsManager(t)
	if err := tm.SetTags("uuid-1", []string{"chill", "chill", "focus"}); err != nil {
		t.Fatalf("SetTags: %v", err)
	}
	tags := tm.GetTags("uuid-1")
	if len(tags) != 2 {
		t.Errorf("expected 2 unique tags after dedup, got %v", tags)
	}
}

func TestMaxTagLengthLimit(t *testing.T) {
	tm, _ := newTestTagsManager(t)
	longTag := "a" + string(make([]byte, 50)) // 51+ chars
	if err := tm.AddTag("uuid-1", longTag); err == nil {
		t.Error("expected error for tag exceeding 50 characters")
	}
}

func TestPlaylistCreatedAt(t *testing.T) {
	tm, _ := newTestTagsManager(t)
	_ = tm.AddTag("uuid-1", "chill")
	before := time.Now()
	_ = tm.CreatePlaylist("TimeTest", []string{"chill"}, "any")
	p := tm.GetPlaylist("TimeTest")
	if p == nil {
		t.Fatal("playlist not found")
	}
	if p.CreatedAt.Before(before.Add(-time.Second)) {
		t.Errorf("CreatedAt %v is suspiciously old", p.CreatedAt)
	}
}
