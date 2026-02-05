package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/shinokada/tera/internal/storage"
	"github.com/shinokada/tera/internal/ui/components"
)

type appearanceState int

const (
	appearanceStateMenu appearanceState = iota
	appearanceStateModeSelect
	appearanceStateTextInput
	appearanceStateAsciiInput
	appearanceStateAlignmentSelect
	appearanceStateWidthInput
	appearanceStateColorInput
	appearanceStatePaddingInput
	appearanceStatePreview
)

type AppearanceSettingsModel struct {
	state          appearanceState
	width          int
	height         int
	message        string
	messageTime    int
	messageSuccess bool

	// Configuration being edited
	config storage.AppearanceConfig

	// Input widgets
	modeList           list.Model
	textInput          textinput.Model
	asciiInput         textarea.Model
	alignmentList      list.Model
	widthInput         textinput.Model
	colorInput         textinput.Model
	paddingTopInput    textinput.Model
	paddingBottomInput textinput.Model
	paddingFocusTop    bool // Track which padding input has focus

	// Menu
	menuList list.Model

	// Help
	helpModel components.HelpModel
}

func NewAppearanceSettingsModel() AppearanceSettingsModel {
	// Load current config
	config, err := storage.LoadAppearanceConfig()
	if err != nil {
		config = storage.DefaultAppearanceConfig()
	}

	m := AppearanceSettingsModel{
		state:  appearanceStateMenu,
		config: config,
	}

	// Initialize menu
	m.initMenu()

	// Initialize mode selector
	m.initModeSelector()

	// Initialize text input
	ti := textinput.New()
	ti.Placeholder = "Enter custom header text"
	ti.Width = 50
	ti.CharLimit = 100
	ti.SetValue(config.Header.CustomText)
	m.textInput = ti

	// Initialize ASCII art input
	ta := textarea.New()
	ta.Placeholder = "Paste your ASCII art here (max 15 lines)"
	ta.SetWidth(70)
	ta.SetHeight(15)
	ta.CharLimit = 2000
	ta.SetValue(config.Header.AsciiArt)
	m.asciiInput = ta

	// Initialize alignment selector
	m.initAlignmentSelector()

	// Initialize width input
	wi := textinput.New()
	wi.Placeholder = "Width (10-120)"
	wi.Width = 20
	wi.CharLimit = 3
	wi.SetValue(fmt.Sprintf("%d", config.Header.Width))
	m.widthInput = wi

	// Initialize color input
	ci := textinput.New()
	ci.Placeholder = "Color (e.g., auto, #FF0000, 33)"
	ci.Width = 30
	ci.CharLimit = 20
	ci.SetValue(config.Header.Color)
	m.colorInput = ci

	// Initialize padding top input
	pti := textinput.New()
	pti.Placeholder = "Padding Top (0-5)"
	pti.Width = 20
	pti.CharLimit = 1
	pti.SetValue(fmt.Sprintf("%d", config.Header.PaddingTop))
	m.paddingTopInput = pti

	// Initialize padding bottom input
	pbi := textinput.New()
	pbi.Placeholder = "Padding Bottom (0-5)"
	pbi.Width = 20
	pbi.CharLimit = 1
	pbi.SetValue(fmt.Sprintf("%d", config.Header.PaddingBottom))
	m.paddingBottomInput = pbi

	// Initialize help
	m.helpModel = components.NewHelpModel(components.CreateAppearanceHelp())

	return m
}

