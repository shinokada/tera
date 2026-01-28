package ui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/shinokada/tera/internal/gist"
)

type gistState int

const (
	gistStateMenu gistState = iota
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
	docStyle          = lipgloss.NewStyle().Margin(1, 2)
	titleStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Bold(true)
	breadcrumbStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Faint(true)
	separatorStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	itemStyle         = lipgloss.NewStyle().PaddingLeft(4)
	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("170"))
	paginationStyle   = list.DefaultStyles().PaginationStyle.PaddingLeft(4)
	helpStyle         = list.DefaultStyles().HelpStyle.PaddingLeft(4).PaddingBottom(1)
	quitTextStyle     = lipgloss.NewStyle().Margin(1, 0, 2, 4)
	errorStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Bold(true)
	successStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Bold(true)
	footerStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Faint(true).MarginTop(1)
)

type item struct {
	title, desc string
	action      gistState
}

func (i item) Title() string       { return i.title }
func (i item) Description() string { return i.desc }
func (i item) FilterValue() string { return i.title }

type gistItem struct {
	meta *gist.GistMetadata
}

func (i gistItem) Title() string       { return i.meta.Description }
func (i gistItem) Description() string { return i.meta.CreatedAt.Format("2006-01-02 15:04") }
func (i gistItem) FilterValue() string { return i.meta.Description }

type GistModel struct {
	state          gistState
	favoritePath   string
	gistClient     *gist.Client
	menuList       list.Model
	gistList       list.Model
	gists          []*gist.GistMetadata
	selectedGist   *gist.GistMetadata
	textInput      textinput.Model
	message        string
	messageIsError bool
	width          int
	height         int
	token          string
	quitting       bool
	inputPurpose   string // "token", "description", "delete"
}

