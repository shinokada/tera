package ui

import (
	"fmt"
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/shinokada/tera/v3/internal/api"
	"github.com/shinokada/tera/v3/internal/blocklist"
	"github.com/shinokada/tera/v3/internal/player"
	"github.com/shinokada/tera/v3/internal/storage"
	"github.com/shinokada/tera/v3/internal/ui/components"
)

// browseTagsState tracks sub-views within the Browse by Tag screen.
type browseTagsState int

const (
	browseTagsStateList       browseTagsState = iota // list of all tags
	browseTagsStateDetail                            // stations for a selected tag
	browseTagsStatePlaying                           // playing a station from a tag
	browseTagsStateTagInput                          // entering a new tag (single via 't')
	browseTagsStateManageTags                        // full tag manager via 'T'
)

// tagStat holds a tag name and how many stations carry it.
type tagStat struct {
	tag   string
	count int
}

// BrowseTagsModel is the model for the "Browse by Tag" screen.
type BrowseTagsModel struct {
	state            browseTagsState
	tagsManager      *storage.TagsManager
	ratingsManager   *storage.RatingsManager
	metadataManager  *storage.MetadataManager
	blocklistManager *blocklist.Manager
	starRenderer     *components.StarRenderer
	tagRenderer      *components.TagRenderer

	// Tag list view
	tagStats        []tagStat
	tagCursor       int
	deleteConfirm   bool // true = waiting for second 'd' to confirm tag deletion

	// Station detail view
	selectedTag    string
	taggedUUIDs    []string      // UUIDs matching selectedTag
	detailStations []api.Station // hydrated stations (from metadata)
	stationCursor  int

	// Playing
	selectedStation *api.Station
	player          *player.MPVPlayer
	ratingMode      bool
	tagInput    components.TagInput
	manageTags  components.ManageTags

	// Help overlay
	helpModel components.HelpModel

	// Shared state
	saveMessage     string
	saveMessageTime int
	width           int
	height          int
}

// NewBrowseTagsModel creates a Browse by Tag model.
func NewBrowseTagsModel(
	tagsManager *storage.TagsManager,
	ratingsManager *storage.RatingsManager,
	metadataManager *storage.MetadataManager,
	starRenderer *components.StarRenderer,
	blocklistManager *blocklist.Manager,
) BrowseTagsModel {
	m := BrowseTagsModel{
		state:            browseTagsStateList,
		tagsManager:      tagsManager,
		ratingsManager:   ratingsManager,
		metadataManager:  metadataManager,
		blocklistManager: blocklistManager,
		starRenderer:     starRenderer,
		tagRenderer:      components.NewTagRenderer(),
		player:           player.NewMPVPlayer(),
		helpModel:        components.NewHelpModel(components.CreateTagsPlayingHelp()),
		width:            80,
		height:           24,
	}
	m.loadTagStats()
	return m
}

// loadTagStats recomputes the tagStats slice from the TagsManager.
func (m *BrowseTagsModel) loadTagStats() {
	if m.tagsManager == nil {
		m.tagStats = nil
		return
	}
	allTags := m.tagsManager.GetAllTags()
	m.tagStats = make([]tagStat, 0, len(allTags))
	for _, tag := range allTags {
		uuids := m.tagsManager.GetStationsByTag(tag)
		if len(uuids) > 0 {
			m.tagStats = append(m.tagStats, tagStat{tag: tag, count: len(uuids)})
		}
	}
	sort.Slice(m.tagStats, func(i, j int) bool {
		return m.tagStats[i].tag < m.tagStats[j].tag
	})
}

// Init satisfies bubbletea.
func (m BrowseTagsModel) Init() tea.Cmd { return tickEverySecond() }

