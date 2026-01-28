package ui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/shinokada/tera/internal/gist"
	"github.com/shinokada/tera/internal/ui/components"
)

type gistState int

const (
	gistStateMenu             gistState = iota
	gistStateCreateVisibility           // Ask public or secret
	gistStateCreateName                 // Enter gist name/description
	gistStateCreate
	gistStateList
	gistStateUpdate
	gistStateDelete
	gistStateRecover
	gistStateTokenMenu
	gistStateTokenSetup
	gistStateTokenView
	gistStateTokenDelete
	gistStateUpdateInput
	gistStateDeleteConfirm
)

var (
	quitTextStyle = lipgloss.NewStyle().Margin(1, 0, 2, 4)
)

type gistItem struct {
	meta *gist.GistMetadata
}

func (i gistItem) Title() string       { return i.meta.Description }
func (i gistItem) Description() string { return i.meta.CreatedAt.Format("2006-01-02 15:04") }
func (i gistItem) FilterValue() string { return i.meta.Description }

type GistModel struct {
	state           gistState
	favoritePath    string
	gistClient      *gist.Client
	menuList        list.Model
	tokenMenuList   list.Model
	visibilityMenu  list.Model
	gistList        list.Model
	gists           []*gist.GistMetadata
	selectedGist    *gist.GistMetadata
	textInput       textinput.Model
	message         string
	messageIsError  bool
	width           int
	height          int
	token           string
	quitting        bool
	inputPurpose    string // "token", "description", "delete"
	createPublic    bool   // true for public gist, false for secret
	gistDescription string // custom gist description/name
}

func NewGistModel(favoritePath string) GistModel {
	// Main Menu
	menuItems := []components.MenuItem{
		components.NewMenuItem("Create a gist", "Upload favorites to a new secret gist", "1"),
		components.NewMenuItem("My Gists", "View and manage your saved gists", "2"),
		components.NewMenuItem("Recover favorites", "Download and restore favorites from a gist", "3"),
		components.NewMenuItem("Update a gist", "Update description of an existing gist", "4"),
		components.NewMenuItem("Delete a gist", "Remove a gist permanently", "5"),
		components.NewMenuItem("Token Management", "Manage your GitHub Personal Access Token", "6"),
	}

	menuList := components.CreateMenu(menuItems, "TERA Gist Menu", 0, 0)

	// Token Management Menu
	tokenMenuItems := []components.MenuItem{
		components.NewMenuItem("Setup/Change Token", "Configure your GitHub Personal Access Token", "1"),
		components.NewMenuItem("View Current Token", "See your masked token", "2"),
		components.NewMenuItem("Validate Token", "Test if your token is valid", "3"),
		components.NewMenuItem("Delete Token", "Remove your stored token", "4"),
	}
	tokenMenuList := components.CreateMenu(tokenMenuItems, "Token Menu", 0, 0)

	// Visibility Menu for create gist
	visibilityMenuItems := []components.MenuItem{
		components.NewMenuItem("Secret gist", "Only you can see this gist (recommended)", "1"),
		components.NewMenuItem("Public gist", "Anyone can see this gist", "2"),
	}
	visibilityMenu := components.CreateMenu(visibilityMenuItems, "Visibility", 0, 0)

	// Gist List (initialized empty)
	// Use styled delegate for consistency
	delegate := createStyledDelegate()
	gistList := list.New([]list.Item{}, delegate, 50, 20) // Default size, will be updated on WindowSizeMsg
	gistList.Title = "My Gists"
	gistList.SetShowStatusBar(false)
	gistList.SetShowHelp(false)
	gistList.SetShowTitle(false)
	gistList.SetShowPagination(false)

	ti := textinput.New()
	ti.Placeholder = "Type here..."
	ti.Focus()
	ti.CharLimit = 156
	ti.Width = 40

	// Check for token
	token, _ := gist.LoadToken()

	m := GistModel{
		state:          gistStateMenu,
		favoritePath:   favoritePath,
		menuList:       menuList,
		tokenMenuList:  tokenMenuList,
		visibilityMenu: visibilityMenu,
		gistList:       gistList,
		textInput:      ti,
		token:          token,
	}

	if token != "" {
		m.gistClient = gist.NewClient(token)
	}

	return m
}

