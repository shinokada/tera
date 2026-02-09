package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestHasV2Config(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "tera-migrate-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			t.Logf("failed to clean up temp dir: %v", err)
		}
	}()

	// No v2 config initially
	if HasV2Config(tmpDir) {
		t.Error("should not detect v2 config in empty directory")
	}

	// Create a v2 config file
	themeContent := `colors:
  primary: "6"
  secondary: "12"
  highlight: "3"
  error: "9"
  success: "2"
  muted: "8"
  text: "7"
padding:
  page_horizontal: 2
  page_vertical: 1
  list_item_left: 2
  box_horizontal: 2
  box_vertical: 1
`
	themePath := filepath.Join(tmpDir, "theme.yaml")
	if err := os.WriteFile(themePath, []byte(themeContent), 0644); err != nil {
		t.Fatalf("failed to create theme.yaml: %v", err)
	}

	// Should now detect v2 config
	if !HasV2Config(tmpDir) {
		t.Error("should detect v2 config with theme.yaml present")
	}
}

func TestDetectV2Config(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "tera-migrate-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			t.Logf("failed to clean up temp dir: %v", err)
		}
	}()

	// Create some v2 config files
	files := map[string]string{
		"theme.yaml": `colors:
  primary: "6"
padding:
  page_horizontal: 2
`,
		"shuffle.yaml": `shuffle:
  auto_advance: false
  interval_minutes: 5
`,
	}

	for filename, content := range files {
		path := filepath.Join(tmpDir, filename)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("failed to create %s: %v", filename, err)
		}
	}

	// Detect v2 config
	detected := DetectV2Config(tmpDir)

	// Check detection results
	if !detected["theme.yaml"] {
		t.Error("should detect theme.yaml")
	}
	if !detected["shuffle.yaml"] {
		t.Error("should detect shuffle.yaml")
	}
	if detected["appearance_config.yaml"] {
		t.Error("should not detect missing appearance_config.yaml")
	}
	if detected["connection_config.yaml"] {
		t.Error("should not detect missing connection_config.yaml")
	}
}

func TestMigrateFromV2(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "tera-migrate-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			t.Logf("failed to clean up temp dir: %v", err)
		}
	}()

	// Create v2 config files
	v2Configs := map[string]string{
		"theme.yaml": `colors:
  primary: "14"
  secondary: "11"
  highlight: "3"
  error: "9"
  success: "10"
  muted: "8"
  text: "7"
padding:
  page_horizontal: 3
  page_vertical: 2
  list_item_left: 3
  box_horizontal: 3
  box_vertical: 2
`,
		"appearance_config.yaml": `header:
  mode: text
  custom_text: "My Radio App"
  ascii_art: ""
  alignment: left
  width: 60
  color: "12"
  bold: false
  padding_top: 2
  padding_bottom: 1
`,
		"connection_config.yaml": `auto_reconnect: false
reconnect_delay: 10
stream_buffer_mb: 100
`,
		"shuffle.yaml": `shuffle:
  auto_advance: true
  interval_minutes: 10
  remember_history: false
  max_history: 7
`,
	}

	for filename, content := range v2Configs {
		path := filepath.Join(tmpDir, filename)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("failed to create %s: %v", filename, err)
		}
	}

	// Migrate v2 config
	cfg, err := MigrateFromV2(tmpDir)
	if err != nil {
		t.Fatalf("migration failed: %v", err)
	}

	// Verify theme colors migrated
	if cfg.UI.Theme.Colors["primary"] != "14" {
		t.Errorf("expected primary color 14, got %s", cfg.UI.Theme.Colors["primary"])
	}
	if cfg.UI.Theme.Colors["secondary"] != "11" {
		t.Errorf("expected secondary color 11, got %s", cfg.UI.Theme.Colors["secondary"])
	}

	// Verify theme padding migrated
	if cfg.UI.Theme.Padding.PageHorizontal != 3 {
		t.Errorf("expected page_horizontal 3, got %d", cfg.UI.Theme.Padding.PageHorizontal)
	}
	if cfg.UI.Theme.Padding.PageVertical != 2 {
		t.Errorf("expected page_vertical 2, got %d", cfg.UI.Theme.Padding.PageVertical)
	}

	// Verify appearance migrated
	if cfg.UI.Appearance.HeaderMode != "text" {
		t.Errorf("expected header mode 'text', got %s", cfg.UI.Appearance.HeaderMode)
	}
	if cfg.UI.Appearance.CustomText != "My Radio App" {
		t.Errorf("expected custom text 'My Radio App', got %s", cfg.UI.Appearance.CustomText)
	}
	if cfg.UI.Appearance.HeaderAlign != "left" {
		t.Errorf("expected header align 'left', got %s", cfg.UI.Appearance.HeaderAlign)
	}
	if cfg.UI.Appearance.HeaderWidth != 60 {
		t.Errorf("expected header width 60, got %d", cfg.UI.Appearance.HeaderWidth)
	}
	if cfg.UI.Appearance.HeaderColor != "12" {
		t.Errorf("expected header color '12', got %s", cfg.UI.Appearance.HeaderColor)
	}
	if cfg.UI.Appearance.HeaderBold {
		t.Error("expected header bold to be false")
	}
	if cfg.UI.Appearance.PaddingTop != 2 {
		t.Errorf("expected padding top 2, got %d", cfg.UI.Appearance.PaddingTop)
	}
	if cfg.UI.Appearance.PaddingBottom != 1 {
		t.Errorf("expected padding bottom 1, got %d", cfg.UI.Appearance.PaddingBottom)
	}

	// Verify network config migrated
	if cfg.Network.AutoReconnect {
		t.Error("expected auto_reconnect to be false")
	}
	if cfg.Network.ReconnectDelay != 10 {
		t.Errorf("expected reconnect delay 10, got %d", cfg.Network.ReconnectDelay)
	}
	if cfg.Network.BufferSizeMB != 100 {
		t.Errorf("expected buffer size 100, got %d", cfg.Network.BufferSizeMB)
	}

	// Verify shuffle config migrated
	if !cfg.Shuffle.AutoAdvance {
		t.Error("expected auto_advance to be true")
	}
	if cfg.Shuffle.IntervalMinutes != 10 {
		t.Errorf("expected interval 10, got %d", cfg.Shuffle.IntervalMinutes)
	}
	if cfg.Shuffle.RememberHistory {
		t.Error("expected remember_history to be false")
	}
	if cfg.Shuffle.MaxHistory != 7 {
		t.Errorf("expected max history 7, got %d", cfg.Shuffle.MaxHistory)
	}
}

