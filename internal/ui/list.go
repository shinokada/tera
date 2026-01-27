package ui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/shinokada/tera/internal/ui/components"
)

// listManagementState represents the current state in the list management menu
type listManagementState int

const (
	listManagementMenu listManagementState = iota
	listManagementCreate
	listManagementDelete
	listManagementSelectListToDelete
	listManagementEdit
	listManagementSelectListToEdit
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

// NewListManagementModel creates a new list management model
func NewListManagementModel(favoritePath string) ListManagementModel {
	ti := textinput.New()
	ti.Placeholder = "Enter list name"
	ti.CharLimit = 50

	items := []list.Item{
		components.NewMenuItem("Create New List", "Create a new favorites list", "1"),
		components.NewMenuItem("Delete List", "Delete an existing list", "2"),
		components.NewMenuItem("Edit List Name", "Rename an existing list", "3"),
		components.NewMenuItem("Show All Lists", "Display all favorite lists", "4"),
	}

	delegate := components.NewMenuDelegate()
	l := list.New(items, delegate, 80, 10)
	l.Title = "📋 List Management"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.SetShowHelp(false)
	l.SetShowPagination(false)
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
		// Ensure enough height for menu items (4 items + title + help)
		h := msg.Height - 4
		if h < 8 {
			h = 8
		}
		m.listModel.SetSize(msg.Width-4, h)
		return m, nil

	case listManagementListsLoadedMsg:
		m.lists = msg.lists
		// If we're in the select to delete state, populate the list model with actual lists
		if m.state == listManagementSelectListToDelete {
			items := make([]list.Item, len(m.lists))
			for i, listName := range m.lists {
				items[i] = components.NewMenuItem(listName, "", fmt.Sprintf("%d", i+1))
			}
			m.listModel.SetItems(items)
			m.listModel.Select(0)
		} else if m.state == listManagementSelectListToEdit {
			items := make([]list.Item, len(m.lists))
			for i, listName := range m.lists {
				items[i] = components.NewMenuItem(listName, "", fmt.Sprintf("%d", i+1))
			}
			m.listModel.SetItems(items)
			m.listModel.Select(0)
		} else if m.state == listManagementMenu {
			// Reset to main menu items
			items := []list.Item{
				components.NewMenuItem("Create New List", "Create a new favorites list", "1"),
				components.NewMenuItem("Delete List", "Delete an existing list", "2"),
				components.NewMenuItem("Edit List Name", "Rename an existing list", "3"),
				components.NewMenuItem("Show All Lists", "Display all favorite lists", "4"),
			}
			m.listModel.SetItems(items)
			m.listModel.Select(0)
		}
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
	case listManagementSelectListToDelete:
		m.listModel, cmd = m.listModel.Update(msg)
	case listManagementSelectListToEdit:
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
	case listManagementSelectListToDelete:
		return m.handleSelectListToDeleteInput(msg)
	case listManagementEdit:
		return m.handleEditInput(msg)
	case listManagementSelectListToEdit:
		return m.handleSelectListToEditInput(msg)
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
	case "esc", "m":
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
		m.state = listManagementSelectListToDelete
		return m, m.loadLists()
	case 2: // Edit list name
		if len(m.lists) == 0 {
			m.message = "No lists available to edit"
			m.messageTime = 150
			return m, nil
		}
		m.state = listManagementSelectListToEdit
		m.textInput.Reset()
		return m, m.loadLists()
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
	case "m":
		// Return to main menu
		m.textInput.Blur()
		return m, func() tea.Msg {
			return navigateMsg{screen: screenMainMenu}
		}
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

// handleSelectListToDeleteInput handles selection of list to delete
func (m ListManagementModel) handleSelectListToDeleteInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "m":
		m.state = listManagementMenu
		return m, nil
	case "q":
		return m, tea.Quit
	case "enter":
		// Get selected item
		item := m.listModel.SelectedItem()
		if item == nil {
			return m, nil
		}
		menuItem, ok := item.(components.MenuItem)
		if !ok {
			return m, nil
		}
		selectedList := menuItem.Title()

		// Check if it's My-favorites (protected)
		if selectedList == "My-favorites" {
			m.message = "Cannot delete My-favorites (protected list)"
			m.messageTime = 150
			return m, nil
		}

		// Move to confirmation
		m.selectedList = selectedList
		m.state = listManagementConfirmDelete
		return m, nil
	case "1", "2", "3", "4", "5", "6", "7", "8", "9":
		// Quick select by number
		idx := -1
		if num := msg.String(); len(num) == 1 && num[0] >= '1' && num[0] <= '9' {
			idx = int(num[0] - '1')
		}
		if idx >= 0 && idx < len(m.listModel.Items()) {
			m.listModel.Select(idx)
			// Simulate enter press
			item := m.listModel.SelectedItem()
			if item != nil {
				if menuItem, ok := item.(components.MenuItem); ok {
					selectedList := menuItem.Title()
					if selectedList != "My-favorites" {
						m.selectedList = selectedList
						m.state = listManagementConfirmDelete
						return m, nil
					} else {
						m.message = "Cannot delete My-favorites (protected list)"
						m.messageTime = 150
					}
				}
			}
		}
	}

	var cmd tea.Cmd
	m.listModel, cmd = m.listModel.Update(msg)
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

// handleSelectListToEditInput handles selection of list to edit
func (m ListManagementModel) handleSelectListToEditInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "m":
		m.state = listManagementMenu
		return m, nil
	case "q":
		return m, tea.Quit
	case "enter":
		// Get selected item
		item := m.listModel.SelectedItem()
		if item == nil {
			return m, nil
		}
		menuItem, ok := item.(components.MenuItem)
		if !ok {
			return m, nil
		}
		selectedList := menuItem.Title()

		// Check if it's My-favorites (protected)
		if selectedList == "My-favorites" {
			m.message = "Cannot rename My-favorites (protected list)"
			m.messageTime = 150
			return m, nil
		}

		// Move to new name input
		m.selectedList = selectedList
		m.state = listManagementEnterNewName
		m.textInput.Reset()
		m.textInput.Placeholder = "Enter new name"
		m.textInput.Focus()
		return m, textinput.Blink
	case "1", "2", "3", "4", "5", "6", "7", "8", "9":
		// Quick select by number
		idx := -1
		if num := msg.String(); len(num) == 1 && num[0] >= '1' && num[0] <= '9' {
			idx = int(num[0] - '1')
		}
		if idx >= 0 && idx < len(m.listModel.Items()) {
			m.listModel.Select(idx)
			// Simulate enter press
			item := m.listModel.SelectedItem()
			if item != nil {
				if menuItem, ok := item.(components.MenuItem); ok {
					selectedList := menuItem.Title()
					if selectedList != "My-favorites" {
						m.selectedList = selectedList
						m.state = listManagementEnterNewName
						m.textInput.Reset()
						m.textInput.Placeholder = "Enter new name"
						m.textInput.Focus()
						return m, textinput.Blink
					} else {
						m.message = "Cannot rename My-favorites (protected list)"
						m.messageTime = 150
					}
				}
			}
		}
	}

	var cmd tea.Cmd
	m.listModel, cmd = m.listModel.Update(msg)
	return m, cmd
}

