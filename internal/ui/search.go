package ui

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/shinokada/tera/v3/internal/api"
	"github.com/shinokada/tera/v3/internal/blocklist"
	"github.com/shinokada/tera/v3/internal/config"
	"github.com/shinokada/tera/v3/internal/player"
	"github.com/shinokada/tera/v3/internal/storage"
	"github.com/shinokada/tera/v3/internal/theme"
	"github.com/shinokada/tera/v3/internal/ui/components"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// Message types for search screen (patterned after lucky.go and play.go)
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

// Message types for search results and errors
type searchErrorMsg struct {
	err error
}
type searchResultsMsg struct {
	results []api.Station
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
		sort.Slice(results, func(i, j int) bool {
			return results[i].Votes > results[j].Votes
		})
		return searchResultsMsg{results: results}
	}
}

// loadAvailableLists loads all available favorite list names from storage.
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

// searchState represents the current state in the search screen
type searchState int

const (
	searchStateMenu searchState = iota
	searchStateInput
	searchStateLoading
	searchStateResults
	searchStateStationInfo
	searchStatePlaying
	searchStateConfirmStop // new state for confirm stop prompt
	searchStateSavePrompt
	searchStateSelectList
	searchStateNewListInput
	searchStateAdvancedForm
	searchStateTagInput
	searchStateManageTags
	searchStateSleepTimer
)

// SearchModel represents the state and data for the search screen
// (fields are modeled after other play screens and previous usage)
type SearchModel struct {
	searchHistory    *storage.SearchHistoryStore
	quickFavorites   []api.Station
	availableLists   []string
	listModel        list.Model
	player           *player.MPVPlayer
	playOptsCfg      config.PlayOptionsConfig
	votedStations    *storage.VotedStations
	ratingsManager   *storage.RatingsManager
	starRenderer     *components.StarRenderer
	tagsManager      *storage.TagsManager
	tagRenderer      *components.TagRenderer
	tagInput         components.TagInput
	manageTags       components.ManageTags
	metadataManager  *storage.MetadataManager
	blocklistManager *blocklist.Manager
	dataPath         string
	favoritePath     string
	saveMessage      string
	saveMessageTime  int
	numberBuffer     string
	err              error
	lastBlockTime    time.Time
	ratingMode       bool
	// ...existing code...
	sleepTimerActive    bool
	sleepTimerDialog    components.SleepTimerDialog
	sleepCountdown      string
	advancedInputs      []textinput.Model
	advancedFocusIdx    int
	advancedBitrate     string
	advancedSortByVotes bool
	// struct fields continue here
	apiClient       *api.Client
	state           searchState
	width           int
	height          int
	menuList        list.Model
	resultsList     list.Model
	stationInfoMenu list.Model
	helpModel       components.HelpModel
	textInput       textinput.Model
	newListInput    textinput.Model
	selectedStation *api.Station
	results         []api.Station
	searchType      api.SearchType
	// ...existing code...
	// Add missing fields for search screen
	resultsItems        []list.Item
	showBlockedInSearch bool
	spinner             spinner.Model
	// ...existing code...
	nowPlayingBar     string // set by App when ContinueOnNavigate is active
	confirmStopTarget string // "back" or "main" — set when entering confirmStop state
}

