# Connection Settings Implementation Plan

## Overview
Add auto-reconnect feature to TERA to handle unstable network connections (GPRS/4G). This addresses GitHub issue #4 where users lose signal and must manually reconnect.

## Feature Summary
- **Auto-reconnect**: Automatically retry connection when stream drops
- **Reconnect delay**: Configurable delay between reconnect attempts
- **Stream buffer**: Cache size to handle brief signal drops
- All settings configurable via Settings menu UI

---

## Implementation Steps

### 1. Add Connection Configuration to Data Model

**File:** `internal/storage/models.go`

**Changes:**
```go
// Add new struct after ShuffleConfig

// ConnectionConfig represents connection/streaming configuration
type ConnectionConfig struct {
	AutoReconnect    bool `yaml:"auto_reconnect"`
	ReconnectDelay   int  `yaml:"reconnect_delay"`   // in seconds
	StreamBufferMB   int  `yaml:"stream_buffer_mb"`  // in megabytes
}

// DefaultConnectionConfig returns default connection configuration
func DefaultConnectionConfig() ConnectionConfig {
	return ConnectionConfig{
		AutoReconnect:  true,   // Enable by default for better UX
		ReconnectDelay: 5,      // 5 seconds between retries
		StreamBufferMB: 50,     // 50MB buffer
	}
}
```

**Rationale:**
- Separate config struct keeps concerns organized
- Default values chosen based on GitHub issue research
- YAML tags for file persistence

---

### 2. Add Connection Config Storage Functions

**File:** Create `internal/storage/connection_config.go`

**Content:**
```go
package storage

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const connectionConfigFile = "connection_config.yaml"

// LoadConnectionConfig loads connection configuration from file
func LoadConnectionConfig() (ConnectionConfig, error) {
	configPath, err := GetConfigPath()
	if err != nil {
		return DefaultConnectionConfig(), err
	}

	filePath := filepath.Join(configPath, connectionConfigFile)

	// If file doesn't exist, return defaults
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return DefaultConnectionConfig(), nil
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return DefaultConnectionConfig(), fmt.Errorf("failed to read connection config: %w", err)
	}

	var config ConnectionConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return DefaultConnectionConfig(), fmt.Errorf("failed to parse connection config: %w", err)
	}

	// Validate and apply bounds
	if config.ReconnectDelay < 1 {
		config.ReconnectDelay = 1
	}
	if config.ReconnectDelay > 30 {
		config.ReconnectDelay = 30
	}
	if config.StreamBufferMB < 10 {
		config.StreamBufferMB = 10
	}
	if config.StreamBufferMB > 200 {
		config.StreamBufferMB = 200
	}

	return config, nil
}

// SaveConnectionConfig saves connection configuration to file
func SaveConnectionConfig(config ConnectionConfig) error {
	configPath, err := GetConfigPath()
	if err != nil {
		return err
	}

	// Ensure config directory exists
	if err := os.MkdirAll(configPath, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	filePath := filepath.Join(configPath, connectionConfigFile)

	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal connection config: %w", err)
	}

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write connection config: %w", err)
	}

	return nil
}
```

**Rationale:**
- Follows existing pattern from `shuffle_config.go`
- Input validation prevents invalid values
- Graceful fallback to defaults if file missing

---

### 3. Update MPV Player to Use Connection Settings

**File:** `internal/player/mpv.go`

**Changes to `Play` method:**

Find this section (around line 67-90):
```go
	// Create mpv command with appropriate flags
	// --no-video: audio only
	// --no-terminal: don't take over terminal
	// --really-quiet: minimal output
	// --no-cache: no buffering for live streams
	// --volume: set initial volume
	// --input-ipc-server: enable IPC for runtime control
	p.cmd = exec.Command("mpv",
		"--no-video",
		"--no-terminal",
		"--really-quiet",
		"--no-cache",
		fmt.Sprintf("--volume=%d", volumeToUse),
		fmt.Sprintf("--input-ipc-server=%s", p.socketPath),
		station.URLResolved,
	)
```

