package timer

import (
	"sync"
	"time"

	"github.com/shinokada/tera/v3/internal/api"
)

// SessionEntry records one station played during a sleep session.
type SessionEntry struct {
	Station   api.Station
	StartedAt time.Time
	Duration  time.Duration // populated when the next station starts or session ends
}

// SleepSession accumulates stations played since the timer was set.
// It is thread-safe.
type SleepSession struct {
	mu        sync.Mutex
	entries   []SessionEntry
	startedAt time.Time
}

// NewSleepSession creates a fresh session starting now.
func NewSleepSession() *SleepSession {
	return &SleepSession{startedAt: time.Now()}
}

// RecordStation is called each time a new station starts playing.
// It closes out the previous entry's duration and opens a new one.
func (s *SleepSession) RecordStation(station api.Station) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()

	// Close the previous entry
	if len(s.entries) > 0 {
		last := &s.entries[len(s.entries)-1]
		if last.Duration == 0 {
			last.Duration = now.Sub(last.StartedAt)
		}
	}

	s.entries = append(s.entries, SessionEntry{
		Station:   station,
		StartedAt: now,
	})
}

// RecordStop closes the final entry when playback stops.
func (s *SleepSession) RecordStop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	if len(s.entries) > 0 {
		last := &s.entries[len(s.entries)-1]
		if last.Duration == 0 {
			last.Duration = now.Sub(last.StartedAt)
		}
	}
}

// Entries returns a snapshot of all session entries.
func (s *SleepSession) Entries() []SessionEntry {
	s.mu.Lock()
	defer s.mu.Unlock()

	cp := make([]SessionEntry, len(s.entries))
	copy(cp, s.entries)
	return cp
}

// Total returns the wall-clock duration of the session.
func (s *SleepSession) Total() time.Duration {
	s.mu.Lock()
	defer s.mu.Unlock()
	return time.Since(s.startedAt)
}