func (m *AppearanceSettingsModel) initMenu() {
	// Format current values for display
	customTextDisplay := m.config.Header.CustomText
	if customTextDisplay == "" {
		customTextDisplay = "(empty)"
	} else if len(customTextDisplay) > 30 {
		customTextDisplay = customTextDisplay[:27] + "..."
	}

	menuItems := []components.MenuItem{
		components.NewMenuItem(fmt.Sprintf("Header Mode: %s", m.config.Header.Mode), string(m.config.Header.Mode), "1"),
		components.NewMenuItem(fmt.Sprintf("Custom Text: %s", customTextDisplay), m.config.Header.CustomText, "2"),
		components.NewMenuItem(fmt.Sprintf("Alignment: %s", m.config.Header.Alignment), m.config.Header.Alignment, "3"),
		components.NewMenuItem(fmt.Sprintf("Width: %d", m.config.Header.Width), fmt.Sprintf("%d", m.config.Header.Width), "4"),
		components.NewMenuItem(fmt.Sprintf("Color: %s", m.config.Header.Color), m.config.Header.Color, "5"),
		components.NewMenuItem(fmt.Sprintf("Bold: %t", m.config.Header.Bold), fmt.Sprintf("%t", m.config.Header.Bold), "6"),
		components.NewMenuItem(fmt.Sprintf("Padding: Top=%d, Bottom=%d", m.config.Header.PaddingTop, m.config.Header.PaddingBottom), "Edit padding", "7"),
		components.NewMenuItem("Edit ASCII Art", "Multi-line editor", "8"),
		components.NewMenuItem("Preview", "See how it looks", "p"),
		components.NewMenuItem("Save", "Save changes", "s"),
		components.NewMenuItem("Reset to Default", "Restore defaults", "r"),
		components.NewMenuItem("Back", "Return to settings", "0"),
	}

	m.menuList = components.CreateMenu(menuItems, "Appearance Settings", 60, len(menuItems)+2)
}

func (m *AppearanceSettingsModel) initModeSelector() {
	menuItems := []components.MenuItem{
		components.NewMenuItem("default", "Show 'TERA' (default)", "1"),
		components.NewMenuItem("text", "Custom text", "2"),
		components.NewMenuItem("ascii", "ASCII art", "3"),
		components.NewMenuItem("none", "No header", "4"),
	}

	m.modeList = components.CreateMenu(menuItems, "Select Header Mode", 50, len(menuItems)+2)

	// Set cursor to current mode
	for i := 0; i < len(menuItems); i++ {
		if menuItems[i].Title() == string(m.config.Header.Mode) {
			m.modeList.Select(i)
			break
		}
	}
}

func (m *AppearanceSettingsModel) initAlignmentSelector() {
	menuItems := []components.MenuItem{
		components.NewMenuItem("left", "Left aligned", "1"),
		components.NewMenuItem("center", "Center aligned", "2"),
		components.NewMenuItem("right", "Right aligned", "3"),
	}

	m.alignmentList = components.CreateMenu(menuItems, "Select Alignment", 50, len(menuItems)+2)

	// Set cursor to current alignment
	for i := 0; i < len(menuItems); i++ {
		if menuItems[i].Title() == m.config.Header.Alignment {
			m.alignmentList.Select(i)
			break
		}
	}
}

func (m AppearanceSettingsModel) Init() tea.Cmd {
	return nil
}

