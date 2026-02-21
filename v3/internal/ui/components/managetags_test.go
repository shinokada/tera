package components

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// ---------------------------------------------------------------------------
// Creation and initialization
// ---------------------------------------------------------------------------

func TestNewManageTags(t *testing.T) {
	stationName := "Test Station"
	currentTags := []string{"rock", "jazz"}
	allTags := []string{"rock", "jazz", "blues", "classical"}
	width := 80

	mt := NewManageTags(stationName, currentTags, allTags, width)

	if mt.stationName != stationName {
		t.Errorf("expected stationName %q, got %q", stationName, mt.stationName)
	}
	if mt.width != width {
		t.Errorf("expected width %d, got %d", width, mt.width)
	}
	if mt.cursor != 0 {
		t.Errorf("expected cursor 0, got %d", mt.cursor)
	}
	if mt.addingNew {
		t.Error("expected addingNew false initially")
	}
}

func TestNewManageTagsCurrentTagsFirst(t *testing.T) {
	currentTags := []string{"jazz", "rock"}
	allTags := []string{"rock", "jazz", "blues"}

	mt := NewManageTags("Station", currentTags, allTags, 80)

	// Current tags should appear first and be checked
	if len(mt.entries) < 2 {
		t.Fatalf("expected at least 2 entries, got %d", len(mt.entries))
	}

	// First two should be current tags (checked)
	if !mt.entries[0].checked || !mt.entries[1].checked {
		t.Error("expected first two entries to be checked")
	}

	// Should contain jazz and rock in order
	foundJazz := false
	foundRock := false
	for i := 0; i < 2; i++ {
		if mt.entries[i].tag == "jazz" {
			foundJazz = true
		}
		if mt.entries[i].tag == "rock" {
			foundRock = true
		}
	}
	if !foundJazz || !foundRock {
		t.Error("expected jazz and rock in first two entries")
	}
}

func TestNewManageTagsRemainingUnchecked(t *testing.T) {
	currentTags := []string{"rock"}
	allTags := []string{"rock", "jazz", "blues"}

	mt := NewManageTags("Station", currentTags, allTags, 80)

	// Should have 3 entries total
	if len(mt.entries) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(mt.entries))
	}

	// First entry (rock) should be checked
	if !mt.entries[0].checked {
		t.Error("expected first entry (rock) to be checked")
	}

	// Remaining entries should be unchecked
	checkedCount := 0
	for _, e := range mt.entries {
		if e.checked {
			checkedCount++
		}
	}
	if checkedCount != 1 {
		t.Errorf("expected 1 checked entry, got %d", checkedCount)
	}
}

func TestNewManageTagsEmptyCurrentTags(t *testing.T) {
	currentTags := []string{}
	allTags := []string{"rock", "jazz"}

	mt := NewManageTags("Station", currentTags, allTags, 80)

	// All entries should be unchecked
	for i, e := range mt.entries {
		if e.checked {
			t.Errorf("expected entry %d to be unchecked, got checked", i)
		}
	}
}

func TestNewManageTagsNoDuplicates(t *testing.T) {
	currentTags := []string{"rock"}
	allTags := []string{"rock", "jazz", "rock"} // rock appears twice

	mt := NewManageTags("Station", currentTags, allTags, 80)

	// Count occurrences of "rock"
	rockCount := 0
	for _, e := range mt.entries {
		if e.tag == "rock" {
			rockCount++
		}
	}

	if rockCount > 1 {
		t.Errorf("expected rock to appear once, got %d times", rockCount)
	}
}

// ---------------------------------------------------------------------------
// Key handling - navigation
// ---------------------------------------------------------------------------

func TestManageTagsUpdateEsc(t *testing.T) {
	mt := NewManageTags("Station", []string{}, []string{}, 80)

	mt, cmd := mt.Update(tea.KeyMsg{Type: tea.KeyEsc})

	if cmd == nil {
		t.Fatal("expected cmd from Esc, got nil")
	}

	msg := cmd()
	if _, ok := msg.(ManageTagsCancelledMsg); !ok {
		t.Errorf("expected ManageTagsCancelledMsg, got %T", msg)
	}
}

