package ui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/shinokada/tera/v3/internal/api"
	"github.com/shinokada/tera/v3/internal/config"
	"github.com/shinokada/tera/v3/internal/storage"
	"github.com/shinokada/tera/v3/internal/theme"
	"github.com/shinokada/tera/v3/internal/ui/components"
)

// settingsState represents the current state in the settings screen
type settingsState int

const (
	settingsStateMenu settingsState = iota
	settingsStateTheme
	settingsStateConnection
	settingsStateHistory        // top-level: Search History or Play History
	settingsStateSearchHistory  // existing search history sub-menu
	settingsStatePlayHistory    // play history settings
	settingsStatePlayOptions    // play options (continuous playback, volume, etc.)
	settingsStateUpdates
	settingsStateAbout
)

// Version is set from main.go
var Version = "dev"

// SettingsModel represents the settings screen
type SettingsModel struct {
	state                settingsState
	menuList             list.Model
	themeList            list.Model
	historyMenuList      list.Model
	historyMenuIndex     int // cursor for the top-level History switcher
	width                int
	height               int
	message              string
	messageTime          int
	messageIsSuccess     bool
	currentTheme         string
	favoritePath         string
	searchHistory        *storage.SearchHistoryStore
	playHistoryCfg       config.PlayHistoryConfig
	metadataManager      *storage.MetadataManager
	// Update checking
	latestVersion   string
	updateAvailable bool
	updateChecked   bool
	updateChecking  bool
	updateError     string
	installInfo     api.InstallInfo
	// Play options
	playOptsCfg config.PlayOptionsConfig
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
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
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
func NewSettingsModel(favoritePath string) SettingsModel {
	// Main settings menu
	menuItems := []components.MenuItem{
		components.NewMenuItem("Theme / Colors", "Choose a color theme", "1"),
		components.NewMenuItem("Appearance", "Customize header and layout", "2"),
		components.NewMenuItem("Connection Settings", "Auto-reconnect and buffering", "3"),
		components.NewMenuItem("Shuffle Settings", "Configure shuffle mode behavior", "4"),
		components.NewMenuItem("Play Options", "Playback behaviour settings", "5"),
		components.NewMenuItem("History", "Search and play history settings", "6"),
		components.NewMenuItem("Check for Updates", "Check for new versions", "7"),
		components.NewMenuItem("About TERA", "Version and information", "8"),
	}
	menuList := components.CreateMenu(menuItems, "", 50, 12)

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

	// Search history settings menu
	historyMenuItems := []components.MenuItem{
		components.NewMenuItem("Increase (+5)", "", "1"),
		components.NewMenuItem("Decrease (-5)", "", "2"),
		components.NewMenuItem("Reset to Default", "", "3"),
		components.NewMenuItem("Clear History", "", "4"),
		components.NewMenuItem("Back", "", "5"),
	}
	historyMenuList := components.CreateMenu(historyMenuItems, "", 50, 10)

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

	// Load search history
	store := storage.NewStorage(favoritePath)
	history, err := store.LoadSearchHistory(context.Background())
	if err != nil || history == nil {
		history = storage.NewSearchHistoryStore()
	}

	// Load play history config
	phCfg, err := storage.LoadPlayHistoryConfigFromUnified()
	if err != nil {
		phCfg = config.DefaultPlayHistoryConfig()
	}

	// Load play options config
	poCfg, err := storage.LoadPlayOptionsConfigFromUnified()
	if err != nil {
		poCfg = config.DefaultPlayOptionsConfig()
	}

	// Detect installation method
	installInfo := api.DetectInstallMethod()

	return SettingsModel{
		state:           settingsStateMenu,
		menuList:        menuList,
		themeList:       themeList,
		historyMenuList: historyMenuList,
		currentTheme:    currentTheme,
		favoritePath:    favoritePath,
		searchHistory:   history,
		playHistoryCfg:  phCfg,
		playOptsCfg:     poCfg,
		// Update fields initialized to defaults
		updateChecked:  false,
		updateChecking: false,
		installInfo:    installInfo,
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
		case settingsStateConnection:
			// Connection settings handled in app.go
			return m, nil
		case settingsStateHistory:
			return m.updateHistoryMenu(msg)
		case settingsStateSearchHistory:
			return m.updateSearchHistory(msg)
		case settingsStatePlayHistory:
			return m.updatePlayHistory(msg)
		case settingsStatePlayOptions:
			return m.updatePlayOptions(msg)
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
		// Navigate to appearance settings
		return m, func() tea.Msg {
			return navigateMsg{screen: screenAppearanceSettings}
		}

	case "3":
		// Navigate to connection settings
		return m, func() tea.Msg {
			return navigateMsg{screen: screenConnectionSettings}
		}

	case "4":
		// Navigate to shuffle settings
		return m, func() tea.Msg {
			return navigateMsg{screen: screenShuffleSettings}
		}

	case "5":
		m.state = settingsStatePlayOptions
		return m, nil

	case "6":
		m.state = settingsStateHistory
		return m, nil

	case "7":
		m.state = settingsStateUpdates
		if !m.updateChecked && !m.updateChecking {
			m.updateChecking = true
			return m, checkForUpdates()
		}
		return m, nil

	case "8":
		m.state = settingsStateAbout
		return m, nil

	case "enter":
		idx := m.menuList.Index()
		switch idx {
		case 0:
			m.state = settingsStateTheme
		case 1:
			// Navigate to appearance settings
			return m, func() tea.Msg {
				return navigateMsg{screen: screenAppearanceSettings}
			}
		case 2:
			// Navigate to connection settings
			return m, func() tea.Msg {
				return navigateMsg{screen: screenConnectionSettings}
			}
		case 3:
			// Navigate to shuffle settings
			return m, func() tea.Msg {
				return navigateMsg{screen: screenShuffleSettings}
			}
		case 4:
			m.state = settingsStatePlayOptions
		case 5:
			m.state = settingsStateHistory
		case 6:
			m.state = settingsStateUpdates
			if !m.updateChecked && !m.updateChecking {
				m.updateChecking = true
				return m, checkForUpdates()
			}
		case 7:
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
	case "r", "enter":
		// Refresh/recheck for updates
		if m.updateChecking {
			return m, nil
		}
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

// updateHistoryMenu handles the top-level History switcher (Search vs Play).
func (m SettingsModel) updateHistoryMenu(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	const historyMenuItems = 3 // Search History, Play History, Back
	switch msg.String() {
	case "esc", "3":
		m.state = settingsStateMenu
		return m, nil
	case "0":
		return m, func() tea.Msg { return backToMainMsg{} }
	case "1":
		m.historyMenuIndex = 0
		m.state = settingsStateSearchHistory
		return m, nil
	case "2":
		m.historyMenuIndex = 0
		m.state = settingsStatePlayHistory
		return m, nil
	case "up", "k":
		if m.historyMenuIndex > 0 {
			m.historyMenuIndex--
		}
		return m, nil
	case "down", "j":
		if m.historyMenuIndex < historyMenuItems-1 {
			m.historyMenuIndex++
		}
		return m, nil
	case "enter":
		switch m.historyMenuIndex {
		case 0:
			m.historyMenuIndex = 0
			m.state = settingsStateSearchHistory
		case 1:
			m.historyMenuIndex = 0
			m.state = settingsStatePlayHistory
		case 2:
			m.state = settingsStateMenu
		}
		return m, nil
	}
	return m, nil
}

func (m SettingsModel) updateSearchHistory(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "5":
		m.state = settingsStateHistory
		return m, nil
	case "0":
		return m, func() tea.Msg { return backToMainMsg{} }
	case "1": // Increase
		newSize := m.searchHistory.MaxSize + 5
		store := storage.NewStorage(m.favoritePath)
		if err := store.UpdateHistorySize(context.Background(), newSize); err == nil {
			m.searchHistory.MaxSize = newSize
			m.message = fmt.Sprintf("✓ History size increased to %d", newSize)
			m.messageIsSuccess = true
			m.messageTime = 3
		} else {
			m.message = fmt.Sprintf("✗ Failed: %v", err)
			m.messageIsSuccess = false
			m.messageTime = 3
		}
		return m, tickEverySecond()
	case "2": // Decrease
		newSize := m.searchHistory.MaxSize - 5
		if newSize < 5 {
			newSize = 5
		}
		store := storage.NewStorage(m.favoritePath)
		if err := store.UpdateHistorySize(context.Background(), newSize); err == nil {
			m.searchHistory.MaxSize = newSize
			m.message = fmt.Sprintf("✓ History size decreased to %d", newSize)
			m.messageIsSuccess = true
			m.messageTime = 3
		} else {
			m.message = fmt.Sprintf("✗ Failed: %v", err)
			m.messageIsSuccess = false
			m.messageTime = 3
		}
		return m, tickEverySecond()
	case "3": // Reset to default
		store := storage.NewStorage(m.favoritePath)
		if err := store.UpdateHistorySize(context.Background(), storage.DefaultMaxHistorySize); err == nil {
			m.searchHistory.MaxSize = storage.DefaultMaxHistorySize
			m.message = fmt.Sprintf("✓ History size reset to %d", storage.DefaultMaxHistorySize)
			m.messageIsSuccess = true
			m.messageTime = 3
		} else {
			m.message = fmt.Sprintf("✗ Failed: %v", err)
			m.messageIsSuccess = false
			m.messageTime = 3
		}
		return m, tickEverySecond()
	case "4": // Clear history
		store := storage.NewStorage(m.favoritePath)
		if err := store.ClearSearchHistory(context.Background()); err == nil {
			m.searchHistory.SearchItems = []storage.SearchHistoryItem{}
			m.searchHistory.LuckyQueries = []string{}
			m.message = "✓ Search history cleared"
			m.messageIsSuccess = true
			m.messageTime = 3
		} else {
			m.message = fmt.Sprintf("✗ Failed: %v", err)
			m.messageIsSuccess = false
			m.messageTime = 3
		}
		return m, tickEverySecond()
	case "enter":
		idx := m.historyMenuList.Index()
		switch idx {
		case 0:
			return m.updateSearchHistory(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'1'}})
		case 1:
			return m.updateSearchHistory(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'2'}})
		case 2:
			return m.updateSearchHistory(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'3'}})
		case 3:
			return m.updateSearchHistory(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'4'}})
		case 4:
			m.state = settingsStateHistory
			return m, nil
		}
	}

	// Handle menu navigation
	var cmd tea.Cmd
	m.historyMenuList, cmd = m.historyMenuList.Update(msg)
	return m, cmd
}