// Update handles messages.
func (m BrowseTagsModel) Update(msg tea.Msg) (BrowseTagsModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.helpModel.SetSize(msg.Width, msg.Height)
		return m, nil

	case components.TagSubmittedMsg:
		if m.state == browseTagsStateManageTags {
			var cmd tea.Cmd
			m.manageTags, cmd = m.manageTags.HandleTagSubmitted(msg.Tag)
			return m, cmd
		}
		if m.selectedStation != nil && m.tagsManager != nil {
			if err := m.tagsManager.AddTag(m.selectedStation.StationUUID, msg.Tag); err != nil {
				m.saveMessage = fmt.Sprintf("✗ %v", err)
			} else {
				m.saveMessage = fmt.Sprintf("✓ Added tag: %s", msg.Tag)
				m.loadTagStats() // keep list counts fresh after a new tag is added
			}
			m.saveMessageTime = messageDisplayShort
		}
		m.state = browseTagsStatePlaying
		return m, nil

	case components.TagCancelledMsg:
		if m.state == browseTagsStateManageTags {
			m.manageTags = m.manageTags.HandleTagCancelled()
			return m, nil
		}
		m.state = browseTagsStatePlaying
		return m, nil

	case components.ManageTagsDoneMsg:
		if m.selectedStation != nil && m.tagsManager != nil {
			if err := m.tagsManager.SetTags(m.selectedStation.StationUUID, msg.Tags); err != nil {
				m.saveMessage = fmt.Sprintf("✗ %v", err)
			} else {
				m.saveMessage = "✓ Tags updated"
				m.loadTagStats()
			}
			m.saveMessageTime = messageDisplayShort
		}
		m.state = browseTagsStatePlaying
		return m, nil

	case components.ManageTagsCancelledMsg:
		m.state = browseTagsStatePlaying
		return m, nil

	case tea.KeyMsg:
		if m.helpModel.IsVisible() {
			var cmd tea.Cmd
			m.helpModel, cmd = m.helpModel.Update(msg)
			return m, cmd
		}
		switch m.state {
		case browseTagsStateList:
			return m.updateTagList(msg)
		case browseTagsStateDetail:
			return m.updateDetail(msg)
		case browseTagsStatePlaying:
			return m.updatePlaying(msg)
		case browseTagsStateTagInput:
			var cmd tea.Cmd
			m.tagInput, cmd = m.tagInput.Update(msg)
			return m, cmd
		case browseTagsStateManageTags:
			var cmd tea.Cmd
			m.manageTags, cmd = m.manageTags.Update(msg)
			return m, cmd
		}

	case playbackStartedMsg:
		return m, nil

	case playbackErrorMsg:
		m.saveMessage = fmt.Sprintf("✗ %v", msg.err)
		m.saveMessageTime = messageDisplayShort
		m.state = browseTagsStateDetail
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

	return m, nil
}

func (m BrowseTagsModel) updateTagList(msg tea.KeyMsg) (BrowseTagsModel, tea.Cmd) {
	switch msg.String() {
	case "esc", "m":
		if m.deleteConfirm {
			m.deleteConfirm = false
			m.saveMessage = ""
			m.saveMessageTime = 0
			return m, nil
		}
		return m, func() tea.Msg { return backToMainMsg{} }
	case "up", "k":
		m.deleteConfirm = false
		if m.tagCursor > 0 {
			m.tagCursor--
		}
	case "down", "j":
		m.deleteConfirm = false
		if m.tagCursor < len(m.tagStats)-1 {
			m.tagCursor++
		}
	case "enter":
		if len(m.tagStats) == 0 {
			break
		}
		m.deleteConfirm = false
		m.saveMessage = ""
		m.selectedTag = m.tagStats[m.tagCursor].tag
		m.loadDetailStations()
		m.stationCursor = 0
		m.state = browseTagsStateDetail
	case "d":
		// Delete tag from all stations (requires a second 'd' to confirm).
		if len(m.tagStats) == 0 {
			break
		}
		tag := m.tagStats[m.tagCursor].tag
		if !m.deleteConfirm {
			m.deleteConfirm = true
			m.saveMessage = fmt.Sprintf("⚠ Delete tag \"%s\" from all stations? Press d again to confirm, Esc to cancel", tag)
			m.saveMessageTime = messageDisplayPersistent
			break
		}
		m.deleteConfirm = false
		failures := m.deleteTagFromAll(tag)
		m.loadTagStats()
		if m.tagCursor >= len(m.tagStats) && m.tagCursor > 0 {
			m.tagCursor = len(m.tagStats) - 1
		}
		if failures == 0 {
			m.saveMessage = fmt.Sprintf("✓ Deleted tag: %s", tag)
		} else {
			m.saveMessage = fmt.Sprintf("✗ Deleted tag: %s (%d station(s) could not be updated)", tag, failures)
		}
		m.saveMessageTime = messageDisplayShort
	}
	return m, nil
}

