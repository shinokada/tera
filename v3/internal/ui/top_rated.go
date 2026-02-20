package ui

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/shinokada/tera/v3/internal/api"
	"github.com/shinokada/tera/v3/internal/blocklist"
	"github.com/shinokada/tera/v3/internal/player"
	"github.com/shinokada/tera/v3/internal/storage"
	"github.com/shinokada/tera/v3/internal/ui/components"
)

// Sort options for Top Rated view
type TopRatedSort int

const (
	sortByRatingHigh TopRatedSort = iota
	sortByRatingLow
	sortByRecentlyRated
	sortByNameRated
)

func (s TopRatedSort) String() string {
	switch s {
	case sortByRatingHigh:
		return "Rating (High→Low)"
	case sortByRatingLow:
		return "Rating (Low→High)"
	case sortByRecentlyRated:
		return "Recently Rated"
	case sortByNameRated:
		return "Station Name"
	default:
		return "Rating (High→Low)"
	}
}

const numTopRatedSortModes = 3 // Excludes sortByNameRated

// Filter options for Top Rated view
type TopRatedFilter int

const (
	filterAllRatings TopRatedFilter = iota
	filterFiveStars
	filterFourPlus
	filterThreePlus
)

func (f TopRatedFilter) String() string {
	switch f {
	case filterAllRatings:
		return "All Ratings"
	case filterFiveStars:
		return "5 Stars Only"
	case filterFourPlus:
		return "4+ Stars"
	case filterThreePlus:
		return "3+ Stars"
	default:
		return "All Ratings"
	}
}

func (f TopRatedFilter) MinRating() int {
	switch f {
	case filterFiveStars:
		return 5
	case filterFourPlus:
		return 4
	case filterThreePlus:
		return 3
	default:
		return 1
	}
}

const numTopRatedFilterModes = 4

// State for Top Rated screen
type topRatedState int

const (
	topRatedStateList topRatedState = iota
	topRatedStatePlaying
	topRatedStateSavePrompt
	topRatedStateSelectList
	topRatedStateRating // Rating mode activated with *
)

// TopRatedModel represents the Top Rated screen
type TopRatedModel struct {
	state              topRatedState
	sortBy             TopRatedSort
	filterBy           TopRatedFilter
	stations           []storage.StationWithRating
	stationItems       []list.Item
	stationListModel   list.Model
	selectedStation    *api.Station
	player             *player.MPVPlayer
	ratingsManager     *storage.RatingsManager
	metadataManager    *storage.MetadataManager
	starRenderer       *components.StarRenderer
	tagsManager        *storage.TagsManager   // for tag pill display
	tagRenderer        *components.TagRenderer // for rendering tag pills
	favoritePath       string
	saveMessage        string
	saveMessageSuccess bool
	saveMessageTime    int
	width              int
	height             int
	err                error
	helpModel          components.HelpModel
	votedStations      *storage.VotedStations
	blocklistManager   *blocklist.Manager
	// For saving to list
	availableLists []string
	listItems      []list.Item
	listModel      list.Model
}

// topRatedStationItem wraps a station with rating for the list
type topRatedStationItem struct {
	station  api.Station
	rating   *storage.StationRating
	tagPills string // pre-rendered tag pills (empty if no tags)
}

func (i topRatedStationItem) FilterValue() string {
	if i.station.Name != "" {
		return i.station.Name
	}
	return i.station.StationUUID
}
func (i topRatedStationItem) Title() string {
	name := i.station.TrimName()
	if name == "" {
		// Fallback for old ratings without cached station info
		if len(i.station.StationUUID) >= 8 {
			name = "Station " + i.station.StationUUID[:8]
		} else {
			name = "Unknown Station"
		}
	}
	if len(name) > 35 {
		name = name[:32] + "..."
	}
	if i.tagPills != "" {
		return name + "  " + i.tagPills
	}
	return name
}
func (i topRatedStationItem) Description() string {
	var parts []string
	if i.rating != nil {
		// Show stars
		stars := storage.RenderStarsCompact(i.rating.Rating, true)
		if stars != "" {
			parts = append(parts, stars)
		}
		if !i.rating.UpdatedAt.IsZero() {
			parts = append(parts, storage.FormatRatedAt(i.rating.UpdatedAt))
		}
	}
	if i.station.Country != "" {
		parts = append(parts, i.station.Country)
	}
	return strings.Join(parts, " • ")
}

