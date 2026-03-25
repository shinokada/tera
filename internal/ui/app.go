package ui

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"

	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/shinokada/tera/v3/internal/api"
	"github.com/shinokada/tera/v3/internal/blocklist"
	"github.com/shinokada/tera/v3/internal/config"
	"github.com/shinokada/tera/v3/internal/player"
	"github.com/shinokada/tera/v3/internal/storage"
	internaltimer "github.com/shinokada/tera/v3/internal/timer"
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

// syncBackupMenuLabel is the label shown in the main menu for the Sync & Backup screen.
const syncBackupMenuLabel = "Sync & Backup"

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
	playHistoryCfg           config.PlayHistoryConfig // cached play history settings
	playOptsCfg              config.PlayOptionsConfig // cached play options settings
	// Continuous playback (v3.11) — non-nil when a screen has handed off its player
	activePlayer        *player.MPVPlayer             // app-level player after a handoff
	activeStation       *api.Station                  // station currently handed off
	activeContextLabel  string                        // context label from the originating screen
	recentlyPlayed      []storage.StationWithMetadata // refreshed on each return to main menu
	numberBuffer        string                        // Buffer for multi-digit number input
	unifiedMenuIndex    int                           // Unified index for navigating both menu and favorites
	helpModel           components.HelpModel          // Help overlay
	volumeDisplay       string                        // Temporary volume display message
	volumeDisplayFrames int                           // Countdown for volume display
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
	// Quick Play Favorites viewport
	qfViewOffset    int // first QF entry index visible on screen
	qfVisibleWindow int // last-known number of QF rows that fit on screen
	// Recently Played viewport
	rpViewOffset    int // first RP entry index visible on screen
	rpVisibleWindow int // last-known number of RP rows that fit on screen
	// Cleanup guard
	cleanupOnce sync.Once // Ensures Cleanup is only called once
	// Bubbletea program handle (set by main) for sending async messages.
	// Uses atomic.Pointer so the timer goroutine can read it race-free.
	program atomic.Pointer[tea.Program]
}

// navigateMsg is sent when changing screens
type navigateMsg struct {
	screen Screen
}

// SetProgram gives the App a reference to the running Bubbletea program so
// background goroutines (e.g. the sleep timer) can send messages back to it.
// It must be called before any sleep timer can be activated; in practice this
// is guaranteed because timer activation requires user interaction.
func (a *App) SetProgram(p *tea.Program) {
	a.program.Store(p)
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

	// Load play history config
	if ph, err := storage.LoadPlayHistoryConfigFromUnified(); err == nil {
		app.playHistoryCfg = ph
	} else {
		app.playHistoryCfg = config.DefaultPlayHistoryConfig()
	}

	// Load play options config
	if po, err := storage.LoadPlayOptionsConfigFromUnified(); err == nil {
		app.playOptsCfg = po
	} else {
		app.playOptsCfg = config.DefaultPlayOptionsConfig()
	}

	// Initialize header renderer
	InitializeHeaderRenderer()

	// Initialize main menu
	app.initMainMenu()

	// Ensure My-favorites.json exists at startup
	app.ensureMyFavorites()

	// Load quick favorites
	app.loadQuickFavorites()

	// Load recently played history
	app.loadRecentlyPlayed()

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
		components.NewMenuItem(syncBackupMenuLabel, "", "0"),
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
	// Check for updates in the background on startup.
	return tea.Batch(checkForUpdates())
}

