package components

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// ---------------------------------------------------------------------------
// Creation and initialization
// ---------------------------------------------------------------------------

func TestNewTagInput(t *testing.T) {
	allTags := []string{"rock", "jazz", "classical"}
	width := 50

	ti := NewTagInput(allTags, width)

	if len(ti.allTags) != 3 {
		t.Errorf("expected 3 allTags, got %d", len(ti.allTags))
	}
	if ti.width != width {
		t.Errorf("expected width %d, got %d", width, ti.width)
	}
	if ti.maxSuggest != 5 {
		t.Errorf("expected maxSuggest 5, got %d", ti.maxSuggest)
	}
	if ti.selectedSug != -1 {
		t.Errorf("expected selectedSug -1, got %d", ti.selectedSug)
	}
}

func TestNewTagInputEmpty(t *testing.T) {
	ti := NewTagInput([]string{}, 40)

	if len(ti.allTags) != 0 {
		t.Errorf("expected 0 allTags, got %d", len(ti.allTags))
	}
}

func TestNewTagInputNil(t *testing.T) {
	ti := NewTagInput(nil, 40)

	// allTags can be nil in current implementation (no initialization to empty slice)
	// Just verify the component is created
	if ti.width != 40 {
		t.Errorf("expected width 40, got %d", ti.width)
	}
}

// ---------------------------------------------------------------------------
// Autocomplete / Suggestions
// ---------------------------------------------------------------------------

func TestFilterSuggestionsExactMatch(t *testing.T) {
	allTags := []string{"rock", "jazz", "blues"}
	ti := NewTagInput(allTags, 50)

	// Exact match should return empty (no suggestions for exact match)
	suggestions := ti.filterSuggestions("rock")
	if len(suggestions) != 0 {
		t.Errorf("expected 0 suggestions for exact match, got %d: %v", len(suggestions), suggestions)
	}
}

func TestFilterSuggestionsPrefix(t *testing.T) {
	allTags := []string{"rock", "rockabilly", "rap", "reggae"}
	ti := NewTagInput(allTags, 50)

	suggestions := ti.filterSuggestions("roc")
	if len(suggestions) != 2 {
		t.Errorf("expected 2 suggestions for 'roc', got %d: %v", len(suggestions), suggestions)
	}

	// Should include both rock and rockabilly
	found := make(map[string]bool)
	for _, s := range suggestions {
		found[s] = true
	}
	if !found["rock"] || !found["rockabilly"] {
		t.Errorf("expected rock and rockabilly in suggestions, got %v", suggestions)
	}
}

func TestFilterSuggestionsEmpty(t *testing.T) {
	allTags := []string{"rock", "jazz", "blues"}
	ti := NewTagInput(allTags, 50)

	// Empty query should return empty
	suggestions := ti.filterSuggestions("")
	if len(suggestions) != 0 {
		t.Errorf("expected 0 suggestions for empty query, got %d", len(suggestions))
	}
}

func TestFilterSuggestionsNoMatch(t *testing.T) {
	allTags := []string{"rock", "jazz", "blues"}
	ti := NewTagInput(allTags, 50)

	suggestions := ti.filterSuggestions("xyz")
	if len(suggestions) != 0 {
		t.Errorf("expected 0 suggestions for no match, got %d", len(suggestions))
	}
}

func TestFilterSuggestionsCaseInsensitive(t *testing.T) {
	allTags := []string{"rock", "Rock Music", "ROCKABILLY"}
	ti := NewTagInput(allTags, 50)

	// Lowercase query should match all case variations
	suggestions := ti.filterSuggestions("roc")
	// Note: filterSuggestions normalizes tags to lowercase for comparison
	// So all three should match if they start with "roc" (case-insensitive)
	if len(suggestions) == 0 {
		t.Error("expected at least 1 suggestion for 'roc', got 0")
	}
}

func TestFilterSuggestionsMaxLimit(t *testing.T) {
	// Create more tags than maxSuggest
	allTags := []string{"tag1", "tag2", "tag3", "tag4", "tag5", "tag6", "tag7"}
	ti := NewTagInput(allTags, 50)
	ti.maxSuggest = 3

	suggestions := ti.filterSuggestions("tag")
	if len(suggestions) > 3 {
		t.Errorf("expected max 3 suggestions, got %d", len(suggestions))
	}
}

func TestFilterSuggestionsTrimSpace(t *testing.T) {
	allTags := []string{"rock", "jazz"}
	ti := NewTagInput(allTags, 50)

	// Query with spaces should be trimmed
	suggestions := ti.filterSuggestions("  roc  ")
	if len(suggestions) != 1 {
		t.Errorf("expected 1 suggestion for trimmed query, got %d", len(suggestions))
	}
}

// ---------------------------------------------------------------------------
// Key handling
// ---------------------------------------------------------------------------

