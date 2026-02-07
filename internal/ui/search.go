package ui

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/shinokada/tera/internal/api"
	"github.com/shinokada/tera/internal/blocklist"
	"github.com/shinokada/tera/internal/player"
	"github.com/shinokada/tera/internal/storage"
	"github.com/shinokada/tera/internal/theme"
	"github.com/shinokada/tera/internal/ui/components"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// searchState represents the current state in the search screen
type searchState int

const (
	searchStateMenu searchState = iota
	searchStateInput
	searchStateLoading
	searchStateResults
	searchStateStationInfo
	searchStatePlaying
	searchStateSavePrompt
	searchStateSelectList
	searchStateNewListInput
	searchStateAdvancedForm
)

// SearchModel represents the search screen
type SearchModel struct {
	state            searchState
	searchType       api.SearchType
	menuList         list.Model // List-based menu navigation
	stationInfoMenu  list.Model // Station info submenu navigation
	apiClient        *api.Client
	textInput        textinput.Model
	newListInput     textinput.Model
	spinner          spinner.Model
	results          []api.Station
	resultsItems     []list.Item
	resultsList      list.Model
	selectedStation  *api.Station
	player           *player.MPVPlayer
	favoritePath     string
	quickFavorites   []api.Station               // My-favorites.json for duplicate checking
	searchHistory    *storage.SearchHistoryStore // Search history
	numberBuffer     string                      // Buffer for number input display
	saveMessage      string
	saveMessageTime  int
	width            int
	height           int
	err              error
	availableLists   []string
	listItems        []list.Item
	listModel        list.Model
	helpModel        components.HelpModel
	votedStations    *storage.VotedStations // Track voted stations
	blocklistManager *blocklist.Manager
	lastBlockTime    time.Time
	// Advanced search form fields
	advancedInputs      [5]textinput.Model // tag, language, country, state, name
	advancedFocusIdx    int                // 0-4: text fields, 5: sort, 6: bitrate
	advancedBitrate     string             // "1", "2", "3", or ""
	advancedSortByVotes bool               // true = votes, false = relevance
}

// Messages for search screen
type searchResultsMsg struct {
	results []api.Station
}

type searchErrorMsg struct {
	err error
}

type quickFavoritesLoadedMsg struct {
	stations []api.Station
}

type playerErrorMsg struct {
	err error
}

type playbackStalledMsg struct {
	station api.Station
}

type checkSignalMsg struct {
	station api.Station
	attempt int
}

// NewSearchModel creates a new search screen model
func NewSearchModel(apiClient *api.Client, favoritePath string, blocklistManager *blocklist.Manager) SearchModel {
	ti := textinput.New()
	ti.Placeholder = "Enter search query..."
	ti.CharLimit = 100
	ti.Width = 50

	// New list input
	nli := textinput.New()
	nli.Placeholder = "Enter new list name..."
	nli.CharLimit = 50
	nli.Width = 50

	// Advanced search inputs
	var advInputs [5]textinput.Model
	placeholders := []string{
		"classical",
		"italian",
		"US",
		"California",
		"BBC Radio",
	}
	for i := 0; i < 5; i++ {
		advInputs[i] = textinput.New()
		advInputs[i].Placeholder = placeholders[i]
		advInputs[i].CharLimit = 100
		advInputs[i].Width = 40
	}

	sp := spinner.New()
	sp.Spinner = spinner.Dot

	// Create search menu items
	menuItems := []components.MenuItem{
		components.NewMenuItem("Search by Tag", "(genre, style, etc.)", "1"),
		components.NewMenuItem("Search by Name", "", "2"),
		components.NewMenuItem("Search by Language", "", "3"),
		components.NewMenuItem("Search by Country Code", "", "4"),
		components.NewMenuItem("Search by State", "", "5"),
		components.NewMenuItem("Advanced Search", "(multiple criteria)", "6"),
	}

	// Set enough height for all 6 menu items + title
	menuList := components.CreateMenu(menuItems, "🔍 Search Radio Stations", 50, 15)

	// Create station info submenu items
	infoMenuItems := []components.MenuItem{
		components.NewMenuItem("Play this station", "", "1"),
		components.NewMenuItem("Save to Quick Favorites", "", "2"),
		components.NewMenuItem("Back to search results", "", "3"),
	}

	// Initial height will be updated on first WindowSizeMsg
	stationInfoMenu := components.CreateMenu(infoMenuItems, "What would you like to do?", 50, 10)

	// Load search history (if it fails, just use empty history)
	store := storage.NewStorage(favoritePath)
	history, err := store.LoadSearchHistory(context.Background())
	if err != nil || history == nil {
		history = storage.NewSearchHistoryStore()
	}

	// Load voted stations
	votedStations, err := storage.LoadVotedStations()
	if err != nil {
		// If we can't load, just create empty list
		votedStations = &storage.VotedStations{Stations: []storage.VotedStation{}}
	}

	model := SearchModel{
		state:               searchStateMenu,
		apiClient:           apiClient,
		menuList:            menuList,
		stationInfoMenu:     stationInfoMenu,
		textInput:           ti,
		newListInput:        nli,
		spinner:             sp,
		favoritePath:        favoritePath,
		player:              player.NewMPVPlayer(),
		quickFavorites:      []api.Station{},
		searchHistory:       history,
		width:               80, // Default width
		height:              24, // Default height
		helpModel:           components.NewHelpModel(components.CreatePlayingHelp()),
		votedStations:       votedStations,
		advancedInputs:      advInputs,
		advancedFocusIdx:    0,
		advancedBitrate:     "",
		advancedSortByVotes: true, // Default to sorting by votes
		blocklistManager:    blocklistManager,
	}

	// Build menu with history items included
	model.rebuildMenuWithHistory()

	return model
}

// Init initializes the search screen
func (m SearchModel) Init() tea.Cmd {
	return tea.Batch(
		m.loadQuickFavorites(),
		m.spinner.Tick,
		tickEverySecond(), // For save message countdown
	)
}

// Note: tickEverySecond() is defined in messages.go and ticks once per second

// loadQuickFavorites loads My-favorites.json for duplicate checking
func (m SearchModel) loadQuickFavorites() tea.Cmd {
	return func() tea.Msg {
		store := storage.NewStorage(m.favoritePath)
		list, err := store.LoadList(context.Background(), "My-favorites")
		if err != nil {
			// It's OK if My-favorites doesn't exist yet
			return quickFavoritesLoadedMsg{stations: []api.Station{}}
		}
		return quickFavoritesLoadedMsg{stations: list.Stations}
	}
}

