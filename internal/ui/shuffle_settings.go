package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/shinokada/tera/v2/internal/storage"
	"github.com/shinokada/tera/v2/internal/theme"
	"github.com/shinokada/tera/v2/internal/ui/components"
)

// shuffleSettingsState represents the current state in shuffle settings
type shuffleSettingsState int

const (
	shuffleSettingsMenu shuffleSettingsState = iota
	shuffleSettingsInterval
	shuffleSettingsHistorySize
)

// ShuffleSettingsModel represents the shuffle settings page
type ShuffleSettingsModel struct {
	state            shuffleSettingsState
	config           storage.ShuffleConfig
	menuList         list.Model
	intervalList     list.Model
	historyList      list.Model
	width            int
	height           int
	message          string
	messageIsSuccess bool
	messageTime      int
}

// NewShuffleSettingsModel creates a new shuffle settings model
func NewShuffleSettingsModel() ShuffleSettingsModel {
	// Load current config
	config, err := storage.LoadShuffleConfig()
	if err != nil {
		config = storage.DefaultShuffleConfig()
	}

	m := ShuffleSettingsModel{
		state:  shuffleSettingsMenu,
		config: config,
		width:  80,
		height: 24,
	}

	m.rebuildMenuList()
	m.buildIntervalList()
	m.buildHistoryList()

	return m
}

// Init initializes the shuffle settings model
func (m ShuffleSettingsModel) Init() tea.Cmd {
	return tickEverySecond()
}

// Update handles messages for shuffle settings
func (m ShuffleSettingsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch m.state {
		case shuffleSettingsMenu:
			return m.updateMenu(msg)
		case shuffleSettingsInterval:
			return m.updateInterval(msg)
		case shuffleSettingsHistorySize:
			return m.updateHistorySize(msg)
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tickMsg:
		// Countdown message
		if m.messageTime > 0 {
			m.messageTime--
			if m.messageTime == 0 {
				m.message = ""
			}
		}
		return m, tickEverySecond()
	}

	return m, nil
}

// updateMenu handles menu navigation
func (m ShuffleSettingsModel) updateMenu(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	// Handle escape/back
	if key == "esc" {
		return m, func() tea.Msg {
			return navigateMsg{screen: screenSettings}
		}
	}
	if key == "0" {
		return m, func() tea.Msg {
			return navigateMsg{screen: screenMainMenu}
		}
	}

	// Handle ctrl+c
	if key == "ctrl+c" {
		return m, tea.Quit
	}

	// Handle menu selection
	newList, selected := components.HandleMenuKey(msg, m.menuList)
	m.menuList = newList

	if selected >= 0 {
		switch selected {
		case 0: // Toggle Auto-advance
			m.config.AutoAdvance = !m.config.AutoAdvance
			m.saveConfig()
			m.rebuildMenuList()
		case 1: // Set Auto-advance Interval
			m.state = shuffleSettingsInterval
		case 2: // Toggle History
			m.config.RememberHistory = !m.config.RememberHistory
			m.saveConfig()
			m.rebuildMenuList()
		case 3: // Set History Size
			m.state = shuffleSettingsHistorySize
		case 4: // Reset to Defaults
			m.config = storage.DefaultShuffleConfig()
			m.saveConfig()
			m.rebuildMenuList()
			m.message = "✓ Reset to default settings"
			m.messageIsSuccess = true
			m.messageTime = 3 // 3 seconds (decremented once per second via tickMsg)
		case 5: // Back to Settings
			return m, func() tea.Msg {
				return navigateMsg{screen: screenSettings}
			}
		}
	}

	// Handle number shortcuts
	if key >= "1" && key <= "6" {
		num := int(key[0] - '0')
		m.menuList.Select(num - 1)
		newModel, cmd := m.updateMenu(tea.KeyMsg{Type: tea.KeyEnter})
		return newModel, cmd
	}

	return m, nil
}

// updateInterval handles interval selection
func (m ShuffleSettingsModel) updateInterval(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	// Handle escape/back
	if key == "esc" {
		m.state = shuffleSettingsMenu
		return m, nil
	}

	if key == "0" {
		return m, func() tea.Msg {
			return navigateMsg{screen: screenMainMenu}
		}
	}

	// Handle ctrl+c
	if key == "ctrl+c" {
		return m, tea.Quit
	}

	// Handle selection
	newList, selected := components.HandleMenuKey(msg, m.intervalList)
	m.intervalList = newList

	if selected >= 0 {
		intervals := []int{1, 3, 5, 10, 15}
		if selected < len(intervals) {
			m.config.IntervalMinutes = intervals[selected]
			m.saveConfig()
			m.rebuildMenuList()
			m.buildIntervalList()
			m.state = shuffleSettingsMenu
			m.message = fmt.Sprintf("✓ Auto-advance interval set to %d minutes", m.config.IntervalMinutes)
			m.messageIsSuccess = true
			m.messageTime = 3 // 3 seconds (decremented once per second via tickMsg)
		} else if selected == len(intervals) {
			// Back option
			m.state = shuffleSettingsMenu
		}
	}

	// Handle number shortcuts
	if key >= "1" && key <= "6" {
		num := int(key[0] - '0')
		m.intervalList.Select(num - 1)
		newModel, cmd := m.updateInterval(tea.KeyMsg{Type: tea.KeyEnter})
		return newModel, cmd
	}

	return m, nil
}