// executeSearchType transitions to the appropriate search state for the given menu index (0-based).
func (m SearchModel) executeSearchType(idx int) (tea.Model, tea.Cmd) {
	switch idx {
	case 0:
		m.searchType = api.SearchByTag
		m.state = searchStateInput
		m.textInput.SetValue("")
		m.textInput.Focus()
		return m, textinput.Blink
	case 1:
		m.searchType = api.SearchByName
		m.state = searchStateInput
		m.textInput.SetValue("")
		m.textInput.Focus()
		return m, textinput.Blink
	case 2:
		m.searchType = api.SearchByLanguage
		m.state = searchStateInput
		m.textInput.SetValue("")
		m.textInput.Focus()
		return m, textinput.Blink
	case 3:
		m.searchType = api.SearchByCountry
		m.state = searchStateInput
		m.textInput.SetValue("")
		m.textInput.Focus()
		return m, textinput.Blink
	case 4:
		m.searchType = api.SearchByState
		m.state = searchStateInput
		m.textInput.SetValue("")
		m.textInput.Focus()
		return m, textinput.Blink
	case 5:
		m.searchType = api.SearchAdvanced
		m.state = searchStateAdvancedForm
		for i := range m.advancedInputs {
			m.advancedInputs[i].SetValue("")
			m.advancedInputs[i].Blur()
		}
		m.advancedFocusIdx = 0
		m.advancedInputs[0].Focus()
		m.advancedBitrate = ""
		m.advancedSortByVotes = true
		return m, textinput.Blink
	}
	return m, nil
}

// saveToList saves the currently selected station to the named favorite list.
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