func (m AppearanceSettingsModel) Update(msg tea.Msg) (AppearanceSettingsModel, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg.(type) {
	case tickMsg:
		// Decrease message timer
		if m.messageTime > 0 {
			m.messageTime--
			if m.messageTime == 0 {
				m.message = ""
			} else {
				return m, tickEverySecond()
			}
		}
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		// Update list sizes
		listHeight := m.height - 10
		m.menuList.SetSize(m.width-4, listHeight)
		m.modeList.SetSize(m.width-4, listHeight)
		m.alignmentList.SetSize(m.width-4, listHeight)

		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			if m.state == appearanceStateMenu {
				return m, func() tea.Msg {
					return navigateMsg{screen: screenSettings}
				}
			}
			// Go back to menu from any sub-screen
			m.state = appearanceStateMenu
			m.initMenu() // Refresh menu with updated values
			return m, nil

		case "esc":
			if m.state == appearanceStateMenu {
				return m, func() tea.Msg {
					return navigateMsg{screen: screenSettings}
				}
			}
			// Go back to menu from any sub-screen
			m.state = appearanceStateMenu
			m.initMenu() // Refresh menu
			return m, nil

		case "?":
			m.helpModel.Toggle()
			return m, nil

		case "enter":
			return m.handleEnter()
		}
	}

	// Update current widget based on state
	switch m.state {
	case appearanceStateMenu:
		m.menuList, cmd = m.menuList.Update(msg)
		cmds = append(cmds, cmd)

	case appearanceStateModeSelect:
		m.modeList, cmd = m.modeList.Update(msg)
		cmds = append(cmds, cmd)

	case appearanceStateTextInput:
		m.textInput, cmd = m.textInput.Update(msg)
		cmds = append(cmds, cmd)

	case appearanceStateAsciiInput:
		m.asciiInput, cmd = m.asciiInput.Update(msg)
		cmds = append(cmds, cmd)

	case appearanceStateAlignmentSelect:
		m.alignmentList, cmd = m.alignmentList.Update(msg)
		cmds = append(cmds, cmd)

	case appearanceStateWidthInput:
		m.widthInput, cmd = m.widthInput.Update(msg)
		cmds = append(cmds, cmd)

	case appearanceStateColorInput:
		m.colorInput, cmd = m.colorInput.Update(msg)
		cmds = append(cmds, cmd)

	case appearanceStatePaddingInput:
		// Handle tab to switch between inputs
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			if keyMsg.String() == "tab" || keyMsg.String() == "shift+tab" {
				m.paddingFocusTop = !m.paddingFocusTop
				if m.paddingFocusTop {
					m.paddingTopInput.Focus()
					m.paddingBottomInput.Blur()
				} else {
					m.paddingBottomInput.Focus()
					m.paddingTopInput.Blur()
				}
				return m, nil
			}
		}
		// Update the focused input
		if m.paddingFocusTop {
			m.paddingTopInput, cmd = m.paddingTopInput.Update(msg)
		} else {
			m.paddingBottomInput, cmd = m.paddingBottomInput.Update(msg)
		}
		cmds = append(cmds, cmd)
	}

	// Update help
	m.helpModel, cmd = m.helpModel.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m *AppearanceSettingsModel) handleEnter() (AppearanceSettingsModel, tea.Cmd) {
	switch m.state {
	case appearanceStateMenu:
		selectedItem := m.menuList.SelectedItem()
		if selectedItem == nil {
			return *m, nil
		}

		// Get shortcut to identify menu item (more reliable than title now)
		shortcut := selectedItem.(components.MenuItem).Shortcut()

		switch shortcut {
		case "1": // Header Mode
			m.state = appearanceStateModeSelect

		case "2": // Custom Text
			m.state = appearanceStateTextInput
			m.textInput.Focus()

		case "3": // Alignment
			m.state = appearanceStateAlignmentSelect

		case "4": // Width
			m.state = appearanceStateWidthInput
			m.widthInput.Focus()

		case "5": // Color
			m.state = appearanceStateColorInput
			m.colorInput.Focus()

		case "6": // Bold
			m.config.Header.Bold = !m.config.Header.Bold
			m.initMenu() // Refresh menu
			m.showMessage("Bold toggled", true)
			return *m, tickEverySecond()

		case "7": // Padding
			m.state = appearanceStatePaddingInput
			m.paddingFocusTop = true
			m.paddingTopInput.Focus()
			m.paddingBottomInput.Blur()

		case "8": // Edit ASCII Art
			m.state = appearanceStateAsciiInput
			m.asciiInput.Focus()

		case "p": // Preview
			m.state = appearanceStatePreview

		case "s": // Save
			return m.saveConfig()

		case "r": // Reset to Default
			m.config = storage.DefaultAppearanceConfig()
			m.initMenu()
			m.textInput.SetValue(m.config.Header.CustomText)
			m.asciiInput.SetValue(m.config.Header.AsciiArt)
			m.widthInput.SetValue(fmt.Sprintf("%d", m.config.Header.Width))
			m.colorInput.SetValue(m.config.Header.Color)
			m.paddingTopInput.SetValue(fmt.Sprintf("%d", m.config.Header.PaddingTop))
			m.paddingBottomInput.SetValue(fmt.Sprintf("%d", m.config.Header.PaddingBottom))
			m.showMessage("Reset to defaults", true)
			return *m, tickEverySecond()

		case "0": // Back
			return *m, func() tea.Msg {
				return navigateMsg{screen: screenSettings}
			}
		}

	case appearanceStateModeSelect:
		selectedItem := m.modeList.SelectedItem()
		if selectedItem != nil {
			m.config.Header.Mode = storage.HeaderMode(selectedItem.(components.MenuItem).Title())
			m.initMenu()
			m.state = appearanceStateMenu
			m.showMessage(fmt.Sprintf("Mode set to: %s", m.config.Header.Mode), true)
			return *m, tickEverySecond()
		}

	case appearanceStateTextInput:
		m.config.Header.CustomText = m.textInput.Value()
		m.initMenu()
		m.state = appearanceStateMenu
		m.showMessage("Custom text updated", true)
		return *m, tickEverySecond()

	case appearanceStateAsciiInput:
		m.config.Header.AsciiArt = m.asciiInput.Value()
		m.initMenu()
		m.state = appearanceStateMenu
		m.showMessage("ASCII art updated", true)
		return *m, tickEverySecond()

	case appearanceStateAlignmentSelect:
		selectedItem := m.alignmentList.SelectedItem()
		if selectedItem != nil {
			m.config.Header.Alignment = selectedItem.(components.MenuItem).Title()
			m.initMenu()
			m.state = appearanceStateMenu
			m.showMessage(fmt.Sprintf("Alignment set to: %s", m.config.Header.Alignment), true)
			return *m, tickEverySecond()
		}

	case appearanceStateWidthInput:
		width := 0
		_, err := fmt.Sscanf(m.widthInput.Value(), "%d", &width)
		if err != nil || width < 10 || width > 120 {
			m.showMessage("Width must be a number between 10-120", false)
			return *m, tickEverySecond()
		} else {
			m.config.Header.Width = width
			m.initMenu()
			m.state = appearanceStateMenu
			m.showMessage(fmt.Sprintf("Width set to: %d", width), true)
			return *m, tickEverySecond()
		}

	case appearanceStateColorInput:
		m.config.Header.Color = m.colorInput.Value()
		m.initMenu()
		m.state = appearanceStateMenu
		m.showMessage("Color updated", true)
		return *m, tickEverySecond()

	case appearanceStatePaddingInput:
		// Validate both padding values
		paddingTop := 0
		paddingBottom := 0
		_, errTop := fmt.Sscanf(m.paddingTopInput.Value(), "%d", &paddingTop)
		_, errBottom := fmt.Sscanf(m.paddingBottomInput.Value(), "%d", &paddingBottom)

		if errTop != nil || paddingTop < 0 || paddingTop > 5 {
			m.showMessage("Padding top must be a number between 0-5", false)
			return *m, tickEverySecond()
		} else if errBottom != nil || paddingBottom < 0 || paddingBottom > 5 {
			m.showMessage("Padding bottom must be a number between 0-5", false)
			return *m, tickEverySecond()
		} else {
			m.config.Header.PaddingTop = paddingTop
			m.config.Header.PaddingBottom = paddingBottom
			m.initMenu()
			m.state = appearanceStateMenu
			m.showMessage(fmt.Sprintf("Padding updated: Top=%d, Bottom=%d", paddingTop, paddingBottom), true)
			return *m, tickEverySecond()
		}

	case appearanceStatePreview:
		m.state = appearanceStateMenu
	}

	return *m, nil
}

