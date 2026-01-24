package ui

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/shinokada/tera/internal/api"
	"github.com/shinokada/tera/internal/player"
	"github.com/shinokada/tera/internal/storage"
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
)

// SearchModel represents the search screen
type SearchModel struct {
	state            searchState
	searchType       api.SearchType
	apiClient        *api.Client
	textInput        textinput.Model
	spinner          spinner.Model
	results          []api.Station
	resultsItems     []list.Item
	resultsList      list.Model
	selectedStation  *api.Station
	player           *player.MPVPlayer
	favoritePath     string
	quickFavorites   []api.Station // My-favorites.json for duplicate checking
	saveMessage      string
	saveMessageTime  int
	width            int
	height           int
	err              error
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

	return SearchModel{
		state:          searchStateMenu,
		apiClient:      apiClient,
		textInput:      ti,
		spinner:        sp,
		favoritePath:   favoritePath,
		player:         player.NewMPVPlayer(),
		quickFavorites: []api.Station{},
	}
}

// Init initializes the search screen
func (m SearchModel) Init() tea.Cmd {
	return tea.Batch(
		m.loadQuickFavorites(),
		m.spinner.Tick,
	)
}

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

		if m.state == searchStateResults && m.resultsList.Items() != nil {
			m.resultsList.SetSize(msg.Width, msg.Height-10)
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

		// Create results list
		delegate := list.NewDefaultDelegate()
		m.resultsList = list.New(m.resultsItems, delegate, m.width, m.height-10)
		m.resultsList.Title = fmt.Sprintf("Search Results (%d stations)", len(m.results))
		m.resultsList.SetShowHelp(true)
		m.resultsList.SetFilteringEnabled(true)
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

	case playbackStoppedMsg:
		// Handle save prompt after playback
		return m.handlePlaybackStopped()

	case saveSuccessMsg:
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
	}

	return m, tea.Batch(cmds...)
}

// handleMenuInput handles input in the search menu state
func (m SearchModel) handleMenuInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "0", "esc":
		// Return to main menu
		return m, func() tea.Msg { return backToMainMsg{} }
	case "1":
		m.searchType = api.SearchByTag
		m.state = searchStateInput
		m.textInput.Placeholder = "Enter tag (e.g., jazz, rock, news)..."
		m.textInput.Focus()
		return m, nil
	case "2":
		m.searchType = api.SearchByName
		m.state = searchStateInput
		m.textInput.Placeholder = "Enter station name..."
		m.textInput.Focus()
		return m, nil
	case "3":
		m.searchType = api.SearchByLanguage
		m.state = searchStateInput
		m.textInput.Placeholder = "Enter language (e.g., english, spanish)..."
		m.textInput.Focus()
		return m, nil
	case "4":
		m.searchType = api.SearchByCountry
		m.state = searchStateInput
		m.textInput.Placeholder = "Enter country code (e.g., US, UK, FR)..."
		m.textInput.Focus()
		return m, nil
	case "5":
		m.searchType = api.SearchByState
		m.state = searchStateInput
		m.textInput.Placeholder = "Enter state (e.g., California, Texas)..."
		m.textInput.Focus()
		return m, nil
	case "6":
		m.searchType = api.SearchAdvanced
		m.state = searchStateInput
		m.textInput.Placeholder = "Enter search query..."
		m.textInput.Focus()
		return m, nil
	}
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
	case "0":
		m.textInput.SetValue("")
		m.textInput.Blur()
		m.state = searchStateMenu
		return m, nil
	case "00", "esc":
		return m, func() tea.Msg { return backToMainMsg{} }
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
		m.state = searchStateMenu
		return m, nil
	case "enter":
		// Show station info and submenu
		if item, ok := m.resultsList.SelectedItem().(stationListItem); ok {
			m.selectedStation = &item.station
			m.state = searchStateStationInfo
			return m, nil
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
	switch msg.String() {
	case "0":
		return m, func() tea.Msg { return backToMainMsg{} }
	case "1":
		// Play station
		m.state = searchStatePlaying
		return m, m.playStation(*m.selectedStation)
	case "2":
		// Save to Quick Favorites
		return m, m.saveToQuickFavorites(*m.selectedStation)
	case "3", "esc":
		// Back to results
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
		return playbackStoppedMsg{}
	}
}

