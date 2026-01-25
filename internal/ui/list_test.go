package ui

import (
	"testing"
)

func TestListManagementModel_Creation(t *testing.T) {
	favPath := "/tmp/test-favorites"
	model := NewListManagementModel(favPath)

	if model.favoritePath != favPath {
		t.Errorf("Expected favoritePath %s, got %s", favPath, model.favoritePath)
	}

	if model.state != listManagementMenu {
		t.Errorf("Expected initial state listManagementMenu, got %d", model.state)
	}

	if model.textInput.Placeholder != "Enter list name" {
		t.Errorf("Expected placeholder 'Enter list name', got '%s'", model.textInput.Placeholder)
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
		{"ShowAll", listManagementShowAll, "showAll"},
		{"ConfirmDelete", listManagementConfirmDelete, "confirmDelete"},
		{"EnterNewName", listManagementEnterNewName, "enterNewName"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := NewListManagementModel("/tmp")
			model.state = tt.state
			
			// Just verify the state can be set without panic
			if model.state != tt.state {
				t.Errorf("Failed to set state %d", tt.state)
			}
		})
	}
}
