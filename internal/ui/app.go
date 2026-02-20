package ui

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/shinokada/tera/v2/internal/api"
	"github.com/shinokada/tera/v2/internal/blocklist"
	"github.com/shinokada/tera/v2/internal/player"
	"github.com/shinokada/tera/v2/internal/storage"
	"github.com/shinokada/tera/v2/internal/ui/components"
)

type Screen int

const (
	screenMainMenu Screen = iota
	screenPlay
	screenSearch
	screenList
	screenLucky
	screenGist
	screenSettings
	screenShuffleSettings
	screenConnectionSettings
	screenAppearanceSettings
	screenBlocklist
)

// Main menu configuration
const mainMenuItemCount = 7

type App struct {
	screen                   Screen
	width                    int
	height                   int
	mainMenuList             list.Model
	playScreen               PlayModel
	searchScreen             SearchModel
	listManagementScreen     ListManagementModel
	luckyScreen              LuckyModel
	gistScreen               GistModel
	settingsScreen           SettingsModel
	shuffleSettingsScreen    ShuffleSettingsModel
	connectionSettingsScreen ConnectionSettingsModel
	appearanceSettingsScreen AppearanceSettingsModel
	blocklistScreen          BlocklistModel
	apiClient                *api.Client
	blocklistManager         *blocklist.Manager
	favoritePath             string
	quickFavorites           []api.Station
	quickFavPlayer           *player.MPVPlayer
	playingFromMain          bool
	playingStation           *api.Station
	numberBuffer             string               // Buffer for multi-digit number input
	unifiedMenuIndex         int                  // Unified index for navigating both menu and favorites
	helpModel                components.HelpModel // Help overlay
	volumeDisplay            string               // Temporary volume display message
	volumeDisplayFrames      int                  // Countdown for volume display
	// Update checking
	latestVersion   string // Latest version from GitHub
	updateAvailable bool   // True if a newer version exists
	updateChecked   bool   // True if version check completed
}

// navigateMsg is sent when changing screens
type navigateMsg struct {
	screen Screen
}

func NewApp() App {
	// Get favorite path from environment or use default
	favPath := os.Getenv("TERA_FAVORITE_PATH")
	if favPath == "" {
		configDir, _ := os.UserConfigDir()
		favPath = filepath.Join(configDir, "tera", "favorites")
	}

	// Ensure favorites directory exists
	if err := os.MkdirAll(favPath, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to create favorites directory: %v\n", err)
	}

	// Initialize blocklist manager
	configDir, _ := os.UserConfigDir()
	blocklistPath := filepath.Join(configDir, "tera", "blocklist.json")
	blocklistMgr := blocklist.NewManager(blocklistPath)
	if err := blocklistMgr.Load(context.Background()); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to load blocklist: %v\n", err)
	}

	app := App{
		screen:           screenMainMenu,
		favoritePath:     favPath,
		apiClient:        api.NewClient(),
		quickFavPlayer:   player.NewMPVPlayer(),
		helpModel:        components.NewHelpModel(components.CreateMainMenuHelp()),
		blocklistManager: blocklistMgr,
	}

	// Initialize header renderer
	InitializeHeaderRenderer()

	// Initialize main menu
	app.initMainMenu()

	// Ensure My-favorites.json exists at startup
	app.ensureMyFavorites()

	// Load quick favorites
	app.loadQuickFavorites()

	return app
}

func (a *App) initMainMenu() {
	items := []components.MenuItem{
		components.NewMenuItem("Play from Favorites", "", "1"),
		components.NewMenuItem("Search Stations", "", "2"),
		components.NewMenuItem("Manage Lists", "", "3"),
		components.NewMenuItem("Block List", "", "4"),
		components.NewMenuItem("I Feel Lucky", "", "5"),
		components.NewMenuItem("Gist Management", "", "6"),
		components.NewMenuItem("Settings", "", "7"),
	}

	// Height will be auto-adjusted by CreateMenu to fit all items
	// Title is empty as TERA header is added by wrapPageWithHeader
	a.mainMenuList = components.CreateMenu(items, "", 50, 20)
}