// Cleanup stops all players and releases resources for graceful shutdown.
// This function is idempotent and safe to call multiple times.
func (a *App) Cleanup() {
	a.cleanupOnce.Do(func() {
		if a.sleepTimer != nil {
			a.sleepTimer.Cancel()
			a.sleepTimer = nil
		}

		// Phase 5: Save LastUsedVolume before stopping players so the
		// GetVolume() calls below reach live players, not nil pointers.
		if a.playOptsCfg.StartVolumeMode == "last_used" {
			var lastVol int
			if a.activePlayer != nil {
				lastVol = a.activePlayer.GetVolume()
			} else if a.quickFavPlayer != nil {
				lastVol = a.quickFavPlayer.GetVolume()
			}
			if lastVol > 0 {
				a.playOptsCfg.LastUsedVolume = lastVol
				_ = storage.SavePlayOptionsConfigToUnified(a.playOptsCfg)
			}
		}

		if a.activePlayer != nil {
			_ = a.activePlayer.Stop()
			a.activePlayer = nil
		}
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
		case "x":
			// Kill the ContinueOnNavigate station (or all mpv when things go
			// wrong) from any screen.
			if a.activeStation != nil || (a.playingFromMain && a.quickFavPlayer != nil) {
				a.stopAllPlayback()
				a.broadcastNowPlayingBar()
				return a, nil
			}
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
			// Stop all screen-owned players so they don't overlap with whatever
			// the user is about to play. When ContinueOnNavigate is on we keep
			// activePlayer alive — the user is just browsing; it will be stopped
			// the moment they actually select and play a new station.
			a.stopScreenPlayers()
			if !a.playOptsCfg.ContinueOnNavigate {
				if a.activePlayer != nil {
					_ = a.activePlayer.Stop()
					a.activePlayer = nil
					a.activeStation = nil
					a.activeContextLabel = ""
				}
				a.playingFromMain = false
				a.playingStation = nil
			}
			a.playScreen = NewPlayModel(a.favoritePath, a.blocklistManager)
			a.playScreen.nowPlayingBar = a.buildNowPlayingBannerText()
			// Pass play options so volume/behaviour is consistent
			a.playScreen.playOptsCfg = a.playOptsCfg
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
			// Sync running timer state so Z cancels rather than reopens the dialog.
			a.playScreen.sleepTimerActive = a.sleepTimer != nil
			// Set dimensions immediately if we have them
			if a.width > 0 && a.height > 0 {
				a.playScreen.width = a.width
				a.playScreen.height = a.height
			}
			return a, a.playScreen.Init()
		case screenSearch:
			a.searchScreen = NewSearchModel(a.apiClient, a.favoritePath, a.dataPath, a.blocklistManager)
			a.searchScreen.nowPlayingBar = a.buildNowPlayingBannerText()
			// Create and wire up the MPV player for search screen playback
			searchPlayer := player.NewMPVPlayer()
			a.searchScreen.player = searchPlayer
			// Load voted stations for the search screen
			if votedStations, err := storage.LoadVotedStations(); err == nil {
				a.searchScreen.votedStations = votedStations
			} else {
				a.searchScreen.votedStations = &storage.VotedStations{Stations: []storage.VotedStation{}}
			}
			// Apply play options so volume/behaviour is consistent
			a.searchScreen.playOptsCfg = a.playOptsCfg
			// Apply search visibility setting
			if cfg, err := storage.LoadBlocklistConfigFromUnified(); err == nil {
				a.searchScreen.showBlockedInSearch = cfg.ShowBlockedInSearch
			} else {
				fmt.Fprintf(os.Stderr, "Warning: failed to load blocklist visibility setting: %v\n", err)
			}
			// Set metadata manager for play tracking and metadata display
			if a.metadataManager != nil {
				a.searchScreen.metadataManager = a.metadataManager
				searchPlayer.SetMetadataManager(a.metadataManager)
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
			// Sync running timer state so Z cancels rather than reopens the dialog.
			a.searchScreen.sleepTimerActive = a.sleepTimer != nil
			// Set dimensions immediately if we have them
			if a.width > 0 && a.height > 0 {
				a.searchScreen.width = a.width
				a.searchScreen.height = a.height
			}
			return a, a.searchScreen.Init()
		case screenList:
			a.listManagementScreen = NewListManagementModel(a.favoritePath)
			a.listManagementScreen.nowPlayingBar = a.buildNowPlayingBannerText()
			// Set dimensions immediately if we have them
			if a.width > 0 && a.height > 0 {
				a.listManagementScreen.width = a.width
				a.listManagementScreen.height = a.height
			}
			return a, a.listManagementScreen.Init()
		case screenLucky:
			a.luckyScreen = NewLuckyModel(a.apiClient, a.favoritePath, a.blocklistManager)
			a.luckyScreen.nowPlayingBar = a.buildNowPlayingBannerText()
			// Pass play options so volume/behaviour is consistent
			a.luckyScreen.playOptsCfg = a.playOptsCfg
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
			a.mostPlayedScreen.nowPlayingBar = a.buildNowPlayingBannerText()
			// Pass play options so volume/behaviour is consistent
			a.mostPlayedScreen.playOptsCfg = a.playOptsCfg
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
			// Stop other screen-owned players (but not activePlayer when
			// ContinueOnNavigate is on — the handoff keeps that alive).
			a.stopScreenPlayers()
			a.topRatedScreen = NewTopRatedModel(a.ratingsManager, a.metadataManager, a.starRenderer, a.favoritePath, a.blocklistManager)
			a.topRatedScreen.nowPlayingBar = a.buildNowPlayingBannerText()
			// Give Top Rated its own dedicated player so it doesn't share
			// quickFavPlayer with the main menu. This prevents stopScreenPlayers()
			// or a ContinueOnNavigate handoff from killing the wrong stream.
			topRatedPlayer := player.NewMPVPlayer()
			a.topRatedScreen.player = topRatedPlayer
			if a.metadataManager != nil {
				topRatedPlayer.SetMetadataManager(a.metadataManager)
			}
			// Pass play options so volume/behaviour is consistent
			a.topRatedScreen.playOptsCfg = a.playOptsCfg
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
				a.browseTagsScreen.nowPlayingBar = a.buildNowPlayingBannerText()
				// Pass play options so volume/behaviour is consistent
				a.browseTagsScreen.playOptsCfg = a.playOptsCfg
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
				a.tagPlaylistsScreen.nowPlayingBar = a.buildNowPlayingBannerText()
				// Pass play options so volume/behaviour is consistent
				a.tagPlaylistsScreen.playOptsCfg = a.playOptsCfg
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
			a.gistScreen.nowPlayingBar = a.buildNowPlayingBannerText()
			// Set dimensions immediately if we have them
			if a.width > 0 && a.height > 0 {
				a.gistScreen.width = a.width
				a.gistScreen.height = a.height
			}
			return a, a.gistScreen.Init()
		case screenSettings:
			a.settingsScreen = NewSettingsModel(a.favoritePath)
			a.settingsScreen.nowPlayingBar = a.buildNowPlayingBannerText()
			// Pass metadata manager for play history clear action
			a.settingsScreen.metadataManager = a.metadataManager
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
			a.shuffleSettingsScreen.nowPlayingBar = a.buildNowPlayingBannerText()
			// Set dimensions immediately if we have them
			if a.width > 0 && a.height > 0 {
				a.shuffleSettingsScreen.width = a.width
				a.shuffleSettingsScreen.height = a.height
			}
			return a, a.shuffleSettingsScreen.Init()
		case screenConnectionSettings:
			a.connectionSettingsScreen = NewConnectionSettingsModel()
			a.connectionSettingsScreen.nowPlayingBar = a.buildNowPlayingBannerText()
			// Set dimensions immediately if we have them
			if a.width > 0 && a.height > 0 {
				a.connectionSettingsScreen.width = a.width
				a.connectionSettingsScreen.height = a.height
			}
			return a, a.connectionSettingsScreen.Init()
		case screenAppearanceSettings:
			a.appearanceSettingsScreen = NewAppearanceSettingsModel()
			a.appearanceSettingsScreen.nowPlayingBar = a.buildNowPlayingBannerText()
			// Set dimensions immediately if we have them
			if a.width > 0 && a.height > 0 {
				a.appearanceSettingsScreen.width = a.width
				a.appearanceSettingsScreen.height = a.height
			}
			return a, a.appearanceSettingsScreen.Init()
		case screenBlocklist:
			a.blocklistScreen = NewBlocklistModel(a.blocklistManager)
			a.blocklistScreen.nowPlayingBar = a.buildNowPlayingBannerText()
			// Set dimensions immediately if we have them
			if a.width > 0 && a.height > 0 {
				a.blocklistScreen.width = a.width
				a.blocklistScreen.height = a.height
			}
			return a, a.blocklistScreen.Init()
		case screenMainMenu:
			// Return to main menu and reload favorites and play history
			a.loadQuickFavorites()
			a.refreshPlayHistoryConfig()
			a.refreshPlayOptionsConfig()
			a.loadRecentlyPlayed()
			a.unifiedMenuIndex = 0
			a.numberBuffer = ""
			a.qfViewOffset = 0
			a.rpViewOffset = 0
			return a, nil
		}
		return a, nil

	case sleepTimerActivateMsg:
		// A player screen confirmed a duration — start the timer.
		d := time.Duration(msg.Minutes) * time.Minute
		a.sleepDuration = d
		a.sleepSession = internaltimer.NewSleepSession()
		if a.sleepTimer != nil {
			a.sleepTimer.Cancel()
		}
		a.sleepTimer = internaltimer.NewSleepTimer(func() {
			if p := a.program.Load(); p != nil {
				p.Send(sleepExpiredMsg{})
			}
		})
		a.sleepTimer.Start(d)
		// Persist the chosen duration off the UI goroutine to avoid blocking on slow filesystems.
		dataPath, mins := a.dataPath, msg.Minutes
		return a, func() tea.Msg {
			_ = storage.SaveSleepTimerConfig(dataPath, &storage.SleepTimerConfig{
				LastDurationMinutes: mins,
			})
			return nil
		}

	case sleepTimerCancelMsg:
		if a.sleepTimer != nil {
			a.sleepTimer.Cancel()
			a.sleepTimer = nil
		}
		a.sleepSession = nil
		a.playScreen.sleepTimerActive = false
		a.playScreen.sleepCountdown = ""
		a.searchScreen.sleepTimerActive = false
		a.searchScreen.sleepCountdown = 0
		return a, nil

	case sleepTimerExtendMsg:
		if a.sleepTimer != nil {
			a.sleepTimer.Extend(time.Duration(msg.Minutes) * time.Minute)
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
		// countdown should be int, not string
		// If countdown is string, convert to int as needed
		// a.searchScreen.sleepCountdown = countdown
		// fall through — don't return; let the screen-routing below forward it

	case sleepExpiredMsg:
		// Sleep timer fired — stop playback on all screens and show summary
		a.stopAllPlayback()
		if a.sleepTimer != nil {
			a.sleepTimer.Cancel()
			a.sleepTimer = nil
		}
		if a.sleepSession != nil {
			a.sleepSummary = NewSleepSummaryModel(a.sleepSession, a.sleepDuration, a.width, a.height)
			a.sleepSummary.nowPlayingBar = a.buildNowPlayingBannerText()
			a.sleepSession = nil
		} else {
			// Session data unavailable (unexpected); fall back to main menu
			a.screen = screenMainMenu
			a.loadQuickFavorites()
			a.playScreen.sleepTimerActive = false
			a.playScreen.sleepCountdown = ""
			a.searchScreen.sleepTimerActive = false
			a.searchScreen.sleepCountdown = 0
			return a, nil
		}
		a.screen = screenSleepSummary
		a.playScreen.sleepTimerActive = false
		a.playScreen.sleepCountdown = ""
		a.searchScreen.sleepTimerActive = false
		a.searchScreen.sleepCountdown = 0
		return a, nil

	case handoffPlaybackMsg:
		// A play screen is navigating away with ContinueOnNavigate=true.
		// Stop any previously handed-off player before accepting the new one.
		if a.activePlayer != nil && a.activePlayer != msg.player {
			_ = a.activePlayer.Stop()
		}
		a.activePlayer = msg.player
		a.activeStation = msg.station
		a.activeContextLabel = msg.contextLabel
		// Set these so Now Playing bar and controls work from main menu
		a.playingFromMain = true
		if msg.station != nil {
			a.playingStation = msg.station
		}
		a.broadcastNowPlayingBar()
		return a, nil

	case stopActivePlaybackMsg:
		// Stop the app-level handed-off player and clear its state.
		if a.activePlayer != nil {
			_ = a.activePlayer.Stop()
			a.activePlayer = nil
		}
		a.activeStation = nil
		a.activeContextLabel = ""
		a.broadcastNowPlayingBar()
		return a, nil

	case backToMainMsg:
		// Handle back to main menu from any screen
		a.screen = screenMainMenu
		// Reload quick favorites and play history in case they were updated
		a.loadQuickFavorites()
		a.refreshPlayHistoryConfig()
		a.refreshPlayOptionsConfig()
		a.loadRecentlyPlayed()
		a.unifiedMenuIndex = 0
		a.numberBuffer = ""
		a.qfViewOffset = 0
		a.rpViewOffset = 0
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

			// maxFavNum is the highest valid shortcut number (QF start at 10, RP follow)
			maxFavNum := 9 + len(a.quickFavorites) + len(a.recentlyPlayed)

			// For 2+ digit buffers, check for QF matches
			if len(a.numberBuffer) >= 2 {
				num := 0
				_, _ = fmt.Sscanf(a.numberBuffer, "%d", &num)

				// Valid QF or RP number — play it immediately
				if num >= 10 && num <= maxFavNum {
					idx := num - 10
					a.numberBuffer = ""
					if idx < len(a.quickFavorites) {
						return a.playQuickFavorite(idx)
					}
					rpIdx := idx - len(a.quickFavorites)
					if rpIdx < len(a.recentlyPlayed) {
						return a.playRecentStation(rpIdx)
					}
					return a, nil
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

				// QF and RP shortcuts (10+) must be checked before the menu-item
				// range, which also covers 10 and 11 (mainMenuItemCount == 11).
				if num >= 10 {
					idx := num - 10
					if idx < len(a.quickFavorites) {
						return a.playQuickFavorite(idx)
					}
					rpIdx := idx - len(a.quickFavorites)
					if rpIdx < len(a.recentlyPlayed) {
						return a.playRecentStation(rpIdx)
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
			}
			// It's a quick favorite or recently played
			offset := a.unifiedMenuIndex - menuItemCount
			if offset < len(a.quickFavorites) {
				return a.playQuickFavorite(offset)
			}
			rpIdx := offset - len(a.quickFavorites)
			if rpIdx < len(a.recentlyPlayed) {
				return a.playRecentStation(rpIdx)
			}
			return a, nil
		}

		// Handle arrow keys for unified menu navigation
		menuItemCount := mainMenuItemCount
		favCount := len(a.quickFavorites)
		totalItems := menuItemCount + favCount + len(a.recentlyPlayed)

		switch msg.String() {
		case "up", "k":
			a.numberBuffer = "" // Clear buffer on navigation
			if a.unifiedMenuIndex > 0 {
				a.unifiedMenuIndex--
				a.updateQFViewOffset(a.qfVisibleWindow)
				a.updateRPViewOffset(a.rpVisibleWindow)
			}
			return a, nil
		case "down", "j":
			a.numberBuffer = "" // Clear buffer on navigation
			if a.unifiedMenuIndex < totalItems-1 {
				a.unifiedMenuIndex++
				a.updateQFViewOffset(a.qfVisibleWindow)
				a.updateRPViewOffset(a.rpVisibleWindow)
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
	// Stop any currently playing quick favorite before navigating.
	// When ContinueOnNavigate is on and something is handed off to App-level,
	// don't stop it — it should keep playing across screens.
	// We do not require quickFavPlayer.IsPlaying() here because Play() is called
	// asynchronously; checking IsPlaying() is a race condition that drops the
	// handoff when the user navigates before the cmd goroutine runs.
	// activePlayer == nil guards the case where a player screen already did a
	// handoffPlaybackMsg — in that scenario playingFromMain/playingStation are
	// also set but we must NOT replace the already-correct activePlayer with the
	// idle quickFavPlayer.
	if a.playingFromMain && a.playingStation != nil && a.activePlayer == nil {
		if a.playOptsCfg.ContinueOnNavigate && a.quickFavPlayer != nil {
			// Hand the main-menu player off to activePlayer so it survives
			// stopScreenPlayers() when the destination screen is initialized.
			// (a.activePlayer is guaranteed nil by the outer condition, so no
			// stop-before-replace is needed here.)
			a.activePlayer = a.quickFavPlayer
			a.activeStation = a.playingStation
			a.activeContextLabel = "Quick Play"
			// Install a fresh quickFavPlayer so stopScreenPlayers doesn't
			// kill the handed-off one.
			a.quickFavPlayer = player.NewMPVPlayer()
			if a.metadataManager != nil {
				a.quickFavPlayer.SetMetadataManager(a.metadataManager)
			}
			a.broadcastNowPlayingBar()
		} else if a.quickFavPlayer != nil {
			_ = a.quickFavPlayer.Stop()
		}
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
	case 9: // Sync & Backup
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

// refreshPlayHistoryConfig reloads the play history config from disk so that
// changes made in Settings (e.g. decreasing the size) are reflected immediately.
func (a *App) refreshPlayHistoryConfig() {
	if ph, err := storage.LoadPlayHistoryConfigFromUnified(); err == nil {
		a.playHistoryCfg = ph
	}
}

// refreshPlayOptionsConfig reloads the play options config from disk so that
// changes made in Settings > Play Options are reflected immediately on return.
func (a *App) refreshPlayOptionsConfig() {
	if po, err := storage.LoadPlayOptionsConfigFromUnified(); err == nil {
		a.playOptsCfg = po
	}
}

// loadRecentlyPlayed loads recently played stations from metadata, respecting
// the play history config (enabled flag and size limit).
// Callers that need a fresh config from disk should call refreshPlayHistoryConfig
// first (e.g. when returning to the main menu from Settings).
func (a *App) loadRecentlyPlayed() {
	if !a.playHistoryCfg.Enabled || a.metadataManager == nil {
		a.recentlyPlayed = nil
		return
	}
	a.recentlyPlayed = a.metadataManager.GetRecentlyPlayed(a.playHistoryCfg.Size)
}

// playQuickFavorite plays a station from the quick favorites list
func (a *App) playQuickFavorite(index int) (tea.Model, tea.Cmd) {
	if index >= len(a.quickFavorites) {
		return a, nil
	}

	station := a.quickFavorites[index]
	a.playingStation = &station
	a.playingFromMain = true

	// Stop any currently playing station. This also cancels any in-flight async
	// Play() goroutine via the killed flag.
	if a.quickFavPlayer != nil {
		_ = a.quickFavPlayer.Stop()
	}
	// Stop any station handed off from a previous ContinueOnNavigate session.
	if a.activePlayer != nil && a.activePlayer != a.quickFavPlayer {
		_ = a.activePlayer.Stop()
		a.activePlayer = nil
		a.activeStation = nil
		a.activeContextLabel = ""
	}

	// Create a fresh player for the new station. The old player's killed flag may
	// be true after Stop() (if it was idle), which would silently block Play().
	fresh := player.NewMPVPlayer()
	if a.metadataManager != nil {
		fresh.SetMetadataManager(a.metadataManager)
	}
	a.quickFavPlayer = fresh
	p := fresh

	// Start playback on the fresh player
	return a, func() tea.Msg {
		if err := p.Play(&station); err != nil {
			return playbackErrorMsg{err}
		}
		return playbackStartedMsg{}
	}
}

// playRecentStation plays a station from the recently played list
func (a *App) playRecentStation(index int) (tea.Model, tea.Cmd) {
	if index >= len(a.recentlyPlayed) {
		return a, nil
	}

	station := a.recentlyPlayed[index].Station
	a.playingStation = &station
	a.playingFromMain = true

	// Stop any currently playing station. This also cancels any in-flight async
	// Play() goroutine via the killed flag.
	if a.quickFavPlayer != nil {
		_ = a.quickFavPlayer.Stop()
	}
	// Stop any station handed off from a previous ContinueOnNavigate session.
	if a.activePlayer != nil && a.activePlayer != a.quickFavPlayer {
		_ = a.activePlayer.Stop()
		a.activePlayer = nil
		a.activeStation = nil
		a.activeContextLabel = ""
	}

	// Create a fresh player for the new station. The old player's killed flag may
	// be true after Stop() (if it was idle), which would silently block Play().
	fresh := player.NewMPVPlayer()
	if a.metadataManager != nil {
		fresh.SetMetadataManager(a.metadataManager)
	}
	a.quickFavPlayer = fresh
	p := fresh

	// Start playback on the fresh player
	return a, func() tea.Msg {
		if err := p.Play(&station); err != nil {
			return playbackErrorMsg{err}
		}
		return playbackStartedMsg{}
	}
}

// isPlayerScreen reports whether the current screen manages its own playback
// UI. The Now Playing bar should not be shown on these screens to avoid
// duplicate/conflicting playback information.
func (a *App) isPlayerScreen() bool {
	switch a.screen {
	case screenPlay, screenSearch, screenLucky,
		screenMostPlayed, screenTopRated,
		screenBrowseTags, screenTagPlaylists:
		return true
	}
	return false
}

// nowPlayingBar returns the now-playing bar string when a station has been
// handed off to the app level (ContinueOnNavigate=true). Returns "" otherwise.
func (a *App) nowPlayingBar() string {
	if a.activeStation == nil {
		return ""
	}
	vol := 0
	if a.activePlayer != nil {
		vol = a.activePlayer.GetVolume()
	}
	return renderNowPlayingBar(a.activeStation, a.activeContextLabel, vol)
}

// buildNowPlayingBannerText returns the static now-playing banner string shown
// above the footer help bar on non-player screens when ContinueOnNavigate is ON.
// The banner includes the station name, optional star rating, and context label.
func (a *App) buildNowPlayingBannerText() string {
	if a.activeStation == nil {
		return ""
	}
	name := a.activeStation.TrimName()
	bar := "♫ Now Playing: " + name
	if a.ratingsManager != nil && a.starRenderer != nil {
		if r := a.ratingsManager.GetRating(a.activeStation.StationUUID); r != nil && r.Rating > 0 {
			bar += " " + a.starRenderer.RenderCompactPlain(r.Rating)
		}
	}
	if a.activeContextLabel != "" {
		bar += "  ·  [" + a.activeContextLabel + "]"
	}
	bar += "  ·  x: Stop"
	return successStyle().Render(bar)
}

// broadcastNowPlayingBar pushes the current now-playing banner to every
// non-player screen model so it appears above each screen's footer help bar.
func (a *App) broadcastNowPlayingBar() {
	bar := a.buildNowPlayingBannerText()
	// Non-player screens
	a.listManagementScreen.nowPlayingBar = bar
	a.gistScreen.nowPlayingBar = bar
	a.settingsScreen.nowPlayingBar = bar
	a.shuffleSettingsScreen.nowPlayingBar = bar
	a.connectionSettingsScreen.nowPlayingBar = bar
	a.appearanceSettingsScreen.nowPlayingBar = bar
	a.blocklistScreen.nowPlayingBar = bar
	a.sleepSummary.nowPlayingBar = bar
	// Player screens (shown in list/browse states when ContinueOnNavigate is on)
	a.playScreen.nowPlayingBar = bar
	a.searchScreen.nowPlayingBar = bar
	a.luckyScreen.nowPlayingBar = bar
	a.mostPlayedScreen.nowPlayingBar = bar
	a.topRatedScreen.nowPlayingBar = bar
	a.browseTagsScreen.nowPlayingBar = bar
	a.tagPlaylistsScreen.nowPlayingBar = bar
}

func (a *App) View() string {
	var view string
	switch a.screen {
	case screenMainMenu:
		return a.viewMainMenu()
	case screenPlay:
		view = a.playScreen.View()
	case screenSearch:
		view = a.searchScreen.View()
	case screenList:
		view = a.listManagementScreen.View()
	case screenLucky:
		view = a.luckyScreen.View()
	case screenGist:
		view = a.gistScreen.View()
	case screenSettings:
		view = a.settingsScreen.View()
	case screenShuffleSettings:
		view = a.shuffleSettingsScreen.View()
	case screenConnectionSettings:
		view = a.connectionSettingsScreen.View()
	case screenAppearanceSettings:
		view = a.appearanceSettingsScreen.View()
	case screenBlocklist:
		view = a.blocklistScreen.View()
	case screenMostPlayed:
		view = a.mostPlayedScreen.View()
	case screenTopRated:
		view = a.topRatedScreen.View()
	case screenBrowseTags:
		view = a.browseTagsScreen.View()
	case screenTagPlaylists:
		view = a.tagPlaylistsScreen.View()
	case screenSleepSummary:
		view = a.sleepSummary.View()
	default:
		return "Unknown screen"
	}
	// Append the global now-playing bar on non-main-menu, non-player screens
	// when ContinueOnNavigate is active and a station has been handed off to App.
	// Player screens (Top Rated, Most Played, Play, Search, etc.) manage their
	// own playback UI, so the bar would be a duplicate there.
	// NOTE: The bar is now injected via PageLayout.NowPlaying above the help
	// text instead of appended here; this block is intentionally removed.
	return view
}

// stopScreenPlayers stops every screen-owned player without touching
// activePlayer or the app-level playingFromMain state. Use this when
// navigating to a new player screen so existing streams don't overlap,
// while still preserving a ContinueOnNavigate handoff.
//
// Important: a screen player pointer may be the same object as activePlayer
// when a handoff just occurred (the screen handed its player to App before
// navigating away). Stopping it here would kill the still-running stream, so
// we skip any player that is identical to activePlayer.
func (a *App) stopScreenPlayers() {
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
		if p != nil && p != a.activePlayer {
			_ = p.Stop()
		}
	}
}

// stopAllPlayback stops mpv on every screen that may be playing.
// App-level players (activePlayer, quickFavPlayer) are stopped unconditionally.
// Screen-owned players are only stopped when actively playing; calling Stop()
// on an idle player sets killed=true which would block the next PlayWithVolume
// call on that screen.
func (a *App) stopAllPlayback() {
	// App-level players: always stop so ContinueOnNavigate state is cleared.
	if a.activePlayer != nil {
		_ = a.activePlayer.Stop()
	}
	if a.quickFavPlayer != nil {
		_ = a.quickFavPlayer.Stop()
	}
	// Screen-owned players: only stop when actually playing.
	for _, p := range []*player.MPVPlayer{
		a.playScreen.player,
		a.searchScreen.player,
		a.luckyScreen.player,
		a.mostPlayedScreen.player,
		a.topRatedScreen.player,
		a.tagPlaylistsScreen.player,
		a.browseTagsScreen.player,
	} {
		if p != nil && p.IsPlaying() {
			_ = p.Stop()
		}
	}
	a.activePlayer = nil
	a.activeStation = nil
	a.activeContextLabel = ""
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

// updateQFViewOffset adjusts qfViewOffset so the currently selected QF entry
// is always visible in the clipped window. Call this after changing unifiedMenuIndex.
func (a *App) updateQFViewOffset(visibleCount ...int) {
	qfCount := len(a.quickFavorites)
	if qfCount == 0 {
		a.qfViewOffset = 0
		return
	}

	win := qfCount
	if len(visibleCount) > 0 && visibleCount[0] > 0 {
		win = visibleCount[0]
	}

	// Cursor position inside the QF list (negative = cursor is not in QF section)
	qfCursor := a.unifiedMenuIndex - mainMenuItemCount
	if qfCursor < 0 {
		a.qfViewOffset = 0
		return
	}
	if qfCursor >= qfCount {
		qfCursor = qfCount - 1
	}

	// Scroll down
	if qfCursor >= a.qfViewOffset+win {
		a.qfViewOffset = qfCursor - win + 1
	}
	// Scroll up
	if qfCursor < a.qfViewOffset {
		a.qfViewOffset = qfCursor
	}
	// Clamp
	if a.qfViewOffset < 0 {
		a.qfViewOffset = 0
	}
	if a.qfViewOffset > qfCount-win {
		a.qfViewOffset = qfCount - win
		if a.qfViewOffset < 0 {
			a.qfViewOffset = 0
		}
	}
}

// updateRPViewOffset adjusts rpViewOffset so the currently selected RP entry
// is always visible in the clipped window. Call this after changing unifiedMenuIndex.
func (a *App) updateRPViewOffset(visibleCount ...int) {
	rpStart := mainMenuItemCount + len(a.quickFavorites)
	rpCount := len(a.recentlyPlayed)
	if rpCount == 0 {
		a.rpViewOffset = 0
		return
	}

	// Determine how many RP rows are currently visible. The caller may supply
	// this value; otherwise fall back to the full list (conservative / safe).
	win := rpCount
	if len(visibleCount) > 0 && visibleCount[0] > 0 {
		win = visibleCount[0]
	}

	// Cursor position inside the RP list (may be negative = not in RP section)
	rpCursor := a.unifiedMenuIndex - rpStart
	if rpCursor < 0 {
		// Cursor is above the RP section — scroll to top
		a.rpViewOffset = 0
		return
	}
	if rpCursor >= rpCount {
		rpCursor = rpCount - 1
	}

	// Scroll down: keep cursor within window
	if rpCursor >= a.rpViewOffset+win {
		a.rpViewOffset = rpCursor - win + 1
	}
	// Scroll up
	if rpCursor < a.rpViewOffset {
		a.rpViewOffset = rpCursor
	}
	// Clamp
	if a.rpViewOffset < 0 {
		a.rpViewOffset = 0
	}
	if a.rpViewOffset > rpCount-win {
		a.rpViewOffset = rpCount - win
		if a.rpViewOffset < 0 {
			a.rpViewOffset = 0
		}
	}
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
		{"0", syncBackupMenuLabel},
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
	} else if a.activeStation != nil {
		// ContinueOnNavigate: station handed off from a player screen.
		content.WriteString("\n")
		content.WriteString(a.buildNowPlayingBannerText())
		content.WriteString("\n")
	}

	// Add quick play favorites if available (also part of unified navigation).
	// The list is viewport-clipped so it never pushes the menu off-screen on
	// small terminals (e.g. tmux panes). The calculation mirrors the RP section.
	if len(a.quickFavorites) > 0 {
		// ── viewport calculation ────────────────────────────────────────────
		// Lines already committed: header + chrome (title/subtitle/blank) +
		// menu items + now-playing block + blank + QF header + scroll indicators
		// + RP section (if non-empty) + volume display + footer.
		headerLines := visibleLineCount(renderHeader())
		p := getPadding()
		const (
			chromeLines   = 5 // same as RP calculation
			footerLines   = 2 // 1 margin + 1 help bar
			qfHeaderLines = 2 // blank + "─── Quick Play Favorites ───"
		)
		menuLinesQF := mainMenuItemCount
		if a.playingFromMain && a.playingStation != nil {
			menuLinesQF += 2
		} else if a.activeStation != nil {
			menuLinesQF += 2
		}
		// Reserve lines for RP section if it will be shown.
		rpReserve := 0
		if len(a.recentlyPlayed) > 0 {
			// At minimum: blank + RP header + 1 entry + possible scroll indicators
			rpReserve = 4
		}
		volumeLines := 0
		if a.volumeDisplay != "" {
			volumeLines = 2
		}
		fixed := headerLines + chromeLines + menuLinesQF + qfHeaderLines + rpReserve + volumeLines + footerLines + p.PageVertical
		visibleQF := a.height - fixed
		if len(a.quickFavorites) > visibleQF {
			// Reserve space for both scroll indicators.
			visibleQF -= 2
		}
		if visibleQF < 1 {
			visibleQF = 1
		}
		if visibleQF > len(a.quickFavorites) {
			visibleQF = len(a.quickFavorites)
		}

		// Cache window size and sync viewport (same pattern as RP).
		a.qfVisibleWindow = visibleQF
		a.updateQFViewOffset(visibleQF)

		content.WriteString("\n")
		content.WriteString(quickFavoritesStyle().Render("─── Quick Play Favorites ───"))
		content.WriteString("\n")

		// Scroll indicator: entries hidden above.
		if a.qfViewOffset > 0 {
			content.WriteString(dimStyle().Render(fmt.Sprintf("  ↑ %d more (↑/k to scroll)", a.qfViewOffset)))
			content.WriteString("\n")
		}

		menuItemCount := len(menuItems)
		endQF := a.qfViewOffset + visibleQF
		if endQF > len(a.quickFavorites) {
			endQF = len(a.quickFavorites)
		}
		for i := a.qfViewOffset; i < endQF; i++ {
			station := a.quickFavorites[i]
			shortcut := fmt.Sprintf("%d", 10+i)

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

			if a.playingFromMain && a.playingStation != nil && a.playingStation.StationUUID == station.StationUUID {
				playingLine := fmt.Sprintf("%s%s. ▶ %s", prefix, shortcut, stationInfo.String())
				if unifiedIdx == a.unifiedMenuIndex {
					content.WriteString(selectedItemStyle().Render(playingLine))
				} else {
					content.WriteString(normalItemStyle().Render(playingLine))
				}
			} else {
				lineContent := fmt.Sprintf("%s%s. %s", prefix, shortcut, stationInfo.String())
				if unifiedIdx == a.unifiedMenuIndex {
					content.WriteString(selectedItemStyle().Render(lineContent))
				} else {
					content.WriteString(normalItemStyle().Render(lineContent))
				}
			}
			content.WriteString("\n")
		}

		// Scroll indicator: entries hidden below.
		hiddenBelowQF := len(a.quickFavorites) - endQF
		if hiddenBelowQF > 0 {
			content.WriteString(dimStyle().Render(fmt.Sprintf("  ↓ %d more (↓/j to scroll)", hiddenBelowQF)))
			content.WriteString("\n")
		}
	}

	// Add recently played section if available, clipped to available screen space.
	if len(a.recentlyPlayed) > 0 {
		// ── viewport calculation ────────────────────────────────────────────
		// Lines already committed: header + chrome (title/subtitle/blank) +
		// menu items + blank + QF header + QF items + blank + RP header + 1 margin + footer.
		headerLines := visibleLineCount(renderHeader())
		p := getPadding()
		const (
			// assemblePageContent always emits 3 lines before Content:
			//   \n (blank-after-header) + title\n + \n (empty subtitle line)
			// Content itself starts with "Choose an option:\n\n" = 2 more lines.
			chromeLines   = 5 // 3 (assemblePageContent) + 2 ("Choose an option:\n\n")
			footerLines   = 2 // 1 margin + 1 help bar
			rpHeaderLines = 2 // blank + "─── Recently Played ───"
		)
		menuLines := mainMenuItemCount // one line per menu item
		if a.playingFromMain && a.playingStation != nil {
			menuLines += 2 // now-playing block: blank + station line
		} else if a.activeStation != nil {
			menuLines += 2 // ContinueOnNavigate banner: blank + banner line
		}
		qfLines := 0
		if len(a.quickFavorites) > 0 {
			// QF is now viewport-clipped; use the cached visible window size.
			// Add 2 for potential scroll indicators (conservative upper bound).
			qfLines = 2 + a.qfVisibleWindow // blank + header + visible items
			if a.qfViewOffset > 0 {
				qfLines++ // top scroll indicator
			}
			if len(a.quickFavorites) > a.qfViewOffset+a.qfVisibleWindow {
				qfLines++ // bottom scroll indicator
			}
		}
		volumeLines := 0
		if a.volumeDisplay != "" {
			// blank line before the banner + banner line itself.
			// The final blank line before the help bar is already covered by footerLines.
			volumeLines = 2
		}
		// p.PageVertical is the bottom padding added by docStyleNoTopPadding;
		// RenderPageWithBottomHelp subtracts it, so we must account for it here
		// to avoid inflating visibleRP and overflowing the terminal height.
		fixed := headerLines + chromeLines + menuLines + qfLines + volumeLines + rpHeaderLines + footerLines + p.PageVertical
		visibleRP := a.height - fixed
		// Reserve lines for scroll indicators conservatively (both can appear).
		// We check before clamping whether the list could ever need scrolling.
		// Checking rpViewOffset > 0 here would be premature: updateRPViewOffset
		// is called below and may change rpViewOffset from 0 to a positive value
		// (e.g. after a resize), causing the up-indicator space to be missed.
		if len(a.recentlyPlayed) > visibleRP {
			// List is scrollable — reserve space for both indicators.
			const maxIndicatorLines = 2
			visibleRP -= maxIndicatorLines
		}
		if visibleRP < 1 {
			visibleRP = 1
		}
		// If the user has set a display row cap, honour it.
		if a.playHistoryCfg.DisplayRows > 0 && visibleRP > a.playHistoryCfg.DisplayRows {
			visibleRP = a.playHistoryCfg.DisplayRows
		}
		if visibleRP > len(a.recentlyPlayed) {
			visibleRP = len(a.recentlyPlayed)
		}

		// NOTE: We mutate state here (unusual in View) because the accurate
		// visible window size is only known at render time and must be cached
		// for key-event handlers to use during navigation.
		a.rpVisibleWindow = visibleRP
		// Keep viewport in sync with cursor now that we know the window size.
		a.updateRPViewOffset(visibleRP)

		content.WriteString("\n")
		content.WriteString(quickFavoritesStyle().Render("─── Recently Played ───"))
		content.WriteString("\n")

		// Scroll indicator: show how many entries are hidden above.
		if a.rpViewOffset > 0 {
			content.WriteString(dimStyle().Render(fmt.Sprintf("  ↑ %d more (↑/k to scroll)", a.rpViewOffset)))
			content.WriteString("\n")
		}

		rpShortcutStart := 10 + len(a.quickFavorites)
		end := a.rpViewOffset + visibleRP
		if end > len(a.recentlyPlayed) {
			end = len(a.recentlyPlayed)
		}
		for i := a.rpViewOffset; i < end; i++ {
			entry := a.recentlyPlayed[i]
			shortcut := fmt.Sprintf("%d", rpShortcutStart+i)

			var stationInfo strings.Builder
			stationInfo.WriteString(entry.Station.TrimName())
			if entry.Station.Country != "" {
				stationInfo.WriteString(" • ")
				stationInfo.WriteString(entry.Station.Country)
			}
			if entry.Metadata != nil {
				stationInfo.WriteString(" • ")
				stationInfo.WriteString(storage.FormatLastPlayed(entry.Metadata.LastPlayed))
			}

			unifiedIdx := mainMenuItemCount + len(a.quickFavorites) + i
			prefix := "  "
			if unifiedIdx == a.unifiedMenuIndex {
				prefix = "> "
			}

			if a.playingFromMain && a.playingStation != nil &&
				a.playingStation.StationUUID == entry.Station.StationUUID {
				playingLine := fmt.Sprintf("%s%s. ▶ %s", prefix, shortcut, stationInfo.String())
				if unifiedIdx == a.unifiedMenuIndex {
					content.WriteString(selectedItemStyle().Render(playingLine))
				} else {
					content.WriteString(normalItemStyle().Render(playingLine))
				}
			} else {
				lineContent := fmt.Sprintf("%s%s. %s", prefix, shortcut, stationInfo.String())
				if unifiedIdx == a.unifiedMenuIndex {
					content.WriteString(selectedItemStyle().Render(lineContent))
				} else {
					content.WriteString(normalItemStyle().Render(lineContent))
				}
			}
			content.WriteString("\n")
		}

		// Scroll indicator: show how many entries are hidden below.
		hiddenBelow := len(a.recentlyPlayed) - end
		if hiddenBelow > 0 {
			content.WriteString(dimStyle().Render(fmt.Sprintf("  ↓ %d more (↓/j to scroll)", hiddenBelow)))
			content.WriteString("\n")
		}
	}

	// Add volume display if visible
	if a.volumeDisplay != "" {
		content.WriteString("\n")
		content.WriteString(highlightStyle().Render(a.volumeDisplay))
		content.WriteString("\n")
	}

	// Guaranteed blank line margin before the footer help bar.
	content.WriteString("\n")

	// Build help text based on playing state
	var helpText string
	if a.playingFromMain {
		helpText = "↑↓/jk: Navigate • Enter: Select • /*: Volume • m: Mute • Esc: Stop • ?: Help"
	} else {
		helpText = "↑↓/jk: Navigate • Enter: Select • 1-0/-: Menu shortcuts • 10+: QF/RP • ?: Help"
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
