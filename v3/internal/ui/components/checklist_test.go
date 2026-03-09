package components

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// helpers

func makeItems() []ChecklistItem {
	return []ChecklistItem{
		{Key: "favorites", Label: "Favorites (playlists)", Checked: true},
		{Key: "settings", Label: "Settings (config.yaml)", Checked: true},
		{Key: "ratings", Label: "Ratings & votes", Checked: false},
		{Key: "blocklist", Label: "Blocklist", Checked: false},
		{Key: "history", Label: "Search history", Checked: false},
	}
}

func pressKey(m ChecklistModel, key string) (ChecklistModel, tea.Cmd) {
	return m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(key)})
}

func pressSpecial(m ChecklistModel, key tea.KeyType) (ChecklistModel, tea.Cmd) {
	return m.Update(tea.KeyMsg{Type: key})
}

// --- construction ---

func TestNewChecklistModel_InitialState(t *testing.T) {
	m := NewChecklistModel("Test title", makeItems())
	if m.Title != "Test title" {
		t.Errorf("Title = %q, want %q", m.Title, "Test title")
	}
	if m.cursor != 0 {
		t.Errorf("cursor = %d, want 0", m.cursor)
	}
	if len(m.Items) != 5 {
		t.Errorf("len(Items) = %d, want 5", len(m.Items))
	}
}

// --- cursor navigation ---

func TestCursorDown(t *testing.T) {
	m := NewChecklistModel("T", makeItems())
	m, _ = pressKey(m, "j")
	if m.cursor != 1 {
		t.Errorf("cursor = %d, want 1", m.cursor)
	}
}

func TestCursorUp(t *testing.T) {
	m := NewChecklistModel("T", makeItems())
	m.cursor = 2
	m, _ = pressKey(m, "k")
	if m.cursor != 1 {
		t.Errorf("cursor = %d, want 1", m.cursor)
	}
}

func TestCursorDoesNotGoAboveZero(t *testing.T) {
	m := NewChecklistModel("T", makeItems())
	m, _ = pressKey(m, "k")
	if m.cursor != 0 {
		t.Errorf("cursor = %d, want 0", m.cursor)
	}
}

func TestCursorDoesNotGoBelowLast(t *testing.T) {
	m := NewChecklistModel("T", makeItems())
	m.cursor = len(m.Items) - 1
	m, _ = pressKey(m, "j")
	if m.cursor != len(m.Items)-1 {
		t.Errorf("cursor = %d, want %d", m.cursor, len(m.Items)-1)
	}
}

func TestArrowKeys(t *testing.T) {
	m := NewChecklistModel("T", makeItems())
	m, _ = pressSpecial(m, tea.KeyDown)
	if m.cursor != 1 {
		t.Errorf("cursor after ↓ = %d, want 1", m.cursor)
	}
	m, _ = pressSpecial(m, tea.KeyUp)
	if m.cursor != 0 {
		t.Errorf("cursor after ↑ = %d, want 0", m.cursor)
	}
}

// --- toggle ---

func TestSpaceTogglesChecked(t *testing.T) {
	m := NewChecklistModel("T", makeItems())
	// item 0 starts checked; toggle it off
	m, _ = pressKey(m, " ")
	if m.Items[0].Checked {
		t.Error("item 0 should be unchecked after Space")
	}
	// toggle it back on
	m, _ = pressKey(m, " ")
	if !m.Items[0].Checked {
		t.Error("item 0 should be checked again after second Space")
	}
}

func TestSpaceTogglesUnchecked(t *testing.T) {
	m := NewChecklistModel("T", makeItems())
	m.cursor = 2 // ratings starts unchecked
	m, _ = pressKey(m, " ")
	if !m.Items[2].Checked {
		t.Error("item 2 should be checked after Space")
	}
}

func TestToggleAll_SelectsAllWhenSomeUnchecked(t *testing.T) {
	m := NewChecklistModel("T", makeItems())
	// items 0,1 checked; 2,3,4 unchecked → 'a' should check all
	m, _ = pressKey(m, "a")
	for i, item := range m.Items {
		if !item.Checked {
			t.Errorf("item %d (%s) should be checked after 'a'", i, item.Key)
		}
	}
}

