package ui

import (
	"strings"
	"testing"

	"github.com/shinokada/tera/v2/internal/ui/components"
)

func TestListManagementModel_Creation(t *testing.T) {
	model := NewListManagementModel("/tmp/test")

	if model.state != listManagementMenu {
		t.Errorf("Expected initial state listManagementMenu, got %v", model.state)
	}

	if model.favoritePath != "/tmp/test" {
		t.Errorf("Expected favoritePath /tmp/test, got %s", model.favoritePath)
	}

	// Verify the list has 4 menu items
	if len(model.listModel.Items()) != 4 {
		t.Errorf("Expected 4 menu items, got %d", len(model.listModel.Items()))
	}

	// Verify menu items are MenuItem type with correct shortcuts
	expectedItems := []struct {
		title    string
		shortcut string
	}{
		{"Create New List", "1"},
		{"Delete List", "2"},
		{"Edit List Name", "3"},
		{"Show All Lists", "4"},
	}

	for i, item := range model.listModel.Items() {
		menuItem, ok := item.(components.MenuItem)
		if !ok {
			t.Errorf("Item %d is not a MenuItem", i)
			continue
		}
		if menuItem.Title() != expectedItems[i].title {
			t.Errorf("Item %d: expected title %q, got %q", i, expectedItems[i].title, menuItem.Title())
		}
		if menuItem.Shortcut() != expectedItems[i].shortcut {
			t.Errorf("Item %d: expected shortcut %q, got %q", i, expectedItems[i].shortcut, menuItem.Shortcut())
		}
	}
}

func TestListManagementModel_States(t *testing.T) {
	tests := []struct {
		name     string
		state    listManagementState
		expected string
	}{
		{"Menu", listManagementMenu, "List Management"},
		{"Create", listManagementCreate, "Create New List"},
		{"Delete", listManagementDelete, "Delete List"},
		{"SelectListToDelete", listManagementSelectListToDelete, "Select"},
		{"Edit", listManagementEdit, "Edit List Name"},
		{"SelectListToEdit", listManagementSelectListToEdit, "Select"},
		{"ShowAll", listManagementShowAll, "All Favorite Lists"},
		{"ConfirmDelete", listManagementConfirmDelete, "Confirm Deletion"},
		{"EnterNewName", listManagementEnterNewName, "Edit List Name"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := NewListManagementModel("/tmp/test")
			model.state = tt.state
			view := model.View()

			if view == "" {
				t.Error("Expected non-empty view")
			}

			if !strings.Contains(view, tt.expected) {
				t.Errorf("Expected view to contain %q, got:\n%s", tt.expected, view)
			}
		})
	}
}

func TestListManagementModel_MenuFooter(t *testing.T) {
	model := NewListManagementModel("/tmp/test")
	model.state = listManagementMenu
	view := model.View()

	expectedFooter := "↑↓/jk: Navigate • Enter: Select • 1-4: Quick select • Esc: Back • Ctrl+C: Quit"
	if !strings.Contains(view, expectedFooter) {
		t.Errorf("Expected menu footer to contain %q in:\n%s", expectedFooter, view)
	}
}

func TestListManagementModel_CreateFooter(t *testing.T) {
	model := NewListManagementModel("/tmp/test")
	model.state = listManagementCreate
	view := model.View()

	expectedFooter := "Enter: Create • Esc: Back • Ctrl+C: Quit"
	if !strings.Contains(view, expectedFooter) {
		t.Errorf("Expected create footer to contain %q in:\n%s", expectedFooter, view)
	}
}

func TestListManagementModel_DeleteFooter(t *testing.T) {
	model := NewListManagementModel("/tmp/test")
	model.state = listManagementDelete
	view := model.View()

	expectedFooter := "Enter: Continue • Esc: Back • Ctrl+C: Quit"
	if !strings.Contains(view, expectedFooter) {
		t.Errorf("Expected delete footer to contain %q in:\n%s", expectedFooter, view)
	}
}