func (m GistModel) Init() tea.Cmd {
	return nil
}

func (m GistModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Handle Ctrl+C globally
		if msg.String() == "ctrl+c" {
			m.quitting = true
			return m, tea.Quit
		}

		// Handle ESC and 0 based on state
		if m.state == gistStateMenu {
			if msg.String() == "esc" || msg.String() == "0" {
				// Go back to main app menu
				m.quitting = true
				return m, func() tea.Msg { return backToMainMsg{} }
			}
		} else if m.state == gistStateCreateVisibility {
			if msg.String() == "esc" {
				m.state = gistStateMenu
				m.message = ""
				return m, nil
			}
		} else if m.state == gistStateList || m.state == gistStateUpdate || m.state == gistStateDelete || m.state == gistStateRecover || m.state == gistStateTokenMenu {
			if msg.String() == "esc" {
				m.state = gistStateMenu
				m.message = ""
				return m, nil
			}
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.menuList.SetWidth(msg.Width)
		m.tokenMenuList.SetWidth(msg.Width)
		m.visibilityMenu.SetWidth(msg.Width)
		m.gistList.SetWidth(msg.Width)
		m.gistList.SetHeight(msg.Height - 10) // Leave room for header and footer
		return m, nil

	case errMsg:
		m.message = msg.err.Error()
		m.messageIsError = true
		// If we're in a gist list state or create state and get an error, go back to menu
		if m.state == gistStateCreate || m.state == gistStateList || m.state == gistStateUpdate || m.state == gistStateDelete || m.state == gistStateRecover {
			m.state = gistStateMenu
		}
		return m, nil

	case successMsg:
		m.message = msg.msg
		m.messageIsError = false
		// After successful create/update/delete/recover, go back to menu
		if m.state == gistStateCreate || m.state == gistStateRecover || m.state == gistStateUpdateInput || m.state == gistStateDeleteConfirm {
			m.state = gistStateMenu
		}
		return m, nil

	case gistsMsg:
		m.gists = msg
		items := make([]list.Item, len(m.gists))
		for i, g := range m.gists {
			items[i] = gistItem{meta: g}
		}
		m.gistList.SetItems(items)
		// Ensure list has proper height
		if m.height > 0 {
			m.gistList.SetHeight(m.height - 10)
		} else {
			m.gistList.SetHeight(20) // Default height
		}
		return m, nil

	case tokenSavedMsg:
		m.token = msg.token
		m.gistClient = gist.NewClient(m.token)
		m.message = fmt.Sprintf("Token saved! User: %s", msg.user)
		m.messageIsError = false
		m.state = gistStateTokenMenu
		return m, nil
	}

	switch m.state {
	case gistStateMenu:
		// Always update the menu list first to handle navigation
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			var selected int
			m.menuList, selected = components.HandleMenuKey(keyMsg, m.menuList)

			// If an item was selected (Enter or number key pressed)
			if selected >= 0 {
				// Clear any previous messages
				m.message = ""
				m.messageIsError = false

				switch selected {
				case 0: // Create - go to visibility selection first
					if m.token == "" {
						m.message = "No token configured!"
						m.messageIsError = true
						return m, nil
					}
					m.state = gistStateCreateVisibility
					return m, nil
				case 1: // My Gists
					m.state = gistStateList
					return m.initGistList()
				case 2: // Recover
					m.state = gistStateRecover
					return m.initGistList()
				case 3: // Update
					m.state = gistStateUpdate
					return m.initGistList()
				case 4: // Delete
					m.state = gistStateDelete
					return m.initGistList()
				case 5: // Token Management
					m.state = gistStateTokenMenu
					return m.initTokenMenu()
				}
			}
		}
		return m, nil

	case gistStateCreateVisibility:
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			// Handle ESC to go back
			if keyMsg.String() == "esc" {
				m.state = gistStateMenu
				return m, nil
			}

			var selected int
			m.visibilityMenu, selected = components.HandleMenuKey(keyMsg, m.visibilityMenu)

			if selected >= 0 {
				switch selected {
				case 0: // Secret gist
					m.createPublic = false
				case 1: // Public gist
					m.createPublic = true
				}
				// Go to name input
				m.state = gistStateCreateName
				m.inputPurpose = "gist_name"
				m.textInput.Placeholder = "Enter gist name (or press Enter for default)"
				timestamp := time.Now().Format("2006-01-02 15:04:05")
				m.textInput.SetValue(fmt.Sprintf("TERA Radio Favorites - %s", timestamp))
				m.textInput.EchoMode = textinput.EchoNormal
				m.textInput.Focus()
				return m, nil
			}
		}
		return m, nil

	case gistStateCreateName:
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			if keyMsg.String() == "enter" {
				m.gistDescription = m.textInput.Value()
				if m.gistDescription == "" {
					timestamp := time.Now().Format("2006-01-02 15:04:05")
					m.gistDescription = fmt.Sprintf("TERA Radio Favorites - %s", timestamp)
				}
				m.state = gistStateCreate
				return m.initCreateGist()
			} else if keyMsg.String() == "esc" {
				m.state = gistStateCreateVisibility
				return m, nil
			}
		}
		m.textInput, cmd = m.textInput.Update(msg)
		cmds = append(cmds, cmd)
		return m, tea.Batch(cmds...)

	case gistStateCreate:
		// This state is transient - handled by initCreateGist
		return m, nil

	case gistStateList, gistStateUpdate, gistStateDelete, gistStateRecover:
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			if keyMsg.String() == "enter" {
				selectedItem, ok := m.gistList.SelectedItem().(gistItem)
				if ok {
					m.selectedGist = selectedItem.meta
					switch m.state {
					case gistStateList:
						// Open in browser
						// For now just show message
						m.message = fmt.Sprintf("Opened %s in browser", m.selectedGist.URL)
						m.messageIsError = false
						return m, nil
					case gistStateUpdate:
						m.inputPurpose = "description"
						m.textInput.Placeholder = "New description"
						m.textInput.SetValue(m.selectedGist.Description)
						m.textInput.EchoMode = textinput.EchoNormal
						m.state = gistStateUpdateInput
						m.textInput.Focus()
						return m, nil
					case gistStateDelete:
						m.inputPurpose = "delete"
						m.textInput.Placeholder = "Type 'yes' to confirm"
						m.textInput.SetValue("")
						m.textInput.EchoMode = textinput.EchoNormal
						m.state = gistStateDeleteConfirm
						m.textInput.Focus()
						return m, nil
					case gistStateRecover:
						return m, m.recoverGistCmd(m.selectedGist.ID)
					}
				}
			}
		}
		m.gistList, cmd = m.gistList.Update(msg)
		cmds = append(cmds, cmd)

	case gistStateUpdateInput:
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			if keyMsg.String() == "enter" {
				return m, m.updateGistCmd(m.selectedGist.ID, m.textInput.Value())
			} else if keyMsg.String() == "esc" {
				m.state = gistStateUpdate // Back to list
				return m, nil
			}
		}
		m.textInput, cmd = m.textInput.Update(msg)
		cmds = append(cmds, cmd)

	case gistStateDeleteConfirm:
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			if keyMsg.String() == "enter" {
				if strings.ToLower(m.textInput.Value()) == "yes" {
					return m, m.deleteGistCmd(m.selectedGist.ID)
				}
				m.message = "Deletion cancelled"
				m.messageIsError = true
				m.state = gistStateMenu
				return m, nil
			} else if keyMsg.String() == "esc" {
				m.state = gistStateDelete // Back to list
				return m, nil
			}
		}
		m.textInput, cmd = m.textInput.Update(msg)
		cmds = append(cmds, cmd)

	case gistStateTokenMenu:
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			// Handle ESC to go back
			if keyMsg.String() == "esc" {
				m.state = gistStateMenu
				m.message = ""
				return m, nil
			}

			var selected int
			m.tokenMenuList, selected = components.HandleMenuKey(keyMsg, m.tokenMenuList)

			if selected >= 0 {
				m.message = ""
				m.messageIsError = false

				switch selected {
				case 0: // Setup
					m.state = gistStateTokenSetup
					m.inputPurpose = "token"
					m.textInput.Placeholder = "Paste GitHub Token"
					m.textInput.SetValue("")
					m.textInput.EchoMode = textinput.EchoPassword
					m.textInput.Focus()
					return m, nil
				case 1: // View
					m.state = gistStateTokenView
					return m, nil
				case 2: // Validate
					if m.token == "" {
						m.message = "No token configured!"
						m.messageIsError = true
						return m, nil
					}
					return m, m.validateTokenCmd()
				case 3: // Delete
					if m.token == "" {
						m.message = "No token to delete!"
						m.messageIsError = true
						return m, nil
					}
					m.state = gistStateTokenDelete
					m.inputPurpose = "delete_token"
					m.textInput.Placeholder = "Type 'yes' to confirm"
					m.textInput.SetValue("")
					m.textInput.EchoMode = textinput.EchoNormal
					m.textInput.Focus()
					return m, nil
				}
			}
		}

	case gistStateTokenSetup:
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			if keyMsg.String() == "enter" {
				token := m.textInput.Value()
				if token != "" {
					return m, m.saveTokenCmd(token)
				}
			} else if keyMsg.String() == "esc" {
				m.state = gistStateTokenMenu
				return m, nil
			}
		}
		m.textInput, cmd = m.textInput.Update(msg)
		cmds = append(cmds, cmd)

	case gistStateTokenView:
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			if keyMsg.String() == "enter" || keyMsg.String() == "esc" {
				m.state = gistStateTokenMenu
				return m, nil
			}
		}

	case gistStateTokenDelete:
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			if keyMsg.String() == "enter" {
				if strings.ToLower(m.textInput.Value()) == "yes" {
					return m, m.deleteTokenCmd()
				}
				m.message = "Token deletion cancelled"
				m.messageIsError = false
				m.state = gistStateTokenMenu
				return m, nil
			} else if keyMsg.String() == "esc" {
				m.state = gistStateTokenMenu
				return m, nil
			}
		}
		m.textInput, cmd = m.textInput.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m GistModel) View() string {
	if m.quitting {
		return ""
	}

	switch m.state {
	case gistStateMenu:
		return RenderPage(PageLayout{
			Title:    "Gist Management",
			Subtitle: "Select an Option",
			Content:  m.menuList.View() + "\n" + m.renderMessage(),
			Help:     "↑↓/jk: Navigate • Enter: Select • 1-6: Quick select • Esc: Back • Ctrl+C: Quit",
		})
	case gistStateCreateVisibility:
		return RenderPage(PageLayout{
			Title:    "Create Gist",
			Subtitle: "Choose Visibility",
			Content:  m.visibilityMenu.View() + "\n" + m.renderMessage(),
			Help:     "↑↓/jk: Navigate • Enter: Select • 1-2: Quick select • Esc: Back",
		})
	case gistStateCreateName:
		visibility := "Secret"
		if m.createPublic {
			visibility = "Public"
		}
		return RenderPage(PageLayout{
			Title:    "Create Gist",
			Subtitle: fmt.Sprintf("Enter Gist Name (%s)", visibility),
			Content:  fmt.Sprintf("Enter a name/description for your gist:\n\n%s", m.textInput.View()) + "\n\n" + m.renderMessage(),
			Help:     "Enter: Create • Esc: Back",
		})
	case gistStateCreate:
		return RenderPage(PageLayout{
			Title:    "Create Gist",
			Subtitle: "Uploading favorites to GitHub",
			Content:  "Creating gist...\n\nPlease wait while your favorites are uploaded.\n\n" + m.renderMessage(),
			Help:     "Please wait...",
		})
	case gistStateList, gistStateUpdate, gistStateDelete, gistStateRecover:
		action := "My Gists"
		switch m.state {
		case gistStateUpdate:
			action = "Update Gist"
		case gistStateDelete:
			action = "Delete Gist"
		case gistStateRecover:
			action = "Recover from Gist"
		}

		content := m.gistList.View()
		if len(m.gists) == 0 {
			content = "No gists available.\n\nCreate a gist first from the main menu."
		}

		return RenderPage(PageLayout{
			Title:    action,
			Subtitle: "Select a Gist",
			Content:  content + "\n" + m.renderMessage(),
			Help:     "↑↓/jk: Navigate • Enter: Select • Esc: Back",
		})
	case gistStateTokenMenu:
		return RenderPage(PageLayout{
			Title:    "Token Management",
			Subtitle: "Manage your GitHub Token",
			Content:  m.tokenMenuList.View() + "\n" + m.renderMessage(),
			Help:     "↑↓/jk: Navigate • Enter: Select • 1-4: Quick select • Esc: Back • Ctrl+C: Quit",
		})
	case gistStateTokenSetup:
		return RenderPage(PageLayout{
			Title:    "Setup Token",
			Subtitle: "Paste your GitHub Token",
			Content:  fmt.Sprintf("Token will be hidden for security.\n\n%s", m.textInput.View()) + "\n\n" + m.renderMessage(),
			Help:     "Enter: Save • Esc: Cancel",
		})
	case gistStateTokenView:
		masked := gist.GetMaskedToken(m.token)
		if m.token == "" {
			masked = "No token configured"
		}
		return RenderPage(PageLayout{
			Title:    "Current Token",
			Subtitle: "View Token Status",
			Content:  fmt.Sprintf("Token: %s", masked) + "\n\n" + m.renderMessage(),
			Help:     "Enter/Esc: Back",
		})
	case gistStateUpdateInput:
		return RenderPage(PageLayout{
			Title:    "Update Gist",
			Subtitle: "Enter new description",
			Content:  fmt.Sprintf("Current: %s\n\nNew Description:\n%s", m.selectedGist.Description, m.textInput.View()) + "\n\n" + m.renderMessage(),
			Help:     "Enter: Update • Esc: Cancel",
		})
	case gistStateDeleteConfirm:
		return RenderPage(PageLayout{
			Title:    "Delete Gist",
			Subtitle: "Confirm Deletion",
			Content:  fmt.Sprintf("Are you sure you want to delete this gist?\n%s\n\nType 'yes' to confirm:\n%s", m.selectedGist.Description, m.textInput.View()) + "\n\n" + m.renderMessage(),
			Help:     "Enter: Confirm • Esc: Cancel",
		})
	case gistStateTokenDelete:
		masked := gist.GetMaskedToken(m.token)
		return RenderPage(PageLayout{
			Title:    "Delete Token",
			Subtitle: "⚠️  WARNING",
			Content:  fmt.Sprintf("This will delete your stored GitHub token!\n\nToken: %s\n\nYou won't be able to use Gist features until you set up a new token.\n\nType 'yes' to confirm deletion:\n%s", masked, m.textInput.View()) + "\n\n" + m.renderMessage(),
			Help:     "Enter: Confirm • Esc: Cancel",
		})
	}

	return ""
}

