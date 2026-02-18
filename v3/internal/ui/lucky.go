package ui

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/shinokada/tera/v3/internal/api"
	"github.com/shinokada/tera/v3/internal/blocklist"
	"github.com/shinokada/tera/v3/internal/player"
	"github.com/shinokada/tera/v3/internal/shuffle"
	"github.com/shinokada/tera/v3/internal/storage"
	"github.com/shinokada/tera/v3/internal/theme"
	"github.com/shinokada/tera/v3/internal/ui/components"
)

// luckyState represents the current state in the lucky screen
type luckyState int

const (
	luckyStateInput luckyState = iota
	luckyStateSearching
	luckyStatePlaying
	luckyStateShufflePlaying
	luckyStateSavePrompt
	luckyStateSelectList
	luckyStateNewListInput
)

// LuckyModel represents the I Feel Lucky screen
type LuckyModel struct {
	state           luckyState
	apiClient       *api.Client
	textInput       textinput.Model
	newListInput    textinput.Model
	menuList        list.Model // Menu for history navigation
	numberBuffer    string     // Buffer for multi-digit number input
	selectedStation *api.Station
	player          *player.MPVPlayer
	favoritePath    string
	searchHistory   *storage.SearchHistoryStore
	saveMessage     string
	saveMessageTime int
	width           int
	height          int
	err             error
	availableLists  []string
	listItems       []list.Item
	listModel       list.Model
	helpModel       components.HelpModel
	votedStations   *storage.VotedStations // Track voted stations
	// Shuffle mode fields
	shuffleEnabled    bool
	shuffleManager    *shuffle.Manager
	shuffleConfig     storage.ShuffleConfig
	allStations       []api.Station // All stations from search for shuffle
	lastSearchKeyword string        // Keyword used for current shuffle session
	blocklistManager  *blocklist.Manager
	metadataManager   *storage.MetadataManager // Track play statistics
	lastBlockTime     time.Time
	// Star rating fields
	ratingsManager *storage.RatingsManager
	starRenderer   *components.StarRenderer
	ratingMode     bool // true when waiting for 1-5 input after pressing R
}

// Messages for lucky screen
type luckySearchResultsMsg struct {
	station *api.Station
}

type luckyShuffleSearchResultsMsg struct {
	stations []api.Station
	keyword  string
}

type luckySearchErrorMsg struct {
	err error
}

type shuffleTimerTickMsg struct{}

type shuffleAdvanceMsg struct{}

type saveToListSuccessMsg struct {
	listName    string
	stationName string
}

type saveToListFailedMsg struct {
	err         error
	isDuplicate bool
}

type luckyPlaybackStalledMsg struct {
	station api.Station
}

type luckyCheckSignalMsg struct {
	station api.Station
	attempt int
}

// NewLuckyModel creates a new lucky screen model
func NewLuckyModel(apiClient *api.Client, favoritePath string, blocklistManager *blocklist.Manager) LuckyModel {
	ti := textinput.New()
	ti.Placeholder = "rock, jazz, classical, meditation..."
	ti.CharLimit = 50
	ti.Width = 50
	ti.Focus()

	// New list input
	nli := textinput.New()
	nli.Placeholder = "Enter new list name..."
	nli.CharLimit = 50
	nli.Width = 50

	// Load search history (if it fails, just use empty history)
	store := storage.NewStorage(favoritePath)
	history, err := store.LoadSearchHistory(context.Background())
	if err != nil || history == nil {
		history = storage.NewSearchHistoryStore()
	}

	// Load shuffle config
	shuffleConfig, err := storage.LoadShuffleConfig()
	if err != nil {
		shuffleConfig = storage.DefaultShuffleConfig()
	}

	// Load voted stations
	votedStations, err := storage.LoadVotedStations()
	if err != nil {
		// If we can't load, create empty list
		votedStations = &storage.VotedStations{Stations: []storage.VotedStation{}}
	}

	m := LuckyModel{
		state:            luckyStateInput,
		apiClient:        apiClient,
		textInput:        ti,
		newListInput:     nli,
		favoritePath:     favoritePath,
		player:           player.NewMPVPlayer(),
		searchHistory:    history,
		width:            80,
		height:           24,
		helpModel:        components.NewHelpModel(components.CreatePlayingHelp()),
		votedStations:    votedStations,
		shuffleEnabled:   false,
		shuffleConfig:    shuffleConfig,
		blocklistManager: blocklistManager,
	}

	// Build menu with history items
	m.rebuildMenuWithHistory()

	return m
}

// Init initializes the lucky screen
func (m LuckyModel) Init() tea.Cmd {
	return tea.Batch(textinput.Blink, tickEverySecond())
}

