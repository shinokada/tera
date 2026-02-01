package ui

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/shinokada/tera/internal/api"
	"github.com/shinokada/tera/internal/player"
	"github.com/shinokada/tera/internal/storage"
	"github.com/shinokada/tera/internal/ui/components"
)

// playState represents the current state in the play screen
type playState int

const (
	playStateListSelection playState = iota
	playStateStationSelection
	playStatePlaying
	playStateSavePrompt
	playStateDeleteConfirm
)

// PlayModel represents the play screen
type PlayModel struct {
	state            playState
	favoritePath     string
	lists            []string
	listItems        []list.Item
	listModel        list.Model
	selectedList     string
	stations         []api.Station
	stationItems     []list.Item
	stationListModel list.Model
	selectedStation  *api.Station
	stationToDelete  *api.Station
	player           *player.MPVPlayer
	saveMessage      string
	saveMessageTime  int // frames to show message
	width            int
	height           int
	err              error
	listsNeedInit    bool                 // Flag to trigger list model initialization
	stationsNeedInit bool                 // Flag to trigger station model initialization
	helpModel        components.HelpModel // Help overlay
}

// playListItem wraps a list name for the bubbles list
type playListItem struct {
	name string
}

func (i playListItem) FilterValue() string { return i.name }
func (i playListItem) Title() string       { return i.name }
func (i playListItem) Description() string { return "" }

// stationListItem wraps a station for the bubbles list
type stationListItem struct {
	station api.Station
}

func (i stationListItem) FilterValue() string { return i.station.Name }
func (i stationListItem) Title() string {
	// Combine name and info into single line
	var parts []string
	parts = append(parts, i.station.TrimName())

	if i.station.Country != "" {
		parts = append(parts, i.station.Country)
	}
	if i.station.Codec != "" {
		codecInfo := i.station.Codec
		if i.station.Bitrate > 0 {
			codecInfo += fmt.Sprintf(" %dkbps", i.station.Bitrate)
		}
		parts = append(parts, codecInfo)
	}
	return strings.Join(parts, " • ")
}
func (i stationListItem) Description() string {
	// Return empty to show single line
	return ""
}

// NewPlayModel creates a new play screen model
func NewPlayModel(favoritePath string) PlayModel {
	// Note: favorites directory and My-favorites.json are created at app startup
	// in app.go's NewApp() function, so no need to check here
	return PlayModel{
		state:        playStateListSelection,
		favoritePath: favoritePath,
		lists:        []string{},
		listItems:    []list.Item{},
		player:       player.NewMPVPlayer(),
		helpModel:    components.NewHelpModel(components.CreateFavoritesHelp()),
	}
}

// Init initializes the play screen
func (m PlayModel) Init() tea.Cmd {
	return m.loadLists()
}

// loadLists loads all available favorite lists
func (m PlayModel) loadLists() tea.Cmd {
	return func() tea.Msg {
		lists, err := m.getAvailableLists()
		if err != nil {
			return errMsg{err}
		}
		return listsLoadedMsg{lists}
	}
}

// getAvailableLists reads all JSON files from the favorite directory
func (m PlayModel) getAvailableLists() ([]string, error) {
	entries, err := os.ReadDir(m.favoritePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read favorites directory: %w", err)
	}

	var lists []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.HasSuffix(name, ".json") {
			// Remove .json extension
			listName := strings.TrimSuffix(name, ".json")
			lists = append(lists, listName)
		}
	}

	if len(lists) == 0 {
		return nil, fmt.Errorf("no favorite lists found in %s", m.favoritePath)
	}

	return lists, nil
}

// loadStations loads stations from the selected list
func (m PlayModel) loadStations() tea.Cmd {
	return func() tea.Msg {
		stations, err := m.getStationsFromList(m.selectedList)
		if err != nil {
			return errMsg{err}
		}
		return stationsLoadedMsg{stations}
	}
}

