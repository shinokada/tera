package ui

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/shinokada/tera/internal/api"
	"github.com/shinokada/tera/internal/ui/components"
)

type Screen int

const (
	screenMainMenu Screen = iota
	screenPlay
	screenSearch
	screenList
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

	app := App{
		screen:       screenMainMenu,
		favoritePath: favPath,
		apiClient:    api.NewClient(),
	}

	// Initialize main menu
	app.initMainMenu()

	return app
}

func (a *App) initMainMenu() {
	items := []components.MenuItem{
		components.NewMenuItem("Play from Favorites", "", "1"),
		components.NewMenuItem("Search Stations", "", "2"),
		components.NewMenuItem("Manage Lists", "", "3"),
		components.NewMenuItem("I Feel Lucky", "(coming soon)", "4"),
		components.NewMenuItem("Delete Station", "(coming soon)", "5"),
		components.NewMenuItem("Gist Management", "(coming soon)", "6"),
		components.NewMenuItem("Exit", "", "0"),
	}

	// Height will be auto-adjusted by CreateMenu to fit all items
	a.mainMenuList = components.CreateMenu(items, "TERA - Terminal Radio", 50, 20)
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
			}
			return a, tea.Quit
		}

	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		
		// Update main menu size
		if a.screen == screenMainMenu {
			h, v := docStyle.GetFrameSize()
			// Ensure enough height for all menu items (7 items + title + help)
			minHeight := 12
			menuHeight := msg.Height - v
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
	}

	return a, nil
}

func (a App) updateMainMenu(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Handle quit - 'q' works from anywhere in the menu
		if msg.String() == "q" {
			// Stop any playing stations before quitting
			if a.playScreen.player != nil {
				a.playScreen.player.Stop()
			}
			if a.searchScreen.player != nil {
				a.searchScreen.player.Stop()
			}
			return a, tea.Quit
		}
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
		// Coming soon
		return a, nil
	case 4: // Delete Station
		// Coming soon
		return a, nil
	case 5: // Gist Management
		// Coming soon
		return a, nil
	case 6: // Exit
		return a, tea.Quit
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
	}
	return "Unknown screen"
}

func (a App) viewMainMenu() string {
	content := a.mainMenuList.View()

	// Add quick play favorites if available
	if len(a.quickFavorites) > 0 {
		content += "\n\n" + quickFavoritesStyle.Render("Quick Play Favorites")
		for i, station := range a.quickFavorites {
			if i >= 10 {
				break // Only show first 10
			}
			shortcut := fmt.Sprintf("%d", 10+i)
			content += fmt.Sprintf("\n  %s. ▶ %s", shortcut, station.TrimName())
		}
	}

	content += "\n\n" + helpStyle.Render("↑↓/jk: Navigate • Enter: Select • 1-6: Quick select • q: Quit")

	return docStyle.Render(content)
}

var (
	docStyle = helpStyle.Copy().Padding(1, 2)

	quickFavoritesStyle = titleStyle.Copy().
				Foreground(lipgloss.Color("99"))
)
