# Patch for /internal/ui/lucky.go

## 1. Add searchHistory field to LuckyModel struct

Find the `LuckyModel` struct definition and add this field:

```go
type LuckyModel struct {
	state           luckyState
	apiClient       *api.Client
	textInput       textinput.Model
	newListInput    textinput.Model
	selectedStation *api.Station
	player          *player.MPVPlayer
	favoritePath    string
	searchHistory   *storage.SearchHistoryStore // ADD THIS LINE
	saveMessage     string
	saveMessageTime int
	width           int
	height          int
	err             error
	availableLists  []string
	listItems       []list.Item
	listModel       list.Model
	helpModel       components.HelpModel
}
```

## 2. Load search history in NewLuckyModel()

In the `NewLuckyModel()` function, add this code BEFORE the `return LuckyModel{` line:

```go
	// Load search history (if it fails, just use empty history)
	store := storage.NewStorage(favoritePath)
	history, err := store.LoadSearchHistory(context.Background())
	if err != nil || history == nil {
		history = storage.NewSearchHistoryStore()
	}
```

Then add `searchHistory: history,` to the returned struct:

```go
	return LuckyModel{
		state:         luckyStateInput,
		apiClient:     apiClient,
		textInput:     ti,
		newListInput:  nli,
		favoritePath:  favoritePath,
		player:        player.NewMPVPlayer(),
		searchHistory: history, // ADD THIS LINE
		width:         80,
		height:        24,
		helpModel:     components.NewHelpModel(components.CreatePlayingHelp()),
	}
```

## 3. Update updateInput() to handle number selection

In the `updateInput()` function, add this code at the very beginning (before the switch statement):

```go
func (m LuckyModel) updateInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Handle quick select for history items (1-10)
	if len(msg.String()) >= 1 && len(msg.String()) <= 2 {
		var histIndex int
		if _, err := fmt.Sscanf(msg.String(), "%d", &histIndex); err == nil {
			actualIndex := histIndex - 1 // 1 = index 0, 2 = index 1, etc.
			if m.searchHistory != nil && actualIndex >= 0 && actualIndex < len(m.searchHistory.LuckyQueries) {
				query := m.searchHistory.LuckyQueries[actualIndex]
				m.state = luckyStateSearching
				m.err = nil
				return m, m.searchAndPickRandom(query)
			}
		}
	}

	// Rest of existing switch...
	switch msg.String() {
	// ... existing code ...
	}
}
```

## 4. Update searchAndPickRandom() to save history

In the `searchAndPickRandom()` function, add this at the very beginning (after `return func() tea.Msg {`):

```go
func (m LuckyModel) searchAndPickRandom(keyword string) tea.Cmd {
	return func() tea.Msg {
		// Save to history in background
		go func() {
			store := storage.NewStorage(m.favoritePath)
			_ = store.AddLuckyQuery(context.Background(), keyword)
		}()

		// Rest of existing code...
		// Search by tag (genre/keyword)
		stations, err := m.apiClient.SearchByTag(context.Background(), keyword)
		// ... etc
	}
}
```

## 5. Replace viewInput() function entirely

Replace the entire `viewInput()` function with this new version:

```go
// viewInput renders the input view
func (m LuckyModel) viewInput() string {
	var content strings.Builder

	// Instructions
	content.WriteString("Type a genre of music: rock, classical, jazz, pop, country, hip, heavy, blues, soul.\n")
	content.WriteString("Or type a keyword like: meditation, relax, mozart, Beatles, etc.\n\n")
	content.WriteString(infoStyle().Render("Use only one word."))
	content.WriteString("\n\n")

	// Input field
	content.WriteString("Genre/keyword: ")
	content.WriteString(m.textInput.View())

	// Show history if available
	if m.searchHistory != nil && len(m.searchHistory.LuckyQueries) > 0 {
		content.WriteString("\n\n")
		dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
		content.WriteString(dimStyle.Render("─── Recent Searches ───"))
		content.WriteString("\n")

		for i, query := range m.searchHistory.LuckyQueries {
			if i >= m.searchHistory.MaxSize {
				break
			}
			line := fmt.Sprintf("%2d. %s", i+1, query)
			content.WriteString(line)
			content.WriteString("\n")
		}
	}

	// Error message if any
	if m.err != nil {
		content.WriteString("\n")
		content.WriteString(errorStyle().Render(m.err.Error()))
	}

	helpText := "Enter: Search"
	if m.searchHistory != nil && len(m.searchHistory.LuckyQueries) > 0 {
		helpText += " • 1-" + fmt.Sprintf("%d", min(len(m.searchHistory.LuckyQueries), m.searchHistory.MaxSize)) + ": Quick search"
	}
	helpText += " • Esc: Back • Ctrl+C: Quit"

	return RenderPageWithBottomHelp(PageLayout{
		Title:   "I Feel Lucky",
		Content: content.String(),
		Help:    helpText,
	}, m.height)
}

// min returns the minimum of two integers (helper for Go < 1.21)
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
```
