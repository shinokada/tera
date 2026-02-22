package ui

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/shinokada/tera/v3/internal/api"
	"github.com/shinokada/tera/v3/internal/blocklist"
	"github.com/shinokada/tera/v3/internal/player"
	"github.com/shinokada/tera/v3/internal/storage"
	internaltimer "github.com/shinokada/tera/v3/internal/timer"
	"time"
	"github.com/shinokada/tera/v3/internal/ui/components"
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
	screenMostPlayed
	screenTopRated
	screenBrowseTags
	screenTagPlaylists
	screenSleepSummary
)

// Main menu configuration
const mainMenuItemCount = 11

type App struct {
	screen                   Screen
	width                    int
	height                   int
	mainMenuList             list.Model
	playScreen               PlayModel
	searchScreen             SearchModel
	listManagementScreen     ListManagementModel
	luckyScreen              LuckyModel
	mostPlayedScreen         MostPlayedModel
	topRatedScreen           TopRatedModel
	browseTagsScreen         BrowseTagsModel
	tagPlaylistsScreen       TagPlaylistsModel
	gistScreen               GistModel
	settingsScreen           SettingsModel
	shuffleSettingsScreen    ShuffleSettingsModel
	connectionSettingsScreen ConnectionSettingsModel
	appearanceSettingsScreen AppearanceSettingsModel
	blocklistScreen          BlocklistModel
	apiClient                *api.Client
	blocklistManager         *blocklist.Manager
	metadataManager          *storage.MetadataManager // Track play statistics
	ratingsManager           *storage.RatingsManager  // Track station ratings
	tagsManager              *storage.TagsManager     // Custom station tags
	starRenderer             *components.StarRenderer // Render star ratings
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
	// Sleep timer (owned here; activated from player screens)
	sleepTimer    *internaltimer.SleepTimer
	sleepSession  *internaltimer.SleepSession
	sleepDuration time.Duration // duration the user set (for the summary)
	dataPath      string        // path for persisting sleep timer config
	sleepSummary  SleepSummaryModel
	// Cleanup guard
	cleanupOnce sync.Once // Ensures Cleanup is only called once
	// Bubbletea program handle (set by main) for sending async messages
	program *tea.Program
}

// navigateMsg is sent when changing screens
type navigateMsg struct {
	screen Screen
}

// SetProgram gives the App a reference to the running Bubbletea program so
// background goroutines (e.g. the sleep timer) can send messages back to it.
func (a *App) SetProgram(p *tea.Program) {
	a.program = p
}