func (m *AppearanceSettingsModel) saveConfig() (AppearanceSettingsModel, tea.Cmd) {
	// Validate before saving
	if err := m.config.Validate(); err != nil {
		m.showMessage(fmt.Sprintf("Validation error: %v", err), false)
		return *m, nil
	}

	// Save config
	if err := storage.SaveAppearanceConfig(m.config); err != nil {
		m.showMessage(fmt.Sprintf("Failed to save: %v", err), false)
		return *m, nil
	}

	// Reload global header renderer
	if globalHeaderRenderer != nil {
		if err := globalHeaderRenderer.Reload(); err != nil {
			m.showMessage("Saved but failed to reload header", false)
			return *m, nil
		}
	}

	m.showMessage("Configuration saved successfully!", true)
	return *m, nil
}

func (m *AppearanceSettingsModel) showMessage(msg string, success bool) {
	m.message = msg
	m.messageSuccess = success
	m.messageTime = 2 // Show for 2 seconds
}

func (m AppearanceSettingsModel) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	var content string

	switch m.state {
	case appearanceStateMenu:
		content = m.menuList.View()

	case appearanceStateModeSelect:
		content = m.modeList.View()

	case appearanceStateTextInput:
		content = m.viewTextInput()

	case appearanceStateAsciiInput:
		content = m.viewAsciiInput()

	case appearanceStateAlignmentSelect:
		content = m.alignmentList.View()

	case appearanceStateWidthInput:
		content = m.viewWidthInput()

	case appearanceStateColorInput:
		content = m.viewColorInput()

	case appearanceStatePaddingInput:
		content = m.viewPaddingInput()

	case appearanceStatePreview:
		content = m.viewPreview()
	}

	// Add message if present
	if m.message != "" {
		var msgStyle lipgloss.Style
		if m.messageSuccess {
			msgStyle = successStyle()
		} else {
			msgStyle = errorStyle()
		}
		content = msgStyle.Render(m.message) + "\n\n" + content
	}

	layout := PageLayout{
		Title:   "Appearance Settings",
		Content: content,
		Help:    "Press ? for help • esc/q to go back",
	}

	// Show help overlay if active
	if m.helpModel.IsVisible() {
		return m.helpModel.View()
	}

	return RenderPageWithBottomHelp(layout, m.height)
}

