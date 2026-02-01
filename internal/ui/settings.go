package ui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/shinokada/tera/internal/api"
	"github.com/shinokada/tera/internal/theme"
	"github.com/shinokada/tera/internal/ui/components"
)

// settingsState represents the current state in the settings screen
type settingsState int

const (
	settingsStateMenu settingsState = iota
	settingsStateTheme
	settingsStateUpdates
	settingsStateAbout
)

// Version is set from main.go
var Version = "dev"

// SettingsModel represents the settings screen
type SettingsModel struct {
	state            settingsState
	menuList         list.Model
	themeList        list.Model
	width            int
	height           int
	message          string
	messageTime      int
	messageIsSuccess bool
	currentTheme     string
	// Update checking
	latestVersion   string
	updateAvailable bool
	updateChecked   bool
	updateChecking  bool
	updateError     string
}

// Predefined themes
var predefinedThemes = []struct {
	name        string
	description string
	colors      theme.ColorConfig
}{
	{
		name:        "Default",
		description: "Cyan and blue tones",
		colors: theme.ColorConfig{
			Primary:   "6",
			Secondary: "12",
			Highlight: "3",
			Error:     "9",
			Success:   "2",
			Muted:     "8",
			Text:      "7",
		},
	},
	{
		name:        "Ocean",
		description: "Deep blue theme",
		colors: theme.ColorConfig{
			Primary:   "33",
			Secondary: "39",
			Highlight: "51",
			Error:     "196",
			Success:   "46",
			Muted:     "240",
			Text:      "255",
		},
	},
	{
		name:        "Forest",
		description: "Green nature theme",
		colors: theme.ColorConfig{
			Primary:   "34",
			Secondary: "28",
			Highlight: "226",
			Error:     "196",
			Success:   "46",
			Muted:     "242",
			Text:      "255",
		},
	},
	{
		name:        "Sunset",
		description: "Warm orange and red",
		colors: theme.ColorConfig{
			Primary:   "208",
			Secondary: "202",
			Highlight: "226",
			Error:     "196",
			Success:   "46",
			Muted:     "242",
			Text:      "255",
		},
	},
	{
		name:        "Purple Haze",
		description: "Purple and magenta",
		colors: theme.ColorConfig{
			Primary:   "135",
			Secondary: "99",
			Highlight: "219",
			Error:     "196",
			Success:   "46",
			Muted:     "242",
			Text:      "255",
		},
	},
	{
		name:        "Monochrome",
		description: "Classic black and white",
		colors: theme.ColorConfig{
			Primary:   "255",
			Secondary: "250",
			Highlight: "226",
			Error:     "196",
			Success:   "46",
			Muted:     "242",
			Text:      "255",
		},
	},
	{
		name:        "Dracula",
		description: "Popular dark theme",
		colors: theme.ColorConfig{
			Primary:   "141",
			Secondary: "212",
			Highlight: "228",
			Error:     "210",
			Success:   "84",
			Muted:     "242",
			Text:      "255",
		},
	},
	{
		name:        "Nord",
		description: "Arctic, north-bluish",
		colors: theme.ColorConfig{
			Primary:   "110",
			Secondary: "109",
			Highlight: "222",
			Error:     "167",
			Success:   "108",
			Muted:     "60",
			Text:      "255",
		},
	},
}

// themeItem for the theme list
type themeItem struct {
	name        string
	description string
}

func (i themeItem) FilterValue() string { return i.name }
func (i themeItem) Title() string       { return i.name }
func (i themeItem) Description() string { return i.description }

// versionCheckMsg is sent when version check completes
type versionCheckMsg struct {
	latestVersion string
	err           error
}

// checkForUpdates performs the version check in the background
func checkForUpdates() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		checker := api.NewVersionChecker()
		release, err := checker.GetLatestRelease(ctx)
		if err != nil {
			return versionCheckMsg{err: err}
		}
		return versionCheckMsg{latestVersion: release.TagName}
	}
}

