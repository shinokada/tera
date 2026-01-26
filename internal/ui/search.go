package ui

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/shinokada/tera/internal/api"
	"github.com/shinokada/tera/internal/player"
	"github.com/shinokada/tera/internal/storage"
	"github.com/shinokada/tera/internal/ui/components"
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
)

// SearchModel represents the search screen
type SearchModel struct {
	state           searchState
	searchType      api.SearchType
	menuList        list.Model // List-based menu navigation
	stationInfoMenu list.Model // Station info submenu navigation
	apiClient       *api.Client
	textInput       textinput.Model
	spinner         spinner.Model
	results         []api.Station
	resultsItems    []list.Item
	resultsList     list.Model
	selectedStation *api.Station
	player          *player.MPVPlayer
	favoritePath    string
	quickFavorites  []api.Station // My-favorites.json for duplicate checking
	saveMessage     string
	saveMessageTime int
	width           int
	height          int
	err             error
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

type backToMainMsg struct{}

type playerErrorMsg struct {
	err error
}

// NewSearchModel creates a new search screen model
func NewSearchModel(apiClient *api.Client, favoritePath string) SearchModel {
	ti := textinput.New()
	ti.Placeholder = "Enter search query..."
	ti.CharLimit = 100
	ti.Width = 50

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

	return SearchModel{
		state:           searchStateMenu,
		apiClient:       apiClient,
		menuList:        menuList,
		stationInfoMenu: stationInfoMenu,
		textInput:       ti,
		spinner:         sp,
		favoritePath:    favoritePath,
		player:          player.NewMPVPlayer(),
		quickFavorites:  []api.Station{},
		width:           80, // Default width
		height:          24, // Default height
	}
}

// Init initializes the search screen
func (m SearchModel) Init() tea.Cmd {
	return tea.Batch(
		m.loadQuickFavorites(),
		m.spinner.Tick,
		ticksEverySecond(), // For save message countdown
	)
}

// ticksEverySecond returns a command that ticks every 60th of a second
func ticksEverySecond() tea.Cmd {
	return tea.Tick(time.Second/60, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

type tickMsg time.Time

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

		// Calculate usable height (leaving room for footer only - title is inside list)
		// Footer needs ~3 lines: empty line + help text + our custom footer
		listHeight := msg.Height - 4
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

	case tea.KeyMsg:
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
		}

	case quickFavoritesLoadedMsg:
		m.quickFavorites = msg.stations

	case searchResultsMsg:
		m.results = msg.results
		m.state = searchStateResults
		m.resultsItems = make([]list.Item, 0, len(m.results))
		for _, station := range m.results {
			m.resultsItems = append(m.resultsItems, stationListItem{station: station})
		}

		// Calculate proper list height
		// Footer needs ~3 lines, so leave 4 lines total for safety
		listHeight := m.height - 4
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

	case playbackStoppedMsg:
		// Handle save prompt after playback
		return m.handlePlaybackStopped()

	case saveSuccessMsg:
		// Update local cache
		m.quickFavorites = append(m.quickFavorites, *msg.station)
		m.saveMessage = fmt.Sprintf("✓ Saved '%s' to Quick Favorites", msg.station.TrimName())
		m.saveMessageTime = 150
		return m, nil

	case saveFailedMsg:
		if msg.isDuplicate {
			m.saveMessage = "Already in Quick Favorites"
		} else {
			m.saveMessage = fmt.Sprintf("✗ Failed to save: %v", msg.err)
		}
		m.saveMessageTime = 150
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
		return m, ticksEverySecond()
	}

	return m, tea.Batch(cmds...)
}