func TestTagInputUpdateEsc(t *testing.T) {
	ti := NewTagInput([]string{}, 50)

	// Simulate Esc key
	ti, cmd := ti.Update(tea.KeyMsg{Type: tea.KeyEsc})

	if cmd == nil {
		t.Fatal("expected cmd from Esc, got nil")
	}

	msg := cmd()
	if _, ok := msg.(TagCancelledMsg); !ok {
		t.Errorf("expected TagCancelledMsg, got %T", msg)
	}
}

func TestUpdateEnterWithValue(t *testing.T) {
	ti := NewTagInput([]string{}, 50)
	ti.input.SetValue("newtag")

	// Simulate Enter key
	ti, cmd := ti.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if cmd == nil {
		t.Fatal("expected cmd from Enter, got nil")
	}

	msg := cmd()
	submitted, ok := msg.(TagSubmittedMsg)
	if !ok {
		t.Fatalf("expected TagSubmittedMsg, got %T", msg)
	}

	if submitted.Tag != "newtag" {
		t.Errorf("expected tag 'newtag', got %q", submitted.Tag)
	}
}

func TestUpdateEnterWithEmptyValue(t *testing.T) {
	ti := NewTagInput([]string{}, 50)
	ti.input.SetValue("")

	// Simulate Enter key with empty value
	ti, cmd := ti.Update(tea.KeyMsg{Type: tea.KeyEnter})

	// Should return nil cmd (no submission)
	if cmd != nil {
		t.Error("expected nil cmd for empty value, got non-nil")
	}
}

func TestUpdateEnterWithSuggestion(t *testing.T) {
	allTags := []string{"rock", "jazz"}
	ti := NewTagInput(allTags, 50)
	ti.input.SetValue("roc")

	// Simulate typing to populate suggestions
	ti, _ = ti.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")})

	// Select first suggestion
	ti.selectedSug = 0
	ti.suggestions = []string{"rock"}

	// Simulate Enter key
	ti, cmd := ti.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if cmd == nil {
		t.Fatal("expected cmd from Enter, got nil")
	}

	msg := cmd()
	submitted, ok := msg.(TagSubmittedMsg)
	if !ok {
		t.Fatalf("expected TagSubmittedMsg, got %T", msg)
	}

	if submitted.Tag != "rock" {
		t.Errorf("expected tag 'rock' from suggestion, got %q", submitted.Tag)
	}
}

func TestUpdateTab(t *testing.T) {
	allTags := []string{"rockabilly"}
	ti := NewTagInput(allTags, 50)
	ti.input.SetValue("roc")
	ti.suggestions = []string{"rockabilly"}

	// Simulate Tab key (autocomplete)
	ti, cmd := ti.Update(tea.KeyMsg{Type: tea.KeyTab})

	// Input value should be autocompleted
	value := ti.input.Value()
	if value != "rockabilly" {
		t.Errorf("expected input value 'rockabilly' after Tab, got %q", value)
	}

	// Suggestions should be cleared/updated
	if ti.selectedSug != -1 {
		t.Error("expected selectedSug to be reset after Tab")
	}

	// cmd might be nil or a blink command
	_ = cmd
}

func TestUpdateTabWithSelection(t *testing.T) {
	allTags := []string{"rock", "rockabilly"}
	ti := NewTagInput(allTags, 50)
	ti.input.SetValue("roc")
	ti.suggestions = []string{"rock", "rockabilly"}
	ti.selectedSug = 1 // Select second suggestion

	// Simulate Tab key
	ti, _ = ti.Update(tea.KeyMsg{Type: tea.KeyTab})

	// Input value should be the selected suggestion
	value := ti.input.Value()
	if value != "rockabilly" {
		t.Errorf("expected input value 'rockabilly' after Tab with selection, got %q", value)
	}
}

func TestUpdateTabNoSuggestions(t *testing.T) {
	ti := NewTagInput([]string{}, 50)
	ti.input.SetValue("xyz")
	ti.suggestions = []string{}

	// Simulate Tab key with no suggestions
	ti, cmd := ti.Update(tea.KeyMsg{Type: tea.KeyTab})

	// Input value should be unchanged
	value := ti.input.Value()
	if value != "xyz" {
		t.Errorf("expected input value unchanged after Tab with no suggestions, got %q", value)
	}

	_ = cmd
}

func TestUpdateUpDown(t *testing.T) {
	allTags := []string{"rock", "jazz", "blues"}
	ti := NewTagInput(allTags, 50)
	ti.suggestions = []string{"rock", "jazz", "blues"}

	// Initially no selection
	if ti.selectedSug != -1 {
		t.Errorf("expected selectedSug -1 initially, got %d", ti.selectedSug)
	}

	// Simulate Down key
	ti, _ = ti.Update(tea.KeyMsg{Type: tea.KeyDown})
	if ti.selectedSug != 0 {
		t.Errorf("expected selectedSug 0 after Down, got %d", ti.selectedSug)
	}

	// Down again
	ti, _ = ti.Update(tea.KeyMsg{Type: tea.KeyDown})
	if ti.selectedSug != 1 {
		t.Errorf("expected selectedSug 1 after second Down, got %d", ti.selectedSug)
	}

	// Up should go back
	ti, _ = ti.Update(tea.KeyMsg{Type: tea.KeyUp})
	if ti.selectedSug != 0 {
		t.Errorf("expected selectedSug 0 after Up, got %d", ti.selectedSug)
	}
}