// getStationsFromList reads and parses stations from a list file
func (m PlayModel) getStationsFromList(listName string) ([]api.Station, error) {
	store := storage.NewStorage(m.favoritePath)
	list, err := store.LoadList(context.Background(), listName)
	if err != nil {
		return nil, fmt.Errorf("failed to load list %s: %w", listName, err)
	}

	if len(list.Stations) == 0 {
		return []api.Station{}, nil
	}

	// Sort stations alphabetically (case-insensitive)
	stations := list.Stations
	sort.Slice(stations, func(i, j int) bool {
		return strings.ToLower(stations[i].TrimName()) < strings.ToLower(stations[j].TrimName())
	})

	return stations, nil
}

// Update handles messages for the play screen
func (m PlayModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Check if we need to initialize models with dimensions we now have
	if m.listsNeedInit && m.width > 0 && m.height > 0 {
		m.initializeListModel()
		m.listsNeedInit = false
	}
	if m.stationsNeedInit && m.width > 0 && m.height > 0 {
		m.initializeStationListModel()
		m.stationsNeedInit = false
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.helpModel.IsVisible() {
			var cmd tea.Cmd
			m.helpModel, cmd = m.helpModel.Update(msg)
			return m, cmd
		}

		switch m.state {
		case playStateListSelection:
			return m.updateListSelection(msg)
		case playStateStationSelection:
			return m.updateStationSelection(msg)
		case playStatePlaying:
			return m.updatePlaying(msg)
		case playStateSavePrompt:
			return m.updateSavePrompt(msg)
		case playStateDeleteConfirm:
			return m.updateDeleteConfirm(msg)
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		// Calculate usable height
		listHeight := msg.Height - 10
		if listHeight < 5 {
			listHeight = 5
		}

		// Initialize models if we have data but they're not initialized yet
		if len(m.listItems) > 0 && m.listModel.Items() == nil {
			m.initializeListModel()
		} else if m.listModel.Items() != nil && len(m.listModel.Items()) > 0 {
			m.listModel.SetSize(msg.Width, listHeight)
		}

		if len(m.stationItems) > 0 && m.stationListModel.Items() == nil {
			m.initializeStationListModel()
		} else if m.stationListModel.Items() != nil && len(m.stationListModel.Items()) > 0 {
			m.stationListModel.SetSize(msg.Width, listHeight)
		}

		m.helpModel.SetSize(msg.Width, msg.Height)

		return m, nil

	case playbackStartedMsg:
		// Playback started successfully
		return m, nil

	case playbackStoppedMsg:
		// Playback stopped
		return m, nil

	case playbackErrorMsg:
		m.err = msg.err
		m.state = playStateStationSelection
		m.selectedStation = nil
		return m, nil

	case saveSuccessMsg:
		m.saveMessage = fmt.Sprintf("✓ Saved '%s' to Quick Favorites", msg.station.TrimName())
		m.saveMessageTime = 150 // Show for ~3 seconds at 60fps
		return m, nil

	case saveFailedMsg:
		if msg.isDuplicate {
			m.saveMessage = "Already in Quick Favorites"
		} else {
			m.saveMessage = fmt.Sprintf("✗ Failed to save: %v", msg.err)
		}
		m.saveMessageTime = 150
		return m, nil

	case deleteSuccessMsg:
		// Reload stations after successful delete
		return m, m.loadStations()

	case deleteFailedMsg:
		m.err = msg.err
		return m, nil

	case voteSuccessMsg:
		m.saveMessage = fmt.Sprintf("✓ %s", msg.message)
		m.saveMessageTime = 150
		return m, nil

	case voteFailedMsg:
		m.saveMessage = fmt.Sprintf("✗ Vote failed: %v", msg.err)
		m.saveMessageTime = 150
		return m, nil

	case listsLoadedMsg:
		m.lists = msg.lists
		m.listItems = make([]list.Item, len(msg.lists))
		for i, name := range msg.lists {
			m.listItems[i] = playListItem{name: name}
		}

		// Initialize now if we have dimensions, otherwise flag for later
		if m.width > 0 && m.height > 0 {
			m.initializeListModel()
		} else {
			m.listsNeedInit = true
		}

		return m, nil

	case stationsLoadedMsg:
		m.stations = msg.stations
		m.stationItems = make([]list.Item, len(msg.stations))
		for i, station := range msg.stations {
			m.stationItems[i] = stationListItem{station: station}
		}

		// Initialize now if we have dimensions, otherwise flag for later
		if m.width > 0 && m.height > 0 {
			m.initializeStationListModel()
		} else {
			m.stationsNeedInit = true
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
		return m, ticksEverySecond()

	case errMsg:
		m.err = msg.err
		return m, nil
	}

	var cmd tea.Cmd
	if m.state == playStateListSelection && m.listModel.Items() != nil {
		m.listModel, cmd = m.listModel.Update(msg)
	} else if m.state == playStateStationSelection && m.stationListModel.Items() != nil {
		m.stationListModel, cmd = m.stationListModel.Update(msg)
	}
	return m, cmd
}

// initializeListModel creates the list model with current dimensions
func (m *PlayModel) initializeListModel() {
	listHeight := m.height - 10
	if listHeight < 5 {
		listHeight = 5
	}

	delegate := createStyledDelegate()

	m.listModel = list.New(m.listItems, delegate, m.width, listHeight)
	m.listModel.Title = "" // No title - we'll add it in the view
	m.listModel.SetShowStatusBar(false)
	m.listModel.SetFilteringEnabled(false)
	m.listModel.SetShowHelp(false) // Disable built-in help to use custom help text
	m.listModel.Styles.Title = titleStyle()
	m.listModel.Styles.PaginationStyle = paginationStyle()
	m.listModel.Styles.HelpStyle = helpStyle()
}

// initializeStationListModel creates the station list model with current dimensions
func (m *PlayModel) initializeStationListModel() {
	listHeight := m.height - 10
	if listHeight < 5 {
		listHeight = 5
	}

	delegate := createStyledDelegate()

	m.stationListModel = list.New(m.stationItems, delegate, m.width, listHeight)
	m.stationListModel.Title = fmt.Sprintf("Stations in %s", m.selectedList)
	m.stationListModel.SetShowStatusBar(true)
	m.stationListModel.SetFilteringEnabled(true) // Enable fzf-style filtering
	m.stationListModel.SetShowHelp(false)        // Disable built-in help to use custom help text
	m.stationListModel.Styles.Title = listTitleStyle()
	m.stationListModel.Styles.PaginationStyle = paginationStyle()
	m.stationListModel.Styles.HelpStyle = helpStyle()
}

// updateListSelection handles input during list selection
func (m PlayModel) updateListSelection(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		// Return to main menu
		return m, func() tea.Msg {
			return navigateMsg{screen: screenMainMenu}
		}
	case "ctrl+c":
		// Quit application
		return m, tea.Quit
	case "q":
		// Prevent 'q' from quitting - do nothing or return to main menu
		return m, nil
	case "enter":
		// Select list and move to station selection
		if i, ok := m.listModel.SelectedItem().(playListItem); ok {
			m.selectedList = i.name
			m.state = playStateStationSelection
			return m, m.loadStations()
		}
	}

	var cmd tea.Cmd
	m.listModel, cmd = m.listModel.Update(msg)
	return m, cmd
}

// updateStationSelection handles input during station selection
func (m PlayModel) updateStationSelection(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		// Go back to list selection
		m.state = playStateListSelection
		m.stations = nil
		m.stationItems = nil
		m.stationListModel = list.Model{}
		return m, nil
	case "0":
		// Return to main menu (Level 2+ shortcut)
		return m, func() tea.Msg {
			return navigateMsg{screen: screenMainMenu}
		}
	case "ctrl+c":
		// Quit application
		return m, tea.Quit
	case "q":
		// Prevent 'q' from quitting - do nothing
		return m, nil
	case "d":
		// Show delete confirmation
		if i, ok := m.stationListModel.SelectedItem().(stationListItem); ok {
			m.stationToDelete = &i.station
			m.state = playStateDeleteConfirm
			return m, nil
		}
	case "enter":
		// Select station and start playback
		if i, ok := m.stationListModel.SelectedItem().(stationListItem); ok {
			m.selectedStation = &i.station
			m.state = playStatePlaying
			// Start playback
			return m, m.startPlayback()
		}
	}

	var cmd tea.Cmd
	m.stationListModel, cmd = m.stationListModel.Update(msg)
	return m, cmd
}

