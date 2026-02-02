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
	"github.com/shinokada/tera/internal/theme"
	"github.com/shinokada/tera/internal/ui/components"
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

	// Load search history (if it fails, just use empty history)
	store := storage.NewStorage(favoritePath)
	history, err := store.LoadSearchHistory(context.Background())
	if err != nil || history == nil {
		history = storage.NewSearchHistoryStore()
	}

	m := LuckyModel{
		state:         luckyStateInput,
		apiClient:     apiClient,
		textInput:     ti,
		newListInput:  nli,
		favoritePath:  favoritePath,
		player:        player.NewMPVPlayer(),
		searchHistory: history,
		width:         80,
		height:        24,
		helpModel:     components.NewHelpModel(components.CreatePlayingHelp()),
	}

	// Build menu with history items
	m.rebuildMenuWithHistory()

	return m
}

// Init initializes the lucky screen
func (m LuckyModel) Init() tea.Cmd {
	return tea.Batch(textinput.Blink, ticksEverySecond())
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
		return m, m.searchAndPickRandom(keyword)
	}

	// Clear buffer on navigation keys
	if key == "up" || key == "down" || key == "j" || key == "k" {
		m.numberBuffer = ""
	}

	// Handle menu navigation for history items
	newList, selected := components.HandleMenuKey(msg, m.menuList)
	m.menuList = newList

	if selected >= 0 {
		// Selected a history item from menu
		if m.searchHistory != nil && selected >= 0 && selected < len(m.searchHistory.LuckyQueries) {
			query := m.searchHistory.LuckyQueries[selected]
			m.state = luckyStateSearching
			m.err = nil
			return m, m.searchAndPickRandom(query)
		}
	}

	// Update text input
	var cmd tea.Cmd
	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

// selectHistoryByNumber selects a history item by number (1-based)
func (m LuckyModel) selectHistoryByNumber(num int) (tea.Model, tea.Cmd) {
	if m.searchHistory == nil {
		return m, nil
	}

	actualIndex := num - 1 // 1 = index 0, 2 = index 1, etc.
	if actualIndex >= 0 && actualIndex < len(m.searchHistory.LuckyQueries) {
		// Update menu selection
		m.menuList.Select(actualIndex)
		query := m.searchHistory.LuckyQueries[actualIndex]
		m.state = luckyStateSearching
		m.err = nil
		return m, m.searchAndPickRandom(query)
	}
	return m, nil
}

// updatePlaying handles input during playback
func (m LuckyModel) updatePlaying(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		// Stop playback and return to I Feel Lucky input
		if err := m.player.Stop(); err != nil {
			m.saveMessage = fmt.Sprintf("✗ Failed to stop playback: %v", err)
			m.saveMessageTime = 150
			return m, nil
		}
		m.state = luckyStateInput
		m.selectedStation = nil
		return m, nil
	case "0":
		// Return to main menu (Level 2+ shortcut)
		if err := m.player.Stop(); err != nil {
			m.saveMessage = fmt.Sprintf("✗ Failed to stop playback: %v", err)
			m.saveMessageTime = 150
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
	case "/":
		// Decrease volume
		newVol := m.player.DecreaseVolume(5)
		if m.selectedStation != nil && newVol >= 0 {
			m.selectedStation.SetVolume(newVol)
			m.saveStationVolume(m.selectedStation)
		}
		m.saveMessage = fmt.Sprintf("Volume: %d%%", newVol)
		m.saveMessageTime = 120 // Show for 2 seconds (60 ticks/sec)
		return m, nil
	case "*":
		// Increase volume
		newVol := m.player.IncreaseVolume(5)
		if m.selectedStation != nil {
			m.selectedStation.SetVolume(newVol)
			m.saveStationVolume(m.selectedStation)
		}
		m.saveMessage = fmt.Sprintf("Volume: %d%%", newVol)
		m.saveMessageTime = 120 // Show for 2 seconds (60 ticks/sec)
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
		m.saveMessageTime = 120 // Show for 2 seconds (60 ticks/sec)
		return m, nil
	case "?":
		m.helpModel.SetSize(m.width, m.height)
		m.helpModel.Toggle()
		return m, nil
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

	helpText := "↑↓/jk: Navigate • Enter: Search"
	if m.searchHistory != nil && len(m.searchHistory.LuckyQueries) > 0 {
		maxItems := len(m.searchHistory.LuckyQueries)
		if maxItems > m.searchHistory.MaxSize {
			maxItems = m.searchHistory.MaxSize
		}
		helpText += fmt.Sprintf(" • 1-%d: Quick search", maxItems)
	}
	helpText += " • Esc: Back • Ctrl+C: Quit"

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

	// Station info (same format as Now Playing in search)
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
		var msgStyle lipgloss.Style
		if strings.Contains(m.saveMessage, "✓") || strings.HasPrefix(m.saveMessage, "Volume:") {
			if strings.Contains(m.saveMessage, "Muted") {
				msgStyle = infoStyle()
			} else {
				msgStyle = successStyle()
			}
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
		Help:    "f: Save to Favorites • s: Save to list • v: Vote • ?: Help",
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