// ensureMyFavorites ensures My-favorites.json exists at startup
func (a *App) ensureMyFavorites() {
	store := storage.NewStorage(a.favoritePath)
	if _, err := store.LoadList(context.Background(), "My-favorites"); err != nil {
		// Only create if file doesn't exist, not on other errors
		if !os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "Warning: failed to load My-favorites: %v\n", err)
			return
		}
		// Create empty My-favorites list
		emptyList := &storage.FavoritesList{
			Name:     "My-favorites",
			Stations: []api.Station{},
		}
		if err := store.SaveList(context.Background(), emptyList); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to create My-favorites: %v\n", err)
		}
	}
}

// loadQuickFavorites loads stations from My-favorites.json for quick play
func (a *App) loadQuickFavorites() {
	store := storage.NewStorage(a.favoritePath)
	list, err := store.LoadList(context.Background(), "My-favorites")
	if err != nil {
		// It's OK if My-favorites doesn't exist or is empty
		a.quickFavorites = []api.Station{}
		return
	}
	a.quickFavorites = list.Stations
}

func (a App) Init() tea.Cmd {
	// Check for updates in the background on startup
	return checkForUpdates()
}

func (a App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case versionCheckMsg:
		// Handle version check result (from startup or settings)
		a.updateChecked = true
		if msg.err == nil {
			a.latestVersion = msg.latestVersion
			a.updateAvailable = api.IsNewerVersion(Version, msg.latestVersion)
		}
		return a, nil

	case tea.KeyMsg:
		// Global key bindings
		switch msg.String() {
		case "ctrl+c":
			// Stop any playing stations before quitting
			if a.quickFavPlayer != nil {
				_ = a.quickFavPlayer.Stop()
			}
			if a.screen == screenPlay && a.playScreen.player != nil {
				_ = a.playScreen.player.Stop()
			} else if a.screen == screenSearch && a.searchScreen.player != nil {
				_ = a.searchScreen.player.Stop()
			} else if a.screen == screenLucky && a.luckyScreen.player != nil {
				_ = a.luckyScreen.player.Stop()
			}
			return a, tea.Quit
		}

	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height

		// Update help model size
		a.helpModel.SetSize(msg.Width, msg.Height)

		// Update main menu size
		if a.screen == screenMainMenu {
			h, v := docStyle().GetFrameSize()
			// Ensure enough height for all menu items (6 items)
			// Reserve 6 lines for TERA header (3 lines) + spacing + help line
			minHeight := 12
			menuHeight := msg.Height - v - 6
			if menuHeight < minHeight {
				menuHeight = minHeight
			}
			a.mainMenuList.SetSize(msg.Width-h, menuHeight)
		}
		// Don't return here - let it fall through to forward to current screen

	case navigateMsg:
		a.screen = msg.screen

		// Initialize screen-specific models with current dimensions
		switch msg.screen {
		case screenPlay:
			a.playScreen = NewPlayModel(a.favoritePath, a.blocklistManager)
			// Set dimensions immediately if we have them
			if a.width > 0 && a.height > 0 {
				a.playScreen.width = a.width
				a.playScreen.height = a.height
			}
			return a, a.playScreen.Init()
		case screenSearch:
			a.searchScreen = NewSearchModel(a.apiClient, a.favoritePath, a.blocklistManager)
			// Set dimensions immediately if we have them
			if a.width > 0 && a.height > 0 {
				a.searchScreen.width = a.width
				a.searchScreen.height = a.height
			}
			return a, a.searchScreen.Init()
		case screenList:
			a.listManagementScreen = NewListManagementModel(a.favoritePath)
			// Set dimensions immediately if we have them
			if a.width > 0 && a.height > 0 {
				a.listManagementScreen.width = a.width
				a.listManagementScreen.height = a.height
			}
			return a, a.listManagementScreen.Init()
		case screenLucky:
			a.luckyScreen = NewLuckyModel(a.apiClient, a.favoritePath, a.blocklistManager)
			// Set dimensions immediately if we have them
			if a.width > 0 && a.height > 0 {
				a.luckyScreen.width = a.width
				a.luckyScreen.height = a.height
			}
			return a, a.luckyScreen.Init()
		case screenGist:
			a.gistScreen = NewGistModel(a.favoritePath)
			// Set dimensions immediately if we have them
			if a.width > 0 && a.height > 0 {
				a.gistScreen.width = a.width
				a.gistScreen.height = a.height
			}
			return a, a.gistScreen.Init()
		case screenSettings:
			a.settingsScreen = NewSettingsModel(a.favoritePath)
			// Set dimensions immediately if we have them
			if a.width > 0 && a.height > 0 {
				a.settingsScreen.width = a.width
				a.settingsScreen.height = a.height
			}
			// Pass update info from App to Settings
			if a.updateChecked {
				a.settingsScreen.updateChecked = true
				a.settingsScreen.latestVersion = a.latestVersion
				a.settingsScreen.updateAvailable = a.updateAvailable
			}
			return a, a.settingsScreen.Init()
		case screenShuffleSettings:
			a.shuffleSettingsScreen = NewShuffleSettingsModel()
			// Set dimensions immediately if we have them
			if a.width > 0 && a.height > 0 {
				a.shuffleSettingsScreen.width = a.width
				a.shuffleSettingsScreen.height = a.height
			}
			return a, a.shuffleSettingsScreen.Init()
		case screenConnectionSettings:
			a.connectionSettingsScreen = NewConnectionSettingsModel()
			// Set dimensions immediately if we have them
			if a.width > 0 && a.height > 0 {
				a.connectionSettingsScreen.width = a.width
				a.connectionSettingsScreen.height = a.height
			}
			return a, a.connectionSettingsScreen.Init()
		case screenAppearanceSettings:
			a.appearanceSettingsScreen = NewAppearanceSettingsModel()
			// Set dimensions immediately if we have them
			if a.width > 0 && a.height > 0 {
				a.appearanceSettingsScreen.width = a.width
				a.appearanceSettingsScreen.height = a.height
			}
			return a, a.appearanceSettingsScreen.Init()
		case screenBlocklist:
			a.blocklistScreen = NewBlocklistModel(a.blocklistManager)
			// Set dimensions immediately if we have them
			if a.width > 0 && a.height > 0 {
				a.blocklistScreen.width = a.width
				a.blocklistScreen.height = a.height
			}
			return a, a.blocklistScreen.Init()
		case screenMainMenu:
			// Return to main menu and reload favorites
			a.loadQuickFavorites()
			a.unifiedMenuIndex = 0
			a.numberBuffer = ""
			return a, nil
		}
		return a, nil

	case backToMainMsg:
		// Handle back to main menu from any screen
		a.screen = screenMainMenu
		// Reload quick favorites in case they were updated
		a.loadQuickFavorites()
		a.unifiedMenuIndex = 0
		a.numberBuffer = ""
		return a, nil
	}

	// Route to current screen
	var cmd tea.Cmd
	switch a.screen {
	case screenMainMenu:
		return a.updateMainMenu(msg)
	case screenPlay:
		var m tea.Model
		m, cmd = a.playScreen.Update(msg)
		a.playScreen = m.(PlayModel)

		// Check if we should return to main menu
		if _, ok := msg.(backToMainMsg); ok {
			a.screen = screenMainMenu
		}
		return a, cmd
	case screenSearch:
		var m tea.Model
		m, cmd = a.searchScreen.Update(msg)
		a.searchScreen = m.(SearchModel)

		// Check if we should return to main menu
		if _, ok := msg.(backToMainMsg); ok {
			a.screen = screenMainMenu
		}
		return a, cmd
	case screenList:
		var m tea.Model
		m, cmd = a.listManagementScreen.Update(msg)
		a.listManagementScreen = m.(ListManagementModel)

		// Check if we should return to main menu
		if _, ok := msg.(backToMainMsg); ok {
			a.screen = screenMainMenu
		}
		return a, cmd
	case screenLucky:
		var m tea.Model
		m, cmd = a.luckyScreen.Update(msg)
		a.luckyScreen = m.(LuckyModel)

		// Check if we should return to main menu
		if _, ok := msg.(backToMainMsg); ok {
			a.screen = screenMainMenu
		}
		return a, cmd
	case screenGist:
		var m tea.Model
		m, cmd = a.gistScreen.Update(msg)
		if model, ok := m.(GistModel); ok {
			a.gistScreen = model
		}

		// Check if we should return to main menu or if the model itself requested quit/back
		// The GistModel handles its own state (quitting -> Back)
		// But in Update() of GistModel, it returns tea.Quit if it wants to quit app,
		// or sets 'quitting' or state to menu?
		// Logic in GistModel: if ESC at menu, returns m, tea.Quit.
		// Wait, if GistModel returns Quit, the whole app quits.
		// We want it to go back to Main Menu.
		// In GistModel.Update:
		// if msg.String() == "esc" {
		//    if m.state == gistStateMenu { m.quitting = true; return m, tea.Quit } ...
		// }
		// We might need to adjust GistModel or handle it here.
		// If GistModel has a 'quitting' flag that we can check?
		if a.gistScreen.quitting {
			a.screen = screenMainMenu
			// Reset quitting so next time it's fresh?
			// But NewGistModel creates a fresh one every time we navigate to it.
		}
		return a, cmd
	case screenSettings:
		var m tea.Model
		m, cmd = a.settingsScreen.Update(msg)
		a.settingsScreen = m.(SettingsModel)

		// Check if we should return to main menu
		if _, ok := msg.(backToMainMsg); ok {
			a.screen = screenMainMenu
		}
		return a, cmd
	case screenShuffleSettings:
		var m tea.Model
		m, cmd = a.shuffleSettingsScreen.Update(msg)
		a.shuffleSettingsScreen = m.(ShuffleSettingsModel)

		// Check if we should return to main menu
		if _, ok := msg.(backToMainMsg); ok {
			a.screen = screenMainMenu
		}
		return a, cmd
	case screenConnectionSettings:
		var m tea.Model
		m, cmd = a.connectionSettingsScreen.Update(msg)
		a.connectionSettingsScreen = m.(ConnectionSettingsModel)

		// Check if we should return to main menu
		if _, ok := msg.(backToMainMsg); ok {
			a.screen = screenMainMenu
		}
		return a, cmd
	case screenAppearanceSettings:
		var m tea.Model
		m, cmd = a.appearanceSettingsScreen.Update(msg)
		a.appearanceSettingsScreen = m.(AppearanceSettingsModel)

		// Check if we should return to main menu
		if _, ok := msg.(backToMainMsg); ok {
			a.screen = screenMainMenu
		}
		return a, cmd
	case screenBlocklist:
		var m tea.Model
		m, cmd = a.blocklistScreen.Update(msg)
		a.blocklistScreen = m.(BlocklistModel)

		// Check if we should return to main menu
		if _, ok := msg.(backToMainMsg); ok {
			a.screen = screenMainMenu
		}
		return a, cmd
	}

	return a, nil
}

