package ui

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/shinokada/tera/internal/api"
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
	playScreen           PlayModel
	searchScreen         SearchModel
	listManagementScreen ListManagementModel
	luckyScreen          LuckyModel
	gistScreen           GistModel
	apiClient            *api.Client
	favoritePath         string
	quickFavorites       []api.Station
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

	app := App{
		screen:       screenMainMenu,
		favoritePath: favPath,
		apiClient:    api.NewClient(),
	}

	// Initialize main menu
	app.initMainMenu()

	// Ensure My-favorites.json exists at startup
	app.ensureMyFavorites()

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

func (a App) Init() tea.Cmd {
	return nil
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
			h, v := docStyle.GetFrameSize()
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
		if msg.String() == "0" {
			// Stop any playing stations before quitting
			if a.playScreen.player != nil {
				a.playScreen.player.Stop()
			}
			if a.searchScreen.player != nil {
				a.searchScreen.player.Stop()
			}
			// Select exit option
			a.mainMenuList.Select(len(a.mainMenuList.Items()) - 1)
			return a, tea.Quit
		}

		// Handle menu navigation
		newList, selected := components.HandleMenuKey(msg, a.mainMenuList)
		a.mainMenuList = newList

		if selected >= 0 {
			// Execute selected action
			return a.executeMenuAction(selected)
		}
	}
	return a, nil
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

	// Menu items
	content.WriteString(a.mainMenuList.View())

	// Add quick play favorites if available
	if len(a.quickFavorites) > 0 {
		content.WriteString("\n\n")
		content.WriteString(quickFavoritesStyle.Render("Quick Play Favorites"))
		for i, station := range a.quickFavorites {
			if i >= 10 {
				break // Only show first 10
			}
			shortcut := fmt.Sprintf("%d", 10+i)
			content.WriteString(fmt.Sprintf("\n  %s. ▶ %s", shortcut, station.TrimName()))
		}
	}

	// Use the consistent page template with title and subtitle
	return RenderPage(PageLayout{
		Title:    "Main Menu",
		Subtitle: "Select an Option",
		Content:  content.String(),
		Help:     "↑↓/jk: Navigate • Enter: Select • 1-5: Quick select • Ctrl+C: Quit",
	})
}
