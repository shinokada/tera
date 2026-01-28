package ui

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/shinokada/tera/internal/api"
	"github.com/shinokada/tera/internal/player"
	"github.com/shinokada/tera/internal/storage"
)

// luckyState represents the current state in the lucky screen
type luckyState int

const (
	luckyStateInput luckyState = iota
	luckyStateSearching
	luckyStatePlaying
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
	selectedStation *api.Station
	player          *player.MPVPlayer
	favoritePath    string
	saveMessage     string
	saveMessageTime int
	width           int
	height          int
	err             error
	availableLists  []string
	listItems       []list.Item
	listModel       list.Model
}

// Messages for lucky screen
type luckySearchResultsMsg struct {
	station *api.Station
}

type luckySearchErrorMsg struct {
	err error
}

type saveToListSuccessMsg struct {
	listName    string
	stationName string
}

type saveToListFailedMsg struct {
	err         error
	isDuplicate bool
}

// NewLuckyModel creates a new lucky screen model
func NewLuckyModel(apiClient *api.Client, favoritePath string) LuckyModel {
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

	return LuckyModel{
		state:        luckyStateInput,
		apiClient:    apiClient,
		textInput:    ti,
		newListInput: nli,
		favoritePath: favoritePath,
		player:       player.NewMPVPlayer(),
		width:        80,
		height:       24,
	}
}

// Init initializes the lucky screen
func (m LuckyModel) Init() tea.Cmd {
	return tea.Batch(textinput.Blink, ticksEverySecond())
}

// Update handles messages for the lucky screen
func (m LuckyModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch m.state {
		case luckyStateInput:
			return m.updateInput(msg)
		case luckyStatePlaying:
			return m.updatePlaying(msg)
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
		return m, nil

	case luckySearchResultsMsg:
		m.selectedStation = msg.station
		m.state = luckyStatePlaying
		// Start playback immediately
		return m, m.startPlayback()

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

	case voteSuccessMsg:
		m.saveMessage = fmt.Sprintf("✓ %s", msg.message)
		m.saveMessageTime = 150
		return m, nil

	case voteFailedMsg:
		m.saveMessage = fmt.Sprintf("✗ Vote failed: %v", msg.err)
		m.saveMessageTime = 150
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
		m.saveMessageTime = 150
		m.state = luckyStatePlaying
		return m, nil

	case saveToListFailedMsg:
		if msg.isDuplicate {
			m.saveMessage = "Already in this list"
		} else {
			m.saveMessage = fmt.Sprintf("✗ Failed to save: %v", msg.err)
		}
		m.saveMessageTime = 150
		m.state = luckyStatePlaying
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
	}

	var cmd tea.Cmd
	if m.state == luckyStateInput {
		m.textInput, cmd = m.textInput.Update(msg)
	}
	return m, cmd
}

// updateInput handles input during the input state
func (m LuckyModel) updateInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		// Return to main menu
		return m, func() tea.Msg {
			return navigateMsg{screen: screenMainMenu}
		}
	case "ctrl+c":
		return m, tea.Quit
	case "enter":
		// Search for stations with the entered keyword
		keyword := strings.TrimSpace(m.textInput.Value())
		if keyword == "" {
			m.err = fmt.Errorf("please enter a keyword")
			return m, nil
		}
		m.err = nil
		m.state = luckyStateSearching
		return m, m.searchAndPickRandom(keyword)
	}

	var cmd tea.Cmd
	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

// updatePlaying handles input during playback
func (m LuckyModel) updatePlaying(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		// Stop playback and show save prompt
		m.player.Stop()
		m.state = luckyStateSavePrompt
		return m, nil
	case "0":
		// Return to main menu (Level 2+ shortcut)
		m.player.Stop()
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
	}
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
	listHeight := m.height - 10
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
		// Search by tag (genre/keyword)
		stations, err := m.apiClient.SearchByTag(context.Background(), keyword)
		if err != nil {
			return luckySearchErrorMsg{err: fmt.Errorf("search failed: %w", err)}
		}

		if len(stations) == 0 {
			return luckySearchErrorMsg{err: fmt.Errorf("no stations found for '%s'", keyword)}
		}

		// Pick a random station (rand is auto-seeded since Go 1.20)
		randomIndex := rand.Intn(len(stations))
		selectedStation := stations[randomIndex]

		return luckySearchResultsMsg{station: &selectedStation}
	}
}