// Update handles messages for the search screen
func (m SearchModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		// Calculate usable height
		listHeight := msg.Height - 14
		if listHeight < 5 {
			listHeight = 5 // Minimum height
		}

		// Update list sizes based on current state
		switch m.state {
		case searchStateMenu:
			m.menuList.SetSize(msg.Width-4, listHeight)
		case searchStateResults:
			if m.resultsList.Items() != nil && len(m.resultsList.Items()) > 0 {
				m.resultsList.SetSize(msg.Width-4, listHeight)
			}
		case searchStateStationInfo:
			// Station info menu is smaller
			infoHeight := 10
			if infoHeight > listHeight {
				infoHeight = listHeight
			}
			m.stationInfoMenu.SetSize(msg.Width-4, infoHeight)
		}

		m.helpModel.SetSize(msg.Width, msg.Height)

	case tea.KeyMsg:
		if m.helpModel.IsVisible() {
			var cmd tea.Cmd
			m.helpModel, cmd = m.helpModel.Update(msg)
			return m, cmd
		}

		switch m.state {
		case searchStateMenu:
			return m.handleMenuInput(msg)
		case searchStateInput:
			return m.handleTextInput(msg)
		case searchStateResults:
			return m.handleResultsInput(msg)
		case searchStateStationInfo:
			return m.handleStationInfoInput(msg)
		case searchStatePlaying:
			return m.handlePlayerUpdate(msg)
		case searchStateSavePrompt:
			return m.handleSavePrompt(msg)
		case searchStateSelectList:
			return m.handleSelectList(msg)
		case searchStateNewListInput:
			return m.handleNewListInput(msg)
		case searchStateAdvancedForm:
			return m.handleAdvancedForm(msg)
		}

	case quickFavoritesLoadedMsg:
		m.quickFavorites = msg.stations

	case searchResultsMsg:
		m.results = msg.results
		m.state = searchStateResults
		m.resultsItems = make([]list.Item, 0, len(m.results))
		for _, station := range m.results {
			isBlocked := false
			if m.blocklistManager != nil {
				isBlocked = m.blocklistManager.IsBlockedByAny(&station)
			}
			m.resultsItems = append(m.resultsItems, stationListItem{station: station, isBlocked: isBlocked})
		}

		// Calculate proper list height
		// Header needs enough space, so use same buffer as search menu
		listHeight := m.height - 14
		if listHeight < 5 {
			listHeight = 5
		}

		// Create results list
		delegate := createStyledDelegate()

		m.resultsList = list.New(m.resultsItems, delegate, m.width, listHeight)
		m.resultsList.Title = fmt.Sprintf("Search Results (%d stations)", len(m.results))
		m.resultsList.SetShowHelp(false)     // We use custom footer instead
		m.resultsList.SetShowStatusBar(true) // Show status bar for filter count
		m.resultsList.SetFilteringEnabled(true)
		// Disable 'q' quit keybinding in the list
		m.resultsList.KeyMap.Quit = key.NewBinding(key.WithDisabled())
		return m, nil

	case searchErrorMsg:
		m.err = msg.err
		m.state = searchStateMenu
		return m, nil

	case spinner.TickMsg:
		if m.state == searchStateLoading {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}

	case playbackStartedMsg:
		// Playback started successfully, stay in playing state
		return m, nil

	case playerErrorMsg:
		m.err = msg.err
		m.state = searchStateResults
		return m, nil

	case playbackStalledMsg:
		// Stop player if it's still "playing" (but silent)
		if m.player != nil {
			_ = m.player.Stop()
		}
		m.saveMessage = "✗ No signal detected"
		m.saveMessageTime = messageDisplayShort
		m.state = searchStateResults
		return m, nil

	case checkSignalMsg:
		if m.state == searchStatePlaying && m.selectedStation != nil && m.selectedStation.StationUUID == msg.station.StationUUID {
			return m, m.checkPlaybackSignal(msg.station, msg.attempt)
		}
		return m, nil

	case playbackStoppedMsg:
		// Handle save prompt after playback
		return m.handlePlaybackStopped()

	case saveSuccessMsg:
		// Update local cache
		m.quickFavorites = append(m.quickFavorites, *msg.station)
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

	case tickMsg:
		// Handle save message countdown
		if m.saveMessageTime > 0 {
			m.saveMessageTime--
			if m.saveMessageTime <= 0 {
				m.saveMessage = ""
			}
		}
		// Continue ticking
		return m, tickEverySecond()

	case listsLoadedMsg:
		m.availableLists = msg.lists
		m.listItems = make([]list.Item, len(msg.lists))
		for i, name := range msg.lists {
			m.listItems[i] = playListItem{name: name}
		}
		if m.width > 0 && m.height > 0 {
			m.initializeListModel()
		}
		return m, nil

	case saveToListFailedMsg:
		if msg.isDuplicate {
			m.saveMessage = "Already in this list"
		} else {
			m.saveMessage = fmt.Sprintf("✗ Failed to save: %v", msg.err)
		}
		m.saveMessageTime = messageDisplayShort
		m.state = searchStatePlaying
		return m, nil

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

			// Return to results
			m.state = searchStateResults

			// Update blocked status in the list items
			if m.resultsList.Items() != nil {
				items := m.resultsList.Items()
				for i, item := range items {
					if si, ok := item.(stationListItem); ok && si.station.StationUUID == msg.stationUUID {
						si.isBlocked = true
						items[i] = si
						break
					}
				}
				m.resultsList.SetItems(items)
			}
			m.selectedStation = nil
		} else {
			// Already blocked
			m.saveMessage = msg.message
			m.saveMessageTime = messageDisplayShort
		}
		return m, nil

	case undoBlockSuccessMsg:
		m.saveMessage = "✓ Block undone"
		m.saveMessageTime = messageDisplayShort
		// Update blocked status in the list items
		if m.resultsList.Items() != nil {
			items := m.resultsList.Items()
			for i, item := range items {
				if si, ok := item.(stationListItem); ok {
					if m.blocklistManager != nil && !m.blocklistManager.IsBlockedByAny(&si.station) {
						if si.isBlocked {
							si.isBlocked = false
							items[i] = si
						}
					}
				}
			}
			m.resultsList.SetItems(items)
		}
		return m, nil

	case undoBlockFailedMsg:
		m.saveMessage = "No recent block to undo"
		m.saveMessageTime = messageDisplayShort
		return m, nil
	}

	return m, tea.Batch(cmds...)
}