func (m BrowseTagsModel) updateDetail(msg tea.KeyMsg) (BrowseTagsModel, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.state = browseTagsStateList
		m.detailStations = nil
		m.loadTagStats()
		if m.tagCursor >= len(m.tagStats) && m.tagCursor > 0 {
			m.tagCursor = len(m.tagStats) - 1
		}
	case "up", "k":
		if m.stationCursor > 0 {
			m.stationCursor--
		}
	case "down", "j":
		if m.stationCursor < len(m.detailStations)-1 {
			m.stationCursor++
		}
	case "enter":
		if len(m.detailStations) == 0 {
			break
		}
		station := m.detailStations[m.stationCursor]
		if station.URLResolved == "" {
			m.saveMessage = "✗ No URL cached for this station — play it from search first"
			m.saveMessageTime = messageDisplayShort
			break
		}
		m.selectedStation = &station
		m.state = browseTagsStatePlaying
		return m, m.playSelected()
	}
	return m, nil
}

func (m BrowseTagsModel) updatePlaying(msg tea.KeyMsg) (BrowseTagsModel, tea.Cmd) {
	if m.ratingMode {
		return m.handleRatingInput(msg)
	}
	switch msg.String() {
	case "esc":
		if m.player != nil {
			_ = m.player.Stop()
		}
		m.state = browseTagsStateDetail
		m.selectedStation = nil
		m.loadDetailStations() // refresh in case tags were modified while playing
	case "0":
		if m.player != nil {
			_ = m.player.Stop()
		}
		return m, func() tea.Msg { return backToMainMsg{} }
	case " ":
		if m.player != nil {
			if err := m.player.TogglePause(); err == nil {
				if m.player.IsPaused() {
					m.saveMessage = "⏸ Paused"
				} else {
					m.saveMessage = "▶ Resumed"
				}
				m.saveMessageTime = messageDisplayShort
			}
		}
	case "r":
		if m.selectedStation != nil && m.ratingsManager != nil {
			m.ratingMode = true
			m.saveMessage = "Press 1-5 to rate, 0 to remove, Esc to cancel"
			m.saveMessageTime = messageDisplayPersistent
		}
	case "/":
		if m.player != nil {
			v := m.player.DecreaseVolume(5)
			m.saveMessage = fmt.Sprintf("Volume: %d%%", v)
			m.saveMessageTime = messageDisplayShort
		}
	case "*":
		if m.player != nil {
			v := m.player.IncreaseVolume(5)
			m.saveMessage = fmt.Sprintf("Volume: %d%%", v)
			m.saveMessageTime = messageDisplayShort
		}
	case "m":
		if m.player != nil {
			muted, vol := m.player.ToggleMute()
			if muted {
				m.saveMessage = "Volume: Muted"
			} else {
				m.saveMessage = fmt.Sprintf("Volume: %d%%", vol)
			}
			m.saveMessageTime = messageDisplayShort
		}
	case "t":
		if m.selectedStation != nil && m.tagsManager != nil {
			allTags := m.tagsManager.GetAllTags()
			w := m.width - 4
			if w < 20 {
				w = 20
			}
			m.tagInput = components.NewTagInput(allTags, w)
			m.state = browseTagsStateTagInput
		}
	case "T":
		if m.selectedStation != nil && m.tagsManager != nil {
			currentTags := m.tagsManager.GetTags(m.selectedStation.StationUUID)
			allTags := m.tagsManager.GetAllTags()
			m.manageTags = components.NewManageTags(m.selectedStation.TrimName(), currentTags, allTags, m.width)
			m.state = browseTagsStateManageTags
		}
	case "?":
		m.helpModel.SetSize(m.width, m.height)
		m.helpModel.Toggle()
		return m, nil
	}
	return m, nil
}

