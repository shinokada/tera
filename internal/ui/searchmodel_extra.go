package ui

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/shinokada/tera/v3/internal/api"
	"github.com/shinokada/tera/v3/internal/blocklist"
	"github.com/shinokada/tera/v3/internal/ui/components"
)

// NewSearchModel constructs a new SearchModel with required dependencies.
func NewSearchModel(apiClient *api.Client, favoritePath, dataPath string, blocklistManager *blocklist.Manager) SearchModel {
	// Initialize 5 advanced-search text inputs (tag, lang, country, state, name).
	advInputs := make([]textinput.Model, 5)
	placeholders := []string{"genre / style", "e.g. english", "e.g. JP or Japan", "e.g. California", "partial name"}
	for i := range advInputs {
		ti := textinput.New()
		ti.Placeholder = placeholders[i]
		advInputs[i] = ti
	}
	advInputs[0].Focus() // default focus on first field

	textIn := textinput.New()
	textIn.Placeholder = "Type to search..."

	newListIn := textinput.New()
	newListIn.Placeholder = "New list name..."

	m := SearchModel{
		apiClient:        apiClient,
		favoritePath:     favoritePath,
		dataPath:         dataPath,
		blocklistManager: blocklistManager,
		tagRenderer:      components.NewTagRenderer(),
		advancedInputs:   advInputs,
		textInput:        textIn,
		newListInput:     newListIn,
	}
	m.reloadSearchHistory()
	return m
}

