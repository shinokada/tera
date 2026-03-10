package config

import (
	"os"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.Version != ConfigVersion {
		t.Errorf("expected version %s, got %s", ConfigVersion, cfg.Version)
	}
	if cfg.Player.DefaultVolume != 100 {
		t.Errorf("expected default volume 100, got %d", cfg.Player.DefaultVolume)
	}
	if cfg.Player.BufferSizeMB != 50 {
		t.Errorf("expected buffer size 50, got %d", cfg.Player.BufferSizeMB)
	}
	if !cfg.Network.AutoReconnect {
		t.Error("expected auto_reconnect to be true")
	}
	if cfg.Network.ReconnectDelay != 5 {
		t.Errorf("expected reconnect delay 5, got %d", cfg.Network.ReconnectDelay)
	}
	if cfg.Shuffle.AutoAdvance {
		t.Error("expected auto_advance to be false")
	}
	if cfg.Shuffle.IntervalMinutes != 5 {
		t.Errorf("expected interval 5, got %d", cfg.Shuffle.IntervalMinutes)
	}
}

func TestPlayerConfigValidation(t *testing.T) {
	tests := []struct {
		name     string
		input    PlayerConfig
		expected PlayerConfig
		hasError bool
	}{
		{
			name:     "valid config",
			input:    PlayerConfig{DefaultVolume: 80, BufferSizeMB: 50},
			expected: PlayerConfig{DefaultVolume: 80, BufferSizeMB: 50},
			hasError: false,
		},
		{
			name:     "volume too low",
			input:    PlayerConfig{DefaultVolume: -10, BufferSizeMB: 50},
			expected: PlayerConfig{DefaultVolume: 0, BufferSizeMB: 50},
			hasError: true,
		},
		{
			name:     "volume too high",
			input:    PlayerConfig{DefaultVolume: 150, BufferSizeMB: 50},
			expected: PlayerConfig{DefaultVolume: 100, BufferSizeMB: 50},
			hasError: true,
		},
		{
			name:     "buffer too small",
			input:    PlayerConfig{DefaultVolume: 80, BufferSizeMB: 5},
			expected: PlayerConfig{DefaultVolume: 80, BufferSizeMB: 10},
			hasError: true,
		},
		{
			name:     "buffer too large",
			input:    PlayerConfig{DefaultVolume: 80, BufferSizeMB: 250},
			expected: PlayerConfig{DefaultVolume: 80, BufferSizeMB: 200},
			hasError: true,
		},
		{
			name:     "buffer disabled (0)",
			input:    PlayerConfig{DefaultVolume: 80, BufferSizeMB: 0},
			expected: PlayerConfig{DefaultVolume: 80, BufferSizeMB: 0},
			hasError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.input.Validate()
			if tt.hasError && err == nil {
				t.Error("expected validation error, got nil")
			}
			if !tt.hasError && err != nil {
				t.Errorf("expected no error, got %v", err)
			}
			if tt.input.DefaultVolume != tt.expected.DefaultVolume {
				t.Errorf("expected volume %d, got %d", tt.expected.DefaultVolume, tt.input.DefaultVolume)
			}
			if tt.input.BufferSizeMB != tt.expected.BufferSizeMB {
				t.Errorf("expected buffer %d, got %d", tt.expected.BufferSizeMB, tt.input.BufferSizeMB)
			}
		})
	}
}