// Update handles messages for the lucky screen
func (m LuckyModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.helpModel.IsVisible() {
			var cmd tea.Cmd
			m.helpModel, cmd = m.helpModel.Update(msg)
			return m, cmd
		}

		switch m.state {
		case luckyStateInput:
			return m.updateInput(msg)
		case luckyStatePlaying:
			return m.updatePlaying(msg)
		case luckyStateShufflePlaying:
			return m.updateShufflePlaying(msg)
		case luckyStateSavePrompt:
			return m.updateSavePrompt(msg)
		case luckyStateSelectList:
			return m.updateSelectList(msg)
		case luckyStateNewListInput:
			return m.updateNewListInput(msg)
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		m.helpModel.SetSize(msg.Width, msg.Height)

		return m, nil

	case luckySearchResultsMsg:
		m.selectedStation = msg.station
		m.state = luckyStatePlaying
		// Start playback immediately
		return m, m.startPlayback()

	case luckyShuffleSearchResultsMsg:
		// Shuffle mode - initialize shuffle manager with all stations
		m.allStations = msg.stations
		m.lastSearchKeyword = msg.keyword

		// Reload config in case it was updated
		config, err := storage.LoadShuffleConfig()
		if err == nil {
			m.shuffleConfig = config
		}

		// Create shuffle manager
		m.shuffleManager = shuffle.NewManager(m.shuffleConfig)
		if err := m.shuffleManager.Initialize(msg.keyword, msg.stations); err != nil {
			m.err = err
			m.state = luckyStateInput
			return m, nil
		}

		// Get first station
		station, err := m.shuffleManager.GetCurrentStation()
		if err != nil {
			m.err = err
			m.state = luckyStateInput
			return m, nil
		}

		m.selectedStation = station
		m.state = luckyStateShufflePlaying

		// Start playback and timer
		return m, tea.Batch(
			m.startPlayback(),
			m.shuffleTimerTick(),
		)

	case luckySearchErrorMsg:
		m.err = msg.err
		m.state = luckyStateInput
		return m, nil

	case playbackStartedMsg:
		// Playback started successfully
		return m, nil

	case playbackStoppedMsg:
		// Playback stopped, show save prompt
		m.state = luckyStateSavePrompt
		return m, nil

	case playbackErrorMsg:
		m.err = msg.err
		m.state = luckyStateSavePrompt
		return m, nil

	case luckyPlaybackStalledMsg:
		// Stop player if it's still "playing" (but silent)
		if m.player != nil {
			_ = m.player.Stop()
		}
		m.saveMessage = "✗ No signal detected"
		m.saveMessageTime = messageDisplayShort
		m.state = luckyStateInput
		m.selectedStation = nil
		return m, nil

	case luckyCheckSignalMsg:
		if (m.state == luckyStatePlaying || m.state == luckyStateShufflePlaying) && m.selectedStation != nil && m.selectedStation.StationUUID == msg.station.StationUUID {
			return m, m.checkPlaybackSignal(msg.station, msg.attempt)
		}
		return m, nil

	case saveSuccessMsg:
		m.saveMessage = fmt.Sprintf("✓ Saved '%s' to Quick Favorites", msg.station.TrimName())
		m.saveMessageTime = messageDisplayShort
		return m, nil

	case saveFailedMsg:
		if msg.isDuplicate {
			m.saveMessage = "Already in Quick Favorites"
		} else {
			m.saveMessage = fmt.Sprintf("✗ Failed to save: %v", msg.err)
		}
		m.saveMessageTime = messageDisplayShort
		return m, nil

	case components.VoteSuccessMsg:
		m.saveMessage = fmt.Sprintf("✓ %s", msg.Message)
		m.saveMessageTime = messageDisplayShort
		return m, nil

	case components.VoteFailedMsg:
		m.saveMessage = fmt.Sprintf("✗ Vote failed: %v", msg.Err)
		m.saveMessageTime = messageDisplayShort
		return m, nil

	case listsLoadedMsg:
		m.availableLists = msg.lists
		m.listItems = make([]list.Item, len(msg.lists))
		for i, name := range msg.lists {
			m.listItems[i] = playListItem{name: name}
		}

		// Initialize list model if we have dimensions
		if m.width > 0 && m.height > 0 {
			m.initializeListModel()
		}
		return m, nil

	case saveToListSuccessMsg:
		m.saveMessage = fmt.Sprintf("✓ Saved '%s' to %s", msg.stationName, msg.listName)
		m.saveMessageTime = messageDisplayShort
		m.state = luckyStatePlaying
		return m, nil

	case saveToListFailedMsg:
		if msg.isDuplicate {
			m.saveMessage = "Already in this list"
		} else {
			m.saveMessage = fmt.Sprintf("✗ Failed to save: %v", msg.err)
		}
		m.saveMessageTime = messageDisplayShort
		m.state = luckyStatePlaying
		return m, nil

	case shuffleTimerTickMsg:
		// Handle shuffle timer tick
		if m.shuffleManager != nil && m.state == luckyStateShufflePlaying {
			if m.shuffleManager.UpdateTimer() {
				// Timer expired, advance to next station
				return m, func() tea.Msg {
					return shuffleAdvanceMsg{}
				}
			}
			// Continue timer
			return m, m.shuffleTimerTick()
		}
		return m, nil

	case shuffleAdvanceMsg:
		// Auto-advance to next shuffle station
		if m.shuffleManager != nil && m.state == luckyStateShufflePlaying {
			nextStation, err := m.shuffleManager.Next(func(s api.Station) bool {
				if m.blocklistManager == nil {
					return true
				}
				return !m.blocklistManager.IsBlockedByAny(&s)
			})
			if err != nil {
				m.saveMessage = fmt.Sprintf("✗ Shuffle error: %v", err)
				m.saveMessageTime = messageDisplayLong
				return m, m.shuffleTimerTick()
			}

			m.selectedStation = nextStation
			// Stop current playback and start new station
			_ = m.player.Stop() // Ignore error, we're starting new playback anyway

			return m, tea.Batch(
				m.startPlayback(),
				m.shuffleTimerTick(),
			)
		}
		return m, nil

	case tickMsg:
		// Countdown save message
		if m.saveMessageTime > 0 {
			m.saveMessageTime--
			if m.saveMessageTime == 0 {
				m.saveMessage = ""
			}
		}
		return m, tickEverySecond()

	case stationBlockedMsg:
		m.lastBlockTime = time.Now()

		if msg.success {
			// Stop playback
			if m.player != nil {
				_ = m.player.Stop()
			}

			// Show message
			m.saveMessage = msg.message + " (press 'u' within 5s to undo)"
			m.saveMessageTime = messageDisplayMedium

			// If shuffle is active, advance to next station
			if m.state == luckyStateShufflePlaying && m.shuffleManager != nil {
				return m, func() tea.Msg {
					return shuffleAdvanceMsg{}
				}
			}

			// Return to input
			m.state = luckyStateInput
			m.selectedStation = nil
			// Reload history
			m.reloadSearchHistory()
			m.rebuildMenuWithHistory()
		} else {
			// Already blocked
			m.saveMessage = msg.message
			m.saveMessageTime = messageDisplayMedium
		}
		return m, nil

	case undoBlockSuccessMsg:
		m.saveMessage = "✓ Block undone"
		m.saveMessageTime = messageDisplayMedium
		return m, nil

	case undoBlockFailedMsg:
		m.saveMessage = "No recent block to undo"
		m.saveMessageTime = messageDisplayMedium
		return m, nil
	}

	var cmd tea.Cmd
	if m.state == luckyStateInput {
		m.textInput, cmd = m.textInput.Update(msg)
	}
	return m, cmd
}

