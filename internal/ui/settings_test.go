package ui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/shinokada/tera/internal/api"
)

func TestNewSettingsModel(t *testing.T) {
	m := NewSettingsModel(t.TempDir())

	if m.state != settingsStateMenu {
		t.Errorf("Expected initial state to be settingsStateMenu, got %v", m.state)
	}

	// currentTheme should be set to one of the predefined themes
	// (depends on user's saved theme, so we just check it's not empty)
	validThemes := []string{"Default", "Ocean", "Forest", "Sunset", "Purple Haze", "Monochrome", "Dracula", "Nord"}
	isValid := false
	for _, theme := range validThemes {
		if m.currentTheme == theme {
			isValid = true
			break
		}
	}
	if !isValid {
		t.Errorf("Expected currentTheme to be a valid predefined theme, got '%s'", m.currentTheme)
	}
}

func TestSettingsModelInit(t *testing.T) {
	m := NewSettingsModel(t.TempDir())
	cmd := m.Init()

	if cmd != nil {
		t.Error("Expected Init to return nil")
	}
}

func TestSettingsMenuNavigation(t *testing.T) {
	tests := []struct {
		name          string
		key           string
		expectedState settingsState
	}{
		{"Press 1 for Theme", "1", settingsStateTheme},
		{"Press 3 for History", "3", settingsStateHistory},
		{"Press 4 for Updates", "4", settingsStateUpdates},
		{"Press 5 for About", "5", settingsStateAbout},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewSettingsModel(t.TempDir())
			m.width = 80
			m.height = 24

			msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(tt.key)}
			newModel, _ := m.Update(msg)
			updatedModel := newModel.(SettingsModel)

			if updatedModel.state != tt.expectedState {
				t.Errorf("Expected state %v, got %v", tt.expectedState, updatedModel.state)
			}
		})
	}
}

func TestSettingsMenuEscNavigation(t *testing.T) {
	tests := []struct {
		name string
		key  string
	}{
		{"Esc key", "esc"},
		{"0 key", "0"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewSettingsModel(t.TempDir())
			m.width = 80
			m.height = 24

			msg := tea.KeyMsg{Type: tea.KeyEscape}
			if tt.key == "0" {
				msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("0")}
			}

			_, cmd := m.Update(msg)

			// Should return a command that produces backToMainMsg
			if cmd == nil {
				t.Error("Expected a command to be returned")
			}
		})
	}
}

func TestSettingsThemeEscBack(t *testing.T) {
	m := NewSettingsModel(t.TempDir())
	m.state = settingsStateTheme
	m.width = 80
	m.height = 24

	msg := tea.KeyMsg{Type: tea.KeyEscape}
	newModel, _ := m.Update(msg)
	updatedModel := newModel.(SettingsModel)

	if updatedModel.state != settingsStateMenu {
		t.Errorf("Expected state to be settingsStateMenu, got %v", updatedModel.state)
	}
}

func TestSettingsAboutEscBack(t *testing.T) {
	m := NewSettingsModel(t.TempDir())
	m.state = settingsStateAbout
	m.width = 80
	m.height = 24

	msg := tea.KeyMsg{Type: tea.KeyEscape}
	newModel, _ := m.Update(msg)
	updatedModel := newModel.(SettingsModel)

	if updatedModel.state != settingsStateMenu {
		t.Errorf("Expected state to be settingsStateMenu, got %v", updatedModel.state)
	}
}

func TestSettingsWindowSizeUpdate(t *testing.T) {
	m := NewSettingsModel(t.TempDir())

	msg := tea.WindowSizeMsg{Width: 100, Height: 50}
	newModel, _ := m.Update(msg)
	updatedModel := newModel.(SettingsModel)

	if updatedModel.width != 100 {
		t.Errorf("Expected width to be 100, got %d", updatedModel.width)
	}

	if updatedModel.height != 50 {
		t.Errorf("Expected height to be 50, got %d", updatedModel.height)
	}
}