func TestNetworkConfigValidation(t *testing.T) {
	tests := []struct {
		name     string
		input    NetworkConfig
		expected NetworkConfig
		hasError bool
	}{
		{
			name:     "valid config",
			input:    NetworkConfig{AutoReconnect: true, ReconnectDelay: 5, BufferSizeMB: 50},
			expected: NetworkConfig{AutoReconnect: true, ReconnectDelay: 5, BufferSizeMB: 50},
			hasError: false,
		},
		{
			name:     "delay too short",
			input:    NetworkConfig{AutoReconnect: true, ReconnectDelay: 0, BufferSizeMB: 50},
			expected: NetworkConfig{AutoReconnect: true, ReconnectDelay: 1, BufferSizeMB: 50},
			hasError: true,
		},
		{
			name:     "delay too long",
			input:    NetworkConfig{AutoReconnect: true, ReconnectDelay: 60, BufferSizeMB: 50},
			expected: NetworkConfig{AutoReconnect: true, ReconnectDelay: 30, BufferSizeMB: 50},
			hasError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.input.Validate()
			if tt.hasError && err == nil {
				t.Error("expected validation error, got nil")
			}
			if !tt.hasError && err != nil {
				t.Errorf("expected no error, got %v", err)
			}
			if tt.input.ReconnectDelay != tt.expected.ReconnectDelay {
				t.Errorf("expected delay %d, got %d", tt.expected.ReconnectDelay, tt.input.ReconnectDelay)
			}
		})
	}
}

func TestShuffleConfigValidation(t *testing.T) {
	tests := []struct {
		name     string
		input    ShuffleConfig
		expected ShuffleConfig
		hasError bool
	}{
		{
			name:     "valid config",
			input:    ShuffleConfig{AutoAdvance: true, IntervalMinutes: 5, RememberHistory: true, MaxHistory: 7},
			expected: ShuffleConfig{AutoAdvance: true, IntervalMinutes: 5, RememberHistory: true, MaxHistory: 7},
			hasError: false,
		},
		{
			name:     "invalid interval",
			input:    ShuffleConfig{AutoAdvance: true, IntervalMinutes: 7, RememberHistory: true, MaxHistory: 5},
			expected: ShuffleConfig{AutoAdvance: true, IntervalMinutes: 5, RememberHistory: true, MaxHistory: 5},
			hasError: true,
		},
		{
			name:     "invalid max history",
			input:    ShuffleConfig{AutoAdvance: true, IntervalMinutes: 5, RememberHistory: true, MaxHistory: 8},
			expected: ShuffleConfig{AutoAdvance: true, IntervalMinutes: 5, RememberHistory: true, MaxHistory: 5},
			hasError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.input.Validate()
			if tt.hasError && err == nil {
				t.Error("expected validation error, got nil")
			}
			if !tt.hasError && err != nil {
				t.Errorf("expected no error, got %v", err)
			}
			if tt.input.IntervalMinutes != tt.expected.IntervalMinutes {
				t.Errorf("expected interval %d, got %d", tt.expected.IntervalMinutes, tt.input.IntervalMinutes)
			}
			if tt.input.MaxHistory != tt.expected.MaxHistory {
				t.Errorf("expected max history %d, got %d", tt.expected.MaxHistory, tt.input.MaxHistory)
			}
		})
	}
}

func TestAppearanceConfigValidation(t *testing.T) {
	tests := []struct {
		name     string
		input    AppearanceConfig
		expected AppearanceConfig
		hasError bool
	}{
		{
			name: "valid config",
			input: AppearanceConfig{
				HeaderMode: "text", HeaderAlign: "center", HeaderWidth: 50,
				CustomText: "My Radio", HeaderColor: "auto", HeaderBold: true,
				PaddingTop: 1, PaddingBottom: 0,
			},
			expected: AppearanceConfig{
				HeaderMode: "text", HeaderAlign: "center", HeaderWidth: 50,
				CustomText: "My Radio", HeaderColor: "auto", HeaderBold: true,
				PaddingTop: 1, PaddingBottom: 0,
			},
			hasError: false,
		},
		{
			name:     "invalid mode",
			input:    AppearanceConfig{HeaderMode: "invalid", HeaderAlign: "center", HeaderWidth: 50},
			expected: AppearanceConfig{HeaderMode: "default", HeaderAlign: "center", HeaderWidth: 50},
			hasError: true,
		},
		{
			name:     "invalid alignment",
			input:    AppearanceConfig{HeaderMode: "text", HeaderAlign: "middle", HeaderWidth: 50},
			expected: AppearanceConfig{HeaderMode: "text", HeaderAlign: "center", HeaderWidth: 50},
			hasError: true,
		},
		{
			name:     "width too small",
			input:    AppearanceConfig{HeaderMode: "text", HeaderAlign: "center", HeaderWidth: 5},
			expected: AppearanceConfig{HeaderMode: "text", HeaderAlign: "center", HeaderWidth: 10},
			hasError: true,
		},
		{
			name:     "width too large",
			input:    AppearanceConfig{HeaderMode: "text", HeaderAlign: "center", HeaderWidth: 150},
			expected: AppearanceConfig{HeaderMode: "text", HeaderAlign: "center", HeaderWidth: 120},
			hasError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.input.Validate()
			if tt.hasError && err == nil {
				t.Error("expected validation error, got nil")
			}
			if !tt.hasError && err != nil {
				t.Errorf("expected no error, got %v", err)
			}
			if tt.input.HeaderMode != tt.expected.HeaderMode {
				t.Errorf("expected mode %s, got %s", tt.expected.HeaderMode, tt.input.HeaderMode)
			}
			if tt.input.HeaderAlign != tt.expected.HeaderAlign {
				t.Errorf("expected align %s, got %s", tt.expected.HeaderAlign, tt.input.HeaderAlign)
			}
			if tt.input.HeaderWidth != tt.expected.HeaderWidth {
				t.Errorf("expected width %d, got %d", tt.expected.HeaderWidth, tt.input.HeaderWidth)
			}
		})
	}
}