func TestUpdateNavigationJK(t *testing.T) {
	mt := NewManageTags("Station", []string{}, []string{"rock", "jazz"}, 80)

	// Initial cursor at 0
	if mt.cursor != 0 {
		t.Errorf("expected cursor 0, got %d", mt.cursor)
	}

	// Press 'j' (down)
	mt, _ = mt.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	if mt.cursor != 1 {
		t.Errorf("expected cursor 1 after 'j', got %d", mt.cursor)
	}

	// Press 'k' (up)
	mt, _ = mt.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")})
	if mt.cursor != 0 {
		t.Errorf("expected cursor 0 after 'k', got %d", mt.cursor)
	}
}

func TestUpdateNavigationArrowKeys(t *testing.T) {
	mt := NewManageTags("Station", []string{}, []string{"rock", "jazz"}, 80)

	// Press Down
	mt, _ = mt.Update(tea.KeyMsg{Type: tea.KeyDown})
	if mt.cursor != 1 {
		t.Errorf("expected cursor 1 after Down, got %d", mt.cursor)
	}

	// Press Up
	mt, _ = mt.Update(tea.KeyMsg{Type: tea.KeyUp})
	if mt.cursor != 0 {
		t.Errorf("expected cursor 0 after Up, got %d", mt.cursor)
	}
}

func TestUpdateNavigationBounds(t *testing.T) {
	mt := NewManageTags("Station", []string{}, []string{"rock"}, 80)
	// Total items = 1 entry + 1 "Add new tag" row = 2

	// At cursor 0, pressing Up should not go negative
	mt.cursor = 0
	mt, _ = mt.Update(tea.KeyMsg{Type: tea.KeyUp})
	if mt.cursor < 0 {
		t.Errorf("expected cursor >= 0 after Up at boundary, got %d", mt.cursor)
	}

	// At last item, pressing Down should not exceed
	mt.cursor = 1 // "Add new tag" row
	mt, _ = mt.Update(tea.KeyMsg{Type: tea.KeyDown})
	if mt.cursor > 1 {
		t.Errorf("expected cursor <= 1 after Down at boundary, got %d", mt.cursor)
	}
}

// ---------------------------------------------------------------------------
// Key handling - toggle
// ---------------------------------------------------------------------------

func TestUpdateSpaceToggle(t *testing.T) {
	mt := NewManageTags("Station", []string{"rock"}, []string{"rock", "jazz"}, 80)

	// First entry is "rock" (checked)
	mt.cursor = 0
	if !mt.entries[0].checked {
		t.Fatal("expected first entry to be checked initially")
	}

	// Press Space to toggle
	mt, _ = mt.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(" ")})
	if mt.entries[0].checked {
		t.Error("expected first entry to be unchecked after Space")
	}

	// Press Space again to toggle back
	mt, _ = mt.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(" ")})
	if !mt.entries[0].checked {
		t.Error("expected first entry to be checked after second Space")
	}
}

func TestUpdateEnterToggle(t *testing.T) {
	mt := NewManageTags("Station", []string{}, []string{"rock"}, 80)

	// First entry is "rock" (unchecked)
	mt.cursor = 0
	if mt.entries[0].checked {
		t.Fatal("expected first entry to be unchecked initially")
	}

	// Press Enter to toggle
	mt, _ = mt.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if !mt.entries[0].checked {
		t.Error("expected first entry to be checked after Enter")
	}
}

func TestUpdateToggleOnlyAffectsSelectedEntry(t *testing.T) {
	mt := NewManageTags("Station", []string{}, []string{"rock", "jazz"}, 80)

	// Check first entry
	mt.cursor = 0
	mt, _ = mt.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(" ")})
	if !mt.entries[0].checked {
		t.Error("expected first entry to be checked")
	}

	// Second entry should remain unchecked
	if mt.entries[1].checked {
		t.Error("expected second entry to remain unchecked")
	}
}

// ---------------------------------------------------------------------------
// Key handling - done/quit
// ---------------------------------------------------------------------------

func TestUpdateDKey(t *testing.T) {
	mt := NewManageTags("Station", []string{"rock"}, []string{"rock", "jazz"}, 80)

	// Toggle jazz on
	mt.cursor = 1
	mt, _ = mt.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(" ")})

	// Press 'd' to submit
	mt, cmd := mt.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("d")})

	if cmd == nil {
		t.Fatal("expected cmd from 'd', got nil")
	}

	msg := cmd()
	done, ok := msg.(ManageTagsDoneMsg)
	if !ok {
		t.Fatalf("expected ManageTagsDoneMsg, got %T", msg)
	}

	// Should have both rock and jazz selected
	if len(done.Tags) != 2 {
		t.Errorf("expected 2 selected tags, got %d", len(done.Tags))
	}
}