// updateInput handles input during the input state
func (m LuckyModel) updateInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	// Handle escape - return to main menu
	if key == "esc" {
		m.numberBuffer = "" // Clear buffer
		return m, func() tea.Msg {
			return navigateMsg{screen: screenMainMenu}
		}
	}

	// Handle ctrl+c
	if key == "ctrl+c" {
		return m, tea.Quit
	}

	// Handle 't' key to toggle shuffle mode (when not typing in input)
	if key == "t" && m.textInput.Value() == "" {
		m.shuffleEnabled = !m.shuffleEnabled
		return m, nil
	}

	// Check if text input is focused (has content being typed)
	inputFocused := m.textInput.Focused() && m.textInput.Value() != ""

	// If text input has content, handle normally
	if inputFocused {
		if key == "enter" {
			keyword := strings.TrimSpace(m.textInput.Value())
			if keyword == "" {
				m.err = fmt.Errorf("please enter a keyword")
				return m, nil
			}
			m.err = nil
			m.state = luckyStateSearching

			// Use shuffle search if enabled
			if m.shuffleEnabled {
				return m, m.searchForShuffle(keyword)
			}
			return m, m.searchAndPickRandom(keyword)
		}
		var cmd tea.Cmd
		m.textInput, cmd = m.textInput.Update(msg)
		return m, cmd
	}

	// Handle number input for multi-digit selection (only when not typing in input)
	if key >= "0" && key <= "9" {
		m.numberBuffer += key

		// Check if we should auto-select
		if len(m.numberBuffer) >= 2 {
			var num int
			_, _ = fmt.Sscanf(m.numberBuffer, "%d", &num)

			// Calculate max valid number
			maxNum := 0
			if m.searchHistory != nil {
				maxHistoryItems := len(m.searchHistory.LuckyQueries)
				if maxHistoryItems > m.searchHistory.MaxSize {
					maxHistoryItems = m.searchHistory.MaxSize
				}
				maxNum = maxHistoryItems
			}

			// If number is valid, select it
			if num >= 1 && num <= maxNum {
				m.numberBuffer = "" // Clear buffer
				return m.selectHistoryByNumber(num)
			}

			// If number is too large or 3+ digits, clear buffer
			if num > maxNum || len(m.numberBuffer) >= 3 {
				m.numberBuffer = ""
			}
		}
		return m, nil
	}

	// Handle Enter to submit buffered number or search
	if key == "enter" {
		if m.numberBuffer != "" {
			var num int
			_, _ = fmt.Sscanf(m.numberBuffer, "%d", &num)
			m.numberBuffer = ""
			return m.selectHistoryByNumber(num)
		}

		// Check if a history item is selected in the menu (index 0 is separator, 1+ are history items)
		selectedIdx := m.menuList.Index()
		if m.searchHistory != nil && selectedIdx > 0 && selectedIdx <= len(m.searchHistory.LuckyQueries) {
			query := m.searchHistory.LuckyQueries[selectedIdx-1] // -1 because index 0 is separator
			m.state = luckyStateSearching
			m.err = nil
			if m.shuffleEnabled {
				return m, m.searchForShuffle(query)
			}
			return m, m.searchAndPickRandom(query)
		}

		// Search with text input value
		keyword := strings.TrimSpace(m.textInput.Value())
		if keyword == "" {
			m.err = fmt.Errorf("please enter a keyword")
			return m, nil
		}
		m.err = nil
		m.state = luckyStateSearching
		if m.shuffleEnabled {
			return m, m.searchForShuffle(keyword)
		}
		return m, m.searchAndPickRandom(keyword)
	}

	// Clear buffer on navigation keys
	if key == "up" || key == "down" || key == "j" || key == "k" {
		m.numberBuffer = ""
	}

	// Handle menu navigation for history items
	newList, selected := components.HandleMenuKey(msg, m.menuList)
	m.menuList = newList

	if selected > 0 { // selected > 0 because index 0 is separator
		// Selected a history item from menu
		actualIndex := selected - 1 // Adjust for separator at index 0
		if m.searchHistory != nil && actualIndex >= 0 && actualIndex < len(m.searchHistory.LuckyQueries) {
			query := m.searchHistory.LuckyQueries[actualIndex]
			m.state = luckyStateSearching
			m.err = nil
			return m, m.searchAndPickRandom(query)
		}
	}

	// Update text input
	// Clear number buffer when user types into text input (non-navigation, non-number)
	if m.numberBuffer != "" {
		m.numberBuffer = ""
	}
	var cmd tea.Cmd
	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

// handleVolumeControl handles volume-related key presses
func (m *LuckyModel) handleVolumeControl(key string) (bool, string) {
	switch key {
	case "/":
		// Decrease volume
		newVol := m.player.DecreaseVolume(5)
		if m.selectedStation != nil && newVol >= 0 {
			m.selectedStation.SetVolume(newVol)
			m.saveStationVolume(m.selectedStation)
		}
		return true, fmt.Sprintf("Volume: %d%%", newVol)
	case "*":
		// Increase volume
		newVol := m.player.IncreaseVolume(5)
		if m.selectedStation != nil {
			m.selectedStation.SetVolume(newVol)
			m.saveStationVolume(m.selectedStation)
		}
		return true, fmt.Sprintf("Volume: %d%%", newVol)
	case "m":
		// Toggle mute
		muted, vol := m.player.ToggleMute()
		if m.selectedStation != nil && !muted && vol >= 0 {
			m.selectedStation.SetVolume(vol)
			m.saveStationVolume(m.selectedStation)
		}
		if muted {
			return true, "Volume: Muted"
		}
		return true, fmt.Sprintf("Volume: %d%%", vol)
	}
	return false, ""
}

// selectHistoryByNumber selects a history item by number (1-based)
func (m LuckyModel) selectHistoryByNumber(num int) (tea.Model, tea.Cmd) {
	if m.searchHistory == nil {
		return m, nil
	}

	actualIndex := num - 1 // 1 = index 0, 2 = index 1, etc.
	if actualIndex >= 0 && actualIndex < len(m.searchHistory.LuckyQueries) {
		// Update menu selection (add 1 to account for separator at index 0)
		m.menuList.Select(actualIndex + 1)
		query := m.searchHistory.LuckyQueries[actualIndex]
		m.state = luckyStateSearching
		m.err = nil
		if m.shuffleEnabled {
			return m, m.searchForShuffle(query)
		}
		return m, m.searchAndPickRandom(query)
	}
	return m, nil
}

// updatePlaying handles input during playback
func (m LuckyModel) updatePlaying(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Handle rating mode input first
	if m.ratingMode {
		return m.handleRatingModeInput(msg)
	}

	switch msg.String() {
	case "b":
		// Block current station
		if m.selectedStation != nil {
			return m, m.blockStation()
		}
		return m, nil
	case "u":
		// Undo last block (within 5 seconds)
		if time.Since(m.lastBlockTime) < 5*time.Second {
			return m, m.undoLastBlock()
		}
		return m, nil
	case "esc":
		// Stop playback and return to I Feel Lucky input
		if err := m.player.Stop(); err != nil {
			m.saveMessage = fmt.Sprintf("✗ Failed to stop playback: %v", err)
			m.saveMessageTime = messageDisplayLong
			return m, nil
		}
		m.state = luckyStateInput
		m.selectedStation = nil
		// Reload history from disk so recent search appears
		m.reloadSearchHistory()
		m.rebuildMenuWithHistory()
		return m, nil
	case "0":
		// Return to main menu (Level 2+ shortcut)
		if err := m.player.Stop(); err != nil {
			m.saveMessage = fmt.Sprintf("✗ Failed to stop playback: %v", err)
			m.saveMessageTime = messageDisplayLong
			return m, nil
		}
		m.selectedStation = nil
		m.state = luckyStateInput
		return m, func() tea.Msg {
			return navigateMsg{screen: screenMainMenu}
		}
	case "f":
		// Save to Quick Favorites during playback
		return m, m.saveToQuickFavorites()
	case "s":
		// Save to a list - show list selection
		m.state = luckyStateSelectList
		return m, m.loadAvailableLists()
	case "v":
		// Vote for this station
		return m, m.voteForStation()
	case " ":
		// Toggle pause/play with space bar
		if err := m.player.TogglePause(); err != nil {
			m.saveMessage = fmt.Sprintf("✗ Pause failed: %v", err)
			m.saveMessageTime = messageDisplayLong
		} else {
			if m.player.IsPaused() {
				m.saveMessage = "⏸ Paused"
			} else {
				m.saveMessage = "▶ Resumed"
			}
			m.saveMessageTime = messageDisplayShort
		}
		return m, nil
	case "r":
		// Enter rating mode
		if m.selectedStation != nil && m.ratingsManager != nil {
			m.ratingMode = true
			m.saveMessage = "Press 1-5 to rate, 0 to remove rating"
			m.saveMessageTime = -1 // Persistent until action
			return m, nil
		}
		return m, nil
	case "/", "*", "m":
		if handled, msg := m.handleVolumeControl(msg.String()); handled {
			m.saveMessage = msg
			m.saveMessageTime = messageDisplayShort
			return m, nil
		}
	case "?":
		m.helpModel.SetSize(m.width, m.height)
		m.helpModel.Toggle()
		return m, nil
	}
	return m, nil
}

