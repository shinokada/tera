package ui

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/shinokada/tera/v3/internal/api"
	"github.com/shinokada/tera/v3/internal/blocklist"
	"github.com/shinokada/tera/v3/internal/player"
	"github.com/shinokada/tera/v3/internal/storage"
	"github.com/shinokada/tera/v3/internal/ui/components"
)

// Sort options for Most Played view
type MostPlayedSort int

const (
	sortByPlayCount MostPlayedSort = iota
	sortByLastPlayed
	sortByFirstPlayed
	sortByName
)

func (s MostPlayedSort) String() string {
	switch s {
	case sortByPlayCount:
		return "Play Count"
	case sortByLastPlayed:
		return "Last Played"
	case sortByFirstPlayed:
		return "First Played"
	default:
		return "Play Count"
	}
}

// numSortModes is the number of valid sort modes (excludes sortByName which requires
// station details not available in metadata-only view).
const numSortModes = 3

// State for Most Played screen
type mostPlayedState int

const (
	mostPlayedStateList mostPlayedState = iota
	mostPlayedStatePlaying
	mostPlayedStateSavePrompt
	mostPlayedStateSelectList
)

// MostPlayedModel represents the Most Played screen
type MostPlayedModel struct {
	state            mostPlayedState
	sortBy           MostPlayedSort
	stations         []storage.StationWithMetadata
	stationItems     []list.Item
	stationListModel list.Model
	selectedStation  *api.Station
	player           *player.MPVPlayer
	metadataManager  *storage.MetadataManager
	favoritePath     string
	saveMessage      string
	saveMessageTime  int
	width            int
	height           int
	err              error
	helpModel        components.HelpModel
	votedStations    *storage.VotedStations
	blocklistManager *blocklist.Manager
	// For saving to list
	availableLists []string
	listItems      []list.Item
	listModel      list.Model
}

// mostPlayedStationItem wraps a station with metadata for the list
type mostPlayedStationItem struct {
	station  api.Station
	metadata *storage.StationMetadata
}

func (i mostPlayedStationItem) FilterValue() string { return i.station.Name }
func (i mostPlayedStationItem) Title() string {
	name := i.station.TrimName()
	if len(name) > 35 {
		name = name[:32] + "..."
	}
	return name
}
func (i mostPlayedStationItem) Description() string {
	var parts []string
	if i.metadata != nil {
		parts = append(parts, fmt.Sprintf("%d plays", i.metadata.PlayCount))
		if !i.metadata.LastPlayed.IsZero() {
			parts = append(parts, storage.FormatLastPlayed(i.metadata.LastPlayed))
		}
	}
	if i.station.Country != "" {
		parts = append(parts, i.station.Country)
	}
	return strings.Join(parts, " • ")
}

// NewMostPlayedModel creates a new Most Played model
func NewMostPlayedModel(metadataManager *storage.MetadataManager, favoritePath string, blocklistManager *blocklist.Manager) MostPlayedModel {
	m := MostPlayedModel{
		state:            mostPlayedStateList,
		sortBy:           sortByPlayCount,
		player:           player.NewMPVPlayer(),
		metadataManager:  metadataManager,
		favoritePath:     favoritePath,
		blocklistManager: blocklistManager,
		helpModel:        components.NewHelpModel(createMostPlayedHelp()),
	}

	// Load voted stations
	votedStations, err := storage.LoadVotedStations()
	if err == nil {
		m.votedStations = votedStations
	}

	// Initialize station list
	delegate := createStyledDelegate()
	m.stationListModel = list.New([]list.Item{}, delegate, 50, 20)
	m.stationListModel.SetShowTitle(false)
	m.stationListModel.SetShowStatusBar(false)
	m.stationListModel.SetFilteringEnabled(true)
	m.stationListModel.SetShowHelp(false)

	return m
}

// createMostPlayedHelp creates the help sections for the Most Played screen
func createMostPlayedHelp() []components.HelpSection {
	return []components.HelpSection{
		{
			Title: "Navigation",
			Items: []components.HelpItem{
				{Key: "↑↓/jk", Description: "Navigate"},
				{Key: "Enter", Description: "Play"},
				{Key: "s", Description: "Sort"},
				{Key: "f", Description: "Add to favorites"},
				{Key: "?", Description: "Help"},
				{Key: "Esc/m", Description: "Back"},
			},
		},
	}
}

