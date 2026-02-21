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

// tagPlaylistsState tracks sub-views within the Tag Playlists screen.
type tagPlaylistsState int

const (
	tagPlaylistsStateList    tagPlaylistsState = iota // list of playlists
	tagPlaylistsStateCreate                           // multi-step create/edit dialog
	tagPlaylistsStateDetail                           // stations for selected playlist
	tagPlaylistsStatePlaying                          // playing a station
)

// createStep tracks which step of the create/edit wizard is active.
type createStep int

const (
	createStepName      createStep = iota // entering playlist name
	createStepTags                        // selecting tags
	createStepMatchMode                   // choosing any/all
)

// playlistEntry is a view model for one row in the playlist list.
type playlistEntry struct {
	name      string
	tags      []string
	matchMode string
	count     int
}

// TagPlaylistsModel is the Bubble Tea model for the Tag Playlists screen.
type TagPlaylistsModel struct {
	state            tagPlaylistsState
	tagsManager      *storage.TagsManager
	ratingsManager   *storage.RatingsManager
	metadataManager  *storage.MetadataManager
	blocklistManager *blocklist.Manager
	starRenderer     *components.StarRenderer
	tagRenderer      *components.TagRenderer
	player           *player.MPVPlayer

	// Playlist list view
	playlists      []playlistEntry
	listCursor     int
	deleteConfirm  bool // true = waiting for second 'd' to confirm playlist deletion

	// Create / edit wizard
	isEditing    bool   // true = editing existing playlist
	editName     string // original name (for edit/delete-then-recreate)
	step         createStep
	inputBuffer  string          // typed playlist name
	allTags      []string        // all known tags for selection
	tagCursor    int             // cursor in tag selector
	selectedTags map[string]bool // which tags are checked

	// Match mode toggle (create step 3)
	matchMode string // "any" or "all"

	// Station detail view
	selectedPlaylist *playlistEntry
	detailStations   []api.Station
	stationCursor    int

	// Playing
	selectedStation *api.Station
	ratingMode      bool

	// Shared
	saveMessage     string
	saveMessageTime int
	width           int
	height          int
}

// NewTagPlaylistsModel creates a Tag Playlists model.
func NewTagPlaylistsModel(
	tagsManager *storage.TagsManager,
	ratingsManager *storage.RatingsManager,
	metadataManager *storage.MetadataManager,
	starRenderer *components.StarRenderer,
	blocklistManager *blocklist.Manager,
) TagPlaylistsModel {
	m := TagPlaylistsModel{
		state:            tagPlaylistsStateList,
		tagsManager:      tagsManager,
		ratingsManager:   ratingsManager,
		metadataManager:  metadataManager,
		blocklistManager: blocklistManager,
		starRenderer:     starRenderer,
		tagRenderer:      components.NewTagRenderer(),
		player:           player.NewMPVPlayer(),
		matchMode:        "any",
		selectedTags:     make(map[string]bool),
		width:            80,
		height:           24,
	}
	m.loadPlaylists()
	return m
}

// Init satisfies the bubbletea Model interface.
func (m TagPlaylistsModel) Init() tea.Cmd { return tickEverySecond() }

// Update handles all incoming messages.
func (m TagPlaylistsModel) Update(msg tea.Msg) (TagPlaylistsModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch m.state {
		case tagPlaylistsStateList:
			return m.updateList(msg)
		case tagPlaylistsStateCreate:
			return m.updateCreate(msg)
		case tagPlaylistsStateDetail:
			return m.updateDetail(msg)
		case tagPlaylistsStatePlaying:
			return m.updatePlaying(msg)
		}

	case tickMsg:
		if m.saveMessageTime > 0 {
			m.saveMessageTime--
			if m.saveMessageTime == 0 {
				m.saveMessage = ""
			}
		}
		return m, tickEverySecond()

	case playbackStartedMsg:
		return m, nil

	case playbackErrorMsg:
		m.saveMessage = fmt.Sprintf("✗ %v", msg.err)
		m.saveMessageTime = messageDisplayShort
		m.state = tagPlaylistsStateDetail
		return m, nil
	}

	return m, nil
}