// handleMenuInput handles input in the search menu state
func (m SearchModel) handleMenuInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Handle back to main menu
	if msg.String() == "esc" || msg.String() == "m" {
		// Stop any playing station when exiting
		if m.player != nil && m.player.IsPlaying() {
			_ = m.player.Stop()
		}
		m.selectedStation = nil
		m.numberBuffer = "" // Clear buffer on exit
		return m, func() tea.Msg { return backToMainMsg{} }
	}

	// Handle number input for multi-digit selection
	key := msg.String()
	if key >= "0" && key <= "9" {
		m.numberBuffer += key

		// Parse current buffer
		var num int
		_, _ = fmt.Sscanf(m.numberBuffer, "%d", &num)

		// Calculate max valid number (6 for search types + history count)
		maxNum := 6
		if m.searchHistory != nil {
			maxHistoryItems := len(m.searchHistory.SearchItems)
			if maxHistoryItems > m.searchHistory.MaxSize {
				maxHistoryItems = m.searchHistory.MaxSize
			}
			if maxHistoryItems > 0 {
				maxNum = 10 + maxHistoryItems - 1 // e.g., 10, 11, 12...
			}
		}

		// Two or more digits: check if valid and execute
		if len(m.numberBuffer) >= 2 {
			// If number is valid, select it immediately
			if num >= 1 && num <= maxNum {
				m.numberBuffer = "" // Clear buffer
				return m.selectByNumber(num)
			}

			// If number is too large or 3+ digits, clear buffer
			if num > maxNum || len(m.numberBuffer) >= 3 {
				m.numberBuffer = ""
			}
		}
		// Single digit: just buffer it, don't execute yet (wait for Enter or second digit)
		return m, nil
	}

	// Handle Enter to submit buffered number
	if key == "enter" && m.numberBuffer != "" {
		var num int
		_, _ = fmt.Sscanf(m.numberBuffer, "%d", &num)
		m.numberBuffer = ""
		return m.selectByNumber(num)
	}

	// Clear buffer on navigation keys
	if key == "up" || key == "down" || key == "j" || key == "k" {
		m.numberBuffer = ""
	}

	// Handle menu navigation and selection
	newList, selected := components.HandleMenuKey(msg, m.menuList)
	m.menuList = newList

	if selected >= 0 {
		// Menu layout: 0-5 = search types, 6 = empty line, 7 = separator, 8+ = history items
		if selected < 6 {
			return m.executeSearchType(selected)
		} else if selected == 6 || selected == 7 {
			// Empty line or separator selected - ignore
			return m, nil
		} else {
			// History item selected (index 8+ maps to history index 0+)
			historyIndex := selected - 8
			if m.searchHistory != nil && historyIndex >= 0 && historyIndex < len(m.searchHistory.SearchItems) {
				item := m.searchHistory.SearchItems[historyIndex]
				return m.executeHistorySearch(item.SearchType, item.Query)
			}
		}
	}

	return m, nil
}

// selectByNumber handles selection by number input (1-6 for search types, 10+ for history)
func (m SearchModel) selectByNumber(num int) (tea.Model, tea.Cmd) {
	// 1-6: Search types
	if num >= 1 && num <= 6 {
		m.menuList.Select(num - 1)
		return m.executeSearchType(num - 1)
	}

	// 10+: History items (10 -> history[0], 11 -> history[1], etc.)
	if num >= 10 && m.searchHistory != nil {
		historyIndex := num - 10
		if historyIndex >= 0 && historyIndex < len(m.searchHistory.SearchItems) {
			// Select the menu item (8 + historyIndex because: 0-5=search types, 6=empty, 7=separator)
			m.menuList.Select(8 + historyIndex)
			item := m.searchHistory.SearchItems[historyIndex]
			return m.executeHistorySearch(item.SearchType, item.Query)
		}
	}

	return m, nil
}

// handleSelectList handles input during list selection
func (m SearchModel) handleSelectList(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		// Cancel and go back to playing
		m.state = searchStatePlaying
		return m, nil
	case "n":
		// Create new list
		m.state = searchStateNewListInput
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

// handleNewListInput handles input for new list name entry
func (m SearchModel) handleNewListInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		// Cancel and go back to list selection
		m.state = searchStateSelectList
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
func (m SearchModel) loadAvailableLists() tea.Cmd {
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
func (m *SearchModel) initializeListModel() {
	listHeight := m.height - 14
	if listHeight < 5 {
		listHeight = 5
	}

	delegate := createStyledDelegate()

	m.listModel = list.New(m.listItems, delegate, m.width, listHeight)
	m.listModel.Title = ""
	m.listModel.SetShowStatusBar(false)
	m.listModel.SetFilteringEnabled(false)
	m.listModel.SetShowHelp(false)
	m.listModel.Styles.Title = titleStyle()
	m.listModel.Styles.PaginationStyle = paginationStyle()
	m.listModel.Styles.HelpStyle = helpStyle()
}

// saveToList saves the current station to a specific list
func (m SearchModel) saveToList(listName string) tea.Cmd {
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

// executeSearchType sets up the search based on selected menu index
func (m SearchModel) executeSearchType(index int) (tea.Model, tea.Cmd) {
	switch index {
	case 0: // Search by Tag
		m.searchType = api.SearchByTag
		m.textInput.Placeholder = "Enter tag (e.g., jazz, rock, news)..."
	case 1: // Search by Name
		m.searchType = api.SearchByName
		m.textInput.Placeholder = "Enter station name..."
	case 2: // Search by Language
		m.searchType = api.SearchByLanguage
		m.textInput.Placeholder = "Enter language (e.g., english, spanish)..."
	case 3: // Search by Country
		m.searchType = api.SearchByCountry
		m.textInput.Placeholder = "Enter country code (e.g., US, UK, FR)..."
	case 4: // Search by State
		m.searchType = api.SearchByState
		m.textInput.Placeholder = "Enter state (e.g., California, Texas)..."
	case 5: // Advanced Search
		m.searchType = api.SearchAdvanced
		// Reset advanced form
		for i := range m.advancedInputs {
			m.advancedInputs[i].SetValue("")
		}
		m.advancedFocusIdx = 0
		m.advancedBitrate = ""
		m.advancedSortByVotes = true
		m.advancedInputs[0].Focus()
		m.state = searchStateAdvancedForm
		return m, textinput.Blink
	}

	m.state = searchStateInput
	m.textInput.Focus()
	return m, nil
}

// handleTextInput handles input in the text input state
func (m SearchModel) handleTextInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		query := strings.TrimSpace(m.textInput.Value())
		if query == "" {
			return m, nil
		}
		m.textInput.SetValue("")
		m.textInput.Blur()
		m.state = searchStateLoading
		return m, m.performSearch(query)
	case "esc":
		m.textInput.SetValue("")
		m.textInput.Blur()
		m.state = searchStateMenu
		// Reload history in case it was updated
		m.reloadSearchHistory()
		m.rebuildMenuWithHistory()
		return m, nil
	default:
		// Pass all other keys to text input for normal typing
		var cmd tea.Cmd
		m.textInput, cmd = m.textInput.Update(msg)
		return m, cmd
	}
}

// performSearch executes the search based on type
func (m SearchModel) performSearch(query string) tea.Cmd {
	// Save search to history in background
	go func() {
		store := storage.NewStorage(m.favoritePath)
		var searchTypeStr string
		switch m.searchType {
		case api.SearchByTag:
			searchTypeStr = "tag"
		case api.SearchByName:
			searchTypeStr = "name"
		case api.SearchByLanguage:
			searchTypeStr = "language"
		case api.SearchByCountry:
			searchTypeStr = "country"
		case api.SearchByState:
			searchTypeStr = "state"
		case api.SearchAdvanced:
			searchTypeStr = "advanced"
		}
		_ = store.AddSearchItem(context.Background(), searchTypeStr, query)
	}()

	return func() tea.Msg {
		var results []api.Station
		var err error
		ctx := context.Background()

		switch m.searchType {
		case api.SearchByTag:
			results, err = m.apiClient.SearchByTag(ctx, query)
		case api.SearchByName:
			results, err = m.apiClient.SearchByName(ctx, query)
		case api.SearchByLanguage:
			results, err = m.apiClient.SearchByLanguage(ctx, query)
		case api.SearchByCountry:
			results, err = m.apiClient.SearchByCountry(ctx, query)
		case api.SearchByState:
			results, err = m.apiClient.SearchByState(ctx, query)
		case api.SearchAdvanced:
			params := api.SearchParams{
				Name:       query,
				Tag:        query,
				Order:      "votes",
				Reverse:    true,
				Limit:      100,
				HideBroken: true,
			}
			results, err = m.apiClient.SearchAdvanced(ctx, params)
		}

		if err != nil {
			return searchErrorMsg{err: err}
		}

		// Sort results by votes (descending)
		sort.Slice(results, func(i, j int) bool {
			return results[i].Votes > results[j].Votes
		})

		return searchResultsMsg{results: results}
	}
}

// handleResultsInput handles input in the results list state
func (m SearchModel) handleResultsInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		// Stop any playing station when going back
		if m.player != nil && m.player.IsPlaying() {
			_ = m.player.Stop()
		}
		m.selectedStation = nil
		// If we came from advanced search, go back to the form
		if m.searchType == api.SearchAdvanced {
			m.state = searchStateAdvancedForm
			// Only focus text input if we're on a text field (0-4), not Sort (5) or Bitrate (6)
			if m.advancedFocusIdx < 5 {
				m.advancedInputs[m.advancedFocusIdx].Focus()
				return m, textinput.Blink
			}
			return m, nil
		}
		// Otherwise go back to search menu
		m.state = searchStateMenu
		// Reload history from disk so recent search appears
		m.reloadSearchHistory()
		m.rebuildMenuWithHistory()
		return m, nil
	case "0":
		// Return to main menu
		if m.player != nil && m.player.IsPlaying() {
			_ = m.player.Stop()
		}
		m.selectedStation = nil
		return m, func() tea.Msg { return backToMainMsg{} }
	case "enter":
		// Play station directly
		if item, ok := m.resultsList.SelectedItem().(stationListItem); ok {
			m.selectedStation = &item.station
			// Stop any currently playing station first
			if m.player != nil && m.player.IsPlaying() {
				_ = m.player.Stop()
			}
			m.state = searchStatePlaying
			return m, m.playStation(item.station)
		}
		return m, nil
	case "u":
		// Undo last block (within 5 seconds)
		if time.Since(m.lastBlockTime) < 5*time.Second {
			return m, m.undoLastBlock()
		}
		return m, nil
	default:
		// Pass all other keys (arrows, etc.) to the list for navigation
		var cmd tea.Cmd
		m.resultsList, cmd = m.resultsList.Update(msg)
		return m, cmd
	}
}