// View renders the settings screen
func (m SettingsModel) View() string {
	switch m.state {
	case settingsStateMenu:
		return m.viewMenu()
	case settingsStateTheme:
		return m.viewTheme()
	case settingsStateHistory:
		return m.viewHistoryMenu()
	case settingsStateSearchHistory:
		return m.viewSearchHistory()
	case settingsStatePlayHistory:
		return m.viewPlayHistory()
	case settingsStatePlayOptions:
		return m.viewPlayOptions()
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
		Help:    "↑↓/jk: Navigate • Enter: Select • 1-8: Shortcut • Esc/0: Back • Ctrl+C: Quit",
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

func (m SettingsModel) viewHistoryMenu() string {
	var content strings.Builder

	// Stats line
	searchCount := 0
	if m.searchHistory != nil {
		searchCount = len(m.searchHistory.SearchItems)
	}
	playCount := 0
	if m.metadataManager != nil {
		playCount = m.metadataManager.GetTotalStations()
	}

	content.WriteString(stationFieldStyle().Render("Search history items: "))
	content.WriteString(highlightStyle().Render(fmt.Sprintf("%d", searchCount)))
	content.WriteString("\n")
	content.WriteString(stationFieldStyle().Render("Play history tracked:  "))
	content.WriteString(highlightStyle().Render(fmt.Sprintf("%d stations", playCount)))
	content.WriteString("\n\n")

	items := []string{
		"1. Search History   Manage search query history",
		"2. Play History     Recently Played display settings",
		"3. Back",
	}
	for i, item := range items {
		if i == m.historyMenuIndex {
			content.WriteString(selectedItemStyle().Render("> " + item))
		} else {
			content.WriteString(normalItemStyle().Render("  " + item))
		}
		content.WriteString("\n")
	}

	if m.message != "" {
		content.WriteString("\n")
		if m.messageIsSuccess {
			content.WriteString(successStyle().Render(m.message))
		} else {
			content.WriteString(errorStyle().Render(m.message))
		}
	}

	return RenderPageWithBottomHelp(PageLayout{
		Title:   "⚙️  Settings > History",
		Content: content.String(),
		Help:    "↑↓/jk: Navigate • Enter: Select • 1-3: Shortcut • Esc: Back • 0: Main Menu • Ctrl+C: Quit",
	}, m.height)
}

func (m SettingsModel) viewSearchHistory() string {
	var content strings.Builder

	// Current history size
	content.WriteString(stationFieldStyle().Render("Current History Size: "))
	content.WriteString(highlightStyle().Render(fmt.Sprintf("%d searches", m.searchHistory.MaxSize)))
	content.WriteString("\n")
	content.WriteString(helpStyle().Render("(Number of recent searches to keep)"))
	content.WriteString("\n\n")

	// Calculate new sizes for display
	newSizeInc := m.searchHistory.MaxSize + 5
	newSizeDec := m.searchHistory.MaxSize - 5
	if newSizeDec < 5 {
		newSizeDec = 5
	}

	// Update menu item descriptions with current calculations
	content.WriteString(m.historyMenuList.View())
	content.WriteString("\n")
	content.WriteString(helpStyle().Render(fmt.Sprintf("  1: Will become %d • 2: Will become %d • 3: Will become 10", newSizeInc, newSizeDec)))

	// Stats
	content.WriteString("\n\n")
	content.WriteString(helpStyle().Render("─────────────────────────────────────────"))
	content.WriteString("\n\n")
	content.WriteString(stationFieldStyle().Render("Current Stats:"))
	content.WriteString("\n")
	fmt.Fprintf(&content, "  Search history items: %d\n", len(m.searchHistory.SearchItems))
	fmt.Fprintf(&content, "  Lucky history items:  %d\n", len(m.searchHistory.LuckyQueries))

	if m.message != "" {
		content.WriteString("\n")
		if m.messageIsSuccess {
			content.WriteString(successStyle().Render(m.message))
		} else {
			content.WriteString(errorStyle().Render(m.message))
		}
	}

	return RenderPageWithBottomHelp(PageLayout{
		Title:   "⚙️  Settings > History > Search History",
		Content: content.String(),
		Help:    "↑↓/jk: Navigate • Enter/1-5: Select • Esc: Back • 0: Main Menu • Ctrl+C: Quit",
	}, m.height)
}

// updatePlayOptions handles keys in the Play Options sub-menu.
func (m SettingsModel) updatePlayOptions(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	cfg := m.playOptsCfg
	switch msg.String() {
	case "esc", "7":
		m.state = settingsStateMenu
		return m, nil
	case "0":
		return m, func() tea.Msg { return backToMainMsg{} }

	case "1": // Toggle ContinueOnNavigate
		cfg.ContinueOnNavigate = !cfg.ContinueOnNavigate
		if err := storage.SavePlayOptionsConfigToUnified(cfg); err == nil {
			m.playOptsCfg = cfg
			v := "Off"
			if cfg.ContinueOnNavigate {
				v = "On"
			}
			m.message = fmt.Sprintf("✓ Continue playback when changing pages: %s", v)
			m.messageIsSuccess = true
		} else {
			m.message = fmt.Sprintf("✗ Failed: %v", err)
			m.messageIsSuccess = false
		}
		m.messageTime = 3
		return m, tickEverySecond()

	case "2": // Increase DefaultVolume
		cfg.DefaultVolume += 5
		if cfg.DefaultVolume > 100 {
			cfg.DefaultVolume = 100
		}
		if err := storage.SavePlayOptionsConfigToUnified(cfg); err == nil {
			m.playOptsCfg = cfg
			m.message = fmt.Sprintf("✓ Default volume: %d%%", cfg.DefaultVolume)
			m.messageIsSuccess = true
		} else {
			m.message = fmt.Sprintf("✗ Failed: %v", err)
			m.messageIsSuccess = false
		}
		m.messageTime = 3
		return m, tickEverySecond()

	case "3": // Decrease DefaultVolume
		cfg.DefaultVolume -= 5
		if cfg.DefaultVolume < 0 {
			cfg.DefaultVolume = 0
		}
		if err := storage.SavePlayOptionsConfigToUnified(cfg); err == nil {
			m.playOptsCfg = cfg
			m.message = fmt.Sprintf("✓ Default volume: %d%%", cfg.DefaultVolume)
			m.messageIsSuccess = true
		} else {
			m.message = fmt.Sprintf("✗ Failed: %v", err)
			m.messageIsSuccess = false
		}
		m.messageTime = 3
		return m, tickEverySecond()

	case "4": // Toggle ConfirmStop
		cfg.ConfirmStop = !cfg.ConfirmStop
		if err := storage.SavePlayOptionsConfigToUnified(cfg); err == nil {
			m.playOptsCfg = cfg
			v := "Off"
			if cfg.ConfirmStop {
				v = "On"
			}
			m.message = fmt.Sprintf("✓ Confirm before stopping: %s", v)
			m.messageIsSuccess = true
		} else {
			m.message = fmt.Sprintf("✗ Failed: %v", err)
			m.messageIsSuccess = false
		}
		m.messageTime = 3
		return m, tickEverySecond()

	case "5": // Toggle ShowMetadata
		cfg.ShowMetadata = !cfg.ShowMetadata
		if err := storage.SavePlayOptionsConfigToUnified(cfg); err == nil {
			m.playOptsCfg = cfg
			v := "Off"
			if cfg.ShowMetadata {
				v = "On"
			}
			m.message = fmt.Sprintf("✓ Show station metadata: %s", v)
			m.messageIsSuccess = true
		} else {
			m.message = fmt.Sprintf("✗ Failed: %v", err)
			m.messageIsSuccess = false
		}
		m.messageTime = 3
		return m, tickEverySecond()

	case "6": // Reset to Defaults
		cfg = config.DefaultPlayOptionsConfig()
		if err := storage.SavePlayOptionsConfigToUnified(cfg); err == nil {
			m.playOptsCfg = cfg
			m.message = "✓ Play Options reset to defaults"
			m.messageIsSuccess = true
		} else {
			m.message = fmt.Sprintf("✗ Failed: %v", err)
			m.messageIsSuccess = false
		}
		m.messageTime = 3
		return m, tickEverySecond()
	}
	return m, nil
}

// viewPlayOptions renders the Play Options settings screen.
func (m SettingsModel) viewPlayOptions() string {
	var content strings.Builder

	boolStr := func(b bool) string {
		if b {
			return "[On] "
		}
		return "[Off]"
	}

	continueStr := boolStr(m.playOptsCfg.ContinueOnNavigate)
	confirmStr := boolStr(m.playOptsCfg.ConfirmStop)
	metaStr := boolStr(m.playOptsCfg.ShowMetadata)

	content.WriteString(normalItemStyle().Render(fmt.Sprintf("  1. Continue playback when changing pages   %s", continueStr)))
	content.WriteString("\n")
	content.WriteString(helpStyle().Render("      Keep station playing while browsing other screens."))
	content.WriteString("\n\n")

	content.WriteString(normalItemStyle().Render(fmt.Sprintf("  2. Default volume (+5)                    → %d%%", min(m.playOptsCfg.DefaultVolume+5, 100))))
	content.WriteString("\n")
	content.WriteString(normalItemStyle().Render(fmt.Sprintf("  3. Default volume (-5)                    → %d%%", max(m.playOptsCfg.DefaultVolume-5, 0))))
	content.WriteString("\n")
	content.WriteString(helpStyle().Render(fmt.Sprintf("      Current default volume: %d%%", m.playOptsCfg.DefaultVolume)))
	content.WriteString("\n\n")

	content.WriteString(normalItemStyle().Render(fmt.Sprintf("  4. Confirm before stopping playback        %s", confirmStr)))
	content.WriteString("\n")
	content.WriteString(helpStyle().Render("      Show a prompt when pressing Esc/0 while a station is playing."))
	content.WriteString("\n\n")

	content.WriteString(normalItemStyle().Render(fmt.Sprintf("  5. Show station metadata while playing     %s", metaStr)))
	content.WriteString("\n")
	content.WriteString(helpStyle().Render("      Display bitrate, codec, country, and tags in the play screen."))
	content.WriteString("\n\n")

	content.WriteString(normalItemStyle().Render("  6. Reset to Defaults"))
	content.WriteString("\n")
	content.WriteString(normalItemStyle().Render("  7. Back"))
	content.WriteString("\n")

	if m.message != "" {
		content.WriteString("\n")
		if m.messageIsSuccess {
			content.WriteString(successStyle().Render(m.message))
		} else {
			content.WriteString(errorStyle().Render(m.message))
		}
	}

	return RenderPageWithBottomHelp(PageLayout{
		Title:   "⚙️  Settings > Play Options",
		Content: content.String(),
		Help:    "1-7: Select • Esc/7: Back • 0: Main Menu • Ctrl+C: Quit",
	}, m.height)
}

// updatePlayHistory handles keys in the Play History settings sub-menu.
func (m SettingsModel) updatePlayHistory(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	cfg := m.playHistoryCfg
	switch msg.String() {
	case "esc", "7":
		m.state = settingsStateHistory
		return m, nil
	case "0":
		return m, func() tea.Msg { return backToMainMsg{} }
	case "1": // Toggle Show
		cfg.Enabled = !cfg.Enabled
		if err := storage.SavePlayHistoryConfigToUnified(cfg); err == nil {
			m.playHistoryCfg = cfg
			showStr := "No"
			if cfg.Enabled {
				showStr = "Yes"
			}
			m.message = fmt.Sprintf("✓ Show Recently Played: %s", showStr)
			m.messageIsSuccess = true
		} else {
			m.message = fmt.Sprintf("✗ Failed: %v", err)
			m.messageIsSuccess = false
		}
		m.messageTime = 3
		return m, tickEverySecond()
	case "2": // Increase size
		oldSize := cfg.Size
		cfg.Size++
		_ = cfg.Validate() // Clamps to [1, 20]
		if cfg.Size == oldSize {
			m.message = fmt.Sprintf("ℹ Size already at maximum (%d)", cfg.Size)
			m.messageIsSuccess = true
		} else if err := storage.SavePlayHistoryConfigToUnified(cfg); err == nil {
			m.playHistoryCfg = cfg
			m.message = fmt.Sprintf("✓ Size increased to %d", cfg.Size)
			m.messageIsSuccess = true
		} else {
			m.message = fmt.Sprintf("✗ Failed: %v", err)
			m.messageIsSuccess = false
		}
		m.messageTime = 3
		return m, tickEverySecond()
	case "3": // Decrease history size
		oldSize := cfg.Size
		cfg.Size--
		_ = cfg.Validate() // Clamps to [1, 20]
		if cfg.Size == oldSize {
			m.message = fmt.Sprintf("ℹ Size already at minimum (%d)", cfg.Size)
			m.messageIsSuccess = true
		} else if err := storage.SavePlayHistoryConfigToUnified(cfg); err == nil {
			m.playHistoryCfg = cfg
			m.message = fmt.Sprintf("✓ Size decreased to %d", cfg.Size)
			m.messageIsSuccess = true
		} else {
			m.message = fmt.Sprintf("✗ Failed: %v", err)
			m.messageIsSuccess = false
		}
		m.messageTime = 3
		return m, tickEverySecond()
	case "4": // Increase display rows
		oldRows := cfg.DisplayRows
		cfg.DisplayRows++
		_ = cfg.Validate() // Clamps to [0, 10]
		if cfg.DisplayRows == oldRows {
			m.message = fmt.Sprintf("ℹ Display rows already at maximum (%d)", cfg.DisplayRows)
			m.messageIsSuccess = true
		} else if err := storage.SavePlayHistoryConfigToUnified(cfg); err == nil {
			m.playHistoryCfg = cfg
			m.message = fmt.Sprintf("✓ Display rows set to %d", cfg.DisplayRows)
			m.messageIsSuccess = true
		} else {
			m.message = fmt.Sprintf("✗ Failed: %v", err)
			m.messageIsSuccess = false
		}
		m.messageTime = 3
		return m, tickEverySecond()
	case "5": // Decrease display rows
		oldRows := cfg.DisplayRows
		cfg.DisplayRows--
		if cfg.DisplayRows < 0 {
			cfg.DisplayRows = 0
		}
		if cfg.DisplayRows == oldRows {
			m.message = "ℹ Display rows already at minimum (0 = auto)"
			m.messageIsSuccess = true
		} else if err := storage.SavePlayHistoryConfigToUnified(cfg); err == nil {
			m.playHistoryCfg = cfg
			if cfg.DisplayRows == 0 {
				m.message = "✓ Display rows set to auto (fills available space)"
			} else {
				m.message = fmt.Sprintf("✓ Display rows set to %d", cfg.DisplayRows)
			}
			m.messageIsSuccess = true
		} else {
			m.message = fmt.Sprintf("✗ Failed: %v", err)
			m.messageIsSuccess = false
		}
		m.messageTime = 3
		return m, tickEverySecond()
	case "6": // Reset All Play Stats
		// ClearAll() wipes the entire metadata store (play counts, last played,
		// duration) — it is intentionally labelled "Reset All Play Stats" in the
		// UI so users understand it also resets Most Played, not just Recently Played.
		if m.metadataManager != nil {
			if err := m.metadataManager.ClearAll(); err == nil {
				m.message = "✓ All play stats reset"
				m.messageIsSuccess = true
			} else {
				m.message = fmt.Sprintf("✗ Failed: %v", err)
				m.messageIsSuccess = false
			}
		} else {
			m.message = "✗ Metadata manager not available"
			m.messageIsSuccess = false
		}
		m.messageTime = 3
		return m, tickEverySecond()
	}
	return m, nil
}

// viewPlayHistory renders the Play History settings screen.
func (m SettingsModel) viewPlayHistory() string {
	var content strings.Builder

	showStr := "No"
	if m.playHistoryCfg.Enabled {
		showStr = "Yes"
	}
	sizeUp := m.playHistoryCfg.Size + 1
	if sizeUp > 20 {
		sizeUp = 20
	}
	sizeDown := m.playHistoryCfg.Size - 1
	if sizeDown < 1 {
		sizeDown = 1
	}

	content.WriteString(stationFieldStyle().Render("  Configured Size:       "))
	content.WriteString(highlightStyle().Render(fmt.Sprintf("%d", m.playHistoryCfg.Size)))
	content.WriteString(helpStyle().Render(" stations stored"))
	content.WriteString("\n")
	if m.metadataManager != nil {
		actual := m.metadataManager.GetRecentlyPlayedCount(m.playHistoryCfg.Size)
		content.WriteString(stationFieldStyle().Render("  Currently in history:  "))
		content.WriteString(highlightStyle().Render(fmt.Sprintf("%d", actual)))
		content.WriteString("\n")
	}
	content.WriteString("\n")

	displayRows := m.playHistoryCfg.DisplayRows
	displayRowsStr := "auto"
	if displayRows > 0 {
		displayRowsStr = fmt.Sprintf("%d rows", displayRows)
	}
	displayRowsUp := displayRows + 1
	if displayRowsUp > 10 {
		displayRowsUp = 10
	}
	displayRowsDown := displayRows - 1
	if displayRowsDown < 0 {
		displayRowsDown = 0
	}
	displayRowsUpStr := fmt.Sprintf("%d rows", displayRowsUp)
	if displayRowsUp == 0 {
		displayRowsUpStr = "auto"
	}
	displayRowsDownStr := fmt.Sprintf("%d rows", displayRowsDown)
	if displayRowsDown == 0 {
		displayRowsDownStr = "auto"
	}

	content.WriteString(normalItemStyle().Render(fmt.Sprintf("  1. Toggle Show                (currently: %s)", showStr)))
	content.WriteString("\n")
	content.WriteString(normalItemStyle().Render(fmt.Sprintf("  2. Increase History Size (+1)  → %d", sizeUp)))
	content.WriteString("\n")
	content.WriteString(normalItemStyle().Render(fmt.Sprintf("  3. Decrease History Size (-1)  → %d", sizeDown)))
	content.WriteString("\n")
	content.WriteString(normalItemStyle().Render(fmt.Sprintf("  4. Increase Display Rows (+1)  → %s", displayRowsUpStr)))
	content.WriteString("\n")
	content.WriteString(normalItemStyle().Render(fmt.Sprintf("  5. Decrease Display Rows (-1)  → %s", displayRowsDownStr)))
	content.WriteString("\n")
	content.WriteString(helpStyle().Render(fmt.Sprintf("      Display rows: %s (0 = fill available space)", displayRowsStr)))
	content.WriteString("\n")
	content.WriteString(normalItemStyle().Render("  6. Reset All Play Stats"))
	content.WriteString("\n")
	content.WriteString(helpStyle().Render("      (clears play counts, Most Played, and Recently Played)"))
	content.WriteString("\n")
	content.WriteString(normalItemStyle().Render("  7. Back"))
	content.WriteString("\n")

	if m.metadataManager != nil {
		total := m.metadataManager.GetTotalStations()
		content.WriteString("\n")
		content.WriteString(helpStyle().Render(fmt.Sprintf("  Total tracked stations: %d", total)))
	}

	if m.message != "" {
		content.WriteString("\n")
		if m.messageIsSuccess {
			content.WriteString(successStyle().Render(m.message))
		} else {
			content.WriteString(errorStyle().Render(m.message))
		}
	}

	return RenderPageWithBottomHelp(PageLayout{
		Title:   "⚙️  Settings > History > Play History",
		Content: content.String(),
		Help:    "1-7: Select • Esc/7: Back • 0: Main Menu • Ctrl+C: Quit",
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
			content.WriteString(helpStyle().Render("─────────────────────────────────────────"))
			content.WriteString("\n\n")

			// Show installation method specific update instructions
			content.WriteString(stationFieldStyle().Render("Detected installation method: "))
			content.WriteString(highlightStyle().Render(m.installInfo.Description))
			content.WriteString("\n\n")

			if m.installInfo.UpdateCommand != "" {
				content.WriteString(stationValueStyle().Render("To update, run:"))
				content.WriteString("\n")
				content.WriteString(highlightStyle().Render("  " + m.installInfo.UpdateCommand))
				content.WriteString("\n\n")
			} else {
				// Manual/Unknown installation
				content.WriteString(stationValueStyle().Render("Visit the release page to download the latest version:"))
				content.WriteString("\n")
				content.WriteString(highlightStyle().Render("  " + api.ReleasePageURL))
				content.WriteString("\n\n")
			}
		} else {
			content.WriteString(successStyle().Render("✓ You're up to date!"))
			content.WriteString("\n\n")
			content.WriteString(stationFieldStyle().Render("Latest version: "))
			content.WriteString(highlightStyle().Render(m.latestVersion))
			content.WriteString("\n\n")
			content.WriteString(stationFieldStyle().Render("Installation method: "))
			content.WriteString(stationValueStyle().Render(m.installInfo.Description))
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