func TestListManagementModel_EditFooter(t *testing.T) {
	model := NewListManagementModel("/tmp/test")
	model.state = listManagementEdit
	view := model.View()

	expectedFooter := "Enter: Continue • Esc: Back • Ctrl+C: Quit"
	if !strings.Contains(view, expectedFooter) {
		t.Errorf("Expected edit footer to contain %q in:\n%s", expectedFooter, view)
	}
}

func TestListManagementModel_ShowAllFooter(t *testing.T) {
	model := NewListManagementModel("/tmp/test")
	model.state = listManagementShowAll
	view := model.View()

	expectedFooter := "Esc: Back • Ctrl+C: Quit"
	if !strings.Contains(view, expectedFooter) {
		t.Errorf("Expected show all footer to contain %q in:\n%s", expectedFooter, view)
	}
}

func TestListManagementModel_ConfirmDeleteFooter(t *testing.T) {
	model := NewListManagementModel("/tmp/test")
	model.state = listManagementConfirmDelete
	model.selectedList = "TestList"
	view := model.View()

	expectedFooter := "y: Yes, Delete • n/Esc: Cancel"
	if !strings.Contains(view, expectedFooter) {
		t.Errorf("Expected confirm delete footer to contain %q in:\n%s", expectedFooter, view)
	}
}

func TestListManagementModel_RenameFooter(t *testing.T) {
	model := NewListManagementModel("/tmp/test")
	model.state = listManagementEnterNewName
	view := model.View()

	expectedFooter := "Enter: Rename • Esc: Back • Ctrl+C: Quit"
	if !strings.Contains(view, expectedFooter) {
		t.Errorf("Expected rename footer to contain %q in:\n%s", expectedFooter, view)
	}
}

func TestListManagementModel_MenuListDisplay(t *testing.T) {
	model := NewListManagementModel("/tmp/test")
	model.state = listManagementMenu
	view := model.View()

	// Check for menu items in the view
	expectedItems := []string{
		"1. Create New List",
		"2. Delete List",
		"3. Edit List Name",
		"4. Show All Lists",
	}

	for _, item := range expectedItems {
		if !strings.Contains(view, item) {
			t.Errorf("Expected view to contain %q in:\n%s", item, view)
		}
	}
}

func TestListManagementModel_SelectListToEditFooter(t *testing.T) {
	model := NewListManagementModel("/tmp/test")
	model.state = listManagementSelectListToEdit
	view := model.View()

	// Should contain navigate and select instructions
	if !strings.Contains(view, "↑↓/jk: Navigate • Enter: Select") {
		t.Errorf("Expected selectListToEdit footer to contain navigation instructions in:\n%s", view)
	}

	if !strings.Contains(view, "Quick select") {
		t.Errorf("Expected selectListToEdit footer to contain quick select in:\n%s", view)
	}

	if !strings.Contains(view, "Esc: Back • Ctrl+C: Quit") {
		t.Errorf("Expected selectListToEdit footer to contain Esc: Back • Ctrl+C: Quit in:\n%s", view)
	}
}

func TestListManagementModel_SelectListToDeleteFooter(t *testing.T) {
	model := NewListManagementModel("/tmp/test")
	model.state = listManagementSelectListToDelete
	view := model.View()

	// Should contain navigate and select instructions
	if !strings.Contains(view, "↑↓/jk: Navigate • Enter: Select") {
		t.Errorf("Expected selectListToDelete footer to contain navigation instructions in:\n%s", view)
	}

	if !strings.Contains(view, "Quick select") {
		t.Errorf("Expected selectListToDelete footer to contain quick select in:\n%s", view)
	}

	if !strings.Contains(view, "Esc: Back • Ctrl+C: Quit") {
		t.Errorf("Expected selectListToDelete footer to contain Esc: Back • Ctrl+C: Quit in:\n%s", view)
	}
}