// handleStationInfoInput handles input in the station info state
func (m SearchModel) handleStationInfoInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {

	// Handle menu navigation and selection
	newList, selected := components.HandleMenuKey(msg, m.stationInfoMenu)
	m.stationInfoMenu = newList

	if selected >= 0 {
		return m.executeStationAction(selected)
	}

	// Handle Esc to go back
	if msg.String() == "esc" {
		// Stop player when going back
		if m.player != nil && m.player.IsPlaying() {
			_ = m.player.Stop()
		}
		m.selectedStation = nil
		m.state = searchStateResults
		return m, nil
	}

	return m, nil
}

// executeStationAction performs the selected action on the station
func (m SearchModel) executeStationAction(index int) (tea.Model, tea.Cmd) {
	switch index {
	case 0: // Play station
		// Stop any currently playing station first
		if m.player != nil && m.player.IsPlaying() {
			_ = m.player.Stop()
		}
		m.state = searchStatePlaying
		return m, m.playStation(*m.selectedStation)
	case 1: // Save to Quick Favorites
		return m, m.saveToQuickFavorites(*m.selectedStation)
	case 2: // Back to results
		// Stop player when going back
		if m.player != nil && m.player.IsPlaying() {
			_ = m.player.Stop()
		}
		m.selectedStation = nil
		m.state = searchStateResults
		return m, nil
	}
	return m, nil
}

// playStation starts playing a station
func (m SearchModel) playStation(station api.Station) tea.Cmd {
	return tea.Batch(
		func() tea.Msg {
			err := m.player.Play(&station)
			if err != nil {
				return playerErrorMsg{err: err}
			}
			return playbackStartedMsg{}
		},
		m.checkPlaybackSignal(station, 1),
	)
}

// checkPlaybackSignal checks for audio bitrate to ensure the station is actually playing
func (m SearchModel) checkPlaybackSignal(station api.Station, attempt int) tea.Cmd {
	return tea.Tick(2*time.Second, func(t time.Time) tea.Msg {
		if m.player == nil || !m.player.IsPlaying() {
			return nil
		}

		bitrate, err := m.player.GetAudioBitrate()
		if err == nil && bitrate > 0 {
			// Signal detected!
			return playbackStartedMsg{}
		}

		if attempt >= 4 { // 4 attempts * 2 seconds = 8 seconds
			return playbackStalledMsg{station: station}
		}

		return checkSignalMsg{station: station, attempt: attempt + 1}
	})
}

// handlePlaybackStopped handles return to results after playback
func (m SearchModel) handlePlaybackStopped() (tea.Model, tea.Cmd) {
	// Check if station is already in Quick Favorites
	if m.selectedStation != nil {
		isDuplicate := false
		for _, s := range m.quickFavorites {
			if s.StationUUID == m.selectedStation.StationUUID {
				isDuplicate = true
				break
			}
		}

		if isDuplicate {
			// Don't show save prompt if already saved
			m.saveMessage = "Already in Quick Favorites"
			m.saveMessageTime = messageDisplayShort
			m.state = searchStateResults
			m.selectedStation = nil
			return m, nil
		}

		// Show save prompt for new stations
		m.state = searchStateSavePrompt
		return m, nil
	}

	// No station selected, just go back
	m.state = searchStateResults
	return m, nil
}

// handleSavePrompt handles the save prompt after playback
func (m SearchModel) handleSavePrompt(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y", "1":
		// Save to Quick Favorites
		m.state = searchStateResults
		station := *m.selectedStation
		m.selectedStation = nil
		return m, m.saveToQuickFavorites(station)
	case "n", "2", "esc":
		// Don't save, go back to results
		m.state = searchStateResults
		m.selectedStation = nil
		return m, nil
	case "q":
		// Quit from save prompt
		return m, tea.Quit
	}
	return m, nil
}

