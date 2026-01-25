package ui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// listManagementState represents the current state in the list management menu
type listManagementState int

const (
	listManagementMenu listManagementState = iota
	listManagementCreate
	listManagementDelete
	listManagementEdit
	listManagementShowAll
	listManagementConfirmDelete
	listManagementEnterNewName
)

// ListManagementModel represents the list management screen
type ListManagementModel struct {
	state        listManagementState
	favoritePath string
	lists        []string
	listItems    []list.Item
	listModel    list.Model
	textInput    textinput.Model
	selectedList string
	newListName  string
	err          error
	message      string
	messageTime  int
	width        int
	height       int
}

// listManagementMenuItem wraps a menu item
type listManagementMenuItem struct {
	title       string
	description string
}

func (i listManagementMenuItem) FilterValue() string { return i.title }
func (i listManagementMenuItem) Title() string       { return i.title }
func (i listManagementMenuItem) Description() string { return i.description }

// NewListManagementModel creates a new list management model
func NewListManagementModel(favoritePath string) ListManagementModel {
	ti := textinput.New()
	ti.Placeholder = "Enter list name"
	ti.CharLimit = 50

	items := []list.Item{
		listManagementMenuItem{title: "Create New List", description: "Create a new favorites list"},
		listManagementMenuItem{title: "Delete List", description: "Delete an existing list"},
		listManagementMenuItem{title: "Edit List Name", description: "Rename an existing list"},
		listManagementMenuItem{title: "Show All Lists", description: "Display all favorite lists"},
	}

	delegate := list.NewDefaultDelegate()
	l := list.New(items, delegate, 0, 0)
	l.Title = "📋 List Management"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Styles.Title = titleStyle
	l.Styles.PaginationStyle = paginationStyle
	l.Styles.HelpStyle = helpStyle

	return ListManagementModel{
		state:        listManagementMenu,
		favoritePath: favoritePath,
		listModel:    l,
		textInput:    ti,
	}
}

// Init initializes the list management screen
func (m ListManagementModel) Init() tea.Cmd {
	return m.loadLists()
}

// loadLists loads all available lists
func (m ListManagementModel) loadLists() tea.Cmd {
	return func() tea.Msg {
		entries, err := os.ReadDir(m.favoritePath)
		if err != nil {
			return errMsg{err}
		}

		var lists []string
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			name := entry.Name()
			if strings.HasSuffix(name, ".json") {
				listName := strings.TrimSuffix(name, ".json")
				lists = append(lists, listName)
			}
		}

		return listManagementListsLoadedMsg{lists}
	}
}

// Update handles messages for the list management screen
func (m ListManagementModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
		// Ensure enough height for menu items (4 items + title + pagination + help)
		h := msg.Height - 8
		if h < 10 {
			h = 10
		}
		m.listModel.SetSize(msg.Width-4, h)
		return m, nil

	case listManagementListsLoadedMsg:
		m.lists = msg.lists
		return m, nil

	case listManagementOperationSuccessMsg:
		m.message = msg.message
		m.messageTime = 150 // ~3 seconds
		m.state = listManagementMenu
		return m, m.loadLists()

	case listManagementOperationErrorMsg:
		m.err = msg.err
		m.message = msg.err.Error()
		m.messageTime = 150
		return m, nil

	case errMsg:
		m.err = msg.err
		return m, nil
	}

	var cmd tea.Cmd
	switch m.state {
	case listManagementMenu:
		m.listModel, cmd = m.listModel.Update(msg)
	case listManagementCreate, listManagementDelete, listManagementEdit, listManagementEnterNewName:
		m.textInput, cmd = m.textInput.Update(msg)
	}
	return m, cmd
}

// handleKeyPress handles keyboard input based on current state
func (m ListManagementModel) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.state {
	case listManagementMenu:
		return m.handleMenuInput(msg)
	case listManagementCreate:
		return m.handleCreateInput(msg)
	case listManagementDelete:
		return m.handleDeleteInput(msg)
	case listManagementEdit:
		return m.handleEditInput(msg)
	case listManagementShowAll:
		return m.handleShowAllInput(msg)
	case listManagementConfirmDelete:
		return m.handleConfirmDeleteInput(msg)
	case listManagementEnterNewName:
		return m.handleEnterNewNameInput(msg)
	}
	return m, nil
}

