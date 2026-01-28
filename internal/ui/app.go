package ui

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/shinokada/tera/internal/api"
	"github.com/shinokada/tera/internal/player"
	"github.com/shinokada/tera/internal/storage"
	"github.com/shinokada/tera/internal/ui/components"
)

type Screen int

const (
	screenMainMenu Screen = iota
	screenPlay
	screenSearch
	screenList
	screenLucky
	screenGist
)

type App struct {
	screen               Screen
	width                int
	height               int
	mainMenuList         list.Model
	menuTextInput        textinput.Model
	playScreen           PlayModel
	searchScreen         SearchModel
	listManagementScreen ListManagementModel
	luckyScreen          LuckyModel
	gistScreen           GistModel
	apiClient            *api.Client
	favoritePath         string
	quickFavorites       []api.Station
	quickFavPlayer       *player.MPVPlayer
	playingFromMain      bool
	playingStation       *api.Station
}

// navigateMsg is sent when changing screens
type navigateMsg struct {
	screen Screen
}

func NewApp() App {
	// Get favorite path from environment or use default
	favPath := os.Getenv("TERA_FAVORITE_PATH")
	if favPath == "" {
		home, _ := os.UserHomeDir()
		favPath = filepath.Join(home, ".config", "tera", "favorites")
	}

	// Ensure favorites directory exists
	if err := os.MkdirAll(favPath, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to create favorites directory: %v\n", err)
	}

	// Initialize text input for menu filtering
	ti := textinput.New()
	ti.Placeholder = "Type to filter..."
	ti.CharLimit = 50
	ti.Width = 40
	ti.Focus()

	app := App{
		screen:         screenMainMenu,
		favoritePath:   favPath,
		apiClient:      api.NewClient(),
		menuTextInput:  ti,
		quickFavPlayer: player.NewMPVPlayer(),
	}

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
		components.NewMenuItem("I Feel Lucky", "", "4"),
		components.NewMenuItem("Gist Management", "(coming soon)", "5"),
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
	return textinput.Blink
}

func (a App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Global key bindings
		switch msg.String() {
		case "ctrl+c":
			// Stop any playing stations before quitting
			if a.screen == screenPlay && a.playScreen.player != nil {
				a.playScreen.player.Stop()
			} else if a.screen == screenSearch && a.searchScreen.player != nil {
				a.searchScreen.player.Stop()
			} else if a.screen == screenLucky && a.luckyScreen.player != nil {
				a.luckyScreen.player.Stop()
			}
			return a, tea.Quit
		}

	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height

		// Update main menu size
		if a.screen == screenMainMenu {
			h, v := docStyle().GetFrameSize()
			// Ensure enough height for all menu items (5 items)
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
			a.playScreen = NewPlayModel(a.favoritePath)
			// Set dimensions immediately if we have them
			if a.width > 0 && a.height > 0 {
				a.playScreen.width = a.width
				a.playScreen.height = a.height
			}
			return a, a.playScreen.Init()
		case screenSearch:
			a.searchScreen = NewSearchModel(a.apiClient, a.favoritePath)
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
			a.luckyScreen = NewLuckyModel(a.apiClient, a.favoritePath)
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
		case screenMainMenu:
			// Return to main menu
			return a, nil
		}
		return a, nil

	case backToMainMsg:
		// Handle back to main menu from any screen
		a.screen = screenMainMenu
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
	}

	return a, nil
}

func (a App) updateMainMenu(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Handle Escape to stop playing if playing from main menu
		if msg.String() == "esc" && a.playingFromMain {
			if a.quickFavPlayer != nil {
				a.quickFavPlayer.Stop()
			}
			a.playingFromMain = false
			a.playingStation = nil
			return a, nil
		}

		if msg.String() == "0" {
			// Stop any playing stations before quitting
			if a.playScreen.player != nil {
				a.playScreen.player.Stop()
			}
			if a.searchScreen.player != nil {
				a.searchScreen.player.Stop()
			}
			if a.quickFavPlayer != nil {
				a.quickFavPlayer.Stop()
			}
			// Select exit option
			a.mainMenuList.Select(len(a.mainMenuList.Items()) - 1)
			return a, tea.Quit
		}

		// Check for quick play shortcuts (a-j for first 10 favorites)
		key := msg.String()
		quickPlayKeys := map[string]int{
			"a": 0, "b": 1, "c": 2, "d": 3, "e": 4,
			"f": 5, "g": 6, "h": 7, "i": 8, "l": 9,
		}
		if idx, ok := quickPlayKeys[key]; ok && idx < len(a.quickFavorites) {
			return a.playQuickFavorite(idx)
		}

		// Handle Enter key to select filtered menu item
		if msg.String() == "enter" {
			filterText := strings.ToLower(strings.TrimSpace(a.menuTextInput.Value()))
			if filterText != "" {
				// Find first matching menu item
				for i, item := range a.mainMenuList.Items() {
					if menuItem, ok := item.(components.MenuItem); ok {
						if strings.Contains(strings.ToLower(menuItem.Title()), filterText) {
							a.menuTextInput.Reset()
							return a.executeMenuAction(i)
						}
					}
				}
			} else {
				// No filter text, use current selection
				selected := a.mainMenuList.Index()
				return a.executeMenuAction(selected)
			}
		}

		// Handle arrow keys for menu navigation when no text is typed
		if a.menuTextInput.Value() == "" {
			switch msg.String() {
			case "up", "k":
				a.mainMenuList.CursorUp()
				return a, nil
			case "down", "j":
				a.mainMenuList.CursorDown()
				return a, nil
			}
		}

		// Handle number shortcuts directly
		for i, item := range a.mainMenuList.Items() {
			if menuItem, ok := item.(components.MenuItem); ok {
				if msg.String() == menuItem.Shortcut() {
					a.menuTextInput.Reset()
					a.mainMenuList.Select(i)
					return a.executeMenuAction(i)
				}
			}
		}

		// Pass other keys to text input
		var cmd tea.Cmd
		a.menuTextInput, cmd = a.menuTextInput.Update(msg)
		return a, cmd

	default:
		// Handle text input updates (like blink)
		var cmd tea.Cmd
		a.menuTextInput, cmd = a.menuTextInput.Update(msg)
		return a, cmd
	}
}