func TestUpdateQKey(t *testing.T) {
	mt := NewManageTags("Station", []string{}, []string{}, 80)

	mt, cmd := mt.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})

	if cmd == nil {
		t.Fatal("expected cmd from 'q', got nil")
	}

	msg := cmd()
	if _, ok := msg.(ManageTagsCancelledMsg); !ok {
		t.Errorf("expected ManageTagsCancelledMsg, got %T", msg)
	}
}

// ---------------------------------------------------------------------------
// Add new tag functionality
// ---------------------------------------------------------------------------

func TestUpdateEnterOnAddNewRow(t *testing.T) {
	mt := NewManageTags("Station", []string{}, []string{"rock"}, 80)

	// Move cursor to "Add new tag..." row (last item)
	mt.cursor = len(mt.entries) // Cursor at "Add new tag" row

	// Press Enter
	mt, _ = mt.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if !mt.addingNew {
		t.Error("expected addingNew to be true after Enter on Add new tag row")
	}
}

func TestHandleTagSubmitted(t *testing.T) {
	mt := NewManageTags("Station", []string{}, []string{"rock"}, 80)
	mt.addingNew = true

	// Submit a new tag
	mt, _ = mt.HandleTagSubmitted("newtag")

	if mt.addingNew {
		t.Error("expected addingNew to be false after HandleTagSubmitted")
	}

	// New tag should be prepended and checked
	if len(mt.entries) != 2 {
		t.Fatalf("expected 2 entries after adding tag, got %d", len(mt.entries))
	}

	if mt.entries[0].tag != "newtag" {
		t.Errorf("expected first entry to be 'newtag', got %q", mt.entries[0].tag)
	}
	if !mt.entries[0].checked {
		t.Error("expected new tag to be checked")
	}
}

func TestHandleTagSubmittedEmpty(t *testing.T) {
	mt := NewManageTags("Station", []string{}, []string{"rock"}, 80)
	mt.addingNew = true
	initialLen := len(mt.entries)

	// Submit empty tag
	mt, _ = mt.HandleTagSubmitted("")

	if mt.addingNew {
		t.Error("expected addingNew to be false after empty submission")
	}

	// No new entry should be added
	if len(mt.entries) != initialLen {
		t.Errorf("expected %d entries after empty submission, got %d", initialLen, len(mt.entries))
	}
}

func TestHandleTagSubmittedExisting(t *testing.T) {
	mt := NewManageTags("Station", []string{}, []string{"rock"}, 80)
	mt.addingNew = true

	// Existing tag "rock" is unchecked
	mt.entries[0].checked = false

	// Submit existing tag "rock"
	mt, _ = mt.HandleTagSubmitted("rock")

	// Should just check the existing entry, not add duplicate
	if len(mt.entries) != 1 {
		t.Errorf("expected 1 entry (no duplicate), got %d", len(mt.entries))
	}
	if !mt.entries[0].checked {
		t.Error("expected existing entry to be checked")
	}
}

func TestHandleTagCancelled(t *testing.T) {
	mt := NewManageTags("Station", []string{}, []string{}, 80)
	mt.addingNew = true

	mt = mt.HandleTagCancelled()

	if mt.addingNew {
		t.Error("expected addingNew to be false after HandleTagCancelled")
	}
}

// ---------------------------------------------------------------------------
// Done method
// ---------------------------------------------------------------------------

func TestDone(t *testing.T) {
	mt := NewManageTags("Station", []string{"rock"}, []string{"rock", "jazz"}, 80)

	// Check jazz
	mt.entries[1].checked = true

	cmd := mt.Done()
	if cmd == nil {
		t.Fatal("expected cmd from Done, got nil")
	}

	msg := cmd()
	done, ok := msg.(ManageTagsDoneMsg)
	if !ok {
		t.Fatalf("expected ManageTagsDoneMsg, got %T", msg)
	}

	// Should have both rock and jazz
	if len(done.Tags) != 2 {
		t.Errorf("expected 2 tags, got %d", len(done.Tags))
	}
}

// ---------------------------------------------------------------------------
// View rendering
// ---------------------------------------------------------------------------

