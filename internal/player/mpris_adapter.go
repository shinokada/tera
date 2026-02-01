package player

import (
	"github.com/go-music-players/mpris"
)

// MPRISAdapter adapts tera's MPVPlayer to the mpris.Player interface
type MPRISAdapter struct {
	player *MPVPlayer
}

// NewMPRISAdapter creates a new MPRIS adapter for tera
func NewMPRISAdapter(p *MPVPlayer) *MPRISAdapter {
	return &MPRISAdapter{
		player: p,
	}
}

// Play starts playback (resume if paused/stopped)
func (a *MPRISAdapter) Play() error {
	station := a.player.GetCurrentStation()
	if station == nil {
		return nil
	}
	return a.player.Play(station)
}

// Pause pauses playback
func (a *MPRISAdapter) Pause() error {
	if a.player.IsPlaying() {
		return a.player.Stop()
	}
	return nil
}

// Stop stops playback
func (a *MPRISAdapter) Stop() error {
	return a.player.Stop()
}

// Next plays next track (not applicable for radio)
func (a *MPRISAdapter) Next() error {
	// Radio player doesn't have "next" concept
	return nil
}

// Previous plays previous track (not applicable for radio)
func (a *MPRISAdapter) Previous() error {
	// Radio player doesn't have "previous" concept
	return nil
}

// GetPlaybackStatus returns current playback status
func (a *MPRISAdapter) GetPlaybackStatus() (mpris.PlaybackStatus, error) {
	if a.player.IsPlaying() {
		return mpris.StatusPlaying, nil
	}
	return mpris.StatusStopped, nil
}

// GetMetadata returns station metadata
func (a *MPRISAdapter) GetMetadata() (*mpris.Metadata, error) {
	station := a.player.GetCurrentStation()
	if station == nil {
		return nil, nil
	}

	metadata := &mpris.Metadata{
		TrackID: station.StationUUID,
		Title:   station.TrimName(),
	}

	// Use country and tags for additional context
	if station.Country != "" {
		metadata.Album = station.Country
	}
	if station.Tags != "" {
		metadata.Artist = []string{station.Tags}
	}

	return metadata, nil
}

// CanPlay returns true if can play
func (a *MPRISAdapter) CanPlay() bool {
	// Radio player can resume if there's a current station
	return a.player.GetCurrentStation() != nil
}

// CanPause returns true if can pause
func (a *MPRISAdapter) CanPause() bool {
	return a.player.IsPlaying()
}

// CanGoNext returns true if can go to next track
func (a *MPRISAdapter) CanGoNext() bool {
	return false // Radio player doesn't have next/previous
}

// CanGoPrevious returns true if can go to previous track
func (a *MPRISAdapter) CanGoPrevious() bool {
	return false // Radio player doesn't have next/previous
}

// CanSeek returns true if can seek
func (a *MPRISAdapter) CanSeek() bool {
	return false // Radio streams don't support seeking
}

// CanControl returns true if player can be controlled
func (a *MPRISAdapter) CanControl() bool {
	return true
}