func (m GistModel) renderMessage() string {
	if m.message == "" {
		return ""
	}
	if m.messageIsError {
		return errorStyle.Render(m.message)
	}
	return successStyle.Render(m.message)
}

// Commands

func (m *GistModel) initCreateGist() (tea.Model, tea.Cmd) {
	if m.token == "" {
		m.message = "No token configured!"
		m.messageIsError = true
		m.state = gistStateMenu
		return *m, nil
	}
	m.message = "Creating gist..."
	return *m, m.createGistCmd
}

func (m *GistModel) createGistCmd() tea.Msg {
	// 1. Read files
	files := make(map[string]string)
	entries, err := os.ReadDir(m.favoritePath)
	if err != nil {
		return errMsg{err}
	}

	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".json") {
			content, err := os.ReadFile(filepath.Join(m.favoritePath, entry.Name()))
			if err == nil {
				files[entry.Name()] = string(content)
			}
		}
	}

	if len(files) == 0 {
		return errMsg{fmt.Errorf("no favorite lists found in %s", m.favoritePath)}
	}

	// 2. Create Gist with user-provided description
	description := m.gistDescription
	if description == "" {
		timestamp := time.Now().Format("2006-01-02 15:04:05")
		description = fmt.Sprintf("TERA Radio Favorites - %s", timestamp)
	}

	newGist, err := m.gistClient.CreateGist(description, files, m.createPublic)
	if err != nil {
		return errMsg{err}
	}

	// 3. Save Metadata
	meta := &gist.GistMetadata{
		ID:          newGist.ID,
		URL:         newGist.URL,
		Description: newGist.Description,
		CreatedAt:   newGist.CreatedAt,
		UpdatedAt:   newGist.UpdatedAt,
	}
	if err := gist.SaveMetadata(meta); err != nil {
		return errMsg{err}
	}

	visibility := "secret"
	if m.createPublic {
		visibility = "public"
	}
	return successMsg{fmt.Sprintf("Gist created (%s)! %s", visibility, newGist.URL)}
}

