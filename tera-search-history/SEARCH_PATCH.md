# Patch for /internal/ui/search.go

## Add these functions at the end of the file (before the closing brace)

```go
// executeHistorySearch executes a search from history
func (m SearchModel) executeHistorySearch(searchType, query string) (tea.Model, tea.Cmd) {
	// Map string search type to api.SearchType
	switch searchType {
	case "tag":
		m.searchType = api.SearchByTag
	case "name":
		m.searchType = api.SearchByName
	case "language":
		m.searchType = api.SearchByLanguage
	case "country":
		m.searchType = api.SearchByCountry
	case "state":
		m.searchType = api.SearchByState
	case "advanced":
		m.searchType = api.SearchAdvanced
	default:
		// Unknown type, go back to menu
		return m, nil
	}

	// Execute search immediately
	m.state = searchStateLoading
	return m, m.performSearch(query)
}

// renderSearchMenu renders the search menu with history
func (m SearchModel) renderSearchMenu() string {
	var content strings.Builder

	// Show main menu
	content.WriteString(m.menuList.View())

	// Add history section if there are items
	if m.searchHistory != nil && len(m.searchHistory.SearchItems) > 0 {
		content.WriteString("\n\n")
		content.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("─── Recent Searches ───"))
		content.WriteString("\n")

		// Show up to MaxSize history items
		for i, item := range m.searchHistory.SearchItems {
			if i >= m.searchHistory.MaxSize {
				break
			}

			// Format: "10. tag: jazz"
			itemNum := i + 10
			prefix := fmt.Sprintf("%2d. ", itemNum)
			typeLabel := fmt.Sprintf("%s: ", item.SearchType)

			dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
			line := prefix + dimStyle.Render(typeLabel) + item.Query
			content.WriteString(line)
			content.WriteString("\n")
		}
	}

	// Error message if any
	if m.err != nil {
		content.WriteString("\n")
		content.WriteString(errorStyle().Render(fmt.Sprintf("Error: %v", m.err)))
	}

	helpText := "↑↓/jk: Navigate • Enter: Select • 1-6: Search Type"
	if m.searchHistory != nil && len(m.searchHistory.SearchItems) > 0 {
		helpText += " • 10+: Quick Search"
	}
	helpText += " • Esc: Back • Ctrl+C: Quit"

	return RenderPageWithBottomHelp(PageLayout{
		Content: content.String(),
		Help:    helpText,
	}, m.height)
}
```

## Modify handleMenuInput() function

Find the `handleMenuInput()` function and add this code right after the line:
```go
if msg.String() == "esc" || msg.String() == "m" {
```

Add this block BEFORE the "Handle menu navigation and selection" comment:

```go
	// Handle quick select for history items (10+)
	if len(msg.String()) >= 2 {
		// Try to parse as a number for history quick select
		var histIndex int
		if _, err := fmt.Sscanf(msg.String(), "%d", &histIndex); err == nil && histIndex >= 10 {
			// Calculate actual history index (10 = index 0, 11 = index 1, etc.)
			actualIndex := histIndex - 10
			if m.searchHistory != nil && actualIndex < len(m.searchHistory.SearchItems) {
				item := m.searchHistory.SearchItems[actualIndex]
				// Set search type based on history item and execute search
				return m.executeHistorySearch(item.SearchType, item.Query)
			}
		}
	}
```

## Modify performSearch() function

At the very beginning of the `performSearch()` function (after `return func() tea.Msg {`), add:

```go
		// Save search to history in background
		go func() {
			store := storage.NewStorage(m.favoritePath)
			var searchTypeStr string
			switch m.searchType {
			case api.SearchByTag:
				searchTypeStr = "tag"
			case api.SearchByName:
				searchTypeStr = "name"
			case api.SearchByLanguage:
				searchTypeStr = "language"
			case api.SearchByCountry:
				searchTypeStr = "country"
			case api.SearchByState:
				searchTypeStr = "state"
			case api.SearchAdvanced:
				searchTypeStr = "advanced"
			}
			_ = store.AddSearchItem(context.Background(), searchTypeStr, query)
		}()
```

## Modify View() function

In the `View()` function, find the case `searchStateMenu:` and replace it with:

```go
	case searchStateMenu:
		return m.renderSearchMenu()
```

This replaces the entire block that was:
```go
	case searchStateMenu:
		var content strings.Builder
		content.WriteString(m.menuList.View())
		if m.err != nil {
			content.WriteString("\n\n")
			content.WriteString(errorStyle().Render(fmt.Sprintf("Error: %v", m.err)))
		}
		return RenderPageWithBottomHelp(PageLayout{
			Content: content.String(),
			Help:    "↑↓/jk: Navigate • Enter: Select • 1-6: Quick select • Esc: Back • Ctrl+C: Quit",
		}, m.height)
```