// startPlayback initiates playback of the selected station
func (m PlayModel) startPlayback() tea.Cmd {
	return func() tea.Msg {
		if m.selectedStation == nil {
			return errMsg{fmt.Errorf("no station selected")}
		}

		if err := m.player.Play(m.selectedStation); err != nil {
			return playbackErrorMsg{err}
		}

		return playbackStartedMsg{}
	}
}

// updatePlaying handles input during playback
func (m PlayModel) updatePlaying(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		// Stop playback and go back
		if err := m.player.Stop(); err != nil {
			m.err = fmt.Errorf("failed to stop playback: %w", err)
			return m, nil
		}
		m.state = playStateStationSelection
		m.selectedStation = nil
		return m, nil
	case "0":
		// Return to main menu (Level 3 shortcut)
		if err := m.player.Stop(); err != nil {
			m.err = fmt.Errorf("failed to stop playback: %w", err)
			return m, nil
		}
		m.selectedStation = nil
		return m, func() tea.Msg {
			return navigateMsg{screen: screenMainMenu}
		}
	case "f":
		// Save to Quick Favorites
		return m, m.saveToQuickFavorites()
	case "s":
		// Save to a list (not implemented yet)
		// TODO: Implement save to custom list
		m.saveMessage = "Save to list feature coming soon"
		m.saveMessageTime = 150
		return m, nil
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
		m.saveMessageTime = 120 // Show for 2 seconds (60 ticks/sec)
		if startTick {
			return m, ticksEverySecond()
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
		m.saveMessageTime = 120 // Show for 2 seconds (60 ticks/sec)
		if startTick {
			return m, ticksEverySecond()
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
		m.saveMessageTime = 120 // Show for 2 seconds (60 ticks/sec)
		if startTick {
			return m, ticksEverySecond()
		}
		return m, nil
	case "?":
		m.helpModel.SetSize(m.width, m.height)
		m.helpModel.Toggle()
		return m, nil
	}
	return m, nil
}

// updateSavePrompt handles input during save prompt
func (m PlayModel) updateSavePrompt(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "n", "2":
		// Don't save, return to station selection
		m.state = playStateStationSelection
		m.selectedStation = nil
		return m, nil
	case "y", "1":
		// Save to Quick Favorites and return to station selection
		cmd := m.saveToQuickFavorites()
		m.state = playStateStationSelection
		return m, cmd
	}
	return m, nil
}

