package ui

import (
	"fmt"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/shinokada/tera/internal/api"
)

func TestSearchModelInit(t *testing.T) {
	client := api.NewClient()
	model := NewSearchModel(client, "/tmp/test")

	if model.state != searchStateMenu {
		t.Errorf("Expected initial state to be searchStateMenu, got %v", model.state)
	}

	if model.apiClient == nil {
		t.Error("Expected apiClient to be set")
	}

	if model.favoritePath != "/tmp/test" {
		t.Errorf("Expected favoritePath to be /tmp/test, got %s", model.favoritePath)
	}
}

func TestSearchMenuNavigation(t *testing.T) {
	client := api.NewClient()
	model := NewSearchModel(client, "/tmp/test")

	tests := []struct {
		name          string
		key           string
		expectedState searchState
		expectedType  api.SearchType
	}{
		{"Select Tag Search", "1", searchStateInput, api.SearchByTag},
		{"Select Name Search", "2", searchStateInput, api.SearchByName},
		{"Select Language Search", "3", searchStateInput, api.SearchByLanguage},
		{"Select Country Search", "4", searchStateInput, api.SearchByCountry},
		{"Select State Search", "5", searchStateInput, api.SearchByState},
		{"Select Advanced Search", "6", searchStateAdvancedForm, api.SearchAdvanced},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset to menu state
			model.state = searchStateMenu
			model.numberBuffer = "" // Clear number buffer

			// First send the number key (gets buffered)
			msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(tt.key)}
			updatedModel, _ := model.Update(msg)
			searchModel := updatedModel.(SearchModel)

			// Then send Enter to confirm the selection
			enterMsg := tea.KeyMsg{Type: tea.KeyEnter}
			updatedModel, _ = searchModel.Update(enterMsg)

			searchModel = updatedModel.(SearchModel)
			if searchModel.state != tt.expectedState {
				t.Errorf("Expected state %v, got %v", tt.expectedState, searchModel.state)
			}

			if searchModel.searchType != tt.expectedType {
				t.Errorf("Expected search type %v, got %v", tt.expectedType, searchModel.searchType)
			}
		})
	}
}

func TestSearchBackNavigation(t *testing.T) {
	client := api.NewClient()
	model := NewSearchModel(client, "/tmp/test")

	tests := []struct {
		name         string
		initialState searchState
		key          string
		keyType      tea.KeyType
		shouldGoBack bool
	}{
		{"Menu - Esc key", searchStateMenu, "esc", tea.KeyEsc, true},
		{"Input - Esc key", searchStateInput, "esc", tea.KeyEsc, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model.state = tt.initialState
			if tt.initialState == searchStateInput {
				model.textInput.Focus()
			}

			var msg tea.KeyMsg
			if tt.keyType == tea.KeyEsc {
				msg = tea.KeyMsg{Type: tea.KeyEsc}
			} else {
				msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(tt.key)}
			}

			updatedModel, cmd := model.Update(msg)
			searchModel := updatedModel.(SearchModel)

			if tt.shouldGoBack {
				switch tt.initialState {
				case searchStateMenu:
					// From menu, should send backToMainMsg
					if cmd == nil {
						t.Error("Expected back command, got nil")
					} else {
						result := cmd()
						if _, ok := result.(backToMainMsg); !ok {
							t.Errorf("Expected backToMainMsg, got %T", result)
						}
					}
				case searchStateInput:
					// From input, should go back to menu
					if searchModel.state != searchStateMenu {
						t.Errorf("Expected state change to menu, got state %v", searchModel.state)
					}
				}
			}
		})
	}
}