func (a App) updateMainMenu(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tickMsg:
		// Handle volume display countdown (only for positive values, not persistent -1)
		if a.volumeDisplayFrames > 0 {
			a.volumeDisplayFrames--
			if a.volumeDisplayFrames == 0 {
				a.volumeDisplay = ""
			}
			return a, tickEverySecond()
		}
		return a, nil

	case tea.KeyMsg:
		// If help is visible, let it handle the key
		if a.helpModel.IsVisible() {
			var cmd tea.Cmd
			a.helpModel, cmd = a.helpModel.Update(msg)
			return a, cmd
		}

		// Handle ? for help
		if msg.String() == "?" {
			a.helpModel.Show()
			return a, nil
		}

		// Handle volume controls when playing
		if a.playingFromMain && a.quickFavPlayer != nil {
			switch msg.String() {
			case " ":
				// Toggle pause/resume
				if err := a.quickFavPlayer.TogglePause(); err == nil {
					if a.quickFavPlayer.IsPaused() {
						// Paused - show persistent message
						a.volumeDisplay = "⏸ Paused - Press Space to resume"
						a.volumeDisplayFrames = -1 // Persistent (negative means persistent)
					} else {
						// Resumed - show temporary message
						a.volumeDisplay = "▶ Resumed"
						startTick := a.volumeDisplayFrames <= 0
						a.volumeDisplayFrames = 2
						if startTick {
							return a, tickEverySecond()
						}
					}
				}
				return a, nil
			case "/":
				// Decrease volume
				newVol := a.quickFavPlayer.DecreaseVolume(5)
				a.volumeDisplay = fmt.Sprintf("Volume: %d%%", newVol)
				startTick := a.volumeDisplayFrames == 0
				a.volumeDisplayFrames = 2 // Show for 2 seconds
				// Update station volume if we have one
				if a.playingStation != nil && newVol >= 0 {
					a.playingStation.SetVolume(newVol)
					// Save updated volume to favorites
					a.saveStationVolume(a.playingStation)
				}
				if startTick {
					return a, tickEverySecond()
				}
				return a, nil
			case "*":
				// Increase volume
				newVol := a.quickFavPlayer.IncreaseVolume(5)
				a.volumeDisplay = fmt.Sprintf("Volume: %d%%", newVol)
				startTick := a.volumeDisplayFrames == 0
				a.volumeDisplayFrames = 2 // Show for 2 seconds
				// Update station volume if we have one
				if a.playingStation != nil {
					a.playingStation.SetVolume(newVol)
					// Save updated volume to favorites
					a.saveStationVolume(a.playingStation)
				}
				if startTick {
					return a, tickEverySecond()
				}
				return a, nil
			case "m":
				// Toggle mute
				muted, vol := a.quickFavPlayer.ToggleMute()
				if muted {
					a.volumeDisplay = "Volume: Muted"
				} else {
					a.volumeDisplay = fmt.Sprintf("Volume: %d%%", vol)
				}
				startTick := a.volumeDisplayFrames == 0
				a.volumeDisplayFrames = 2 // Show for 2 seconds
				// Update station volume if we have one
				if a.playingStation != nil && !muted && vol >= 0 {
					a.playingStation.SetVolume(vol)
					// Save updated volume to favorites
					a.saveStationVolume(a.playingStation)
				}
				if startTick {
					return a, tickEverySecond()
				}
				return a, nil
			}
		}

		// Handle Escape to stop playing if playing from main menu
		if msg.String() == "esc" && a.playingFromMain {
			if a.quickFavPlayer != nil {
				_ = a.quickFavPlayer.Stop()
			}
			a.playingFromMain = false
			a.playingStation = nil
			a.numberBuffer = "" // Clear buffer on escape
			return a, nil
		}

		// Handle number input for menu and quick favorites
		key := msg.String()
		if len(key) == 1 && key >= "0" && key <= "9" {
			a.numberBuffer += key

			// For numbers >= 10, we need at least 2 digits
			// Allow up to 3 digits for larger lists (e.g., 100+)
			if len(a.numberBuffer) >= 2 {
				// Check if this could be a valid selection
				num := 0
				_, _ = fmt.Sscanf(a.numberBuffer, "%d", &num)

				// Calculate max valid number
				maxFavNum := 9 + len(a.quickFavorites) // 10-based indexing

				// If the number is valid for favorites, play it
				if num >= 10 && num <= maxFavNum {
					idx := num - 10
					a.numberBuffer = "" // Clear buffer
					return a.playQuickFavorite(idx)
				}

				// If number is too large or we have 3+ digits, clear and ignore
				if num > maxFavNum || len(a.numberBuffer) >= 3 {
					a.numberBuffer = ""
					return a, nil
				}
			}

			// Single digit - could be menu shortcut (1-6) or start of larger number
			return a, nil
		}

		// Handle Enter key to select filtered menu item or confirm number
		if msg.String() == "enter" {
			// Check if there's a buffered number to process
			if a.numberBuffer != "" {
				num := 0
				_, _ = fmt.Sscanf(a.numberBuffer, "%d", &num)
				a.numberBuffer = "" // Clear buffer

				// Numbers 1-6 are for menu items
				if num >= 1 && num <= mainMenuItemCount {
					a.unifiedMenuIndex = num - 1
					return a.executeMenuAction(num - 1)
				}
				// Numbers 10+ are for quick favorites
				if num >= 10 {
					idx := num - 10
					if idx < len(a.quickFavorites) {
						return a.playQuickFavorite(idx)
					}
				}
				return a, nil
			}
			// No buffered number, use current unified selection
			menuItemCount := mainMenuItemCount
			if a.unifiedMenuIndex < menuItemCount {
				return a.executeMenuAction(a.unifiedMenuIndex)
			} else {
				// It's a quick favorite
				favIndex := a.unifiedMenuIndex - menuItemCount
				if favIndex < len(a.quickFavorites) {
					return a.playQuickFavorite(favIndex)
				}
			}
			return a, nil
		}

		// Handle arrow keys for unified menu navigation
		menuItemCount := mainMenuItemCount
		favCount := len(a.quickFavorites)
		totalItems := menuItemCount + favCount

		switch msg.String() {
		case "up", "k":
			a.numberBuffer = "" // Clear buffer on navigation
			if a.unifiedMenuIndex > 0 {
				a.unifiedMenuIndex--
			}
			return a, nil
		case "down", "j":
			a.numberBuffer = "" // Clear buffer on navigation
			if a.unifiedMenuIndex < totalItems-1 {
				a.unifiedMenuIndex++
			}
			return a, nil
		}

		// Clear buffer on other keys
		a.numberBuffer = ""
		return a, nil

	default:
		return a, nil
	}
}