Replace with:
```go
	// Load connection configuration
	connConfig, err := storage.LoadConnectionConfig()
	if err != nil {
		// Fall back to defaults on error
		connConfig = storage.DefaultConnectionConfig()
	}

	// Build mpv arguments
	args := []string{
		"--no-video",
		"--no-terminal",
		"--really-quiet",
		fmt.Sprintf("--volume=%d", volumeToUse),
		fmt.Sprintf("--input-ipc-server=%s", p.socketPath),
	}

	// Add connection-related flags based on config
	if connConfig.AutoReconnect {
		// Enable force loop to retry after stream drops
		args = append(args, "--loop-playlist=force")
		
		// FFmpeg reconnect flags for network-level reconnection
		args = append(args,
			fmt.Sprintf("--stream-lavf-o=reconnect_streamed=1,reconnect_delay_max=%d", connConfig.ReconnectDelay),
		)
	}

	// Add caching/buffering based on config
	if connConfig.StreamBufferMB > 0 {
		args = append(args,
			"--cache=yes",
			fmt.Sprintf("--demuxer-max-bytes=%dM", connConfig.StreamBufferMB),
		)
	} else {
		// No buffering (original behavior)
		args = append(args, "--no-cache")
	}

	// Add URL as final argument
	args = append(args, station.URLResolved)

	// Create mpv command
	p.cmd = exec.Command("mpv", args...)
```

**Add import at top of file:**
```go
import (
	// ... existing imports ...
	"github.com/shinokada/tera/internal/storage"
)
```

**Rationale:**
- Dynamically builds mpv arguments based on config
- Maintains backward compatibility (defaults work like current version)
- Uses proven mpv/FFmpeg flags from GitHub issue research

---

### 4. Create Connection Settings UI Model

**File:** Create `internal/ui/connection_settings.go`

**Content:**
```go
package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/shinokada/tera/internal/storage"
	"github.com/shinokada/tera/internal/theme"
	"github.com/shinokada/tera/internal/ui/components"
)

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
	return ticksEverySecond()
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
		return m, ticksEverySecond()
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
			m.message = "✓ Reset to default settings"
			m.messageIsSuccess = true
			m.messageTime = 180
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
		delays := []int{1, 3, 5, 10, 15, 30}
		if selected < len(delays) {
			m.config.ReconnectDelay = delays[selected]
			m.saveConfig()
			m.rebuildMenuList()
			m.buildDelayList()
			m.state = connectionSettingsMenu
			m.message = fmt.Sprintf("✓ Reconnect delay set to %d seconds", m.config.ReconnectDelay)
			m.messageIsSuccess = true
			m.messageTime = 180
		} else if selected == len(delays) {
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
		buffers := []int{10, 25, 50, 100, 150, 200}
		if selected < len(buffers) {
			m.config.StreamBufferMB = buffers[selected]
			m.saveConfig()
			m.rebuildMenuList()
			m.buildBufferList()
			m.state = connectionSettingsMenu
			m.message = fmt.Sprintf("✓ Stream buffer set to %d MB", m.config.StreamBufferMB)
			m.messageIsSuccess = true
			m.messageTime = 180
		} else if selected == len(buffers) {
			// Back option
			m.state = connectionSettingsMenu
		}
	}

	// Handle number shortcuts
	if key >= "1" && key <= "7" {
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
		m.messageTime = 180
	}
}

// rebuildMenuList rebuilds the main menu list
func (m *ConnectionSettingsModel) rebuildMenuList() {
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
			fmt.Sprintf("Set Stream Buffer (%d MB)", m.config.StreamBufferMB),
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
	delays := []struct {
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

	menuItems := []components.MenuItem{}
	for i, delay := range delays {
		shortcut := fmt.Sprintf("%d", i+1)
		desc := ""
		if delay.seconds == m.config.ReconnectDelay {
			desc = "← Current"
		}
		menuItems = append(menuItems, components.NewMenuItem(delay.label, desc, shortcut))
	}
	menuItems = append(menuItems, components.NewMenuItem("Back", "", "7"))

	m.delayList = components.CreateMenu(menuItems, "", 50, len(menuItems)+2)
}

// buildBufferList builds the stream buffer selection list
func (m *ConnectionSettingsModel) buildBufferList() {
	buffers := []struct {
		mb    int
		label string
	}{
		{10, "10 MB (Minimal)"},
		{25, "25 MB (Light)"},
		{50, "50 MB (Default)"},
		{100, "100 MB (Heavy)"},
		{150, "150 MB (Maximum)"},
		{200, "200 MB (Extreme)"},
	}

	menuItems := []components.MenuItem{}
	for i, buffer := range buffers {
		shortcut := fmt.Sprintf("%d", i+1)
		desc := ""
		if buffer.mb == m.config.StreamBufferMB {
			desc = "← Current"
		}
		menuItems = append(menuItems, components.NewMenuItem(buffer.label, desc, shortcut))
	}
	menuItems = append(menuItems, components.NewMenuItem("Back", "", "7"))

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
	content.WriteString(fmt.Sprintf("  Auto-reconnect:         %s\n", boolToEnabledDisabled(m.config.AutoReconnect)))
	content.WriteString(fmt.Sprintf("  Reconnect delay:        %d seconds\n", m.config.ReconnectDelay))
	content.WriteString(fmt.Sprintf("  Stream buffer:          %d MB\n", m.config.StreamBufferMB))
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
	content.WriteString(fmt.Sprintf("  Current: %d seconds\n", m.config.ReconnectDelay))
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
	content.WriteString(fmt.Sprintf("  Current: %d MB\n", m.config.StreamBufferMB))
	content.WriteString("\n")

	// Buffer list
	content.WriteString(m.bufferList.View())

	content.WriteString("\n\n")
	content.WriteString(infoStyle().Render("Larger buffers handle longer signal drops but use more memory"))

	return RenderPageWithBottomHelp(PageLayout{
		Content: content.String(),
		Help:    "↑↓/jk: Navigate • Enter: Select • 1-7: Shortcut • Esc: Back • 0: Main Menu",
	}, m.height)
}
```