func NewApp() *App {
	// Get config directory once
	configDir, err := os.UserConfigDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to determine config directory: %v\n", err)
		configDir = "." // Fallback to current directory
	}

	// Get favorite path from environment or use default
	favPath := os.Getenv("TERA_FAVORITE_PATH")
	if favPath == "" {
		favPath = filepath.Join(configDir, "tera", "data", "favorites")
	}

	// Ensure favorites directory exists
	if err := os.MkdirAll(favPath, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to create favorites directory: %v\n", err)
	}

	// Get cache path from environment or use default
	cachePath := os.Getenv("TERA_CACHE_PATH")
	if cachePath == "" {
		cachePath = filepath.Join(configDir, "tera", "data", "cache")
	}

	// Ensure cache directory exists
	if err := os.MkdirAll(cachePath, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to create cache directory: %v\n", err)
	}

	// Initialize blocklist manager
	blocklistPath := filepath.Join(configDir, "tera", "data", "blocklist.json")
	blocklistMgr := blocklist.NewManager(blocklistPath)
	if err := blocklistMgr.Load(context.Background()); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to load blocklist: %v\n", err)
	}

	// Initialize metadata manager for play statistics
	dataPath := filepath.Join(configDir, "tera", "data")
	metadataMgr, err := storage.NewMetadataManager(dataPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to initialize metadata manager: %v\n", err)
	}

	// Initialize ratings manager for star ratings
	ratingsMgr, err := storage.NewRatingsManager(dataPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to initialize ratings manager: %v\n", err)
	}

	// Initialize star renderer
	starRenderer := components.NewStarRenderer(true) // Use unicode by default

	// Initialize tags manager for custom station tags
	tagsMgr, err := storage.NewTagsManager(dataPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to initialize tags manager: %v\n", err)
	}

	app := &App{
		screen:           screenMainMenu,
		favoritePath:     favPath,
		apiClient:        api.NewClient(),
		quickFavPlayer:   player.NewMPVPlayer(),
		helpModel:        components.NewHelpModel(components.CreateMainMenuHelp()),
		blocklistManager: blocklistMgr,
		metadataManager:  metadataMgr,
		ratingsManager:   ratingsMgr,
		tagsManager:      tagsMgr,
		starRenderer:     starRenderer,
		dataPath:         dataPath,
	}

	// Set metadata manager on players for play tracking
	if metadataMgr != nil {
		app.quickFavPlayer.SetMetadataManager(metadataMgr)
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
		components.NewMenuItem("Most Played", "", "3"),
		components.NewMenuItem("Top Rated", "", "4"),
		components.NewMenuItem("Browse by Tag", "", "5"),
		components.NewMenuItem("Tag Playlists", "", "6"),
		components.NewMenuItem("Manage Lists", "", "7"),
		components.NewMenuItem("Block List", "", "8"),
		components.NewMenuItem("I Feel Lucky", "", "9"),
		components.NewMenuItem("Gist Management", "", "0"),
		components.NewMenuItem("Settings", "", "-"),
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

func (a *App) Init() tea.Cmd {
	// Check for updates in the background on startup
	return checkForUpdates()
}

// Cleanup stops all players and releases resources for graceful shutdown
// This function is idempotent and safe to call multiple times
func (a *App) Cleanup() {
	a.cleanupOnce.Do(func() {
		if a.quickFavPlayer != nil {
			_ = a.quickFavPlayer.Stop()
		}
		if a.playScreen.player != nil {
			_ = a.playScreen.player.Stop()
		}
		if a.searchScreen.player != nil {
			_ = a.searchScreen.player.Stop()
		}
		if a.luckyScreen.player != nil {
			_ = a.luckyScreen.player.Stop()
		}
		if a.mostPlayedScreen.player != nil {
			_ = a.mostPlayedScreen.player.Stop()
		}
		if a.topRatedScreen.player != nil {
			_ = a.topRatedScreen.player.Stop()
		}
		if a.tagPlaylistsScreen.player != nil {
			_ = a.tagPlaylistsScreen.player.Stop()
		}
		if a.browseTagsScreen.player != nil {
			_ = a.browseTagsScreen.player.Stop()
		}
		// Close metadata manager to save pending changes
		if a.metadataManager != nil {
			_ = a.metadataManager.Close()
		}
		// Close ratings manager to save pending changes
		if a.ratingsManager != nil {
			_ = a.ratingsManager.Close()
		}
		// Close tags manager to stop background goroutine and flush pending changes
		if a.tagsManager != nil {
			_ = a.tagsManager.Close()
		}
	})
}

func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
			a.Cleanup()
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
			// Set metadata manager for play tracking and metadata display
			if a.metadataManager != nil {
				a.playScreen.metadataManager = a.metadataManager
				if a.playScreen.player != nil {
					a.playScreen.player.SetMetadataManager(a.metadataManager)
				}
			}
			// Set ratings manager and star renderer for star rating feature
			if a.ratingsManager != nil {
				a.playScreen.ratingsManager = a.ratingsManager
			}
			if a.starRenderer != nil {
				a.playScreen.starRenderer = a.starRenderer
			}
			// Set tags manager and renderer
			if a.tagsManager != nil {
				a.playScreen.tagsManager = a.tagsManager
				a.playScreen.tagRenderer = components.NewTagRenderer()
			}
			// Pass data path for sleep timer config
			a.playScreen.dataPath = a.dataPath
			// Set dimensions immediately if we have them
			if a.width > 0 && a.height > 0 {
				a.playScreen.width = a.width
				a.playScreen.height = a.height
			}
			return a, a.playScreen.Init()
		case screenSearch:
			a.searchScreen = NewSearchModel(a.apiClient, a.favoritePath, a.blocklistManager)
			// Set metadata manager for play tracking and metadata display
			if a.metadataManager != nil {
				a.searchScreen.metadataManager = a.metadataManager
				if a.searchScreen.player != nil {
					a.searchScreen.player.SetMetadataManager(a.metadataManager)
				}
			}
			// Set ratings manager and star renderer for star rating feature
			if a.ratingsManager != nil {
				a.searchScreen.ratingsManager = a.ratingsManager
			}
			if a.starRenderer != nil {
				a.searchScreen.starRenderer = a.starRenderer
			}
			// Set tags manager and renderer
			if a.tagsManager != nil {
				a.searchScreen.tagsManager = a.tagsManager
				a.searchScreen.tagRenderer = components.NewTagRenderer()
			}
			// Pass data path for sleep timer config
			a.searchScreen.dataPath = a.dataPath
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
			// Set metadata manager for play tracking and metadata display
			if a.metadataManager != nil {
				a.luckyScreen.metadataManager = a.metadataManager
				if a.luckyScreen.player != nil {
					a.luckyScreen.player.SetMetadataManager(a.metadataManager)
				}
			}
			// Set ratings manager and star renderer for star rating feature
			if a.ratingsManager != nil {
				a.luckyScreen.ratingsManager = a.ratingsManager
			}
			if a.starRenderer != nil {
				a.luckyScreen.starRenderer = a.starRenderer
			}
			// Set tags manager and renderer
			if a.tagsManager != nil {
				a.luckyScreen.tagsManager = a.tagsManager
				a.luckyScreen.tagRenderer = components.NewTagRenderer()
			}
			// Set dimensions immediately if we have them
			if a.width > 0 && a.height > 0 {
				a.luckyScreen.width = a.width
				a.luckyScreen.height = a.height
			}
			return a, a.luckyScreen.Init()
		case screenMostPlayed:
			a.mostPlayedScreen = NewMostPlayedModel(a.metadataManager, a.favoritePath, a.blocklistManager)
			// Set metadata manager for play tracking
			if a.metadataManager != nil && a.mostPlayedScreen.player != nil {
				a.mostPlayedScreen.player.SetMetadataManager(a.metadataManager)
			}
			// Set ratings manager and star renderer for star rating feature
			if a.ratingsManager != nil {
				a.mostPlayedScreen.ratingsManager = a.ratingsManager
			}
			if a.starRenderer != nil {
				a.mostPlayedScreen.starRenderer = a.starRenderer
			}
			// Set tags manager for tag pill display
			if a.tagsManager != nil {
				a.mostPlayedScreen.tagsManager = a.tagsManager
				a.mostPlayedScreen.tagRenderer = components.NewTagRenderer()
			}
			// Set dimensions immediately if we have them
			if a.width > 0 && a.height > 0 {
				a.mostPlayedScreen.width = a.width
				a.mostPlayedScreen.height = a.height
			}
			return a, a.mostPlayedScreen.Init()
		case screenTopRated:
			a.topRatedScreen = NewTopRatedModel(a.ratingsManager, a.metadataManager, a.starRenderer, a.favoritePath, a.blocklistManager)
			// Set metadata manager for play tracking
			if a.metadataManager != nil && a.topRatedScreen.player != nil {
				a.topRatedScreen.player.SetMetadataManager(a.metadataManager)
			}
			// Set tags manager for tag pill display
			if a.tagsManager != nil {
				a.topRatedScreen.tagsManager = a.tagsManager
				a.topRatedScreen.tagRenderer = components.NewTagRenderer()
			}
			// Set dimensions immediately if we have them
			if a.width > 0 && a.height > 0 {
				a.topRatedScreen.width = a.width
				a.topRatedScreen.height = a.height
			}
			return a, a.topRatedScreen.Init()
		case screenBrowseTags:
			if a.tagsManager != nil {
				a.browseTagsScreen = NewBrowseTagsModel(a.tagsManager, a.ratingsManager, a.metadataManager, a.starRenderer, a.blocklistManager)
				// Set metadata manager for play tracking
				if a.metadataManager != nil && a.browseTagsScreen.player != nil {
					a.browseTagsScreen.player.SetMetadataManager(a.metadataManager)
				}
				if a.width > 0 && a.height > 0 {
					a.browseTagsScreen.width = a.width
					a.browseTagsScreen.height = a.height
				}
				return a, a.browseTagsScreen.Init()
			}
			a.screen = screenMainMenu
			return a, func() tea.Msg { return backToMainMsg{} }
		case screenTagPlaylists:
			if a.tagsManager != nil {
				a.tagPlaylistsScreen = NewTagPlaylistsModel(a.tagsManager, a.ratingsManager, a.metadataManager, a.starRenderer, a.blocklistManager)
				// Set metadata manager for play tracking
				if a.metadataManager != nil && a.tagPlaylistsScreen.player != nil {
					a.tagPlaylistsScreen.player.SetMetadataManager(a.metadataManager)
				}
				if a.width > 0 && a.height > 0 {
					a.tagPlaylistsScreen.width = a.width
					a.tagPlaylistsScreen.height = a.height
				}
				return a, a.tagPlaylistsScreen.Init()
			}
			a.screen = screenMainMenu
			return a, func() tea.Msg { return backToMainMsg{} }
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

	case sleepTimerActivateMsg:
		// A player screen confirmed a duration — start the timer.
		d := time.Duration(msg.Minutes) * time.Minute
		a.sleepDuration = d
		a.sleepSession = internaltimer.NewSleepSession()
		_ = storage.SaveSleepTimerConfig(a.dataPath, &storage.SleepTimerConfig{
			LastDurationMinutes: msg.Minutes,
		})
		if a.sleepTimer != nil {
			a.sleepTimer.Cancel()
		}
		a.sleepTimer = internaltimer.NewSleepTimer(func() {
			if a.program != nil {
				a.program.Send(sleepExpiredMsg{})
			}
		})
		a.sleepTimer.Start(d)
		return a, nil

	case sleepTimerCancelMsg:
		if a.sleepTimer != nil {
			a.sleepTimer.Cancel()
			a.sleepTimer = nil
		}
		a.sleepSession = nil
		return a, nil

	case sleepTimerExtendMsg:
		if a.sleepTimer != nil {
			a.sleepTimer.Extend(15 * time.Minute)
		}
		return a, nil

	case tickMsg:
		// Refresh sleep timer countdown on all player screens
		countdown := ""
		if a.sleepTimer != nil {
			if rem, active := a.sleepTimer.Remaining(); active {
				countdown = fmt.Sprintf("Stops in %d:%02d", int(rem.Minutes()), int(rem.Seconds())%60)
			}
		}
		a.playScreen.sleepCountdown = countdown
		a.searchScreen.sleepCountdown = countdown
		// fall through — don't return; let the screen-routing below forward it

	case sleepExpiredMsg:
		// Sleep timer fired — stop playback on all screens and show summary
		a.stopAllPlayback()
		if a.sleepSession != nil {
			a.sleepSummary = NewSleepSummaryModel(a.sleepSession, a.sleepDuration, a.width, a.height)
			a.sleepSession = nil
		}
		a.screen = screenSleepSummary
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
	case screenMostPlayed:
		a.mostPlayedScreen, cmd = a.mostPlayedScreen.Update(msg)
		return a, cmd
	case screenTopRated:
		a.topRatedScreen, cmd = a.topRatedScreen.Update(msg)
		return a, cmd
	case screenBrowseTags:
		a.browseTagsScreen, cmd = a.browseTagsScreen.Update(msg)
		if _, ok := msg.(backToMainMsg); ok {
			a.screen = screenMainMenu
		}
		return a, cmd
	case screenTagPlaylists:
		a.tagPlaylistsScreen, cmd = a.tagPlaylistsScreen.Update(msg)
		if _, ok := msg.(backToMainMsg); ok {
			a.screen = screenMainMenu
		}
		return a, cmd
	case screenSleepSummary:
		var m tea.Model
		m, cmd = a.sleepSummary.Update(msg)
		a.sleepSummary = m.(SleepSummaryModel)
		// navigateMsg from the summary (e.g. 0 → Main Menu) is handled by the
		// top-level navigateMsg case above on the next Update cycle, so just
		// return the cmd here and let it propagate normally.
		return a, cmd
	}

	return a, nil
}

