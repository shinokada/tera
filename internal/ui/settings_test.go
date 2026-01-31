package ui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestNewSettingsModel(t *testing.T) {
	m := NewSettingsModel()

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
	m := NewSettingsModel()
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
		{"Press 2 for About", "2", settingsStateAbout},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewSettingsModel()
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
			m := NewSettingsModel()
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
	m := NewSettingsModel()
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
	m := NewSettingsModel()
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
	m := NewSettingsModel()

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
	m := NewSettingsModel()
	m.width = 80
	m.height = 24

	view := m.View()

	if !strings.Contains(view, "Settings") {
		t.Error("Expected view to contain 'Settings'")
	}
}

func TestSettingsViewTheme(t *testing.T) {
	m := NewSettingsModel()
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
	m := NewSettingsModel()
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