// NewSettingsModel creates a new settings screen model
func NewSettingsModel() SettingsModel {
	// Main settings menu
	menuItems := []components.MenuItem{
		components.NewMenuItem("Theme / Colors", "Choose a color theme", "1"),
		components.NewMenuItem("Check for Updates", "Check for new versions", "2"),
		components.NewMenuItem("About TERA", "Version and information", "3"),
	}
	menuList := components.CreateMenu(menuItems, "", 50, 10)

	// Theme selection list
	themeItems := make([]list.Item, len(predefinedThemes))
	for i, t := range predefinedThemes {
		themeItems[i] = themeItem{
			name:        t.name,
			description: t.description,
		}
	}

	delegate := createStyledDelegate()
	themeList := list.New(themeItems, delegate, 50, 15)
	themeList.Title = "Select a Theme"
	themeList.SetShowStatusBar(false)
	themeList.SetFilteringEnabled(false)
	themeList.SetShowHelp(false)

	// Get current theme name - detect from saved theme
	currentTheme := "Default"
	if current := theme.Current(); current != nil {
		// Try to match current colors to a predefined theme
		for _, t := range predefinedThemes {
			if t.colors == current.Colors {
				currentTheme = t.name
				break
			}
		}
	}

	return SettingsModel{
		state:        settingsStateMenu,
		menuList:     menuList,
		themeList:    themeList,
		currentTheme: currentTheme,
		// Update fields initialized to defaults
		updateChecked:  false,
		updateChecking: false,
	}
}

// Init initializes the settings screen
func (m SettingsModel) Init() tea.Cmd {
	return nil
}

// Update handles messages for the settings screen
func (m SettingsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case versionCheckMsg:
		m.updateChecking = false
		m.updateChecked = true
		if msg.err != nil {
			m.updateError = msg.err.Error()
			return m, nil
		}
		m.latestVersion = msg.latestVersion
		m.updateAvailable = api.IsNewerVersion(Version, msg.latestVersion)
		return m, nil

	case tea.KeyMsg:
		switch m.state {
		case settingsStateMenu:
			return m.updateMenu(msg)
		case settingsStateTheme:
			return m.updateTheme(msg)
		case settingsStateUpdates:
			return m.updateUpdates(msg)
		case settingsStateAbout:
			return m.updateAbout(msg)
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		h, v := docStyle().GetFrameSize()
		m.menuList.SetSize(msg.Width-h, msg.Height-v-8)
		m.themeList.SetSize(msg.Width-h, msg.Height-v-10)
		return m, nil

	case tickMsg:
		if m.messageTime > 0 {
			m.messageTime--
			if m.messageTime == 0 {
				m.message = ""
			}
			return m, tickEverySecond()
		}
		return m, nil
	}

	return m, nil
}

func (m SettingsModel) updateMenu(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "0":
		return m, func() tea.Msg { return backToMainMsg{} }

	case "1":
		m.state = settingsStateTheme
		return m, nil

	case "2":
		m.state = settingsStateUpdates
		if !m.updateChecked && !m.updateChecking {
			m.updateChecking = true
			return m, checkForUpdates()
		}
		return m, nil

	case "3":
		m.state = settingsStateAbout
		return m, nil

	case "enter":
		idx := m.menuList.Index()
		switch idx {
		case 0:
			m.state = settingsStateTheme
		case 1:
			m.state = settingsStateUpdates
			if !m.updateChecked && !m.updateChecking {
				m.updateChecking = true
				return m, checkForUpdates()
			}
		case 2:
			m.state = settingsStateAbout
		}
		return m, nil
	}

	// Handle menu navigation
	var cmd tea.Cmd
	m.menuList, cmd = m.menuList.Update(msg)
	return m, cmd
}