// handlePlayerUpdate handles player-related updates during playback
func (m SearchModel) handlePlayerUpdate(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q":
		// Quit application
		if m.player != nil {
			_ = m.player.Stop()
		}
		return m, tea.Quit
	case "0":
		// Return to main menu (Level 3 shortcut)
		if m.player != nil {
			_ = m.player.Stop()
		}
		m.selectedStation = nil
		return m, func() tea.Msg { return backToMainMsg{} }
	case "esc":
		// Esc during playback goes back without save prompt
		if m.player != nil {
			_ = m.player.Stop()
		}
		m.selectedStation = nil
		m.state = searchStateResults
		return m, nil
	case "1":
		// Stop playback and trigger save prompt flow
		if m.player != nil {
			_ = m.player.Stop()
		}
		return m.handlePlaybackStopped()
	case "f":
		// Save to Quick Favorites during playback
		return m, m.saveToQuickFavorites(*m.selectedStation)
	case "s":
		// Save to a list - show list selection
		m.state = searchStateSelectList
		return m, m.loadAvailableLists()
	case "v":
		// Vote for this station
		return m, m.voteForStation()
	case "/":
		// Decrease volume
		newVol := m.player.DecreaseVolume(5)
		if m.selectedStation != nil && newVol >= 0 {
			m.selectedStation.SetVolume(newVol)
			m.saveStationVolume(m.selectedStation)
		}
		m.saveMessage = fmt.Sprintf("Volume: %d%%", newVol)
		startTick := m.saveMessageTime == 0
		m.saveMessageTime = 2 // Show for 2 seconds
		if startTick {
			return m, tickEverySecond()
		}
		return m, nil
	case "*":
		// Increase volume
		newVol := m.player.IncreaseVolume(5)
		if m.selectedStation != nil {
			m.selectedStation.SetVolume(newVol)
			m.saveStationVolume(m.selectedStation)
		}
		m.saveMessage = fmt.Sprintf("Volume: %d%%", newVol)
		startTick := m.saveMessageTime == 0
		m.saveMessageTime = 2 // Show for 2 seconds
		if startTick {
			return m, tickEverySecond()
		}
		return m, nil
	case "m":
		// Toggle mute
		muted, vol := m.player.ToggleMute()
		if muted {
			m.saveMessage = "Volume: Muted"
		} else {
			m.saveMessage = fmt.Sprintf("Volume: %d%%", vol)
		}
		if m.selectedStation != nil && !muted && vol >= 0 {
			m.selectedStation.SetVolume(vol)
			m.saveStationVolume(m.selectedStation)
		}
		startTick := m.saveMessageTime == 0
		m.saveMessageTime = 2 // Show for 2 seconds
		if startTick {
			return m, tickEverySecond()
		}
		return m, nil
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
	case "?":
		m.helpModel.SetSize(m.width, m.height)
		m.helpModel.Toggle()
		return m, nil
	case " ":
		// Toggle pause/resume
		if m.player != nil {
			if err := m.player.TogglePause(); err == nil {
				if m.player.IsPaused() {
					// Paused - show persistent message
					m.saveMessage = "⏸ Paused - Press Space to resume"
					m.saveMessageTime = -1 // Persistent
				} else {
					// Resumed - show temporary message
					m.saveMessage = "▶ Resumed"
					startTick := m.saveMessageTime <= 0
					m.saveMessageTime = 2
					if startTick {
						return m, tickEverySecond()
					}
				}
			}
		}
		return m, nil
	}
	return m, nil
}

// saveToQuickFavorites saves a station to My-favorites.json
func (m SearchModel) saveToQuickFavorites(station api.Station) tea.Cmd {
	return func() tea.Msg {
		// Check for duplicates
		for _, s := range m.quickFavorites {
			if s.StationUUID == station.StationUUID {
				return saveFailedMsg{
					isDuplicate: true,
				}
			}
		}

		// Add to Quick Favorites
		store := storage.NewStorage(m.favoritePath)
		ctx := context.Background()

		// Load existing list
		list, err := store.LoadList(ctx, "My-favorites")
		if err != nil {
			// Create new list if it doesn't exist
			list = &storage.FavoritesList{
				Name:     "My-favorites",
				Stations: []api.Station{},
			}
		}

		// Add station
		list.Stations = append(list.Stations, station)

		// Save
		if err := store.SaveList(ctx, list); err != nil {
			return saveFailedMsg{
				err:         err,
				isDuplicate: false,
			}
		}

		return saveSuccessMsg{
			station: &station,
		}
	}
}

// voteForStation votes for the currently selected station
func (m SearchModel) voteForStation() tea.Cmd {
	return components.ExecuteVote(m.selectedStation, m.votedStations, m.apiClient)
}

