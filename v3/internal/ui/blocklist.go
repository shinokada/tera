package ui

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/shinokada/tera/v3/internal/blocklist"
	"github.com/shinokada/tera/v3/internal/config"
	"github.com/shinokada/tera/v3/internal/storage"
	"github.com/shinokada/tera/v3/internal/ui/components"
)

// blocklistState represents the current state in the blocklist screen
type blocklistState int

const (
	blocklistMainMenu blocklistState = iota
	blocklistViewStations
	blocklistConfirmClear
	blocklistRulesMenu
	blocklistBlockByCountry
	blocklistBlockByLanguage
	blocklistBlockByTag
	blocklistViewRules
	blocklistImportExport
	blocklistConfirmDeleteRule
	blocklistConfirmAddRule
	blocklistSearchVisibility
)

// BlocklistModel represents the blocklist screen
type BlocklistModel struct {
	state                blocklistState
	manager              *blocklist.Manager
	mainMenu             list.Model
	rulesMenu            list.Model
	listModel            list.Model
	rulesListModel       list.Model
	stations             []blocklist.BlockedStation
	rules                []blocklist.BlockRule
	selectedRuleIndex    int
	pendingRuleType      blocklist.BlockRuleType
	pendingRuleValue     string
	previousState        blocklistState
	textInput            textinput.Model
	message              string
	messageTime          int
	err                  error
	width                int
	height               int
	showBlockedInSearch  bool // current setting value
}

// blocklistItem wraps a BlockedStation for list.Item interface
type blocklistItem struct {
	station blocklist.BlockedStation
}

func (b blocklistItem) Title() string {
	return b.station.Name
}

func (b blocklistItem) Description() string {
	parts := []string{}

	if b.station.Country != "" {
		parts = append(parts, b.station.Country)
	}
	if b.station.Language != "" {
		parts = append(parts, b.station.Language)
	}
	if b.station.Codec != "" {
		codec := b.station.Codec
		if b.station.Bitrate > 0 {
			codec += fmt.Sprintf(" %dkbps", b.station.Bitrate)
		}
		parts = append(parts, codec)
	}

	return strings.Join(parts, " • ")
}

func (b blocklistItem) FilterValue() string {
	return b.station.Name
}

// NewBlocklistModel creates a new blocklist model
func NewBlocklistModel(manager *blocklist.Manager) BlocklistModel {
	// Load search visibility setting
	showBlocked := false
	if cfg, err := storage.LoadBlocklistConfigFromUnified(); err == nil {
		showBlocked = cfg.ShowBlockedInSearch
	}

	// Create main menu
	visibilityLabel := "Off (hidden from search)"
	if showBlocked {
		visibilityLabel = "On (shown in search)"
	}
	mainMenuItems := []components.MenuItem{
		components.NewMenuItem("View Blocked Stations", "Manage individually blocked stations", "1"),
		components.NewMenuItem("Manage Block Rules", "Block by country/language/tag", "2"),
		components.NewMenuItem("Import/Export Blocklist", "Backup and restore blocklist", "3"),
		components.NewMenuItem("Search Visibility", visibilityLabel, "4"),
	}
	mainMenu := components.CreateMenu(mainMenuItems, "📋 Block List Management", 80, 10)

	// Create rules submenu
	rulesMenuItems := []components.MenuItem{
		components.NewMenuItem("Block by Country", "Block all stations from specific countries", "1"),
		components.NewMenuItem("Block by Language", "Block all stations in specific languages", "2"),
		components.NewMenuItem("Block by Tag", "Block all stations with specific tags", "3"),
		components.NewMenuItem("View Active Rules", "See all block rules currently in effect", "4"),
	}
	rulesMenu := components.CreateMenu(rulesMenuItems, "🚫 Block Rules", 80, 10)

	// Create stations list
	delegate := createStyledDelegate()
	l := list.New([]list.Item{}, delegate, 80, 20)
	l.Title = "🚫 Blocked Stations"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.SetShowHelp(false)
	l.SetShowPagination(true)
	l.Styles.Title = listTitleStyle()
	l.Styles.PaginationStyle = paginationStyle()

	// Create text input for rules
	ti := textinput.New()
	ti.Placeholder = "Enter value..."
	ti.CharLimit = 50
	ti.Width = 50

	return BlocklistModel{
		state:               blocklistMainMenu,
		manager:             manager,
		mainMenu:            mainMenu,
		rulesMenu:           rulesMenu,
		listModel:           l,
		textInput:           ti,
		showBlockedInSearch: showBlocked,
	}
}

