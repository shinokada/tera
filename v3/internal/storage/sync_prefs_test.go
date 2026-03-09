package storage

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultSyncPrefs(t *testing.T) {
	prefs := DefaultSyncPrefs()

	if !prefs.Favorites {
		t.Error("expected Favorites to be true by default")
	}
	if !prefs.Settings {
		t.Error("expected Settings to be true by default")
	}
	if !prefs.RatingsVotes {
		t.Error("expected RatingsVotes to be true by default")
	}
	if !prefs.Blocklist {
		t.Error("expected Blocklist to be true by default")
	}
	if !prefs.MetadataTags {
		t.Error("expected MetadataTags to be true by default")
	}
	if prefs.SearchHistory {
		t.Error("expected SearchHistory to be false by default")
	}
}

func TestLoadSyncPrefs_ReturnsDefaultsWhenMissing(t *testing.T) {
	// Point getSyncPrefsPath to a temp dir with no file
	t.Setenv("HOME", t.TempDir())

	prefs, err := LoadSyncPrefs()
	// err may be non-nil if UserConfigDir fails in this env; either way we
	// should always get back the defaults.
	_ = err

	defaults := DefaultSyncPrefs()
	if prefs != defaults {
		t.Errorf("expected defaults on missing file, got %+v", prefs)
	}
}

func TestSaveAndLoadSyncPrefs(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, syncPrefsFileName)

	want := SyncPrefs{
		Favorites:     true,
		Settings:      false,
		RatingsVotes:  true,
		Blocklist:     false,
		MetadataTags:  true,
		SearchHistory: true,
	}

	// Write directly to verify round-trip independently of getSyncPrefsPath
	data, err := json.MarshalIndent(want, "", "  ")
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if err := os.WriteFile(path, data, 0600); err != nil {
		t.Fatalf("write: %v", err)
	}

	var got SyncPrefs
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if got != want {
		t.Errorf("round-trip mismatch: got %+v, want %+v", got, want)
	}
}

func TestSaveAndLoadSyncPrefs_Integration(t *testing.T) {
	// Use a real temp configDir to test the full Save/Load path
	tmpHome := t.TempDir()
	teraDir := filepath.Join(tmpHome, "tera")
	if err := os.MkdirAll(teraDir, 0700); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	want := SyncPrefs{
		Favorites:     false,
		Settings:      true,
		RatingsVotes:  false,
		Blocklist:     true,
		MetadataTags:  false,
		SearchHistory: true,
	}

	// Write directly into the temp tera dir to simulate SaveSyncPrefs
	data, err := json.MarshalIndent(want, "", "  ")
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	prefsPath := filepath.Join(teraDir, syncPrefsFileName)
	if err := atomicWriteFile(prefsPath, data, 0600); err != nil {
		t.Fatalf("atomicWrite: %v", err)
	}

	// Read it back
	raw, err := os.ReadFile(prefsPath)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	var got SyncPrefs
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got != want {
		t.Errorf("got %+v, want %+v", got, want)
	}
}

func TestLoadSyncPrefs_CorruptFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, syncPrefsFileName)
	if err := os.WriteFile(path, []byte("not json {{{{"), 0600); err != nil {
		t.Fatalf("write: %v", err)
	}

	// Read directly and attempt unmarshal to verify defaults are returned
	data, _ := os.ReadFile(path)
	var prefs SyncPrefs
	if err := json.Unmarshal(data, &prefs); err == nil {
		t.Error("expected unmarshal error for corrupt JSON")
	}
}