// handleRatingModeInput handles input when in rating mode for lucky screen
func (m LuckyModel) handleRatingModeInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	m.ratingMode = false // Exit rating mode regardless of key

	if m.selectedStation == nil || m.ratingsManager == nil {
		return m, nil
	}

	key := msg.String()

	// Handle rating keys 1-5
	if len(key) == 1 && key[0] >= '1' && key[0] <= '5' {
		rating := int(key[0] - '0')
		if err := m.ratingsManager.SetRating(m.selectedStation, rating); err == nil {
			stars := ""
			if m.starRenderer != nil {
				stars = m.starRenderer.RenderCompactPlain(rating) + " "
			}
			m.saveMessage = fmt.Sprintf("✓ %sRated!", stars)
			m.saveMessageTime = messageDisplayShort
		}
		return m, nil
	}

	// Handle remove rating (0 or r)
	if key == "0" || key == "r" {
		_ = m.ratingsManager.RemoveRating(m.selectedStation.StationUUID)
		m.saveMessage = "✓ Rating removed"
		m.saveMessageTime = messageDisplayShort
		return m, nil
	}

	// Any other key - just clear the message
	m.saveMessage = ""
	m.saveMessageTime = 0
	return m, nil
}

// updateSelectList handles input during list selection
func (m LuckyModel) updateSelectList(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		// Cancel and go back to playing
		m.state = luckyStatePlaying
		return m, nil
	case "n":
		// Create new list
		m.state = luckyStateNewListInput
		m.newListInput.SetValue("")
		m.newListInput.Focus()
		return m, textinput.Blink
	case "enter":
		// Save to selected list
		if i, ok := m.listModel.SelectedItem().(playListItem); ok {
			return m, m.saveToList(i.name)
		}
	}

	var cmd tea.Cmd
	m.listModel, cmd = m.listModel.Update(msg)
	return m, cmd
}

// updateNewListInput handles input for new list name entry
func (m LuckyModel) updateNewListInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		// Cancel and go back to list selection
		m.state = luckyStateSelectList
		return m, nil
	case "enter":
		// Save to new list
		listName := strings.TrimSpace(m.newListInput.Value())
		if listName == "" {
			return m, nil
		}
		// saveToList handles both existing and new lists
		return m, m.saveToList(listName)
	}

	var cmd tea.Cmd
	m.newListInput, cmd = m.newListInput.Update(msg)
	return m, cmd
}

// loadAvailableLists loads all favorite lists
func (m LuckyModel) loadAvailableLists() tea.Cmd {
	return func() tea.Msg {
		store := storage.NewStorage(m.favoritePath)
		lists, err := store.GetAllLists(context.Background())
		if err != nil {
			return saveToListFailedMsg{err: fmt.Errorf("failed to load lists: %w", err)}
		}
		return listsLoadedMsg{lists: lists}
	}
}

// initializeListModel creates the list model with current dimensions
func (m *LuckyModel) initializeListModel() {
	listHeight := m.height - 14
	if listHeight < 5 {
		listHeight = 5
	}

	delegate := createStyledDelegate()

	m.listModel = list.New(m.listItems, delegate, m.width, listHeight)
	m.listModel.Title = "" // No title - we'll add it in the view
	m.listModel.SetShowStatusBar(false)
	m.listModel.SetFilteringEnabled(false)
	m.listModel.SetShowHelp(false)
	m.listModel.Styles.Title = titleStyle()
	m.listModel.Styles.PaginationStyle = paginationStyle()
	m.listModel.Styles.HelpStyle = helpStyle()
}

// saveToList saves the current station to a specific list
func (m LuckyModel) saveToList(listName string) tea.Cmd {
	return func() tea.Msg {
		if m.selectedStation == nil {
			return saveToListFailedMsg{err: fmt.Errorf("no station selected")}
		}

		store := storage.NewStorage(m.favoritePath)
		err := store.AddStation(context.Background(), listName, *m.selectedStation)

		if err != nil {
			if err == storage.ErrDuplicateStation {
				return saveToListFailedMsg{err: err, isDuplicate: true}
			}
			return saveToListFailedMsg{err: err}
		}

		return saveToListSuccessMsg{
			listName:    listName,
			stationName: m.selectedStation.TrimName(),
		}
	}
}

// updateSavePrompt handles input during save prompt
func (m LuckyModel) updateSavePrompt(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "n", "2":
		// Don't save, return to main menu
		return m, func() tea.Msg {
			return navigateMsg{screen: screenMainMenu}
		}
	case "y", "1":
		// Save to Quick Favorites and return to main menu
		cmd := m.saveToQuickFavorites()
		// Wait a moment to show the save message, then return to main menu
		return m, tea.Batch(cmd, tea.Tick(2*time.Second, func(t time.Time) tea.Msg {
			return navigateMsg{screen: screenMainMenu}
		}))
	}
	return m, nil
}

// searchAndPickRandom searches for stations and picks one randomly
func (m LuckyModel) searchAndPickRandom(keyword string) tea.Cmd {
	return func() tea.Msg {
		// Save to history in background
		go func() {
			store := storage.NewStorage(m.favoritePath)
			_ = store.AddLuckyQuery(context.Background(), keyword)
		}()

		// Search by tag (genre/keyword)
		stations, err := m.apiClient.SearchByTag(context.Background(), keyword)
		if err != nil {
			return luckySearchErrorMsg{err: fmt.Errorf("search failed: %w", err)}
		}

		if len(stations) == 0 {
			return luckySearchErrorMsg{err: fmt.Errorf("no stations found for '%s'", keyword)}
		}

		// Filter out blocked stations
		if m.blocklistManager != nil {
			filtered := make([]api.Station, 0, len(stations))
			for _, s := range stations {
				if !m.blocklistManager.IsBlockedByAny(&s) {
					filtered = append(filtered, s)
				}
			}
			stations = filtered
		}

		if len(stations) == 0 {
			return luckySearchErrorMsg{err: fmt.Errorf("all stations found for '%s' are blocked", keyword)}
		}

		// Pick a random station (rand is auto-seeded since Go 1.20)
		randomIndex := rand.Intn(len(stations))
		selectedStation := stations[randomIndex]

		return luckySearchResultsMsg{station: &selectedStation}
	}
}