func (m MostPlayedModel) Init() tea.Cmd {
	return m.loadStations
}

// loadStations loads stations with metadata
func (m MostPlayedModel) loadStations() tea.Msg {
	return mostPlayedLoadedMsg{}
}

type mostPlayedLoadedMsg struct{}

func (m MostPlayedModel) Update(msg tea.Msg) (MostPlayedModel, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case mostPlayedLoadedMsg:
		m.refreshStationList()
		return m, nil

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.helpModel.SetSize(msg.Width, msg.Height)
		h, v := docStyle().GetFrameSize()
		m.stationListModel.SetSize(msg.Width-h, msg.Height-v-6)
		return m, nil

	case tea.KeyMsg:
		// Handle help overlay first
		if m.helpModel.IsVisible() {
			m.helpModel, cmd = m.helpModel.Update(msg)
			return m, cmd
		}

		switch m.state {
		case mostPlayedStateList:
			return m.handleListInput(msg)
		case mostPlayedStatePlaying:
			return m.handlePlayingInput(msg)
		case mostPlayedStateSavePrompt:
			return m.handleSavePromptInput(msg)
		case mostPlayedStateSelectList:
			return m.handleSelectListInput(msg)
		}

	case tickMsg:
		// Decrement save message timer
		if m.saveMessageTime > 0 {
			m.saveMessageTime--
			if m.saveMessageTime == 0 {
				m.saveMessage = ""
			}
			return m, tickEverySecond()
		}
	}

	// Update list model
	m.stationListModel, cmd = m.stationListModel.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m *MostPlayedModel) refreshStationList() {
	if m.metadataManager == nil {
		m.stations = []storage.StationWithMetadata{}
		return
	}

	// Get stations based on sort order
	switch m.sortBy {
	case sortByPlayCount:
		m.stations = m.metadataManager.GetTopPlayed(100)
	case sortByLastPlayed:
		m.stations = m.metadataManager.GetRecentlyPlayed(100)
	case sortByFirstPlayed:
		m.stations = m.metadataManager.GetFirstPlayed(100)
	default:
		m.stations = m.metadataManager.GetTopPlayed(100)
	}

	// We need to look up full station info from favorites or API
	// For now, we'll just display what we have from metadata
	// This is a limitation - we only have UUIDs in metadata
	// Future enhancement: cache station info in metadata

	// Convert to list items
	m.stationItems = make([]list.Item, len(m.stations))
	for i, s := range m.stations {
		m.stationItems[i] = mostPlayedStationItem{
			station:  s.Station,
			metadata: s.Metadata,
		}
	}
	m.stationListModel.SetItems(m.stationItems)
}

func (m MostPlayedModel) handleListInput(msg tea.KeyMsg) (MostPlayedModel, tea.Cmd) {
	switch msg.String() {
	case "q", "esc", "m":
		// Stop player before returning
		if m.player != nil && m.player.IsPlaying() {
			_ = m.player.Stop()
		}
		return m, func() tea.Msg { return navigateMsg{screen: screenMainMenu} }

	case "?":
		m.helpModel.Show()
		return m, nil

	case "enter":
		// Play selected station
		if len(m.stationItems) > 0 {
			selected := m.stationListModel.SelectedItem()
			if item, ok := selected.(mostPlayedStationItem); ok {
				m.selectedStation = &item.station
				// Check if we have the URL to play
				if item.station.URLResolved != "" {
					if err := m.player.Play(&item.station); err != nil {
						m.err = err
					} else {
						m.state = mostPlayedStatePlaying
					}
				} else {
					m.saveMessage = "Station URL not available (needs lookup)"
					m.saveMessageTime = 3
					return m, tickEverySecond()
				}
			}
		}
		return m, nil

	case "s":
		// Cycle through implemented sort options (Name excluded: station names not stored in metadata)
		m.sortBy = (m.sortBy + 1) % numSortModes
		m.refreshStationList()
		m.saveMessage = fmt.Sprintf("Sorted by: %s", m.sortBy.String())
		m.saveMessageTime = 2
		return m, tickEverySecond()

	case "f":
		// Add to favorites
		if len(m.stationItems) > 0 {
			selected := m.stationListModel.SelectedItem()
			if item, ok := selected.(mostPlayedStationItem); ok {
				m.selectedStation = &item.station
				m.state = mostPlayedStateSavePrompt
			}
		}
		return m, nil
	}

	// Pass to list model for navigation
	var cmd tea.Cmd
	m.stationListModel, cmd = m.stationListModel.Update(msg)
	return m, cmd
}