// Init initializes the blocklist screen
func (m BlocklistModel) Init() tea.Cmd {
	return m.loadBlockedStations()
}

// loadBlockedStations loads all blocked stations into the list
func (m BlocklistModel) loadBlockedStations() tea.Cmd {
	return func() tea.Msg {
		stations := m.manager.GetAll()
		return blocklistLoadedMsg{stations}
	}
}

// Update handles messages for the blocklist screen
func (m BlocklistModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Decrement message timer
	if m.messageTime > 0 {
		m.messageTime--
		if m.messageTime == 0 {
			m.message = ""
		}
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKeyPress(msg)

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		// listHeight accounts for: ASCII header (~8 lines) + page chrome
		// (blank line + title + subtitle + help = ~4 lines) + bottom padding (~2)
		// Use 14 to match the same overhead used in play.go.
		listHeight := msg.Height - 14
		if listHeight < 5 {
			listHeight = 5
		}
		// Update all list sizes
		m.mainMenu.SetSize(msg.Width-4, listHeight)
		m.rulesMenu.SetSize(msg.Width-4, listHeight)
		m.listModel.SetSize(msg.Width-4, listHeight)
		m.rulesListModel.SetSize(msg.Width-4, listHeight)
		return m, nil

	case blocklistLoadedMsg:
		m.stations = msg.stations
		items := make([]list.Item, len(m.stations))
		for i, station := range m.stations {
			items[i] = blocklistItem{station: station}
		}
		m.listModel.SetItems(items)
		if len(items) > 0 {
			m.listModel.Select(0)
		}
		return m, nil

	case blocklistUnblockedMsg:
		m.message = fmt.Sprintf("✓ Unblocked: %s", msg.stationName)
		m.messageTime = 180 // 3 seconds (at ~60fps)
		return m, m.loadBlockedStations()

	case blocklistClearedMsg:
		m.message = "✓ Cleared all blocked stations"
		m.messageTime = 180 // 3 seconds (at ~60fps)
		return m, m.loadBlockedStations()

	case blockRuleAddedMsg:
		m.message = fmt.Sprintf("✓ Added rule: %s = %s", msg.ruleType, msg.value)
		m.messageTime = 180 // 3 seconds (at ~60fps)
		m.state = blocklistRulesMenu
		m.textInput.Blur()
		return m, nil

	case blockRuleErrorMsg:
		m.message = fmt.Sprintf("✗ %v", msg.err)
		m.messageTime = 180 // 3 seconds (at ~60fps)
		return m, nil

	case blockRulesLoadedMsg:
		m.rules = msg.rules
		m.rulesListModel = createRulesListModel(msg.rules)
		if m.width > 0 && m.height > 0 {
			listHeight := m.height - 14
			if listHeight < 5 {
				listHeight = 5
			}
			m.rulesListModel.SetSize(m.width-4, listHeight)
		}
		if len(msg.rules) > 0 {
			m.rulesListModel.Select(0)
		}
		return m, nil

	case blockRuleDeletedMsg:
		m.message = fmt.Sprintf("✓ Deleted rule: %s", msg.rule.String())
		m.messageTime = 180 // 3 seconds (at ~60fps)
		m.state = blocklistViewRules
		return m, m.loadBlockRules()

	case blocklistExportedMsg:
		m.message = fmt.Sprintf("✓ Exported to: %s", msg.path)
		m.messageTime = 200
		m.state = blocklistMainMenu
		return m, nil

	case blocklistImportedMsg:
		m.message = fmt.Sprintf("✓ Imported %d rules and %d stations", msg.rulesCount, msg.stationsCount)
		m.messageTime = 200
		m.state = blocklistMainMenu
		return m, m.loadBlockedStations()

	case errMsg:
		m.err = msg.err
		m.message = fmt.Sprintf("✗ %v", msg.err)
		m.messageTime = 180 // 3 seconds (at ~60fps)
		return m, nil
	}

	var cmd tea.Cmd
	m.listModel, cmd = m.listModel.Update(msg)
	return m, cmd
}

