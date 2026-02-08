package ui

import (
	"fmt"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/shinokada/tera/v3/internal/api"
	"github.com/shinokada/tera/v3/internal/blocklist"
	"github.com/shinokada/tera/v3/internal/ui/components"
)

func TestNewLuckyModel(t *testing.T) {
	client := api.NewClient()
	model := NewLuckyModel(client, "/tmp/test", blocklist.NewManager("/tmp/blocklist"))

	if model.state != luckyStateInput {
		t.Errorf("Expected initial state to be luckyStateInput, got %v", model.state)
	}

	if model.apiClient == nil {
		t.Error("Expected apiClient to be set")
	}

	if model.favoritePath != "/tmp/test" {
		t.Errorf("Expected favoritePath to be /tmp/test, got %s", model.favoritePath)
	}

	if model.player == nil {
		t.Error("Expected player to be initialized")
	}

	if model.textInput.Placeholder != "rock, jazz, classical, meditation..." {
		t.Errorf("Expected placeholder text, got %s", model.textInput.Placeholder)
	}

	if model.textInput.CharLimit != 50 {
		t.Errorf("Expected char limit 50, got %d", model.textInput.CharLimit)
	}

	if model.width != 80 {
		t.Errorf("Expected default width 80, got %d", model.width)
	}

	if model.height != 24 {
		t.Errorf("Expected default height 24, got %d", model.height)
	}
}

func TestLuckyModelInit(t *testing.T) {
	client := api.NewClient()
	model := NewLuckyModel(client, "/tmp/test", blocklist.NewManager("/tmp/blocklist"))

	cmd := model.Init()
	if cmd == nil {
		t.Error("Expected Init to return a command (textinput.Blink)")
	}
}

func TestLuckyInputStateEscNavigation(t *testing.T) {
	client := api.NewClient()
	model := NewLuckyModel(client, "/tmp/test", blocklist.NewManager("/tmp/blocklist"))
	model.state = luckyStateInput

	msg := tea.KeyMsg{Type: tea.KeyEsc}
	updatedModel, cmd := model.Update(msg)

	luckyModel := updatedModel.(LuckyModel)
	// State shouldn't change immediately - the command should trigger navigation
	if luckyModel.state != luckyStateInput {
		t.Errorf("Expected state to remain luckyStateInput, got %v", luckyModel.state)
	}

	if cmd == nil {
		t.Error("Expected command to be returned for navigation")
	}

	// Execute the command to verify it returns navigateMsg
	resultMsg := cmd()
	if navMsg, ok := resultMsg.(navigateMsg); ok {
		if navMsg.screen != screenMainMenu {
			t.Errorf("Expected navigation to screenMainMenu, got %v", navMsg.screen)
		}
	} else {
		t.Error("Expected navigateMsg from command")
	}
}

func TestLuckyInputStateEmptyKeyword(t *testing.T) {
	client := api.NewClient()
	model := NewLuckyModel(client, "/tmp/test", blocklist.NewManager("/tmp/blocklist"))
	model.state = luckyStateInput
	model.textInput.SetValue("")

	msg := tea.KeyMsg{Type: tea.KeyEnter}
	updatedModel, _ := model.Update(msg)

	luckyModel := updatedModel.(LuckyModel)
	if luckyModel.err == nil {
		t.Error("Expected error for empty keyword")
	}

	if luckyModel.err.Error() != "please enter a keyword" {
		t.Errorf("Expected 'please enter a keyword' error, got '%s'", luckyModel.err.Error())
	}

	// State should remain in input
	if luckyModel.state != luckyStateInput {
		t.Errorf("Expected state to remain luckyStateInput, got %v", luckyModel.state)
	}
}

func TestLuckyInputStateValidKeyword(t *testing.T) {
	client := api.NewClient()
	model := NewLuckyModel(client, "/tmp/test", blocklist.NewManager("/tmp/blocklist"))
	model.state = luckyStateInput
	model.textInput.SetValue("jazz")

	msg := tea.KeyMsg{Type: tea.KeyEnter}
	updatedModel, cmd := model.Update(msg)

	luckyModel := updatedModel.(LuckyModel)
	if luckyModel.err != nil {
		t.Errorf("Expected no error, got %v", luckyModel.err)
	}

	if luckyModel.state != luckyStateSearching {
		t.Errorf("Expected state to be luckyStateSearching, got %v", luckyModel.state)
	}

	if cmd == nil {
		t.Error("Expected search command to be returned")
	}
}