func (m MostPlayedModel) handlePlayingInput(msg tea.KeyMsg) (MostPlayedModel, tea.Cmd) {
	switch msg.String() {
	case "q", "esc", "m":
		// Stop playback and return to list
		if m.player != nil {
			_ = m.player.Stop()
		}
		m.state = mostPlayedStateList
		return m, nil

	case "s":
		// Stop playback
		if m.player != nil {
			_ = m.player.Stop()
		}
		return m, nil

	case "p":
		// Toggle pause
		if m.player != nil {
			_ = m.player.TogglePause()
		}
		return m, nil

	case "f":
		// Save to favorites
		m.state = mostPlayedStateSavePrompt
		return m, nil

	case "+", "=":
		// Volume up (IncreaseVolume handles clamping to 0-100)
		if m.player != nil {
			vol := m.player.IncreaseVolume(5)
			m.saveMessage = fmt.Sprintf("Volume: %d%%", vol)
			m.saveMessageTime = 2
			return m, tickEverySecond()
		}

	case "-":
		// Volume down (DecreaseVolume handles clamping to 0-100)
		if m.player != nil {
			vol := m.player.DecreaseVolume(5)
			m.saveMessage = fmt.Sprintf("Volume: %d%%", vol)
			m.saveMessageTime = 2
			return m, tickEverySecond()
		}
	}

	return m, nil
}

func (m MostPlayedModel) handleSavePromptInput(msg tea.KeyMsg) (MostPlayedModel, tea.Cmd) {
	switch msg.String() {
	case "y", "Y":
		// Save to My-favorites
		if m.selectedStation != nil {
			store := storage.NewStorage(m.favoritePath)
			err := store.AddStation(context.TODO(), "My-favorites", *m.selectedStation)
			if err != nil {
				if err == storage.ErrDuplicateStation {
					m.saveMessage = "Already in My-favorites"
				} else {
					m.saveMessage = fmt.Sprintf("Error: %v", err)
				}
			} else {
				m.saveMessage = "✓ Saved to My-favorites"
			}
			m.saveMessageTime = 3
		}
		m.state = mostPlayedStateList
		return m, tickEverySecond()

	case "n", "N", "esc":
		m.state = mostPlayedStateList
		return m, nil

	case "l", "L":
		// Select from list
		m.loadAvailableLists()
		m.state = mostPlayedStateSelectList
		return m, nil
	}

	return m, nil
}

func (m MostPlayedModel) handleSelectListInput(msg tea.KeyMsg) (MostPlayedModel, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.state = mostPlayedStateList
		return m, nil

	case "enter":
		// Save to selected list
		if len(m.listItems) > 0 && m.selectedStation != nil {
			selected := m.listModel.SelectedItem()
			if item, ok := selected.(components.MenuItem); ok {
				listName := item.Title()
				store := storage.NewStorage(m.favoritePath)
				err := store.AddStation(context.TODO(), listName, *m.selectedStation)
				if err != nil {
					if err == storage.ErrDuplicateStation {
						m.saveMessage = fmt.Sprintf("Already in %s", listName)
					} else {
						m.saveMessage = fmt.Sprintf("Error: %v", err)
					}
				} else {
					m.saveMessage = fmt.Sprintf("✓ Saved to %s", listName)
				}
				m.saveMessageTime = 3
			}
		}
		m.state = mostPlayedStateList
		return m, tickEverySecond()
	}

	// Pass to list model for navigation
	var cmd tea.Cmd
	m.listModel, cmd = m.listModel.Update(msg)
	return m, cmd
}

func (m *MostPlayedModel) loadAvailableLists() {
	store := storage.NewStorage(m.favoritePath)
	lists, _ := store.GetAllLists(context.TODO())
	m.availableLists = lists

	items := make([]list.Item, len(lists))
	for i, name := range lists {
		items[i] = components.NewMenuItem(name, "", fmt.Sprintf("%d", i+1))
	}
	m.listItems = items

	m.listModel = list.New(items, createStyledDelegate(), 40, 10)
	m.listModel.SetShowTitle(false)
	m.listModel.SetShowStatusBar(false)
	m.listModel.SetFilteringEnabled(false)
	m.listModel.SetShowHelp(false)
}