**Rationale:**
- Follows exact pattern of `shuffle_settings.go` for consistency
- Three sub-screens: menu, delay selection, buffer selection
- Real-time feedback with auto-save
- Helper text explains what each setting does

---

### 5. Integrate Connection Settings into Settings Menu

**File:** `internal/ui/settings.go`

**Changes:**

1. Add new constant after line ~18:
```go
const (
	settingsStateMenu settingsState = iota
	settingsStateTheme
	settingsStateConnection  // ADD THIS
	settingsStateHistory
	settingsStateUpdates
	settingsStateAbout
)
```

2. Find the settings menu creation (search for "Theme / Colors") and update:
```go
menuItems := []components.MenuItem{
	components.NewMenuItem(
		"Theme / Colors",
		"Customize TERA's appearance",
		"1",
	),
	components.NewMenuItem(
		"Connection Settings",  // ADD THIS
		"Auto-reconnect and buffering",
		"2",
	),
	components.NewMenuItem(
		"Shuffle Settings",
		"Configure shuffle mode behavior",
		"3",  // Changed from "2"
	),
	components.NewMenuItem(
		"Search History",
		"Manage search history",
		"4",  // Changed from "3"
	),
	components.NewMenuItem(
		"Check for Updates",
		"Check for newer versions",
		"5",  // Changed from "4"
	),
	components.NewMenuItem(
		"About TERA",
		fmt.Sprintf("Version %s", Version),
		"6",  // Changed from "5"
	),
}
```

3. Update the menu selection handler to route to connection settings:
```go
if selected >= 0 {
	switch selected {
	case 0: // Theme
		m.state = settingsStateTheme
	case 1: // Connection Settings - ADD THIS CASE
		return m, func() tea.Msg {
			return navigateMsg{screen: screenConnectionSettings}
		}
	case 2: // Shuffle Settings (was case 1)
		return m, func() tea.Msg {
			return navigateMsg{screen: screenShuffleSettings}
		}
	case 3: // Search History (was case 2)
		m.state = settingsStateHistory
	case 4: // Check for Updates (was case 3)
		// ... existing code ...
	case 5: // About (was case 4)
		m.state = settingsStateAbout
	}
}
```