func (m BrowseTagsModel) handleRatingInput(msg tea.KeyMsg) (BrowseTagsModel, tea.Cmd) {
	m.ratingMode = false
	if m.selectedStation == nil || m.ratingsManager == nil {
		return m, nil
	}
	k := msg.String()
	if len(k) == 1 && k[0] >= '1' && k[0] <= '5' {
		rating := int(k[0] - '0')
		if err := m.ratingsManager.SetRating(m.selectedStation, rating); err == nil {
			stars := ""
			if m.starRenderer != nil {
				stars = m.starRenderer.RenderCompactPlain(rating) + " "
			}
			m.saveMessage = fmt.Sprintf("✓ %sRated!", stars)
		} else {
			m.saveMessage = fmt.Sprintf("✗ %v", err)
		}
		m.saveMessageTime = messageDisplayShort
		return m, nil
	}
	if k == "0" {
		if err := m.ratingsManager.RemoveRating(m.selectedStation.StationUUID); err == nil {
			m.saveMessage = "✓ Rating removed"
		} else {
			m.saveMessage = fmt.Sprintf("✗ %v", err)
		}
		m.saveMessageTime = messageDisplayShort
		return m, nil
	}
	m.saveMessage = ""
	m.saveMessageTime = 0
	return m, nil
}

// loadDetailStations fetches station UUIDs tagged with selectedTag and hydrates
// them from metadata (using stored name/country/codec/bitrate).
func (m *BrowseTagsModel) loadDetailStations() {
	m.taggedUUIDs = m.tagsManager.GetStationsByTag(m.selectedTag)
	m.detailStations = hydrateStations(m.metadataManager, m.taggedUUIDs)
	sort.Slice(m.detailStations, func(i, j int) bool {
		return strings.ToLower(m.detailStations[i].TrimName()) < strings.ToLower(m.detailStations[j].TrimName())
	})
}

// deleteTagFromAll removes a tag from every station that has it.
// Returns the number of stations from which removal failed.
func (m *BrowseTagsModel) deleteTagFromAll(tag string) int {
	uuids := m.tagsManager.GetStationsByTag(tag)
	var failures int
	for _, uuid := range uuids {
		if err := m.tagsManager.RemoveTag(uuid, tag); err != nil {
			failures++
		}
	}
	return failures
}

// playSelected starts playing the currently selected station.
func (m BrowseTagsModel) playSelected() tea.Cmd {
	if m.selectedStation == nil {
		return nil
	}
	station := *m.selectedStation
	return func() tea.Msg {
		if err := m.player.Play(&station); err != nil {
			return playbackErrorMsg{err}
		}
		return playbackStartedMsg{}
	}
}

// View renders the Browse by Tag screen.
func (m BrowseTagsModel) View() string {
	if m.helpModel.IsVisible() {
		return m.helpModel.View()
	}
	switch m.state {
	case browseTagsStateList:
		return m.viewTagList()
	case browseTagsStateDetail:
		return m.viewDetail()
	case browseTagsStatePlaying:
		return m.viewPlaying()
	case browseTagsStateTagInput:
		return m.viewTagInput()
	case browseTagsStateManageTags:
		return m.viewManageTags()
	}
	return ""
}

// viewManageTags renders the ManageTags dialog overlay.
func (m BrowseTagsModel) viewManageTags() string {
	return RenderPageWithBottomHelp(PageLayout{
		Title:   "🏷 Manage Tags",
		Content: m.manageTags.View(),
		Help:    "Space/Enter: Toggle • ↑↓/jk: Navigate • d: Done • Esc: Cancel",
	}, m.height)
}

// viewTagInput renders the tag input overlay.
func (m BrowseTagsModel) viewTagInput() string {
	var sb strings.Builder
	if m.selectedStation != nil {
		sb.WriteString(boldStyle().Render(m.selectedStation.TrimName()))
		sb.WriteString("\n\n")
	}
	sb.WriteString(m.tagInput.View())
	return RenderPageWithBottomHelp(PageLayout{
		Title:   "🏷 Add Tag",
		Content: sb.String(),
		Help:    "Enter: Add • Tab: Complete • ↑↓: Navigate • Esc: Cancel",
	}, m.height)
}

func (m BrowseTagsModel) viewTagList() string {
	var sb strings.Builder

	if len(m.tagStats) == 0 {
		sb.WriteString(infoStyle().Render("ℹ No tagged stations yet — press t while playing to add tags!"))
	} else {
		sb.WriteString(subtitleStyle().Render(fmt.Sprintf("Your Tags (%d total)", len(m.tagStats))))
		sb.WriteString("\n\n")
		for i, ts := range m.tagStats {
			line := fmt.Sprintf("%-30s  %d station", ts.tag, ts.count)
			if ts.count != 1 {
				line += "s"
			}
			if i == m.tagCursor {
				sb.WriteString(selectedItemStyle().Render("> " + line))
			} else {
				sb.WriteString(normalItemStyle().Render("  " + line))
			}
			sb.WriteString("\n")
		}
	}

	renderSaveMessage(&sb, m.saveMessage)

	return RenderPageWithBottomHelp(PageLayout{
		Title:   "🏷 Browse by Tag",
		Content: sb.String(),
		Help:    "↑↓/jk: Navigate • Enter: View stations • d: Delete tag (confirm) • Esc/m: Back",
	}, m.height)
}