// startPlayback initiates playback of the selected station
func (m LuckyModel) startPlayback() tea.Cmd {
	if m.selectedStation == nil {
		return func() tea.Msg {
			return playbackErrorMsg{fmt.Errorf("no station selected")}
		}
	}
	station := *m.selectedStation
	return tea.Batch(
		func() tea.Msg {
			if err := m.player.Play(&station); err != nil {
				return playbackErrorMsg{err}
			}
			return playbackStartedMsg{}
		},
		m.checkPlaybackSignal(station, 1),
	)
}

// checkPlaybackSignal checks for audio bitrate to ensure the station is actually playing
func (m LuckyModel) checkPlaybackSignal(station api.Station, attempt int) tea.Cmd {
	return tea.Tick(2*time.Second, func(t time.Time) tea.Msg {
		if m.player == nil || !m.player.IsPlaying() {
			return nil
		}

		// Check for audio bitrate
		bitrate, err := m.player.GetAudioBitrate()
		if err == nil && bitrate > 0 {
			// Signal detected via bitrate!
			return playbackStartedMsg{}
		}

		// Also check for media-title as fallback (some streams don't report bitrate)
		if track, err := m.player.GetCurrentTrack(); err == nil && track != "" {
			// Signal detected via media title!
			return playbackStartedMsg{}
		}

		if attempt >= 4 { // 4 attempts * 2 seconds = 8 seconds
			return luckyPlaybackStalledMsg{station: station}
		}

		return luckyCheckSignalMsg{station: station, attempt: attempt + 1}
	})
}

// saveToQuickFavorites saves the current station to My-favorites.json
func (m LuckyModel) saveToQuickFavorites() tea.Cmd {
	return func() tea.Msg {
		if m.selectedStation == nil {
			return saveFailedMsg{err: fmt.Errorf("no station selected")}
		}

		store := storage.NewStorage(m.favoritePath)
		err := store.AddStation(context.Background(), "My-favorites", *m.selectedStation)

		if err != nil {
			if err == storage.ErrDuplicateStation {
				return saveFailedMsg{err: err, isDuplicate: true}
			}
			return saveFailedMsg{err: err}
		}

		return saveSuccessMsg{station: m.selectedStation}
	}
}

// voteForStation votes for the currently playing station
func (m LuckyModel) voteForStation() tea.Cmd {
	return components.ExecuteVote(m.selectedStation, m.votedStations, m.apiClient)
}

// saveStationVolume saves the updated volume for a station in My-favorites
func (m LuckyModel) saveStationVolume(station *api.Station) {
	if station == nil {
		return
	}

	store := storage.NewStorage(m.favoritePath)
	// Save to My-favorites if the station exists there
	list, err := store.LoadList(context.Background(), "My-favorites")
	if err != nil {
		return
	}

	for i := range list.Stations {
		if list.Stations[i].StationUUID == station.StationUUID {
			list.Stations[i].Volume = station.Volume
			break
		}
	}
	_ = store.SaveList(context.Background(), list)
}

// View renders the lucky screen
func (m LuckyModel) View() string {
	if m.helpModel.IsVisible() {
		return m.helpModel.View()
	}

	switch m.state {
	case luckyStateInput:
		return m.viewInput()
	case luckyStateSearching:
		return m.viewSearching()
	case luckyStatePlaying:
		return m.viewPlaying()
	case luckyStateShufflePlaying:
		return m.viewShufflePlaying()
	case luckyStateSavePrompt:
		return m.viewSavePrompt()
	case luckyStateSelectList:
		return m.viewSelectList()
	case luckyStateNewListInput:
		return m.viewNewListInput()
	}
	return "Unknown state"
}

// viewInput renders the input view
func (m LuckyModel) viewInput() string {
	var content strings.Builder

	t := theme.Current()
	titleStyle := lipgloss.NewStyle().
		Foreground(t.HighlightColor()).
		Bold(true).
		PaddingLeft(t.Padding.ListItemLeft)

	// Title
	content.WriteString(titleStyle.Render("🎲 I Feel Lucky"))
	content.WriteString("\n\n")

	// "Choose an option:" with number buffer
	content.WriteString(subtitleStyle().Render("Choose an option:"))
	if m.numberBuffer != "" {
		content.WriteString(" ")
		content.WriteString(highlightStyle().Render(m.numberBuffer + "_"))
	}
	content.WriteString("\n\n")

	// Instructions
	content.WriteString("Type a genre of music: rock, classical, jazz, pop, country, hip, heavy, blues, soul.\n")
	content.WriteString("Or type a keyword like: meditation, relax, mozart, Beatles, etc.\n\n")
	content.WriteString(infoStyle().Render("Use only one word."))
	content.WriteString("\n\n")

	// Input field
	content.WriteString("Genre/keyword: ")
	content.WriteString(m.textInput.View())
	content.WriteString("\n\n")

	// Shuffle mode toggle
	if m.shuffleEnabled {
		content.WriteString("Shuffle mode: ")
		content.WriteString(highlightStyle().Render("[✓] On"))
		content.WriteString("  (press 't' to disable)")
		if m.shuffleConfig.AutoAdvance {
			fmt.Fprintf(&content, "\n              Auto-advance in %d min • History: %d stations",
				m.shuffleConfig.IntervalMinutes, m.shuffleConfig.MaxHistory)
		}
	} else {
		content.WriteString("Shuffle mode: ")
		content.WriteString(infoStyle().Render("[ ] Off"))
		content.WriteString(" (press 't' to enable)")
	}

	// Show history menu if available
	if m.searchHistory != nil && len(m.searchHistory.LuckyQueries) > 0 {
		content.WriteString("\n\n")
		content.WriteString(m.menuList.View())
	}

	// Error message if any
	if m.err != nil {
		content.WriteString("\n")
		content.WriteString(errorStyle().Render(m.err.Error()))
	}

	// Save message if any
	if m.saveMessage != "" {
		content.WriteString("\n")
		if strings.Contains(m.saveMessage, "✓") || strings.Contains(m.saveMessage, "🚫") {
			content.WriteString(successStyle().Render(m.saveMessage))
		} else if strings.Contains(m.saveMessage, "✗") {
			content.WriteString(errorStyle().Render(m.saveMessage))
		} else {
			content.WriteString(infoStyle().Render(m.saveMessage))
		}
	}

	helpText := "↑↓/jk: Navigate • Enter: Search • t: Toggle shuffle"
	if m.searchHistory != nil && len(m.searchHistory.LuckyQueries) > 0 {
		maxItems := len(m.searchHistory.LuckyQueries)
		if maxItems > m.searchHistory.MaxSize {
			maxItems = m.searchHistory.MaxSize
		}
		helpText += fmt.Sprintf(" • 1-%d: Quick search", maxItems)
	}
	helpText += " • Esc: Back"

	return RenderPageWithBottomHelp(PageLayout{
		Content: content.String(),
		Help:    helpText,
	}, m.height)
}