func TestPaddingConfigValidation(t *testing.T) {
	tests := []struct {
		name     string
		input    PaddingConfig
		expected PaddingConfig
		hasError bool
	}{
		{
			name: "valid config",
			input: PaddingConfig{
				PageHorizontal: 2, PageVertical: 1, ListItemLeft: 2,
				BoxHorizontal: 2, BoxVertical: 1,
			},
			expected: PaddingConfig{
				PageHorizontal: 2, PageVertical: 1, ListItemLeft: 2,
				BoxHorizontal: 2, BoxVertical: 1,
			},
			hasError: false,
		},
		{
			name: "negative values",
			input: PaddingConfig{
				PageHorizontal: -1, PageVertical: -1, ListItemLeft: -1,
				BoxHorizontal: -1, BoxVertical: -1,
			},
			expected: PaddingConfig{
				PageHorizontal: 0, PageVertical: 0, ListItemLeft: 0,
				BoxHorizontal: 0, BoxVertical: 0,
			},
			hasError: true,
		},
		{
			name: "values too large",
			input: PaddingConfig{
				PageHorizontal: 30, PageVertical: 30, ListItemLeft: 30,
				BoxHorizontal: 30, BoxVertical: 30,
			},
			expected: PaddingConfig{
				PageHorizontal: 20, PageVertical: 20, ListItemLeft: 20,
				BoxHorizontal: 20, BoxVertical: 20,
			},
			hasError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.input.Validate()
			if tt.hasError && err == nil {
				t.Error("expected validation error, got nil")
			}
			if !tt.hasError && err != nil {
				t.Errorf("expected no error, got %v", err)
			}
			if tt.input.PageHorizontal != tt.expected.PageHorizontal {
				t.Errorf("expected page_horizontal %d, got %d", tt.expected.PageHorizontal, tt.input.PageHorizontal)
			}
		})
	}
}

func TestLoadAndSave(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "tera-config-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			t.Logf("failed to clean up temp dir: %v", err)
		}
	}()

	originalFunc := userConfigDirFunc
	defer func() { userConfigDirFunc = originalFunc }()
	userConfigDirFunc = func() (string, error) { return tmpDir, nil }

	cfg, err := Load()
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}
	if cfg.Version != ConfigVersion {
		t.Errorf("expected version %s, got %s", ConfigVersion, cfg.Version)
	}

	cfg.Player.DefaultVolume = 80
	cfg.UI.Theme.Name = "custom"

	if err := Save(cfg); err != nil {
		t.Fatalf("failed to save config: %v", err)
	}

	cfg2, err := Load()
	if err != nil {
		t.Fatalf("failed to reload config: %v", err)
	}
	if cfg2.Player.DefaultVolume != 80 {
		t.Errorf("expected volume 80, got %d", cfg2.Player.DefaultVolume)
	}
	if cfg2.UI.Theme.Name != "custom" {
		t.Errorf("expected theme 'custom', got %s", cfg2.UI.Theme.Name)
	}
}