func (m BrowseTagsModel) viewDetail() string {
	var sb strings.Builder

	sb.WriteString(subtitleStyle().Render(fmt.Sprintf("Stations tagged \"%s\" (%d)", m.selectedTag, len(m.detailStations))))
	sb.WriteString("\n\n")

	if len(m.detailStations) == 0 {
		sb.WriteString(infoStyle().Render("No stations with this tag."))
	} else {
		for i, s := range m.detailStations {
			var parts []string
			name := s.TrimName()
			if s.URLResolved == "" {
				name += dimStyle().Render(" (no URL)")
			}
			parts = append(parts, name)
			if s.Country != "" {
				parts = append(parts, s.Country)
			}
			if s.Codec != "" {
				codec := s.Codec
				if s.Bitrate > 0 {
					codec += fmt.Sprintf(" %dkbps", s.Bitrate)
				}
				parts = append(parts, codec)
			}
			// Append all tags for this station.
			tags := m.tagsManager.GetTags(s.StationUUID)
			if len(tags) > 0 && m.tagRenderer != nil {
				parts = append(parts, m.tagRenderer.RenderPills(tags))
			}
			line := strings.Join(parts, " • ")
			if i == m.stationCursor {
				sb.WriteString(selectedItemStyle().Render("> " + line))
			} else {
				sb.WriteString(normalItemStyle().Render("  " + line))
			}
			sb.WriteString("\n")
		}
	}

	renderSaveMessage(&sb, m.saveMessage)

	return RenderPageWithBottomHelp(PageLayout{
		Title:   fmt.Sprintf("🏷 Tag: %s", m.selectedTag),
		Content: sb.String(),
		Help:    "↑↓/jk: Navigate • Enter: Play • Esc: Back",
	}, m.height)
}

func (m BrowseTagsModel) viewPlaying() string {
	if m.selectedStation == nil {
		return ""
	}
	var sb strings.Builder

	var metadata *storage.StationMetadata
	if m.metadataManager != nil {
		metadata = m.metadataManager.GetMetadata(m.selectedStation.StationUUID)
	}
	var rating int
	if m.ratingsManager != nil {
		if r := m.ratingsManager.GetRating(m.selectedStation.StationUUID); r != nil {
			rating = r.Rating
		}
	}
	sb.WriteString(RenderStationDetailsWithRating(*m.selectedStation, false, metadata, rating, m.starRenderer))

	sb.WriteString("\n")
	if m.player != nil && m.player.IsPlaying() {
		if track := m.player.GetCachedTrack(); IsValidTrackMetadata(track, m.selectedStation.TrimName()) {
			sb.WriteString(successStyle().Render("▶ Now Playing:") + " " + infoStyle().Render(track))
		} else {
			sb.WriteString(successStyle().Render("▶ Playing..."))
		}
	} else {
		sb.WriteString(infoStyle().Render("⏸ Stopped"))
	}

	// Tag display.
	if m.tagsManager != nil && m.tagRenderer != nil {
		tags := m.tagsManager.GetTags(m.selectedStation.StationUUID)
		sb.WriteString("\n")
		if len(tags) > 0 {
			fmt.Fprintf(&sb, "Tags: %s", m.tagRenderer.RenderList(tags))
		} else {
			sb.WriteString(dimStyle().Render("No tags — press t to add one"))
		}
	}

	if m.saveMessage != "" {
		sb.WriteString("\n")
	}
	renderSaveMessage(&sb, m.saveMessage)

	helpText := "Space: Pause • r: Rate • t: Tag • /*: Volume • 0: Main Menu • ?: Help • Esc: Back"
	return RenderPageWithBottomHelp(PageLayout{
		Title:   "🎵 Now Playing",
		Content: sb.String(),
		Help:    helpText,
	}, m.height)
}
