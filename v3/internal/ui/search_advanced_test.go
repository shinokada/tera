package ui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/shinokada/tera/v3/internal/api"
	"github.com/shinokada/tera/v3/internal/blocklist"
)

func TestBuildAdvancedSearchParams(t *testing.T) {
	// Setup helper to create a model with inputs filled
	createModelWithInputs := func(tag, lang, country, state, name string, sortByVotes bool) SearchModel {
		model := NewSearchModel(api.NewClient(), "", blocklist.NewManager("/tmp/blocklist"))

		// Fill inputs (order: tag, lang, country, state, name)
		model.advancedInputs[0].SetValue(tag)
		model.advancedInputs[1].SetValue(lang)
		model.advancedInputs[2].SetValue(country)
		model.advancedInputs[3].SetValue(state)
		model.advancedInputs[4].SetValue(name)

		model.advancedSortByVotes = sortByVotes

		return model
	}

	tests := []struct {
		name           string
		inputs         []string // tag, lang, country, state, name
		sortByVotes    bool
		expectedParams api.SearchParams
	}{
		{
			name:        "Country Code Logic (2 letters)",
			inputs:      []string{"", "", "jp", "", ""},
			sortByVotes: true, // Match model default
			expectedParams: api.SearchParams{
				CountryCode: "JP",
				Country:     "",
				Limit:       100,
				HideBroken:  true,
				Order:       "votes",
				Reverse:     true,
			},
		},
		{
			name:        "Country Name Logic (>2 letters)",
			inputs:      []string{"", "", "japan", "", ""},
			sortByVotes: true, // Match model default
			expectedParams: api.SearchParams{
				CountryCode: "",
				Country:     "Japan", // Title cased
				Limit:       100,
				HideBroken:  true,
				Order:       "votes",
				Reverse:     true,
			},
		},
		{
			name:        "Language Lowercase Logic",
			inputs:      []string{"", "English", "", "", ""},
			sortByVotes: true, // Match model default
			expectedParams: api.SearchParams{
				Language:   "english",
				Limit:      100,
				HideBroken: true,
				Order:      "votes",
				Reverse:    true,
			},
		},
		{
			name:        "Mixed Logic",
			inputs:      []string{"Jazz", "English", "usa", "California", "Test"},
			sortByVotes: true, // Match model default
			expectedParams: api.SearchParams{
				Tag:         "Jazz",
				Language:    "english",
				Country:     "Usa", // Title cased because 3 letters
				CountryCode: "",
				State:       "California",
				Name:        "Test",
				Limit:       100,
				HideBroken:  true,
				Order:       "votes",
				Reverse:     true,
			},
		},
		{
			name:        "Mixed Logic with Country Code",
			inputs:      []string{"", "", "us", "", ""},
			sortByVotes: true, // Match model default
			expectedParams: api.SearchParams{
				CountryCode: "US", // Uppercased because 2 letters
				Country:     "",
				Limit:       100,
				HideBroken:  true,
				Order:       "votes",
				Reverse:     true,
			},
		},
		{
			name:        "Sort by Votes (explicit true)",
			inputs:      []string{"", "", "", "", ""},
			sortByVotes: true,
			expectedParams: api.SearchParams{
				Limit:      100,
				HideBroken: true,
				Order:      "votes",
				Reverse:    true,
			},
		},
		{
			name:        "Sort by Relevance (explicit false)",
			inputs:      []string{"Jazz", "", "", "", ""},
			sortByVotes: false,
			expectedParams: api.SearchParams{
				Tag:        "Jazz",
				Limit:      100,
				HideBroken: true,
				Order:      "",
				Reverse:    false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// sortByVotes defaults to false in Go, but the helper defaults to true
			// So we need to be explicit about which tests set it
			model := createModelWithInputs(tt.inputs[0], tt.inputs[1], tt.inputs[2], tt.inputs[3], tt.inputs[4], tt.sortByVotes)
			params := model.buildAdvancedSearchParams()

			if params.Tag != tt.expectedParams.Tag {
				t.Errorf("Tag: expected %q, got %q", tt.expectedParams.Tag, params.Tag)
			}
			if params.Language != tt.expectedParams.Language {
				t.Errorf("Language: expected %q, got %q", tt.expectedParams.Language, params.Language)
			}
			if params.Country != tt.expectedParams.Country {
				t.Errorf("Country: expected %q, got %q", tt.expectedParams.Country, params.Country)
			}
			if params.CountryCode != tt.expectedParams.CountryCode {
				t.Errorf("CountryCode: expected %q, got %q", tt.expectedParams.CountryCode, params.CountryCode)
			}
			if params.State != tt.expectedParams.State {
				t.Errorf("State: expected %q, got %q", tt.expectedParams.State, params.State)
			}
			if params.Name != tt.expectedParams.Name {
				t.Errorf("Name: expected %q, got %q", tt.expectedParams.Name, params.Name)
			}
			if params.Order != tt.expectedParams.Order {
				t.Errorf("Order: expected %q, got %q", tt.expectedParams.Order, params.Order)
			}
			if params.Reverse != tt.expectedParams.Reverse {
				t.Errorf("Reverse: expected %v, got %v", tt.expectedParams.Reverse, params.Reverse)
			}
			if params.Limit != tt.expectedParams.Limit {
				t.Errorf("Limit: expected %d, got %d", tt.expectedParams.Limit, params.Limit)
			}
			if params.HideBroken != tt.expectedParams.HideBroken {
				t.Errorf("HideBroken: expected %v, got %v", tt.expectedParams.HideBroken, params.HideBroken)
			}
		})
	}
}

func TestBitrateToggle(t *testing.T) {
	model := NewSearchModel(api.NewClient(), "", blocklist.NewManager("/tmp/blocklist"))
	model.state = searchStateAdvancedForm
	model.advancedFocusIdx = 6 // Focus on bitrate
	model.advancedBitrate = "1"

	// Press '1' again should toggle off
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("1")}
	updatedModel, _ := model.Update(msg)
	searchModel := updatedModel.(SearchModel)

	if searchModel.advancedBitrate != "" {
		t.Errorf("Expected bitrate to be empty after toggling, got %q", searchModel.advancedBitrate)
	}

	// Press '2' should select 2
	msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("2")}
	updatedModel, _ = searchModel.Update(msg)
	searchModel = updatedModel.(SearchModel)

	if searchModel.advancedBitrate != "2" {
		t.Errorf("Expected bitrate to be '2', got %q", searchModel.advancedBitrate)
	}
}