// ---------------------------------------------------------------------------
// Playlist list view
// ---------------------------------------------------------------------------

func (m TagPlaylistsModel) updateList(msg tea.KeyMsg) (TagPlaylistsModel, tea.Cmd) {
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
		if m.listCursor > 0 {
			m.listCursor--
		}
	case "down", "j":
		m.deleteConfirm = false
		if m.listCursor < len(m.playlists)-1 {
			m.listCursor++
		}
	case "enter":
		if len(m.playlists) == 0 {
			break
		}
		m.deleteConfirm = false
		m.saveMessage = ""
		pl := m.playlists[m.listCursor]
		m.selectedPlaylist = &pl
		m.loadDetailStations()
		m.stationCursor = 0
		m.state = tagPlaylistsStateDetail

	case "n":
		// New playlist
		m.deleteConfirm = false
		m.beginCreate(false, "")
	case "e":
		// Edit selected playlist
		m.deleteConfirm = false
		if len(m.playlists) > 0 {
			m.beginCreate(true, m.playlists[m.listCursor].name)
		}
	case "d":
		// Delete selected playlist (requires a second 'd' to confirm).
		if len(m.playlists) == 0 {
			break
		}
		name := m.playlists[m.listCursor].name
		if !m.deleteConfirm {
			m.deleteConfirm = true
			m.saveMessage = fmt.Sprintf("⚠ Delete playlist \"%s\"? Press d again to confirm, Esc to cancel", name)
			m.saveMessageTime = -1
			break
		}
		m.deleteConfirm = false
		if err := m.tagsManager.DeletePlaylist(name); err == nil {
			m.saveMessage = fmt.Sprintf("✓ Deleted: %s", name)
			m.saveMessageTime = messageDisplayShort
			m.loadPlaylists()
			if m.listCursor >= len(m.playlists) && m.listCursor > 0 {
				m.listCursor = len(m.playlists) - 1
			}
		} else {
			m.saveMessage = fmt.Sprintf("✗ %v", err)
			m.saveMessageTime = messageDisplayShort
		}
	}
	return m, nil
}

// beginCreate initialises the wizard for a new or edited playlist.
func (m *TagPlaylistsModel) beginCreate(editing bool, existingName string) {
	m.isEditing = editing
	m.editName = existingName
	m.inputBuffer = existingName
	m.selectedTags = make(map[string]bool)
	m.matchMode = "any"
	m.allTags = m.tagsManager.GetAllTags()
	m.tagCursor = 0
	m.step = createStepName
	m.saveMessage = ""
	m.saveMessageTime = 0

	if editing {
		pl := m.tagsManager.GetPlaylist(existingName)
		if pl != nil {
			for _, t := range pl.Tags {
				m.selectedTags[t] = true
			}
			m.matchMode = pl.MatchMode
			// Ensure all existing selected tags appear in allTags.
			for t := range m.selectedTags {
				found := false
				for _, at := range m.allTags {
					if at == t {
						found = true
						break
					}
				}
				if !found {
					m.allTags = append(m.allTags, t)
				}
			}
			sort.Strings(m.allTags)
		}
	}

	m.state = tagPlaylistsStateCreate
}

// ---------------------------------------------------------------------------
// Create / Edit wizard
// ---------------------------------------------------------------------------

func (m TagPlaylistsModel) updateCreate(msg tea.KeyMsg) (TagPlaylistsModel, tea.Cmd) {
	switch m.step {
	case createStepName:
		return m.updateCreateName(msg)
	case createStepTags:
		return m.updateCreateTags(msg)
	case createStepMatchMode:
		return m.updateCreateMatchMode(msg)
	}
	return m, nil
}

func (m TagPlaylistsModel) updateCreateName(msg tea.KeyMsg) (TagPlaylistsModel, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		m.state = tagPlaylistsStateList
	case tea.KeyEnter:
		name := strings.TrimSpace(m.inputBuffer)
		if name == "" {
			m.saveMessage = "✗ Name cannot be empty"
			m.saveMessageTime = messageDisplayShort
			break
		}
		if len(m.allTags) == 0 {
			m.saveMessage = "✗ No tags available — tag some stations first"
			m.saveMessageTime = messageDisplayShort
			break
		}
		m.inputBuffer = name
		m.step = createStepTags
	case tea.KeyBackspace, tea.KeyDelete:
		if runes := []rune(m.inputBuffer); len(runes) > 0 {
			m.inputBuffer = string(runes[:len(runes)-1])
		}
	case tea.KeyRunes:
		m.inputBuffer += msg.String()
	}
	return m, nil
}

