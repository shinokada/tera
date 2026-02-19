package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/shinokada/tera/v3/internal/storage"
	"github.com/shinokada/tera/v3/internal/theme"
	"github.com/shinokada/tera/v3/internal/ui/components"
)

// Shared configuration for reconnect delay options
var reconnectDelayOptions = []struct {
	seconds int
	label   string
}{
	{1, "1 second (Fastest)"},
	{3, "3 seconds"},
	{5, "5 seconds (Default)"},
	{10, "10 seconds"},
	{15, "15 seconds"},
	{30, "30 seconds (Slowest)"},
}

// Shared configuration for stream buffer options
var streamBufferOptions = []struct {
	mb    int
	label string
}{
	{0, "No buffering (Original behavior)"},
	{10, "10 MB (Minimal)"},
	{25, "25 MB (Light)"},
	{50, "50 MB (Default)"},
	{100, "100 MB (Heavy)"},
	{150, "150 MB (Maximum)"},
	{200, "200 MB (Extreme)"},
}

// connectionSettingsState represents the current state in connection settings
type connectionSettingsState int

const (
	connectionSettingsMenu connectionSettingsState = iota
	connectionSettingsDelay
	connectionSettingsBuffer
)

// ConnectionSettingsModel represents the connection settings page
type ConnectionSettingsModel struct {
	state            connectionSettingsState
	config           storage.ConnectionConfig
	menuList         list.Model
	delayList        list.Model
	bufferList       list.Model
	width            int
	height           int
	message          string
	messageIsSuccess bool
	messageTime      int
}

// NewConnectionSettingsModel creates a new connection settings model
func NewConnectionSettingsModel() ConnectionSettingsModel {
	// Load current config
	config, err := storage.LoadConnectionConfig()
	if err != nil {
		config = storage.DefaultConnectionConfig()
	}

	m := ConnectionSettingsModel{
		state:  connectionSettingsMenu,
		config: config,
		width:  80,
		height: 24,
	}

	m.rebuildMenuList()
	m.buildDelayList()
	m.buildBufferList()

	return m
}

// Init initializes the connection settings model
func (m ConnectionSettingsModel) Init() tea.Cmd {
	return tickEverySecond()
}

// Update handles messages for connection settings
func (m ConnectionSettingsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch m.state {
		case connectionSettingsMenu:
			return m.updateMenu(msg)
		case connectionSettingsDelay:
			return m.updateDelay(msg)
		case connectionSettingsBuffer:
			return m.updateBuffer(msg)
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
func (m ConnectionSettingsModel) updateMenu(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
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
		case 0: // Toggle Auto-reconnect
			m.config.AutoReconnect = !m.config.AutoReconnect
			m.saveConfig()
			m.rebuildMenuList()
		case 1: // Set Reconnect Delay
			m.state = connectionSettingsDelay
		case 2: // Set Stream Buffer
			m.state = connectionSettingsBuffer
		case 3: // Reset to Defaults
			m.config = storage.DefaultConnectionConfig()
			m.saveConfig()
			m.rebuildMenuList()
			m.buildDelayList()
			m.buildBufferList()
			m.message = "✓ Reset to default settings"
			m.messageIsSuccess = true
			m.messageTime = 3 // 3 seconds (decremented once per second via tickMsg)
		case 4: // Back to Settings
			return m, func() tea.Msg {
				return navigateMsg{screen: screenSettings}
			}
		}
	}

	// Handle number shortcuts
	if key >= "1" && key <= "5" {
		num := int(key[0] - '0')
		m.menuList.Select(num - 1)
		newModel, cmd := m.updateMenu(tea.KeyMsg{Type: tea.KeyEnter})
		return newModel, cmd
	}

	return m, nil
}