// handleMenuInput handles input in the search menu state
func (m SearchModel) handleMenuInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Handle back to main menu
	if msg.String() == "esc" || msg.String() == "m" {
		// Stop any playing station when exiting
		if m.player != nil && m.player.IsPlaying() {
			m.player.Stop()
		}
		m.selectedStation = nil
		return m, func() tea.Msg { return backToMainMsg{} }
	}

	// Handle menu navigation and selection
	newList, selected := components.HandleMenuKey(msg, m.menuList)
	m.menuList = newList

	if selected >= 0 {
		// Execute selected search type
		return m.executeSearchType(selected)
	}

	return m, nil
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
		m.textInput.Placeholder = "Enter search query..."
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
			m.player.Stop()
		}
		m.selectedStation = nil
		m.state = searchStateMenu
		return m, nil
	case "0":
		// Return to main menu
		if m.player != nil && m.player.IsPlaying() {
			m.player.Stop()
		}
		m.selectedStation = nil
		return m, func() tea.Msg { return backToMainMsg{} }
	case "enter":
		// Play station directly
		if item, ok := m.resultsList.SelectedItem().(stationListItem); ok {
			m.selectedStation = &item.station
			// Stop any currently playing station first
			if m.player != nil && m.player.IsPlaying() {
				m.player.Stop()
			}
			m.state = searchStatePlaying
			return m, m.playStation(item.station)
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
			m.player.Stop()
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
			m.player.Stop()
		}
		m.state = searchStatePlaying
		return m, m.playStation(*m.selectedStation)
	case 1: // Save to Quick Favorites
		return m, m.saveToQuickFavorites(*m.selectedStation)
	case 2: // Back to results
		// Stop player when going back
		if m.player != nil && m.player.IsPlaying() {
			m.player.Stop()
		}
		m.selectedStation = nil
		m.state = searchStateResults
		return m, nil
	}
	return m, nil
}

// playStation starts playing a station
func (m SearchModel) playStation(station api.Station) tea.Cmd {
	return func() tea.Msg {
		err := m.player.Play(&station)
		if err != nil {
			return playerErrorMsg{err: err}
		}
		// Return started message, not stopped
		// Player will continue running until user stops it
		return playbackStartedMsg{}
	}
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
			m.saveMessageTime = 150
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
			m.player.Stop()
		}
		return m, tea.Quit
	case "esc":
		// Esc during playback goes back without save prompt
		if m.player != nil {
			m.player.Stop()
		}
		m.selectedStation = nil
		m.state = searchStateResults
		return m, nil
	case "1":
		// Stop playback and trigger save prompt flow
		if m.player != nil {
			m.player.Stop()
		}
		return m.handlePlaybackStopped()
	case "f":
		// Save to Quick Favorites during playback
		return m, m.saveToQuickFavorites(*m.selectedStation)
	case "s":
		// Save to a list (not implemented yet)
		// TODO: Implement save to custom list
		m.saveMessage = "Save to list feature coming soon"
		m.saveMessageTime = 150
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