4. Update number shortcuts (find the section with `key >= "1" && key <= "5"`):
```go
// Handle number shortcuts
if key >= "1" && key <= "6" {  // Change from "5" to "6"
	num := int(key[0] - '0')
	m.menuList.Select(num - 1)
	newModel, cmd := m.updateMenu(tea.KeyMsg{Type: tea.KeyEnter})
	return newModel, cmd
}
```

**Rationale:**
- Adds Connection Settings as option #2 (between Theme and Shuffle)
- Updates all subsequent numbers
- Follows existing navigation pattern

---

### 6. Update App Model Navigation

**File:** `internal/ui/app.go`

**Changes:**

1. Add screen constant (find `screenType` enum):
```go
const (
	screenMainMenu screenType = iota
	screenList
	screenPlay
	screenLucky
	screenSearch
	screenSettings
	screenShuffleSettings
	screenConnectionSettings  // ADD THIS
	screenGist
)
```

2. Add field to App struct (find the `App` struct definition):
```go
type App struct {
	state               screenType
	mainMenu            MainMenuModel
	listModel           ListModel
	playModel           PlayModel
	luckyModel          LuckyModel
	searchModel         SearchModel
	settingsModel       SettingsModel
	shuffleSettingsModel ShuffleSettingsModel
	connectionSettingsModel ConnectionSettingsModel  // ADD THIS
	gistModel           GistModel
	width               int
	height              int
	searchHistoryStore  *storage.SearchHistoryStore
	shuffleManager      *shuffle.Manager
	client              *api.Client
}
```

3. Initialize in `NewApp` function (find where other models are initialized):
```go
func NewApp(client *api.Client, favPath string, history *storage.SearchHistoryStore, shuffleManager *shuffle.Manager) App {
	return App{
		// ... existing initializations ...
		shuffleSettingsModel:    NewShuffleSettingsModel(),
		connectionSettingsModel: NewConnectionSettingsModel(),  // ADD THIS
		gistModel:              NewGistModel(),
		// ... rest of initializations ...
	}
}
```

4. Add to `Update` switch statement (find the big switch on `a.state`):
```go
func (a App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case navigateMsg:
		a.state = msg.screen
		switch msg.screen {
		// ... existing cases ...
		case screenShuffleSettings:
			a.shuffleSettingsModel = NewShuffleSettingsModel()
			return a, a.shuffleSettingsModel.Init()
		case screenConnectionSettings:  // ADD THIS CASE
			a.connectionSettingsModel = NewConnectionSettingsModel()
			return a, a.connectionSettingsModel.Init()
		// ... rest of cases ...
		}
	}

	// ... later in Update, find the screen routing switch ...
	switch a.state {
	// ... existing cases ...
	case screenShuffleSettings:
		newModel, cmd := a.shuffleSettingsModel.Update(msg)
		a.shuffleSettingsModel = newModel.(ShuffleSettingsModel)
		return a, cmd
	case screenConnectionSettings:  // ADD THIS CASE
		newModel, cmd := a.connectionSettingsModel.Update(msg)
		a.connectionSettingsModel = newModel.(ConnectionSettingsModel)
		return a, cmd
	// ... rest of cases ...
	}
}
```

5. Add to `View` method (find the screen view switch):
```go
func (a App) View() string {
	switch a.state {
	// ... existing cases ...
	case screenShuffleSettings:
		return a.shuffleSettingsModel.View()
	case screenConnectionSettings:  // ADD THIS CASE
		return a.connectionSettingsModel.View()
	// ... rest of cases ...
	}
}
```

**Rationale:**
- Integrates new screen into app navigation flow
- Follows exact pattern used by shuffle settings
- Enables navigation: Settings → Connection Settings → back

---

## Testing Plan

### Manual Testing Checklist

1. **Configuration Persistence**
   - [ ] Change auto-reconnect setting → restart app → verify setting persisted
   - [ ] Change reconnect delay → restart app → verify setting persisted
   - [ ] Change buffer size → restart app → verify setting persisted
   - [ ] Reset to defaults → verify all settings return to defaults