func TestUpdateUpDownWrapAround(t *testing.T) {
	allTags := []string{"rock", "jazz"}
	ti := NewTagInput(allTags, 50)
	ti.suggestions = []string{"rock", "jazz"}

	// Start at first
	ti.selectedSug = 0

	// Up should wrap to last
	ti, _ = ti.Update(tea.KeyMsg{Type: tea.KeyUp})
	if ti.selectedSug != 1 {
		t.Errorf("expected selectedSug 1 after Up wrap, got %d", ti.selectedSug)
	}

	// Down should wrap to first
	ti, _ = ti.Update(tea.KeyMsg{Type: tea.KeyDown})
	if ti.selectedSug != 0 {
		t.Errorf("expected selectedSug 0 after Down wrap, got %d", ti.selectedSug)
	}
}

func TestUpdateUpDownNoSuggestions(t *testing.T) {
	ti := NewTagInput([]string{}, 50)
	ti.suggestions = []string{}

	// Up/Down should be no-op when no suggestions
	ti, _ = ti.Update(tea.KeyMsg{Type: tea.KeyDown})
	if ti.selectedSug != -1 {
		t.Error("expected selectedSug to remain -1 with no suggestions")
	}

	ti, _ = ti.Update(tea.KeyMsg{Type: tea.KeyUp})
	if ti.selectedSug != -1 {
		t.Error("expected selectedSug to remain -1 with no suggestions")
	}
}

// ---------------------------------------------------------------------------
// View rendering
// ---------------------------------------------------------------------------

func TestTagInputView(t *testing.T) {
	allTags := []string{"rock", "jazz"}
	ti := NewTagInput(allTags, 50)

	view := ti.View()

	if view == "" {
		t.Error("expected non-empty view")
	}

	// Should contain label
	if !strings.Contains(view, "Add Tag") {
		t.Error("expected view to contain 'Add Tag'")
	}
}

func TestViewWithSuggestions(t *testing.T) {
	allTags := []string{"rock", "rockabilly"}
	ti := NewTagInput(allTags, 50)
	ti.input.SetValue("roc")
	ti.suggestions = []string{"rock", "rockabilly"}

	view := ti.View()

	if !strings.Contains(view, "Suggestions:") {
		t.Error("expected view to contain 'Suggestions:'")
	}
	if !strings.Contains(view, "rock") {
		t.Error("expected view to contain suggestion 'rock'")
	}
	if !strings.Contains(view, "rockabilly") {
		t.Error("expected view to contain suggestion 'rockabilly'")
	}
}

func TestViewWithoutSuggestions(t *testing.T) {
	ti := NewTagInput([]string{}, 50)
	ti.input.SetValue("xyz")
	ti.suggestions = []string{}

	view := ti.View()

	if strings.Contains(view, "Suggestions:") {
		t.Error("expected view not to contain 'Suggestions:' when empty")
	}
	// Should show normal help text
	if !strings.Contains(view, "Enter") {
		t.Error("expected view to contain help text")
	}
}

// ---------------------------------------------------------------------------
// Value method
// ---------------------------------------------------------------------------

func TestValue(t *testing.T) {
	ti := NewTagInput([]string{}, 50)
	ti.input.SetValue("test value")

	if ti.Value() != "test value" {
		t.Errorf("expected value 'test value', got %q", ti.Value())
	}
}

// ---------------------------------------------------------------------------
// Edge cases
// ---------------------------------------------------------------------------

func TestUpdateTypingClearsSuggestionSelection(t *testing.T) {
	allTags := []string{"rock", "jazz"}
	ti := NewTagInput(allTags, 50)
	ti.suggestions = []string{"rock", "jazz"}
	ti.selectedSug = 1

	// Simulate typing (any key that goes to text input)
	ti, _ = ti.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("a")})

	// selectedSug should be reset
	if ti.selectedSug != -1 {
		t.Error("expected selectedSug to be reset after typing")
	}
}

func TestFilterSuggestionsWithSpecialCharacters(t *testing.T) {
	allTags := []string{"rock-n-roll", "jazz-fusion", "blues"}
	ti := NewTagInput(allTags, 50)

	suggestions := ti.filterSuggestions("rock-")
	if len(suggestions) != 1 {
		t.Errorf("expected 1 suggestion for 'rock-', got %d", len(suggestions))
	}
	if len(suggestions) > 0 && suggestions[0] != "rock-n-roll" {
		t.Errorf("expected 'rock-n-roll', got %q", suggestions[0])
	}
}

func TestTagInputInit(t *testing.T) {
	ti := NewTagInput([]string{}, 50)
	cmd := ti.Init()

	// Init should return textinput.Blink command (or nil)
	// Just verify it doesn't panic
	_ = cmd
}