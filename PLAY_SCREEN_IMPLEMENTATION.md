# Play Screen Implementation Guide

## Changes needed in `internal/ui/play.go`

### 1. Update the Update() method to handle tick messages and help model size

Add this at the beginning of the `Update()` method, right after the dimension checks:

```go
func (m PlayModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Check if we need to initialize models with dimensions we now have
	if m.listsNeedInit && m.width > 0 && m.height > 0 {
		m.initializeListModel()
		m.listsNeedInit = false
	}
	if m.stationsNeedInit && m.width > 0 && m.height > 0 {
		m.initializeStationListModel()
		m.stationsNeedInit = false
	}

	switch msg := msg.(type) {
	case tickMsg:
		// Handle volume display countdown
		if m.volumeDisplayFrames > 0 {
			m.volumeDisplayFrames--
			if m.volumeDisplayFrames == 0 {
				m.volumeDisplay = ""
			}
			// Decrement save message time as well
			if m.saveMessageTime > 0 {
				m.saveMessageTime--
				if m.saveMessageTime == 0 {
					m.saveMessage = ""
				}
			}
			return m, tickEverySecond()
		}
		
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		
		// Update help model size
		m.helpModel.SetSize(msg.Width, msg.Height)
		
		// ... rest of existing WindowSizeMsg handling ...
```

### 2. Replace the `updatePlaying` method entirely:

```go
// updatePlaying handles input during playback
func (m PlayModel) updatePlaying(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// If help is visible, let it handle the key
	if m.helpModel.IsVisible() {
		var cmd tea.Cmd
		m.helpModel, cmd = m.helpModel.Update(msg)
		return m, cmd
	}

	switch msg.String() {
	case "?":
		// Show help
		m.helpModel.Show()
		return m, nil
		
	case "/":
		// Decrease volume
		if m.player != nil && m.player.IsPlaying() {
			newVol := m.player.DecreaseVolume(5)
			m.volumeDisplay = fmt.Sprintf("Volume: %d%%", newVol)
			m.volumeDisplayFrames = 2 // Show for 2 seconds
			// Update station volume
			if m.selectedStation != nil {
				m.selectedStation.Volume = newVol
				m.saveStationVolume(m.selectedStation)
			}
			return m, tickEverySecond()
		}
		
	case "*":
		// Increase volume
		if m.player != nil && m.player.IsPlaying() {
			newVol := m.player.IncreaseVolume(5)
			m.volumeDisplay = fmt.Sprintf("Volume: %d%%", newVol)
			m.volumeDisplayFrames = 2 // Show for 2 seconds
			// Update station volume
			if m.selectedStation != nil {
				m.selectedStation.Volume = newVol
				m.saveStationVolume(m.selectedStation)
			}
			return m, tickEverySecond()
		}
		
	case "esc":
		// Stop playback and return to main menu (no save prompt)
		if err := m.player.Stop(); err != nil {
			m.err = fmt.Errorf("failed to stop playback: %w", err)
			return m, nil
		}
		m.selectedStation = nil
		return m, func() tea.Msg {
			return backToMainMsg{}
		}
		
	case "0":
		// Return to main menu (Level 3 shortcut)
		if err := m.player.Stop(); err != nil {
			m.err = fmt.Errorf("failed to stop playback: %w", err)
			return m, nil
		}
		m.selectedStation = nil
		return m, func() tea.Msg {
			return backToMainMsg{}
		}
		
	case "f":
		// Save to Quick Favorites
		return m, m.saveToQuickFavorites()
		
	case "s":
		// Save to a list (not implemented yet)
		// TODO: Implement save to custom list
		m.saveMessage = "Save to list feature coming soon"
		m.saveMessageTime = 150
		return m, nil
		
	case "v":
		// Vote for this station
		return m, m.voteForStation()
	}
	return m, nil
}
```

### 3. Remove `playStateSavePrompt` completely

