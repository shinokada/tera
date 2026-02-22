package timer

import (
	"sync/atomic"
	"testing"
	"time"
)

// TestTimerFires verifies that the callback is invoked after the duration.
func TestTimerFires(t *testing.T) {
	var fired atomic.Bool
	st := NewSleepTimer(func() { fired.Store(true) })
	st.Start(50 * time.Millisecond)

	time.Sleep(150 * time.Millisecond)
	if !fired.Load() {
		t.Error("expected timer to fire, but it did not")
	}
	if st.IsActive() {
		t.Error("timer should be inactive after firing")
	}
}

// TestTimerCancel verifies that Cancel prevents the callback from firing.
func TestTimerCancel(t *testing.T) {
	var fired atomic.Bool
	st := NewSleepTimer(func() { fired.Store(true) })
	st.Start(200 * time.Millisecond)
	st.Cancel()

	time.Sleep(300 * time.Millisecond)
	if fired.Load() {
		t.Error("callback should not fire after Cancel")
	}
	if st.IsActive() {
		t.Error("timer should be inactive after Cancel")
	}
}

// TestTimerIdempotentCancel verifies Cancel is safe to call multiple times.
func TestTimerIdempotentCancel(t *testing.T) {
	st := NewSleepTimer(nil)
	// Should not panic even when never started
	st.Cancel()
	st.Cancel()
}

// TestTimerExtend verifies that Extend postpones the expiry.
func TestTimerExtend(t *testing.T) {
	var fired atomic.Bool
	st := NewSleepTimer(func() { fired.Store(true) })
	st.Start(80 * time.Millisecond)

	// Extend before expiry
	time.Sleep(40 * time.Millisecond)
	st.Extend(80 * time.Millisecond)

	// At the original expiry the timer should NOT have fired yet
	time.Sleep(60 * time.Millisecond)
	if fired.Load() {
		t.Error("timer fired too early after Extend")
	}

	// After the extended expiry it should have fired
	time.Sleep(100 * time.Millisecond)
	if !fired.Load() {
		t.Error("timer did not fire after extended duration")
	}
}

// TestTimerExtendInactive verifies Extend is a no-op when inactive.
func TestTimerExtendInactive(t *testing.T) {
	var fired atomic.Bool
	st := NewSleepTimer(func() { fired.Store(true) })
	// Extend on an inactive timer must not start one
	st.Extend(50 * time.Millisecond)
	time.Sleep(150 * time.Millisecond)
	if fired.Load() {
		t.Error("Extend on inactive timer should not start it")
	}
}

// TestTimerRemaining verifies Remaining returns sensible values.
func TestTimerRemaining(t *testing.T) {
	st := NewSleepTimer(nil)

	// Inactive timer
	rem, active := st.Remaining()
	if active {
		t.Error("should be inactive before Start")
	}
	if rem != 0 {
		t.Errorf("remaining should be 0 when inactive, got %v", rem)
	}

	// Active timer
	st.Start(500 * time.Millisecond)
	rem, active = st.Remaining()
	if !active {
		t.Error("should be active after Start")
	}
	if rem <= 0 || rem > 500*time.Millisecond {
		t.Errorf("unexpected remaining: %v", rem)
	}

	st.Cancel()
}

// TestTimerToggle verifies that starting a new timer cancels the previous one.
func TestTimerToggle(t *testing.T) {
	var count atomic.Int32
	st := NewSleepTimer(func() { count.Add(1) })

	st.Start(300 * time.Millisecond)
	// Restart with a longer duration — first callback must NOT fire
	time.Sleep(50 * time.Millisecond)
	st.Start(300 * time.Millisecond)

	time.Sleep(200 * time.Millisecond)
	if count.Load() != 0 {
		t.Error("first timer should have been cancelled by second Start")
	}

	// Wait for second timer
	time.Sleep(200 * time.Millisecond)
	if count.Load() != 1 {
		t.Errorf("expected 1 callback, got %d", count.Load())
	}
}