func (m *GistModel) initGistList() (tea.Model, tea.Cmd) {
	// Check if token is configured
	if m.token == "" || m.gistClient == nil {
		m.message = "No GitHub token configured! Please set up a token first."
		m.messageIsError = true
		m.state = gistStateMenu
		return *m, nil
	}
	// Return a command that loads gists
	return *m, func() tea.Msg {
		gists, err := gist.GetAllGists()
		if err != nil {
			return errMsg{err}
		}
		if len(gists) == 0 {
			return errMsg{fmt.Errorf("no gists found - create a gist first")}
		}
		return gistsMsg(gists)
	}
}

func (m *GistModel) initTokenMenu() (tea.Model, tea.Cmd) {
	return *m, nil
}

func (m *GistModel) saveTokenCmd(token string) tea.Cmd {
	return func() tea.Msg {
		// Validate
		client := gist.NewClient(token)
		user, err := client.ValidateToken()
		if err != nil {
			return errMsg{fmt.Errorf("invalid token: %v", err)}
		}

		if err := gist.SaveToken(token); err != nil {
			return errMsg{err}
		}

		return tokenSavedMsg{token, user}
	}
}

func (m *GistModel) recoverGistCmd(gistID string) tea.Cmd {
	return func() tea.Msg {
		g, err := m.gistClient.GetGist(gistID)
		if err != nil {
			return errMsg{err}
		}

		// Restore files
		for filename, file := range g.Files {
			// Basic validation to prevent directory traversal
			if strings.Contains(filename, "..") || strings.Contains(filename, "/") {
				continue
			}
			path := filepath.Join(m.favoritePath, filename)
			if err := os.WriteFile(path, []byte(file.Content), 0644); err != nil {
				return errMsg{err}
			}
		}

		return successMsg{"Favorites restored successfully!"}
	}
}