// updateDeleteConfirm handles input during delete confirmation
func (m PlayModel) updateDeleteConfirm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "n":
		// Cancel delete, return to station selection
		m.state = playStateStationSelection
		m.stationToDelete = nil
		return m, nil
	case "y":
		// Confirm delete
		cmd := m.deleteStationFromList(m.stationToDelete)
		m.state = playStateStationSelection
		m.stationToDelete = nil
		return m, cmd
	}
	return m, nil
}

// saveToQuickFavorites saves the current station to My-favorites.json
func (m PlayModel) saveToQuickFavorites() tea.Cmd {
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

// deleteStationFromList removes a station from the current list
func (m PlayModel) deleteStationFromList(station *api.Station) tea.Cmd {
	return func() tea.Msg {
		if station == nil {
			return deleteFailedMsg{err: fmt.Errorf("no station to delete")}
		}

		store := storage.NewStorage(m.favoritePath)
		err := store.RemoveStation(context.Background(), m.selectedList, station.StationUUID)

		if err != nil {
			return deleteFailedMsg{err: err}
		}

		return deleteSuccessMsg{stationName: station.TrimName()}
	}
}

// saveStationVolume saves the updated volume for a station in the current list
func (m PlayModel) saveStationVolume(station *api.Station) {
	if station == nil || m.selectedList == "" {
		return
	}

	store := storage.NewStorage(m.favoritePath)
	list, err := store.LoadList(context.Background(), m.selectedList)
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
}

// voteForStation votes for the currently playing station
func (m PlayModel) voteForStation() tea.Cmd {
	return func() tea.Msg {
		if m.selectedStation == nil {
			return voteFailedMsg{err: fmt.Errorf("no station selected")}
		}

		client := api.NewClient()
		result, err := client.Vote(context.Background(), m.selectedStation.StationUUID)
		if err != nil {
			return voteFailedMsg{err: err}
		}

		if !result.OK {
			return voteFailedMsg{err: fmt.Errorf("%s", result.Message)}
		}

		return voteSuccessMsg{message: "Voted for " + m.selectedStation.TrimName()}
	}
}

// View renders the play screen
func (m PlayModel) View() string {
	if m.helpModel.IsVisible() {
		return m.helpModel.View()
	}

	if m.err != nil {
		return errorView(m.err)
	}

	switch m.state {
	case playStateListSelection:
		return m.viewListSelection()
	case playStateStationSelection:
		return m.viewStationSelection()
	case playStatePlaying:
		return m.viewPlaying()
	case playStateSavePrompt:
		return m.viewSavePrompt()
	case playStateDeleteConfirm:
		return m.viewDeleteConfirm()
	}

	return "Unknown state"
}

// viewListSelection renders the list selection view
func (m PlayModel) viewListSelection() string {
	// Check if we have lists but no model yet (waiting for dimensions)
	if len(m.lists) > 0 && m.listModel.Items() == nil {
		return "Loading..."
	}

	if len(m.lists) == 0 {
		return noListsView()
	}

	// Use the consistent page template
	return RenderPage(PageLayout{
		Title:    "Play from Favorites",
		Subtitle: "Select a Favorite List",
		Content:  m.listModel.View(),
		Help:     "↑↓/jk: Navigate • Enter: Select • Esc: Back • Ctrl+C: Quit",
	})
}

// viewPlaying renders the playback view
func (m PlayModel) viewPlaying() string {
	if m.selectedStation == nil {
		return "No station selected"
	}

	var content strings.Builder

	// Station info (consistent format across all playing views)
	content.WriteString(RenderStationDetails(*m.selectedStation))

	// Playback status with proper spacing
	content.WriteString("\n")
	if m.player.IsPlaying() {
		content.WriteString(successStyle().Render("▶ Playing..."))
	} else {
		content.WriteString(infoStyle().Render("⏸ Stopped"))
	}

	// Save message (if any)
	if m.saveMessage != "" {
		content.WriteString("\n\n")
		// Determine style based on message content
		var style lipgloss.Style
		if strings.Contains(m.saveMessage, "✓") || strings.HasPrefix(m.saveMessage, "Volume:") {
			if strings.Contains(m.saveMessage, "Muted") {
				style = infoStyle()
			} else {
				style = successStyle()
			}
		} else if strings.Contains(m.saveMessage, "Already") {
			style = infoStyle()
		} else {
			style = errorStyle()
		}
		content.WriteString(style.Render(m.saveMessage))
	}

	// Use the consistent page template with bottom-aligned help
	return RenderPageWithBottomHelp(PageLayout{
		Title:   "🎵 Now Playing",
		Content: content.String(),
		Help:    "f: Favorites • v: Vote • 0: Main Menu • ?: Help",
	}, m.height)
}

// viewStationSelection renders the station selection view
func (m PlayModel) viewStationSelection() string {
	// Check if we have stations but no model yet (waiting for dimensions)
	if len(m.stations) > 0 && m.stationListModel.Items() == nil {
		return "Loading..."
	}

	if len(m.stations) == 0 {
		return noStationsView(m.selectedList)
	}

	// Use the consistent page template
	return RenderPage(PageLayout{
		Title:   "Play from Favorites",
		Content: m.stationListModel.View(),
		Help:    "↑↓/jk: Navigate • Enter: Play • d: Delete • Esc: Back • 0: Main Menu • Ctrl+C: Quit",
	})
}

// noStationsView renders the view when a list is empty
func noStationsView(listName string) string {
	var content strings.Builder

	content.WriteString(infoStyle().Render("This list is empty!"))
	content.WriteString("\n\n")
	content.WriteString("Add stations to this list using Search or List Management.")

	return RenderPage(PageLayout{
		Title:    "Play from Favorites",
		Subtitle: fmt.Sprintf("List: %s", listName),
		Content:  content.String(),
		Help:     "Esc: Back • Ctrl+C: Quit",
	})
}

// noListsView renders the view when no lists are available
func noListsView() string {
	var content strings.Builder

	content.WriteString(errorStyle().Render("No favorite lists found!"))
	content.WriteString("\n\n")
	content.WriteString("Create your first list using the List Management menu.")

	return RenderPage(PageLayout{
		Title:   "Play from Favorites",
		Content: content.String(),
		Help:    "Esc: Back to main menu • Ctrl+C: Quit",
	})
}

// errorView renders an error message
func errorView(err error) string {
	return RenderPage(PageLayout{
		Title:   "Error",
		Content: errorStyle().Render(err.Error()),
		Help:    "Esc: Back to main menu • Ctrl+C: Quit",
	})
}

// viewSavePrompt renders the save prompt after playback
func (m PlayModel) viewSavePrompt() string {
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
	content.WriteString("2) Return to station list")

	// Use the consistent page template
	return RenderPage(PageLayout{
		Title:   "💾 Save Station?",
		Content: content.String(),
		Help:    "y/1: Yes • n/2/Esc: No",
	})
}

// viewDeleteConfirm renders the delete confirmation prompt
func (m PlayModel) viewDeleteConfirm() string {
	if m.stationToDelete == nil {
		return "No station selected"
	}

	var content strings.Builder

	// Warning message
	content.WriteString(errorStyle().Render("⚠ Delete Station?"))
	content.WriteString("\n\n")

	// Station name
	content.WriteString("Station: ")
	content.WriteString(stationNameStyle().Render(m.stationToDelete.TrimName()))
	content.WriteString("\n")
	content.WriteString("From list: ")
	content.WriteString(stationValueStyle().Render(m.selectedList))
	content.WriteString("\n\n")

	// Confirmation question
	content.WriteString("Are you sure you want to delete this station?\n")
	content.WriteString(infoStyle().Render("This action cannot be undone."))

	// Use the consistent page template
	return RenderPage(PageLayout{
		Title:   "⚠️  Confirm Delete",
		Content: content.String(),
		Help:    "y: Yes, delete • n/Esc: No, cancel",
	})
}

// Messages

type listsLoadedMsg struct {
	lists []string
}

type stationsLoadedMsg struct {
	stations []api.Station
}

type playbackStartedMsg struct{}

type playbackStoppedMsg struct{}

type playbackErrorMsg struct {
	err error
}

type saveSuccessMsg struct {
	station *api.Station
}

type saveFailedMsg struct {
	err         error
	isDuplicate bool
}

type deleteSuccessMsg struct {
	stationName string
}

type deleteFailedMsg struct {
	err error
}

type voteSuccessMsg struct {
	message string
}

type voteFailedMsg struct {
	err error
}

type errMsg struct {
	err error
}