// handleMenuInput handles input on the main menu
func (m ListManagementModel) handleMenuInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		return m, func() tea.Msg {
			return navigateMsg{screen: screenMainMenu}
		}
	case "q":
		return m, tea.Quit
	case "enter":
		// Get selected item index
		idx := m.listModel.Index()
		return m.executeMenuAction(idx)
	case "1":
		// Create new list
		return m.executeMenuAction(0)
	case "2":
		// Delete list
		return m.executeMenuAction(1)
	case "3":
		// Edit list name
		return m.executeMenuAction(2)
	case "4":
		// Show all lists
		return m.executeMenuAction(3)
	}

	var cmd tea.Cmd
	m.listModel, cmd = m.listModel.Update(msg)
	return m, cmd
}

// executeMenuAction executes the selected menu action
func (m ListManagementModel) executeMenuAction(index int) (tea.Model, tea.Cmd) {
	switch index {
	case 0: // Create new list
		m.state = listManagementCreate
		m.textInput.Reset()
		m.textInput.Placeholder = "Enter new list name"
		m.textInput.Focus()
		return m, tea.Batch(m.loadLists(), textinput.Blink)
	case 1: // Delete list
		if len(m.lists) == 0 {
			m.message = "No lists available to delete"
			m.messageTime = 150
			return m, nil
		}
		m.state = listManagementDelete
		m.textInput.Reset()
		m.textInput.Placeholder = "Enter list name to delete"
		m.textInput.Focus()
		return m, tea.Batch(m.loadLists(), textinput.Blink)
	case 2: // Edit list name
		if len(m.lists) == 0 {
			m.message = "No lists available to edit"
			m.messageTime = 150
			return m, nil
		}
		m.state = listManagementEdit
		m.textInput.Reset()
		m.textInput.Placeholder = "Enter list name to rename"
		m.textInput.Focus()
		return m, tea.Batch(m.loadLists(), textinput.Blink)
	case 3: // Show all lists
		m.state = listManagementShowAll
		return m, m.loadLists()
	}
	return m, nil
}

// handleCreateInput handles input during list creation
func (m ListManagementModel) handleCreateInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.state = listManagementMenu
		m.textInput.Blur()
		return m, nil
	case "enter":
		name := strings.TrimSpace(m.textInput.Value())
		if name == "" {
			m.message = "List name cannot be empty"
			m.messageTime = 150
			return m, nil
		}

		// Replace spaces with hyphens
		name = strings.ReplaceAll(name, " ", "-")

		// Check if list exists
		for _, existing := range m.lists {
			if existing == name {
				m.message = fmt.Sprintf("List '%s' already exists", name)
				m.messageTime = 150
				return m, nil
			}
		}

		// Create new list file
		return m, m.createList(name)
	}

	var cmd tea.Cmd
	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

// handleDeleteInput handles input during list deletion
func (m ListManagementModel) handleDeleteInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.state = listManagementMenu
		m.textInput.Blur()
		return m, nil
	case "enter":
		name := strings.TrimSpace(m.textInput.Value())
		if name == "" {
			m.message = "List name cannot be empty"
			m.messageTime = 150
			return m, nil
		}

		// Check if it's My-favorites (protected)
		if name == "My-favorites" {
			m.message = "Cannot delete My-favorites (protected list)"
			m.messageTime = 150
			return m, nil
		}

		// Check if list exists
		found := false
		for _, existing := range m.lists {
			if existing == name {
				found = true
				break
			}
		}

		if !found {
			m.message = fmt.Sprintf("List '%s' does not exist", name)
			m.messageTime = 150
			return m, nil
		}

		// Move to confirmation
		m.selectedList = name
		m.state = listManagementConfirmDelete
		m.textInput.Blur()
		return m, nil
	}

	var cmd tea.Cmd
	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

// handleConfirmDeleteInput handles delete confirmation
func (m ListManagementModel) handleConfirmDeleteInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "n":
		m.state = listManagementMenu
		m.selectedList = ""
		return m, nil
	case "y":
		return m, m.deleteList(m.selectedList)
	}
	return m, nil
}

