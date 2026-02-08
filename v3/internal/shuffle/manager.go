package shuffle

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/shinokada/tera/v3/internal/api"
	"github.com/shinokada/tera/v3/internal/storage"
)

// Manager handles shuffle mode logic
type Manager struct {
	config          storage.ShuffleConfig
	stations        []api.Station // All available stations from search
	shuffledIndices []int         // Shuffled order of indices
	currentIndex    int           // Current position in shuffled list
	history         []api.Station // History of played stations
	keyword         string        // Original search keyword
	sessionCount    int           // Number of stations played in this session
	ticker          *time.Ticker  // Auto-advance ticker
	timerPaused     bool          // Whether timer is paused
	timeRemaining   time.Duration // Time remaining on current timer
	lastTickTime    time.Time     // Last time we updated the timer
}

// NewManager creates a new shuffle manager
func NewManager(config storage.ShuffleConfig) *Manager {
	return &Manager{
		config:       config,
		history:      make([]api.Station, 0),
		timerPaused:  false,
		sessionCount: 0,
	}
}

// Initialize sets up shuffle with a new set of stations
func (m *Manager) Initialize(keyword string, stations []api.Station) error {
	if len(stations) == 0 {
		return fmt.Errorf("no stations provided for shuffle")
	}

	m.keyword = keyword
	m.stations = stations
	m.currentIndex = 0
	m.sessionCount = 0
	m.history = make([]api.Station, 0)

	// Create shuffled indices
	m.shuffledIndices = rand.Perm(len(stations))

	// Initialize timer if auto-advance is enabled
	if m.config.AutoAdvance {
		m.startTimer()
	}

	return nil
}

// GetCurrentStation returns the current station
func (m *Manager) GetCurrentStation() (*api.Station, error) {
	if len(m.stations) == 0 || m.currentIndex >= len(m.shuffledIndices) {
		return nil, fmt.Errorf("no current station")
	}

	stationIdx := m.shuffledIndices[m.currentIndex]
	station := m.stations[stationIdx]
	return &station, nil
}

// Next advances to the next station in shuffle.
// filter is an optional function to skip stations (returns true to keep, false to skip).
func (m *Manager) Next(filter func(api.Station) bool) (*api.Station, error) {
	if len(m.stations) == 0 {
		return nil, fmt.Errorf("no stations available")
	}

	// Save current station to history if we have one
	if m.currentIndex < len(m.shuffledIndices) {
		currentStation, _ := m.GetCurrentStation()
		if currentStation != nil {
			m.addToHistory(*currentStation)
		}
	}

	// Try to find the next valid station
	stationsTried := 0
	for stationsTried < len(m.stations) {
		// Move to next station
		m.currentIndex++
		// If we've exhausted all stations, reshuffle
		if m.currentIndex >= len(m.shuffledIndices) {
			m.shuffledIndices = rand.Perm(len(m.stations))
			m.currentIndex = 0
		}

		station, err := m.GetCurrentStation()
		if err != nil {
			return nil, err
		}

		// If no filter or filter accepts it, we're done
		if filter == nil || filter(*station) {
			m.sessionCount++
			// Restart timer if auto-advance is enabled
			if m.config.AutoAdvance {
				m.restartTimer()
			}
			return station, nil
		}
		stationsTried++
	}

	return nil, fmt.Errorf("no unblocked stations available")
}

// Previous goes back to the previous station in history
func (m *Manager) Previous() (*api.Station, error) {
	if !m.config.RememberHistory || len(m.history) == 0 {
		return nil, fmt.Errorf("no history available")
	}

	// Pop last station from history
	station := m.history[len(m.history)-1]
	m.history = m.history[:len(m.history)-1]

	// Pause timer when going back in history
	m.PauseTimer()

	return &station, nil
}

// addToHistory adds a station to history
func (m *Manager) addToHistory(station api.Station) {
	if !m.config.RememberHistory {
		return
	}

	m.history = append(m.history, station)

	// Limit history size
	if len(m.history) > m.config.MaxHistory {
		m.history = m.history[len(m.history)-m.config.MaxHistory:]
	}
}

// GetHistory returns the current history
func (m *Manager) GetHistory() []api.Station {
	historyCopy := make([]api.Station, len(m.history))
	copy(historyCopy, m.history)
	return historyCopy
}

// GetSessionCount returns the number of stations played in this session
func (m *Manager) GetSessionCount() int {
	return m.sessionCount
}

// GetKeyword returns the search keyword
func (m *Manager) GetKeyword() string {
	return m.keyword
}

// startTimer starts the auto-advance timer
func (m *Manager) startTimer() {
	if m.ticker != nil {
		m.ticker.Stop()
	}

	interval := time.Duration(m.config.IntervalMinutes) * time.Minute
	m.timeRemaining = interval
	m.lastTickTime = time.Now()
	m.ticker = time.NewTicker(time.Second)
	m.timerPaused = false
}