func (m SettingsModel) updateTheme(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.state = settingsStateMenu
		return m, nil

	case "0":
		return m, func() tea.Msg { return backToMainMsg{} }

	case "enter":
		// Apply selected theme
		idx := m.themeList.Index()
		if idx >= 0 && idx < len(predefinedThemes) {
			selectedTheme := predefinedThemes[idx]
			currentTheme := theme.Current()
			if currentTheme == nil {
				m.message = "✗ Failed to load current theme"
				m.messageIsSuccess = false
				m.messageTime = 3
				return m, tickEverySecond()
			}
			currentTheme.Colors = selectedTheme.colors

			// Save the theme
			if err := theme.Save(currentTheme); err != nil {
				m.message = fmt.Sprintf("✗ Failed to save theme: %v", err)
				m.messageIsSuccess = false
			} else {
				m.message = fmt.Sprintf("✓ Theme '%s' applied!", selectedTheme.name)
				m.messageIsSuccess = true
				m.currentTheme = selectedTheme.name
			}
			// messageTime is in seconds (tickEverySecond ticks once per second)
			m.messageTime = 3
		}
		return m, tickEverySecond()
	}

	// Handle list navigation
	var cmd tea.Cmd
	m.themeList, cmd = m.themeList.Update(msg)
	return m, cmd
}

func (m SettingsModel) updateUpdates(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.state = settingsStateMenu
		return m, nil
	case "0":
		return m, func() tea.Msg { return backToMainMsg{} }
	case "r":
		// Refresh/recheck for updates
		m.updateChecking = true
		m.updateChecked = false
		m.updateError = ""
		return m, checkForUpdates()
	}
	return m, nil
}

func (m SettingsModel) updateAbout(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "enter":
		m.state = settingsStateMenu
		return m, nil
	case "0":
		return m, func() tea.Msg { return backToMainMsg{} }
	}
	return m, nil
}

// View renders the settings screen
func (m SettingsModel) View() string {
	switch m.state {
	case settingsStateMenu:
		return m.viewMenu()
	case settingsStateTheme:
		return m.viewTheme()
	case settingsStateUpdates:
		return m.viewUpdates()
	case settingsStateAbout:
		return m.viewAbout()
	}
	return "Unknown state"
}

func (m SettingsModel) viewMenu() string {
	var content strings.Builder

	content.WriteString(m.menuList.View())

	if m.message != "" {
		content.WriteString("\n\n")
		if m.messageIsSuccess {
			content.WriteString(successStyle().Render(m.message))
		} else {
			content.WriteString(errorStyle().Render(m.message))
		}
	}

	return RenderPageWithBottomHelp(PageLayout{
		Title:   "⚙️  Settings",
		Content: content.String(),
		Help:    "↑↓/jk: Navigate • Enter: Select • 1-3: Shortcut • Esc/0: Back • Ctrl+C: Quit",
	}, m.height)
}

func (m SettingsModel) viewTheme() string {
	var content strings.Builder

	content.WriteString(subtitleStyle().Render("Current theme: "))
	content.WriteString(highlightStyle().Render(m.currentTheme))
	content.WriteString("\n\n")

	content.WriteString(m.themeList.View())

	if m.message != "" {
		content.WriteString("\n\n")
		if m.messageIsSuccess {
			content.WriteString(successStyle().Render(m.message))
		} else {
			content.WriteString(errorStyle().Render(m.message))
		}
	}

	return RenderPageWithBottomHelp(PageLayout{
		Title:   "🎨 Theme / Colors",
		Content: content.String(),
		Help:    "↑↓/jk: Navigate • Enter: Apply Theme • Esc: Back • 0: Main Menu • Ctrl+C: Quit",
	}, m.height)
}

