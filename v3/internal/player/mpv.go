package player

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"github.com/shinokada/tera/v3/internal/api"
	"github.com/shinokada/tera/v3/internal/storage"
)

// MPVPlayer manages the MPV process for playing radio streams
type MPVPlayer struct {
	cmd             *exec.Cmd
	playing         bool
	paused          bool // Pause state
	station         *api.Station
	volume          int // Current volume (0-100)
	muted           bool
	lastVolume      int // Volume before mute
	mu              sync.Mutex
	stopCh          chan struct{}
	socketPath      string                   // IPC socket path for runtime control
	conn            net.Conn                 // Connection to IPC socket
	trackHistory    []string                 // Last 5 track names
	currentTrack    string                   // Current playing track
	trackMu         sync.Mutex               // Protect track history
	metadataManager *storage.MetadataManager // Track play statistics
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
	if station.Volume != nil {
		volumeToUse = *station.Volume
	}
	// Clamp volume to valid range (0-100) to prevent unexpectedly loud playback
	if volumeToUse < 0 {
		volumeToUse = 0
	}
	if volumeToUse > 100 {
		volumeToUse = 100
	}

	// Create unique socket path for IPC (platform-specific)
	if runtime.GOOS == "windows" {
		// Windows: Use TCP socket for IPC (more reliable than named pipes)
		p.socketPath = fmt.Sprintf("127.0.0.1:%d", 10000+os.Getpid()%50000)
	} else {
		// Unix/Linux/macOS use Unix sockets
		p.socketPath = filepath.Join(os.TempDir(), fmt.Sprintf("tera-mpv-%d.sock", os.Getpid()))
		// Remove any existing socket file (Unix only)
		_ = os.Remove(p.socketPath)
	}

	// Load connection configuration
	connConfig, err := storage.LoadConnectionConfig()
	if err != nil {
		// Fall back to defaults on error
		connConfig = storage.DefaultConnectionConfig()
	}

	// Build mpv arguments
	args := []string{
		"--no-video",
		"--no-terminal",
		"--really-quiet",
		fmt.Sprintf("--volume=%d", volumeToUse),
		fmt.Sprintf("--input-ipc-server=%s", p.socketPath),
	}

	// Add connection-related flags based on config
	if connConfig.AutoReconnect {
		// Enable force loop to retry after stream drops
		args = append(args, "--loop-playlist=force")

		// FFmpeg reconnect flags for network-level reconnection
		args = append(args,
			fmt.Sprintf("--stream-lavf-o=reconnect_streamed=1,reconnect_delay_max=%d", connConfig.ReconnectDelay),
		)
	}

	// Add caching/buffering based on config
	if connConfig.StreamBufferMB > 0 {
		args = append(args,
			"--cache=yes",
			fmt.Sprintf("--demuxer-max-bytes=%dM", connConfig.StreamBufferMB),
		)
	} else {
		// No buffering (original behavior)
		args = append(args, "--no-cache")
	}

	// Add URL as final argument
	args = append(args, station.URLResolved)

	// Create mpv command
	p.cmd = exec.Command("mpv", args...)

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
	p.paused = false
	p.station = station
	p.stopCh = make(chan struct{})

	// Record play start for statistics (errors are non-fatal)
	if p.metadataManager != nil {
		_ = p.metadataManager.StartPlay(station.StationUUID)
	}

	// Connect to IPC socket (with retry for socket creation delay)
	go p.connectToSocket()

	// Monitor the process in a goroutine
	go p.monitor()

	// Monitor metadata for track changes
	go p.monitorMetadata()

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

		// Platform-specific connection
		var conn net.Conn
		var err error
		if runtime.GOOS == "windows" {
			// On Windows, connect via TCP
			conn, err = net.Dial("tcp", socketPath)
		} else {
			// On Unix-like systems, connect via Unix socket
			conn, err = net.Dial("unix", socketPath)
		}
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

// getProperty retrieves a property value from mpv via IPC
func (p *MPVPlayer) getProperty(name string) (interface{}, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.conn == nil {
		return nil, fmt.Errorf("not connected to mpv")
	}

	msg := map[string]interface{}{
		"command": []interface{}{"get_property", name},
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return nil, err
	}
	data = append(data, '\n')

	// Set deadline for write and read
	_ = p.conn.SetDeadline(time.Now().Add(500 * time.Millisecond))
	defer func() { _ = p.conn.SetDeadline(time.Time{}) }()

	if _, err := p.conn.Write(data); err != nil {
		return nil, err
	}

	// Read response
	buf := make([]byte, 1024)
	n, err := p.conn.Read(buf)
	if err != nil {
		return nil, err
	}

	var resp struct {
		Data  interface{} `json:"data"`
		Error string      `json:"error"`
	}
	if err := json.Unmarshal(buf[:n], &resp); err != nil {
		return nil, err
	}

	if resp.Error != "success" {
		return nil, fmt.Errorf("mpv error: %s", resp.Error)
	}

	return resp.Data, nil
}

// GetAudioBitrate returns the current audio bitrate (useful for checking signal)
func (p *MPVPlayer) GetAudioBitrate() (int, error) {
	val, err := p.getProperty("audio-bitrate")
	if err != nil {
		return 0, err
	}

	// mpv returns bitrate as float64 via JSON
	if bitrate, ok := val.(float64); ok {
		return int(bitrate), nil
	}

	return 0, nil
}

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

// GetCachedTrack returns the current track name from the in-memory cache without IPC.
// Use this in render paths to avoid potential UI jank from synchronous IPC calls.
func (p *MPVPlayer) GetCachedTrack() string {
	p.trackMu.Lock()
	defer p.trackMu.Unlock()
	return p.currentTrack
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

	// Skip empty or very short tracks (likely station name, not song)
	if len(track) < 3 {
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
	// Capture stopCh to avoid data race with Play() reassigning it
	p.mu.Lock()
	stopCh := p.stopCh
	p.mu.Unlock()

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

		case <-stopCh:
			return
		}
	}
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
	// Record play stop for statistics (errors are non-fatal)
	if p.metadataManager != nil && p.station != nil {
		_ = p.metadataManager.StopPlay(p.station.StationUUID)
	}

	// Close IPC connection
	if p.conn != nil {
		_ = p.conn.Close()
		p.conn = nil
	}

	// Remove socket file (Unix only, Windows named pipes auto-cleanup)
	if p.socketPath != "" && runtime.GOOS != "windows" {
		_ = os.Remove(p.socketPath)
	}
	p.socketPath = ""

	if p.cmd != nil && p.cmd.Process != nil {
		// Send termination signal
		// On Windows, Process.Kill() should work, but we add a timeout
		if err := p.cmd.Process.Kill(); err != nil {
			// Process may have already exited
			if !errors.Is(err, os.ErrProcessDone) {
				return fmt.Errorf("failed to stop mpv: %w", err)
			}
		}

		// Wait for process to finish (with timeout to prevent hanging)
		done := make(chan error, 1)
		go func() {
			done <- p.cmd.Wait()
		}()

		select {
		case <-done:
			// Process exited cleanly
		case <-time.After(2 * time.Second):
			// Process didn't exit in time, force kill on Windows
			if runtime.GOOS == "windows" {
				// Use taskkill as last resort on Windows
				_ = exec.Command("taskkill", "/F", "/PID", fmt.Sprintf("%d", p.cmd.Process.Pid)).Run()
			}
		}
	}

	// Signal monitoring goroutine to stop
	close(p.stopCh)

	p.playing = false
	p.paused = false
	p.station = nil
	p.cmd = nil

	// Clear track history
	p.trackMu.Lock()
	p.trackHistory = []string{}
	p.currentTrack = ""
	p.trackMu.Unlock()

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

// TogglePause toggles pause/resume state
func (p *MPVPlayer) TogglePause() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.playing {
		return fmt.Errorf("not playing")
	}

	if p.conn == nil {
		return fmt.Errorf("not connected to mpv")
	}

	// Cycle the pause property (toggles pause/unpause)
	if err := p.sendCommand([]interface{}{"cycle", "pause"}); err != nil {
		return err
	}

	// Toggle the pause state only after successful command
	p.paused = !p.paused
	return nil
}

// IsPaused returns whether the player is currently paused
func (p *MPVPlayer) IsPaused() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.paused
}

// SetMetadataManager sets the metadata manager for play statistics tracking
func (p *MPVPlayer) SetMetadataManager(mgr *storage.MetadataManager) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.metadataManager = mgr
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