// restartTimer restarts the auto-advance timer
func (m *Manager) restartTimer() {
	m.startTimer()
}

// StopTimer stops the auto-advance timer
func (m *Manager) StopTimer() {
	if m.ticker != nil {
		m.ticker.Stop()
		m.ticker = nil
	}
	m.timerPaused = false
}

// PauseTimer pauses the auto-advance timer
func (m *Manager) PauseTimer() {
	if m.ticker != nil && !m.timerPaused {
		m.timerPaused = true
	}
}

// ResumeTimer resumes the auto-advance timer
func (m *Manager) ResumeTimer() {
	if m.ticker != nil && m.timerPaused {
		m.timerPaused = false
		m.lastTickTime = time.Now()
	}
}

// ToggleTimer toggles the timer pause state
func (m *Manager) ToggleTimer() bool {
	if m.timerPaused {
		m.ResumeTimer()
		return false // Not paused
	} else {
		m.PauseTimer()
		return true // Paused
	}
}

// GetTimerTick returns the timer's tick channel (or nil if no timer)
func (m *Manager) GetTimerTick() <-chan time.Time {
	if m.ticker == nil {
		return nil
	}
	return m.ticker.C
}

// UpdateTimer updates the remaining time (call on each tick)
func (m *Manager) UpdateTimer() bool {
	if m.ticker == nil || m.timerPaused {
		return false
	}

	now := time.Now()
	elapsed := now.Sub(m.lastTickTime)
	m.lastTickTime = now

	m.timeRemaining -= elapsed

	// Check if timer expired
	if m.timeRemaining <= 0 {
		return true // Signal to advance
	}

	return false
}

// GetTimeRemaining returns the remaining time as a string
func (m *Manager) GetTimeRemaining() string {
	if m.ticker == nil {
		return ""
	}

	if m.timerPaused {
		return "Paused"
	}

	seconds := int(m.timeRemaining.Seconds())
	if seconds < 0 {
		seconds = 0
	}

	minutes := seconds / 60
	secs := seconds % 60

	return fmt.Sprintf("%d:%02d", minutes, secs)
}

// IsTimerPaused returns whether the timer is paused
func (m *Manager) IsTimerPaused() bool {
	return m.timerPaused
}

// IsAutoAdvanceEnabled returns whether auto-advance is enabled
func (m *Manager) IsAutoAdvanceEnabled() bool {
	return m.config.AutoAdvance
}

// UpdateConfig updates the shuffle configuration
func (m *Manager) UpdateConfig(config storage.ShuffleConfig) {
	oldAutoAdvance := m.config.AutoAdvance
	oldInterval := m.config.IntervalMinutes
	oldRememberHistory := m.config.RememberHistory
	m.config = config

	// If auto-advance was toggled on, start timer
	if config.AutoAdvance && !oldAutoAdvance {
		m.startTimer()
	}

	// If auto-advance was toggled off, stop timer
	if !config.AutoAdvance && oldAutoAdvance {
		m.StopTimer()
	}

	// If interval changed while auto-advance stays enabled, restart timer
	if config.AutoAdvance && oldAutoAdvance && config.IntervalMinutes != oldInterval {
		m.startTimer()
	}

	// If history was disabled, clear existing history
	if oldRememberHistory && !config.RememberHistory {
		m.history = nil
	}

	// Trim history if max size decreased
	if config.RememberHistory && len(m.history) > config.MaxHistory {
		m.history = m.history[len(m.history)-config.MaxHistory:]
	}
}

// GetConfig returns the current configuration
func (m *Manager) GetConfig() storage.ShuffleConfig {
	return m.config
}

// Stop stops the shuffle session and cleans up
func (m *Manager) Stop() {
	m.Cleanup()
}

// TogglePause toggles the timer pause state and returns the new paused state
func (m *Manager) TogglePause() bool {
	return m.ToggleTimer()
}

// ShuffleStatus represents the current state of shuffle mode
type ShuffleStatus struct {
	Keyword       string
	CurrentIndex  int
	SessionCount  int
	History       []api.Station
	TimeRemaining time.Duration
	TimerPaused   bool
	AutoAdvance   bool
}

// GetStatus returns the current shuffle status
func (m *Manager) GetStatus() ShuffleStatus {
	historyCopy := make([]api.Station, len(m.history))
	copy(historyCopy, m.history)
	return ShuffleStatus{
		Keyword:       m.keyword,
		CurrentIndex:  m.currentIndex,
		SessionCount:  m.sessionCount,
		History:       historyCopy,
		TimeRemaining: m.timeRemaining,
		TimerPaused:   m.timerPaused,
		AutoAdvance:   m.config.AutoAdvance,
	}
}

// Cleanup stops timers and cleans up resources
func (m *Manager) Cleanup() {
	m.StopTimer()
}