func TestLuckyInputStateKeywordWithWhitespace(t *testing.T) {
	client := api.NewClient()
	model := NewLuckyModel(client, "/tmp/test", blocklist.NewManager("/tmp/blocklist"))
	model.state = luckyStateInput
	model.textInput.SetValue("   rock   ")

	msg := tea.KeyMsg{Type: tea.KeyEnter}
	updatedModel, cmd := model.Update(msg)

	luckyModel := updatedModel.(LuckyModel)
	// Should trim and proceed
	if luckyModel.state != luckyStateSearching {
		t.Errorf("Expected state to be luckyStateSearching, got %v", luckyModel.state)
	}

	if cmd == nil {
		t.Error("Expected search command to be returned")
	}
}

func TestLuckyInputStateWhitespaceOnlyKeyword(t *testing.T) {
	client := api.NewClient()
	model := NewLuckyModel(client, "/tmp/test", blocklist.NewManager("/tmp/blocklist"))
	model.state = luckyStateInput
	model.textInput.SetValue("   ")

	msg := tea.KeyMsg{Type: tea.KeyEnter}
	updatedModel, _ := model.Update(msg)

	luckyModel := updatedModel.(LuckyModel)
	if luckyModel.err == nil {
		t.Error("Expected error for whitespace-only keyword")
	}

	if luckyModel.state != luckyStateInput {
		t.Errorf("Expected state to remain luckyStateInput, got %v", luckyModel.state)
	}
}

func TestLuckyPlayingStateEscNavigation(t *testing.T) {
	client := api.NewClient()
	model := NewLuckyModel(client, "/tmp/test", blocklist.NewManager("/tmp/blocklist"))
	model.state = luckyStatePlaying
	model.selectedStation = &api.Station{Name: "Test Station"}

	msg := tea.KeyMsg{Type: tea.KeyEsc}
	updatedModel, cmd := model.Update(msg)

	luckyModel := updatedModel.(LuckyModel)
	// ESC during playback should stop and return to lucky input state
	if luckyModel.state != luckyStateInput {
		t.Errorf("Expected state to be luckyStateInput, got %v", luckyModel.state)
	}

	// ESC returns to input state without navigation command (similar to search.go behavior)
	if luckyModel.selectedStation != nil {
		t.Error("Expected selectedStation to be nil after pressing ESC")
	}

	// No navigation command is expected - ESC just returns to input state
	_ = cmd // cmd may be nil, which is expected
}

func TestLuckyPlayingStateZeroToMainMenu(t *testing.T) {
	client := api.NewClient()
	model := NewLuckyModel(client, "/tmp/test", blocklist.NewManager("/tmp/blocklist"))
	model.state = luckyStatePlaying
	model.selectedStation = &api.Station{Name: "Test Station"}

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("0")}
	updatedModel, cmd := model.Update(msg)

	luckyModel := updatedModel.(LuckyModel)
	// Station should be cleared
	if luckyModel.selectedStation != nil {
		t.Error("Expected selectedStation to be nil after pressing 0")
	}

	if cmd == nil {
		t.Error("Expected command to be returned for navigation")
	}

	// Verify navigation to main menu
	resultMsg := cmd()
	if navMsg, ok := resultMsg.(navigateMsg); ok {
		if navMsg.screen != screenMainMenu {
			t.Errorf("Expected navigation to screenMainMenu, got %v", navMsg.screen)
		}
	} else {
		t.Error("Expected navigateMsg from command")
	}
}

func TestLuckyPlayingStateFavoriteShortcut(t *testing.T) {
	client := api.NewClient()
	model := NewLuckyModel(client, "/tmp/test", blocklist.NewManager("/tmp/blocklist"))
	model.state = luckyStatePlaying
	model.selectedStation = &api.Station{Name: "Test Station"}

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("f")}
	_, cmd := model.Update(msg)

	if cmd == nil {
		t.Error("Expected save command to be returned")
	}
}

func TestLuckyPlayingStateSaveToListShortcut(t *testing.T) {
	client := api.NewClient()
	model := NewLuckyModel(client, "/tmp/test", blocklist.NewManager("/tmp/blocklist"))
	model.state = luckyStatePlaying
	model.selectedStation = &api.Station{Name: "Test Station"}

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("s")}
	updatedModel, cmd := model.Update(msg)

	luckyModel := updatedModel.(LuckyModel)
	if luckyModel.state != luckyStateSelectList {
		t.Errorf("Expected state to be luckyStateSelectList, got %v", luckyModel.state)
	}

	if cmd == nil {
		t.Error("Expected load lists command to be returned")
	}
}