// handleKeyPress handles keyboard input for blocklist
func (m BlocklistModel) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.state {
	case blocklistMainMenu:
		return m.handleMainMenuInput(msg)

	case blocklistViewStations:
		return m.handleViewStationsInput(msg)

	case blocklistConfirmClear:
		return m.handleConfirmClearInput(msg)

	case blocklistRulesMenu:
		return m.handleRulesMenuInput(msg)

	case blocklistBlockByCountry:
		return m.handleBlockByCountryInput(msg)

	case blocklistBlockByLanguage:
		return m.handleBlockByLanguageInput(msg)

	case blocklistBlockByTag:
		return m.handleBlockByTagInput(msg)

	case blocklistViewRules:
		return m.handleViewRulesInput(msg)

	case blocklistConfirmDeleteRule:
		return m.handleConfirmDeleteRuleInput(msg)

	case blocklistConfirmAddRule:
		return m.handleConfirmAddRuleInput(msg)

	case blocklistImportExport:
		// Placeholder view - allow back to main menu
		if msg.String() == "esc" {
			m.state = blocklistMainMenu
			return m, nil
		}
		return m, nil

	case blocklistSearchVisibility:
		return m.handleSearchVisibilityInput(msg)
	}

	return m, nil
}

// handleMainMenuInput handles input on the main blocklist menu
func (m BlocklistModel) handleMainMenuInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "m":
		// Return to main menu
		return m, func() tea.Msg {
			return navigateMsg{screen: screenMainMenu}
		}
	case "q":
		return m, tea.Quit
	case "enter":
		// Get selected item index
		idx := m.mainMenu.Index()
		return m.executeMainMenuAction(idx)
	case "1":
		return m.executeMainMenuAction(0)
	case "2":
		return m.executeMainMenuAction(1)
	case "3":
		return m.executeMainMenuAction(2)
	case "4":
		return m.executeMainMenuAction(3)
	}

	var cmd tea.Cmd
	m.mainMenu, cmd = m.mainMenu.Update(msg)
	return m, cmd
}

// executeMainMenuAction executes the selected menu action
func (m BlocklistModel) executeMainMenuAction(index int) (tea.Model, tea.Cmd) {
	switch index {
	case 0: // View Blocked Stations
		m.state = blocklistViewStations
		return m, m.loadBlockedStations()
	case 1: // Manage Block Rules
		m.state = blocklistRulesMenu
		return m, nil
	case 2: // Import/Export
		m.state = blocklistImportExport
		return m, nil
	case 3: // Search Visibility
		m.state = blocklistSearchVisibility
		return m, nil
	}
	return m, nil
}

// handleRulesMenuInput handles input on the rules submenu
func (m BlocklistModel) handleRulesMenuInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		// Return to main blocklist menu
		m.state = blocklistMainMenu
		return m, nil
	case "q":
		return m, tea.Quit
	case "enter":
		// Get selected item index
		idx := m.rulesMenu.Index()
		return m.executeRulesMenuAction(idx)
	case "1":
		return m.executeRulesMenuAction(0)
	case "2":
		return m.executeRulesMenuAction(1)
	case "3":
		return m.executeRulesMenuAction(2)
	case "4":
		return m.executeRulesMenuAction(3)
	}

	var cmd tea.Cmd
	m.rulesMenu, cmd = m.rulesMenu.Update(msg)
	return m, cmd
}

// executeRulesMenuAction executes the selected rules menu action
func (m BlocklistModel) executeRulesMenuAction(index int) (tea.Model, tea.Cmd) {
	switch index {
	case 0: // Block by Country
		m.state = blocklistBlockByCountry
		m.textInput.Reset()
		m.textInput.Placeholder = "Enter country name or code (e.g., US, United States)..."
		m.textInput.Focus()
		return m, textinput.Blink
	case 1: // Block by Language
		m.state = blocklistBlockByLanguage
		m.textInput.Reset()
		m.textInput.Placeholder = "Enter language (e.g., english, arabic, spanish)..."
		m.textInput.Focus()
		return m, textinput.Blink
	case 2: // Block by Tag
		m.state = blocklistBlockByTag
		m.textInput.Reset()
		m.textInput.Placeholder = "Enter tag (e.g., jazz, sports, news)..."
		m.textInput.Focus()
		return m, textinput.Blink
	case 3: // View Active Rules
		m.state = blocklistViewRules
		return m, m.loadBlockRules()
	}
	return m, nil
}

