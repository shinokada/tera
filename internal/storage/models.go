package storage

import (
	"errors"

	"github.com/shinokada/tera/internal/api"
)

// Errors
var (
	ErrDuplicateStation = errors.New("station already exists in list")
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