// handleEditInput handles input during list editing
func (m ListManagementModel) handleEditInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.state = listManagementMenu
		m.textInput.Blur()
		return m, nil
	case "enter":
		name := strings.TrimSpace(m.textInput.Value())
		if name == "" {
			m.message = "List name cannot be empty"
			m.messageTime = 150
			return m, nil
		}

		// Check if it's My-favorites (protected)
		if name == "My-favorites" {
			m.message = "Cannot rename My-favorites (protected list)"
			m.messageTime = 150
			return m, nil
		}

		// Check if list exists
		found := false
		for _, existing := range m.lists {
			if existing == name {
				found = true
				break
			}
		}

		if !found {
			m.message = fmt.Sprintf("List '%s' does not exist", name)
			m.messageTime = 150
			return m, nil
		}

		// Move to new name input
		m.selectedList = name
		m.state = listManagementEnterNewName
		m.textInput.Reset()
		m.textInput.Placeholder = "Enter new name"
		m.textInput.Focus()
		return m, textinput.Blink
	}

	var cmd tea.Cmd
	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

// handleEnterNewNameInput handles input for new list name
func (m ListManagementModel) handleEnterNewNameInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.state = listManagementMenu
		m.textInput.Blur()
		m.selectedList = ""
		return m, nil
	case "enter":
		newName := strings.TrimSpace(m.textInput.Value())
		if newName == "" {
			m.message = "List name cannot be empty"
			m.messageTime = 150
			return m, nil
		}

		// Replace spaces with hyphens
		newName = strings.ReplaceAll(newName, " ", "-")

		// Check if new name already exists
		for _, existing := range m.lists {
			if existing == newName {
				m.message = fmt.Sprintf("List '%s' already exists", newName)
				m.messageTime = 150
				return m, nil
			}
		}

		// Rename list
		return m, m.renameList(m.selectedList, newName)
	}

	var cmd tea.Cmd
	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

// handleShowAllInput handles input when showing all lists
func (m ListManagementModel) handleShowAllInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter", "esc":
		m.state = listManagementMenu
		return m, nil
	}
	return m, nil
}

// createList creates a new list file
func (m ListManagementModel) createList(name string) tea.Cmd {
	return func() tea.Msg {
		path := filepath.Join(m.favoritePath, name+".json")

		// Create empty JSON array
		if err := os.WriteFile(path, []byte("[]"), 0644); err != nil {
			return listManagementOperationErrorMsg{err}
		}

		return listManagementOperationSuccessMsg{
			message: fmt.Sprintf("✓ Created list '%s'", name),
		}
	}
}

// deleteList deletes a list file
func (m ListManagementModel) deleteList(name string) tea.Cmd {
	return func() tea.Msg {
		path := filepath.Join(m.favoritePath, name+".json")

		if err := os.Remove(path); err != nil {
			return listManagementOperationErrorMsg{err}
		}

		return listManagementOperationSuccessMsg{
			message: fmt.Sprintf("✓ Deleted list '%s'", name),
		}
	}
}

// renameList renames a list file
func (m ListManagementModel) renameList(oldName, newName string) tea.Cmd {
	return func() tea.Msg {
		oldPath := filepath.Join(m.favoritePath, oldName+".json")
		newPath := filepath.Join(m.favoritePath, newName+".json")

		if err := os.Rename(oldPath, newPath); err != nil {
			return listManagementOperationErrorMsg{err}
		}

		return listManagementOperationSuccessMsg{
			message: fmt.Sprintf("✓ Renamed '%s' to '%s'", oldName, newName),
		}
	}
}

// View renders the list management screen
func (m ListManagementModel) View() string {
	switch m.state {
	case listManagementMenu:
		return m.viewMenu()
	case listManagementCreate:
		return m.viewCreate()
	case listManagementDelete:
		return m.viewDelete()
	case listManagementEdit:
		return m.viewEdit()
	case listManagementShowAll:
		return m.viewShowAll()
	case listManagementConfirmDelete:
		return m.viewConfirmDelete()
	case listManagementEnterNewName:
		return m.viewEnterNewName()
	}
	return ""
}

// viewMenu renders the main menu
func (m ListManagementModel) viewMenu() string {
	var b strings.Builder

	if m.message != "" {
		style := successStyle
		if strings.Contains(m.message, "✗") || m.err != nil {
			style = errorStyle
		}
		b.WriteString(style.Render(m.message))
		b.WriteString("\n\n")
	}

	b.WriteString(m.listModel.View())
	b.WriteString("\n\n")

	help := helpStyle.Render("↑↓/jk: navigate • enter: select • 1-4: quick select • esc: back • q: quit")
	b.WriteString(help)

	return b.String()
}

