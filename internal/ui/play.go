package ui

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/lipgloss"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/shinokada/tera/internal/api"
	"github.com/shinokada/tera/internal/player"
	"github.com/shinokada/tera/internal/storage"
)

// playState represents the current state in the play screen
type playState int

const (
	playStateListSelection playState = iota
	playStateStationSelection
	playStatePlaying
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
	player           *player.MPVPlayer
	saveMessage      string
	saveMessageTime  int // frames to show message
	width            int
	height           int
	err              error
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
func (i stationListItem) Title() string       { return i.station.TrimName() }
func (i stationListItem) Description() string {
	// Show country and codec/bitrate info
	var parts []string
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

// NewPlayModel creates a new play screen model
func NewPlayModel(favoritePath string) PlayModel {
	return PlayModel{
		state:        playStateListSelection,
		favoritePath: favoritePath,
		lists:        []string{},
		listItems:    []list.Item{},
		player:       player.NewMPVPlayer(),
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
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch m.state {
		case playStateListSelection:
			return m.updateListSelection(msg)
		case playStateStationSelection:
			return m.updateStationSelection(msg)
		case playStatePlaying:
			return m.updatePlaying(msg)
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		if m.listModel.Items() != nil {
			m.listModel.SetSize(msg.Width, msg.Height-10)
		}
		if m.stationListModel.Items() != nil {
			m.stationListModel.SetSize(msg.Width, msg.Height-10)
		}
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

	case listsLoadedMsg:
		m.lists = msg.lists
		m.listItems = make([]list.Item, len(msg.lists))
		for i, name := range msg.lists {
			m.listItems[i] = playListItem{name: name}
		}

		// Initialize the list model
		delegate := list.NewDefaultDelegate()
		m.listModel = list.New(m.listItems, delegate, m.width, m.height-10)
		m.listModel.Title = "Select a Favorite List"
		m.listModel.SetShowStatusBar(false)
		m.listModel.SetFilteringEnabled(false)
		m.listModel.Styles.Title = titleStyle
		m.listModel.Styles.PaginationStyle = paginationStyle
		m.listModel.Styles.HelpStyle = helpStyle

		return m, nil

	case stationsLoadedMsg:
		m.stations = msg.stations
		m.stationItems = make([]list.Item, len(msg.stations))
		for i, station := range msg.stations {
			m.stationItems[i] = stationListItem{station: station}
		}

		// Initialize the station list model with filtering enabled
		delegate := list.NewDefaultDelegate()
		m.stationListModel = list.New(m.stationItems, delegate, m.width, m.height-10)
		m.stationListModel.Title = fmt.Sprintf("Stations in %s", m.selectedList)
		m.stationListModel.SetShowStatusBar(true)
		m.stationListModel.SetFilteringEnabled(true) // Enable fzf-style filtering
		m.stationListModel.Styles.Title = titleStyle
		m.stationListModel.Styles.PaginationStyle = paginationStyle
		m.stationListModel.Styles.HelpStyle = helpStyle

		return m, nil

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

// updateListSelection handles input during list selection
func (m PlayModel) updateListSelection(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "0":
		// Return to main menu
		return m, func() tea.Msg {
			return navigateMsg{screen: screenMainMenu}
		}
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
	case "esc", "0":
		// Go back to list selection
		m.state = playStateListSelection
		m.stations = nil
		m.stationItems = nil
		m.stationListModel = list.Model{}
		return m, nil
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

// stopPlayback stops the current playback
func (m PlayModel) stopPlayback() tea.Cmd {
	return func() tea.Msg {
		if err := m.player.Stop(); err != nil {
			return errMsg{err}
		}
		return playbackStoppedMsg{}
	}
}

// updatePlaying handles input during playback
func (m PlayModel) updatePlaying(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "esc", "0":
		// Stop playback and return to station selection
		m.player.Stop()
		m.state = playStateStationSelection
		m.selectedStation = nil
		m.saveMessage = ""
		m.saveMessageTime = 0
		return m, nil
	case "s":
		// Save to Quick Favorites
		return m, m.saveToQuickFavorites()
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

// View renders the play screen
func (m PlayModel) View() string {
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
	}

	return "Unknown state"
}

// viewListSelection renders the list selection view
func (m PlayModel) viewListSelection() string {
	if len(m.lists) == 0 {
		return noListsView()
	}

	var b strings.Builder

	// Title
	b.WriteString(titleStyle.Render("Play from Favorites"))
	b.WriteString("\n\n")

	// List
	b.WriteString(m.listModel.View())
	b.WriteString("\n\n")

	// Help
	help := helpStyle.Render("↑/↓: navigate • enter: select • esc/0: back to menu")
	b.WriteString(help)

	return b.String()
}

// viewPlaying renders the playback view
func (m PlayModel) viewPlaying() string {
	if m.selectedStation == nil {
		return "No station selected"
	}

	var b strings.Builder

	// Title
	b.WriteString(titleStyle.Render("Now Playing"))
	b.WriteString("\n\n")

	// Station info box
	info := m.formatStationInfo(m.selectedStation)
	b.WriteString(boxStyle.Render(info))
	b.WriteString("\n\n")

	// Playback status
	if m.player.IsPlaying() {
		b.WriteString(successStyle.Render("▶ Playing..."))
	} else {
		b.WriteString(infoStyle.Render("⏸ Stopped"))
	}
	b.WriteString("\n\n")

	// Save message (if any)
	if m.saveMessage != "" {
		// Determine style based on message content
		var style lipgloss.Style
		if strings.Contains(m.saveMessage, "✓") {
			style = successStyle
		} else if strings.Contains(m.saveMessage, "Already") {
			style = infoStyle
		} else {
			style = errorStyle
		}
		b.WriteString(style.Render(m.saveMessage))
		b.WriteString("\n\n")
	}

	// Help
	help := helpStyle.Render("q/esc/0: stop • s: save to favorites")
	b.WriteString(help)

	return b.String()
}

// formatStationInfo formats station information for display
func (m PlayModel) formatStationInfo(station *api.Station) string {
	var b strings.Builder

	// Station name
	b.WriteString(stationNameStyle.Render(station.TrimName()))
	b.WriteString("\n\n")

	// Details
	if station.Country != "" {
		b.WriteString(stationFieldStyle.Render("Country: "))
		b.WriteString(stationValueStyle.Render(station.Country))
		b.WriteString("\n")
	}

	if station.Codec != "" {
		b.WriteString(stationFieldStyle.Render("Codec: "))
		codecInfo := station.Codec
		if station.Bitrate > 0 {
			codecInfo += fmt.Sprintf(" (%d kbps)", station.Bitrate)
		}
		b.WriteString(stationValueStyle.Render(codecInfo))
		b.WriteString("\n")
	}

	if station.Tags != "" {
		b.WriteString(stationFieldStyle.Render("Tags: "))
		b.WriteString(stationValueStyle.Render(station.Tags))
		b.WriteString("\n")
	}

	if station.Votes > 0 {
		b.WriteString(stationFieldStyle.Render("Votes: "))
		b.WriteString(stationValueStyle.Render(fmt.Sprintf("%d", station.Votes)))
	}

	return b.String()
}

// viewStationSelection renders the station selection view
func (m PlayModel) viewStationSelection() string {
	if len(m.stations) == 0 {
		return noStationsView(m.selectedList)
	}

	var b strings.Builder

	// Title
	b.WriteString(titleStyle.Render("Play from Favorites"))
	b.WriteString("\n")
	b.WriteString(subtitleStyle.Render(fmt.Sprintf("List: %s", m.selectedList)))
	b.WriteString("\n\n")

	// Station list
	b.WriteString(m.stationListModel.View())
	b.WriteString("\n\n")

	// Help
	help := helpStyle.Render("↑/↓: navigate • /: filter • enter: play • esc/0: back")
	b.WriteString(help)

	return b.String()
}

// noStationsView renders the view when a list is empty
func noStationsView(listName string) string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("Play from Favorites"))
	b.WriteString("\n")
	b.WriteString(subtitleStyle.Render(fmt.Sprintf("List: %s", listName)))
	b.WriteString("\n\n")
	b.WriteString(infoStyle.Render("This list is empty!"))
	b.WriteString("\n\n")
	b.WriteString("Add stations to this list using Search or List Management.\n\n")
	b.WriteString(helpStyle.Render("Press esc or 0 to go back"))

	return b.String()
}

// noListsView renders the view when no lists are available
func noListsView() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("Play from Favorites"))
	b.WriteString("\n\n")
	b.WriteString(errorStyle.Render("No favorite lists found!"))
	b.WriteString("\n\n")
	b.WriteString("Create your first list using the List Management menu.\n\n")
	b.WriteString(helpStyle.Render("Press esc or 0 to return to main menu"))

	return b.String()
}

// errorView renders an error message
func errorView(err error) string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("Error"))
	b.WriteString("\n\n")
	b.WriteString(errorStyle.Render(err.Error()))
	b.WriteString("\n\n")
	b.WriteString(helpStyle.Render("Press esc or 0 to return to main menu"))

	return b.String()
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

type errMsg struct {
	err error
}