// handleViewStationsInput handles input when viewing blocked stations
func (m BlocklistModel) handleViewStationsInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		// Return to main blocklist menu
		m.state = blocklistMainMenu
		return m, nil
	case "q":
		return m, tea.Quit

	case "u":
		// Unblock selected station
		if len(m.stations) == 0 {
			m.message = "No blocked stations"
			m.messageTime = 150
			return m, nil
		}

		selected := m.listModel.Index()
		if selected < 0 || selected >= len(m.stations) {
			return m, nil
		}

		station := m.stations[selected]
		return m, m.unblockStation(station)

	case "c":
		// Confirm clear all
		if len(m.stations) == 0 {
			m.message = "No blocked stations to clear"
			m.messageTime = 150
			return m, nil
		}
		m.state = blocklistConfirmClear
		m.message = ""
		return m, nil

	case "?":
		// Show help
		return m, nil

	default:
		// Pass navigation keys to list model (up/down, j/k, etc.)
		var cmd tea.Cmd
		m.listModel, cmd = m.listModel.Update(msg)
		return m, cmd
	}
}

// handleConfirmClearInput handles input during clear confirmation
func (m BlocklistModel) handleConfirmClearInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y", "Y":
		// Confirm clear
		m.state = blocklistViewStations
		return m, m.clearAll()

	case "n", "N", "esc":
		// Cancel
		m.state = blocklistViewStations
		m.message = "Clear cancelled"
		m.messageTime = 150
		return m, nil
	}
	return m, nil
}

// unblockStation removes a station from the blocklist
func (m BlocklistModel) unblockStation(station blocklist.BlockedStation) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		if err := m.manager.Unblock(ctx, station.StationUUID); err != nil {
			return errMsg{err}
		}
		return blocklistUnblockedMsg{stationName: station.Name}
	}
}

// clearAll clears all blocked stations
func (m BlocklistModel) clearAll() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		if err := m.manager.Clear(ctx); err != nil {
			return errMsg{err}
		}
		return blocklistClearedMsg{}
	}
}

// View renders the blocklist screen
func (m BlocklistModel) View() string {
	switch m.state {
	case blocklistMainMenu:
		return m.viewMainMenu()

	case blocklistViewStations:
		return m.viewBlockedStations()

	case blocklistConfirmClear:
		return m.viewConfirmClear()

	case blocklistRulesMenu:
		return m.viewRulesMenu()

	case blocklistBlockByCountry:
		return m.viewBlockByCountry()

	case blocklistBlockByLanguage:
		return m.viewBlockByLanguage()

	case blocklistBlockByTag:
		return m.viewBlockByTag()

	case blocklistViewRules:
		return m.viewActiveRules()

	case blocklistConfirmDeleteRule:
		return m.viewConfirmDeleteRule()

	case blocklistConfirmAddRule:
		return m.viewConfirmAddRule()

	case blocklistImportExport:
		return m.viewPlaceholder("Import/Export Blocklist")
	case blocklistSearchVisibility:
		return m.viewSearchVisibility()
	}

	return ""
}

// viewMainMenu renders the main blocklist menu
func (m BlocklistModel) viewMainMenu() string {
	var content strings.Builder

	if m.message != "" {
		style := successStyle()
		if strings.Contains(m.message, "✗") {
			style = errorStyle()
		}
		content.WriteString(style.Render(m.message))
		content.WriteString("\n\n")
	}

	content.WriteString(m.mainMenu.View())

	return RenderPageWithBottomHelp(PageLayout{
		Content: content.String(),
		Help:    "↑↓/jk: Navigate • Enter: Select • 1-4: Quick select • Esc: Back • Ctrl+C: Quit",
	}, m.height)
}

// handleSearchVisibilityInput handles input on the search visibility screen
func (m BlocklistModel) handleSearchVisibilityInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.state = blocklistMainMenu
		return m, nil
	case "y", "1":
		return m.setSearchVisibility(true)
	case "n", "2":
		return m.setSearchVisibility(false)
	}
	return m, nil
}

// setSearchVisibility saves the setting and refreshes the menu label
func (m BlocklistModel) setSearchVisibility(show bool) (tea.Model, tea.Cmd) {
	if err := storage.SaveBlocklistConfigToUnified(config.BlocklistConfig{
		ShowBlockedInSearch: show,
	}); err != nil {
		m.message = fmt.Sprintf("✗ Failed to save: %v", err)
		m.messageTime = 180
		m.state = blocklistMainMenu
		return m, nil
	} else {
		m.showBlockedInSearch = show
		if show {
			m.message = "✓ Blocked stations will appear in search (marked 🚫)"
		} else {
			m.message = "✓ Blocked stations will be hidden from search"
		}
	}
	m.messageTime = 180
	// Refresh the menu item description to reflect new value
	items := m.mainMenu.Items()
	if len(items) >= 4 {
		visibilityLabel := "Off (hidden from search)"
		if show {
			visibilityLabel = "On (shown in search)"
		}
		items[3] = components.NewMenuItem("Search Visibility", visibilityLabel, "4")
		m.mainMenu.SetItems(items)
	}
	m.state = blocklistMainMenu
	return m, nil
}