// View renders the search screen
func (m SearchModel) View() string {
	var s strings.Builder

	switch m.state {
	case searchStateMenu:
		s.WriteString(m.menuList.View())
		s.WriteString("\n")
		s.WriteString(subtleStyle.Render("↑↓/jk: Navigate • Enter: Select • 1-6: Quick select • Esc: Back • Ctrl+C: Quit"))

		if m.err != nil {
			s.WriteString("\n\n")
			s.WriteString(errorStyle.Render(fmt.Sprintf("Error: %v", m.err)))
		}
		return wrapPageWithHeader(s.String())

	case searchStateInput:
		s.WriteString(titleStyle.Render("🔍 Search Radio Stations"))
		s.WriteString("\n\n")
		s.WriteString(m.getSearchTypeLabel())
		s.WriteString("\n\n")
		s.WriteString(m.textInput.View())
		s.WriteString("\n\n")
		s.WriteString(subtleStyle.Render("Enter) Search  |  Esc) Back  |  0) Main Menu  |  Ctrl+C) Quit"))
		return wrapPageWithHeader(s.String())

	case searchStateLoading:
		s.WriteString(titleStyle.Render("🔍 Searching..."))
		s.WriteString("\n\n")
		s.WriteString(m.spinner.View())
		s.WriteString(" Searching for stations...")
		return wrapPageWithHeader(s.String())

	case searchStateResults:
		if len(m.results) == 0 {
			s.WriteString(titleStyle.Render("🔍 No Results"))
			s.WriteString("\n\n")
			s.WriteString("No stations found matching your search.\n\n")
			s.WriteString(subtleStyle.Render("Press Esc to return to search menu"))
		} else {
			s.WriteString(m.resultsList.View())
			s.WriteString("\n")
			s.WriteString(subtleStyle.Render("↑↓/jk: Navigate • Enter: Select  |  Esc) Back  |  0) Main Menu  |  Ctrl+C) Quit"))
		}
		return wrapPageWithHeader(s.String())

	case searchStateStationInfo:
		return wrapPageWithHeader(m.renderStationInfo())

	case searchStateSavePrompt:
		return wrapPageWithHeader(m.renderSavePrompt())

	case searchStatePlaying:
		s.WriteString(titleStyle.Render("🎵 Now Playing"))
		s.WriteString("\n\n")
		if m.selectedStation != nil {
			s.WriteString(renderStationDetails(*m.selectedStation))
		}
		s.WriteString("\n\n")
		s.WriteString(subtleStyle.Render("Esc) Back  |  f) Save to Quick Favorites  |  s) Save to list  |  Ctrl+C) Quit"))

		if m.saveMessage != "" {
			s.WriteString("\n\n")
			if strings.Contains(m.saveMessage, "✓") {
				s.WriteString(successStyle.Render(m.saveMessage))
			} else if strings.Contains(m.saveMessage, "Already") {
				s.WriteString(infoStyle.Render(m.saveMessage))
			} else {
				s.WriteString(errorStyle.Render(m.saveMessage))
			}
		}
		return wrapPageWithHeader(s.String())
	}

	return wrapPageWithHeader(s.String())
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

// renderStationInfo renders the station information and submenu
func (m SearchModel) renderStationInfo() string {
	var s strings.Builder

	s.WriteString(titleStyle.Render("📻 Station Information"))
	s.WriteString("\n\n")

	if m.selectedStation != nil {
		s.WriteString(renderStationDetails(*m.selectedStation))
	}

	s.WriteString("\n\n")
	s.WriteString(m.stationInfoMenu.View())
	s.WriteString("\n")
	s.WriteString(subtleStyle.Render("↑↓/jk: Navigate • Enter: Select • 1-3: Quick select • Esc: Back • q: Quit"))

	if m.saveMessage != "" {
		s.WriteString("\n\n")
		if strings.Contains(m.saveMessage, "✓") {
			s.WriteString(successStyle.Render(m.saveMessage))
		} else if strings.Contains(m.saveMessage, "Already") {
			s.WriteString(infoStyle.Render(m.saveMessage))
		} else {
			s.WriteString(errorStyle.Render(m.saveMessage))
		}
	}

	return s.String()
}

// renderSavePrompt renders the save prompt after playback
func (m SearchModel) renderSavePrompt() string {
	var s strings.Builder

	s.WriteString(titleStyle.Render("💾 Save Station?"))
	s.WriteString("\n\n")

	if m.selectedStation != nil {
		s.WriteString("Did you enjoy this station?\n\n")
		s.WriteString(boldStyle.Render(m.selectedStation.TrimName()))
		s.WriteString("\n\n")
	}

	s.WriteString("1) ⭐ Add to Quick Favorites\n")
	s.WriteString("2) Return to search results\n\n")
	s.WriteString(subtleStyle.Render("y/1: Yes • n/2/Esc: No • q: Quit"))

	return s.String()
}

// renderStationDetails renders station details in a formatted way
func renderStationDetails(station api.Station) string {
	var s strings.Builder

	s.WriteString(fmt.Sprintf("Name:    %s\n", boldStyle.Render(station.TrimName())))

	if station.Tags != "" {
		s.WriteString(fmt.Sprintf("Tags:    %s\n", station.Tags))
	}

	if station.Country != "" {
		s.WriteString(fmt.Sprintf("Country: %s", station.Country))
		if station.State != "" {
			s.WriteString(fmt.Sprintf(", %s", station.State))
		}
		s.WriteString("\n")
	}

	if station.Language != "" {
		s.WriteString(fmt.Sprintf("Language: %s\n", station.Language))
	}

	s.WriteString(fmt.Sprintf("Votes:   %d\n", station.Votes))

	if station.Codec != "" {
		s.WriteString(fmt.Sprintf("Codec:   %s", station.Codec))
		if station.Bitrate > 0 {
			s.WriteString(fmt.Sprintf(" @ %d kbps", station.Bitrate))
		}
		s.WriteString("\n")
	}

	return s.String()
}