// handleStationInfoInput handles input in the station info state
func (m SearchModel) handleStationInfoInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Handle Esc to go back
	if msg.String() == "esc" {
		if m.player != nil && m.player.IsPlaying() {
			_ = m.player.Stop()
		}
		m.selectedStation = nil
		m.state = searchStateResults
		return m, nil
	}

	// Number quick-select shortcuts (1-based, matching menu items)
	switch msg.String() {
	case "1":
		return m.executeStationAction(0)
	case "2":
		return m.executeStationAction(1)
	case "3":
		return m.executeStationAction(2)
	}

	// Handle menu navigation and selection via arrow keys / enter
	newList, selected := components.HandleMenuKey(msg, m.stationInfoMenu)
	m.stationInfoMenu = newList
	if selected >= 0 {
		return m.executeStationAction(selected)
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

// playStation starts playing a station.
// Phase 5: volume is resolved from PlayOptions (DefaultVolume / StartVolumeMode).
func (m SearchModel) playStation(station api.Station) tea.Cmd {
	startVol := m.playOptsCfg.DefaultVolume
	if m.playOptsCfg.StartVolumeMode == "last_used" && m.playOptsCfg.LastUsedVolume > 0 {
		startVol = m.playOptsCfg.LastUsedVolume
	}
	if station.Volume != nil {
		startVol = *station.Volume
	}
	return tea.Batch(
		// Stop any app-level handed-off player (e.g. from Play from Favorites
		// with ContinueOnNavigate on) before starting the new stream.
		func() tea.Msg { return stopActivePlaybackMsg{} },
		func() tea.Msg {
			err := m.player.PlayWithVolume(&station, startVol)
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
			// Only report stalled if the process actually died.
			// IPC checks can fail while audio is playing (slow socket,
			// buffering, missing metadata), so trust IsPlaying() here.
			if !m.player.IsPlaying() {
				return playbackStalledMsg{station: station}
			}
			return playbackStartedMsg{}
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

// handOffPlayer hands m.player to App and replaces it with a fresh player so
// the old stream is exclusively owned by App while SearchModel has a clean
// player for the next station. Returns the updated model and handoff command.
func (m SearchModel) handOffPlayer() (SearchModel, tea.Cmd) {
	station := m.selectedStation
	oldPlayer := m.player
	newP := player.NewMPVPlayer()
	if m.metadataManager != nil {
		newP.SetMetadataManager(m.metadataManager)
	}
	m.player = newP
	cmd := func() tea.Msg {
		return handoffPlaybackMsg{
			player:       oldPlayer,
			station:      station,
			contextLabel: "Search",
		}
	}
	return m, cmd
}

// navigateToMainCmd returns the appropriate command when the user presses 0
// during playback. When ContinueOnNavigate is on it hands the player off to
// App (with a fresh replacement player) and navigates to main menu.
// NOTE: callers must use the returned model, not the original.
func (m SearchModel) navigateToMainCmd() (SearchModel, tea.Cmd) {
	navCmd := func() tea.Msg { return backToMainMsg{} }
	if m.playOptsCfg.ContinueOnNavigate && m.selectedStation != nil {
		m, handoffCmd := m.handOffPlayer()
		return m, tea.Batch(handoffCmd, navCmd)
	}
	// ContinueOnNavigate off — stop and navigate.
	if m.player != nil {
		_ = m.player.Stop()
	}
	return m, navCmd
}

// handlePlayerUpdate handles player-related updates during playback
func (m SearchModel) handlePlayerUpdate(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Handle rating mode input first
	if m.ratingMode {
		return m.handleRatingModeInput(msg)
	}

	switch msg.String() {
	case "q":
		// Quit application
		if m.player != nil {
			_ = m.player.Stop()
		}
		return m, tea.Quit
	case "0":
		// Phase 5: gate on ConfirmStop before navigating away.
		if m.playOptsCfg.ConfirmStop {
			m.confirmStopTarget = "main"
			m.state = searchStateConfirmStop
			return m, nil
		}
		// Hand off (with fresh player) or stop, then go to main menu.
		m, cmd := m.navigateToMainCmd()
		m.selectedStation = nil
		return m, cmd
	case "esc":
		// Phase 5: gate on ConfirmStop before navigating away.
		if m.playOptsCfg.ConfirmStop {
			m.confirmStopTarget = "back"
			m.state = searchStateConfirmStop
			return m, nil
		}
		// Go back to results. With ContinueOnNavigate ON, hand the player off
		// to App and install a fresh player so the next station selection
		// doesn't share the same MPVPlayer pointer.
		if m.playOptsCfg.ContinueOnNavigate && m.selectedStation != nil {
			m, handoffCmd := m.handOffPlayer()
			m.selectedStation = nil
			m.state = searchStateResults
			return m, handoffCmd
		}
		// ContinueOnNavigate off — stop the player.
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
		m.saveMessageTime = messageDisplayShort
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
		m.saveMessageTime = messageDisplayShort
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
		m.saveMessageTime = messageDisplayShort
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
					m.saveMessageTime = messageDisplayShort
					if startTick {
						return m, tickEverySecond()
					}
				}
			}
		}
		return m, nil
	case "r":
		// Enter rating mode
		if m.selectedStation != nil && m.ratingsManager != nil {
			m.ratingMode = true
			m.saveMessage = "Press 1-5 to rate, 0 to remove rating, Esc to cancel"
			m.saveMessageTime = -1 // Persistent until action
			return m, nil
		}
		return m, nil
	case "t":
		// Enter tag input mode
		if m.selectedStation != nil && m.tagsManager != nil {
			allTags := m.tagsManager.GetAllTags()
			w := m.width
			if w < 24 {
				w = 24
			}
			m.tagInput = components.NewTagInput(allTags, w)
			m.state = searchStateTagInput
			return m, nil
		}
		return m, nil
	case "T":
		// Enter manage tags dialog
		if m.selectedStation != nil && m.tagsManager != nil {
			currentTags := m.tagsManager.GetTags(m.selectedStation.StationUUID)
			allTags := m.tagsManager.GetAllTags()
			w := m.width
			if w < 24 {
				w = 24
			}
			m.manageTags = components.NewManageTags(m.selectedStation.TrimName(), currentTags, allTags, w)
			m.state = searchStateManageTags
			return m, nil
		}
		return m, nil
	case "Z":
		// If a sleep timer is already running, Z cancels it immediately.
		// Otherwise, open the dialog to set a new duration.
		if m.sleepTimerActive {
			return m, func() tea.Msg { return sleepTimerCancelMsg{} }
		}
		last := 30
		if m.dataPath != "" {
			if cfg, err := storage.LoadSleepTimerConfig(m.dataPath); err == nil && cfg.LastDurationMinutes > 0 {
				last = cfg.LastDurationMinutes
			}
		}
		w := m.width
		if w < 24 {
			w = 24
		}
		m.sleepTimerDialog = components.NewSleepTimerDialog(last, w)
		m.state = searchStateSleepTimer
		return m, nil
	case "+":
		// Extend active sleep timer by 15 minutes (no-op when timer is not running)
		if m.sleepTimerActive {
			return m, func() tea.Msg { return sleepTimerExtendMsg{Minutes: 15} }
		}
		return m, nil
	}
	return m, nil
}

// handleSleepTimerDialogKey delegates key events to the SleepTimerDialog component.
func (m SearchModel) handleSleepTimerDialogKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.sleepTimerDialog, cmd = m.sleepTimerDialog.Update(msg)
	return m, cmd
}

