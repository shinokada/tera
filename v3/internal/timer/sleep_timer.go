package timer

import (
	"context"
	"sync"
	"time"
)

// SleepTimer manages a countdown that fires a callback when it expires.
// It is thread-safe and supports cancellation and extension.
type SleepTimer struct {
	mu        sync.Mutex
	active    bool
	expiresAt time.Time
	cancelFn  context.CancelFunc
	onExpire  func()
}

// NewSleepTimer creates a new SleepTimer.
// onExpire is called in a background goroutine when the timer fires naturally.
func NewSleepTimer(onExpire func()) *SleepTimer {
	return &SleepTimer{onExpire: onExpire}
}

// Start begins the countdown for duration d.
// Any previously running timer is cancelled first.
func (s *SleepTimer) Start(d time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Cancel any existing timer
	if s.cancelFn != nil {
		s.cancelFn()
	}

	ctx, cancel := context.WithCancel(context.Background())
	s.cancelFn = cancel
	s.active = true
	s.expiresAt = time.Now().Add(d)

	go func() {
		t := time.NewTimer(d)
		defer t.Stop()
		select {
		case <-t.C:
			s.mu.Lock()
			s.active = false
			s.mu.Unlock()
			if s.onExpire != nil {
				s.onExpire()
			}
		case <-ctx.Done():
			// Cancelled — do nothing
		}
	}()
}

// Cancel stops the timer without firing the callback.
// Safe to call even when no timer is running.
func (s *SleepTimer) Cancel() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.cancelFn != nil {
		s.cancelFn()
		s.cancelFn = nil
	}
	s.active = false
}

// Extend adds d to the current deadline. If no timer is active, it is a no-op.
func (s *SleepTimer) Extend(d time.Duration) {
	s.mu.Lock()

	if !s.active {
		s.mu.Unlock()
		return
	}

	remaining := time.Until(s.expiresAt) + d
	s.mu.Unlock()

	s.Start(remaining)
}

// Remaining returns the time left and whether a timer is currently active.
func (s *SleepTimer) Remaining() (time.Duration, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.active {
		return 0, false
	}
	rem := time.Until(s.expiresAt)
	if rem < 0 {
		rem = 0
	}
	return rem, true
}

// IsActive reports whether the timer is currently running.
func (s *SleepTimer) IsActive() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.active
}