// handleEditInput handles input during list editing (deprecated - kept for compatibility)
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
	case listManagementSelectListToDelete:
		return m.viewSelectListToDelete()
	case listManagementEdit:
		return m.viewEdit()
	case listManagementSelectListToEdit:
		return m.viewSelectListToEdit()
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
// viewMenu renders the list management menu
func (m ListManagementModel) viewMenu() string {
	var content strings.Builder

	if m.message != "" {
		style := successStyle
		if strings.Contains(m.message, "✗") || m.err != nil {
			style = errorStyle
		}
		content.WriteString(style.Render(m.message))
		content.WriteString("\n\n")
	}

	content.WriteString(m.listModel.View())

	return RenderPage(PageLayout{
		Content: content.String(),
		Help:    "↑↓/jk: Navigate • Enter: Select • 1-4: Quick select • Esc: Back • Ctrl+C: Quit",
	})
}

// viewCreate renders the create list view
func (m ListManagementModel) viewCreate() string {
	var content strings.Builder

	if len(m.lists) > 0 {
		content.WriteString(subtitleStyle.Render("Current lists:"))
		content.WriteString("\n")
		for _, list := range m.lists {
			content.WriteString(fmt.Sprintf("  • %s\n", list))
		}
		content.WriteString("\n")
	}

	content.WriteString(m.textInput.View())

	if m.message != "" {
		content.WriteString("\n\n")
		content.WriteString(errorStyle.Render(m.message))
	}

	return RenderPage(PageLayout{
		Title:   "Create New List",
		Content: content.String(),
		Help:    "Enter: Create • Esc: Back • Ctrl+C: Quit",
	})
}

// viewDelete renders the delete list view
func (m ListManagementModel) viewDelete() string {
	var content strings.Builder

	content.WriteString(subtitleStyle.Render("Available lists:"))
	content.WriteString("\n")
	for _, list := range m.lists {
		if list == "My-favorites" {
			content.WriteString(fmt.Sprintf("  • %s (protected)\n", list))
		} else {
			content.WriteString(fmt.Sprintf("  • %s\n", list))
		}
	}
	content.WriteString("\n")

	content.WriteString(m.textInput.View())

	if m.message != "" {
		content.WriteString("\n\n")
		content.WriteString(errorStyle.Render(m.message))
	}

	return RenderPage(PageLayout{
		Title:   "Delete List",
		Content: content.String(),
		Help:    "Enter: Continue • Esc: Back • Ctrl+C: Quit",
	})
}