// saveStationVolume saves the updated volume for a station in My-favorites
func (m SearchModel) saveStationVolume(station *api.Station) {
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

// View renders the search screen
func (m SearchModel) View() string {
	if m.helpModel.IsVisible() {
		return m.helpModel.View()
	}

	switch m.state {
	case searchStateMenu:
		return m.renderSearchMenu()

	case searchStateInput:
		var content strings.Builder
		content.WriteString(m.getSearchTypeLabel())
		content.WriteString("\n\n")
		// Add helpful description based on search type
		content.WriteString(m.getSearchTypeDescription())
		content.WriteString("\n\n")
		content.WriteString(m.textInput.View())
		return RenderPageWithBottomHelp(PageLayout{
			Title:   "🔍 Search Radio Stations",
			Content: content.String(),
			Help:    "Enter: Search • Esc: Back • Ctrl+C: Quit",
		}, m.height)

	case searchStateLoading:
		var content strings.Builder
		content.WriteString(m.spinner.View())
		content.WriteString(" Searching for stations...")
		return RenderPage(PageLayout{
			Title:   "🔍 Searching...",
			Content: content.String(),
			Help:    "",
		})

	case searchStateResults:
		if len(m.results) == 0 {
			// Build search criteria display for no results
			var criteria strings.Builder
			criteria.WriteString("No stations found matching your search.\n\n")

			if m.searchType == api.SearchAdvanced {
				criteria.WriteString("Search criteria:\n")
				if tag := strings.TrimSpace(m.advancedInputs[0].Value()); tag != "" {
					criteria.WriteString(fmt.Sprintf("  Tag: %s\n", tag))
				}
				if lang := strings.TrimSpace(m.advancedInputs[1].Value()); lang != "" {
					criteria.WriteString(fmt.Sprintf("  Language: %s\n", lang))
				}
				if country := strings.TrimSpace(m.advancedInputs[2].Value()); country != "" {
					criteria.WriteString(fmt.Sprintf("  Country: %s\n", country))
				}
				if state := strings.TrimSpace(m.advancedInputs[3].Value()); state != "" {
					criteria.WriteString(fmt.Sprintf("  State: %s\n", state))
				}
				if name := strings.TrimSpace(m.advancedInputs[4].Value()); name != "" {
					criteria.WriteString(fmt.Sprintf("  Name: %s\n", name))
				}
				if m.advancedBitrate != "" {
					bitrateText := map[string]string{
						"1": "Low (≤ 64 kbps)",
						"2": "Medium (96-128 kbps)",
						"3": "High (≥ 192 kbps)",
					}
					criteria.WriteString(fmt.Sprintf("  Bitrate: %s\n", bitrateText[m.advancedBitrate]))
				}
			}

			return RenderPage(PageLayout{
				Title:   "🔍 No Results",
				Content: criteria.String(),
				Help:    "Esc: Back to search menu",
			})
		}

		var content strings.Builder
		content.WriteString(m.resultsList.View())

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
			Content: content.String(),
			Help:    "↑↓/jk: Navigate • Enter: Play • Esc: Back • 0: Main Menu • Ctrl+C: Quit",
		})

	case searchStateStationInfo:
		return m.renderStationInfo()

	case searchStateSavePrompt:
		return m.renderSavePrompt()

	case searchStatePlaying:
		var content strings.Builder
		if m.selectedStation != nil {
			// Check if user has voted for this station
			hasVoted := m.votedStations != nil && m.votedStations.HasVoted(m.selectedStation.StationUUID)
			content.WriteString(RenderStationDetailsWithVote(*m.selectedStation, hasVoted))
			// Playback status with proper spacing
			content.WriteString("\n")
			if m.player.IsPlaying() {
				content.WriteString(successStyle().Render("▶ Playing..."))
			} else {
				content.WriteString(infoStyle().Render("⏸ Stopped"))
			}
		}
		if m.saveMessage != "" {
			content.WriteString("\n\n")
			if strings.Contains(m.saveMessage, "✓") || strings.HasPrefix(m.saveMessage, "Volume:") {
				if strings.Contains(m.saveMessage, "Muted") {
					content.WriteString(infoStyle().Render(m.saveMessage))
				} else {
					content.WriteString(successStyle().Render(m.saveMessage))
				}
			} else if strings.Contains(m.saveMessage, "Already") {
				content.WriteString(infoStyle().Render(m.saveMessage))
			} else {
				content.WriteString(errorStyle().Render(m.saveMessage))
			}
		}
		return RenderPageWithBottomHelp(PageLayout{
			Title:   "🎵 Now Playing",
			Content: content.String(),
			Help:    "b: Block • u: Undo • f: Favorites • s: Save to list • v: Vote • ?: Help",
		}, m.height)

	case searchStateSelectList:
		return m.viewSelectList()

	case searchStateNewListInput:
		return m.viewNewListInput()

	case searchStateAdvancedForm:
		return m.viewAdvancedForm()
	}

	return RenderPage(PageLayout{
		Content: "Unknown state",
		Help:    "",
	})
}

// getSearchTypeLabel returns a label for the current search type
func (m SearchModel) getSearchTypeLabel() string {
	switch m.searchType {
	case api.SearchByTag:
		return "Search by Tag (genre, style, etc.)"
	case api.SearchByName:
		return "Search by Station Name"
	case api.SearchByLanguage:
		return "Search by Language"
	case api.SearchByCountry:
		return "Search by Country Code"
	case api.SearchByState:
		return "Search by State"
	case api.SearchAdvanced:
		return "Advanced Search (multiple criteria)"
	default:
		return "Search"
	}
}

// getSearchTypeDescription returns a helpful description for the current search type
func (m SearchModel) getSearchTypeDescription() string {
	switch m.searchType {
	case api.SearchByTag:
		return "Enter a genre or style tag (e.g., jazz, rock, news, classical)"
	case api.SearchByName:
		return "Enter a station name or partial name (e.g., BBC, NPR)"
	case api.SearchByLanguage:
		return "Enter a language (e.g., english, spanish, japanese)"
	case api.SearchByCountry:
		return "Enter a country code (e.g., US, UK, FR, JP)"
	case api.SearchByState:
		return "Enter a state or region (e.g., California, Bavaria)"
	case api.SearchAdvanced:
		return "Searches both station names AND tags.\nUse a word or phrase (e.g., smooth jazz, classic rock)"
	default:
		return ""
	}
}

// renderStationInfo renders the station information and submenu
func (m SearchModel) renderStationInfo() string {
	var content strings.Builder

	if m.selectedStation != nil {
		content.WriteString(RenderStationDetails(*m.selectedStation))
		content.WriteString("\n\n")
	}

	content.WriteString(m.stationInfoMenu.View())

	if m.saveMessage != "" {
		content.WriteString("\n\n")
		if strings.Contains(m.saveMessage, "✓") {
			content.WriteString(successStyle().Render(m.saveMessage))
		} else if strings.Contains(m.saveMessage, "Already") {
			content.WriteString(infoStyle().Render(m.saveMessage))
		} else {
			content.WriteString(errorStyle().Render(m.saveMessage))
		}
	}

	return RenderPage(PageLayout{
		Title:   "📻 Station Information",
		Content: content.String(),
		Help:    "↑↓/jk: Navigate • Enter: Select • 1-3: Quick select • Esc: Back • q: Quit",
	})
}

// renderSavePrompt renders the save prompt after playback
func (m SearchModel) renderSavePrompt() string {
	var content strings.Builder

	if m.selectedStation != nil {
		content.WriteString("Did you enjoy this station?\n\n")
		content.WriteString(boldStyle().Render(m.selectedStation.TrimName()))
		content.WriteString("\n\n")
	}

	content.WriteString("1) ⭐ Add to Quick Favorites\n")
	content.WriteString("2) Return to search results")

	return RenderPage(PageLayout{
		Title:   "💾 Save Station?",
		Content: content.String(),
		Help:    "y/1: Yes • n/2/Esc: No • q: Quit",
	})
}