func (a App) executeMenuAction(index int) (tea.Model, tea.Cmd) {
	// Stop any currently playing quick favorite before navigating
	if a.playingFromMain && a.quickFavPlayer != nil {
		_ = a.quickFavPlayer.Stop()
	}
	a.playingFromMain = false
	a.playingStation = nil

	switch index {
	case 0: // Play from Favorites
		return a, func() tea.Msg {
			return navigateMsg{screen: screenPlay}
		}
	case 1: // Search Stations
		return a, func() tea.Msg {
			return navigateMsg{screen: screenSearch}
		}
	case 2: // Manage Lists
		return a, func() tea.Msg {
			return navigateMsg{screen: screenList}
		}
	case 3: // Block List
		return a, func() tea.Msg {
			return navigateMsg{screen: screenBlocklist}
		}
	case 4: // I Feel Lucky
		return a, func() tea.Msg {
			return navigateMsg{screen: screenLucky}
		}
	case 5: // Gist Management
		return a, func() tea.Msg {
			return navigateMsg{screen: screenGist}
		}
	case 6: // Settings
		return a, func() tea.Msg {
			return navigateMsg{screen: screenSettings}
		}
	}
	return a, nil
}

// playQuickFavorite plays a station from the quick favorites list
func (a App) playQuickFavorite(index int) (tea.Model, tea.Cmd) {
	if index >= len(a.quickFavorites) {
		return a, nil
	}

	station := a.quickFavorites[index]
	a.playingStation = &station
	a.playingFromMain = true

	// Stop any currently playing station
	if a.quickFavPlayer != nil {
		_ = a.quickFavPlayer.Stop()
	}

	// Start playback
	return a, func() tea.Msg {
		if a.quickFavPlayer != nil {
			if err := a.quickFavPlayer.Play(&station); err != nil {
				return playbackErrorMsg{err}
			}
		}
		return playbackStartedMsg{}
	}
}