// handlePlaybackStopped handles return to results after playback
func (m SearchModel) handlePlaybackStopped() (tea.Model, tea.Cmd) {
	m.state = searchStateResults
	return m, nil
}

// handlePlayerUpdate handles player-related updates during playback
func (m SearchModel) handlePlayerUpdate(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "esc", "0":
		// Stop playback
		if m.player != nil {
			m.player.Stop()
		}
		m.state = searchStateResults
		return m, nil
	case "s":
		// Save to Quick Favorites during playback
		return m, m.saveToQuickFavorites(*m.selectedStation)
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

		// Update local cache
		m.quickFavorites = append(m.quickFavorites, station)

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
		s.WriteString(titleStyle.Render("🔍 Search Radio Stations"))
		s.WriteString("\n\n")
		s.WriteString("1) Search by Tag\n")
		s.WriteString("2) Search by Name\n")
		s.WriteString("3) Search by Language\n")
		s.WriteString("4) Search by Country Code\n")
		s.WriteString("5) Search by State\n")
		s.WriteString("6) Advanced Search\n")
		s.WriteString("\n")
		s.WriteString(subtleStyle.Render("0/Esc) Back to Main Menu"))

		if m.err != nil {
			s.WriteString("\n\n")
			s.WriteString(errorStyle.Render(fmt.Sprintf("Error: %v", m.err)))
		}

	case searchStateInput:
		s.WriteString(titleStyle.Render("🔍 Search Radio Stations"))
		s.WriteString("\n\n")
		s.WriteString(m.getSearchTypeLabel())
		s.WriteString("\n\n")
		s.WriteString(m.textInput.View())
		s.WriteString("\n\n")
		s.WriteString(subtleStyle.Render("Enter) Search  |  0) Back  |  00/Esc) Main Menu"))

	case searchStateLoading:
		s.WriteString(titleStyle.Render("🔍 Searching..."))
		s.WriteString("\n\n")
		s.WriteString(m.spinner.View())
		s.WriteString(" Searching for stations...")

	case searchStateResults:
		if len(m.results) == 0 {
			s.WriteString(titleStyle.Render("🔍 No Results"))
			s.WriteString("\n\n")
			s.WriteString("No stations found matching your search.\n\n")
			s.WriteString(subtleStyle.Render("Press Esc to return to search menu"))
		} else {
			s.WriteString(m.resultsList.View())
			s.WriteString("\n")
			s.WriteString(subtleStyle.Render("Enter) Select  |  /) Filter  |  Esc) Back"))
		}

	case searchStateStationInfo:
		s.WriteString(m.renderStationInfo())

	case searchStatePlaying:
		s.WriteString(titleStyle.Render("🎵 Now Playing"))
		s.WriteString("\n\n")
		if m.selectedStation != nil {
			s.WriteString(renderStationDetails(*m.selectedStation))
		}
		s.WriteString("\n\n")
		s.WriteString(subtleStyle.Render("q/Esc/0) Stop  |  s) Save to Quick Favorites"))

		if m.saveMessage != "" {
			s.WriteString("\n\n")
			if strings.Contains(m.saveMessage, "✓") {
				s.WriteString(successStyle.Render(m.saveMessage))
			} else if strings.Contains(m.saveMessage, "Already") {
				s.WriteString(infoStyle.Render(m.saveMessage))
			} else {
				s.WriteString(errorStyle.Render(m.saveMessage))
			}
			m.saveMessageTime--
			if m.saveMessageTime <= 0 {
				m.saveMessage = ""
			}
		}
	}

	return s.String()
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
	s.WriteString(titleStyle.Render("What would you like to do?"))
	s.WriteString("\n\n")
	s.WriteString("1) Play this station\n")
	s.WriteString("2) Save to Quick Favorites\n")
	s.WriteString("3) Back to search results\n")
	s.WriteString("\n")
	s.WriteString(subtleStyle.Render("0) Main Menu  |  Esc) Back"))

	if m.saveMessage != "" {
		s.WriteString("\n\n")
		if strings.Contains(m.saveMessage, "✓") {
			s.WriteString(successStyle.Render(m.saveMessage))
		} else if strings.Contains(m.saveMessage, "Already") {
			s.WriteString(infoStyle.Render(m.saveMessage))
		} else {
			s.WriteString(errorStyle.Render(m.saveMessage))
		}
		m.saveMessageTime--
		if m.saveMessageTime <= 0 {
			m.saveMessage = ""
		}
	}

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