// viewSearching renders the searching view
func (m LuckyModel) viewSearching() string {
	var content strings.Builder

	content.WriteString(infoStyle().Render("🔍 Searching for stations..."))
	content.WriteString("\n\n")
	content.WriteString("Finding a random station for you...")

	// Save message if any
	if m.saveMessage != "" {
		content.WriteString("\n\n")
		if strings.Contains(m.saveMessage, "✓") || strings.Contains(m.saveMessage, "🚫") {
			content.WriteString(successStyle().Render(m.saveMessage))
		} else if strings.Contains(m.saveMessage, "✗") {
			content.WriteString(errorStyle().Render(m.saveMessage))
		} else {
			content.WriteString(infoStyle().Render(m.saveMessage))
		}
	}

	return RenderPage(PageLayout{
		Title:   "I Feel Lucky",
		Content: content.String(),
		Help:    "Please wait...",
	})
}

// viewPlaying renders the playback view
func (m LuckyModel) viewPlaying() string {
	if m.selectedStation == nil {
		return "No station selected"
	}

	var content strings.Builder

	// Station info with rating
	// Get rating for display
	var rating int
	if m.ratingsManager != nil {
		if r := m.ratingsManager.GetRating(m.selectedStation.StationUUID); r != nil {
			rating = r.Rating
		}
	}
	// Get metadata for display
	var metadata *storage.StationMetadata
	if m.metadataManager != nil {
		metadata = m.metadataManager.GetMetadata(m.selectedStation.StationUUID)
	}
	content.WriteString(RenderStationDetailsWithRating(*m.selectedStation, false, metadata, rating, m.starRenderer))

	// Playback status with proper spacing
	content.WriteString("\n")
	if m.player.IsPlaying() {
		// Show current track if available
		if track, err := m.player.GetCurrentTrack(); err == nil && IsValidTrackMetadata(track, m.selectedStation.Name) {
			content.WriteString(successStyle().Render("▶ Now Playing:"))
			content.WriteString(" ")
			content.WriteString(infoStyle().Render(track))
		} else {
			content.WriteString(successStyle().Render("▶ Playing..."))
		}
	} else {
		content.WriteString(infoStyle().Render("⏸ Stopped"))
	}

	// Save message (if any)
	if m.saveMessage != "" {
		content.WriteString("\n\n")
		var msgStyle lipgloss.Style
		if strings.Contains(m.saveMessage, "✓") || strings.HasPrefix(m.saveMessage, "Volume:") {
			if strings.Contains(m.saveMessage, "Muted") {
				msgStyle = infoStyle()
			} else {
				msgStyle = successStyle()
			}
		} else if strings.Contains(m.saveMessage, "Already") ||
			strings.Contains(m.saveMessage, "Paused") ||
			strings.Contains(m.saveMessage, "Resumed") {
			msgStyle = infoStyle()
		} else {
			msgStyle = errorStyle()
		}
		content.WriteString(msgStyle.Render(m.saveMessage))
	}

	return RenderPageWithBottomHelp(PageLayout{
		Title:   "🎵 Now Playing",
		Content: content.String(),
		Help:    "b: Block • u: Undo • Space: Pause/Play • r: Rate • f: Save to Favorites • s: Save to list • v: Vote • ?: Help",
	}, m.height)
}

// viewSavePrompt renders the save prompt after playback
func (m LuckyModel) viewSavePrompt() string {
	if m.selectedStation == nil {
		return "No station selected"
	}

	var content strings.Builder

	// Message
	content.WriteString("Did you enjoy this station?\n\n")

	// Station name
	content.WriteString(stationNameStyle().Render(m.selectedStation.TrimName()))
	content.WriteString("\n\n")

	// Options
	content.WriteString("1) ⭐ Add to Quick Favorites\n")
	content.WriteString("2) Return to Main Menu")

	// Show save message if any (from pressing 'f' during playback)
	if m.saveMessage != "" {
		content.WriteString("\n\n")
		var msgStyle lipgloss.Style
		if strings.Contains(m.saveMessage, "✓") {
			msgStyle = successStyle()
		} else if strings.Contains(m.saveMessage, "Already") {
			msgStyle = infoStyle()
		} else {
			msgStyle = errorStyle()
		}
		content.WriteString(msgStyle.Render(m.saveMessage))
	}

	return RenderPage(PageLayout{
		Title:   "💾 Save Station?",
		Content: content.String(),
		Help:    "y/1: Yes • n/2/Esc: No",
	})
}

// viewSelectList renders the list selection view
func (m LuckyModel) viewSelectList() string {
	if m.selectedStation == nil {
		return "No station selected"
	}

	var content strings.Builder

	// Station name
	content.WriteString("Station: ")
	content.WriteString(stationNameStyle().Render(m.selectedStation.TrimName()))
	content.WriteString("\n\n")

	// Instruction
	content.WriteString("Select a list to save to:\n\n")

	// List selection
	if len(m.availableLists) == 0 {
		content.WriteString(infoStyle().Render("No existing lists."))
		content.WriteString("\n")
	} else if m.listModel.Items() != nil {
		content.WriteString(m.listModel.View())
	} else {
		content.WriteString("Loading lists...")
	}

	return RenderPage(PageLayout{
		Title:   "💾 Save to List",
		Content: content.String(),
		Help:    "↑↓/jk: Navigate • Enter: Select • n: New list • Esc: Cancel",
	})
}

// viewNewListInput renders the new list name input view
func (m LuckyModel) viewNewListInput() string {
	if m.selectedStation == nil {
		return "No station selected"
	}

	var content strings.Builder

	// Station name
	content.WriteString("Station: ")
	content.WriteString(stationNameStyle().Render(m.selectedStation.TrimName()))
	content.WriteString("\n\n")

	// Instruction
	content.WriteString("Enter new list name:\n\n")

	// Text input
	content.WriteString(m.newListInput.View())

	return RenderPage(PageLayout{
		Title:   "💾 Create New List",
		Content: content.String(),
		Help:    "Enter: Save • Esc: Cancel",
	})
}