// viewSearchVisibility renders the search visibility toggle screen
func (m BlocklistModel) viewSearchVisibility() string {
	var content strings.Builder

	content.WriteString("Control whether blocked stations appear in search results.\n\n")

	current := "Hidden from search results (default)"
	if m.showBlockedInSearch {
		current = "Shown in search results (marked 🚫)"
	}
	content.WriteString(stationFieldStyle().Render("Current setting: "))
	content.WriteString(highlightStyle().Render(current))
	content.WriteString("\n\n")
	content.WriteString(helpStyle().Render("─────────────────────────────────────────"))
	content.WriteString("\n\n")

	if m.showBlockedInSearch {
		content.WriteString("  1) y — Keep showing blocked stations (🚫 prefix)\n")
		content.WriteString("  2) n — Hide blocked stations from search\n")
	} else {
		content.WriteString("  1) y — Show blocked stations in search (🚫 prefix)\n")
		content.WriteString("  2) n — Keep hiding blocked stations from search\n")
	}

	return RenderPageWithBottomHelp(PageLayout{
		Title:   "🔍 Search Visibility",
		Content: content.String(),
		Help:    "y/1: Yes • n/2: No • Esc: Back",
	}, m.height)
}

// viewRulesMenu renders the rules submenu
func (m BlocklistModel) viewRulesMenu() string {
	var content strings.Builder

	if m.message != "" {
		style := successStyle()
		if strings.Contains(m.message, "✗") {
			style = errorStyle()
		}
		content.WriteString(style.Render(m.message))
		content.WriteString("\n\n")
	}

	content.WriteString(m.rulesMenu.View())

	return RenderPageWithBottomHelp(PageLayout{
		Content: content.String(),
		Help:    "↑↓/jk: Navigate • Enter: Select • 1-4: Quick select • Esc: Back • Ctrl+C: Quit",
	}, m.height)
}

// viewBlockedStations renders the blocked stations list
func (m BlocklistModel) viewBlockedStations() string {
	count := m.manager.Count()
	subtitle := fmt.Sprintf("%d station(s) blocked", count)

	var content strings.Builder

	if m.message != "" {
		style := successStyle()
		if strings.Contains(m.message, "✗") {
			style = errorStyle()
		}
		content.WriteString(style.Render(m.message))
		content.WriteString("\n\n")
	}

	// Show list or empty message
	if count == 0 {
		content.WriteString(infoStyle().Render("No blocked stations yet.\n\nPress 'b' while playing a station to block it."))
	} else {
		content.WriteString(m.listModel.View())
	}

	return RenderPageWithBottomHelp(PageLayout{
		Title:    "Blocked Stations",
		Subtitle: subtitle,
		Content:  content.String(),
		Help:     "↑↓/jk: Navigate • u: Unblock • c: Clear all • Esc: Back • ?: Help",
	}, m.height)
}

// viewConfirmClear renders the clear confirmation dialog
func (m BlocklistModel) viewConfirmClear() string {
	count := m.manager.Count()
	var content strings.Builder

	fmt.Fprintf(&content, "Clear all %d blocked stations?\n\n", count)
	content.WriteString(errorStyle().Render("⚠ This cannot be undone!"))

	return RenderPageWithBottomHelp(PageLayout{
		Title:   "Confirm Clear",
		Content: content.String(),
		Help:    "y: Yes, clear all • n/Esc: No, cancel",
	}, m.height)
}

// viewPlaceholder renders a placeholder view for features not yet implemented
func (m BlocklistModel) viewPlaceholder(title string) string {
	var content strings.Builder

	content.WriteString(infoStyle().Render("🚧 Coming Soon"))
	content.WriteString("\n\n")
	content.WriteString("This feature is under development.")

	if m.message != "" {
		content.WriteString("\n\n")
		content.WriteString(infoStyle().Render(m.message))
	}

	return RenderPageWithBottomHelp(PageLayout{
		Title:   title,
		Content: content.String(),
		Help:    "Esc: Back",
	}, m.height)
}

