package player

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"

	"github.com/shinokada/tera/internal/api"
)

// MPVPlayer manages the MPV process for playing radio streams
type MPVPlayer struct {
	cmd        *exec.Cmd
	playing    bool
	station    *api.Station
	volume     int // Current volume (0-100)
	muted      bool
	lastVolume int // Volume before mute
	mu         sync.Mutex
	stopCh     chan struct{}
	socketPath string   // IPC socket path for runtime control
	conn       net.Conn // Connection to IPC socket
}

// NewMPVPlayer creates a new MPV player instance
func NewMPVPlayer() *MPVPlayer {
	return &MPVPlayer{
		playing:    false,
		volume:     100, // Default volume
		lastVolume: 100,
		stopCh:     make(chan struct{}),
	}
}

// Play starts playing a radio station
func (p *MPVPlayer) Play(station *api.Station) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Stop any existing playback
	if p.playing {
		_ = p.stopInternal()
	}

	// Check if mpv is available
	if _, err := exec.LookPath("mpv"); err != nil {
		return fmt.Errorf("mpv not found in PATH. Please install mpv: %w", err)
	}

	// Determine volume to use: station-specific volume or current player volume
	volumeToUse := p.volume
	if station.Volume > 0 {
		volumeToUse = station.Volume
	}
	// Clamp volume to valid range (0-100) to prevent unexpectedly loud playback
	if volumeToUse < 0 {
		volumeToUse = 0
	}
	if volumeToUse > 100 {
		volumeToUse = 100
	}

	// Create unique socket path for IPC
	p.socketPath = filepath.Join(os.TempDir(), fmt.Sprintf("tera-mpv-%d.sock", os.Getpid()))

	// Remove any existing socket file
	_ = os.Remove(p.socketPath)

	// Create mpv command with appropriate flags
	// --no-video: audio only
	// --no-terminal: don't take over terminal
	// --really-quiet: minimal output
	// --no-cache: no buffering for live streams
	// --volume: set initial volume
	// --input-ipc-server: enable IPC for runtime control
	p.cmd = exec.Command("mpv",
		"--no-video",
		"--no-terminal",
		"--really-quiet",
		"--no-cache",
		fmt.Sprintf("--volume=%d", volumeToUse),
		fmt.Sprintf("--input-ipc-server=%s", p.socketPath),
		station.URLResolved,
	)

	// Update current volume
	p.volume = volumeToUse
	if volumeToUse > 0 {
		p.lastVolume = volumeToUse
	}
	p.muted = (volumeToUse == 0)

	// Start the process
	if err := p.cmd.Start(); err != nil {
		return fmt.Errorf("failed to start mpv: %w", err)
	}

	p.playing = true
	p.station = station
	p.stopCh = make(chan struct{})

	// Connect to IPC socket (with retry for socket creation delay)
	go p.connectToSocket()

	// Monitor the process in a goroutine
	go p.monitor()

	return nil
}

// connectToSocket attempts to connect to the mpv IPC socket
func (p *MPVPlayer) connectToSocket() {
	// Wait a bit for mpv to create the socket
	for i := 0; i < 10; i++ {
		time.Sleep(100 * time.Millisecond)

		p.mu.Lock()
		if !p.playing {
			p.mu.Unlock()
			return
		}
		socketPath := p.socketPath
		p.mu.Unlock()

		conn, err := net.Dial("unix", socketPath)
		if err == nil {
			p.mu.Lock()
			// Guard against stale IPC connections when Play restarts quickly
			if !p.playing || p.socketPath != socketPath || p.conn != nil {
				p.mu.Unlock()
				_ = conn.Close()
				return
			}
			p.conn = conn
			currentVol := p.volume
			muted := p.muted
			// Sync volume state to mpv after connection establishes
			if muted {
				currentVol = 0
			}
			_ = p.sendCommand([]interface{}{"set_property", "volume", float64(currentVol)})
			p.mu.Unlock()
			return
		}
	}
}

// sendCommand sends a JSON command to mpv via IPC
func (p *MPVPlayer) sendCommand(command []interface{}) error {
	if p.conn == nil {
		return fmt.Errorf("not connected to mpv")
	}

	msg := map[string]interface{}{
		"command": command,
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	// Add newline as message terminator
	data = append(data, '\n')

	// Prevent UI stalls on blocked IPC writes
	_ = p.conn.SetWriteDeadline(time.Now().Add(250 * time.Millisecond))
	defer func() { _ = p.conn.SetWriteDeadline(time.Time{}) }()

	_, err = p.conn.Write(data)
	return err
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
	// Close IPC connection
	if p.conn != nil {
		_ = p.conn.Close()
		p.conn = nil
	}

	// Remove socket file
	if p.socketPath != "" {
		_ = os.Remove(p.socketPath)
		p.socketPath = ""
	}

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

// GetVolume returns the current volume level
func (p *MPVPlayer) GetVolume() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.volume
}

// IsMuted returns whether the player is currently muted
func (p *MPVPlayer) IsMuted() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.muted
}

// SetVolume sets the volume level (0-100) with immediate effect
func (p *MPVPlayer) SetVolume(volume int) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if volume < 0 {
		volume = 0
	}
	if volume > 100 {
		volume = 100
	}
	if volume > 0 {
		p.lastVolume = volume
	}
	p.volume = volume
	p.muted = (volume == 0)

	// Send volume command to mpv via IPC
	if p.conn != nil {
		_ = p.sendCommand([]interface{}{"set_property", "volume", float64(volume)})
	}
}

// IncreaseVolume increases volume by specified amount with immediate effect
func (p *MPVPlayer) IncreaseVolume(amount int) int {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.volume += amount
	if p.volume > 100 {
		p.volume = 100
	}
	p.muted = false
	p.lastVolume = p.volume

	// Send volume command to mpv via IPC
	if p.conn != nil {
		_ = p.sendCommand([]interface{}{"set_property", "volume", float64(p.volume)})
	}

	return p.volume
}

// DecreaseVolume decreases volume by specified amount with immediate effect
func (p *MPVPlayer) DecreaseVolume(amount int) int {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.volume -= amount
	if p.volume < 0 {
		p.volume = 0
	}
	p.muted = (p.volume == 0)
	if p.volume > 0 {
		p.lastVolume = p.volume
	}

	// Send volume command to mpv via IPC
	if p.conn != nil {
		_ = p.sendCommand([]interface{}{"set_property", "volume", float64(p.volume)})
	}

	return p.volume
}

// ToggleMute toggles mute state with immediate effect
func (p *MPVPlayer) ToggleMute() (muted bool, volume int) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.muted {
		// Unmute - restore previous volume
		p.volume = p.lastVolume
		if p.volume == 0 {
			p.volume = 100
		}
		p.muted = false
	} else {
		// Mute - save current volume and set to 0
		if p.volume > 0 {
			p.lastVolume = p.volume
		}
		p.volume = 0
		p.muted = true
	}

	// Send volume command to mpv via IPC
	if p.conn != nil {
		_ = p.sendCommand([]interface{}{"set_property", "volume", float64(p.volume)})
	}

	return p.muted, p.volume
}

// monitor watches the mpv process and updates state when it exits
func (p *MPVPlayer) monitor() {
	p.mu.Lock()
	cmd := p.cmd
	p.mu.Unlock()

	if cmd == nil || cmd.Process == nil {
		return
	}

	// Wait for either the process to end or stop signal
	done := make(chan error, 1)
	go func() {
		if cmd != nil && cmd.Process != nil {
			done <- cmd.Wait()
		} else {
			done <- nil
		}
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