// rebuildMenuWithHistory rebuilds the menu list with history items
func (m *LuckyModel) rebuildMenuWithHistory() {
	if m.searchHistory == nil || len(m.searchHistory.LuckyQueries) == 0 {
		// Create empty menu
		m.menuList = components.CreateMenu([]components.MenuItem{}, "", 50, 5)
		return
	}

	// Build menu items from history
	menuItems := []components.MenuItem{}

	// Add separator
	menuItems = append(menuItems, components.NewMenuItem("─── Recent Searches ───", "", ""))

	// Add history items with numbers 1, 2, 3, etc.
	for i, query := range m.searchHistory.LuckyQueries {
		if i >= m.searchHistory.MaxSize {
			break
		}
		shortcut := fmt.Sprintf("%d", i+1)
		menuItems = append(menuItems, components.NewMenuItem(query, "", shortcut))
	}

	// Calculate appropriate height
	height := len(menuItems) + 2
	if height > 15 {
		height = 15
	}

	// Use empty title - we render the title manually
	m.menuList = components.CreateMenu(menuItems, "", 50, height)
}

// updateShufflePlaying handles input during shuffle playback
func (m LuckyModel) updateShufflePlaying(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		// Stop shuffle and playback, return to I Feel Lucky input
		if m.shuffleManager != nil {
			m.shuffleManager.Stop()
		}
		if err := m.player.Stop(); err != nil {
			m.saveMessage = fmt.Sprintf("✗ Failed to stop playback: %v", err)
			m.saveMessageTime = messageDisplayLong
			return m, nil
		}
		m.state = luckyStateInput
		m.selectedStation = nil
		m.shuffleEnabled = false
		m.shuffleManager = nil
		// Reload history from disk
		m.reloadSearchHistory()
		m.rebuildMenuWithHistory()
		return m, nil
	case "0":
		// Return to main menu
		if m.shuffleManager != nil {
			m.shuffleManager.Stop()
		}
		if err := m.player.Stop(); err != nil {
			m.saveMessage = fmt.Sprintf("✗ Failed to stop playback: %v", err)
			m.saveMessageTime = messageDisplayLong
			return m, nil
		}
		m.selectedStation = nil
		m.state = luckyStateInput
		m.shuffleEnabled = false
		m.shuffleManager = nil
		return m, func() tea.Msg {
			return navigateMsg{screen: screenMainMenu}
		}
	case "h":
		// Stop shuffle but keep playing current station
		if m.shuffleManager != nil {
			m.shuffleManager.Stop()
		}
		m.shuffleEnabled = false
		m.shuffleManager = nil
		m.state = luckyStatePlaying
		m.saveMessage = "Shuffle stopped - continuing with current station"
		m.saveMessageTime = messageDisplayMedium
		return m, nil
	case "n":
		// Next shuffle station (manual advance)
		if m.shuffleManager == nil {
			return m, nil
		}
		nextStation, err := m.shuffleManager.Next(func(s api.Station) bool {
			if m.blocklistManager == nil {
				return true
			}
			return !m.blocklistManager.IsBlockedByAny(&s)
		})
		if err != nil {
			m.saveMessage = fmt.Sprintf("✗ Shuffle error: %v", err)
			m.saveMessageTime = messageDisplayShort
			return m, nil
		}
		m.selectedStation = nextStation
		// Stop current playback and start new station
		_ = m.player.Stop() // Ignore error, we're starting new playback anyway
		return m, tea.Batch(
			m.startPlayback(),
			m.shuffleTimerTick(),
		)
	case "b":
		// Block current station
		if m.selectedStation != nil {
			return m, m.blockStation()
		}
		return m, nil
	case "u":
		// Undo last block (within 5 seconds)
		if time.Since(m.lastBlockTime) < 5*time.Second {
			return m, m.undoLastBlock()
		}
		return m, nil
	case "[":
		// Previous shuffle station (from history)
		if m.shuffleManager == nil {
			return m, nil
		}
		prevStation, err := m.shuffleManager.Previous()
		if err != nil {
			m.saveMessage = fmt.Sprintf("✗ %v", err)
			m.saveMessageTime = messageDisplayLong
			return m, nil
		}
		m.selectedStation = prevStation
		// Stop current playback and start new station
		_ = m.player.Stop() // Ignore error, we're starting new playback anyway
		return m, tea.Batch(
			m.startPlayback(),
			m.shuffleTimerTick(),
		)
	case "p":
		// Pause/resume auto-advance timer
		if m.shuffleManager == nil {
			return m, nil
		}
		paused := m.shuffleManager.TogglePause()
		if paused {
			m.saveMessage = "⏸ Auto-advance paused"
		} else {
			m.saveMessage = "▶ Auto-advance resumed"
		}
		m.saveMessageTime = messageDisplayMedium
		return m, nil
	case "f":
		// Save to Quick Favorites during shuffle
		return m, m.saveToQuickFavorites()
	case "s":
		// Save to a list - show list selection
		m.state = luckyStateSelectList
		return m, m.loadAvailableLists()
	case "v":
		// Vote for this station
		return m, m.voteForStation()
	case " ":
		// Toggle pause/play with space bar
		if err := m.player.TogglePause(); err != nil {
			m.saveMessage = fmt.Sprintf("✗ Pause failed: %v", err)
			m.saveMessageTime = messageDisplayLong
		} else {
			if m.player.IsPaused() {
				m.saveMessage = "⏸ Paused"
			} else {
				m.saveMessage = "▶ Resumed"
			}
			m.saveMessageTime = messageDisplayShort
		}
		return m, nil
	case "/", "*", "m":
		if handled, msg := m.handleVolumeControl(msg.String()); handled {
			m.saveMessage = msg
			m.saveMessageTime = messageDisplayShort
			return m, nil
		}
	case "?":
		m.helpModel.SetSize(m.width, m.height)
		m.helpModel.Toggle()
		return m, nil
	}
	return m, nil
}

// searchForShuffle searches and returns all stations for shuffle mode
func (m LuckyModel) searchForShuffle(keyword string) tea.Cmd {
	return func() tea.Msg {
		// Save to history in background
		go func() {
			store := storage.NewStorage(m.favoritePath)
			_ = store.AddLuckyQuery(context.Background(), keyword)
		}()

		// Search by tag (genre/keyword)
		stations, err := m.apiClient.SearchByTag(context.Background(), keyword)
		if err != nil {
			return luckySearchErrorMsg{err: fmt.Errorf("search failed: %w", err)}
		}

		if len(stations) == 0 {
			return luckySearchErrorMsg{err: fmt.Errorf("no stations found for '%s'", keyword)}
		}

		// Filter out blocked stations
		if m.blocklistManager != nil {
			filtered := make([]api.Station, 0, len(stations))
			for _, s := range stations {
				if !m.blocklistManager.IsBlockedByAny(&s) {
					filtered = append(filtered, s)
				}
			}
			stations = filtered
		}

		if len(stations) == 0 {
			return luckySearchErrorMsg{err: fmt.Errorf("all stations found for '%s' are blocked", keyword)}
		}

		return luckyShuffleSearchResultsMsg{
			stations: stations,
			keyword:  keyword,
		}
	}
}