func TestSettingsViewMenu(t *testing.T) {
	m := NewSettingsModel(t.TempDir())
	m.width = 80
	m.height = 24

	view := m.View()

	if !strings.Contains(view, "Settings") {
		t.Error("Expected view to contain 'Settings'")
	}
}

func TestSettingsViewTheme(t *testing.T) {
	m := NewSettingsModel(t.TempDir())
	m.state = settingsStateTheme
	m.width = 80
	m.height = 24

	view := m.View()

	if !strings.Contains(view, "Theme") {
		t.Error("Expected view to contain 'Theme'")
	}

	if !strings.Contains(view, "Current theme") {
		t.Error("Expected view to contain 'Current theme'")
	}
}

func TestSettingsViewAbout(t *testing.T) {
	m := NewSettingsModel(t.TempDir())
	m.state = settingsStateAbout
	m.width = 80
	m.height = 24

	view := m.View()

	if !strings.Contains(view, "About") {
		t.Error("Expected view to contain 'About'")
	}

	if !strings.Contains(view, "Version") {
		t.Error("Expected view to contain 'Version'")
	}

	if !strings.Contains(view, "Author") {
		t.Error("Expected view to contain 'Author'")
	}

	if !strings.Contains(view, "Repository") {
		t.Error("Expected view to contain 'Repository'")
	}

	if !strings.Contains(view, "License") {
		t.Error("Expected view to contain 'License'")
	}
}

func TestPredefinedThemesExist(t *testing.T) {
	expectedThemes := []string{
		"Default",
		"Ocean",
		"Forest",
		"Sunset",
		"Purple Haze",
		"Monochrome",
		"Dracula",
		"Nord",
	}

	if len(predefinedThemes) != len(expectedThemes) {
		t.Errorf("Expected %d predefined themes, got %d", len(expectedThemes), len(predefinedThemes))
	}

	for i, expected := range expectedThemes {
		if predefinedThemes[i].name != expected {
			t.Errorf("Expected theme %d to be '%s', got '%s'", i, expected, predefinedThemes[i].name)
		}
	}
}

func TestThemeItemInterface(t *testing.T) {
	item := themeItem{
		name:        "Test Theme",
		description: "A test theme",
	}

	if item.FilterValue() != "Test Theme" {
		t.Errorf("Expected FilterValue to return 'Test Theme', got '%s'", item.FilterValue())
	}

	if item.Title() != "Test Theme" {
		t.Errorf("Expected Title to return 'Test Theme', got '%s'", item.Title())
	}

	if item.Description() != "A test theme" {
		t.Errorf("Expected Description to return 'A test theme', got '%s'", item.Description())
	}
}

func TestVersionVariable(t *testing.T) {
	// Version should be set (at least to "dev" by default)
	if Version == "" {
		t.Error("Version should not be empty")
	}
}

func TestSettingsUpdatesEscBack(t *testing.T) {
	m := NewSettingsModel(t.TempDir())
	m.state = settingsStateUpdates
	m.width = 80
	m.height = 24

	msg := tea.KeyMsg{Type: tea.KeyEscape}
	newModel, _ := m.Update(msg)
	updatedModel := newModel.(SettingsModel)

	if updatedModel.state != settingsStateMenu {
		t.Errorf("Expected state to be settingsStateMenu, got %v", updatedModel.state)
	}
}

func TestSettingsUpdatesRefresh(t *testing.T) {
	m := NewSettingsModel(t.TempDir())
	m.state = settingsStateUpdates
	m.width = 80
	m.height = 24

	// Press 'r' to refresh
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("r")}
	newModel, cmd := m.Update(msg)
	updatedModel := newModel.(SettingsModel)

	// Should be checking for updates
	if !updatedModel.updateChecking {
		t.Error("Expected updateChecking to be true after pressing 'r'")
	}

	// Should return a command (the version check)
	if cmd == nil {
		t.Error("Expected a command to be returned for version check")
	}
}