func TestLuckyPlayingStateVoteShortcut(t *testing.T) {
	client := api.NewClient()
	model := NewLuckyModel(client, "/tmp/test", blocklist.NewManager("/tmp/blocklist"))
	model.state = luckyStatePlaying
	model.selectedStation = &api.Station{Name: "Test Station", StationUUID: "test-uuid"}

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("v")}
	_, cmd := model.Update(msg)

	if cmd == nil {
		t.Error("Expected vote command to be returned")
	}
}

func TestLuckySavePromptYes(t *testing.T) {
	client := api.NewClient()
	model := NewLuckyModel(client, "/tmp/test", blocklist.NewManager("/tmp/blocklist"))
	model.state = luckyStateSavePrompt
	model.selectedStation = &api.Station{Name: "Test Station"}

	tests := []struct {
		name string
		key  string
	}{
		{"y key", "y"},
		{"1 key", "1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model.state = luckyStateSavePrompt
			msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(tt.key)}
			_, cmd := model.Update(msg)

			if cmd == nil {
				t.Error("Expected batch command to be returned")
			}
		})
	}
}

func TestLuckySavePromptNo(t *testing.T) {
	client := api.NewClient()
	model := NewLuckyModel(client, "/tmp/test", blocklist.NewManager("/tmp/blocklist"))
	model.state = luckyStateSavePrompt
	model.selectedStation = &api.Station{Name: "Test Station"}

	tests := []struct {
		name    string
		key     string
		keyType tea.KeyType
	}{
		{"n key", "n", tea.KeyRunes},
		{"2 key", "2", tea.KeyRunes},
		{"esc key", "", tea.KeyEsc},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model.state = luckyStateSavePrompt
			var msg tea.KeyMsg
			if tt.keyType == tea.KeyEsc {
				msg = tea.KeyMsg{Type: tea.KeyEsc}
			} else {
				msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(tt.key)}
			}
			_, cmd := model.Update(msg)

			if cmd == nil {
				t.Error("Expected navigation command to be returned")
			}

			resultMsg := cmd()
			if navMsg, ok := resultMsg.(navigateMsg); ok {
				if navMsg.screen != screenMainMenu {
					t.Errorf("Expected navigation to screenMainMenu, got %v", navMsg.screen)
				}
			} else {
				t.Error("Expected navigateMsg from command")
			}
		})
	}
}

func TestLuckySelectListEscNavigation(t *testing.T) {
	client := api.NewClient()
	model := NewLuckyModel(client, "/tmp/test", blocklist.NewManager("/tmp/blocklist"))
	model.state = luckyStateSelectList
	model.selectedStation = &api.Station{Name: "Test Station"}

	msg := tea.KeyMsg{Type: tea.KeyEsc}
	updatedModel, _ := model.Update(msg)

	luckyModel := updatedModel.(LuckyModel)
	// ESC should go back to playing state
	if luckyModel.state != luckyStatePlaying {
		t.Errorf("Expected state to be luckyStatePlaying, got %v", luckyModel.state)
	}
}

func TestLuckySelectListNewListShortcut(t *testing.T) {
	client := api.NewClient()
	model := NewLuckyModel(client, "/tmp/test", blocklist.NewManager("/tmp/blocklist"))
	model.state = luckyStateSelectList
	model.selectedStation = &api.Station{Name: "Test Station"}

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("n")}
	updatedModel, cmd := model.Update(msg)

	luckyModel := updatedModel.(LuckyModel)
	// N should go to new list input state
	if luckyModel.state != luckyStateNewListInput {
		t.Errorf("Expected state to be luckyStateNewListInput, got %v", luckyModel.state)
	}

	if cmd == nil {
		t.Error("Expected textinput.Blink command to be returned")
	}
}

func TestLuckyNewListInputEscNavigation(t *testing.T) {
	client := api.NewClient()
	model := NewLuckyModel(client, "/tmp/test", blocklist.NewManager("/tmp/blocklist"))
	model.state = luckyStateNewListInput
	model.selectedStation = &api.Station{Name: "Test Station"}

	msg := tea.KeyMsg{Type: tea.KeyEsc}
	updatedModel, _ := model.Update(msg)

	luckyModel := updatedModel.(LuckyModel)
	// ESC should go back to select list state
	if luckyModel.state != luckyStateSelectList {
		t.Errorf("Expected state to be luckyStateSelectList, got %v", luckyModel.state)
	}
}