// shuffleTimerTick sends a timer tick message for shuffle mode
func (m LuckyModel) shuffleTimerTick() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return shuffleTimerTickMsg{}
	})
}

// viewShufflePlaying renders the shuffle playback view
func (m LuckyModel) viewShufflePlaying() string {
	if m.selectedStation == nil || m.shuffleManager == nil {
		return "No shuffle session active"
	}

	var content strings.Builder

	// Station info with rating
	var rating int
	if m.ratingsManager != nil {
		if r := m.ratingsManager.GetRating(m.selectedStation.StationUUID); r != nil {
			rating = r.Rating
		}
	}
	var metadata *storage.StationMetadata
	if m.metadataManager != nil {
		metadata = m.metadataManager.GetMetadata(m.selectedStation.StationUUID)
	}
	content.WriteString(RenderStationDetailsWithRating(*m.selectedStation, false, metadata, rating, m.starRenderer))

	// Playback status
	content.WriteString("\n")
	if m.player.IsPlaying() {
		// Show current track if available
		if track, err := m.player.GetCurrentTrack(); err == nil && IsValidTrackMetadata(track, m.selectedStation.Name) {
			content.WriteString(successStyle().Render("▶ Now Playing:"))
			content.WriteString(" ")
			content.WriteString(infoStyle().Render(track))
		} else {
			content.WriteString(successStyle().Render("▶ Playing..."))
		}
	} else {
		content.WriteString(infoStyle().Render("⏸ Stopped"))
	}

	// Shuffle info
	content.WriteString("\n\n")
	shuffleInfo := m.shuffleManager.GetStatus()
	content.WriteString(highlightStyle().Render("🔀 Shuffle Active"))

	// Auto-advance timer (if enabled)
	if m.shuffleConfig.AutoAdvance {
		if shuffleInfo.TimerPaused {
			content.WriteString(" • ")
			content.WriteString(infoStyle().Render("⏸ Timer paused"))
		} else if shuffleInfo.TimeRemaining > 0 {
			minutes := int(shuffleInfo.TimeRemaining.Minutes())
			seconds := int(shuffleInfo.TimeRemaining.Seconds()) % 60
			fmt.Fprintf(&content, " • Next in: %d:%02d", minutes, seconds)
		}
	} else {
		content.WriteString(" • ")
		content.WriteString(infoStyle().Render("Manual mode"))
	}

	// Station counter
	fmt.Fprintf(&content, "\n   Station %d of session", shuffleInfo.SessionCount+1)

	// Shuffle history (append current station for display)
	history := shuffleInfo.History
	if m.selectedStation != nil {
		history = append(history, *m.selectedStation)
	}
	if len(history) > 0 {
		content.WriteString("\n\n")
		content.WriteString(subtitleStyle().Render("─── Shuffle History ───"))
		content.WriteString("\n")

		// Show up to 3 most recent stations
		startIdx := 0
		if len(history) > 3 {
			startIdx = len(history) - 3
		}

		for i := startIdx; i < len(history); i++ {
			station := history[i]
			isBlocked := false
			if m.blocklistManager != nil {
				isBlocked = m.blocklistManager.IsBlockedByAny(&station)
			}
			name := station.TrimName()
			if isBlocked {
				name = "🚫 " + name
			}

			if i == len(history)-1 {
				// Current station
				content.WriteString("  → ")
				content.WriteString(highlightStyle().Render(name))
				content.WriteString(" ← Current")
			} else {
				// Previous station
				content.WriteString("  ← ")
				content.WriteString(name)
			}
			content.WriteString("\n")
		}
	}

	// Save message (if any)
	if m.saveMessage != "" {
		content.WriteString("\n")
		var msgStyle lipgloss.Style
		if strings.Contains(m.saveMessage, "✓") || strings.HasPrefix(m.saveMessage, "Volume:") {
			if strings.Contains(m.saveMessage, "Muted") {
				msgStyle = infoStyle()
			} else {
				msgStyle = successStyle()
			}
		} else if strings.Contains(m.saveMessage, "Already") ||
			strings.Contains(m.saveMessage, "Paused") ||
			strings.Contains(m.saveMessage, "Resumed") ||
			strings.Contains(m.saveMessage, "paused") ||
			strings.Contains(m.saveMessage, "resumed") ||
			strings.Contains(m.saveMessage, "stopped") {
			msgStyle = infoStyle()
		} else {
			msgStyle = errorStyle()
		}
		content.WriteString(msgStyle.Render(m.saveMessage))
	}

	title := fmt.Sprintf("🎵 Now Playing (🔀 Shuffle: %s)", m.lastSearchKeyword)
	help := "Space: Pause/Play • b: Block • u: Undo • f: Fav • s: List • v: Vote • n: Next • [: Prev • p: Pause timer • h: Stop shuffle • ?: Help"

	return RenderPageWithBottomHelp(PageLayout{
		Title:   title,
		Content: content.String(),
		Help:    help,
	}, m.height)
}

// blockStation blocks the currently playing station
func (m LuckyModel) blockStation() tea.Cmd {
	return func() tea.Msg {
		if m.selectedStation == nil {
			return luckySearchErrorMsg{fmt.Errorf("no station selected")}
		}
		if m.blocklistManager == nil {
			return luckySearchErrorMsg{fmt.Errorf("blocklist not available")}
		}

		ctx := context.Background()
		msg, err := m.blocklistManager.Block(ctx, m.selectedStation)
		if err != nil {
			// Check if already blocked
			if errors.Is(err, blocklist.ErrStationAlreadyBlocked) {
				return stationBlockedMsg{
					message:     "Station is already blocked",
					stationUUID: m.selectedStation.StationUUID,
					success:     false,
				}
			}
			return luckySearchErrorMsg{err}
		}

		return stationBlockedMsg{
			message:     msg,
			stationUUID: m.selectedStation.StationUUID,
			success:     true,
		}
	}
}

// undoLastBlock undoes the last block operation
func (m LuckyModel) undoLastBlock() tea.Cmd {
	return func() tea.Msg {
		if m.blocklistManager == nil {
			return undoBlockFailedMsg{}
		}
		ctx := context.Background()
		undone, err := m.blocklistManager.UndoLastBlock(ctx)
		if err != nil {
			return luckySearchErrorMsg{err}
		}

		if undone {
			return undoBlockSuccessMsg{}
		}
		return undoBlockFailedMsg{}
	}
}

// reloadSearchHistory reloads history from disk
func (m *LuckyModel) reloadSearchHistory() {
	store := storage.NewStorage(m.favoritePath)
	history, err := store.LoadSearchHistory(context.Background())
	if err != nil || history == nil {
		history = storage.NewSearchHistoryStore()
	}
	m.searchHistory = history
}
