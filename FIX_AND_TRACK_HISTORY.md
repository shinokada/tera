# Fix Instructions

## Issue 1: Remove Duplicate File

Run this command to remove the duplicate file:

```bash
cd /Users/shinichiokada/Terminal-Tools/tera
rm internal/ui/blocklist_enhanced.go
```

Or run the provided script:

```bash
cd /Users/shinichiokada/Terminal-Tools/tera
chmod +x fix_duplicates.sh
./fix_duplicates.sh
```

Then rebuild:

```bash
make clean-all && make lint && make build
```

## Issue 2: Track History Feature - Implementation Guide

### Overview
Yes, it's absolutely possible to display the last 5 track names! MPV provides metadata through its IPC interface, including `icy-title` (the track name sent by radio stations).

### Files to Modify

1. **`internal/player/mpv.go`** - Add track history tracking
2. **`internal/ui/play.go`** - Display track history in the UI

### Step 1: Update MPV Player (internal/player/mpv.go)

Add these fields to the `MPVPlayer` struct:

```go
type MPVPlayer struct {
    // ... existing fields ...
    
    trackHistory []string      // Last 5 track names
    currentTrack string         // Current playing track
    trackMu      sync.Mutex     // Protect track history
}
```

Add method to get track metadata:

```go
// GetCurrentTrack returns the current track title from stream metadata
func (p *MPVPlayer) GetCurrentTrack() (string, error) {
    val, err := p.getProperty("media-title")
    if err != nil {
        return "", err
    }
    
    if title, ok := val.(string); ok {
        return title, nil
    }
    
    return "", nil
}

// GetTrackHistory returns the last 5 track names
func (p *MPVPlayer) GetTrackHistory() []string {
    p.trackMu.Lock()
    defer p.trackMu.Unlock()
    
    // Return a copy
    history := make([]string, len(p.trackHistory))
    copy(history, p.trackHistory)
    return history
}

// addToTrackHistory adds a new track to history
func (p *MPVPlayer) addToTrackHistory(track string) {
    p.trackMu.Lock()
    defer p.trackMu.Unlock()
    
    // Skip if same as current track
    if track == p.currentTrack {
        return
    }
    
    p.currentTrack = track
    
    // Add to history (newest first)
    p.trackHistory = append([]string{track}, p.trackHistory...)
    
    // Keep only last 5
    if len(p.trackHistory) > 5 {
        p.trackHistory = p.trackHistory[:5]
    }
}

// monitorMetadata monitors for metadata changes (track info)
func (p *MPVPlayer) monitorMetadata() {
    ticker := time.NewTicker(5 * time.Second) // Check every 5 seconds
    defer ticker.Stop()
    
    for {
        select {
        case <-ticker.C:
            p.mu.Lock()
            playing := p.playing
            p.mu.Unlock()
            
            if !playing {
                return
            }
            
            // Get current track
            track, err := p.GetCurrentTrack()
            if err == nil && track != "" {
                p.addToTrackHistory(track)
            }
            
        case <-p.stopCh:
            return
        }
    }
}
```

Update the `Play` method to start metadata monitoring:

```go
// In the Play method, after starting the monitor goroutine, add:
go p.monitorMetadata()
```

Update `stopInternal` to clear track history:

```go
// In stopInternal, before returning, add:
p.trackMu.Lock()
p.trackHistory = []string{}
p.currentTrack = ""
p.trackMu.Unlock()
```

### Step 2: Update Play UI (internal/ui/play.go)

Add field to `PlayModel`:

```go
type PlayModel struct {
    // ... existing fields ...
    
    trackHistory []string  // Last 5 tracks played
}
```

Add message type:

```go
type trackHistoryMsg struct {
    tracks []string
}
```

Add polling command:

```go
// pollTrackHistory polls for track updates every 5 seconds
func (m PlayModel) pollTrackHistory() tea.Cmd {
    return tea.Tick(5*time.Second, func(t time.Time) tea.Msg {
        if m.player == nil || !m.player.IsPlaying() {
            return nil
        }
        
        tracks := m.player.GetTrackHistory()
        return trackHistoryMsg{tracks: tracks}
    })
}
```

Update the `Update` method to handle track history:

```go
// In the Update method, add this case:
case trackHistoryMsg:
    m.trackHistory = msg.tracks
    // Continue polling if still playing
    if m.state == playStatePlaying {
        return m, m.pollTrackHistory()
    }
    return m, nil

// Also update playbackStartedMsg to start polling:
case playbackStartedMsg:
    // ... existing code ...
    return m, tea.Batch(
        ticksEverySecond(),
        m.pollTrackHistory(), // Add this
    )
```

Update `viewPlaying` to display track history:

```go
// In viewPlaying function, after the playback status, add:

// Track history
if len(m.trackHistory) > 0 {
    content.WriteString("\n\n")
    content.WriteString(subtitleStyle().Render("Recently Played:"))
    content.WriteString("\n")
    
    for i, track := range m.trackHistory {
        if i >= 5 {
            break
        }
        
        // Show newest first with indicators
        indicator := "  "
        if i == 0 {
            indicator = "🎵"
        }
        
        trackStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
        if i == 0 {
            trackStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("blue"))
        }
        
        content.WriteString(fmt.Sprintf("%s %s\n", indicator, trackStyle.Render(track)))
    }
}
```

Clear track history when stopping:

```go
// In updatePlaying, when handling "esc":
case "esc":
    // Stop playback and go back
    if err := m.player.Stop(); err != nil {
        m.err = fmt.Errorf("failed to stop playback: %w", err)
        return m, nil
    }
    m.state = playStateStationSelection
    m.selectedStation = nil
    m.trackHistory = []string{}  // Add this
    return m, nil
```

### Step 3: Test the Feature

1. Build and run:
```bash
make clean-all && make lint && make build && ./tera
```

2. Play a radio station

3. Wait for track changes (most stations update every 3-5 minutes)

4. You should see the track history update automatically

### Example Output

When playing a station, you'll see something like:

```
🎵 Now Playing
────────────────────────────────────────
Station: Jazz FM
Country: United States
Codec: MP3 128kbps

▶ Playing...

Recently Played:
🎵 Miles Davis - So What
   John Coltrane - Blue Train
   Bill Evans - Waltz for Debby
   Thelonious Monk - Round Midnight
   Chet Baker - My Funny Valentine
```

### Notes

- Track info comes from the `icy-title` metadata field in the stream
- Not all stations send this metadata
- Update frequency depends on the radio station (typically every 3-5 minutes)
- The feature automatically clears history when you stop playback
- Shows maximum of 5 tracks
- Current track is highlighted with 🎵 emoji

This implementation is efficient and non-intrusive, polling every 5 seconds only when actively playing.
