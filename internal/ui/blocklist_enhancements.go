package ui

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/shinokada/tera/internal/blocklist"
)

// ruleListItem wraps a BlockRule for list.Item interface
type ruleListItem struct {
	rule  blocklist.BlockRule
	index int
}

func (r ruleListItem) Title() string {
	return r.rule.String()
}

func (r ruleListItem) Description() string {
	typeStr := ""
	switch r.rule.Type {
	case blocklist.BlockRuleCountry:
		typeStr = "Blocks all stations from this country"
	case blocklist.BlockRuleLanguage:
		typeStr = "Blocks all stations in this language"
	case blocklist.BlockRuleTag:
		typeStr = "Blocks all stations with this tag/genre"
	}
	return typeStr
}

func (r ruleListItem) FilterValue() string {
	return r.rule.String()
}

// Enhanced message types
type blockRulesLoadedMsg struct {
	rules []blocklist.BlockRule
}

type blockRuleDeletedMsg struct {
	rule blocklist.BlockRule
}

type blocklistExportedMsg struct {
	path string
}

type blocklistImportedMsg struct {
	rulesCount    int
	stationsCount int
}

// loadBlockRules loads all block rules into the list
func (m *BlocklistModel) loadBlockRules() tea.Cmd {
	return func() tea.Msg {
		rules := m.manager.GetBlockRules()
		return blockRulesLoadedMsg{rules}
	}
}

// deleteBlockRule removes a block rule
func (m *BlocklistModel) deleteBlockRule(rule blocklist.BlockRule) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		if err := m.manager.RemoveBlockRule(ctx, rule.Type, rule.Value); err != nil {
			return blockRuleErrorMsg{err: err}
		}
		return blockRuleDeletedMsg{rule: rule}
	}
}

// addBlockRuleWithConfirmation initiates the confirmation flow
func (m *BlocklistModel) addBlockRuleWithConfirmation(ruleType blocklist.BlockRuleType, value string) (BlocklistModel, tea.Cmd) {
	m.pendingRuleType = ruleType
	m.pendingRuleValue = value
	m.previousState = m.state
	m.state = blocklistConfirmAddRule
	m.textInput.Blur()
	return *m, nil
}

// confirmAddBlockRule actually adds the rule after confirmation
func (m *BlocklistModel) confirmAddBlockRule() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		if err := m.manager.AddBlockRule(ctx, m.pendingRuleType, m.pendingRuleValue); err != nil {
			return blockRuleErrorMsg{err: err}
		}
		return blockRuleAddedMsg{
			ruleType: m.pendingRuleType,
			value:    m.pendingRuleValue,
		}
	}
}

// exportBlocklist exports the blocklist to a JSON file
// TODO: Implement export/import UI flow
// nolint:unused
func (m *BlocklistModel) exportBlocklist(filename string) tea.Cmd {
	return func() tea.Msg {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return blockRuleErrorMsg{err: fmt.Errorf("failed to get home directory: %w", err)}
		}

		// Create exports directory
		exportDir := filepath.Join(homeDir, ".tera", "exports")
		if err := os.MkdirAll(exportDir, 0755); err != nil {
			return blockRuleErrorMsg{err: fmt.Errorf("failed to create export directory: %w", err)}
		}

		// Add timestamp to filename if not present
		if filename == "" {
			filename = fmt.Sprintf("blocklist-%s.json", time.Now().Format("2006-01-02-150405"))
		} else if filepath.Ext(filename) != ".json" {
			filename = filename + ".json"
		}

		exportPath := filepath.Join(exportDir, filename)

		// Get all data to export
		stations := m.manager.GetAll()
		rules := m.manager.GetBlockRules()

		// Create export structure
		data := blocklist.Blocklist{
			Version:         "1.0",
			BlockedStations: stations,
			BlockRules:      rules,
		}

		// Marshal to JSON
		jsonData, err := json.MarshalIndent(data, "", "  ")
		if err != nil {
			return blockRuleErrorMsg{err: fmt.Errorf("failed to marshal blocklist: %w", err)}
		}

		// Write to file
		if err := os.WriteFile(exportPath, jsonData, 0644); err != nil {
			return blockRuleErrorMsg{err: fmt.Errorf("failed to write export file: %w", err)}
		}

		return blocklistExportedMsg{path: exportPath}
	}
}

// importBlocklist imports a blocklist from a JSON file
// TODO: Implement export/import UI flow
// nolint:unused
func (m *BlocklistModel) importBlocklist(filepath string, merge bool) tea.Cmd {
	return func() tea.Msg {
		// Read file
		data, err := os.ReadFile(filepath)
		if err != nil {
			return blockRuleErrorMsg{err: fmt.Errorf("failed to read import file: %w", err)}
		}

		// Parse JSON
		var imported blocklist.Blocklist
		if err := json.Unmarshal(data, &imported); err != nil {
			return blockRuleErrorMsg{err: fmt.Errorf("failed to parse blocklist JSON: %w", err)}
		}

		ctx := context.Background()

		// If not merging, clear current blocklist
		if !merge {
			if err := m.manager.Clear(ctx); err != nil {
				return blockRuleErrorMsg{err: fmt.Errorf("failed to clear blocklist: %w", err)}
			}
		}

		// Import rules
		rulesCount := 0
		for _, rule := range imported.BlockRules {
			// AddBlockRule handles duplicates
			_ = m.manager.AddBlockRule(ctx, rule.Type, rule.Value)
			rulesCount++
		}

		// Import stations would require manager enhancement
		stationsCount := len(imported.BlockedStations)

		return blocklistImportedMsg{
			rulesCount:    rulesCount,
			stationsCount: stationsCount,
		}
	}
}

// createRulesListModel creates a list model for displaying rules
func createRulesListModel(rules []blocklist.BlockRule) list.Model {
	items := make([]list.Item, len(rules))
	for i, rule := range rules {
		items[i] = ruleListItem{rule: rule, index: i}
	}

	delegate := createStyledDelegate()
	l := list.New(items, delegate, 80, 20)
	l.Title = "🚫 Active Block Rules"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.SetShowHelp(false)
	l.SetShowPagination(true)
	l.Styles.Title = listTitleStyle()
	l.Styles.PaginationStyle = paginationStyle()

	return l
}
