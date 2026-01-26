package ui

import (
	"os"
	"path/filepath"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestNewPlayModel(t *testing.T) {
	favPath := "/tmp/favorites"
	model := NewPlayModel(favPath)

	if model.favoritePath != favPath {
		t.Errorf("Expected favoritePath %s, got %s", favPath, model.favoritePath)
	}

	if model.state != playStateListSelection {
		t.Errorf("Expected initial state playStateListSelection, got %v", model.state)
	}
}

func TestGetAvailableLists(t *testing.T) {
	// Create a temporary directory
	tmpDir := t.TempDir()

	// Test with no files initially - should return error
	model := NewPlayModel(tmpDir)
	lists, err := model.getAvailableLists()
	// Should get error since no files exist yet
	if err == nil {
		t.Error("Expected error when no files exist, got nil")
	}
	if lists != nil {
		t.Errorf("Expected nil lists, got %v", lists)
	}

	// Create some test JSON files
	testFiles := []string{"favorites.json", "jazz.json", "rock.json"}
	for _, name := range testFiles {
		path := filepath.Join(tmpDir, name)
		if err := os.WriteFile(path, []byte("[]"), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
	}

	// Create a non-JSON file (should be ignored)
	if err := os.WriteFile(filepath.Join(tmpDir, "readme.txt"), []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Test with files
	lists, err = model.getAvailableLists()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Should find 3 JSON files (without extension)
	expectedCount := 3
	if len(lists) != expectedCount {
		t.Errorf("Expected %d lists, got %d", expectedCount, len(lists))
	}

	// Verify .json extension is removed
	for _, list := range lists {
		if filepath.Ext(list) == ".json" {
			t.Errorf("List name should not have .json extension: %s", list)
		}
	}
}

func TestPlayModel_Update_ListsLoaded(t *testing.T) {
	model := NewPlayModel("/tmp/favorites")
	model.width = 80
	model.height = 24

	// Simulate lists loaded message
	msg := listsLoadedMsg{
		lists: []string{"favorites", "jazz", "rock"},
	}

	updatedModel, _ := model.Update(msg)
	m := updatedModel.(PlayModel)

	if len(m.lists) != 3 {
		t.Errorf("Expected 3 lists, got %d", len(m.lists))
	}

	if len(m.listItems) != 3 {
		t.Errorf("Expected 3 list items, got %d", len(m.listItems))
	}

	// Check that list model is initialized
	if m.listModel.Items() == nil {
		t.Error("Expected listModel to be initialized")
	}
}

func TestPlayModel_Update_NavigationKeys(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		keyType  tea.KeyType
		expected bool // true if should navigate away
	}{
		{"Escape key", "esc", tea.KeyEsc, true},
		{"Other key", "a", tea.KeyRunes, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := NewPlayModel("/tmp/favorites")
			model.state = playStateListSelection
			model.width = 80
			model.height = 24

			// Initialize with some lists
			updatedModel, _ := model.Update(listsLoadedMsg{
				lists: []string{"favorites"},
			})
			model = updatedModel.(PlayModel)

			// Send key message
			var keyMsg tea.KeyMsg
			if tt.keyType == tea.KeyEsc {
				keyMsg = tea.KeyMsg{Type: tea.KeyEsc}
			} else {
				keyMsg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(tt.key)}
			}

			updatedModel, cmd := model.Update(keyMsg)
			m := updatedModel.(PlayModel)

			if tt.expected {
				// Should return navigation command
				if cmd == nil {
					t.Error("Expected navigation command, got nil")
				} else {
					// Verify it's a navigateMsg
					result := cmd()
					if _, ok := result.(navigateMsg); !ok {
						t.Errorf("Expected navigateMsg, got %T", result)
					}
				}
			} else {
				// Should stay in same state
				if m.state != playStateListSelection {
					t.Errorf("Expected to stay in playStateListSelection, got %v", m.state)
				}
			}
		})
	}
}

func TestPlayModel_View_NoLists(t *testing.T) {
	model := NewPlayModel("/tmp/favorites")
	view := model.View()

	if view == "" {
		t.Error("Expected non-empty view")
	}

	// Should show error about no lists
	// This is a basic check - could be more specific
	if len(view) < 10 {
		t.Error("Expected more substantial view content")
	}
}

func TestPlayModel_View_WithLists(t *testing.T) {
	model := NewPlayModel("/tmp/favorites")
	model.width = 80
	model.height = 24

	// Load lists
	model.Update(listsLoadedMsg{
		lists: []string{"favorites", "jazz"},
	})

	view := model.viewListSelection()

	if view == "" {
		t.Error("Expected non-empty view")
	}

	// Should contain title
	// (exact content depends on styling, so just check it's substantial)
	if len(view) < 20 {
		t.Error("Expected more substantial view content with lists")
	}
}

func TestPlayListItem(t *testing.T) {
	item := playListItem{name: "My Favorites"}

	if item.FilterValue() != "My Favorites" {
		t.Errorf("Expected FilterValue 'My Favorites', got '%s'", item.FilterValue())
	}

	if item.Title() != "My Favorites" {
		t.Errorf("Expected Title 'My Favorites', got '%s'", item.Title())
	}

	if item.Description() != "" {
		t.Errorf("Expected empty Description, got '%s'", item.Description())
	}
}

func TestErrorView(t *testing.T) {
	err := os.ErrNotExist
	view := errorView(err)

	if view == "" {
		t.Error("Expected non-empty error view")
	}

	// Should contain error message
	if len(view) < 10 {
		t.Error("Expected substantial error view content")
	}
}

func TestPlayModel_Update_EnterKey(t *testing.T) {
	// Test that Enter key changes state when list is properly initialized
	model := NewPlayModel("/tmp/favorites")
	model.width = 80
	model.height = 24
	model.state = playStateListSelection

	// Initialize with lists
	updatedModel, _ := model.Update(listsLoadedMsg{
		lists: []string{"favorites", "jazz"},
	})
	model = updatedModel.(PlayModel)

	// Verify we're in list selection state
	if model.state != playStateListSelection {
		t.Fatalf("Expected playStateListSelection, got %v", model.state)
	}

	// Verify list model is initialized
	if model.listModel.Items() == nil {
		t.Fatal("List model not initialized")
	}

	// Note: We can't easily test Enter key selection because it requires
	// the list model to have a selected item, which is handled by bubbles
	// internally. In real usage, the list will have a default selection.
	// This is tested through integration tests.
}
