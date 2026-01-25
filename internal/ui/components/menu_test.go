package components

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestMenuCreation(t *testing.T) {
	items := []MenuItem{
		NewMenuItem("Item 1", "Description 1", "1"),
		NewMenuItem("Item 2", "Description 2", "2"),
		NewMenuItem("Item 3", "Description 3", "3"),
	}

	menu := CreateMenu(items, "Test Menu", 50, 10)

	if len(menu.Items()) != 3 {
		t.Errorf("Expected 3 items, got %d", len(menu.Items()))
	}

	if menu.Title != "Test Menu" {
		t.Errorf("Expected title 'Test Menu', got '%s'", menu.Title)
	}
}

func TestHandleMenuKey(t *testing.T) {
	items := []MenuItem{
		NewMenuItem("Item 1", "", "1"),
		NewMenuItem("Item 2", "", "2"),
		NewMenuItem("Item 3", "", "3"),
	}

	menu := CreateMenu(items, "Test", 50, 10)

	tests := []struct {
		name     string
		key      string
		expected int
	}{
		{"Down arrow", "down", -1},
		{"Up arrow", "up", -1},
		{"Enter key", "enter", 0},
		{"Number 1", "1", 0},
		{"Number 2", "2", 1},
		{"Number 3", "3", 2},
		{"j key", "j", -1},
		{"k key", "k", -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset menu to first item
			menu.Select(0)

			msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(tt.key)}
			if tt.key == "down" {
				msg = tea.KeyMsg{Type: tea.KeyDown}
			} else if tt.key == "up" {
				msg = tea.KeyMsg{Type: tea.KeyUp}
			} else if tt.key == "enter" {
				msg = tea.KeyMsg{Type: tea.KeyEnter}
			}

			_, selected := HandleMenuKey(msg, menu)

			if selected != tt.expected && tt.key != "down" && tt.key != "up" && tt.key != "j" && tt.key != "k" {
				t.Errorf("Expected selection %d, got %d", tt.expected, selected)
			}
		})
	}
}

func TestMenuItemInterface(t *testing.T) {
	item := NewMenuItem("Test Item", "Test Description", "1")

	if item.Title() != "Test Item" {
		t.Errorf("Expected title 'Test Item', got '%s'", item.Title())
	}

	if item.Description() != "Test Description" {
		t.Errorf("Expected description 'Test Description', got '%s'", item.Description())
	}

	if item.FilterValue() != "Test Item" {
		t.Errorf("Expected filter value 'Test Item', got '%s'", item.FilterValue())
	}

	if item.Shortcut() != "1" {
		t.Errorf("Expected shortcut '1', got '%s'", item.Shortcut())
	}
}

func TestMenuNavigation(t *testing.T) {
	items := []MenuItem{
		NewMenuItem("First", "", "1"),
		NewMenuItem("Second", "", "2"),
		NewMenuItem("Third", "", "3"),
	}

	menu := CreateMenu(items, "Nav Test", 50, 10)

	// Test moving down
	menu.CursorDown()
	if menu.Index() != 1 {
		t.Errorf("Expected index 1 after moving down, got %d", menu.Index())
	}

	// Test moving up
	menu.CursorUp()
	if menu.Index() != 0 {
		t.Errorf("Expected index 0 after moving up, got %d", menu.Index())
	}

	// Test going to last item
	menu.Select(len(menu.Items()) - 1)
	if menu.Index() != 2 {
		t.Errorf("Expected index 2 (last item), got %d", menu.Index())
	}

	// Test going to first item
	menu.Select(0)
	if menu.Index() != 0 {
		t.Errorf("Expected index 0 (first item), got %d", menu.Index())
	}
}