func (m TagPlaylistsModel) updateCreateTags(msg tea.KeyMsg) (TagPlaylistsModel, tea.Cmd) {
	total := len(m.allTags) + 1 // +1 for the "Next →" row
	switch msg.String() {
	case "esc":
		m.step = createStepName
	case "up", "k":
		if m.tagCursor > 0 {
			m.tagCursor--
		}
	case "down", "j":
		if m.tagCursor < total-1 {
			m.tagCursor++
		}
	case " ":
		if m.tagCursor < len(m.allTags) {
			tag := m.allTags[m.tagCursor]
			m.selectedTags[tag] = !m.selectedTags[tag]
		}
	case "enter":
		if m.tagCursor == len(m.allTags) {
			// "Next" row
			if !m.anyTagSelected() {
				m.saveMessage = "✗ Select at least one tag"
				m.saveMessageTime = messageDisplayShort
				break
			}
			m.step = createStepMatchMode
		} else {
			tag := m.allTags[m.tagCursor]
			m.selectedTags[tag] = !m.selectedTags[tag]
		}
	case "n":
		if m.anyTagSelected() {
			m.step = createStepMatchMode
		} else {
			m.saveMessage = "✗ Select at least one tag"
			m.saveMessageTime = messageDisplayShort
		}
	}
	return m, nil
}

func (m TagPlaylistsModel) updateCreateMatchMode(msg tea.KeyMsg) (TagPlaylistsModel, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.step = createStepTags
	case "left", "right", " ", "h", "l":
		if m.matchMode == "any" {
			m.matchMode = "all"
		} else {
			m.matchMode = "any"
		}
	case "enter":
		return m.commitPlaylist()
	}
	return m, nil
}

// anyTagSelected returns true if at least one tag is checked.
func (m *TagPlaylistsModel) anyTagSelected() bool {
	for _, v := range m.selectedTags {
		if v {
			return true
		}
	}
	return false
}

// selectedTagsList returns the checked tags as a sorted slice.
func (m *TagPlaylistsModel) selectedTagsList() []string {
	out := make([]string, 0, len(m.selectedTags))
	for t, sel := range m.selectedTags {
		if sel {
			out = append(out, t)
		}
	}
	sort.Strings(out)
	return out
}

// previewCount returns how many stations would match current wizard selections.
func (m *TagPlaylistsModel) previewCount() int {
	tags := m.selectedTagsList()
	if len(tags) == 0 {
		return 0
	}
	return len(m.tagsManager.GetStationsByTags(tags, m.matchMode == "all"))
}

// commitPlaylist validates and saves the playlist, then returns to the list.
func (m TagPlaylistsModel) commitPlaylist() (TagPlaylistsModel, tea.Cmd) {
	name := strings.TrimSpace(m.inputBuffer)
	tags := m.selectedTagsList()

	var err error
	if m.isEditing {
		// UpdatePlaylist handles both same-name and rename atomically.
		err = m.tagsManager.UpdatePlaylist(m.editName, name, tags, m.matchMode)
	} else {
		err = m.tagsManager.CreatePlaylist(name, tags, m.matchMode)
	}
	if err != nil {
		m.saveMessage = fmt.Sprintf("✗ %v", err)
		m.saveMessageTime = messageDisplayShort
		m.step = createStepName
		return m, nil
	}

	m.loadPlaylists()
	// Move cursor to the newly saved playlist.
	for i, p := range m.playlists {
		if p.name == name {
			m.listCursor = i
			break
		}
	}
	m.state = tagPlaylistsStateList
	action := "Created"
	if m.isEditing {
		action = "Updated"
	}
	m.saveMessage = fmt.Sprintf("✓ %s playlist: %s", action, name)
	m.saveMessageTime = messageDisplayShort
	return m, nil
}