func TestToggleAll_DeselectsAllWhenAllChecked(t *testing.T) {
	items := makeItems()
	for i := range items {
		items[i].Checked = true
	}
	m := NewChecklistModel("T", items)
	m, _ = pressKey(m, "a")
	for i, item := range m.Items {
		if item.Checked {
			t.Errorf("item %d (%s) should be unchecked after 'a' when all checked", i, item.Key)
		}
	}
}

// --- confirm / cancel ---

func TestEnterEmitsConfirmedMsg(t *testing.T) {
	m := NewChecklistModel("T", makeItems())
	_, cmd := pressSpecial(m, tea.KeyEnter)
	if cmd == nil {
		t.Fatal("expected a Cmd after Enter, got nil")
	}
	msg := cmd()
	confirmed, ok := msg.(ChecklistConfirmedMsg)
	if !ok {
		t.Fatalf("expected ChecklistConfirmedMsg, got %T", msg)
	}
	if len(confirmed.Items) != len(m.Items) {
		t.Errorf("confirmed items len = %d, want %d", len(confirmed.Items), len(m.Items))
	}
}

func TestEscEmitsCancelledMsg(t *testing.T) {
	m := NewChecklistModel("T", makeItems())
	_, cmd := pressSpecial(m, tea.KeyEsc)
	if cmd == nil {
		t.Fatal("expected a Cmd after Esc, got nil")
	}
	msg := cmd()
	if _, ok := msg.(ChecklistCancelledMsg); !ok {
		t.Fatalf("expected ChecklistCancelledMsg, got %T", msg)
	}
}

func TestQKeyEmitsCancelledMsg(t *testing.T) {
	m := NewChecklistModel("T", makeItems())
	_, cmd := pressKey(m, "q")
	if cmd == nil {
		t.Fatal("expected a Cmd after q, got nil")
	}
	if _, ok := cmd().(ChecklistCancelledMsg); !ok {
		t.Fatalf("expected ChecklistCancelledMsg")
	}
}

// --- helpers ---

func TestCheckedKeys(t *testing.T) {
	m := NewChecklistModel("T", makeItems())
	// items 0,1 checked by default
	keys := m.CheckedKeys()
	if len(keys) != 2 {
		t.Errorf("CheckedKeys len = %d, want 2", len(keys))
	}
	if keys[0] != "favorites" || keys[1] != "settings" {
		t.Errorf("CheckedKeys = %v, want [favorites settings]", keys)
	}
}

func TestAnyChecked_True(t *testing.T) {
	m := NewChecklistModel("T", makeItems())
	if !m.AnyChecked() {
		t.Error("AnyChecked should be true")
	}
}

func TestAnyChecked_False(t *testing.T) {
	items := makeItems()
	for i := range items {
		items[i].Checked = false
	}
	m := NewChecklistModel("T", items)
	if m.AnyChecked() {
		t.Error("AnyChecked should be false when nothing is checked")
	}
}

func TestSetItems_ResetsCursor(t *testing.T) {
	m := NewChecklistModel("T", makeItems())
	m.cursor = 3
	m.SetItems(makeItems()[:2])
	if m.cursor != 0 {
		t.Errorf("cursor after SetItems = %d, want 0", m.cursor)
	}
	if len(m.Items) != 2 {
		t.Errorf("len(Items) after SetItems = %d, want 2", len(m.Items))
	}
}

func TestView_NonEmpty(t *testing.T) {
	m := NewChecklistModel("Pick categories", makeItems())
	v := m.View()
	if v == "" {
		t.Error("View should return non-empty string")
	}
	if !strings.Contains(v, "Pick categories") {
		t.Error("View should contain the title")
	}
	if !strings.Contains(v, "Favorites") {
		t.Error("View should contain item label")
	}
	if !strings.Contains(v, "[x]") {
		t.Error("View should show [x] for checked items")
	}
	if !strings.Contains(v, "[ ]") {
		t.Error("View should show [ ] for unchecked items")
	}
}

func TestInit_ReturnsNil(t *testing.T) {
	m := NewChecklistModel("T", makeItems())
	if m.Init() != nil {
		t.Error("Init should return nil")
	}
}