// viewSelectList renders the list selection view
func (m SearchModel) viewSelectList() string {
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
func (m SearchModel) viewNewListInput() string {
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

// executeHistorySearch executes a search from history
func (m SearchModel) executeHistorySearch(searchType, query string) (tea.Model, tea.Cmd) {
	// Map string search type to api.SearchType
	switch searchType {
	case "tag":
		m.searchType = api.SearchByTag
	case "name":
		m.searchType = api.SearchByName
	case "language":
		m.searchType = api.SearchByLanguage
	case "country":
		m.searchType = api.SearchByCountry
	case "state":
		m.searchType = api.SearchByState
	case "advanced":
		m.searchType = api.SearchAdvanced
	default:
		// Unknown type, go back to menu
		return m, nil
	}

	// Execute search immediately
	m.state = searchStateLoading
	return m, m.performSearch(query)
}

// renderSearchMenu renders the search menu with history
func (m SearchModel) renderSearchMenu() string {
	var content strings.Builder

	t := theme.Current()
	titleStyle := lipgloss.NewStyle().
		Foreground(t.HighlightColor()).
		Bold(true).
		PaddingLeft(t.Padding.ListItemLeft)

	// Render title first
	content.WriteString(titleStyle.Render("🔍 Search Radio Stations"))
	content.WriteString("\n\n")

	// Add "Choose an option:" with number buffer display below title
	content.WriteString(subtitleStyle().Render("Choose an option:"))
	if m.numberBuffer != "" {
		content.WriteString(" ")
		content.WriteString(highlightStyle().Render(m.numberBuffer + "_"))
	}
	content.WriteString("\n\n")

	// Show main menu (includes history items if any)
	content.WriteString(m.menuList.View())

	// Error message if any
	if m.err != nil {
		content.WriteString("\n")
		content.WriteString(errorStyle().Render(fmt.Sprintf("Error: %v", m.err)))
	}

	helpText := "↑↓/jk: Navigate • Enter: Select • 1-6+Enter: Search • 10,11,12...: History • Esc: Back • Ctrl+C: Quit"

	return RenderPageWithBottomHelp(PageLayout{
		Content: content.String(),
		Help:    helpText,
	}, m.height)
}

// reloadSearchHistory reloads history from disk
func (m *SearchModel) reloadSearchHistory() {
	store := storage.NewStorage(m.favoritePath)
	history, err := store.LoadSearchHistory(context.Background())
	if err != nil || history == nil {
		history = storage.NewSearchHistoryStore()
	}
	m.searchHistory = history
}

// rebuildMenuWithHistory rebuilds the menu list to include history items
func (m *SearchModel) rebuildMenuWithHistory() {
	// Base menu items
	menuItems := []components.MenuItem{
		components.NewMenuItem("Search by Tag", "(genre, style, etc.)", "1"),
		components.NewMenuItem("Search by Name", "", "2"),
		components.NewMenuItem("Search by Language", "", "3"),
		components.NewMenuItem("Search by Country Code", "", "4"),
		components.NewMenuItem("Search by State", "", "5"),
		components.NewMenuItem("Advanced Search", "(multiple criteria)", "6"),
	}

	// Add history items if available
	if m.searchHistory != nil && len(m.searchHistory.SearchItems) > 0 {
		// Add empty lines for spacing
		menuItems = append(menuItems, components.NewMenuItem("", "", ""))
		// Add separator-like item
		menuItems = append(menuItems, components.NewMenuItem("─── Recent Searches ───", "", ""))

		// Add history items with numbers starting from 10
		for i, item := range m.searchHistory.SearchItems {
			if i >= m.searchHistory.MaxSize {
				break
			}
			title := fmt.Sprintf("%s: %s", item.SearchType, item.Query)
			// Shortcut keys: 10, 11, 12, etc.
			shortcut := fmt.Sprintf("%d", 10+i)
			menuItems = append(menuItems, components.NewMenuItem(title, "", shortcut))
		}
	}

	// Calculate appropriate height
	height := len(menuItems) + 5
	if height > 20 {
		height = 20
	}

	// Use empty title - we render the title manually in renderSearchMenu
	m.menuList = components.CreateMenu(menuItems, "", 50, height)
}

// handleAdvancedForm handles input in the advanced search form
func (m SearchModel) handleAdvancedForm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		// Cancel and go back to menu
		m.err = nil
		m.state = searchStateMenu
		for i := range m.advancedInputs {
			m.advancedInputs[i].Blur()
		}
		return m, nil

	case "enter":
		// Execute search if at least one field is filled
		if m.hasAtLeastOneField() {
			m.err = nil
			m.state = searchStateLoading
			for i := range m.advancedInputs {
				m.advancedInputs[i].Blur()
			}
			return m, m.performAdvancedSearch()
		}
		// Show error if no fields filled
		m.err = fmt.Errorf("at least one field is required")
		return m, nil

	case "tab", "down":
		// Clear error on navigation
		m.err = nil
		// Blur current text input if on text field
		if m.advancedFocusIdx < 5 {
			m.advancedInputs[m.advancedFocusIdx].Blur()
		}
		// Move to next field (0-6: 5 text fields + sort + bitrate)
		m.advancedFocusIdx = (m.advancedFocusIdx + 1) % 7
		// Focus new text input if on text field
		if m.advancedFocusIdx < 5 {
			m.advancedInputs[m.advancedFocusIdx].Focus()
			return m, textinput.Blink
		}
		return m, nil

	case "shift+tab", "up":
		// Clear error on navigation
		m.err = nil
		// Blur current text input if on text field
		if m.advancedFocusIdx < 5 {
			m.advancedInputs[m.advancedFocusIdx].Blur()
		}
		// Move to previous field (0-6: 5 text fields + sort + bitrate)
		m.advancedFocusIdx = (m.advancedFocusIdx - 1 + 7) % 7
		// Focus new text input if on text field
		if m.advancedFocusIdx < 5 {
			m.advancedInputs[m.advancedFocusIdx].Focus()
			return m, textinput.Blink
		}
		return m, nil

	case "left", "right", " ":
		// Clear error when user takes action
		m.err = nil
		// Handle actions based on focused section
		if m.advancedFocusIdx == 5 {
			// Sort section: left/right/space toggles sort
			m.advancedSortByVotes = !m.advancedSortByVotes
			return m, nil
		}
		// For other sections (like text fields), pass to default handler
		// Fall through to default

	case "1", "2", "3":
		// Clear error when user takes action
		m.err = nil
		// If focused on bitrate section or always allow number input
		if m.advancedFocusIdx == 6 || m.advancedBitrate != "" {
			// Toggle bitrate: if selecting same value, unset it
			if m.advancedBitrate == msg.String() {
				m.advancedBitrate = ""
			} else {
				m.advancedBitrate = msg.String()
			}
			return m, nil
		}
		// Otherwise pass to text input handler
		// Fall through to default
	}

	// Default handler for all other keys
	// Clear error when user starts typing
	m.err = nil
	// Only update text input if we're on a text field (0-4)
	if m.advancedFocusIdx < 5 {
		var cmd tea.Cmd
		m.advancedInputs[m.advancedFocusIdx], cmd = m.advancedInputs[m.advancedFocusIdx].Update(msg)
		return m, cmd
	}
	// If on Sort or Bitrate sections, ignore other keys
	return m, nil
}

// hasAtLeastOneField checks if at least one search field is filled
func (m SearchModel) hasAtLeastOneField() bool {
	for i := range m.advancedInputs {
		if strings.TrimSpace(m.advancedInputs[i].Value()) != "" {
			return true
		}
	}
	return false
}