func (m AppearanceSettingsModel) viewTextInput() string {
	var b strings.Builder
	b.WriteString("Enter custom header text:\n\n")
	b.WriteString(m.textInput.View())
	b.WriteString("\n\nPress Enter to save, Esc to cancel")
	return b.String()
}

func (m AppearanceSettingsModel) viewAsciiInput() string {
	var b strings.Builder
	b.WriteString("Paste or type your ASCII art (max 15 lines):\n\n")
	b.WriteString(m.asciiInput.View())
	b.WriteString("\n\nPress Enter to save, Esc to cancel")
	b.WriteString("\nTip: Use https://patorjk.com/software/taag/ or 'figlet' to generate ASCII art")
	return b.String()
}

func (m AppearanceSettingsModel) viewWidthInput() string {
	var b strings.Builder
	b.WriteString("Enter header width (10-120):\n\n")
	b.WriteString(m.widthInput.View())
	b.WriteString("\n\nPress Enter to save, Esc to cancel")
	return b.String()
}

func (m AppearanceSettingsModel) viewColorInput() string {
	var b strings.Builder
	b.WriteString("Enter header color:\n\n")
	b.WriteString(m.colorInput.View())
	b.WriteString("\n\nExamples: auto, #FF0000, 33 (ANSI code)")
	b.WriteString("\nPress Enter to save, Esc to cancel")
	return b.String()
}

func (m AppearanceSettingsModel) viewPaddingInput() string {
	var b strings.Builder
	b.WriteString("Configure header padding:\n\n")
	
	b.WriteString("Padding Top (0-5):\n")
	b.WriteString(m.paddingTopInput.View())
	b.WriteString("\n\n")
	
	b.WriteString("Padding Bottom (0-5):\n")
	b.WriteString(m.paddingBottomInput.View())
	b.WriteString("\n\n")
	
	b.WriteString("Press Tab to switch between fields\n")
	b.WriteString("Press Enter to save, Esc to cancel")
	return b.String()
}

func (m AppearanceSettingsModel) viewPreview() string {
	var b strings.Builder

	b.WriteString("Preview of your header configuration:\n\n")
	b.WriteString(strings.Repeat("─", 60))
	b.WriteString("\n")

	// Create a temporary renderer with current config
	tempRenderer := &HeaderRenderer{config: m.config}
	preview := tempRenderer.Render()

	if preview == "" {
		b.WriteString("(No header will be shown)\n")
	} else {
		b.WriteString(preview)
	}

	b.WriteString("\n")
	b.WriteString(strings.Repeat("─", 60))
	b.WriteString("\n\nPress Enter to return to menu")

	return b.String()
}