func TestLuckyWindowSizeUpdate(t *testing.T) {
	client := api.NewClient()
	model := NewLuckyModel(client, "/tmp/test", blocklist.NewManager("/tmp/blocklist"))

	msg := tea.WindowSizeMsg{Width: 120, Height: 40}
	updatedModel, _ := model.Update(msg)

	luckyModel := updatedModel.(LuckyModel)
	if luckyModel.width != 120 {
		t.Errorf("Expected width 120, got %d", luckyModel.width)
	}
	if luckyModel.height != 40 {
		t.Errorf("Expected height 40, got %d", luckyModel.height)
	}
}

func TestLuckySearchResultsMsg(t *testing.T) {
	client := api.NewClient()
	model := NewLuckyModel(client, "/tmp/test", blocklist.NewManager("/tmp/blocklist"))
	model.state = luckyStateSearching

	station := &api.Station{Name: "Found Station", URLResolved: "http://example.com/stream"}
	msg := luckySearchResultsMsg{station: station}
	updatedModel, cmd := model.Update(msg)

	luckyModel := updatedModel.(LuckyModel)
	if luckyModel.selectedStation == nil {
		t.Error("Expected selectedStation to be set")
	}

	if luckyModel.selectedStation.Name != "Found Station" {
		t.Errorf("Expected station name 'Found Station', got '%s'", luckyModel.selectedStation.Name)
	}

	if luckyModel.state != luckyStatePlaying {
		t.Errorf("Expected state to be luckyStatePlaying, got %v", luckyModel.state)
	}

	if cmd == nil {
		t.Error("Expected playback command to be returned")
	}
}

func TestLuckySearchErrorMsg(t *testing.T) {
	client := api.NewClient()
	model := NewLuckyModel(client, "/tmp/test", blocklist.NewManager("/tmp/blocklist"))
	model.state = luckyStateSearching

	msg := luckySearchErrorMsg{err: fmt.Errorf("no stations found")}
	updatedModel, _ := model.Update(msg)

	luckyModel := updatedModel.(LuckyModel)
	if luckyModel.err == nil {
		t.Error("Expected error to be set")
	}

	if luckyModel.state != luckyStateInput {
		t.Errorf("Expected state to return to luckyStateInput, got %v", luckyModel.state)
	}
}

func TestLuckySaveSuccessMsg(t *testing.T) {
	client := api.NewClient()
	model := NewLuckyModel(client, "/tmp/test", blocklist.NewManager("/tmp/blocklist"))
	model.state = luckyStatePlaying

	station := &api.Station{Name: "Test Station"}
	msg := saveSuccessMsg{station: station}
	updatedModel, _ := model.Update(msg)

	luckyModel := updatedModel.(LuckyModel)
	if luckyModel.saveMessage == "" {
		t.Error("Expected save message to be set")
	}

	if luckyModel.saveMessageTime != 3 {
		t.Errorf("Expected saveMessageTime 3 (3 seconds), got %d", luckyModel.saveMessageTime)
	}
}

func TestLuckySaveFailedMsgDuplicate(t *testing.T) {
	client := api.NewClient()
	model := NewLuckyModel(client, "/tmp/test", blocklist.NewManager("/tmp/blocklist"))
	model.state = luckyStatePlaying

	msg := saveFailedMsg{err: fmt.Errorf("duplicate"), isDuplicate: true}
	updatedModel, _ := model.Update(msg)

	luckyModel := updatedModel.(LuckyModel)
	if luckyModel.saveMessage != "Already in Quick Favorites" {
		t.Errorf("Expected duplicate message, got '%s'", luckyModel.saveMessage)
	}
}

func TestLuckyVoteSuccessMsg(t *testing.T) {
	client := api.NewClient()
	model := NewLuckyModel(client, "/tmp/test", blocklist.NewManager("/tmp/blocklist"))
	model.state = luckyStatePlaying

	msg := components.VoteSuccessMsg{Message: "Voted for Test Station"}
	updatedModel, _ := model.Update(msg)

	luckyModel := updatedModel.(LuckyModel)
	if luckyModel.saveMessage == "" {
		t.Error("Expected message to be set")
	}

	if luckyModel.saveMessageTime != 3 {
		t.Errorf("Expected saveMessageTime 3 (3 seconds), got %d", luckyModel.saveMessageTime)
	}
}

