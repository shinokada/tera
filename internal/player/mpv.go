package player

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/shinokada/tera/v3/internal/api"
	"github.com/shinokada/tera/v3/internal/storage"
)

// PlayWithVolume starts playing a radio station with a specific volume.
// It is a thin wrapper around Play that temporarily overrides the station's
// volume field, ensuring full IPC/monitor setup identical to Play.
func (p *MPVPlayer) PlayWithVolume(station *api.Station, volume int) error {
	// Clone the station so we don't mutate the caller's value.
	cloned := *station
	cloned.Volume = &volume
	return p.Play(&cloned)
}

// playerInstanceCounter provides a process-wide unique ID for each MPVPlayer
// so that concurrent instances never collide on the same IPC socket path.
var playerInstanceCounter atomic.Uint64

// MPVPlayer manages the MPV process for playing radio streams
type MPVPlayer struct {
	cmd             *exec.Cmd
	playing         bool
	killed          bool // set by Stop() on a not-yet-playing player to reject a late Play()
	paused          bool // Pause state
	station         *api.Station
	volume          int // Current volume (0-100)
	muted           bool
	lastVolume      int // Volume before mute
	mu              sync.Mutex
	stopCh          chan struct{}
	instanceID      uint64                   // Unique ID for this player instance (socket path)
	socketPath      string                   // IPC socket path for runtime control
	conn            net.Conn                 // Connection to IPC socket
	connReader      *bufio.Reader            // Buffered reader for newline-delimited IPC responses
	nextRequestID   atomic.Uint64            // Monotonically increasing IPC request ID
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
		instanceID: playerInstanceCounter.Add(1),
	}
}

