package ui

import (
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/shinokada/tera/internal/api"
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
	screen       Screen
	width        int
	height       int
	playScreen   PlayModel
	searchScreen SearchModel
	apiClient    *api.Client
	favoritePath string
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

	return App{
		screen:       screenMainMenu,
		favoritePath: favPath,
		apiClient:    api.NewClient(),
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
			return a, tea.Quit
		case "q":
			// Only quit from main menu
			if a.screen == screenMainMenu {
				return a, tea.Quit
			}
		}

	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height

	case navigateMsg:
		a.screen = msg.screen

		// Initialize screen-specific models
		switch msg.screen {
		case screenPlay:
			a.playScreen = NewPlayModel(a.favoritePath)
			return a, a.playScreen.Init()
		case screenSearch:
			a.searchScreen = NewSearchModel(a.apiClient, a.favoritePath)
			return a, a.searchScreen.Init()
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
	}

	return a, nil
}

func (a App) updateMainMenu(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "1":
			// Navigate to Play screen
			return a, func() tea.Msg {
				return navigateMsg{screen: screenPlay}
			}
		case "2":
			// Navigate to Search screen
			return a, func() tea.Msg {
				return navigateMsg{screen: screenSearch}
			}
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
	}
	return "Unknown screen"
}

func (a App) viewMainMenu() string {
	return titleStyle.Render("TERA - Terminal Radio") + "\n\n" +
		"1. Play from Favorites\n" +
		"2. Search Stations\n" +
		"3. Manage Lists (coming soon)\n" +
		"6. Gist Management (coming soon)\n\n" +
		helpStyle.Render("q: quit • 1: play • 2: search")
}