// viewSelectListToDelete renders the list selection view for deletion
func (m ListManagementModel) viewSelectListToDelete() string {
	var content strings.Builder

	if m.message != "" {
		style := successStyle
		if strings.Contains(m.message, "✗") || m.message == "Cannot delete My-favorites (protected list)" {
			style = errorStyle
		}
		content.WriteString(style.Render(m.message))
		content.WriteString("\n\n")
	}

	content.WriteString(m.listModel.View())

	numLists := len(m.lists)
	maxNum := numLists
	if maxNum > 9 {
		maxNum = 9
	}

	return RenderPage(PageLayout{
		Content: content.String(),
		Help:    fmt.Sprintf("↑↓/jk: Navigate • Enter: Select • 1-%d: Quick select • Esc: Back • Ctrl+C: Quit", maxNum),
	})
}

// viewConfirmDelete renders the delete confirmation view
func (m ListManagementModel) viewConfirmDelete() string {
	var content strings.Builder

	warning := fmt.Sprintf("⚠ Are you sure you want to delete '%s'?", m.selectedList)
	content.WriteString(errorStyle.Render(warning))
	content.WriteString("\n\n")

	content.WriteString("This action cannot be undone.")

	return RenderPage(PageLayout{
		Title:   "Confirm Deletion",
		Content: content.String(),
		Help:    "y: Yes, Delete • n/Esc: Cancel",
	})
}

// viewEdit renders the edit list view
func (m ListManagementModel) viewEdit() string {
	var content strings.Builder

	content.WriteString(subtitleStyle.Render("Available lists:"))
	content.WriteString("\n")
	for _, list := range m.lists {
		if list == "My-favorites" {
			content.WriteString(fmt.Sprintf("  • %s (protected)\n", list))
		} else {
			content.WriteString(fmt.Sprintf("  • %s\n", list))
		}
	}
	content.WriteString("\n")

	content.WriteString(m.textInput.View())

	if m.message != "" {
		content.WriteString("\n\n")
		content.WriteString(errorStyle.Render(m.message))
	}

	return RenderPage(PageLayout{
		Title:   "Edit List Name",
		Content: content.String(),
		Help:    "Enter: Continue • Esc: Back • Ctrl+C: Quit",
	})
}

// viewSelectListToEdit renders the list selection view for editing
func (m ListManagementModel) viewSelectListToEdit() string {
	var content strings.Builder

	if m.message != "" {
		style := successStyle
		if strings.Contains(m.message, "✗") || m.message == "Cannot rename My-favorites (protected list)" {
			style = errorStyle
		}
		content.WriteString(style.Render(m.message))
		content.WriteString("\n\n")
	}

	content.WriteString(m.listModel.View())

	numLists := len(m.lists)
	maxNum := numLists
	if maxNum > 9 {
		maxNum = 9
	}

	return RenderPage(PageLayout{
		Content: content.String(),
		Help:    fmt.Sprintf("↑↓/jk: Navigate • Enter: Select • 1-%d: Quick select • Esc: Back • Ctrl+C: Quit", maxNum),
	})
}

// viewEnterNewName renders the new name input view
func (m ListManagementModel) viewEnterNewName() string {
	var content strings.Builder

	content.WriteString(m.textInput.View())

	if m.message != "" {
		content.WriteString("\n\n")
		content.WriteString(errorStyle.Render(m.message))
	}

	return RenderPage(PageLayout{
		Title:    "Edit List Name",
		Subtitle: fmt.Sprintf("Renaming: %s", m.selectedList),
		Content:  content.String(),
		Help:     "Enter: Rename • Esc: Back • Ctrl+C: Quit",
	})
}

// viewShowAll renders all lists
func (m ListManagementModel) viewShowAll() string {
	var content strings.Builder

	if len(m.lists) == 0 {
		content.WriteString(infoStyle.Render("No lists found"))
		content.WriteString("\n\n")
		content.WriteString("Create your first list using option 1.")
	} else {
		for i, list := range m.lists {
			if list == "My-favorites" {
				content.WriteString(fmt.Sprintf("%d. %s (Quick Favorites)\n", i+1, list))
			} else {
				content.WriteString(fmt.Sprintf("%d. %s\n", i+1, list))
			}
		}
	}

	return RenderPage(PageLayout{
		Title:   "All Favorite Lists",
		Content: content.String(),
		Help:    "Esc: Back • Ctrl+C: Quit",
	})
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