// performAdvancedSearch executes the advanced search with multiple criteria
func (m SearchModel) performAdvancedSearch() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()

		params := m.buildAdvancedSearchParams()

		// Execute search
		results, err := m.apiClient.SearchAdvanced(ctx, params)
		if err != nil {
			return searchErrorMsg{err: err}
		}

		// Apply bitrate filter if specified
		if m.advancedBitrate != "" {
			results = m.filterByBitrate(results, m.advancedBitrate)
		}

		// Sort results if needed (votes sorting)
		if m.advancedSortByVotes {
			sort.Slice(results, func(i, j int) bool {
				return results[i].Votes > results[j].Votes
			})
		}

		return searchResultsMsg{results: results}
	}
}

// buildAdvancedSearchParams constructs search parameters from the form
func (m SearchModel) buildAdvancedSearchParams() api.SearchParams {
	// Build search params from form
	countryInput := strings.TrimSpace(m.advancedInputs[2].Value())
	var country, countryCode string

	// Uppercase country code if it looks like a 2-letter code
	if len(countryInput) == 2 {
		countryCode = strings.ToUpper(countryInput)
	} else if len(countryInput) > 2 {
		// Title case for full country name
		country = cases.Title(language.English).String(strings.ToLower(countryInput))
	}

	params := api.SearchParams{
		Tag:         strings.TrimSpace(m.advancedInputs[0].Value()),
		Language:    strings.ToLower(strings.TrimSpace(m.advancedInputs[1].Value())),
		Country:     country,
		CountryCode: countryCode,
		State:       strings.TrimSpace(m.advancedInputs[3].Value()),
		Name:        strings.TrimSpace(m.advancedInputs[4].Value()),
		Limit:       100,
		HideBroken:  true,
	}

	// Set sort order
	if m.advancedSortByVotes {
		params.Order = "votes"
		params.Reverse = true
	} else {
		// Relevance = Radio Browser default (no order specified)
		params.Order = ""
		params.Reverse = false
	}

	return params
}

// filterByBitrate filters stations by bitrate range
func (m SearchModel) filterByBitrate(stations []api.Station, bitrateOption string) []api.Station {
	var filtered []api.Station
	for _, station := range stations {
		switch bitrateOption {
		case "1": // Low (≤ 64 kbps)
			if station.Bitrate <= 64 {
				filtered = append(filtered, station)
			}
		case "2": // Medium (96-128 kbps)
			if station.Bitrate >= 96 && station.Bitrate <= 128 {
				filtered = append(filtered, station)
			}
		case "3": // High (≥ 192 kbps)
			if station.Bitrate >= 192 {
				filtered = append(filtered, station)
			}
		}
	}
	return filtered
}

// viewAdvancedForm renders the advanced search form
func (m SearchModel) viewAdvancedForm() string {
	var content strings.Builder

	t := theme.Current()
	labelStyle := lipgloss.NewStyle().
		Foreground(t.TextColor()).
		Width(30)

	focusedLabelStyle := lipgloss.NewStyle().
		Foreground(t.HighlightColor()).
		Bold(true).
		Width(30)

	// Title
	content.WriteString(boldStyle().Render("🔍 Advanced Search"))
	content.WriteString("\n\n")

	// Field labels and inputs
	labels := []string{
		"Tag (optional):",
		"Language (optional):",
		"Country / Country code (optional):",
		"State (optional):",
		"Name contains (optional):",
	}

	for i, label := range labels {
		if i == m.advancedFocusIdx {
			content.WriteString(focusedLabelStyle.Render(label))
		} else {
			content.WriteString(labelStyle.Render(label))
		}
		content.WriteString("  ")
		content.WriteString(m.advancedInputs[i].View())
		content.WriteString("\n")
	}

	content.WriteString("\n")

	// Sort by option (focus index 5)
	sortLabel := "Sort by: "
	isSortFocused := m.advancedFocusIdx == 5
	if isSortFocused {
		// Focused: show as highlighted with arrow
		if m.advancedSortByVotes {
			sortLabel = focusedLabelStyle.Render("▶ Sort by: ") + highlightStyle().Render("votes") + " (default) | relevance"
		} else {
			sortLabel = focusedLabelStyle.Render("▶ Sort by: ") + "votes (default) | " + highlightStyle().Render("relevance")
		}
	} else {
		// Not focused: show normally
		if m.advancedSortByVotes {
			sortLabel += boldStyle().Render("votes") + " (default) | relevance"
		} else {
			sortLabel += "votes (default) | " + boldStyle().Render("relevance")
		}
	}
	content.WriteString(sortLabel)
	content.WriteString("\n\n")

	// Bitrate filter (focus index 6)
	isBitrateFocused := m.advancedFocusIdx == 6
	if isBitrateFocused {
		content.WriteString(focusedLabelStyle.Render("▶ Bitrate (optional):"))
	} else {
		content.WriteString("Bitrate (optional):")
	}
	content.WriteString("\n")

	bitrateOptions := []string{
		"1) Low   (≤ 64 kbps)",
		"2) Medium (96–128 kbps)",
		"3) High  (≥ 192 kbps)",
	}
	for i, option := range bitrateOptions {
		optionNum := fmt.Sprintf("%d", i+1)
		if m.advancedBitrate == optionNum {
			if isBitrateFocused {
				content.WriteString(highlightStyle().Render("✓ " + option))
			} else {
				content.WriteString("✓ " + option)
			}
		} else {
			if isBitrateFocused {
				content.WriteString(labelStyle.Render("  " + option))
			} else {
				content.WriteString("  " + option)
			}
		}
		content.WriteString("\n")
	}

	// Error message if any
	if m.err != nil {
		content.WriteString("\n")
		content.WriteString(errorStyle().Render(fmt.Sprintf("✗ %v", m.err)))
	}

	helpText := "Tab/↑↓: Navigate all fields • Space/←→: Toggle sort • 1/2/3: Select bitrate • Enter: Search • Esc: Cancel"

	return RenderPageWithBottomHelp(PageLayout{
		Content: content.String(),
		Help:    helpText,
	}, m.height)
}

// blockStation blocks the currently playing station
func (m SearchModel) blockStation() tea.Cmd {
	return func() tea.Msg {
		if m.selectedStation == nil {
			return searchErrorMsg{fmt.Errorf("no station selected")}
		}
		if m.blocklistManager == nil {
			return searchErrorMsg{fmt.Errorf("blocklist not available")}
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
			return searchErrorMsg{err}
		}

		return stationBlockedMsg{
			message:     msg,
			stationUUID: m.selectedStation.StationUUID,
			success:     true,
		}
	}
}

// undoLastBlock undoes the last block operation
func (m SearchModel) undoLastBlock() tea.Cmd {
	return func() tea.Msg {
		if m.blocklistManager == nil {
			return undoBlockFailedMsg{}
		}
		ctx := context.Background()
		undone, err := m.blocklistManager.UndoLastBlock(ctx)
		if err != nil {
			return searchErrorMsg{err}
		}

		if undone {
			return undoBlockSuccessMsg{}
		}
		return undoBlockFailedMsg{}
	}
}