// Init implements tea.Model for SearchModel.
func (m SearchModel) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model for SearchModel.
func (m SearchModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.helpModel.IsVisible() {
			var cmd tea.Cmd
			m.helpModel, cmd = m.helpModel.Update(msg)
			return m, cmd
		}
		switch m.state {
		case searchStateMenu:
			return m.handleMenuKey(msg)
		case searchStateResults:
			return m.handleResultsKey(msg)
		case searchStateInput:
			return m.handleInputKey(msg)
		case searchStatePlaying:
			return m.handlePlayerUpdate(msg)
		case searchStateConfirmStop:
			return m.handleConfirmStopKey(msg)
		case searchStateSavePrompt:
			return m.handleSavePrompt(msg)
		case searchStateStationInfo:
			return m.handleStationInfoInput(msg)
		case searchStateSelectList:
			return m.handleSelectList(msg)
		case searchStateNewListInput:
			return m.handleNewListInput(msg)
		case searchStateAdvancedForm:
			return m.handleAdvancedForm(msg)
		case searchStateTagInput:
			return m.handleTagInputKey(msg)
		case searchStateManageTags:
			return m.handleManageTagsKey(msg)
		case searchStateSleepTimer:
			return m.handleSleepTimerDialogKey(msg)
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.helpModel.SetSize(msg.Width, msg.Height)
		return m, nil

	case searchResultsMsg:
		m.results = msg.results
		m.state = searchStateResults
		m.resultsItems = make([]list.Item, len(msg.results))
		for i, station := range msg.results {
			isBlocked := false
			if m.blocklistManager != nil {
				isBlocked = m.blocklistManager.IsBlockedByAny(&station)
			}
			tagPills := ""
			if m.tagsManager != nil {
				if tags := m.tagsManager.GetTags(station.StationUUID); len(tags) > 0 {
					tagPills = m.tagRenderer.RenderPills(tags)
				}
			}
			m.resultsItems[i] = stationListItem{station: station, isBlocked: isBlocked, tagPills: tagPills}
		}
		height := availableListHeight(m.height)
		delegate := createStyledDelegate()
		m.resultsList = list.New(m.resultsItems, delegate, m.width, height)
		m.resultsList.SetShowStatusBar(true)
		m.resultsList.SetFilteringEnabled(true)
		m.resultsList.SetShowHelp(false)
		return m, nil

	case searchErrorMsg:
		m.err = msg.err
		m.state = searchStateMenu
		return m, nil

	case playbackStartedMsg:
		if m.saveMessageTime <= 0 {
			return m, tickEverySecond()
		}
		return m, nil

	case playerErrorMsg:
		m.err = msg.err
		m.state = searchStateResults
		return m, nil

	case playbackStalledMsg:
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

	case saveSuccessMsg:
		name := ""
		if msg.station != nil {
			name = msg.station.TrimName()
		}
		m.saveMessage = fmt.Sprintf("✓ Saved '%s' to Quick Favorites", name)
		m.saveMessageTime = messageDisplayShort
		// Refresh quickFavorites list
		m.reloadSearchHistory()
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
		startTick := m.saveMessageTime == 0
		m.saveMessageTime = messageDisplayShort
		if startTick {
			return m, tickEverySecond()
		}
		return m, nil

	case components.VoteFailedMsg:
		m.saveMessage = fmt.Sprintf("✗ Vote failed: %v", msg.Err)
		startTick := m.saveMessageTime == 0
		m.saveMessageTime = messageDisplayShort
		if startTick {
			return m, tickEverySecond()
		}
		return m, nil

	case stationBlockedMsg:
		m.lastBlockTime = time.Now()
		if msg.success {
			if m.player != nil {
				_ = m.player.Stop()
			}
			m.saveMessage = msg.message + " (press 'u' within 5s to undo)"
			m.saveMessageTime = messageDisplayMedium
			m.state = searchStateResults
			m.selectedStation = nil
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
			return m, tickEverySecond()
		}
		m.saveMessage = msg.message
		m.saveMessageTime = messageDisplayShort
		return m, nil

	case undoBlockSuccessMsg:
		m.saveMessage = "✓ Block undone"
		m.saveMessageTime = messageDisplayShort
		return m, nil

	case undoBlockFailedMsg:
		m.saveMessage = "No recent block to undo"
		m.saveMessageTime = messageDisplayShort
		return m, nil

	case components.TagSubmittedMsg:
		if m.state == searchStateManageTags {
			var cmd tea.Cmd
			m.manageTags, cmd = m.manageTags.HandleTagSubmitted(msg.Tag)
			return m, cmd
		}
		if m.selectedStation != nil && m.tagsManager != nil {
			if err := m.tagsManager.AddTag(m.selectedStation.StationUUID, msg.Tag); err != nil {
				m.saveMessage = fmt.Sprintf("✗ %v", err)
			} else {
				m.saveMessage = fmt.Sprintf("✓ Added tag: %s", msg.Tag)
				m.refreshResultsTagPills(m.selectedStation.StationUUID)
			}
			startTick := m.saveMessageTime == 0
			m.saveMessageTime = messageDisplayShort
			m.state = searchStatePlaying
			if startTick {
				return m, tickEverySecond()
			}
		}
		m.state = searchStatePlaying
		return m, nil

	case components.TagCancelledMsg:
		if m.state == searchStateManageTags {
			m.manageTags = m.manageTags.HandleTagCancelled()
			return m, nil
		}
		m.state = searchStatePlaying
		return m, nil

	case components.ManageTagsDoneMsg:
		if m.selectedStation != nil && m.tagsManager != nil {
			if err := m.tagsManager.SetTags(m.selectedStation.StationUUID, msg.Tags); err != nil {
				m.saveMessage = fmt.Sprintf("✗ %v", err)
			} else if len(msg.Tags) == 0 {
				m.saveMessage = "✓ All tags removed"
				m.refreshResultsTagPills(m.selectedStation.StationUUID)
			} else {
				m.saveMessage = fmt.Sprintf("✓ Tags saved (%d)", len(msg.Tags))
				m.refreshResultsTagPills(m.selectedStation.StationUUID)
			}
			startTick := m.saveMessageTime == 0
			m.saveMessageTime = messageDisplayShort
			m.state = searchStatePlaying
			if startTick {
				return m, tickEverySecond()
			}
		}
		m.state = searchStatePlaying
		return m, nil

	case components.ManageTagsCancelledMsg:
		m.state = searchStatePlaying
		return m, nil

	case components.SleepTimerSelectedMsg:
		m.state = searchStatePlaying
		m.sleepTimerActive = true
		return m, func() tea.Msg { return sleepTimerActivateMsg{Minutes: msg.Minutes} }

	case components.SleepTimerCancelledMsg:
		m.state = searchStatePlaying
		return m, nil

	case tickMsg:
		if m.saveMessageTime > 0 {
			m.saveMessageTime--
			if m.saveMessageTime == 0 {
				m.saveMessage = ""
			}
		}
		return m, tickEverySecond()
	}

	// Pass through to list models for navigation
	var cmd tea.Cmd
	switch m.state {
	case searchStateMenu:
		m.menuList, cmd = m.menuList.Update(msg)
	case searchStateResults:
		if m.resultsList.Items() != nil {
			m.resultsList, cmd = m.resultsList.Update(msg)
		}
	}
	return m, cmd
}