// handleBlockRuleInput is a shared helper for blocking by country, language, or tag
func (m BlocklistModel) handleBlockRuleInput(msg tea.KeyMsg, ruleType blocklist.BlockRuleType, emptyMsg string) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.state = blocklistRulesMenu
		m.textInput.Blur()
		return m, nil
	case "enter":
		value := strings.TrimSpace(m.textInput.Value())
		if value == "" {
			m.message = emptyMsg
			m.messageTime = 150
			return m, nil
		}
		newM, cmd := m.addBlockRuleWithConfirmation(ruleType, value)
		return newM, cmd
	}

	var cmd tea.Cmd
	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

// handleBlockByCountryInput handles input for blocking by country
func (m BlocklistModel) handleBlockByCountryInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	return m.handleBlockRuleInput(msg, blocklist.BlockRuleCountry, "Country cannot be empty")
}

// handleBlockByLanguageInput handles input for blocking by language
func (m BlocklistModel) handleBlockByLanguageInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	return m.handleBlockRuleInput(msg, blocklist.BlockRuleLanguage, "Language cannot be empty")
}

// handleBlockByTagInput handles input for blocking by tag
func (m BlocklistModel) handleBlockByTagInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	return m.handleBlockRuleInput(msg, blocklist.BlockRuleTag, "Tag cannot be empty")
}

// handleViewRulesInput handles input when viewing active rules
func (m BlocklistModel) handleViewRulesInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.state = blocklistRulesMenu
		return m, nil
	case "d":
		// Delete selected rule
		if len(m.rules) == 0 {
			m.message = "No rules to delete"
			m.messageTime = 150
			return m, nil
		}
		selected := m.rulesListModel.Index()
		if selected < 0 || selected >= len(m.rules) {
			return m, nil
		}
		m.selectedRuleIndex = selected
		m.state = blocklistConfirmDeleteRule
		return m, nil
	default:
		// Pass navigation keys to list model
		var cmd tea.Cmd
		m.rulesListModel, cmd = m.rulesListModel.Update(msg)
		return m, cmd
	}
}

// handleConfirmDeleteRuleInput handles input during rule deletion confirmation
func (m BlocklistModel) handleConfirmDeleteRuleInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y", "Y":
		// Confirm deletion
		if m.selectedRuleIndex >= 0 && m.selectedRuleIndex < len(m.rules) {
			rule := m.rules[m.selectedRuleIndex]
			return m, m.deleteBlockRule(rule)
		}
		m.state = blocklistViewRules
		return m, nil
	case "n", "N", "esc":
		// Cancel deletion
		m.state = blocklistViewRules
		m.message = "Deletion cancelled"
		m.messageTime = 150
		return m, nil
	}
	return m, nil
}

// handleConfirmAddRuleInput handles input during rule addition confirmation
func (m BlocklistModel) handleConfirmAddRuleInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y", "Y":
		// Confirm addition
		return m, m.confirmAddBlockRule()
	case "n", "N", "esc":
		// Cancel addition
		m.state = m.previousState
		m.textInput.Focus()
		m.message = "Rule addition cancelled"
		m.messageTime = 150
		return m, textinput.Blink
	}
	return m, nil
}

// viewBlockByCountry renders the block by country input view
func (m BlocklistModel) viewBlockByCountry() string {
	var content strings.Builder

	content.WriteString("Enter a country name or 2-letter country code.\n")
	content.WriteString("Examples: US, United States, FR, France\n\n")

	if m.message != "" {
		style := errorStyle()
		if strings.Contains(m.message, "✓") {
			style = successStyle()
		}
		content.WriteString(style.Render(m.message))
		content.WriteString("\n\n")
	}

	content.WriteString(m.textInput.View())

	return RenderPageWithBottomHelp(PageLayout{
		Title:   "Block by Country",
		Content: content.String(),
		Help:    "Enter: Add rule • Esc: Back",
	}, m.height)
}

// viewBlockByLanguage renders the block by language input view
func (m BlocklistModel) viewBlockByLanguage() string {
	var content strings.Builder

	content.WriteString("Enter a language name (case-insensitive).\n")
	content.WriteString("Examples: english, arabic, spanish, french\n\n")

	if m.message != "" {
		style := errorStyle()
		if strings.Contains(m.message, "✓") {
			style = successStyle()
		}
		content.WriteString(style.Render(m.message))
		content.WriteString("\n\n")
	}

	content.WriteString(m.textInput.View())

	return RenderPageWithBottomHelp(PageLayout{
		Title:   "Block by Language",
		Content: content.String(),
		Help:    "Enter: Add rule • Esc: Back",
	}, m.height)
}