func TestLuckyListsLoadedMsg(t *testing.T) {
	client := api.NewClient()
	model := NewLuckyModel(client, "/tmp/test", blocklist.NewManager("/tmp/blocklist"))
	model.state = luckyStateSelectList
	model.width = 80
	model.height = 24

	msg := listsLoadedMsg{lists: []string{"My-favorites", "Jazz", "Classical"}}
	updatedModel, _ := model.Update(msg)

	luckyModel := updatedModel.(LuckyModel)
	if len(luckyModel.availableLists) != 3 {
		t.Errorf("Expected 3 lists, got %d", len(luckyModel.availableLists))
	}

	if len(luckyModel.listItems) != 3 {
		t.Errorf("Expected 3 list items, got %d", len(luckyModel.listItems))
	}
}

func TestLuckyViewStates(t *testing.T) {
	client := api.NewClient()
	model := NewLuckyModel(client, "/tmp/test", blocklist.NewManager("/tmp/blocklist"))

	tests := []struct {
		name  string
		state luckyState
	}{
		{"Input state", luckyStateInput},
		{"Searching state", luckyStateSearching},
		{"Playing state", luckyStatePlaying},
		{"Save prompt state", luckyStateSavePrompt},
		{"Select list state", luckyStateSelectList},
		{"New list input state", luckyStateNewListInput},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model.state = tt.state
			if tt.state == luckyStatePlaying || tt.state == luckyStateSavePrompt || tt.state == luckyStateSelectList || tt.state == luckyStateNewListInput {
				model.selectedStation = &api.Station{Name: "Test Station", URLResolved: "http://example.com"}
			}
			view := model.View()
			if view == "" {
				t.Error("Expected non-empty view")
			}
			if view == "Unknown state" {
				t.Errorf("Got 'Unknown state' for state %v", tt.state)
			}
		})
	}
}

// Shuffle mode tests

func TestLuckyShuffleToggle(t *testing.T) {
	client := api.NewClient()
	model := NewLuckyModel(client, "/tmp/test", blocklist.NewManager("/tmp/blocklist"))
	model.state = luckyStateInput
	model.textInput.SetValue("") // Empty input

	// Initially shuffle should be disabled
	if model.shuffleEnabled {
		t.Error("Expected shuffle to be disabled initially")
	}

	// Toggle shuffle on
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("t")}
	updatedModel, _ := model.Update(msg)
	luckyModel := updatedModel.(LuckyModel)

	if !luckyModel.shuffleEnabled {
		t.Error("Expected shuffle to be enabled after pressing 't'")
	}

	// Toggle shuffle off
	msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("t")}
	updatedModel, _ = luckyModel.Update(msg)
	luckyModel = updatedModel.(LuckyModel)

	if luckyModel.shuffleEnabled {
		t.Error("Expected shuffle to be disabled after pressing 't' again")
	}
}

func TestLuckyShuffleSearchTrigger(t *testing.T) {
	client := api.NewClient()
	model := NewLuckyModel(client, "/tmp/test", blocklist.NewManager("/tmp/blocklist"))
	model.state = luckyStateInput
	model.textInput.SetValue("jazz")
	model.shuffleEnabled = true

	msg := tea.KeyMsg{Type: tea.KeyEnter}
	updatedModel, cmd := model.Update(msg)

	luckyModel := updatedModel.(LuckyModel)
	if luckyModel.state != luckyStateSearching {
		t.Errorf("Expected state to be luckyStateSearching, got %v", luckyModel.state)
	}

	if cmd == nil {
		t.Error("Expected shuffle search command to be returned")
	}
}

func TestLuckyShufflePlayingStateStopShuffle(t *testing.T) {
	client := api.NewClient()
	model := NewLuckyModel(client, "/tmp/test", blocklist.NewManager("/tmp/blocklist"))
	model.state = luckyStateShufflePlaying
	model.selectedStation = &api.Station{Name: "Test Station"}
	model.shuffleEnabled = true
	// Mock shuffle manager would be set here in real scenario

	// Press 'h' to stop shuffle but keep playing
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("h")}
	updatedModel, _ := model.Update(msg)

	luckyModel := updatedModel.(LuckyModel)
	if luckyModel.state != luckyStatePlaying {
		t.Errorf("Expected state to be luckyStatePlaying, got %v", luckyModel.state)
	}

	if luckyModel.shuffleEnabled {
		t.Error("Expected shuffle to be disabled after stopping")
	}

	if luckyModel.shuffleManager != nil {
		t.Error("Expected shuffleManager to be nil after stopping")
	}
}