func (a App) View() string {
	switch a.screen {
	case screenMainMenu:
		return a.viewMainMenu()
	case screenPlay:
		return a.playScreen.View()
	case screenSearch:
		return a.searchScreen.View()
	case screenList:
		return a.listManagementScreen.View()
	case screenLucky:
		return a.luckyScreen.View()
	case screenGist:
		return a.gistScreen.View()
	case screenSettings:
		return a.settingsScreen.View()
	case screenShuffleSettings:
		return a.shuffleSettingsScreen.View()
	case screenConnectionSettings:
		return a.connectionSettingsScreen.View()
	case screenAppearanceSettings:
		return a.appearanceSettingsScreen.View()
	case screenBlocklist:
		return a.blocklistScreen.View()
	}
	return "Unknown screen"
}

// saveStationVolume saves the updated volume for a station in the favorites list
func (a *App) saveStationVolume(station *api.Station) {
	if station == nil {
		return
	}

	store := storage.NewStorage(a.favoritePath)
	list, err := store.LoadList(context.Background(), "My-favorites")
	if err != nil {
		return
	}

	// Find and update the station
	for i := range list.Stations {
		if list.Stations[i].StationUUID == station.StationUUID {
			list.Stations[i].Volume = station.Volume
			break
		}
	}

	// Save the updated list
	_ = store.SaveList(context.Background(), list)

	// Reload quick favorites to reflect the change
	a.loadQuickFavorites()
}