func TestExists(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "tera-config-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			t.Logf("failed to clean up temp dir: %v", err)
		}
	}()

	originalFunc := userConfigDirFunc
	defer func() { userConfigDirFunc = originalFunc }()
	userConfigDirFunc = func() (string, error) { return tmpDir, nil }

	exists, err := Exists()
	if err != nil {
		t.Fatalf("failed to check existence: %v", err)
	}
	if exists {
		t.Error("config should not exist initially")
	}

	_, err = Load()
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	exists, err = Exists()
	if err != nil {
		t.Fatalf("failed to check existence: %v", err)
	}
	if !exists {
		t.Error("config should exist after load")
	}
}

func TestBackup(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "tera-config-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			t.Logf("failed to clean up temp dir: %v", err)
		}
	}()

	originalFunc := userConfigDirFunc
	defer func() { userConfigDirFunc = originalFunc }()
	userConfigDirFunc = func() (string, error) { return tmpDir, nil }

	cfg := DefaultConfig()
	cfg.Player.DefaultVolume = 80
	if err := Save(&cfg); err != nil {
		t.Fatalf("failed to save config: %v", err)
	}

	if err := Backup("test"); err != nil {
		t.Fatalf("failed to create backup: %v", err)
	}

	configPath, _ := GetConfigPath()
	backupPath := configPath + ".test"
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		t.Error("backup file should exist")
	}
}

func TestReset(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "tera-config-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			t.Logf("failed to clean up temp dir: %v", err)
		}
	}()

	originalFunc := userConfigDirFunc
	defer func() { userConfigDirFunc = originalFunc }()
	userConfigDirFunc = func() (string, error) { return tmpDir, nil }

	cfg := DefaultConfig()
	cfg.Player.DefaultVolume = 50
	if err := Save(&cfg); err != nil {
		t.Fatalf("failed to save config: %v", err)
	}

	if err := Reset(); err != nil {
		t.Fatalf("failed to reset config: %v", err)
	}

	cfg2, err := Load()
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}
	if cfg2.Player.DefaultVolume != 100 {
		t.Errorf("expected default volume 100, got %d", cfg2.Player.DefaultVolume)
	}
}

func TestThemeConfigValidation(t *testing.T) {
	tests := []struct {
		name     string
		input    ThemeConfig
		hasError bool
	}{
		{
			name: "valid theme",
			input: ThemeConfig{
				Name: "custom",
				Colors: map[string]string{
					"primary": "6", "secondary": "12", "highlight": "3",
					"error": "9", "success": "2", "muted": "8", "text": "7",
				},
				Padding: PaddingConfig{
					PageHorizontal: 2, PageVertical: 1, ListItemLeft: 2,
					BoxHorizontal: 2, BoxVertical: 1,
				},
			},
			hasError: false,
		},
		{
			name: "missing colors",
			input: ThemeConfig{
				Name:    "custom",
				Colors:  map[string]string{"primary": "6"},
				Padding: PaddingConfig{},
			},
			hasError: true,
		},
		{
			name: "empty name",
			input: ThemeConfig{
				Name: "",
				Colors: map[string]string{
					"primary": "6", "secondary": "12", "highlight": "3",
					"error": "9", "success": "2", "muted": "8", "text": "7",
				},
				Padding: PaddingConfig{},
			},
			hasError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.input.Validate()
			if tt.hasError && err == nil {
				t.Error("expected validation error, got nil")
			}
			if !tt.hasError && err != nil {
				t.Errorf("expected no error, got %v", err)
			}
			requiredColors := []string{"primary", "secondary", "highlight", "error", "success", "muted", "text"}
			for _, color := range requiredColors {
				if _, exists := tt.input.Colors[color]; !exists {
					t.Errorf("missing required color: %s", color)
				}
			}
		})
	}
}