2. **UI Navigation**
   - [ ] Main Menu → Settings → Connection Settings
   - [ ] Navigate with arrow keys and j/k
   - [ ] Use number shortcuts (1-7)
   - [ ] Escape key returns to Settings
   - [ ] 0 key returns to Main Menu

3. **Auto-Reconnect Functionality**
   - [ ] Enable auto-reconnect → play station → simulate network drop (disconnect WiFi briefly) → verify auto-reconnects
   - [ ] Disable auto-reconnect → play station → simulate network drop → verify stream stops
   - [ ] Test different reconnect delays (1s, 5s, 10s)

4. **Buffer Testing**
   - [ ] Set buffer to 10MB → play station → monitor memory usage
   - [ ] Set buffer to 200MB → play station → monitor memory usage
   - [ ] Test buffering during brief connection interruptions

5. **Edge Cases**
   - [ ] What happens if config file is corrupted? (Should fall back to defaults)
   - [ ] What happens if mpv is not installed? (Existing error handling should work)
   - [ ] Can user set extreme values? (Validation should prevent this)

### Test Scenarios for GitHub Issue #4

**Scenario 1: Highway driving with intermittent 4G**
- Set: Auto-reconnect ON, Delay 5s, Buffer 50MB
- Expected: Brief signal drops (< 10s) handled by buffer; longer drops trigger reconnect

**Scenario 2: Rural area with weak GPRS**
- Set: Auto-reconnect ON, Delay 10s, Buffer 100MB
- Expected: Reconnects after connection loss; larger buffer handles longer gaps

**Scenario 3: User prefers no auto-reconnect**
- Set: Auto-reconnect OFF
- Expected: Stream stops on connection loss (original behavior)

---

## Files to Create

1. `internal/storage/connection_config.go` - New file
2. `internal/ui/connection_settings.go` - New file

## Files to Modify

1. `internal/storage/models.go` - Add ConnectionConfig struct
2. `internal/player/mpv.go` - Update Play method to use config
3. `internal/ui/settings.go` - Add menu item and navigation
4. `internal/ui/app.go` - Integrate new screen

---

## Configuration File Location

Config will be stored at:
- **Linux/macOS**: `~/.config/tera/connection_config.yaml`
- **Windows**: `%APPDATA%\tera\connection_config.yaml`

**Example `connection_config.yaml`:**
```yaml
auto_reconnect: true
reconnect_delay: 5
stream_buffer_mb: 50
```

---

## Rollback Plan

If issues arise, users can:
1. Delete `connection_config.yaml` to restore defaults
2. Disable auto-reconnect via Settings menu
3. Set buffer to 0MB to disable caching (returns to original `--no-cache` behavior)

---

## Future Enhancements (Out of Scope)

- Advanced mode: Allow custom mpv flags
- Connection quality indicator
- Network statistics display
- Automatic buffer adjustment based on connection quality
- Per-station connection settings

---

## Implementation Order

1. ✅ Create this plan document
2. Add ConnectionConfig to models.go
3. Create connection_config.go storage functions
4. Create connection_settings.go UI
5. Update settings.go to add menu item
6. Update app.go for navigation
7. Update mpv.go to use connection settings
8. Test manually with checklist
9. Update documentation (README)
10. Close GitHub issue #4

---

## Estimated Time

- Step 2-4: ~30 minutes (data models and storage)
- Step 5-6: ~20 minutes (UI and navigation)
- Step 7: ~15 minutes (MPV integration)
- Step 8: ~30 minutes (testing)
- **Total: ~2 hours**

---

## Notes for Future Maintainer

- The connection settings follow the same pattern as shuffle settings for consistency
- MPV flags are well-documented and battle-tested (from GitHub issue research)
- Default values (5s delay, 50MB buffer) are reasonable for most use cases
- Users with very stable connections can disable auto-reconnect
- Validation prevents unreasonable values (e.g., 1000MB buffer)

---

## Questions to Consider

1. Should auto-reconnect be enabled by default? **Recommendation: Yes** (better UX for mobile users)
2. Should there be a "test connection" button? **Recommendation: No** (adds complexity, user can just play a station)
3. Should we show a "reconnecting..." indicator? **Recommendation: Future enhancement** (out of scope for initial implementation)

---

End of Implementation Plan