// updateHistorySize handles history size selection
func (m ShuffleSettingsModel) updateHistorySize(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	// Handle escape/back
	if key == "esc" {
		m.state = shuffleSettingsMenu
		return m, nil
	}

	if key == "0" {
		return m, func() tea.Msg {
			return navigateMsg{screen: screenMainMenu}
		}
	}

	// Handle ctrl+c
	if key == "ctrl+c" {
		return m, tea.Quit
	}

	// Handle selection
	newList, selected := components.HandleMenuKey(msg, m.historyList)
	m.historyList = newList

	if selected >= 0 {
		historySizes := []int{3, 5, 7, 10}
		if selected < len(historySizes) {
			m.config.MaxHistory = historySizes[selected]
			m.saveConfig()
			m.rebuildMenuList()
			m.buildHistoryList()
			m.state = shuffleSettingsMenu
			m.message = fmt.Sprintf("✓ Shuffle history size set to %d stations", m.config.MaxHistory)
			m.messageIsSuccess = true
			m.messageTime = 3 // 3 seconds (decremented once per second via tickMsg)
		} else if selected == len(historySizes) {
			// Back option
			m.state = shuffleSettingsMenu
		}
	}

	// Handle number shortcuts
	if key >= "1" && key <= "5" {
		num := int(key[0] - '0')
		m.historyList.Select(num - 1)
		newModel, cmd := m.updateHistorySize(tea.KeyMsg{Type: tea.KeyEnter})
		return newModel, cmd
	}

	return m, nil
}

// saveConfig saves the current configuration
func (m *ShuffleSettingsModel) saveConfig() {
	if err := storage.SaveShuffleConfig(m.config); err != nil {
		m.message = fmt.Sprintf("✗ Failed to save: %v", err)
		m.messageIsSuccess = false
		m.messageTime = 3 // 3 seconds (decremented once per second via tickMsg)
	}
}

// rebuildMenuList rebuilds the main menu list
func (m *ShuffleSettingsModel) rebuildMenuList() {
	menuItems := []components.MenuItem{
		components.NewMenuItem(
			fmt.Sprintf("Toggle Auto-advance (%s)", boolToOnOff(m.config.AutoAdvance)),
			"Automatically skip to next shuffle station",
			"1",
		),
		components.NewMenuItem(
			fmt.Sprintf("Set Auto-advance Interval (%d min)", m.config.IntervalMinutes),
			"How long to play each station",
			"2",
		),
		components.NewMenuItem(
			fmt.Sprintf("Toggle History (%s)", boolToOnOff(m.config.RememberHistory)),
			"Remember previous shuffle stations",
			"3",
		),
		components.NewMenuItem(
			fmt.Sprintf("Set History Size (%d stations)", m.config.MaxHistory),
			"Number of previous stations to remember",
			"4",
		),
		components.NewMenuItem(
			"Reset to Defaults",
			"Restore default shuffle settings",
			"5",
		),
		components.NewMenuItem(
			"Back to Settings",
			"",
			"6",
		),
	}

	m.menuList = components.CreateMenu(menuItems, "", 60, len(menuItems)+2)
}

// buildIntervalList builds the interval selection list
func (m *ShuffleSettingsModel) buildIntervalList() {
	intervals := []struct {
		minutes int
		label   string
	}{
		{1, "1 minute"},
		{3, "3 minutes"},
		{5, "5 minutes"},
		{10, "10 minutes"},
		{15, "15 minutes"},
	}

	menuItems := []components.MenuItem{}
	for i, interval := range intervals {
		shortcut := fmt.Sprintf("%d", i+1)
		desc := ""
		if interval.minutes == m.config.IntervalMinutes {
			desc = "← Current"
		}
		menuItems = append(menuItems, components.NewMenuItem(interval.label, desc, shortcut))
	}
	menuItems = append(menuItems, components.NewMenuItem("Back", "", "6"))

	m.intervalList = components.CreateMenu(menuItems, "", 50, len(menuItems)+2)
}

// buildHistoryList builds the history size selection list
func (m *ShuffleSettingsModel) buildHistoryList() {
	sizes := []struct {
		size  int
		label string
	}{
		{3, "3 stations (Minimal)"},
		{5, "5 stations (Default)"},
		{7, "7 stations"},
		{10, "10 stations (Maximum)"},
	}

	menuItems := []components.MenuItem{}
	for i, size := range sizes {
		shortcut := fmt.Sprintf("%d", i+1)
		desc := ""
		if size.size == m.config.MaxHistory {
			desc = "← Current"
		}
		menuItems = append(menuItems, components.NewMenuItem(size.label, desc, shortcut))
	}
	menuItems = append(menuItems, components.NewMenuItem("Back", "", "5"))

	m.historyList = components.CreateMenu(menuItems, "", 50, len(menuItems)+2)
}