// TestLoadLegacyConfig_MissingPlayHistory verifies that loading a pre-3.7 config
// file that has no play_history section yields DefaultPlayHistoryConfig values
// rather than Go zero-values ({enabled:false, size:0}).
func TestLoadLegacyConfig_MissingPlayHistory(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "tera-legacy-config-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			t.Logf("failed to clean up temp dir: %v", err)
		}
	}()

	originalFunc := userConfigDirFunc
	defer func() { userConfigDirFunc = originalFunc }()
	userConfigDirFunc = func() (string, error) { return tmpDir, nil }

	// Write a legacy config that intentionally omits the play_history section.
	legacyYAML := `version: "3.0"
player:
  default_volume: 80
  buffer_size_mb: 50
`
	configDir, _ := GetConfigDir()
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}
	configPath, _ := GetConfigPath()
	if err := os.WriteFile(configPath, []byte(legacyYAML), 0644); err != nil {
		t.Fatalf("failed to write legacy config: %v", err)
	}

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	defaults := DefaultPlayHistoryConfig()
	if cfg.PlayHistory.Enabled != defaults.Enabled {
		t.Errorf("PlayHistory.Enabled: got %v, want %v", cfg.PlayHistory.Enabled, defaults.Enabled)
	}
	if cfg.PlayHistory.Size != defaults.Size {
		t.Errorf("PlayHistory.Size: got %d, want %d", cfg.PlayHistory.Size, defaults.Size)
	}
	if cfg.PlayHistory.AllowDuplicate != defaults.AllowDuplicate {
		t.Errorf("PlayHistory.AllowDuplicate: got %v, want %v", cfg.PlayHistory.AllowDuplicate, defaults.AllowDuplicate)
	}
}

func TestPlayHistoryConfigDefaults(t *testing.T) {
	cfg := DefaultConfig()

	if !cfg.PlayHistory.Enabled {
		t.Error("expected PlayHistory.Enabled to be true")
	}
	if cfg.PlayHistory.Size != 5 {
		t.Errorf("expected PlayHistory.Size 5, got %d", cfg.PlayHistory.Size)
	}
	if cfg.PlayHistory.AllowDuplicate {
		t.Error("expected PlayHistory.AllowDuplicate to be false")
	}
}

func TestPlayHistoryConfigValidation(t *testing.T) {
	tests := []struct {
		name         string
		input        PlayHistoryConfig
		expectedSize int
		hasError     bool
	}{
		{
			name:         "valid config",
			input:        PlayHistoryConfig{Enabled: true, Size: 5, AllowDuplicate: false},
			expectedSize: 5,
			hasError:     false,
		},
		{
			name:         "size at minimum boundary",
			input:        PlayHistoryConfig{Enabled: true, Size: 1, AllowDuplicate: false},
			expectedSize: 1,
			hasError:     false,
		},
		{
			name:         "size at maximum boundary",
			input:        PlayHistoryConfig{Enabled: true, Size: 20, AllowDuplicate: false},
			expectedSize: 20,
			hasError:     false,
		},
		{
			name:         "size zero — clamped to 1",
			input:        PlayHistoryConfig{Enabled: true, Size: 0, AllowDuplicate: false},
			expectedSize: 1,
			hasError:     true,
		},
		{
			name:         "size negative — clamped to 1",
			input:        PlayHistoryConfig{Enabled: true, Size: -3, AllowDuplicate: false},
			expectedSize: 1,
			hasError:     true,
		},
		{
			name:         "size too large — clamped to 20",
			input:        PlayHistoryConfig{Enabled: true, Size: 25, AllowDuplicate: false},
			expectedSize: 20,
			hasError:     true,
		},
		{
			name:         "disabled with valid size",
			input:        PlayHistoryConfig{Enabled: false, Size: 10, AllowDuplicate: true},
			expectedSize: 10,
			hasError:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.input.Validate()
			if tt.hasError && err == nil {
				t.Error("expected validation error, got nil")
			}
			if !tt.hasError && err != nil {
				t.Errorf("expected no error, got %v", err)
			}
			if tt.input.Size != tt.expectedSize {
				t.Errorf("expected size %d after validation, got %d", tt.expectedSize, tt.input.Size)
			}
		})
	}
}