func NewGistModel(favoritePath string) GistModel {
	// Main Menu
	menuItems := []list.Item{
		item{title: "Create a gist", desc: "Upload favorites to a new secret gist", action: gistStateCreate},
		item{title: "My Gists", desc: "View and manage your saved gists", action: gistStateList},
		item{title: "Recover favorites", desc: "Download and restore favorites from a gist", action: gistStateRecover},
		item{title: "Update a gist", desc: "Update description of an existing gist", action: gistStateUpdate},
		item{title: "Delete a gist", desc: "Remove a gist permanently", action: gistStateDelete},
		item{title: "Token Management", desc: "Manage your GitHub Personal Access Token", action: gistStateTokenMenu},
	}

	menuList := list.New(menuItems, list.NewDefaultDelegate(), 0, 0)
	menuList.Title = "TERA Gist Menu"

	// Gist List (initialized empty)
	gistList := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	gistList.Title = "My Gists"

	ti := textinput.New()
	ti.Placeholder = "Type here..."
	ti.Focus()
	ti.CharLimit = 156
	ti.Width = 40

	// Check for token
	token, _ := gist.LoadToken()

	m := GistModel{
		state:        gistStateMenu,
		favoritePath: favoritePath,
		menuList:     menuList,
		gistList:     gistList,
		textInput:    ti,
		token:        token,
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
		if m.state == gistStateMenu || m.state == gistStateList || m.state == gistStateTokenMenu {
			if msg.String() == "ctrl+c" {
				m.quitting = true
				return m, tea.Quit
			}
			if msg.String() == "esc" {
				if m.state == gistStateMenu {
					m.quitting = true
					return m, tea.Quit // Or return to main app menu if this is a submodel
				}
				m.state = gistStateMenu
				m.message = ""
				return m, nil
			}
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.menuList.SetWidth(msg.Width)
		m.gistList.SetWidth(msg.Width)
		return m, nil

	case errMsg:
		m.message = msg.err.Error()
		m.messageIsError = true
		return m, nil

	case successMsg:
		m.message = msg.msg
		m.messageIsError = false
		if m.state == gistStateCreate {
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
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			if keyMsg.String() == "enter" {
				selectedItem, ok := m.menuList.SelectedItem().(item)
				if ok {
					m.state = selectedItem.action
					m.message = ""

					// Initialize specific state logic
					switch m.state {
					case gistStateCreate:
						return m.initCreateGist()
					case gistStateList, gistStateUpdate, gistStateDelete, gistStateRecover:
						return m.initGistList()
					case gistStateTokenMenu:
						return m.initTokenMenu()
					}
					return m, nil
				}
			}
		}
		m.menuList, cmd = m.menuList.Update(msg)
		cmds = append(cmds, cmd)

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
			switch keyMsg.String() {
			case "1": // Setup
				m.state = gistStateTokenSetup
				m.inputPurpose = "token"
				m.textInput.Placeholder = "Paste GitHub Token"
				m.textInput.SetValue("")
				m.textInput.EchoMode = textinput.EchoPassword
				m.textInput.Focus()
				return m, nil
			case "2": // View
				m.state = gistStateTokenView
				return m, nil
			case "4": // Delete
				m.state = gistStateTokenDelete
				m.inputPurpose = "delete_token"
				return m, nil
			case "esc", "0":
				m.state = gistStateMenu
				m.message = ""
				return m, nil
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
	}

	return m, tea.Batch(cmds...)
}

func (m GistModel) View() string {
	if m.quitting {
		return ""
	}

	var title string
	var content string
	breadcrumb := m.renderBreadcrumb()
	footer := m.renderFooter()

	switch m.state {
	case gistStateMenu:
		return docStyle.Render(
			lipgloss.JoinVertical(
				lipgloss.Left,
				titleStyle.Render("TERA Gist Menu"),
				breadcrumb,
				"",
				m.menuList.View(),
				m.renderMessage(),
				footer,
			),
		)
	case gistStateList, gistStateUpdate, gistStateDelete, gistStateRecover:
		return docStyle.Render(
			lipgloss.JoinVertical(
				lipgloss.Left,
				titleStyle.Render(m.getStateTitle()),
				breadcrumb,
				"",
				m.gistList.View(),
				m.renderMessage(),
				footer,
			),
		)
	case gistStateTokenMenu:
		title = "Token Management"
		content = "1) Setup/Change Token\n2) View Current Token\n4) Delete Token\n0) Back"
	case gistStateTokenSetup:
		title = "Setup Token"
		content = fmt.Sprintf("Paste your GitHub Token:\n\n%s", m.textInput.View())
	case gistStateTokenView:
		title = "Current Token"
		masked := gist.GetMaskedToken(m.token)
		if m.token == "" {
			masked = "No token configured"
		}
		content = fmt.Sprintf("Token: %s\n\nPress Enter to back", masked)
	case gistStateUpdateInput:
		title = "Update Gist Description"
		content = fmt.Sprintf("Current: %s\n\nNew Description:\n%s", m.selectedGist.Description, m.textInput.View())
	case gistStateDeleteConfirm:
		title = "Delete Gist"
		content = fmt.Sprintf("Are you sure you want to delete this gist?\n%s\n\nType 'yes' to confirm:\n%s", m.selectedGist.Description, m.textInput.View())
	}

	return docStyle.Render(
		lipgloss.JoinVertical(
			lipgloss.Left,
			titleStyle.Render(title),
			breadcrumb,
			"",
			content,
			m.renderMessage(),
			footer,
		),
	)
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

// renderBreadcrumb generates a navigation breadcrumb based on current state
func (m GistModel) renderBreadcrumb() string {
	var parts []string
	parts = append(parts, "Main Menu")
	parts = append(parts, "Gist")

	switch m.state {
	case gistStateMenu:
		// Just "Main Menu > Gist"
	case gistStateCreate:
		parts = append(parts, "Create")
	case gistStateList:
		parts = append(parts, "My Gists")
	case gistStateUpdate:
		parts = append(parts, "Update")
	case gistStateUpdateInput:
		parts = append(parts, "Update", "Edit Description")
	case gistStateDelete:
		parts = append(parts, "Delete")
	case gistStateDeleteConfirm:
		parts = append(parts, "Delete", "Confirm")
	case gistStateRecover:
		parts = append(parts, "Recover")
	case gistStateTokenMenu:
		parts = append(parts, "Token Management")
	case gistStateTokenSetup:
		parts = append(parts, "Token Management", "Setup")
	case gistStateTokenView:
		parts = append(parts, "Token Management", "View")
	case gistStateTokenDelete:
		parts = append(parts, "Token Management", "Delete")
	}

	// Join with separator
	separator := separatorStyle.Render(" > ")
	breadcrumbParts := make([]string, len(parts))
	for i, part := range parts {
		breadcrumbParts[i] = breadcrumbStyle.Render(part)
	}

	return strings.Join(breadcrumbParts, separator)
}

// renderFooter generates context-sensitive footer with navigation hints
func (m GistModel) renderFooter() string {
	var hints []string

	switch m.state {
	case gistStateMenu:
		hints = append(hints, "Enter: Select", "Esc: Back", "Ctrl+C: Quit")
	case gistStateList, gistStateUpdate, gistStateDelete, gistStateRecover:
		hints = append(hints, "Enter: Select", "Esc: Back to Menu")
	case gistStateTokenMenu:
		hints = append(hints, "1-4: Select Option", "0/Esc: Back")
	case gistStateTokenSetup, gistStateUpdateInput, gistStateDeleteConfirm:
		hints = append(hints, "Enter: Confirm", "Esc: Cancel")
	case gistStateTokenView:
		hints = append(hints, "Enter/Esc: Back")
	default:
		hints = append(hints, "Esc: Back", "Ctrl+C: Quit")
	}

	return footerStyle.Render(strings.Join(hints, " | "))
}

// getStateTitle returns the display title for list-based states
func (m GistModel) getStateTitle() string {
	switch m.state {
	case gistStateList:
		return "My Gists"
	case gistStateUpdate:
		return "Select Gist to Update"
	case gistStateDelete:
		return "Select Gist to Delete"
	case gistStateRecover:
		return "Select Gist to Recover"
	default:
		return "My Gists"
	}
}

// Commands

func (m *GistModel) initCreateGist() (tea.Model, tea.Cmd) {
	if m.token == "" {
		m.message = "No token configured!"
		m.messageIsError = true
		m.state = gistStateMenu
		return m, nil
	}
	m.message = "Creating gist..."
	return m, m.createGistCmd
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

	// 2. Create Gist
	newGist, err := m.gistClient.CreateGist("TERA Radio Favorites Backup", files)
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

	return successMsg{fmt.Sprintf("Gist created! %s", newGist.URL)}
}

func (m *GistModel) initGistList() (tea.Model, tea.Cmd) {
	return m, m.loadGistsCmd
}

func (m *GistModel) loadGistsCmd() tea.Msg {
	gists, err := gist.GetAllGists()
	if err != nil {
		return errMsg{err}
	}
	return gistsMsg(gists)
}

func (m *GistModel) initTokenMenu() (tea.Model, tea.Cmd) {
	return m, nil
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

// Messages

type errMsg struct{ err error }
type successMsg struct{ msg string }
type gistsMsg []*gist.GistMetadata
type tokenSavedMsg struct {
	token, user string
}