1. Remove `playStateSavePrompt` from the `playState` const
2. Remove the `case playStateSavePrompt:` from Update() switch
3. Remove the `updateSavePrompt()` method
4. Remove the `viewSavePrompt()` method

### 4. Add saveStationVolume helper method:

```go
// saveStationVolume saves the updated volume for a station
func (m *PlayModel) saveStationVolume(station *api.Station) {
	if station == nil {
		return
	}

	store := storage.NewStorage(m.favoritePath)
	
	// Try to update in current list first
	if m.selectedList != "" {
		list, err := store.LoadList(context.Background(), m.selectedList)
		if err == nil {
			// Find and update the station
			for i := range list.Stations {
				if list.Stations[i].StationUUID == station.StationUUID {
					list.Stations[i].Volume = station.Volume
					break
				}
			}
			// Save the updated list
			_ = store.SaveList(context.Background(), list)
		}
	}
}
```

### 5. Update `viewPlaying` method:

Replace the current `viewPlaying()` with:

```go
// viewPlaying renders the playback view
func (m PlayModel) viewPlaying() string {
	if m.selectedStation == nil {
		return "No station selected"
	}

	var content strings.Builder

	// Station info (consistent format across all playing views)
	content.WriteString(renderStationDetails(*m.selectedStation))

	// Playback status with proper spacing
	content.WriteString("\n")
	if m.player.IsPlaying() {
		content.WriteString(successStyle().Render("▶ Playing..."))
	} else {
		content.WriteString(infoStyle().Render("⏸ Stopped"))
	}

	// Volume display (if visible)
	if m.volumeDisplay != "" {
		content.WriteString("\n\n")
		content.WriteString(highlightStyle().Render(m.volumeDisplay))
	}

	// Save message (if any)
	if m.saveMessage != "" {
		content.WriteString("\n\n")
		// Determine style based on message content
		var style lipgloss.Style
		if strings.Contains(m.saveMessage, "✓") {
			style = successStyle()
		} else if strings.Contains(m.saveMessage, "Already") {
			style = infoStyle()
		} else {
			style = errorStyle()
		}
		content.WriteString(style.Render(m.saveMessage))
	}

	// Build the page
	page := RenderPageWithBottomHelp(PageLayout{
		Title:   "🎵 Now Playing",
		Content: content.String(),
		Help:    "f: Favorites • v: Vote • 0: Main Menu • ?: Help",
	}, m.height)

	// Overlay help if visible
	if m.helpModel.IsVisible() {
		return m.helpModel.View()
	}

	return page
}
```

## Summary of Changes

✅ Added help model to PlayModel struct
✅ Added volume display fields
✅ Initialized help model in NewPlayModel
✅ Added help model size update in WindowSizeMsg
✅ Added tick message handling for volume display countdown
✅ Completely rewrote updatePlaying() to handle:
  - `?` for help
  - `/` and `*` for volume control
  - `Esc` now stops and returns to main menu (NO save prompt)
  - Existing `f`, `s`, `v` keys
✅ Added saveStationVolume() helper method
✅ Updated viewPlaying() to show volume display and help overlay
✅ Removed save prompt state and methods entirely
✅ Updated footer to match spec: `f: Favorites • v: Vote • 0: Main Menu • ?: Help`

## What This Achieves

1. ✅ Per-station volume control that saves automatically
2. ✅ Context-sensitive help overlay
3. ✅ Removed annoying "Save Station?" confirmation
4. ✅ Clean, minimal footer
5. ✅ Consistent UX with main menu

## Test Checklist

- [ ] Play a station from favorites
- [ ] Adjust volume with `/` and `*`
- [ ] Verify volume persists when you replay the same station
- [ ] Press `?` to see help overlay
- [ ] Press any key to close help
- [ ] Press `f` to save to favorites (shows message)
- [ ] Press `v` to vote (shows message)
- [ ] Press `Esc` - should stop and return to main menu
- [ ] Press `0` - should also stop and return to main menu
- [ ] Verify NO save prompt appears
