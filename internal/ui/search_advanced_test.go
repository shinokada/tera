package ui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/shinokada/tera/internal/api"
)

func TestBuildAdvancedSearchParams(t *testing.T) {
	// Setup helper to create a model with inputs filled
	createModelWithInputs := func(tag, lang, country, state, name string, sortByVotes bool) SearchModel {
		model := NewSearchModel(api.NewClient(), "")

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
			name:   "Country Code Logic (2 letters)",
			inputs: []string{"", "", "jp", "", ""},
			expectedParams: api.SearchParams{
				CountryCode: "JP",
				Country:     "",
				Limit:       100,
				HideBroken:  true,
				Order:       "", // Default false for test helper
			},
		},
		{
			name:   "Country Name Logic (>2 letters)",
			inputs: []string{"", "", "japan", "", ""},
			expectedParams: api.SearchParams{
				CountryCode: "",
				Country:     "Japan", // Title cased
				Limit:       100,
				HideBroken:  true,
				Order:       "",
			},
		},
		{
			name:   "Language Lowercase Logic",
			inputs: []string{"", "English", "", "", ""},
			expectedParams: api.SearchParams{
				Language:   "english",
				Limit:      100,
				HideBroken: true,
				Order:      "",
			},
		},
		{
			name:   "Mixed Logic",
			inputs: []string{"Jazz", "English", "usa", "California", "Test"},
			expectedParams: api.SearchParams{
				Tag:         "Jazz",
				Language:    "english",
				Country:     "Usa", // Title cased because 3 letters
				CountryCode: "",
				State:       "California",
				Name:        "Test",
				Limit:       100,
				HideBroken:  true,
				Order:       "",
			},
		},
		{
			name:   "Mixed Logic with Country Code",
			inputs: []string{"", "", "us", "", ""},
			expectedParams: api.SearchParams{
				CountryCode: "US", // Uppercased because 2 letters
				Country:     "",
				Limit:       100,
				HideBroken:  true,
				Order:       "",
			},
		},
		{
			name:        "Sort by Votes",
			inputs:      []string{"", "", "", "", ""},
			sortByVotes: true,
			expectedParams: api.SearchParams{
				Limit:      100,
				HideBroken: true,
				Order:      "votes",
				Reverse:    true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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
		})
	}
}

func TestBitrateToggle(t *testing.T) {
	model := NewSearchModel(api.NewClient(), "")
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