func TestSearchTextInput(t *testing.T) {
	client := api.NewClient()
	model := NewSearchModel(client, "/tmp/test")
	model.state = searchStateInput
	model.searchType = api.SearchByTag
	model.textInput.Focus()

	// Type some text
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")}
	updatedModel, _ := model.Update(msg)
	model = updatedModel.(SearchModel)

	msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("a")}
	updatedModel, _ = model.Update(msg)
	model = updatedModel.(SearchModel)

	msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("z")}
	updatedModel, _ = model.Update(msg)
	model = updatedModel.(SearchModel)

	msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("z")}
	updatedModel, _ = model.Update(msg)
	model = updatedModel.(SearchModel)

	if model.textInput.Value() != "jazz" {
		t.Errorf("Expected text input to be 'jazz', got '%s'", model.textInput.Value())
	}

	// Press enter to search
	msg = tea.KeyMsg{Type: tea.KeyEnter}
	updatedModel, cmd := model.Update(msg)
	model = updatedModel.(SearchModel)

	if model.state != searchStateLoading {
		t.Errorf("Expected state to be searchStateLoading after enter, got %v", model.state)
	}

	if cmd == nil {
		t.Error("Expected search command after enter, got nil")
	}

	// Check that text input was cleared
	if model.textInput.Value() != "" {
		t.Errorf("Expected text input to be cleared, got '%s'", model.textInput.Value())
	}
}

func TestSearchEmptyQuery(t *testing.T) {
	client := api.NewClient()
	model := NewSearchModel(client, "/tmp/test")
	model.state = searchStateInput
	model.textInput.Focus()

	// Press enter with empty input
	msg := tea.KeyMsg{Type: tea.KeyEnter}
	updatedModel, cmd := model.Update(msg)
	model = updatedModel.(SearchModel)

	// Should not trigger search
	if model.state != searchStateInput {
		t.Errorf("Expected state to remain searchStateInput with empty query, got %v", model.state)
	}

	if cmd != nil {
		t.Error("Expected no command with empty query, got command")
	}
}

func TestSearchResults(t *testing.T) {
	client := api.NewClient()
	model := NewSearchModel(client, "/tmp/test")

	// Simulate search results
	stations := []api.Station{
		{
			StationUUID: "test-1",
			Name:        "Test Jazz Station",
			Tags:        "jazz",
			Votes:       100,
		},
		{
			StationUUID: "test-2",
			Name:        "Test Rock Station",
			Tags:        "rock",
			Votes:       50,
		},
	}

	msg := searchResultsMsg{results: stations}
	updatedModel, _ := model.Update(msg)
	model = updatedModel.(SearchModel)

	if model.state != searchStateResults {
		t.Errorf("Expected state to be searchStateResults, got %v", model.state)
	}

	if len(model.results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(model.results))
	}

	if len(model.resultsItems) != 2 {
		t.Errorf("Expected 2 result items, got %d", len(model.resultsItems))
	}
}

func TestSearchError(t *testing.T) {
	client := api.NewClient()
	model := NewSearchModel(client, "/tmp/test")
	model.state = searchStateLoading

	// Simulate search error
	msg := searchErrorMsg{err: fmt.Errorf("search failed")}
	updatedModel, _ := model.Update(msg)
	model = updatedModel.(SearchModel)

	if model.state != searchStateMenu {
		t.Errorf("Expected state to return to searchStateMenu on error, got %v", model.state)
	}

	if model.err == nil {
		t.Error("Expected error to be set")
	}
}

func TestSearchStationSelection(t *testing.T) {
	client := api.NewClient()
	model := NewSearchModel(client, "/tmp/test")
	model.width = 80
	model.height = 24

	// Set up results
	stations := []api.Station{
		{
			StationUUID: "test-1",
			Name:        "Test Station",
			URLResolved: "http://example.com/stream",
			Votes:       100,
		},
	}

	// Simulate receiving search results to properly initialize the list
	msg := searchResultsMsg{results: stations}
	updatedModel, _ := model.Update(msg)
	model = updatedModel.(SearchModel)

	// Now select station
	keyMsg := tea.KeyMsg{Type: tea.KeyEnter}
	updatedModel, _ = model.Update(keyMsg)
	model = updatedModel.(SearchModel)

	if model.state != searchStatePlaying {
		t.Errorf("Expected state to be searchStatePlaying after selection, got %v", model.state)
	}

	if model.selectedStation == nil {
		t.Error("Expected selected station to be set")
		return
	}

	if model.selectedStation.StationUUID != "test-1" {
		t.Errorf("Expected selected station UUID to be test-1, got %s", model.selectedStation.StationUUID)
	}
}