func (m *GistModel) updateGistCmd(id, description string) tea.Cmd {
	return func() tea.Msg {
		if err := m.gistClient.UpdateGist(id, description); err != nil {
			return errMsg{err}
		}
		if err := gist.UpdateMetadata(id, description); err != nil {
			// Don't fail UI if local update fails, but warn?
			// ignoring for now or could return errMsg
		}
		return successMsg{"Gist updated!"}
	}
}

func (m *GistModel) validateTokenCmd() tea.Cmd {
	return func() tea.Msg {
		if m.token == "" {
			return errMsg{fmt.Errorf("no token configured")}
		}

		client := gist.NewClient(m.token)
		user, err := client.ValidateToken()
		if err != nil {
			return errMsg{fmt.Errorf("token validation failed: %v", err)}
		}

		return successMsg{fmt.Sprintf("✓ Token is VALID! GitHub user: %s", user)}
	}
}

func (m *GistModel) deleteGistCmd(id string) tea.Cmd {
	return func() tea.Msg {
		if err := m.gistClient.DeleteGist(id); err != nil {
			return errMsg{err}
		}
		if err := gist.DeleteMetadata(id); err != nil {
			// ignoring
		}
		return successMsg{"Gist deleted!"}
	}
}

func (m *GistModel) deleteTokenCmd() tea.Cmd {
	return func() tea.Msg {
		if err := gist.DeleteToken(); err != nil {
			return errMsg{fmt.Errorf("failed to delete token: %v", err)}
		}
		// Clear the token from memory
		m.token = ""
		m.gistClient = nil
		return successMsg{"✓ Token has been deleted successfully!"}
	}
}

// Messages

type successMsg struct{ msg string }
type gistsMsg []*gist.GistMetadata
type tokenSavedMsg struct {
	token, user string
}
