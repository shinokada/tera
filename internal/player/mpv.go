package player

import (
	"fmt"
	"os/exec"
	"sync"

	"github.com/shinokada/tera/internal/api"
)

// MPVPlayer manages the MPV process for playing radio streams
type MPVPlayer struct {
	cmd     *exec.Cmd
	playing bool
	station *api.Station
	mu      sync.Mutex
	stopCh  chan struct{}
}

// NewMPVPlayer creates a new MPV player instance
func NewMPVPlayer() *MPVPlayer {
	return &MPVPlayer{
		playing: false,
		stopCh:  make(chan struct{}),
	}
}

// Play starts playing a radio station
func (p *MPVPlayer) Play(station *api.Station) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Stop any existing playback
	if p.playing {
		p.stopInternal()
	}

	// Check if mpv is available
	if _, err := exec.LookPath("mpv"); err != nil {
		return fmt.Errorf("mpv not found in PATH. Please install mpv: %w", err)
	}

	// Create mpv command with appropriate flags
	// --no-video: audio only
	// --no-terminal: don't take over terminal
	// --really-quiet: minimal output
	// --no-cache: no buffering for live streams
	p.cmd = exec.Command("mpv",
		"--no-video",
		"--no-terminal",
		"--really-quiet",
		"--no-cache",
		station.URLResolved,
	)

	// Start the process
	if err := p.cmd.Start(); err != nil {
		return fmt.Errorf("failed to start mpv: %w", err)
	}

	p.playing = true
	p.station = station
	p.stopCh = make(chan struct{})

	// Monitor the process in a goroutine
	go p.monitor()

	return nil
}

// Stop stops the current playback
func (p *MPVPlayer) Stop() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.playing {
		return nil
	}

	return p.stopInternal()
}

// stopInternal stops playback without locking (internal use)
func (p *MPVPlayer) stopInternal() error {
	if p.cmd != nil && p.cmd.Process != nil {
		// Send termination signal
		if err := p.cmd.Process.Kill(); err != nil {
			return fmt.Errorf("failed to stop mpv: %w", err)
		}
		// Wait for process to finish
		_ = p.cmd.Wait()
	}

	// Signal monitoring goroutine to stop
	close(p.stopCh)

	p.playing = false
	p.station = nil
	p.cmd = nil

	return nil
}

// IsPlaying returns whether the player is currently playing
func (p *MPVPlayer) IsPlaying() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.playing
}

// GetCurrentStation returns the currently playing station, or nil
func (p *MPVPlayer) GetCurrentStation() *api.Station {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.station
}

// monitor watches the mpv process and updates state when it exits
func (p *MPVPlayer) monitor() {
	if p.cmd == nil {
		return
	}

	// Wait for either the process to end or stop signal
	done := make(chan error, 1)
	go func() {
		done <- p.cmd.Wait()
	}()

	select {
	case <-done:
		// Process ended (could be error or natural end)
		p.mu.Lock()
		p.playing = false
		p.station = nil
		p.cmd = nil
		p.mu.Unlock()
	case <-p.stopCh:
		// Stop was called, process already killed
		return
	}
}
