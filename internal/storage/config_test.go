package storage

import (
	"path/filepath"
	"testing"

	"github.com/shinokada/tera/v3/internal/config"
)

// redirectConfigHome redirects all config-directory env vars to a fresh temp
// directory so tests never read from or write to the real user config.
func redirectConfigHome(t *testing.T) {
	t.Helper()
	root := t.TempDir()
	t.Setenv("HOME", root)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(root, ".config"))
	t.Setenv("APPDATA", filepath.Join(root, "AppData", "Roaming"))
}

func TestLoadPlayOptionsConfigFromUnified(t *testing.T) {
	// Redirect all config-dir env vars so the test never touches the real config.
	// config.Load() will auto-create a default config under the redirected dir.
	redirectConfigHome(t)

	po, err := LoadPlayOptionsConfigFromUnified()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	defaults := config.DefaultPlayOptionsConfig()
	if po.ContinueOnNavigate != defaults.ContinueOnNavigate {
		t.Errorf("ContinueOnNavigate: got %v, want %v", po.ContinueOnNavigate, defaults.ContinueOnNavigate)
	}
	if po.DefaultVolume != defaults.DefaultVolume {
		t.Errorf("DefaultVolume: got %d, want %d", po.DefaultVolume, defaults.DefaultVolume)
	}
	if po.ConfirmStop != defaults.ConfirmStop {
		t.Errorf("ConfirmStop: got %v, want %v", po.ConfirmStop, defaults.ConfirmStop)
	}
	if po.ShowMetadata != defaults.ShowMetadata {
		t.Errorf("ShowMetadata: got %v, want %v", po.ShowMetadata, defaults.ShowMetadata)
	}
	if po.StartVolumeMode != defaults.StartVolumeMode {
		t.Errorf("StartVolumeMode: got %q, want %q", po.StartVolumeMode, defaults.StartVolumeMode)
	}
	if po.LastUsedVolume != defaults.LastUsedVolume {
		t.Errorf("LastUsedVolume: got %d, want %d", po.LastUsedVolume, defaults.LastUsedVolume)
	}
}

func TestSavePlayOptionsConfigToUnified(t *testing.T) {
	// Redirect all config-dir env vars so the test never touches the real config.
	redirectConfigHome(t)

	custom := config.PlayOptionsConfig{
		ContinueOnNavigate: true,
		DefaultVolume:      70,
		ConfirmStop:        true,
		ShowMetadata:       false,
		StartVolumeMode:    "last_used",
		LastUsedVolume:     65,
	}

	if err := SavePlayOptionsConfigToUnified(custom); err != nil {
		t.Fatalf("SavePlayOptionsConfigToUnified failed: %v", err)
	}

	got, err := LoadPlayOptionsConfigFromUnified()
	if err != nil {
		t.Fatalf("LoadPlayOptionsConfigFromUnified failed: %v", err)
	}

	if got.ContinueOnNavigate != custom.ContinueOnNavigate {
		t.Errorf("ContinueOnNavigate: got %v, want %v", got.ContinueOnNavigate, custom.ContinueOnNavigate)
	}
	if got.DefaultVolume != custom.DefaultVolume {
		t.Errorf("DefaultVolume: got %d, want %d", got.DefaultVolume, custom.DefaultVolume)
	}
	if got.ConfirmStop != custom.ConfirmStop {
		t.Errorf("ConfirmStop: got %v, want %v", got.ConfirmStop, custom.ConfirmStop)
	}
	if got.ShowMetadata != custom.ShowMetadata {
		t.Errorf("ShowMetadata: got %v, want %v", got.ShowMetadata, custom.ShowMetadata)
	}
	if got.StartVolumeMode != custom.StartVolumeMode {
		t.Errorf("StartVolumeMode: got %q, want %q", got.StartVolumeMode, custom.StartVolumeMode)
	}
	if got.LastUsedVolume != custom.LastUsedVolume {
		t.Errorf("LastUsedVolume: got %d, want %d", got.LastUsedVolume, custom.LastUsedVolume)
	}
}