func (a *App) updateMainMenu(msg tea.Msg) (tea.Model, tea.Cmd) {
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
				startTick := a.volumeDisplayFrames <= 0
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
				startTick := a.volumeDisplayFrames <= 0
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
				startTick := a.volumeDisplayFrames <= 0
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

		// Handle '-' shortcut for Settings (immediate, never a QF prefix)
		if msg.String() == "-" {
			a.numberBuffer = ""
			a.unifiedMenuIndex = mainMenuItemCount - 1 // Settings is last menu item
			return a.executeMenuAction(mainMenuItemCount - 1)
		}

		// Handle number input for menu and quick favorites
		key := msg.String()
		if len(key) == 1 && key >= "0" && key <= "9" {
			a.numberBuffer += key

			// maxFavNum is the highest valid QF number (quick favorites start at 10)
			maxFavNum := 9 + len(a.quickFavorites)

			// For 2+ digit buffers, check for QF matches
			if len(a.numberBuffer) >= 2 {
				num := 0
				_, _ = fmt.Sscanf(a.numberBuffer, "%d", &num)

				// Valid QF number — play it immediately
				if num >= 10 && num <= maxFavNum {
					idx := num - 10
					a.numberBuffer = ""
					return a.playQuickFavorite(idx)
				}

				// Out of range or 3+ digits — clear and ignore
				if num > maxFavNum || len(a.numberBuffer) >= 3 {
					a.numberBuffer = ""
					return a, nil
				}
			}

			// Single digit: execute immediately if it cannot start a valid QF number.
			// A digit d can start a QF number only if d*10..d*10+9 overlaps [10, maxFavNum],
			// i.e. d >= 1 && d*10 <= maxFavNum.
			// '0' never starts a QF (00-09 < 10), so always execute immediately.
			if len(a.numberBuffer) == 1 {
				digit := 0
				_, _ = fmt.Sscanf(a.numberBuffer, "%d", &digit)
				canStartQF := digit >= 1 && digit*10 <= maxFavNum
				if !canStartQF {
					a.numberBuffer = ""
					if digit == 0 {
						// '0' → Gist Management
						a.unifiedMenuIndex = 9
						return a.executeMenuAction(9)
					}
					if digit >= 1 && digit <= mainMenuItemCount {
						a.unifiedMenuIndex = digit - 1
						return a.executeMenuAction(digit - 1)
					}
					return a, nil
				}
			}

			// Ambiguous single digit (could start a QF number) — buffer and wait
			return a, nil
		}

		// Handle Enter key to select filtered menu item or confirm number
		if msg.String() == "enter" {
			// Check if there's a buffered number to process
			if a.numberBuffer != "" {
				num := 0
				_, _ = fmt.Sscanf(a.numberBuffer, "%d", &num)
				a.numberBuffer = ""

				// Quick favorites 10+ must be checked before the menu-item range,
				// which also covers 10 and 11 (mainMenuItemCount == 11).
				if num >= 10 {
					idx := num - 10
					if idx < len(a.quickFavorites) {
						return a.playQuickFavorite(idx)
					}
					return a, nil
				}
				// '0' shortcut → Gist Management
				if num == 0 {
					a.unifiedMenuIndex = 9
					return a.executeMenuAction(9)
				}
				// Menu items 1–9
				if num >= 1 && num <= mainMenuItemCount {
					a.unifiedMenuIndex = num - 1
					return a.executeMenuAction(num - 1)
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

func (a *App) executeMenuAction(index int) (tea.Model, tea.Cmd) {
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
	case 2: // Most Played
		return a, func() tea.Msg {
			return navigateMsg{screen: screenMostPlayed}
		}
	case 3: // Top Rated
		return a, func() tea.Msg {
			return navigateMsg{screen: screenTopRated}
		}
	case 4: // Browse by Tag
		return a, func() tea.Msg {
			return navigateMsg{screen: screenBrowseTags}
		}
	case 5: // Tag Playlists
		return a, func() tea.Msg {
			return navigateMsg{screen: screenTagPlaylists}
		}
	case 6: // Manage Lists
		return a, func() tea.Msg {
			return navigateMsg{screen: screenList}
		}
	case 7: // Block List
		return a, func() tea.Msg {
			return navigateMsg{screen: screenBlocklist}
		}
	case 8: // I Feel Lucky
		return a, func() tea.Msg {
			return navigateMsg{screen: screenLucky}
		}
	case 9: // Gist Management
		return a, func() tea.Msg {
			return navigateMsg{screen: screenGist}
		}
	case 10: // Settings
		return a, func() tea.Msg {
			return navigateMsg{screen: screenSettings}
		}
	}
	return a, nil
}

// playQuickFavorite plays a station from the quick favorites list
func (a *App) playQuickFavorite(index int) (tea.Model, tea.Cmd) {
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

func (a *App) View() string {
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
	case screenMostPlayed:
		return a.mostPlayedScreen.View()
	case screenTopRated:
		return a.topRatedScreen.View()
	case screenBrowseTags:
		return a.browseTagsScreen.View()
	case screenTagPlaylists:
		return a.tagPlaylistsScreen.View()
	case screenSleepSummary:
		return a.sleepSummary.View()
	}
	return "Unknown screen"
}

// stopAllPlayback stops mpv on every screen that may be playing.
func (a *App) stopAllPlayback() {
	for _, p := range []*player.MPVPlayer{
		a.quickFavPlayer,
		a.playScreen.player,
		a.searchScreen.player,
		a.luckyScreen.player,
		a.mostPlayedScreen.player,
		a.topRatedScreen.player,
		a.tagPlaylistsScreen.player,
		a.browseTagsScreen.player,
	} {
		if p != nil {
			_ = p.Stop()
		}
	}
	a.playingFromMain = false
	a.playingStation = nil
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

func (a *App) viewMainMenu() string {
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
		{"3", "Most Played"},
		{"4", "Top Rated"},
		{"5", "Browse by Tag"},
		{"6", "Tag Playlists"},
		{"7", "Manage Lists"},
		{"8", "Block List"},
		{"9", "I Feel Lucky"},
		{"0", "Gist Management"},
		{"-", "Settings"},
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
		// Show star rating if rated
		if a.ratingsManager != nil {
			if r := a.ratingsManager.GetRating(a.playingStation.StationUUID); r != nil && a.starRenderer != nil {
				content.WriteString(" ")
				content.WriteString(a.starRenderer.RenderCompactPlain(r.Rating))
			}
		}
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
		helpText = "↑↓/jk: Navigate • Enter: Select • 1-0/-: Menu shortcuts • 10+: Quick Play • ?: Help"
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