// handleMenuKey handles key input in the search menu state.
func (m SearchModel) handleMenuKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		return m, func() tea.Msg { return backToMainMsg{} }
	case "ctrl+c":
		return m, tea.Quit
	case "q":
		return m, tea.Quit
	case "enter":
		// If there's a buffered number, execute it
		if m.numberBuffer != "" {
			var num int
			if _, err := fmt.Sscanf(m.numberBuffer, "%d", &num); err == nil {
				m.numberBuffer = ""
				return m.selectByNumber(num)
			}
			m.numberBuffer = ""
		}
		// Otherwise use the highlighted list item.
		// Map the raw list index back to a logical action:
		//   0-5  → search types
		//   6    → blank spacer (ignore)
		//   7    → separator header (ignore)
		//   8+   → history item (index = raw - 8)
		idx := m.menuList.Index()
		if idx <= 5 {
			return m.executeSearchType(idx)
		}
		if idx >= 8 && m.searchHistory != nil {
			historyIndex := idx - 8
			if historyIndex < len(m.searchHistory.SearchItems) {
				item := m.searchHistory.SearchItems[historyIndex]
				return m.executeHistorySearch(item.SearchType, item.Query)
			}
		}
		return m, nil
	case "0":
		if m.numberBuffer != "" {
			m.numberBuffer += "0"
			var num int
			if _, err := fmt.Sscanf(m.numberBuffer, "%d", &num); err == nil {
				m.numberBuffer = ""
				return m.selectByNumber(num)
			}
			m.numberBuffer = ""
			return m, nil
		}
		return m, func() tea.Msg { return backToMainMsg{} }
	default:
		// Number buffer for quick selection.
		// Single digits 1-6 are NOT executed immediately — they are buffered so
		// that two-digit history shortcuts like "10", "11", etc. can be entered.
		if len(msg.String()) == 1 && msg.String()[0] >= '1' && msg.String()[0] <= '9' {
			m.numberBuffer += msg.String()
			// Two digits: execute immediately
			if len(m.numberBuffer) >= 2 {
				var num int
				if _, err := fmt.Sscanf(m.numberBuffer, "%d", &num); err == nil {
					m.numberBuffer = ""
					return m.selectByNumber(num)
				}
				m.numberBuffer = ""
			}
			// Single digit: buffer it and wait for Enter or a second digit
			return m, nil
		}
		// Non-digit key: clear buffer
		m.numberBuffer = ""
	}
	var cmd tea.Cmd
	m.menuList, cmd = m.menuList.Update(msg)
	return m, cmd
}

// handleInputKey handles key input in the text-input state.
func (m SearchModel) handleInputKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.state = searchStateMenu
		return m, nil
	case "enter":
		query := m.textInput.Value()
		if query == "" {
			return m, nil
		}
		m.textInput.SetValue("") // clear input after submitting
		m.state = searchStateLoading
		return m, m.performSearch(query)
	}
	var cmd tea.Cmd
	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

// handleResultsKey handles key input in the search results state.
func (m SearchModel) handleResultsKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "0":
		m.state = searchStateMenu
		return m, nil
	case "enter":
		if item, ok := m.resultsList.SelectedItem().(stationListItem); ok {
			station := item.station
			m.selectedStation = &station
			m.state = searchStatePlaying
			if m.player != nil {
				return m, m.playStation(station)
			}
			return m, nil
		}
	}
	var cmd tea.Cmd
	m.resultsList, cmd = m.resultsList.Update(msg)
	return m, cmd
}

// handleConfirmStopKey handles the confirm-stop prompt (Phase 5).
func (m SearchModel) handleConfirmStopKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y", "Y":
		// Stop and navigate back
		if m.player != nil {
			_ = m.player.Stop()
		}
		m.selectedStation = nil
		m.state = searchStateResults
		return m, nil
	case "n", "N", "esc":
		// Resume / stay
		m.state = searchStatePlaying
		return m, nil
	}
	return m, nil
}
