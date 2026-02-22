package timer

import (
	"testing"
	"time"

	"github.com/shinokada/tera/v3/internal/api"
)

func makeStation(name string) api.Station {
	return api.Station{Name: name, StationUUID: name}
}

// TestSessionRecordsStations verifies entries are appended in order.
func TestSessionRecordsStations(t *testing.T) {
	s := NewSleepSession()
	s.RecordStation(makeStation("Jazz FM"))
	s.RecordStation(makeStation("Rock 101"))

	entries := s.Entries()
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
	if entries[0].Station.Name != "Jazz FM" {
		t.Errorf("first entry: want Jazz FM, got %s", entries[0].Station.Name)
	}
	if entries[1].Station.Name != "Rock 101" {
		t.Errorf("second entry: want Rock 101, got %s", entries[1].Station.Name)
	}
}

// TestSessionDuration verifies that switching stations closes the previous duration.
func TestSessionDuration(t *testing.T) {
	s := NewSleepSession()
	s.RecordStation(makeStation("A"))
	time.Sleep(50 * time.Millisecond)
	s.RecordStation(makeStation("B")) // closes A's duration

	entries := s.Entries()
	if entries[0].Duration <= 0 {
		t.Errorf("first entry duration should be > 0, got %v", entries[0].Duration)
	}
}

// TestSessionSingleStation verifies RecordStop closes the only entry.
func TestSessionSingleStation(t *testing.T) {
	s := NewSleepSession()
	s.RecordStation(makeStation("Solo"))
	time.Sleep(30 * time.Millisecond)
	s.RecordStop()

	entries := s.Entries()
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].Duration <= 0 {
		t.Errorf("duration should be > 0 after RecordStop, got %v", entries[0].Duration)
	}
}

// TestSessionEmptySession verifies an empty session is safe to query.
func TestSessionEmptySession(t *testing.T) {
	s := NewSleepSession()
	s.RecordStop() // must not panic

	entries := s.Entries()
	if len(entries) != 0 {
		t.Errorf("expected 0 entries, got %d", len(entries))
	}

	total := s.Total()
	if total < 0 {
		t.Errorf("total should be >= 0, got %v", total)
	}
}