func (a App) viewMainMenu() string {
	var content strings.Builder

	// Add "Choose an option:" with number buffer display
	content.WriteString(subtitleStyle().Render("Choose an option:"))
	if a.numberBuffer != "" {
		content.WriteString(" ")
		content.WriteString(highlightStyle().Render(a.numberBuffer + "_"))
	}
	content.WriteString("\n\n")

	// Display menu items with unified index
	menuItems := []struct {
		shortcut string
		title    string
	}{
		{"1", "Play from Favorites"},
		{"2", "Search Stations"},
		{"3", "Manage Lists"},
		{"4", "Block List"},
		{"5", "I Feel Lucky"},
		{"6", "Gist Management"},
		{"7", "Settings"},
	}

	for i, item := range menuItems {
		prefix := "  "
		if i == a.unifiedMenuIndex {
			prefix = "> "
			content.WriteString(selectedItemStyle().Render(fmt.Sprintf("%s%s. %s", prefix, item.shortcut, item.title)))
		} else {
			content.WriteString(normalItemStyle().Render(fmt.Sprintf("%s%s. %s", prefix, item.shortcut, item.title)))
		}
		content.WriteString("\n")
	}

	// Show currently playing station if playing from main menu
	if a.playingFromMain && a.playingStation != nil {
		content.WriteString("\n")
		content.WriteString(successStyle().Render("♫ Now Playing: "))
		content.WriteString(stationNameStyle().Render(a.playingStation.TrimName()))
		content.WriteString("\n")
	}

	// Add quick play favorites if available (also part of unified navigation)
	if len(a.quickFavorites) > 0 {
		content.WriteString("\n")
		content.WriteString(quickFavoritesStyle().Render("─── Quick Play Favorites ───"))
		content.WriteString("\n")

		menuItemCount := len(menuItems)
		for i, station := range a.quickFavorites {
			// Use numbers 10+ for quick favorites (no limit)
			shortcut := fmt.Sprintf("%d", 10+i)

			// Build station info line
			var stationInfo strings.Builder
			stationInfo.WriteString(station.TrimName())

			if station.Country != "" {
				stationInfo.WriteString(" • ")
				stationInfo.WriteString(station.Country)
			}
			if station.Codec != "" {
				stationInfo.WriteString(" • ")
				stationInfo.WriteString(station.Codec)
				if station.Bitrate > 0 {
					fmt.Fprintf(&stationInfo, " %dkbps", station.Bitrate)
				}
			}

			unifiedIdx := menuItemCount + i
			prefix := "  "
			if unifiedIdx == a.unifiedMenuIndex {
				prefix = "> "
			}

			// Build the line content
			lineContent := fmt.Sprintf("%s%s. %s", prefix, shortcut, stationInfo.String())

			// Highlight if this is the playing station
			if a.playingFromMain && a.playingStation != nil && a.playingStation.StationUUID == station.StationUUID {
				// Playing station - show with play icon
				playingLine := fmt.Sprintf("%s%s. ▶ %s", prefix, shortcut, stationInfo.String())
				if unifiedIdx == a.unifiedMenuIndex {
					content.WriteString(selectedItemStyle().Render(playingLine))
				} else {
					content.WriteString(normalItemStyle().Render(playingLine))
				}
			} else {
				// Normal station
				if unifiedIdx == a.unifiedMenuIndex {
					content.WriteString(selectedItemStyle().Render(lineContent))
				} else {
					content.WriteString(normalItemStyle().Render(lineContent))
				}
			}
			content.WriteString("\n")
		}
	}

	// Add volume display if visible
	if a.volumeDisplay != "" {
		content.WriteString("\n")
		content.WriteString(highlightStyle().Render(a.volumeDisplay))
		content.WriteString("\n")
	}

	// Build help text based on playing state
	var helpText string
	if a.playingFromMain {
		helpText = "↑↓/jk: Navigate • Enter: Select • /*: Volume • m: Mute • Esc: Stop • ?: Help"
	} else {
		helpText = "↑↓/jk: Navigate • Enter: Select • 1-7: Menu • 10+: Quick Play • ?: Help"
	}

	// Add update indicator if available (yellow)
	if a.updateAvailable {
		helpText += " • " + highlightStyle().Render("⬆ Update")
	}

	// Render the page
	page := RenderPageWithBottomHelp(PageLayout{
		Title:    "Main Menu & Quick Play",
		Subtitle: "",
		Content:  content.String(),
		Help:     helpText,
	}, a.height)

	// Overlay help if visible
	if a.helpModel.IsVisible() {
		return a.helpModel.View()
	}

	return page
}