// ---------------------------------------------------------------------------
// Detail view (stations in a playlist)
// ---------------------------------------------------------------------------

func (m TagPlaylistsModel) updateDetail(msg tea.KeyMsg) (TagPlaylistsModel, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.state = tagPlaylistsStateList
		m.selectedPlaylist = nil
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
		m.state = tagPlaylistsStatePlaying
		return m, m.startPlayback()
	}
	return m, nil
}

// ---------------------------------------------------------------------------
// Playing view
// ---------------------------------------------------------------------------

func (m TagPlaylistsModel) updatePlaying(msg tea.KeyMsg) (TagPlaylistsModel, tea.Cmd) {
	if m.ratingMode {
		return m.handleRatingInput(msg)
	}
	switch msg.String() {
	case "esc":
		if m.player != nil {
			_ = m.player.Stop()
		}
		m.state = tagPlaylistsStateDetail
		m.selectedStation = nil
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
			m.saveMessageTime = -1
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
	}
	return m, nil
}

func (m TagPlaylistsModel) handleRatingInput(msg tea.KeyMsg) (TagPlaylistsModel, tea.Cmd) {
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

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func (m *TagPlaylistsModel) loadPlaylists() {
	all := m.tagsManager.GetAllPlaylists()
	m.playlists = make([]playlistEntry, 0, len(all))
	for name, p := range all {
		count := len(m.tagsManager.GetPlaylistStations(name))
		m.playlists = append(m.playlists, playlistEntry{
			name:      name,
			tags:      p.Tags,
			matchMode: p.MatchMode,
			count:     count,
		})
	}
	sort.Slice(m.playlists, func(i, j int) bool {
		return m.playlists[i].name < m.playlists[j].name
	})
}

func (m *TagPlaylistsModel) loadDetailStations() {
	if m.selectedPlaylist == nil {
		return
	}
	uuids := m.tagsManager.GetStationsByTags(
		m.selectedPlaylist.tags,
		m.selectedPlaylist.matchMode == "all",
	)
	m.detailStations = make([]api.Station, 0, len(uuids))
	for _, uuid := range uuids {
		var s api.Station
		s.StationUUID = uuid
		// Hydrate from cached station info if available.
		if m.metadataManager != nil {
			if cached := m.metadataManager.GetCachedStation(uuid); cached != nil {
				s.Name = cached.Name
				s.Country = cached.Country
				s.Codec = cached.Codec
				s.Bitrate = cached.Bitrate
				s.URLResolved = cached.URL
			}
		}
		if s.Name == "" {
			s.Name = uuid
		}
		m.detailStations = append(m.detailStations, s)
	}
	sort.Slice(m.detailStations, func(i, j int) bool {
		return strings.ToLower(m.detailStations[i].TrimName()) <
			strings.ToLower(m.detailStations[j].TrimName())
	})
}

func (m TagPlaylistsModel) startPlayback() tea.Cmd {
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

// ---------------------------------------------------------------------------
// View
// ---------------------------------------------------------------------------

func (m TagPlaylistsModel) View() string {
	switch m.state {
	case tagPlaylistsStateList:
		return m.viewList()
	case tagPlaylistsStateCreate:
		return m.viewCreate()
	case tagPlaylistsStateDetail:
		return m.viewDetail()
	case tagPlaylistsStatePlaying:
		return m.viewPlaying()
	}
	return ""
}

func (m TagPlaylistsModel) viewList() string {
	var sb strings.Builder

	if len(m.playlists) == 0 {
		sb.WriteString(infoStyle().Render("ℹ No playlists yet — press n to create one!"))
	} else {
		sb.WriteString(subtitleStyle().Render(fmt.Sprintf("Tag Playlists (%d)", len(m.playlists))))
		sb.WriteString("\n\n")
		for i, pl := range m.playlists {
			mode := pl.matchMode
			tagStr := strings.Join(pl.tags, ", ")
			const maxTagLen = 35
			if runes := []rune(tagStr); len(runes) > maxTagLen {
				tagStr = string(runes[:maxTagLen-3]) + "..."
			}
			line := fmt.Sprintf("%-22s  [%s]  %-38s  %d station", pl.name, mode, tagStr, pl.count)
			if pl.count != 1 {
				line += "s"
			}
			if i == m.listCursor {
				sb.WriteString(selectedItemStyle().Render("> " + line))
			} else {
				sb.WriteString(normalItemStyle().Render("  " + line))
			}
			sb.WriteString("\n")
		}
	}

	if m.saveMessage != "" {
		sb.WriteString("\n")
		if strings.Contains(m.saveMessage, "✓") {
			sb.WriteString(successStyle().Render(m.saveMessage))
		} else if strings.Contains(m.saveMessage, "✗") {
			sb.WriteString(errorStyle().Render(m.saveMessage))
		} else {
			sb.WriteString(infoStyle().Render(m.saveMessage))
		}
	}

	return RenderPageWithBottomHelp(PageLayout{
		Title:   "🎵 Tag Playlists",
		Content: sb.String(),
		Help:    "↑↓/jk: Navigate • Enter: Play stations • n: New • e: Edit • d: Delete (confirm) • Esc: Back",
	}, m.height)
}

func (m TagPlaylistsModel) viewCreate() string {
	title := "Create Tag Playlist"
	if m.isEditing {
		title = fmt.Sprintf("Edit Playlist: %s", m.editName)
	}

	switch m.step {
	case createStepName:
		return m.viewCreateName(title)
	case createStepTags:
		return m.viewCreateTags(title)
	case createStepMatchMode:
		return m.viewCreateMatchMode(title)
	}
	return ""
}

func (m TagPlaylistsModel) viewCreateName(title string) string {
	var sb strings.Builder
	sb.WriteString(subtitleStyle().Render("Playlist Name:"))
	sb.WriteString("\n\n")
	fmt.Fprintf(&sb, "  %s█", m.inputBuffer)
	sb.WriteString("\n\n")
	sb.WriteString(dimStyle().Render("Type a name and press Enter to continue"))
	if m.saveMessage != "" {
		sb.WriteString("\n")
		sb.WriteString(errorStyle().Render(m.saveMessage))
	}
	return RenderPageWithBottomHelp(PageLayout{
		Title:   title,
		Content: sb.String(),
		Help:    "Enter: Next step • Esc: Cancel",
	}, m.height)
}

func (m TagPlaylistsModel) viewCreateTags(title string) string {
	var sb strings.Builder
	sb.WriteString(subtitleStyle().Render(fmt.Sprintf("Name: %s", m.inputBuffer)))
	sb.WriteString("\n\n")
	sb.WriteString(boldStyle().Render("Select tags (Space/Enter to toggle):"))
	sb.WriteString("\n\n")

	for i, tag := range m.allTags {
		selected := m.selectedTags[tag]
		var box string
		if selected {
			box = successStyle().Render("[✓]")
		} else {
			box = dimStyle().Render("[ ]")
		}
		line := fmt.Sprintf("%s %s", box, tag)
		if i == m.tagCursor {
			sb.WriteString(selectedItemStyle().Render("> " + line))
		} else {
			sb.WriteString("  " + line)
		}
		sb.WriteString("\n")
	}

	// "Next" row
	nextRow := "[ Next → ]"
	if m.tagCursor == len(m.allTags) {
		sb.WriteString(selectedItemStyle().Render("> " + nextRow))
	} else {
		sb.WriteString(dimStyle().Render("  " + nextRow))
	}

	preview := m.previewCount()
	fmt.Fprintf(&sb, "\n\n%s", infoStyle().Render(fmt.Sprintf("Preview: %d station(s) match", preview)))
	if m.saveMessage != "" {
		sb.WriteString("\n")
		sb.WriteString(errorStyle().Render(m.saveMessage))
	}

	return RenderPageWithBottomHelp(PageLayout{
		Title:   title,
		Content: sb.String(),
		Help:    "Space/Enter: Toggle • ↑↓/jk: Navigate • n: Next step • Esc: Back",
	}, m.height)
}

func (m TagPlaylistsModel) viewCreateMatchMode(title string) string {
	var sb strings.Builder
	sb.WriteString(subtitleStyle().Render(fmt.Sprintf("Name: %s", m.inputBuffer)))
	sb.WriteString("\n\n")

	// Selected tags summary.
	tags := m.selectedTagsList()
	sb.WriteString(boldStyle().Render(fmt.Sprintf("Tags: %s", strings.Join(tags, ", "))))
	sb.WriteString("\n\n")

	sb.WriteString(boldStyle().Render("Match mode:"))
	sb.WriteString("\n\n")
	if m.matchMode == "any" {
		sb.WriteString(selectedItemStyle().Render("  (•) Any tag (OR)  — station has at least one of the selected tags"))
		sb.WriteString("\n")
		sb.WriteString(dimStyle().Render("  ( ) All tags (AND) — station has every selected tag"))
	} else {
		sb.WriteString(dimStyle().Render("  ( ) Any tag (OR)  — station has at least one of the selected tags"))
		sb.WriteString("\n")
		sb.WriteString(selectedItemStyle().Render("  (•) All tags (AND) — station has every selected tag"))
	}

	preview := m.previewCount()
	fmt.Fprintf(&sb, "\n\n%s", infoStyle().Render(fmt.Sprintf("Preview: %d station(s) match", preview)))

	return RenderPageWithBottomHelp(PageLayout{
		Title:   title,
		Content: sb.String(),
		Help:    "←→/h/l/Space: Toggle • Enter: Save • Esc: Back",
	}, m.height)
}

func (m TagPlaylistsModel) viewDetail() string {
	if m.selectedPlaylist == nil {
		return ""
	}
	var sb strings.Builder

	mode := m.selectedPlaylist.matchMode
	tagStr := strings.Join(m.selectedPlaylist.tags, ", ")
	sb.WriteString(subtitleStyle().Render(
		fmt.Sprintf(`"%s" — match %s: %s`, m.selectedPlaylist.name, mode, tagStr),
	))
	fmt.Fprintf(&sb, " (%d station", len(m.detailStations))
	if len(m.detailStations) != 1 {
		sb.WriteString("s")
	}
	sb.WriteString(")\n\n")

	if len(m.detailStations) == 0 {
		sb.WriteString(infoStyle().Render("No stations match this playlist."))
		sb.WriteString("\n")
		sb.WriteString(dimStyle().Render("Add tags to stations from any Now Playing view."))
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
			// Show tags for station.
			if m.tagsManager != nil && m.tagRenderer != nil {
				tags := m.tagsManager.GetTags(s.StationUUID)
				if len(tags) > 0 {
					parts = append(parts, m.tagRenderer.RenderPills(tags))
				}
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

	if m.saveMessage != "" {
		sb.WriteString("\n")
		if strings.Contains(m.saveMessage, "✓") {
			sb.WriteString(successStyle().Render(m.saveMessage))
		} else if strings.Contains(m.saveMessage, "✗") {
			sb.WriteString(errorStyle().Render(m.saveMessage))
		} else {
			sb.WriteString(infoStyle().Render(m.saveMessage))
		}
	}

	return RenderPageWithBottomHelp(PageLayout{
		Title:   fmt.Sprintf("🎵 %s", m.selectedPlaylist.name),
		Content: sb.String(),
		Help:    "↑↓/jk: Navigate • Enter: Play • Esc: Back",
	}, m.height)
}

func (m TagPlaylistsModel) viewPlaying() string {
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
			sb.WriteString(dimStyle().Render("No tags"))
		}
	}

	if m.saveMessage != "" {
		sb.WriteString("\n\n")
		if strings.Contains(m.saveMessage, "✓") {
			sb.WriteString(successStyle().Render(m.saveMessage))
		} else if strings.Contains(m.saveMessage, "✗") {
			sb.WriteString(errorStyle().Render(m.saveMessage))
		} else {
			sb.WriteString(infoStyle().Render(m.saveMessage))
		}
	}

	return RenderPageWithBottomHelp(PageLayout{
		Title:   "🎵 Now Playing",
		Content: sb.String(),
		Help:    "Space: Pause/Play • r: Rate • /*: Volume • m: Mute • 0: Main Menu • Esc: Back",
	}, m.height)
}