func (m MostPlayedModel) View() string {
	// Help overlay
	if m.helpModel.IsVisible() {
		return m.helpModel.View()
	}

	switch m.state {
	case mostPlayedStateList:
		return m.viewList()
	case mostPlayedStatePlaying:
		return m.viewPlaying()
	case mostPlayedStateSavePrompt:
		return m.viewSavePrompt()
	case mostPlayedStateSelectList:
		return m.viewSelectList()
	}

	return m.viewList()
}

func (m MostPlayedModel) viewList() string {
	var content strings.Builder

	// Sort indicator
	sortInfo := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Render(fmt.Sprintf("Sort by: %s (press 's' to change)", m.sortBy.String()))
	content.WriteString(sortInfo)
	content.WriteString("\n\n")

	// Check if we have any stations
	if len(m.stationItems) == 0 {
		content.WriteString(infoStyle().Render("ℹ No play history yet - start listening!"))
		content.WriteString("\n\n")
		content.WriteString("Play some stations to see your listening statistics here.")
	} else {
		content.WriteString(m.stationListModel.View())
	}

	// Show save message if any
	if m.saveMessage != "" {
		content.WriteString("\n\n")
		if strings.Contains(m.saveMessage, "✓") || strings.HasPrefix(m.saveMessage, "Sorted") || strings.HasPrefix(m.saveMessage, "Volume") {
			content.WriteString(successStyle().Render(m.saveMessage))
		} else {
			content.WriteString(infoStyle().Render(m.saveMessage))
		}
	}

	return RenderPage(PageLayout{
		Title:   "📊 Most Played Stations",
		Content: content.String(),
		Help:    "↑↓: Navigate • Enter: Play • s: Sort • f: Favorites • Esc: Back",
	})
}

func (m MostPlayedModel) viewPlaying() string {
	if m.selectedStation == nil {
		return "No station selected"
	}

	var content strings.Builder

	// Station info with metadata
	hasVoted := m.votedStations != nil && m.votedStations.HasVoted(m.selectedStation.StationUUID)
	var metadata *storage.StationMetadata
	if m.metadataManager != nil {
		metadata = m.metadataManager.GetMetadata(m.selectedStation.StationUUID)
	}
	content.WriteString(RenderStationDetailsWithMetadata(*m.selectedStation, hasVoted, metadata))

	// Playback status
	content.WriteString("\n")
	if m.player.IsPlaying() {
		if m.player.IsPaused() {
			content.WriteString(infoStyle().Render("⏸ Paused"))
		} else {
			// Use cached track to avoid IPC call in the render path
			track := m.player.GetCachedTrack()
			if track != "" && track != m.selectedStation.Name {
				content.WriteString(successStyle().Render("▶ Now Playing:"))
				content.WriteString(" ")
				content.WriteString(infoStyle().Render(track))
			} else {
				content.WriteString(successStyle().Render("▶ Playing..."))
			}
		}
	} else {
		content.WriteString(infoStyle().Render("⏹ Stopped"))
	}

	// Volume
	content.WriteString(fmt.Sprintf("\n\nVolume: %d%%", m.player.GetVolume()))

	// Show save message
	if m.saveMessage != "" {
		content.WriteString("\n\n")
		content.WriteString(successStyle().Render(m.saveMessage))
	}

	return RenderPage(PageLayout{
		Title:   "📊 Most Played - Now Playing",
		Content: content.String(),
		Help:    "p: Pause • s: Stop • +/-: Volume • f: Favorites • Esc: Back",
	})
}

func (m MostPlayedModel) viewSavePrompt() string {
	var content strings.Builder

	if m.selectedStation != nil {
		content.WriteString(fmt.Sprintf("Save \"%s\" to favorites?\n\n", m.selectedStation.TrimName()))
	}
	content.WriteString("[Y] Save to My-favorites\n")
	content.WriteString("[L] Choose from list\n")
	content.WriteString("[N] Cancel")

	return RenderPage(PageLayout{
		Title:   "💾 Save Station",
		Content: content.String(),
		Help:    "Y: My-favorites • L: Choose list • N: Cancel",
	})
}

func (m MostPlayedModel) viewSelectList() string {
	var content strings.Builder

	content.WriteString("Select a list:\n\n")
	content.WriteString(m.listModel.View())

	return RenderPage(PageLayout{
		Title:   "📁 Select List",
		Content: content.String(),
		Help:    "↑↓: Navigate • Enter: Select • Esc: Cancel",
	})
}