// NewTopRatedModel creates a new Top Rated model
func NewTopRatedModel(ratingsManager *storage.RatingsManager, metadataManager *storage.MetadataManager, starRenderer *components.StarRenderer, favoritePath string, blocklistManager *blocklist.Manager) TopRatedModel {
	m := TopRatedModel{
		state:            topRatedStateList,
		sortBy:           sortByRatingHigh,
		filterBy:         filterAllRatings,
		player:           player.NewMPVPlayer(),
		ratingsManager:   ratingsManager,
		metadataManager:  metadataManager,
		starRenderer:     starRenderer,
		favoritePath:     favoritePath,
		blocklistManager: blocklistManager,
		helpModel:        components.NewHelpModel(createTopRatedHelp()),
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

// createTopRatedHelp creates the help sections for the Top Rated screen
func createTopRatedHelp() []components.HelpSection {
	return []components.HelpSection{
		{
			Title: "Navigation",
			Items: []components.HelpItem{
				{Key: "↑↓/jk", Description: "Navigate"},
				{Key: "Enter", Description: "Play"},
				{Key: "r then 1-5", Description: "Rate station"},
				{Key: "r then 0", Description: "Remove rating"},
				{Key: "s", Description: "Sort"},
				{Key: "f", Description: "Filter"},
				{Key: "a", Description: "Add to favorites"},
				{Key: "?", Description: "Help"},
				{Key: "Esc/m", Description: "Back"},
			},
		},
	}
}

func (m TopRatedModel) Init() tea.Cmd {
	return m.loadStations
}

// loadStations loads stations with ratings
func (m TopRatedModel) loadStations() tea.Msg {
	return topRatedLoadedMsg{}
}

type topRatedLoadedMsg struct{}

func (m TopRatedModel) Update(msg tea.Msg) (TopRatedModel, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case topRatedLoadedMsg:
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
		case topRatedStateList:
			return m.handleListInput(msg)
		case topRatedStatePlaying:
			return m.handlePlayingInput(msg)
		case topRatedStateSavePrompt:
			return m.handleSavePromptInput(msg)
		case topRatedStateSelectList:
			return m.handleSelectListInput(msg)
		case topRatedStateRating:
			return m.handleRatingInput(msg)
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

func (m *TopRatedModel) refreshStationList() {
	if m.ratingsManager == nil {
		m.stations = []storage.StationWithRating{}
		return
	}

	// Sort based on sort order
	switch m.sortBy {
	case sortByRatingHigh:
		m.stations = m.ratingsManager.GetTopRated(0) // 0 = no limit
		// Apply filter
		if m.filterBy != filterAllRatings {
			filtered := make([]storage.StationWithRating, 0)
			for _, s := range m.stations {
				if s.Rating != nil && s.Rating.Rating >= m.filterBy.MinRating() {
					filtered = append(filtered, s)
				}
			}
			m.stations = filtered
		}
	case sortByRatingLow:
		// Get all and reverse
		all := m.ratingsManager.GetTopRated(0)
		m.stations = make([]storage.StationWithRating, 0, len(all))
		for i := len(all) - 1; i >= 0; i-- {
			if m.filterBy == filterAllRatings || (all[i].Rating != nil && all[i].Rating.Rating >= m.filterBy.MinRating()) {
				m.stations = append(m.stations, all[i])
			}
		}
	case sortByRecentlyRated:
		m.stations = m.ratingsManager.GetRecentlyRated(0)
		// Apply filter
		if m.filterBy != filterAllRatings {
			filtered := make([]storage.StationWithRating, 0)
			for _, s := range m.stations {
				if s.Rating != nil && s.Rating.Rating >= m.filterBy.MinRating() {
					filtered = append(filtered, s)
				}
			}
			m.stations = filtered
		}
	default:
		m.stations = m.ratingsManager.GetTopRated(0)
	}

	// Convert to list items
	m.stationItems = make([]list.Item, len(m.stations))
	for i, s := range m.stations {
		tagPills := ""
		if m.tagsManager != nil && m.tagRenderer != nil {
			if tags := m.tagsManager.GetTags(s.Station.StationUUID); len(tags) > 0 {
				tagPills = m.tagRenderer.RenderPills(tags)
			}
		}
		m.stationItems[i] = topRatedStationItem{
			station:  s.Station,
			rating:   s.Rating,
			tagPills: tagPills,
		}
	}
	m.stationListModel.SetItems(m.stationItems)
}

func (m TopRatedModel) handleListInput(msg tea.KeyMsg) (TopRatedModel, tea.Cmd) {
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

	case "*":
		// Enter rating mode
		if len(m.stationItems) > 0 {
			m.state = topRatedStateRating
			m.saveMessage = "Press 1-5 to rate, 0 or r to clear"
			m.saveMessageSuccess = true
			m.saveMessageTime = 3
			return m, tickEverySecond()
		}
		return m, nil

	case "enter":
		// Play selected station
		if len(m.stationItems) > 0 {
			selected := m.stationListModel.SelectedItem()
			if item, ok := selected.(topRatedStationItem); ok {
				m.selectedStation = &item.station
				// Check if we have the URL to play
				if item.station.URLResolved != "" {
					if err := m.player.Play(&item.station); err != nil {
						m.err = err
					} else {
						m.state = topRatedStatePlaying
					}
				} else {
					m.saveMessage = "Station URL not available (needs lookup)"
					m.saveMessageSuccess = false
					m.saveMessageTime = 3
					return m, tickEverySecond()
				}
			}
		}
		return m, nil

	case "s":
		// Cycle through sort options
		m.sortBy = (m.sortBy + 1) % TopRatedSort(numTopRatedSortModes)
		m.refreshStationList()
		m.saveMessage = fmt.Sprintf("Sorted by: %s", m.sortBy.String())
		m.saveMessageSuccess = true
		m.saveMessageTime = 2
		return m, tickEverySecond()

	case "f":
		// Cycle through filter options
		m.filterBy = (m.filterBy + 1) % TopRatedFilter(numTopRatedFilterModes)
		m.refreshStationList()
		m.saveMessage = fmt.Sprintf("Filter: %s", m.filterBy.String())
		m.saveMessageSuccess = true
		m.saveMessageTime = 2
		return m, tickEverySecond()

	case "a":
		// Add to favorites
		if len(m.stationItems) > 0 {
			selected := m.stationListModel.SelectedItem()
			if item, ok := selected.(topRatedStationItem); ok {
				m.selectedStation = &item.station
				m.state = topRatedStateSavePrompt
			}
		}
		return m, nil
	}

	// Pass other keys to list
	var cmd tea.Cmd
	m.stationListModel, cmd = m.stationListModel.Update(msg)
	return m, cmd
}

func (m TopRatedModel) handleRatingInput(msg tea.KeyMsg) (TopRatedModel, tea.Cmd) {
	switch msg.String() {
	case "1", "2", "3", "4", "5":
		// Rate the selected station
		if len(m.stationItems) > 0 && m.ratingsManager != nil {
			selected := m.stationListModel.SelectedItem()
			if item, ok := selected.(topRatedStationItem); ok {
				rating := int(msg.String()[0] - '0')
				if err := m.ratingsManager.SetRating(&item.station, rating); err != nil {
					m.saveMessage = "Failed to save rating"
					m.saveMessageSuccess = false
				} else {
					stars := storage.RenderStars(rating, true)
					m.saveMessage = fmt.Sprintf("%s Rating saved", stars)
					m.saveMessageSuccess = true
					m.refreshStationList()
				}
				m.saveMessageTime = 2
				m.state = topRatedStateList
				return m, tickEverySecond()
			}
		}
		m.state = topRatedStateList
		return m, nil

	case "0", "r":
		// Remove rating
		if len(m.stationItems) > 0 && m.ratingsManager != nil {
			selected := m.stationListModel.SelectedItem()
			if item, ok := selected.(topRatedStationItem); ok {
				if err := m.ratingsManager.RemoveRating(item.station.StationUUID); err != nil {
					m.saveMessage = "Failed to remove rating"
					m.saveMessageSuccess = false
				} else {
					m.saveMessage = "Rating removed"
					m.saveMessageSuccess = true
					m.refreshStationList()
				}
				m.saveMessageTime = 2
				m.state = topRatedStateList
				return m, tickEverySecond()
			}
		}
		m.state = topRatedStateList
		return m, nil

	case "esc", "q":
		// Cancel rating mode
		m.state = topRatedStateList
		m.saveMessage = ""
		return m, nil
	}

	// Timeout - exit rating mode
	m.state = topRatedStateList
	return m, nil
}

func (m TopRatedModel) handlePlayingInput(msg tea.KeyMsg) (TopRatedModel, tea.Cmd) {
	switch msg.String() {
	case "q", "esc", "m":
		if m.player != nil {
			_ = m.player.Stop()
		}
		m.state = topRatedStateList
		return m, nil

	case "*":
		// Enter rating mode while playing
		m.state = topRatedStateRating
		m.saveMessage = "Press 1-5 to rate, 0 or r to clear"
		m.saveMessageSuccess = true
		m.saveMessageTime = 3
		return m, tickEverySecond()

	case "s":
		// Stop playback
		if m.player != nil {
			_ = m.player.Stop()
		}
		m.state = topRatedStateList
		return m, nil
	}
	return m, nil
}

func (m TopRatedModel) handleSavePromptInput(msg tea.KeyMsg) (TopRatedModel, tea.Cmd) {
	switch msg.String() {
	case "y", "Y", "enter":
		// Save to My-favorites
		if m.selectedStation != nil {
			store := storage.NewStorage(m.favoritePath)
			err := store.AddStation(context.TODO(), "My-favorites", *m.selectedStation)
			if err != nil {
				if errors.Is(err, storage.ErrDuplicateStation) {
					m.saveMessage = "Already in My-favorites"
				} else {
					m.saveMessage = fmt.Sprintf("Error: %v", err)
				}
				m.saveMessageSuccess = false
			} else {
				m.saveMessage = "✓ Saved to My-favorites"
				m.saveMessageSuccess = true
			}
			m.saveMessageTime = 3
		}
		m.state = topRatedStateList
		return m, tickEverySecond()

	case "n", "N", "esc", "q":
		m.state = topRatedStateList
		return m, nil

	case "l", "L":
		// Select list to save to
		m.loadAvailableLists()
		m.state = topRatedStateSelectList
		return m, nil
	}
	return m, nil
}

func (m TopRatedModel) handleSelectListInput(msg tea.KeyMsg) (TopRatedModel, tea.Cmd) {
	switch msg.String() {
	case "esc", "q":
		m.state = topRatedStateList
		return m, nil

	case "enter":
		// Save to selected list
		if m.selectedStation != nil && len(m.listItems) > 0 {
			selected := m.listModel.SelectedItem()
			if item, ok := selected.(components.MenuItem); ok {
				listName := item.Title()
				store := storage.NewStorage(m.favoritePath)
				err := store.AddStation(context.TODO(), listName, *m.selectedStation)
				if err != nil {
					if errors.Is(err, storage.ErrDuplicateStation) {
						m.saveMessage = fmt.Sprintf("Already in %s", listName)
					} else {
						m.saveMessage = fmt.Sprintf("Error: %v", err)
					}
					m.saveMessageSuccess = false
				} else {
					m.saveMessage = fmt.Sprintf("✓ Saved to %s", listName)
					m.saveMessageSuccess = true
				}
				m.saveMessageTime = 3
				m.state = topRatedStateList
				return m, tickEverySecond()
			}
		}
		m.state = topRatedStateList
		return m, nil
	}

	var cmd tea.Cmd
	m.listModel, cmd = m.listModel.Update(msg)
	return m, cmd
}

func (m *TopRatedModel) loadAvailableLists() {
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

// SetMetadataManager sets the metadata manager for play tracking
func (m *TopRatedModel) SetMetadataManager(mgr *storage.MetadataManager) {
	m.metadataManager = mgr
	if m.player != nil {
		m.player.SetMetadataManager(mgr)
	}
}

func (m TopRatedModel) View() string {
	// Handle help overlay
	if m.helpModel.IsVisible() {
		return m.helpModel.View()
	}

	var content strings.Builder

	// Title
	title := titleStyle().Render("★ Top Rated Stations")
	content.WriteString(title)
	content.WriteString("\n\n")

	// Filter and sort info
	filterInfo := helpStyle().Render(fmt.Sprintf("Filter: %s  |  Sort: %s", m.filterBy.String(), m.sortBy.String()))
	content.WriteString(filterInfo)
	content.WriteString("\n\n")

	switch m.state {
	case topRatedStateList, topRatedStateRating:
		if len(m.stationItems) == 0 {
			content.WriteString(helpStyle().Render("No rated stations yet. Press * then 1-5 to rate stations!"))
		} else {
			content.WriteString(m.stationListModel.View())
		}

	case topRatedStatePlaying:
		if m.selectedStation != nil {
			content.WriteString(stationNameStyle().Render(m.selectedStation.Name))
			content.WriteString("\n")
			if m.selectedStation.Country != "" {
				content.WriteString(helpStyle().Render(m.selectedStation.Country))
				content.WriteString("\n")
			}
			// Show current rating
			if m.ratingsManager != nil {
				rating := m.ratingsManager.GetRating(m.selectedStation.StationUUID)
				if rating != nil {
					stars := storage.RenderStars(rating.Rating, true)
					content.WriteString(highlightStyle().Render(stars))
					content.WriteString("\n")
				}
			}
			content.WriteString("\n")
			content.WriteString(successStyle().Render("▶ Now Playing"))
		}

	case topRatedStateSavePrompt:
		if m.selectedStation != nil {
			fmt.Fprintf(&content, "Save \"%s\" to favorites?\n\n", m.selectedStation.TrimName())
			content.WriteString("[Enter/y] Save to My-favorites\n")
			content.WriteString("[l] Choose list\n")
			content.WriteString("[Esc] Cancel")
		}

	case topRatedStateSelectList:
		content.WriteString("Select list to save to:\n\n")
		content.WriteString(m.listModel.View())
	}

	// Show save message
	if m.saveMessage != "" {
		content.WriteString("\n")
		if m.saveMessageSuccess {
			content.WriteString(successStyle().Render(m.saveMessage))
		} else {
			content.WriteString(errorStyle().Render(m.saveMessage))
		}
	}

	// Help text
	var helpText string
	switch m.state {
	case topRatedStateList:
		helpText = "↑↓/jk: Navigate • Enter: Play • *1-5: Rate • s: Sort • f: Filter • ?: Help • Esc: Back"
	case topRatedStateRating:
		helpText = "1-5: Set rating • 0/r: Remove rating • Esc: Cancel"
	case topRatedStatePlaying:
		helpText = "s: Stop • *1-5: Rate • Esc: Back"
	}

	return RenderPageWithBottomHelp(PageLayout{
		Content: content.String(),
		Help:    helpText,
	}, m.height)
}