// viewCreate renders the create list view
func (m ListManagementModel) viewCreate() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("Create New List"))
	b.WriteString("\n\n")

	if len(m.lists) > 0 {
		b.WriteString(subtitleStyle.Render("Current lists:"))
		b.WriteString("\n")
		for _, list := range m.lists {
			b.WriteString(fmt.Sprintf("  • %s\n", list))
		}
		b.WriteString("\n")
	}

	b.WriteString(m.textInput.View())
	b.WriteString("\n\n")

	if m.message != "" {
		b.WriteString(errorStyle.Render(m.message))
		b.WriteString("\n\n")
	}

	help := helpStyle.Render("enter: create • esc: cancel")
	b.WriteString(help)

	return b.String()
}

// viewDelete renders the delete list view
func (m ListManagementModel) viewDelete() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("Delete List"))
	b.WriteString("\n\n")

	b.WriteString(subtitleStyle.Render("Available lists:"))
	b.WriteString("\n")
	for _, list := range m.lists {
		if list == "My-favorites" {
			b.WriteString(fmt.Sprintf("  • %s (protected)\n", list))
		} else {
			b.WriteString(fmt.Sprintf("  • %s\n", list))
		}
	}
	b.WriteString("\n")

	b.WriteString(m.textInput.View())
	b.WriteString("\n\n")

	if m.message != "" {
		b.WriteString(errorStyle.Render(m.message))
		b.WriteString("\n\n")
	}

	help := helpStyle.Render("enter: continue • esc: cancel")
	b.WriteString(help)

	return b.String()
}

// viewConfirmDelete renders the delete confirmation view
func (m ListManagementModel) viewConfirmDelete() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("Confirm Deletion"))
	b.WriteString("\n\n")

	warning := fmt.Sprintf("⚠ Are you sure you want to delete '%s'?", m.selectedList)
	b.WriteString(errorStyle.Render(warning))
	b.WriteString("\n\n")

	b.WriteString("This action cannot be undone.\n\n")

	help := helpStyle.Render("y: yes, delete • n/esc: cancel")
	b.WriteString(help)

	return b.String()
}

// viewEdit renders the edit list view
func (m ListManagementModel) viewEdit() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("Edit List Name"))
	b.WriteString("\n\n")

	b.WriteString(subtitleStyle.Render("Available lists:"))
	b.WriteString("\n")
	for _, list := range m.lists {
		if list == "My-favorites" {
			b.WriteString(fmt.Sprintf("  • %s (protected)\n", list))
		} else {
			b.WriteString(fmt.Sprintf("  • %s\n", list))
		}
	}
	b.WriteString("\n")

	b.WriteString(m.textInput.View())
	b.WriteString("\n\n")

	if m.message != "" {
		b.WriteString(errorStyle.Render(m.message))
		b.WriteString("\n\n")
	}

	help := helpStyle.Render("enter: continue • esc: cancel")
	b.WriteString(help)

	return b.String()
}

// viewEnterNewName renders the new name input view
func (m ListManagementModel) viewEnterNewName() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("Edit List Name"))
	b.WriteString("\n\n")

	b.WriteString(subtitleStyle.Render(fmt.Sprintf("Renaming: %s", m.selectedList)))
	b.WriteString("\n\n")

	b.WriteString(m.textInput.View())
	b.WriteString("\n\n")

	if m.message != "" {
		b.WriteString(errorStyle.Render(m.message))
		b.WriteString("\n\n")
	}

	help := helpStyle.Render("enter: rename • esc: cancel")
	b.WriteString(help)

	return b.String()
}

// viewShowAll renders all lists
func (m ListManagementModel) viewShowAll() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("All Favorite Lists"))
	b.WriteString("\n\n")

	if len(m.lists) == 0 {
		b.WriteString(infoStyle.Render("No lists found"))
		b.WriteString("\n\n")
		b.WriteString("Create your first list using option 1.\n\n")
	} else {
		for i, list := range m.lists {
			if list == "My-favorites" {
				b.WriteString(fmt.Sprintf("%d. %s (Quick Favorites)\n", i+1, list))
			} else {
				b.WriteString(fmt.Sprintf("%d. %s\n", i+1, list))
			}
		}
		b.WriteString("\n")
	}

	help := helpStyle.Render("enter/esc: back")
	b.WriteString(help)

	return b.String()
}

// Messages

type listManagementListsLoadedMsg struct {
	lists []string
}

type listManagementOperationSuccessMsg struct {
	message string
}

type listManagementOperationErrorMsg struct {
	err error
}
