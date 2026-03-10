package storage

import (
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
	configDir, err := os.UserConfigDir()
	if err != nil {
		t.Skipf("cannot determine config dir: %v", err)
	}
	teraDir := filepath.Join(configDir, "tera")
	if err := os.MkdirAll(teraDir, 0700); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	path := filepath.Join(teraDir, syncPrefsFileName)

	// Back up any existing file and restore it when the test ends.
	original, readErr := os.ReadFile(path)
	t.Cleanup(func() {
		if readErr == nil {
			_ = os.WriteFile(path, original, 0600)
		} else {
			_ = os.Remove(path)
		}
	})

	want := SyncPrefs{
		Favorites:     true,
		Settings:      false,
		RatingsVotes:  true,
		Blocklist:     false,
		MetadataTags:  true,
		SearchHistory: true,
	}

	if err := SaveSyncPrefs(want); err != nil {
		t.Fatalf("SaveSyncPrefs: %v", err)
	}

	got, err := LoadSyncPrefs()
	if err != nil {
		t.Fatalf("LoadSyncPrefs: %v", err)
	}
	if got != want {
		t.Errorf("round-trip mismatch: got %+v, want %+v", got, want)
	}
}

func TestSaveAndLoadSyncPrefs_Integration(t *testing.T) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		t.Skipf("cannot determine config dir: %v", err)
	}
	teraDir := filepath.Join(configDir, "tera")
	if err := os.MkdirAll(teraDir, 0700); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	path := filepath.Join(teraDir, syncPrefsFileName)

	// Back up any existing file and restore it when the test ends.
	original, readErr := os.ReadFile(path)
	t.Cleanup(func() {
		if readErr == nil {
			_ = os.WriteFile(path, original, 0600)
		} else {
			_ = os.Remove(path)
		}
	})

	want := SyncPrefs{
		Favorites:     false,
		Settings:      true,
		RatingsVotes:  false,
		Blocklist:     true,
		MetadataTags:  false,
		SearchHistory: true,
	}

	if err := SaveSyncPrefs(want); err != nil {
		t.Fatalf("SaveSyncPrefs: %v", err)
	}

	got, err := LoadSyncPrefs()
	if err != nil {
		t.Fatalf("LoadSyncPrefs: %v", err)
	}
	if got != want {
		t.Errorf("got %+v, want %+v", got, want)
	}
}

func TestLoadSyncPrefs_CorruptFile(t *testing.T) {
	// Resolve the real config dir so we write the corrupt file where
	// LoadSyncPrefs will look for it (os.UserConfigDir() is platform-specific).
	configDir, err := os.UserConfigDir()
	if err != nil {
		t.Skipf("cannot determine config dir: %v", err)
	}
	teraDir := filepath.Join(configDir, "tera")
	if err := os.MkdirAll(teraDir, 0700); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	path := filepath.Join(teraDir, syncPrefsFileName)

	// Back up any existing file and restore it when the test ends.
	original, readErr := os.ReadFile(path)
	t.Cleanup(func() {
		if readErr == nil {
			_ = os.WriteFile(path, original, 0600)
		} else {
			_ = os.Remove(path)
		}
	})

	if err := os.WriteFile(path, []byte("not json {{{{"), 0600); err != nil {
		t.Fatalf("write: %v", err)
	}

	prefs, err := LoadSyncPrefs()
	if err == nil {
		t.Error("expected error for corrupt JSON")
	}
	if prefs != DefaultSyncPrefs() {
		t.Errorf("expected defaults on corrupt file, got %+v", prefs)
	}
}