// handleTagInputKey delegates key events to the TagInput component.
func (m SearchModel) handleTagInputKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.tagInput, cmd = m.tagInput.Update(msg)
	return m, cmd
}

// handleManageTagsKey delegates key events to the ManageTags component.
func (m SearchModel) handleManageTagsKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.manageTags, cmd = m.manageTags.Update(msg)
	return m, cmd
}

// handleRatingModeInput handles input when in rating mode
func (m SearchModel) handleRatingModeInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
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
		} else {
			m.saveMessage = fmt.Sprintf("✗ Rating failed: %v", err)
		}
		m.saveMessageTime = messageDisplayShort
		return m, nil
	}

	// Handle remove rating (0 only); r is treated as cancel to match play.go
	if key == "0" {
		if err := m.ratingsManager.RemoveRating(m.selectedStation.StationUUID); err != nil {
			m.saveMessage = fmt.Sprintf("✗ Remove failed: %v", err)
		} else {
			m.saveMessage = "✓ Rating removed"
		}
		m.saveMessageTime = messageDisplayShort
		return m, nil
	}

	// Any other key (including r, esc) - cancel rating mode, clear the message
	m.saveMessage = ""
	m.saveMessageTime = 0
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
		return m.renderPageWithBottomHelp(PageLayout{
			Title:   "🔍 Search Radio Stations",
			Content: content.String(),
			Help:    "Enter: Search • Esc: Back • Ctrl+C: Quit",
		}, m.height)

	case searchStateLoading:
		var content strings.Builder
		content.WriteString(m.spinner.View())
		content.WriteString(" Searching for stations...")
		return m.renderPage(PageLayout{
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
					fmt.Fprintf(&criteria, "  Tag: %s\n", tag)
				}
				if lang := strings.TrimSpace(m.advancedInputs[1].Value()); lang != "" {
					fmt.Fprintf(&criteria, "  Language: %s\n", lang)
				}
				if country := strings.TrimSpace(m.advancedInputs[2].Value()); country != "" {
					fmt.Fprintf(&criteria, "  Country: %s\n", country)
				}
				if state := strings.TrimSpace(m.advancedInputs[3].Value()); state != "" {
					fmt.Fprintf(&criteria, "  State: %s\n", state)
				}
				if name := strings.TrimSpace(m.advancedInputs[4].Value()); name != "" {
					fmt.Fprintf(&criteria, "  Name: %s\n", name)
				}
				if m.advancedBitrate != "" {
					bitrateText := map[string]string{
						"1": "Low (≤ 64 kbps)",
						"2": "Medium (96-128 kbps)",
						"3": "High (≥ 192 kbps)",
					}
					fmt.Fprintf(&criteria, "  Bitrate: %s\n", bitrateText[m.advancedBitrate])
				}
			}

			return m.renderPage(PageLayout{
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

		return m.renderPageWithBottomHelp(PageLayout{
			Content: content.String(),
			Help:    "↑↓/jk: Navigate • Enter: Play • Esc: Back • 0: Main Menu • Ctrl+C: Quit",
		}, m.height)

	case searchStateStationInfo:
		return m.renderStationInfo()

	case searchStateSavePrompt:
		return m.renderSavePrompt()

	case searchStatePlaying:
		var content strings.Builder
		if m.selectedStation != nil {
			// Check if user has voted for this station
			hasVoted := m.votedStations != nil && m.votedStations.HasVoted(m.selectedStation.StationUUID)
			// Get metadata for display
			var metadata *storage.StationMetadata
			if m.metadataManager != nil {
				metadata = m.metadataManager.GetMetadata(m.selectedStation.StationUUID)
			}
			// Get rating for display
			var rating int
			if m.ratingsManager != nil {
				if r := m.ratingsManager.GetRating(m.selectedStation.StationUUID); r != nil {
					rating = r.Rating
				}
			}
			content.WriteString(RenderStationDetailsWithRating(*m.selectedStation, hasVoted, metadata, rating, m.starRenderer))
			// Playback status with proper spacing
			content.WriteString("\n")
			if m.player.IsPlaying() {
				// Use the cached track (kept fresh by monitorMetadata every 5 s) to
				// avoid a blocking IPC socket call inside the render path.
				if track := m.player.GetCachedTrack(); IsValidTrackMetadata(track, m.selectedStation.TrimName()) {
					content.WriteString(successStyle().Render("▶ Now Playing:"))
					content.WriteString(" ")
					content.WriteString(infoStyle().Render(track))
				} else {
					content.WriteString(successStyle().Render("▶ Playing..."))
				}
			} else {
				content.WriteString(infoStyle().Render("⏸ Stopped"))
			}
			// Tag display
			if m.tagsManager != nil && m.tagRenderer != nil {
				tags := m.tagsManager.GetTags(m.selectedStation.StationUUID)
				content.WriteString("\n")
				if len(tags) > 0 {
					fmt.Fprintf(&content, "Tags: %s", m.tagRenderer.RenderList(tags))
				} else {
					content.WriteString(helpStyle().Render("No tags — press t to add one"))
				}
			}
		}
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
		// Sleep timer countdown
		if timerInfo := m.sleepTimerCountdown(); timerInfo != "" {
			content.WriteString("\n")
			content.WriteString(highlightStyle().Render(timerInfo))
		}
		helpText := "Space: Pause • f: Fav • s: List • v: Vote • b: Block • Z: Sleep • +: Extend • 0: Main Menu • ?: Help"
		return m.renderPageWithBottomHelp(PageLayout{
			Title:   "🎵 Now Playing",
			Content: content.String(),
			Help:    helpText,
		}, m.height)

	case searchStateSelectList:
		return m.viewSelectList()

	case searchStateNewListInput:
		return m.viewNewListInput()

	case searchStateAdvancedForm:
		return m.viewAdvancedForm()

	case searchStateTagInput:
		return m.viewTagInput()
	case searchStateManageTags:
		var sb strings.Builder
		if m.selectedStation != nil {
			sb.WriteString(boldStyle().Render(m.selectedStation.TrimName()))
			sb.WriteString("\n\n")
		}
		sb.WriteString(m.manageTags.View())
		return m.renderPageWithBottomHelp(PageLayout{
			Title:   "🏷 Manage Tags",
			Content: sb.String(),
			Help:    "Space/Enter: Toggle • ↑↓/jk: Navigate • d: Done • Esc: Cancel",
		}, m.height)
	case searchStateSleepTimer:
		return m.renderPageWithBottomHelp(PageLayout{
			Title:   "💤 Sleep Timer",
			Content: m.sleepTimerDialog.View(),
			Help:    "Enter: Set • ↑↓/jk: Navigate • Esc: Cancel",
		}, m.height)

	case searchStateConfirmStop:
		return m.renderPageWithBottomHelp(PageLayout{
			Title:   "⚠ Confirm Stop",
			Content: "Stop playback and leave this screen?\n\ny/1: Yes, stop\nn/2/Esc: No, keep playing",
			Help:    "y/1: Yes • n/2/Esc: No",
		}, m.height)
	}

	return m.renderPage(PageLayout{
		Content: "Unknown state",
		Help:    "",
	})
}