func TestSearchTypeLabels(t *testing.T) {
	client := api.NewClient()
	model := NewSearchModel(client, "/tmp/test")

	tests := []struct {
		searchType    api.SearchType
		expectedLabel string
	}{
		{api.SearchByTag, "Search by Tag (genre, style, etc.)"},
		{api.SearchByName, "Search by Station Name"},
		{api.SearchByLanguage, "Search by Language"},
		{api.SearchByCountry, "Search by Country Code"},
		{api.SearchByState, "Search by State"},
		{api.SearchAdvanced, "Advanced Search (multiple criteria)"},
	}

	for _, tt := range tests {
		t.Run(tt.expectedLabel, func(t *testing.T) {
			model.searchType = tt.searchType
			label := model.getSearchTypeLabel()
			if label != tt.expectedLabel {
				t.Errorf("Expected label '%s', got '%s'", tt.expectedLabel, label)
			}
		})
	}
}

func TestStationInfoMenu(t *testing.T) {
	client := api.NewClient()
	model := NewSearchModel(client, "/tmp/test")

	station := api.Station{
		StationUUID: "test-1",
		Name:        "Test Station",
		URLResolved: "http://example.com/stream",
	}

	model.state = searchStateStationInfo
	model.selectedStation = &station

	tests := []struct {
		name          string
		key           string
		expectedState searchState
	}{
		{"Play Station", "1", searchStatePlaying},
		{"Back to Results", "3", searchStateResults},
		{"Esc to Results", "esc", searchStateResults},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model.state = searchStateStationInfo

			var msg tea.KeyMsg
			if tt.key == "esc" {
				msg = tea.KeyMsg{Type: tea.KeyEsc}
			} else {
				msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(tt.key)}
			}

			updatedModel, _ := model.Update(msg)
			model = updatedModel.(SearchModel)

			if model.state != tt.expectedState {
				t.Errorf("Expected state %v, got %v", tt.expectedState, model.state)
			}
		})
	}
}

func TestWindowResize(t *testing.T) {
	client := api.NewClient()
	model := NewSearchModel(client, "/tmp/test")

	msg := tea.WindowSizeMsg{Width: 100, Height: 50}
	updatedModel, _ := model.Update(msg)
	model = updatedModel.(SearchModel)

	if model.width != 100 {
		t.Errorf("Expected width to be 100, got %d", model.width)
	}

	if model.height != 50 {
		t.Errorf("Expected height to be 50, got %d", model.height)
	}
}

func TestQuickFavoritesLoading(t *testing.T) {
	client := api.NewClient()
	model := NewSearchModel(client, "/tmp/test")

	stations := []api.Station{
		{StationUUID: "fav-1", Name: "Favorite 1"},
		{StationUUID: "fav-2", Name: "Favorite 2"},
	}

	msg := quickFavoritesLoadedMsg{stations: stations}
	updatedModel, _ := model.Update(msg)
	model = updatedModel.(SearchModel)

	if len(model.quickFavorites) != 2 {
		t.Errorf("Expected 2 quick favorites, got %d", len(model.quickFavorites))
	}
}

func TestRenderStationDetails(t *testing.T) {
	station := api.Station{
		StationUUID: "test-1",
		Name:        "  Test Station  ", // With whitespace
		Tags:        "jazz,smooth",
		Country:     "United States",
		State:       "California",
		Language:    "english",
		Votes:       100,
		Codec:       "MP3",
		Bitrate:     128,
	}

	details := RenderStationDetails(station)

	// Check that details contain expected information
	if !contains(details, "Test Station") {
		t.Error("Expected details to contain station name")
	}

	if !contains(details, "jazz,smooth") {
		t.Error("Expected details to contain tags")
	}

	if !contains(details, "United States") {
		t.Error("Expected details to contain country")
	}

	if !contains(details, "California") {
		t.Error("Expected details to contain state")
	}

	if !contains(details, "english") {
		t.Error("Expected details to contain language")
	}

	if !contains(details, "100") {
		t.Error("Expected details to contain votes")
	}

	if !contains(details, "MP3") {
		t.Error("Expected details to contain codec")
	}

	if !contains(details, "128") {
		t.Error("Expected details to contain bitrate")
	}
}