// View renders the shuffle settings screen
func (m ShuffleSettingsModel) View() string {
	switch m.state {
	case shuffleSettingsMenu:
		return m.viewMenu()
	case shuffleSettingsInterval:
		return m.viewInterval()
	case shuffleSettingsHistorySize:
		return m.viewHistorySize()
	}
	return "Unknown state"
}

// viewMenu renders the main menu
func (m ShuffleSettingsModel) viewMenu() string {
	var content strings.Builder

	t := theme.Current()
	titleStyle := lipgloss.NewStyle().
		Foreground(t.HighlightColor()).
		Bold(true).
		PaddingLeft(t.Padding.ListItemLeft)

	// Title
	content.WriteString(titleStyle.Render("⚙️  Settings > Shuffle Settings"))
	content.WriteString("\n\n")

	// Current settings summary
	content.WriteString(subtitleStyle().Render("Current Settings:"))
	content.WriteString("\n\n")
	fmt.Fprintf(&content, "  Auto-advance:           %s\n", boolToEnabledDisabled(m.config.AutoAdvance))
	fmt.Fprintf(&content, "  Auto-advance interval:  %d minutes\n", m.config.IntervalMinutes)
	fmt.Fprintf(&content, "  Remember history:       %s (Last %d stations)\n", boolToEnabledDisabled(m.config.RememberHistory), m.config.MaxHistory)
	content.WriteString("\n")

	// Menu
	content.WriteString(m.menuList.View())

	// Success message
	if m.message != "" {
		content.WriteString("\n\n")
		if m.messageIsSuccess {
			content.WriteString(successStyle().Render(m.message))
		} else {
			content.WriteString(errorStyle().Render(m.message))
		}
	} else {
		content.WriteString("\n\n")
		content.WriteString(infoStyle().Render("✓ Shuffle settings saved automatically"))
	}

	return RenderPageWithBottomHelp(PageLayout{
		Content: content.String(),
		Help:    "↑↓/jk: Navigate • Enter: Select • 1-6: Shortcut • Esc: Back • 0: Main Menu",
	}, m.height)
}

// viewInterval renders the interval selection screen
func (m ShuffleSettingsModel) viewInterval() string {
	var content strings.Builder

	t := theme.Current()
	titleStyle := lipgloss.NewStyle().
		Foreground(t.HighlightColor()).
		Bold(true).
		PaddingLeft(t.Padding.ListItemLeft)

	// Title
	content.WriteString(titleStyle.Render("⚙️  Settings > Shuffle Settings > Auto-advance Interval"))
	content.WriteString("\n\n")

	content.WriteString(subtitleStyle().Render("Select auto-advance interval:"))
	content.WriteString("\n\n")
	fmt.Fprintf(&content, "  Current: %d minutes\n", m.config.IntervalMinutes)
	content.WriteString("\n")

	// Interval list
	content.WriteString(m.intervalList.View())

	content.WriteString("\n\n")
	content.WriteString(infoStyle().Render("Auto-advance will skip to next shuffle station after this interval"))

	return RenderPageWithBottomHelp(PageLayout{
		Content: content.String(),
		Help:    "↑↓/jk: Navigate • Enter: Select • 1-6: Shortcut • Esc: Back • 0: Main Menu",
	}, m.height)
}

// viewHistorySize renders the history size selection screen
func (m ShuffleSettingsModel) viewHistorySize() string {
	var content strings.Builder

	t := theme.Current()
	titleStyle := lipgloss.NewStyle().
		Foreground(t.HighlightColor()).
		Bold(true).
		PaddingLeft(t.Padding.ListItemLeft)

	// Title
	content.WriteString(titleStyle.Render("⚙️  Settings > Shuffle Settings > History Size"))
	content.WriteString("\n\n")

	content.WriteString(subtitleStyle().Render("Select shuffle history size:"))
	content.WriteString("\n\n")
	fmt.Fprintf(&content, "  Current: %d stations\n", m.config.MaxHistory)
	content.WriteString("\n")

	// History size list
	content.WriteString(m.historyList.View())

	content.WriteString("\n\n")
	content.WriteString(infoStyle().Render("History allows you to go back to previous shuffle stations"))

	return RenderPageWithBottomHelp(PageLayout{
		Content: content.String(),
		Help:    "↑↓/jk: Navigate • Enter: Select • 1-5: Shortcut • Esc: Back • 0: Main Menu",
	}, m.height)
}

// Helper functions
func boolToOnOff(b bool) string {
	if b {
		return "On"
	}
	return "Off"
}

func boolToEnabledDisabled(b bool) string {
	if b {
		return "Enabled"
	}
	return "Disabled"
}