// updateDelay handles reconnect delay selection
func (m ConnectionSettingsModel) updateDelay(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	// Handle escape/back
	if key == "esc" {
		m.state = connectionSettingsMenu
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
	newList, selected := components.HandleMenuKey(msg, m.delayList)
	m.delayList = newList

	if selected >= 0 {
		if selected < len(reconnectDelayOptions) {
			m.config.ReconnectDelay = reconnectDelayOptions[selected].seconds
			m.saveConfig()
			m.rebuildMenuList()
			m.buildDelayList()
			m.state = connectionSettingsMenu
			m.message = fmt.Sprintf("✓ Reconnect delay set to %d seconds", m.config.ReconnectDelay)
			m.messageIsSuccess = true
			m.messageTime = 3 // 3 seconds (decremented once per second via tickMsg)
		} else if selected == len(reconnectDelayOptions) {
			// Back option
			m.state = connectionSettingsMenu
		}
	}

	// Handle number shortcuts
	if key >= "1" && key <= "7" {
		num := int(key[0] - '0')
		m.delayList.Select(num - 1)
		newModel, cmd := m.updateDelay(tea.KeyMsg{Type: tea.KeyEnter})
		return newModel, cmd
	}

	return m, nil
}

// updateBuffer handles stream buffer selection
func (m ConnectionSettingsModel) updateBuffer(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	// Handle escape/back
	if key == "esc" {
		m.state = connectionSettingsMenu
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
	newList, selected := components.HandleMenuKey(msg, m.bufferList)
	m.bufferList = newList

	if selected >= 0 {
		if selected < len(streamBufferOptions) {
			m.config.StreamBufferMB = streamBufferOptions[selected].mb
			m.saveConfig()
			m.rebuildMenuList()
			m.buildBufferList()
			m.state = connectionSettingsMenu
			if m.config.StreamBufferMB == 0 {
				m.message = "✓ Stream buffering disabled"
			} else {
				m.message = fmt.Sprintf("✓ Stream buffer set to %d MB", m.config.StreamBufferMB)
			}
			m.messageIsSuccess = true
			m.messageTime = 3 // 3 seconds (decremented once per second via tickMsg)
		} else if selected == len(streamBufferOptions) {
			// Back option
			m.state = connectionSettingsMenu
		}
	}

	// Handle number shortcuts
	if key >= "1" && key <= "8" {
		num := int(key[0] - '0')
		m.bufferList.Select(num - 1)
		newModel, cmd := m.updateBuffer(tea.KeyMsg{Type: tea.KeyEnter})
		return newModel, cmd
	}

	return m, nil
}

// saveConfig saves the current configuration
func (m *ConnectionSettingsModel) saveConfig() {
	if err := storage.SaveConnectionConfig(m.config); err != nil {
		m.message = fmt.Sprintf("✗ Failed to save: %v", err)
		m.messageIsSuccess = false
		m.messageTime = 3 // 3 seconds (decremented once per second via tickMsg)
	}
}

// rebuildMenuList rebuilds the main menu list
func (m *ConnectionSettingsModel) rebuildMenuList() {
	bufferLabel := fmt.Sprintf("Set Stream Buffer (%d MB)", m.config.StreamBufferMB)
	if m.config.StreamBufferMB == 0 {
		bufferLabel = "Set Stream Buffer (Disabled)"
	}

	menuItems := []components.MenuItem{
		components.NewMenuItem(
			fmt.Sprintf("Toggle Auto-reconnect (%s)", boolToOnOff(m.config.AutoReconnect)),
			"Automatically retry connection when stream drops",
			"1",
		),
		components.NewMenuItem(
			fmt.Sprintf("Set Reconnect Delay (%d sec)", m.config.ReconnectDelay),
			"Wait time between reconnection attempts",
			"2",
		),
		components.NewMenuItem(
			bufferLabel,
			"Buffer size to handle brief signal drops",
			"3",
		),
		components.NewMenuItem(
			"Reset to Defaults",
			"Restore default connection settings",
			"4",
		),
		components.NewMenuItem(
			"Back to Settings",
			"",
			"5",
		),
	}

	m.menuList = components.CreateMenu(menuItems, "", 60, len(menuItems)+2)
}

// buildDelayList builds the reconnect delay selection list
func (m *ConnectionSettingsModel) buildDelayList() {
	menuItems := []components.MenuItem{}
	for i, delay := range reconnectDelayOptions {
		shortcut := fmt.Sprintf("%d", i+1)
		desc := ""
		if delay.seconds == m.config.ReconnectDelay {
			desc = "← Current"
		}
		menuItems = append(menuItems, components.NewMenuItem(delay.label, desc, shortcut))
	}
	menuItems = append(menuItems, components.NewMenuItem("Back", "", fmt.Sprintf("%d", len(reconnectDelayOptions)+1)))

	m.delayList = components.CreateMenu(menuItems, "", 50, len(menuItems)+2)
}

// buildBufferList builds the stream buffer selection list
func (m *ConnectionSettingsModel) buildBufferList() {
	menuItems := []components.MenuItem{}
	for i, buffer := range streamBufferOptions {
		shortcut := fmt.Sprintf("%d", i+1)
		desc := ""
		if buffer.mb == m.config.StreamBufferMB {
			desc = "← Current"
		}
		menuItems = append(menuItems, components.NewMenuItem(buffer.label, desc, shortcut))
	}
	menuItems = append(menuItems, components.NewMenuItem("Back", "", fmt.Sprintf("%d", len(streamBufferOptions)+1)))

	m.bufferList = components.CreateMenu(menuItems, "", 50, len(menuItems)+2)
}

// View renders the connection settings screen
func (m ConnectionSettingsModel) View() string {
	switch m.state {
	case connectionSettingsMenu:
		return m.viewMenu()
	case connectionSettingsDelay:
		return m.viewDelay()
	case connectionSettingsBuffer:
		return m.viewBuffer()
	}
	return "Unknown state"
}

// viewMenu renders the main menu
func (m ConnectionSettingsModel) viewMenu() string {
	var content strings.Builder

	t := theme.Current()
	titleStyle := lipgloss.NewStyle().
		Foreground(t.HighlightColor()).
		Bold(true).
		PaddingLeft(t.Padding.ListItemLeft)

	// Title
	content.WriteString(titleStyle.Render("⚙️  Settings > Connection Settings"))
	content.WriteString("\n\n")

	// Current settings summary
	content.WriteString(subtitleStyle().Render("Current Settings:"))
	content.WriteString("\n\n")
	fmt.Fprintf(&content, "  Auto-reconnect:         %s\n", boolToEnabledDisabled(m.config.AutoReconnect))
	fmt.Fprintf(&content, "  Reconnect delay:        %d seconds\n", m.config.ReconnectDelay)
	if m.config.StreamBufferMB == 0 {
		content.WriteString("  Stream buffer:          Disabled\n")
	} else {
		fmt.Fprintf(&content, "  Stream buffer:          %d MB\n", m.config.StreamBufferMB)
	}
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
		content.WriteString(infoStyle().Render("ℹ️  Helps maintain stable playback on unstable networks (4G/GPRS)"))
	}

	return RenderPageWithBottomHelp(PageLayout{
		Content: content.String(),
		Help:    "↑↓/jk: Navigate • Enter: Select • 1-5: Shortcut • Esc: Back • 0: Main Menu",
	}, m.height)
}

// viewDelay renders the delay selection screen
func (m ConnectionSettingsModel) viewDelay() string {
	var content strings.Builder

	t := theme.Current()
	titleStyle := lipgloss.NewStyle().
		Foreground(t.HighlightColor()).
		Bold(true).
		PaddingLeft(t.Padding.ListItemLeft)

	// Title
	content.WriteString(titleStyle.Render("⚙️  Settings > Connection Settings > Reconnect Delay"))
	content.WriteString("\n\n")

	content.WriteString(subtitleStyle().Render("Select reconnect delay:"))
	content.WriteString("\n\n")
	fmt.Fprintf(&content, "  Current: %d seconds\n", m.config.ReconnectDelay)
	content.WriteString("\n")

	// Delay list
	content.WriteString(m.delayList.View())

	content.WriteString("\n\n")
	content.WriteString(infoStyle().Render("Shorter delays reconnect faster but may strain weak connections"))

	return RenderPageWithBottomHelp(PageLayout{
		Content: content.String(),
		Help:    "↑↓/jk: Navigate • Enter: Select • 1-7: Shortcut • Esc: Back • 0: Main Menu",
	}, m.height)
}

// viewBuffer renders the buffer selection screen
func (m ConnectionSettingsModel) viewBuffer() string {
	var content strings.Builder

	t := theme.Current()
	titleStyle := lipgloss.NewStyle().
		Foreground(t.HighlightColor()).
		Bold(true).
		PaddingLeft(t.Padding.ListItemLeft)

	// Title
	content.WriteString(titleStyle.Render("⚙️  Settings > Connection Settings > Stream Buffer"))
	content.WriteString("\n\n")

	content.WriteString(subtitleStyle().Render("Select stream buffer size:"))
	content.WriteString("\n\n")
	if m.config.StreamBufferMB == 0 {
		content.WriteString("  Current: Disabled\n")
	} else {
		fmt.Fprintf(&content, "  Current: %d MB\n", m.config.StreamBufferMB)
	}
	content.WriteString("\n")

	// Buffer list
	content.WriteString(m.bufferList.View())

	content.WriteString("\n\n")
	content.WriteString(infoStyle().Render("Larger buffers handle longer signal drops but use more memory"))

	return RenderPageWithBottomHelp(PageLayout{
		Content: content.String(),
		Help:    "↑↓/jk: Navigate • Enter: Select • 1-8: Shortcut • Esc: Back • 0: Main Menu",
	}, m.height)
}