// viewBlockByTag renders the block by tag input view
func (m BlocklistModel) viewBlockByTag() string {
	var content strings.Builder

	content.WriteString("Enter a tag/genre name (case-insensitive).\n")
	content.WriteString("Examples: jazz, rock, news, sports, classical\n\n")

	if m.message != "" {
		style := errorStyle()
		if strings.Contains(m.message, "✓") {
			style = successStyle()
		}
		content.WriteString(style.Render(m.message))
		content.WriteString("\n\n")
	}

	content.WriteString(m.textInput.View())

	return RenderPageWithBottomHelp(PageLayout{
		Title:   "Block by Tag",
		Content: content.String(),
		Help:    "Enter: Add rule • Esc: Back",
	}, m.height)
}

// viewActiveRules renders the active rules list
func (m BlocklistModel) viewActiveRules() string {
	// Use Subtitle slot for transient messages so the list starts at a
	// consistent vertical position and the header is never clipped.
	subtitle := ""
	if m.message != "" {
		if strings.Contains(m.message, "✗") {
			subtitle = errorStyle().Render(m.message)
		} else {
			subtitle = successStyle().Render(m.message)
		}
	}

	var content strings.Builder
	if len(m.rules) == 0 {
		content.WriteString(infoStyle().Render("No block rules defined yet.\n\n"))
		content.WriteString("Use the Block Rules menu to add rules.")
	} else {
		// Use the interactive list (title is rendered by PageLayout.Title)
		content.WriteString(m.rulesListModel.View())
	}

	return RenderPageWithBottomHelp(PageLayout{
		Title:    "🚫 Active Block Rules",
		Subtitle: subtitle,
		Content:  content.String(),
		Help:     "↑↓/jk: Navigate • d: Delete rule • Esc: Back",
	}, m.height)
}

// viewConfirmDeleteRule renders the delete rule confirmation dialog
func (m BlocklistModel) viewConfirmDeleteRule() string {
	var content strings.Builder

	if m.selectedRuleIndex >= 0 && m.selectedRuleIndex < len(m.rules) {
		rule := m.rules[m.selectedRuleIndex]
		content.WriteString("Delete this blocking rule?\n\n")
		fmt.Fprintf(&content, "Rule: %s\n\n", rule.String())
		content.WriteString(errorStyle().Render("⚠ This will allow matching stations to appear again!"))
	} else {
		content.WriteString("No rule selected")
	}

	return RenderPageWithBottomHelp(PageLayout{
		Title:   "Confirm Delete Rule",
		Content: content.String(),
		Help:    "y: Yes, delete • n/Esc: No, cancel",
	}, m.height)
}

// viewConfirmAddRule renders the add rule confirmation dialog
func (m BlocklistModel) viewConfirmAddRule() string {
	var content strings.Builder

	content.WriteString("Add this blocking rule?\n\n")
	fmt.Fprintf(&content, "Type: %s\n", m.pendingRuleType)
	fmt.Fprintf(&content, "Value: %s\n\n", m.pendingRuleValue)

	// Add description based on type
	switch m.pendingRuleType {
	case blocklist.BlockRuleCountry:
		content.WriteString(infoStyle().Render("This will block all stations from this country."))
	case blocklist.BlockRuleLanguage:
		content.WriteString(infoStyle().Render("This will block all stations in this language."))
	case blocklist.BlockRuleTag:
		content.WriteString(infoStyle().Render("This will block all stations with this tag/genre."))
	}

	return RenderPageWithBottomHelp(PageLayout{
		Title:   "Confirm Add Rule",
		Content: content.String(),
		Help:    "y: Yes, add rule • n/Esc: No, cancel",
	}, m.height)
}

// Messages for blocklist operations
type blocklistLoadedMsg struct {
	stations []blocklist.BlockedStation
}

type blocklistUnblockedMsg struct {
	stationName string
}

type blocklistClearedMsg struct{}

type blockRuleAddedMsg struct {
	ruleType blocklist.BlockRuleType
	value    string
}

type blockRuleErrorMsg struct {
	err error
}