// viewTagInput renders the tag input overlay.
func (m SearchModel) viewTagInput() string {
	var content strings.Builder
	if m.selectedStation != nil {
		content.WriteString(boldStyle().Render(m.selectedStation.TrimName()))
		content.WriteString("\n\n")
	}
	content.WriteString(m.tagInput.View())
	return m.renderPageWithBottomHelp(PageLayout{
		Title:   "🏷 Add Tag",
		Content: content.String(),
		Help:    "Enter: Add • Tab: Complete • ↑↓: Navigate • Esc: Cancel",
	}, m.height)
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
		// Get metadata for display
		var metadata *storage.StationMetadata
		if m.metadataManager != nil {
			metadata = m.metadataManager.GetMetadata(m.selectedStation.StationUUID)
		}
		hasVoted := m.votedStations != nil && m.votedStations.HasVoted(m.selectedStation.StationUUID)
		// Station info view does not handle interactive rating; use metadata-only rendering.
		content.WriteString(RenderStationDetailsWithMetadata(*m.selectedStation, hasVoted, metadata))
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

	return m.renderPage(PageLayout{
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

	return m.renderPage(PageLayout{
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

	return m.renderPage(PageLayout{
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

	return m.renderPage(PageLayout{
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

	return m.renderPageWithBottomHelp(PageLayout{
		Content: content.String(),
		Help:    helpText,
	}, m.height)
}

// reloadSearchHistory reloads history from disk and rebuilds the menu.
func (m *SearchModel) reloadSearchHistory() {
	store := storage.NewStorage(m.favoritePath)
	history, err := store.LoadSearchHistory(context.Background())
	if err != nil || history == nil {
		history = storage.NewSearchHistoryStore()
	}
	m.searchHistory = history
	m.rebuildMenuWithHistory()
}

// reloadQuickFavorites reloads the My-favorites station list from disk so that
// duplicate detection in saveToQuickFavorites and handlePlaybackStopped stays
// current within the same Search session.
func (m *SearchModel) reloadQuickFavorites() {
	store := storage.NewStorage(m.favoritePath)
	if list, err := store.LoadList(context.Background(), "My-favorites"); err == nil {
		m.quickFavorites = list.Stations
	}
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

	return m.renderPageWithBottomHelp(PageLayout{
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

// refreshResultsTagPills updates tag pills for a single station in the results list.
func (m *SearchModel) refreshResultsTagPills(stationUUID string) {
	if m.tagsManager == nil || m.resultsList.Items() == nil {
		return
	}
	tr := components.NewTagRenderer()
	tags := m.tagsManager.GetTags(stationUUID)
	pills := ""
	if len(tags) > 0 {
		pills = tr.RenderPills(tags)
	}
	items := m.resultsList.Items()
	for i, item := range items {
		if si, ok := item.(stationListItem); ok && si.station.StationUUID == stationUUID {
			si.tagPills = pills
			items[i] = si
			break
		}
	}
	m.resultsList.SetItems(items)
}

// sleepTimerCountdown returns a formatted countdown string when a sleep timer
// is active, or an empty string. The App refreshes sleepCountdown on every tick.
func (m SearchModel) sleepTimerCountdown() string {
	return formatSleepCountdown(m.sleepCountdown)
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

// renderPage injects the now-playing bar when the model's own player is not
// actively playing (so viewPlaying is unaffected).
func (m SearchModel) renderPage(layout PageLayout) string {
	if m.player == nil || !m.player.IsPlaying() {
		layout.NowPlaying = m.nowPlayingBar
	}
	return RenderPage(layout)
}

func (m SearchModel) renderPageWithBottomHelp(layout PageLayout, height int) string {
	if m.player == nil || !m.player.IsPlaying() {
		layout.NowPlaying = m.nowPlayingBar
	}
	return RenderPageWithBottomHelp(layout, height)
}