func (a App) executeMenuAction(index int) (tea.Model, tea.Cmd) {
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
	case 3: // I Feel Lucky
		return a, func() tea.Msg {
			return navigateMsg{screen: screenLucky}
		}
	case 4: // Gist Management
		return a, func() tea.Msg {
			return navigateMsg{screen: screenGist}
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
		a.quickFavPlayer.Stop()
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
	}
	return "Unknown screen"
}

func (a App) viewMainMenu() string {
	var content strings.Builder

	// Add "Choose an option:" with text input field
	content.WriteString(subtitleStyle().Render("Choose an option:"))
	content.WriteString(" ")
	content.WriteString(a.menuTextInput.View())
	content.WriteString("\n\n")

	// Filter and display menu items based on text input
	filterText := strings.ToLower(strings.TrimSpace(a.menuTextInput.Value()))
	menuContent := ""
	for i, item := range a.mainMenuList.Items() {
		if menuItem, ok := item.(components.MenuItem); ok {
			itemTitle := menuItem.Title()
			// Show all items if no filter, or matching items if filter is set
			if filterText == "" || strings.Contains(strings.ToLower(itemTitle), filterText) {
				prefix := "  "
				if i == a.mainMenuList.Index() && filterText == "" {
					prefix = "> "
					menuContent += selectedItemStyle().Render(fmt.Sprintf("%s%s. %s", prefix, menuItem.Shortcut(), itemTitle)) + "\n"
				} else {
					menuContent += normalItemStyle().Render(fmt.Sprintf("%s%s. %s", prefix, menuItem.Shortcut(), itemTitle)) + "\n"
				}
			}
		}
	}
	content.WriteString(menuContent)

	// Show currently playing station if playing from main menu
	if a.playingFromMain && a.playingStation != nil {
		content.WriteString("\n")
		content.WriteString(successStyle().Render("♫ Now Playing: "))
		content.WriteString(stationNameStyle().Render(a.playingStation.TrimName()))
		content.WriteString("\n")
	}

	// Add quick play favorites if available
	if len(a.quickFavorites) > 0 {
		content.WriteString("\n")
		content.WriteString(quickFavoritesStyle().Render("─── Quick Play Favorites ───"))
		content.WriteString("\n")

		// Define shortcut keys
		shortcutKeys := []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "l"}

		for i, station := range a.quickFavorites {
			if i >= 10 {
				break // Only show first 10
			}
			shortcut := shortcutKeys[i]

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
					stationInfo.WriteString(fmt.Sprintf(" %dkbps", station.Bitrate))
				}
			}

			// Highlight if this is the playing station
			if a.playingFromMain && a.playingStation != nil && a.playingStation.StationUUID == station.StationUUID {
				content.WriteString(fmt.Sprintf("  %s) ", shortcut))
				content.WriteString(successStyle().Render("▶ " + stationInfo.String()))
			} else {
				content.WriteString(fmt.Sprintf("  %s) %s", shortcut, stationInfo.String()))
			}
			content.WriteString("\n")
		}
	}

	// Build help text
	helpText := "↑↓/jk: Navigate • Enter: Select • 1-5: Menu • a-l: Quick play"
	if a.playingFromMain {
		helpText += " • Esc: Stop"
	}
	helpText += " • Ctrl+C: Quit"

	// Use the consistent page template with title and subtitle
	return RenderPage(PageLayout{
		Title:    "Main Menu",
		Subtitle: "Select an Option",
		Content:  content.String(),
		Help:     helpText,
	})
}