// Play starts playing a radio station
func (p *MPVPlayer) Play(station *api.Station) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	// If Stop() was called before Play() ran (race between async cmd and
	// navigation), honour the stop request and refuse to start.
	if p.killed {
		return nil
	}

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

	// Create unique socket path for IPC (platform-specific).
	// Use both PID and per-instance ID so that multiple concurrent MPVPlayer
	// instances within the same process never share a socket path — which
	// would cause one player's Stop() to remove the other player's socket.
	if runtime.GOOS == "windows" {
		// Windows: Use TCP socket for IPC (more reliable than named pipes).
		// Incorporate instanceID to guarantee a unique port per instance.
		p.socketPath = fmt.Sprintf("127.0.0.1:%d", 10000+(os.Getpid()*1000+int(p.instanceID))%50000)
	} else {
		// Unix/Linux/macOS use Unix sockets
		p.socketPath = filepath.Join(os.TempDir(), fmt.Sprintf("tera-mpv-%d-%d.sock", os.Getpid(), p.instanceID))
		// Remove any stale socket file from a previous Play() call on this instance
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

	// Validate URL scheme before passing to mpv to prevent local file access
	// via file:// or other unexpected schemes from a malicious API response.
	safeURL, err := validateStreamURL(station.URLResolved)
	if err != nil {
		return err
	}

	// Add URL as final argument
	args = append(args, safeURL)

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
		_ = p.metadataManager.StartPlay(station)
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
			p.connReader = bufio.NewReader(conn)
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

// sendCommand sends a JSON command to mpv via IPC.
// It tags every outbound message with a unique request_id and reads back
// lines until it finds the matching reply, discarding unrelated events
// (e.g. property-change notifications, or replies to prior set_property
// calls whose responses were never consumed).
// Caller must hold p.mu.
func (p *MPVPlayer) sendCommand(command []interface{}) error {
	if p.conn == nil || p.connReader == nil {
		return fmt.Errorf("not connected to mpv")
	}

	reqID := p.nextRequestID.Add(1)
	msg := map[string]interface{}{
		"command":    command,
		"request_id": reqID,
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	data = append(data, '\n')

	// Prevent UI stalls on blocked IPC writes.
	_ = p.conn.SetWriteDeadline(time.Now().Add(250 * time.Millisecond))
	defer func() { _ = p.conn.SetWriteDeadline(time.Time{}) }()

	if _, err := p.conn.Write(data); err != nil {
		return err
	}

	// Drain the reply for this command so it is never left in the buffer
	// to confuse a subsequent getProperty call.
	_ = p.conn.SetReadDeadline(time.Now().Add(250 * time.Millisecond))
	defer func() { _ = p.conn.SetReadDeadline(time.Time{}) }()

	for {
		line, err := p.connReader.ReadString('\n')
		if err != nil {
			return err
		}
		var reply ipcReply
		if jsonErr := json.Unmarshal([]byte(line), &reply); jsonErr != nil {
			continue
		}
		if reply.RequestID == reqID {
			if reply.Error != "" && reply.Error != "success" {
				return fmt.Errorf("mpv error: %s", reply.Error)
			}
			return nil
		}
		// Discard unrelated events/replies and keep reading.
	}
}

// getProperty retrieves a property value from mpv via IPC.
// It tags every request with a unique request_id and skips lines until
// the matching reply arrives, so stale set_property responses left in the
// buffer by sendCommand can never corrupt the returned value.
func (p *MPVPlayer) getProperty(name string) (interface{}, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.conn == nil || p.connReader == nil {
		return nil, fmt.Errorf("not connected to mpv")
	}

	reqID := p.nextRequestID.Add(1)
	msg := map[string]interface{}{
		"command":    []interface{}{"get_property", name},
		"request_id": reqID,
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return nil, err
	}
	data = append(data, '\n')

	// Set deadline for write and read.
	_ = p.conn.SetDeadline(time.Now().Add(500 * time.Millisecond))
	defer func() { _ = p.conn.SetDeadline(time.Time{}) }()

	if _, err := p.conn.Write(data); err != nil {
		return nil, err
	}

	// Read lines until we find the reply that matches our request_id,
	// discarding unrelated events (property notifications, prior command acks).
	for {
		line, err := p.connReader.ReadString('\n')
		if err != nil {
			return nil, err
		}
		var reply ipcReply
		if jsonErr := json.Unmarshal([]byte(line), &reply); jsonErr != nil {
			continue
		}
		if reply.RequestID != reqID {
			continue // discard: belongs to a different command or is an event
		}
		if reply.Error != "success" {
			return nil, fmt.Errorf("mpv error: %s", reply.Error)
		}
		return reply.Data, nil
	}
}

// ipcReply is the subset of fields present in every mpv IPC response line.
type ipcReply struct {
	RequestID uint64      `json:"request_id"`
	Data      interface{} `json:"data"`
	Error     string      `json:"error"`
}

// validateStreamURL checks that the URL uses a safe streaming scheme and
// returns the trimmed, validated URL. This prevents a malicious or compromised
// API response from supplying a file:// or fd:// URL that would cause mpv to
// open local resources. A URL with leading/trailing whitespace is trimmed so
// the sanitized value is forwarded to mpv rather than the raw input.
func validateStreamURL(rawURL string) (string, error) {
	cleaned := strings.TrimSpace(rawURL)
	if cleaned == "" {
		return "", fmt.Errorf("station URL is empty")
	}

	u, err := url.Parse(cleaned)
	if err != nil || u.Scheme == "" {
		return "", fmt.Errorf("station URL is invalid: %q", rawURL)
	}

	switch strings.ToLower(u.Scheme) {
	case "http", "https", "rtsp", "rtmp", "rtsps", "rtmps":
		if u.Host == "" {
			return "", fmt.Errorf("station URL must include a host: %q", rawURL)
		}
		return cleaned, nil
	default:
		return "", fmt.Errorf("station URL has disallowed scheme (must be http/https/rtsp/rtmp/rtsps/rtmps): %q", rawURL)
	}
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
		// Not yet playing — mark as killed so any in-flight Play() cmd
		// that arrives after this Stop() will not start the player.
		p.killed = true
		return nil
	}

	return p.stopInternal()
}

// cleanupResourcesLocked releases IPC connection, socket file, and all player
// state fields. Must be called with p.mu already held.
func (p *MPVPlayer) cleanupResourcesLocked() {
	// Close IPC connection
	if p.conn != nil {
		_ = p.conn.Close()
		p.conn = nil
		p.connReader = nil
	}

	// Remove socket file (Unix only, Windows named pipes auto-cleanup)
	if p.socketPath != "" && runtime.GOOS != "windows" {
		_ = os.Remove(p.socketPath)
	}
	p.socketPath = ""

	// Signal monitorMetadata goroutine to stop
	close(p.stopCh)

	p.playing = false
	p.killed = false // reset so the player instance may be reused after a full stop
	p.paused = false
	p.station = nil
	p.cmd = nil

	// Clear track history
	p.trackMu.Lock()
	p.trackHistory = []string{}
	p.currentTrack = ""
	p.trackMu.Unlock()
}

// stopInternal stops playback without locking (internal use)
func (p *MPVPlayer) stopInternal() error {
	// Record play stop for statistics (errors are non-fatal)
	if p.metadataManager != nil && p.station != nil {
		_ = p.metadataManager.StopPlay(p.station.StationUUID)
	}

	if p.cmd != nil && p.cmd.Process != nil {
		if runtime.GOOS == "windows" {
			// On Windows, mpv installed via package managers (e.g. scoop) uses a
			// shim executable that spawns the real mpv as a child process.
			// Process.Kill() only terminates the shim, leaving the real mpv running
			// as an orphan. taskkill with /T kills the entire process tree so all
			// descendant processes (including the real mpv) are terminated.
			_ = exec.Command("taskkill", "/F", "/T", "/PID", fmt.Sprintf("%d", p.cmd.Process.Pid)).Run()
			// Also call Kill to ensure the Go process handle is cleaned up.
			_ = p.cmd.Process.Kill()
		} else {
			// Send termination signal
			if err := p.cmd.Process.Kill(); err != nil {
				// Process may have already exited
				if !errors.Is(err, os.ErrProcessDone) {
					return fmt.Errorf("failed to stop mpv: %w", err)
				}
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
			// Timeout: process did not exit within grace period
		}
	}

	p.cleanupResourcesLocked()
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

// Done returns a channel that is closed when playback ends for any reason:
// a natural stream drop, an external process kill, or an explicit Stop call.
// Callers must not rely on this channel to distinguish between these cases;
// it only signals that the player is no longer active.
func (p *MPVPlayer) Done() <-chan struct{} {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.stopCh
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
		// Process ended naturally (stream drop, network loss, error).
		// Guard with p.playing: stopInternal may have raced us here (it kills the
		// process, then calls cleanupResourcesLocked which sets playing=false and
		// closes stopCh). If it won the lock first, skip to avoid a double-close.
		p.mu.Lock()
		if p.playing {
			if p.metadataManager != nil && p.station != nil {
				_ = p.metadataManager.StopPlay(p.station.StationUUID)
			}
			p.cleanupResourcesLocked()
		}
		p.mu.Unlock()
	case <-p.stopCh:
		// Stop was called, process already killed and resources cleaned up
		return
	}
}