func TestMigrateFromV2_PartialConfigs(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "tera-migrate-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			t.Logf("failed to clean up temp dir: %v", err)
		}
	}()

	// Create only theme.yaml (partial v2 config)
	themeContent := `colors:
  primary: "14"
  secondary: "11"
  highlight: "3"
  error: "9"
  success: "10"
  muted: "8"
  text: "7"
padding:
  page_horizontal: 3
  page_vertical: 2
  list_item_left: 3
  box_horizontal: 3
  box_vertical: 2
`
	themePath := filepath.Join(tmpDir, "theme.yaml")
	if err := os.WriteFile(themePath, []byte(themeContent), 0644); err != nil {
		t.Fatalf("failed to create theme.yaml: %v", err)
	}

	// Migrate should work with partial configs, using defaults for missing files
	cfg, err := MigrateFromV2(tmpDir)
	if err != nil {
		t.Fatalf("migration failed: %v", err)
	}

	// Theme should be migrated
	if cfg.UI.Theme.Colors["primary"] != "14" {
		t.Errorf("expected primary color 14, got %s", cfg.UI.Theme.Colors["primary"])
	}

	// Other configs should use defaults
	if cfg.Network.AutoReconnect != true {
		t.Error("expected default auto_reconnect (true)")
	}
	if cfg.Shuffle.IntervalMinutes != 5 {
		t.Errorf("expected default interval (5), got %d", cfg.Shuffle.IntervalMinutes)
	}
}

func TestBackupV2Configs(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "tera-migrate-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			t.Logf("failed to clean up temp dir: %v", err)
		}
	}()

	// Create v2 config files
	v2Files := []string{"theme.yaml", "shuffle.yaml"}
	for _, filename := range v2Files {
		path := filepath.Join(tmpDir, filename)
		content := "test: value\n"
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("failed to create %s: %v", filename, err)
		}
	}

	// Backup v2 configs
	if err := BackupV2Configs(tmpDir); err != nil {
		t.Fatalf("backup failed: %v", err)
	}

	// Find backup directory
	entries, err := os.ReadDir(tmpDir)
	if err != nil {
		t.Fatalf("failed to read dir: %v", err)
	}

	backupDirFound := false
	for _, entry := range entries {
		if entry.IsDir() && len(entry.Name()) > 10 && entry.Name()[:11] == ".v2-backup-" {
			backupDirFound = true

			// Verify backup files exist
			backupDir := filepath.Join(tmpDir, entry.Name())
			for _, filename := range v2Files {
				backupPath := filepath.Join(backupDir, filename)
				if _, err := os.Stat(backupPath); os.IsNotExist(err) {
					t.Errorf("backup file %s should exist", filename)
				}
			}
		}
	}

	if !backupDirFound {
		t.Error("backup directory should be created")
	}
}