func TestSettingsViewUpdates(t *testing.T) {
	m := NewSettingsModel(t.TempDir())
	m.state = settingsStateUpdates
	m.width = 80
	m.height = 24

	view := m.View()

	if !strings.Contains(view, "Check for Updates") {
		t.Error("Expected view to contain 'Check for Updates'")
	}

	if !strings.Contains(view, "Current version") {
		t.Error("Expected view to contain 'Current version'")
	}
}

func TestSettingsViewUpdatesChecking(t *testing.T) {
	m := NewSettingsModel(t.TempDir())
	m.state = settingsStateUpdates
	m.updateChecking = true
	m.width = 80
	m.height = 24

	view := m.View()

	if !strings.Contains(view, "Checking for updates") {
		t.Error("Expected view to contain 'Checking for updates' when checking")
	}
}

func TestSettingsViewUpdatesAvailable(t *testing.T) {
	m := NewSettingsModel(t.TempDir())
	m.state = settingsStateUpdates
	m.updateChecked = true
	m.updateAvailable = true
	m.latestVersion = "v2.0.0"
	// Set installInfo to Manual so UpdateCommand is empty and URL is shown
	m.installInfo = api.InstallInfo{
		Method:        api.InstallMethodManual,
		UpdateCommand: "",
		Description:   "Manual/Binary",
	}
	m.width = 80
	m.height = 24

	view := m.View()

	if !strings.Contains(view, "New version available") {
		t.Error("Expected view to contain 'New version available'")
	}

	if !strings.Contains(view, "v2.0.0") {
		t.Error("Expected view to contain the latest version 'v2.0.0'")
	}

	if !strings.Contains(view, "github.com/shinokada/tera/releases/latest") {
		t.Error("Expected view to contain release page URL")
	}
}

func TestSettingsViewUpdatesUpToDate(t *testing.T) {
	m := NewSettingsModel(t.TempDir())
	m.state = settingsStateUpdates
	m.updateChecked = true
	m.updateAvailable = false
	m.latestVersion = "v1.0.0"
	m.width = 80
	m.height = 24

	view := m.View()

	if !strings.Contains(view, "up to date") {
		t.Error("Expected view to contain 'up to date' when on latest version")
	}
}

func TestSettingsViewUpdatesError(t *testing.T) {
	m := NewSettingsModel(t.TempDir())
	m.state = settingsStateUpdates
	m.updateChecked = true
	m.updateError = "network error"
	m.width = 80
	m.height = 24

	view := m.View()

	if !strings.Contains(view, "Error") {
		t.Error("Expected view to contain 'Error' when update check failed")
	}

	if !strings.Contains(view, "network error") {
		t.Error("Expected view to contain the error message")
	}
}

func TestVersionCheckMsgHandling(t *testing.T) {
	m := NewSettingsModel(t.TempDir())
	m.state = settingsStateUpdates
	m.updateChecking = true
	m.width = 80
	m.height = 24

	// Simulate receiving version check result
	msg := versionCheckMsg{latestVersion: "v2.0.0", err: nil}
	newModel, _ := m.Update(msg)
	updatedModel := newModel.(SettingsModel)

	if updatedModel.updateChecking {
		t.Error("Expected updateChecking to be false after receiving result")
	}

	if !updatedModel.updateChecked {
		t.Error("Expected updateChecked to be true after receiving result")
	}

	if updatedModel.latestVersion != "v2.0.0" {
		t.Errorf("Expected latestVersion to be 'v2.0.0', got '%s'", updatedModel.latestVersion)
	}
}

func TestSettingsMenuNavigateToShuffleSettings(t *testing.T) {
	m := NewSettingsModel(t.TempDir())
	m.width = 80
	m.height = 24

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("2")}
	_, cmd := m.Update(msg)

	// Should return a command that produces navigateMsg to screenShuffleSettings
	if cmd == nil {
		t.Error("Expected a command to be returned for shuffle settings navigation")
	}
}
