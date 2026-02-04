package storage

import (
	"errors"

	"github.com/shinokada/tera/internal/api"
)

// Errors
var (
	ErrDuplicateStation = errors.New("station already exists in list")
	ErrStationNotFound  = errors.New("station not found in list")
)

// FavoritesList represents a collection of favorite stations
type FavoritesList struct {
	Name     string        `json:"-"`
	Stations []api.Station `json:"stations"`
}

// Config represents application configuration
type Config struct {
	FavoritePath string       `json:"favorite_path"`
	CachePath    string       `json:"cache_path"`
	LastPlayed   *api.Station `json:"last_played,omitempty"`
}

// ShuffleConfig represents shuffle mode configuration
type ShuffleConfig struct {
	AutoAdvance     bool `yaml:"auto_advance"`
	IntervalMinutes int  `yaml:"interval_minutes"`
	RememberHistory bool `yaml:"remember_history"`
	MaxHistory      int  `yaml:"max_history"`
}

// DefaultShuffleConfig returns default shuffle configuration
func DefaultShuffleConfig() ShuffleConfig {
	return ShuffleConfig{
		AutoAdvance:     false,
		IntervalMinutes: 5,
		RememberHistory: true,
		MaxHistory:      5,
	}
}