func TestLuckyShufflePlayingStateEscNavigation(t *testing.T) {
	client := api.NewClient()
	model := NewLuckyModel(client, "/tmp/test", blocklist.NewManager("/tmp/blocklist"))
	model.state = luckyStateShufflePlaying
	model.selectedStation = &api.Station{Name: "Test Station"}
	model.shuffleEnabled = true

	// Press Esc to stop shuffle and return to input
	msg := tea.KeyMsg{Type: tea.KeyEsc}
	updatedModel, _ := model.Update(msg)

	luckyModel := updatedModel.(LuckyModel)
	if luckyModel.state != luckyStateInput {
		t.Errorf("Expected state to be luckyStateInput, got %v", luckyModel.state)
	}

	if luckyModel.selectedStation != nil {
		t.Error("Expected selectedStation to be nil after Esc")
	}

	if luckyModel.shuffleEnabled {
		t.Error("Expected shuffle to be disabled after Esc")
	}

	if luckyModel.shuffleManager != nil {
		t.Error("Expected shuffleManager to be nil after Esc")
	}
}

func TestLuckyShufflePlayingStateZeroToMainMenu(t *testing.T) {
	client := api.NewClient()
	model := NewLuckyModel(client, "/tmp/test", blocklist.NewManager("/tmp/blocklist"))
	model.state = luckyStateShufflePlaying
	model.selectedStation = &api.Station{Name: "Test Station"}
	model.shuffleEnabled = true

	// Press '0' to return to main menu
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("0")}
	updatedModel, cmd := model.Update(msg)

	luckyModel := updatedModel.(LuckyModel)
	if luckyModel.selectedStation != nil {
		t.Error("Expected selectedStation to be nil after pressing 0")
	}

	if luckyModel.shuffleEnabled {
		t.Error("Expected shuffle to be disabled after pressing 0")
	}

	if cmd == nil {
		t.Error("Expected command to be returned for navigation")
	}

	// Verify navigation to main menu
	resultMsg := cmd()
	if navMsg, ok := resultMsg.(navigateMsg); ok {
		if navMsg.screen != screenMainMenu {
			t.Errorf("Expected navigation to screenMainMenu, got %v", navMsg.screen)
		}
	} else {
		t.Error("Expected navigateMsg from command")
	}
}

func TestLuckyShufflePlayingStateFavoriteShortcut(t *testing.T) {
	client := api.NewClient()
	model := NewLuckyModel(client, "/tmp/test", blocklist.NewManager("/tmp/blocklist"))
	model.state = luckyStateShufflePlaying
	model.selectedStation = &api.Station{Name: "Test Station"}

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("f")}
	_, cmd := model.Update(msg)

	if cmd == nil {
		t.Error("Expected save command to be returned")
	}
}

func TestLuckyShufflePlayingStateSaveToListShortcut(t *testing.T) {
	client := api.NewClient()
	model := NewLuckyModel(client, "/tmp/test", blocklist.NewManager("/tmp/blocklist"))
	model.state = luckyStateShufflePlaying
	model.selectedStation = &api.Station{Name: "Test Station"}

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("s")}
	updatedModel, cmd := model.Update(msg)

	luckyModel := updatedModel.(LuckyModel)
	if luckyModel.state != luckyStateSelectList {
		t.Errorf("Expected state to be luckyStateSelectList, got %v", luckyModel.state)
	}

	if cmd == nil {
		t.Error("Expected load lists command to be returned")
	}
}

func TestLuckyShufflePlayingStateVoteShortcut(t *testing.T) {
	client := api.NewClient()
	model := NewLuckyModel(client, "/tmp/test", blocklist.NewManager("/tmp/blocklist"))
	model.state = luckyStateShufflePlaying
	model.selectedStation = &api.Station{Name: "Test Station", StationUUID: "test-uuid"}

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("v")}
	_, cmd := model.Update(msg)

	if cmd == nil {
		t.Error("Expected vote command to be returned")
	}
}

func TestLuckyShuffleViewShufflePlaying(t *testing.T) {
	client := api.NewClient()
	model := NewLuckyModel(client, "/tmp/test", blocklist.NewManager("/tmp/blocklist"))
	model.state = luckyStateShufflePlaying
	model.selectedStation = &api.Station{Name: "Test Station", URLResolved: "http://example.com"}
	// In real scenario, shuffleManager would be initialized
	// For this test, we just check it doesn't crash

	view := model.View()
	if view == "" {
		t.Error("Expected non-empty view")
	}
	// Without a proper shuffleManager, we expect a fallback message
	// We expect a fallback message like "No shuffle session active", but any non-empty view is acceptable here.
}