// startPlayback initiates playback of the selected station
func (m LuckyModel) startPlayback() tea.Cmd {
	return func() tea.Msg {
		if m.selectedStation == nil {
			return playbackErrorMsg{fmt.Errorf("no station selected")}
		}

		if err := m.player.Play(m.selectedStation); err != nil {
			return playbackErrorMsg{err}
		}

		return playbackStartedMsg{}
	}
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
	return func() tea.Msg {
		if m.selectedStation == nil {
			return voteFailedMsg{err: fmt.Errorf("no station selected")}
		}

		// Reuse injected API client for consistency and testability
		client := m.apiClient
		if client == nil {
			client = api.NewClient()
		}
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

// View renders the lucky screen
func (m LuckyModel) View() string {
	switch m.state {
	case luckyStateInput:
		return m.viewInput()
	case luckyStateSearching:
		return m.viewSearching()
	case luckyStatePlaying:
		return m.viewPlaying()
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

	// Instructions
	content.WriteString("Type a genre of music: rock, classical, jazz, pop, country, hip, heavy, blues, soul.\n")
	content.WriteString("Or type a keyword like: meditation, relax, mozart, Beatles, etc.\n\n")
	content.WriteString(infoStyle().Render("Use only one word."))
	content.WriteString("\n\n")

	// Input field
	content.WriteString("Genre/keyword: ")
	content.WriteString(m.textInput.View())

	// Error message if any
	if m.err != nil {
		content.WriteString("\n\n")
		content.WriteString(errorStyle().Render(m.err.Error()))
	}

	return RenderPageWithBottomHelp(PageLayout{
		Title:   "I Feel Lucky",
		Content: content.String(),
		Help:    "Enter: Search • Esc: Back • Ctrl+C: Quit",
	}, m.height)
}

// viewSearching renders the searching view
func (m LuckyModel) viewSearching() string {
	var content strings.Builder

	content.WriteString(infoStyle().Render("🔍 Searching for stations..."))
	content.WriteString("\n\n")
	content.WriteString("Finding a random station for you...")

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

	// Station info box
	info := m.formatStationInfo(m.selectedStation)
	content.WriteString(boxStyle().Render(info))
	content.WriteString("\n\n")

	// Playback status
	if m.player.IsPlaying() {
		content.WriteString(successStyle().Render("▶ Playing..."))
	} else {
		content.WriteString(infoStyle().Render("⏸ Stopped"))
	}

	// Save message (if any)
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

	return RenderPageWithBottomHelp(PageLayout{
		Title:   "🎵 Now Playing",
		Content: content.String(),
		Help:    "Esc: Stop • f: Save to Favorites • s: Save to list • v: Vote • 0: Main Menu • Ctrl+C: Quit",
	}, m.height)
}

// formatStationInfo formats station information for display
func (m LuckyModel) formatStationInfo(station *api.Station) string {
	var b strings.Builder

	// Station name
	b.WriteString(stationNameStyle().Render(station.TrimName()))
	b.WriteString("\n\n")

	// Details
	if station.Country != "" {
		b.WriteString(stationFieldStyle().Render("Country: "))
		b.WriteString(stationValueStyle().Render(station.Country))
		b.WriteString("\n")
	}

	if station.Codec != "" {
		b.WriteString(stationFieldStyle().Render("Codec: "))
		codecInfo := station.Codec
		if station.Bitrate > 0 {
			codecInfo += fmt.Sprintf(" (%d kbps)", station.Bitrate)
		}
		b.WriteString(stationValueStyle().Render(codecInfo))
		b.WriteString("\n")
	}

	if station.Tags != "" {
		b.WriteString(stationFieldStyle().Render("Tags: "))
		b.WriteString(stationValueStyle().Render(station.Tags))
		b.WriteString("\n")
	}

	if station.Votes > 0 {
		b.WriteString(stationFieldStyle().Render("Votes: "))
		b.WriteString(stationValueStyle().Render(fmt.Sprintf("%d", station.Votes)))
	}

	return b.String()
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
