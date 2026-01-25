package ui

import (
	"testing"
)

func TestListManagementModel_Creation(t *testing.T) {
	model := NewListManagementModel("/tmp/test")

	if model.state != listManagementMenu {
		t.Errorf("Expected initial state listManagementMenu, got %v", model.state)
	}

	if model.favoritePath != "/tmp/test" {
		t.Errorf("Expected favoritePath /tmp/test, got %s", model.favoritePath)
	}
}

func TestListManagementModel_States(t *testing.T) {
	tests := []struct {
		name     string
		state    listManagementState
		expected string
	}{
		{"Menu", listManagementMenu, "menu"},
		{"Create", listManagementCreate, "create"},
		{"Delete", listManagementDelete, "delete"},
		{"Edit", listManagementEdit, "edit"},
		{"ShowAll", listManagementShowAll, "show"},
		{"ConfirmDelete", listManagementConfirmDelete, "confirm"},
		{"EnterNewName", listManagementEnterNewName, "newname"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := NewListManagementModel("/tmp/test")
			model.state = tt.state
			view := model.View()

			if view == "" {
				t.Error("Expected non-empty view")
			}
		})
	}
}