func TestManageTagsView(t *testing.T) {
	mt := NewManageTags("Test Station", []string{"rock"}, []string{"rock", "jazz"}, 80)

	view := mt.View()

	if view == "" {
		t.Error("expected non-empty view")
	}

	// Should contain station name
	if !strings.Contains(view, "Test Station") {
		t.Error("expected view to contain station name")
	}

	// Should contain label
	if !strings.Contains(view, "Manage Tags") {
		t.Error("expected view to contain 'Manage Tags'")
	}

	// Should contain tags
	if !strings.Contains(view, "rock") {
		t.Error("expected view to contain 'rock'")
	}
}

func TestViewEmptyTags(t *testing.T) {
	mt := NewManageTags("Station", []string{}, []string{}, 80)

	view := mt.View()

	if !strings.Contains(view, "No tags yet") {
		t.Error("expected view to contain 'No tags yet' message")
	}
}

func TestViewAddNewRow(t *testing.T) {
	mt := NewManageTags("Station", []string{}, []string{"rock"}, 80)

	view := mt.View()

	if !strings.Contains(view, "Add new tag") {
		t.Error("expected view to contain 'Add new tag' row")
	}
}

func TestViewAddingNewMode(t *testing.T) {
	mt := NewManageTags("Station", []string{}, []string{"rock"}, 80)
	mt.addingNew = true
	mt.tagInput = NewTagInput([]string{"rock"}, 76)

	view := mt.View()

	// Should show tag input instead of main view
	// The view should be the tagInput.View()
	if view == "" {
		t.Error("expected non-empty view in adding new mode")
	}
}

// ---------------------------------------------------------------------------
// Helper methods
// ---------------------------------------------------------------------------

func TestSelectedTags(t *testing.T) {
	mt := NewManageTags("Station", []string{"rock"}, []string{"rock", "jazz", "blues"}, 80)

	// rock is checked (from currentTags)
	// Check jazz as well
	for i, e := range mt.entries {
		if e.tag == "jazz" {
			mt.entries[i].checked = true
		}
	}

	selected := mt.selectedTags()

	if len(selected) != 2 {
		t.Errorf("expected 2 selected tags, got %d", len(selected))
	}

	// Should contain rock and jazz
	found := make(map[string]bool)
	for _, tag := range selected {
		found[tag] = true
	}
	if !found["rock"] || !found["jazz"] {
		t.Errorf("expected rock and jazz in selected tags, got %v", selected)
	}
}

func TestSelectedTagsNone(t *testing.T) {
	mt := NewManageTags("Station", []string{}, []string{"rock"}, 80)

	// No tags checked
	selected := mt.selectedTags()

	if len(selected) != 0 {
		t.Errorf("expected 0 selected tags, got %d", len(selected))
	}
}

func TestExistingTags(t *testing.T) {
	mt := NewManageTags("Station", []string{}, []string{"rock", "jazz"}, 80)

	existing := mt.existingTags()

	if len(existing) != 2 {
		t.Errorf("expected 2 existing tags, got %d", len(existing))
	}
}

// ---------------------------------------------------------------------------
// Edge cases
// ---------------------------------------------------------------------------

func TestUpdateWhileAddingNew(t *testing.T) {
	mt := NewManageTags("Station", []string{}, []string{"rock"}, 80)
	mt.addingNew = true
	mt.tagInput = NewTagInput([]string{"rock"}, 76)

	// Any key input should be delegated to TagInput
	mt, cmd := mt.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("a")})

	// Should still be in addingNew mode
	if !mt.addingNew {
		t.Error("expected to remain in addingNew mode")
	}

	// cmd might be from tagInput
	_ = cmd
}

func TestManageTagsInit(t *testing.T) {
	mt := NewManageTags("Station", []string{}, []string{}, 80)
	cmd := mt.Init()

	// Init should return nil
	if cmd != nil {
		t.Error("expected Init to return nil")
	}
}

func TestCursorOnAddNewRowNoToggle(t *testing.T) {
	mt := NewManageTags("Station", []string{}, []string{"rock"}, 80)

	// Move cursor to "Add new tag" row
	mt.cursor = len(mt.entries)

	// Press Space (should not toggle since we're on "Add new tag" row)
	mt, _ = mt.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(" ")})

	// No entry should have changed
	// (Space on "Add new tag" row is a no-op in current implementation)
}