func (m SettingsModel) viewUpdates() string {
	var content strings.Builder

	// Current version
	currentVersion := Version
	if currentVersion == "" {
		currentVersion = "dev"
	}
	content.WriteString(stationFieldStyle().Render("Current version: "))
	content.WriteString(highlightStyle().Render(currentVersion))
	content.WriteString("\n\n")

	content.WriteString(helpStyle().Render("─────────────────────────────────────────"))
	content.WriteString("\n\n")

	if m.updateChecking {
		content.WriteString(stationValueStyle().Render("⏳ Checking for updates..."))
	} else if m.updateError != "" {
		content.WriteString(errorStyle().Render("✗ Error checking for updates:"))
		content.WriteString("\n")
		content.WriteString(helpStyle().Render("  " + m.updateError))
		content.WriteString("\n\n")
		content.WriteString(stationValueStyle().Render("Press 'r' to retry"))
	} else if m.updateChecked {
		if m.updateAvailable {
			content.WriteString(successStyle().Render("⬆ New version available!"))
			content.WriteString("\n\n")
			content.WriteString(stationFieldStyle().Render("Latest version: "))
			content.WriteString(highlightStyle().Render(m.latestVersion))
			content.WriteString("\n\n")
			content.WriteString(stationFieldStyle().Render("Release page:"))
			content.WriteString("\n")
			content.WriteString(stationValueStyle().Render("  " + api.ReleasePageURL))
			content.WriteString("\n\n")
			content.WriteString(helpStyle().Render("─────────────────────────────────────────"))
			content.WriteString("\n\n")
			content.WriteString(helpStyle().Render("Update instructions:"))
			content.WriteString("\n\n")
			content.WriteString(stationValueStyle().Render("  If installed via Go:"))
			content.WriteString("\n")
			content.WriteString(highlightStyle().Render("    go install github.com/shinokada/tera/cmd/tera@latest"))
			content.WriteString("\n\n")
			content.WriteString(stationValueStyle().Render("  Or visit the release page to download binaries."))
		} else {
			content.WriteString(successStyle().Render("✓ You're up to date!"))
			content.WriteString("\n\n")
			content.WriteString(stationFieldStyle().Render("Latest version: "))
			content.WriteString(highlightStyle().Render(m.latestVersion))
		}
	} else {
		content.WriteString(stationValueStyle().Render("Press Enter or 'r' to check for updates"))
	}

	return RenderPageWithBottomHelp(PageLayout{
		Title:   "🔄 Check for Updates",
		Content: content.String(),
		Help:    "r: Refresh • Esc: Back • 0: Main Menu • Ctrl+C: Quit",
	}, m.height)
}

func (m SettingsModel) viewAbout() string {
	var content strings.Builder

	// Version (with fallback)
	version := Version
	if version == "" {
		version = "dev"
	}
	content.WriteString(stationFieldStyle().Render("Version:    "))
	content.WriteString(highlightStyle().Render(version))
	content.WriteString("\n\n")

	// Author
	content.WriteString(stationFieldStyle().Render("Author:     "))
	content.WriteString(stationValueStyle().Render("Shinichi Okada"))
	content.WriteString("\n\n")

	// Repository
	content.WriteString(stationFieldStyle().Render("Repository: "))
	content.WriteString(stationValueStyle().Render("https://github.com/shinokada/tera"))
	content.WriteString("\n\n")

	// Website
	content.WriteString(stationFieldStyle().Render("Website:    "))
	content.WriteString(stationValueStyle().Render("https://tera.codewithshin.com"))
	content.WriteString("\n\n")

	// License
	content.WriteString(stationFieldStyle().Render("License:    "))
	content.WriteString(stationValueStyle().Render("MIT"))
	content.WriteString("\n\n")

	// Description
	content.WriteString(helpStyle().Render("─────────────────────────────────────────"))
	content.WriteString("\n\n")
	content.WriteString(stationValueStyle().Render("TERA is a terminal-based internet radio player"))
	content.WriteString("\n")
	content.WriteString(stationValueStyle().Render("powered by Radio Browser API."))
	content.WriteString("\n\n")
	content.WriteString(helpStyle().Render("Requires: mpv for audio playback"))

	return RenderPageWithBottomHelp(PageLayout{
		Title:   "ℹ️  About TERA",
		Content: content.String(),
		Help:    "Esc/Enter: Back • 0: Main Menu • Ctrl+C: Quit",
	}, m.height)
}