func TestRemoveV2Configs(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "tera-migrate-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			t.Logf("failed to clean up temp dir: %v", err)
		}
	}()

	// Create v2 config files
	v2Files := []string{"theme.yaml", "appearance_config.yaml", "connection_config.yaml", "shuffle.yaml"}
	for _, filename := range v2Files {
		path := filepath.Join(tmpDir, filename)
		content := "test: value\n"
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("failed to create %s: %v", filename, err)
		}
	}

	// Remove v2 configs
	if err := RemoveV2Configs(tmpDir); err != nil {
		t.Fatalf("remove failed: %v", err)
	}

	// Verify files are removed
	for _, filename := range v2Files {
		path := filepath.Join(tmpDir, filename)
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Errorf("file %s should be removed", filename)
		}
	}
}

func TestReadV2Theme(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "tera-migrate-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			t.Logf("failed to clean up temp dir: %v", err)
		}
	}()

	// Create v2 theme.yaml
	themeContent := `colors:
  primary: "14"
  secondary: "11"
  highlight: "3"
  error: "9"
  success: "10"
  muted: "8"
  text: "7"
padding:
  page_horizontal: 3
  page_vertical: 2
  list_item_left: 3
  box_horizontal: 3
  box_vertical: 2
`
	themePath := filepath.Join(tmpDir, "theme.yaml")
	if err := os.WriteFile(themePath, []byte(themeContent), 0644); err != nil {
		t.Fatalf("failed to create theme.yaml: %v", err)
	}

	// Read theme colors
	colors, err := readV2Theme(themePath)
	if err != nil {
		t.Fatalf("failed to read theme: %v", err)
	}

	if colors["primary"] != "14" {
		t.Errorf("expected primary 14, got %s", colors["primary"])
	}
	if colors["secondary"] != "11" {
		t.Errorf("expected secondary 11, got %s", colors["secondary"])
	}

	// Read theme padding
	padding, err := readV2ThemePadding(themePath)
	if err != nil {
		t.Fatalf("failed to read padding: %v", err)
	}

	if padding.PageHorizontal != 3 {
		t.Errorf("expected page_horizontal 3, got %d", padding.PageHorizontal)
	}
	if padding.PageVertical != 2 {
		t.Errorf("expected page_vertical 2, got %d", padding.PageVertical)
	}
}

func TestReadV2Appearance(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "tera-migrate-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			t.Logf("failed to clean up temp dir: %v", err)
		}
	}()

	// Create v2 appearance_config.yaml
	appearanceContent := `header:
  mode: ascii
  custom_text: ""
  ascii_art: |
    ╔═══════╗
    ║ RADIO ║
    ╚═══════╝
  alignment: center
  width: 70
  color: "14"
  bold: true
  padding_top: 2
  padding_bottom: 1
`
	appearancePath := filepath.Join(tmpDir, "appearance_config.yaml")
	if err := os.WriteFile(appearancePath, []byte(appearanceContent), 0644); err != nil {
		t.Fatalf("failed to create appearance_config.yaml: %v", err)
	}

	// Read appearance config
	appearance, err := readV2Appearance(appearancePath)
	if err != nil {
		t.Fatalf("failed to read appearance: %v", err)
	}

	if appearance.HeaderMode != "ascii" {
		t.Errorf("expected mode ascii, got %s", appearance.HeaderMode)
	}
	if appearance.HeaderAlign != "center" {
		t.Errorf("expected align center, got %s", appearance.HeaderAlign)
	}
	if appearance.HeaderWidth != 70 {
		t.Errorf("expected width 70, got %d", appearance.HeaderWidth)
	}
	if appearance.HeaderColor != "14" {
		t.Errorf("expected color 14, got %s", appearance.HeaderColor)
	}
	if !appearance.HeaderBold {
		t.Error("expected bold to be true")
	}
	if appearance.PaddingTop != 2 {
		t.Errorf("expected padding_top 2, got %d", appearance.PaddingTop)
	}
}
